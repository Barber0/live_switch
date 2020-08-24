package rtmp

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"live/src/utils"
	"net"
	"sync"
	"time"
)

type connection struct {
	tcpConn         net.Conn
	chunkSize       uint32
	remoteChunkSize uint32
	windowAckSize   uint32
	windowReceived  uint32
	pool            *utils.Pool
	bufPool         *sync.Pool
	chunkMap        map[byte]*chunk
	handshaker      *handshake
	transactionId   int
	connInfo        utils.AMFObj
	publishInfo     publishInfo
	done            bool
	prevCh          chunk
}

type publishInfo struct {
	name        string
	publishType string
}

func newConn(c net.Conn) *connection {
	conn := &connection{
		tcpConn:         c,
		chunkSize:       128,
		remoteChunkSize: 128,
		windowAckSize:   2500000,
		handshaker:      &handshake{},
		chunkMap:        make(map[byte]*chunk),
		bufPool: &sync.Pool{
			New: func() interface{} {
				return bytes.NewBuffer(nil)
			},
		},
	}
	conn.pool = utils.NetPool(int(conn.chunkSize))
	return conn
}

func (c *connection) Close() {
	pubName := c.getPublisherName()
	if publisherMap.Has(pubName) {
		publisherMap.Remove(pubName)
	}
	c.tcpConn.Close()
}

func (c *connection) handshake() (err error) {

	// C0C1
	bs1537 := c.pool.Get(1537)
	// bsC0C1 := make([]byte, 1537)
	_, err = c.tcpConn.Read(*bs1537)
	if err != nil {
		return
	}

	buf := bytes.NewBuffer(*bs1537)

	if c.handshaker.version, err = buf.ReadByte(); err != nil {
		return
	}

	timeC1Bs := buf.Next(4)
	c.handshaker.timeC1 = binary.BigEndian.Uint32(timeC1Bs)

	buf.Next(4)
	c.handshaker.randomBytesC1 = buf.Bytes()

	// send S0S1S2
	buf.Reset()
	// s0s1s2Buf := bytes.NewBuffer(nil)
	buf.WriteByte(c.handshaker.version)

	timeS1 := uint32(time.Now().Unix())
	timeS1Bs := make([]byte, 4)
	binary.BigEndian.PutUint32(timeS1Bs, timeS1)
	buf.Write(timeS1Bs)

	buf.Write(make([]byte, 4))

	randBsS1 := utils.RandBytes(1528)
	buf.Write(randBsS1)

	buf.Write(timeC1Bs)

	timeS2 := uint32(time.Now().Unix())
	timeS2Bs := make([]byte, 4)
	binary.BigEndian.PutUint32(timeS2Bs, timeS2)
	buf.Write(timeS2Bs)

	randBsS2 := utils.RandBytes(1528)
	buf.Write(randBsS2)

	if _, err = buf.WriteTo(c.tcpConn); err != nil {
		return
	}

	// C2
	bsC2 := make([]byte, 1536)
	_, err = c.tcpConn.Read(bsC2)
	if err != nil {
		return
	}
	// var bsC2 []byte
	// bsC2, err = utils.NetReadBytes(decoder.conn, 1536)

	buf.Reset()
	buf.Write(bsC2)

	timeC2Bs := buf.Next(4)
	c.handshaker.timeC2 = binary.BigEndian.Uint32(timeC2Bs)

	buf.Next(4)

	c.handshaker.randomBytesC2 = buf.Bytes()

	timeC2Match := c.handshaker.timeC2 == timeS1
	randC2Match := bytes.Equal(c.handshaker.randomBytesC2, randBsS1)

	if !(timeC2Match && randC2Match) {
		return errors.New("C2 info not match")
	}
	return
}

