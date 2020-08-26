package httpflv

import (
	"encoding/binary"
	"fmt"
	"github.com/sirupsen/logrus"
	"live/src/protocol/rtmp"
	"live/src/utils"
	"live/src/video"
	"net/http"
)

const (
	HEADER_LEN    = 11
	MAX_QUEUE_LEN = (1 << 10) * 100
)

var _ rtmp.Consumer = &FLVWriter{}

type FLVWriter struct {
	//utils.TimeMeter
	packetQueue chan *video.Packet
	stopChan    chan struct{}
	ctx         http.ResponseWriter
	buf         []byte
	closed      bool
	status      bool
}

func (fw *FLVWriter) Close() {
	logrus.Debug("flv writer closed")
	if !fw.closed {
		close(fw.packetQueue)
		close(fw.stopChan)
	}
	fw.closed = true
}

func (fw *FLVWriter) GetStatus() bool {
	return fw.status
}

func (fw *FLVWriter) SetStatus(v bool) {
	fw.status = v
}

func NewFLVWriter(ctx http.ResponseWriter) *FLVWriter {
	fw := &FLVWriter{
		packetQueue: make(chan *video.Packet, MAX_QUEUE_LEN),
		ctx:         ctx,
		buf:         make([]byte, HEADER_LEN),
	}
	fw.ctx.Write([]byte{0x46, 0x4c, 0x56, 0x01, 0x05, 0x00, 0x00, 0x00, 0x09})
	binary.BigEndian.PutUint32(fw.buf[:4], 0)
	fw.ctx.Write(fw.buf[:4])
	go func() {
		if err := fw.sendPacket(); err != nil {
			logrus.Error("send packet err: ", err)
			fw.closed = true
		}
	}()
	return fw
}

func (fw *FLVWriter) Name() string {
	return "http_flv_writer"
}

func (fw *FLVWriter) Write(p *video.Packet) (err error) {
	if fw.closed {
		return fmt.Errorf("flv write source closed")
	}

	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("flv writer write panic: %v", e)
		}
	}()

	if len(fw.packetQueue) >= MAX_QUEUE_LEN-84 {
		fw.dropPacket()
	} else {
		fw.packetQueue <- p
	}
	return
}

func (fw *FLVWriter) dropPacket() {
	logrus.Warningf("[%v] packet queue max!!!", fw.Name())
	for i := 0; i < MAX_QUEUE_LEN-84; i++ {
		tmpPkt, ok := <-fw.packetQueue
		if ok {

			if len(fw.packetQueue) > MAX_QUEUE_LEN-10 {
				<-fw.packetQueue
			}
			<-fw.packetQueue

			if tmpPkt.DataType == video.DATA_TYPE_AUDIO {
				//fw.packetQueue <- tmpPkt
			}
		}
	}
	logrus.Debug("packet queue len: ", len(fw.packetQueue))
}

func (fw *FLVWriter) Wait() {
	<-fw.stopChan
}

func (fw *FLVWriter) sendPacket() error {
	for {
		p, ok := <-fw.packetQueue
		if ok {
			//fw.SetPreTime()
			hbts := fw.buf[:HEADER_LEN]
			typeID := rtmp.TYPE_ID_VIDEO_MSG
			switch p.DataType {
			case video.DATA_TYPE_VIDEO:
				typeID = rtmp.TYPE_ID_VIDEO_MSG
			case video.DATA_TYPE_AUDIO:
				typeID = rtmp.TYPE_ID_AUDIO_MSG
			case video.DATA_TYPE_META:
				var err error
				typeID = rtmp.TYPE_ID_CMD_MSG_AMF0
				if p.Data, err = utils.GetAMFHandler().MetaDataReform(p.Data, utils.META_DATA_REFORM_FLAG_DEL); err != nil {
					return err
				}
			}

			dataLen := len(p.Data)

			preDataLen := dataLen + HEADER_LEN
			timestamp := p.Timestamp
			//timestamp := p.Timestamp + fw.BaseTimestamp()
			timestampBase := timestamp & 0xffffff
			timestampExt := timestamp >> 24 & 0xff

			binary.BigEndian.PutUint32(hbts[:4], uint32(typeID)<<24|uint32(dataLen))
			binary.BigEndian.PutUint32(hbts[4:8], timestampBase<<8|timestampExt)

			if _, err := fw.ctx.Write(hbts); err != nil {
				return err
			}

			if _, err := fw.ctx.Write(p.Data); err != nil {
				return err
			}

			binary.BigEndian.PutUint32(hbts[:4], uint32(preDataLen))
			if _, err := fw.ctx.Write(hbts[:4]); err != nil {
				return err
			}

		} else {
			return fmt.Errorf("closed")
		}
	}
	return nil
}