func (c *connection) readMsg() (retCh *chunk, err error) {
	defer func() {
		c.prevCh = *retCh
	}()
	var ch *chunk
	for {
		bs1 := c.pool.Get(1)
		_, err = c.tcpConn.Read(*bs1)
		if err != nil {
			return
		}

		tmpFmt := (*bs1)[0] >> 6

		csid1 := (*bs1)[0] & 0x3f
		c.pool.ResetByBsp(bs1)

		var ok bool
		if ch, ok = c.chunkMap[csid1]; !ok || ch.finished {
			ch = &chunk{}
			c.chunkMap[csid1] = ch
		}

		switch csid1 {
		case 0:
			_, err = c.tcpConn.Read(*bs1)
			if err != nil {
				return
			}
			ch.csid = uint32((*bs1)[0]) + 64
			c.pool.ResetByBsp(bs1)
		case 1:
			bs4 := c.pool.Get(4)
			_, err = c.tcpConn.Read((*bs4)[:2])
			if err != nil {
				return
			}
			ch.csid = binary.LittleEndian.Uint32(*bs4) + 64
			c.pool.ResetByBsp(bs4)
		default:
			ch.csid = uint32(csid1)
		}

		if csid1 > 4 || csid1 < 2 {
			fmt.Println("test point")
		}

		if ch.csid > 4 || ch.csid < 2 {
			fmt.Println("test point1")
		}

		ch.basicHeader.tmpFmt = tmpFmt

		if err = c.readChunk(ch); err != nil {
			return
		}

		if ch.finished {
			retCh = ch
			//delete(c.chunkMap, csid1)
			return
		}
	}
}

func (c *connection) readChunk(ch *chunk) (err error) {
	if ch.remain > 0 && ch.tmpFmt != 3 {
		return errors.New("message finished, but still have remain data")
	}

	bs4 := c.pool.Get(4)

	if ch.tmpFmt < 3 {
		ch.format = ch.tmpFmt

		_, err = c.tcpConn.Read((*bs4)[1:])
		if err != nil {
			return
		}
		ch.timestamp += binary.BigEndian.Uint32(*bs4)
		c.pool.ResetByBsp(bs4)

		if ch.tmpFmt < 2 {
			if _, err = c.tcpConn.Read(*bs4); err != nil {
				return
			}
			lenAndTypeId := binary.BigEndian.Uint32(*bs4)
			ch.length = lenAndTypeId >> 8

			if ch.length > c.remoteChunkSize {
				fmt.Println("too long")
			}

			ch.typeId = byte(lenAndTypeId & 0xff)

			ch.data = make([]byte, ch.length)
			ch.index = 0
			ch.remain = ch.length
			c.pool.ResetByBsp(bs4)

			if ch.tmpFmt < 1 {
				if _, err = c.tcpConn.Read(*bs4); err != nil {
					return
				}
				ch.streamId = binary.LittleEndian.Uint32(*bs4)
				c.pool.ResetByBsp(bs4)
			}

			if ch.timestamp == 0xffffff {
				_, err = c.tcpConn.Read(*bs4)
				if err != nil {
					return
				}
				ch.timestamp = binary.BigEndian.Uint32(*bs4)
				c.pool.ResetByBsp(bs4)
				ch.hasExtTs = true
			}
		} else {
			ch.data = make([]byte, ch.length)
		}
	}

	if ch.hasExtTs {
		rw := bufio.NewReader(c.tcpConn)
		var bts []byte
		bts, err = rw.Peek(4)
		if err != nil {
			return err
		}
		tmpts := binary.BigEndian.Uint32(bts)
		if tmpts == ch.timestamp {
			rw.Discard(4)
		}
	}

	size := ch.remain
	if ch.remain > c.remoteChunkSize {
		size = c.remoteChunkSize
	}
	if _, err = c.tcpConn.Read(ch.data[ch.index : ch.index+int(size)]); err != nil {
		return
	}
	ch.index += int(size)
	ch.remain -= size

	if ch.remain == 0 {
		ch.finished = true
	}

	return
}

func (c *connection) handleCtrlMsg(msg *chunk) {
	switch msg.typeId {
	case TYPE_ID_SET_CHUNK_SIZE:
		c.remoteChunkSize = binary.BigEndian.Uint32(msg.data)
	case TYPE_ID_WINDOW_ACK_SIZE:
		c.windowAckSize = binary.BigEndian.Uint32(msg.data)
	}
}

func (c *connection) handleCmdMsg(msg *chunk) (err error) {
	if msg.typeId == TYPE_ID_CMD_MSG_AMF3 {
		msg.data = msg.data[1:]
	}
	var datas []interface{}
	datas, err = utils.GetAMFHandler().DecodeBatch(msg.data, utils.AMF0)
	if err != nil {
		return
	}
	switch datas[0].(type) {
	case string:
		if _, ok := cmdHandlers[datas[0].(string)]; ok {
			err = cmdHandlers[datas[0].(string)](c, msg, datas[1:])
			if err != nil {
				return
			}
		}
	}
	return
}

func (c *connection) ack(size uint32) error {
	c.windowReceived += size
	if c.windowReceived >= c.windowAckSize {
		bs := make([]byte, 4)
		binary.BigEndian.PutUint32(bs, c.windowReceived)
		ch := newChunk(TYPE_ID_ACK, CSID_AUTO, 0, bs)
		wbuf := bytes.NewBuffer(nil)
		if err := c.writeChunk(ch); err != nil {
			return err
		}
		if _, err := wbuf.WriteTo(c.tcpConn); err != nil {
			return err
		}
		c.windowReceived = 0
	}
	return nil
}

func (c *connection) writeAmfMsg(typeId byte, csid, streamId uint32, args ...interface{}) (err error) {
	var bs []byte
	var amfVer utils.AMFVersion
	switch typeId {
	case TYPE_ID_CMD_MSG_AMF3, TYPE_ID_DATA_MSG_AMF3, TYPE_ID_SHARED_OBJ_MSG_AMF3:
		amfVer = utils.AMF3
	default:
		amfVer = utils.AMF0
	}
	bs, err = utils.GetAMFHandler().EncodeBatch(amfVer, args...)
	if err != nil {
		return
	}

	ch := newChunk(typeId, csid, streamId, bs)
	err = c.writeChunk(ch)
	if err != nil {
		return err
	}
	return
}

func (c *connection) writeChunk(ch *chunk) (err error) {
	if ch.typeId == TYPE_ID_SET_CHUNK_SIZE {
		c.chunkSize = binary.BigEndian.Uint32(ch.data)
	}
	numChunk := int(ch.length/c.chunkSize + 1)
	for i := 0; i < numChunk; i++ {
		if i > 0 {
			ch.format = 3
		}
		if err = c.writeChunkHeader(ch); err != nil {
			return
		}
		inc := c.chunkSize
		start := uint32(i) * c.chunkSize
		if tmpInc := ch.length - start; tmpInc <= inc {
			inc = tmpInc
		}
		end := start + inc
		_, err = c.tcpConn.Write(ch.data[start:end])
		if err != nil {
			return
		}
	}
	return nil
}

func (c *connection) writeChunkHeader(ch *chunk) (err error) {
	var bsBH []byte
	fmtH := ch.format << 6
	if ch.csid < 64 {
		bsBH = make([]byte, 1)
		bsBH[0] = fmtH | byte(ch.csid)
	} else if tmpCsid := ch.csid - 64; tmpCsid < 0x100 {
		bsBH = make([]byte, 2)
		bsBH[0] = fmtH
		bsBH[1] = byte(tmpCsid)
	} else if tmpCsid < 0x10000 {
		bsBH = make([]byte, 3)
		bsBH[0] = fmtH | 1
		bsBH[1] = byte(tmpCsid)
		bsBH[2] = byte(tmpCsid >> 8)
	}
	_, err = c.tcpConn.Write(bsBH)
	if err != nil {
		return
	}

	bsMH1 := make([]byte, 4)
	hasExtTs := false
	if ch.format < 3 {
		tsLimit := uint32(0xffffff)
		hasExtTs = ch.timestamp >= tsLimit
		if hasExtTs {
			binary.BigEndian.PutUint32(bsMH1, tsLimit)
		} else {
			binary.BigEndian.PutUint32(bsMH1, ch.timestamp)
		}
		_, err = c.tcpConn.Write(bsMH1[1:])
		if err != nil {
			return
		}
	}

	bsMH2 := make([]byte, 4)
	if ch.format < 2 {
		lineVal := ch.length<<8 | uint32(ch.typeId)
		binary.BigEndian.PutUint32(bsMH2, lineVal)
		_, err = c.tcpConn.Write(bsMH2)
		if err != nil {
			return
		}
	}

	bsMH3 := make([]byte, 4)
	if ch.format < 1 {
		binary.LittleEndian.PutUint32(bsMH3, ch.streamId)
		_, err = c.tcpConn.Write(bsMH3)
		if err != nil {
			return
		}
	}

	if hasExtTs {
		bsExtTs := make([]byte, 4)
		binary.BigEndian.PutUint32(bsExtTs, ch.timestamp)
		_, err = c.tcpConn.Write(bsExtTs)
		if err != nil {
			return
		}
	}
	return
}

func (c *connection) getPublisherName() (URL string) {
	if tmpTcUrl, ok := c.connInfo["tcUrl"]; ok {
		URL = tmpTcUrl.(string) + "/" + c.publishInfo.name
	} else {
		URL = c.publishInfo.name
	}
	return
}
