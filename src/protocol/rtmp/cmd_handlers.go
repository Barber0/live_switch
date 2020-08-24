package rtmp

import (
	"encoding/binary"
	"live/src/utils"

	amf "github.com/Barber0/goamf-1"
)

var cmdHandlers = map[string]func(*connection, *chunk, []interface{}) error{
	CMD_CONNECT:       cmdConnectHandler,
	CMD_CREATE_STREAM: cmdCreateStreamHandler,
	CMD_PUBLISH:       cmdPublishHandler,
}

func cmdConnectHandler(conn *connection, ch *chunk, datas []interface{}) (err error) {
	for _, d := range datas {
		switch d.(type) {
		case float64:
			conn.transactionId = int(d.(float64))
		case amf.Object:
			objmap := d.(amf.Object)
			conn.connInfo = utils.AMFObj(objmap)
		}
	}

	bs4 := make([]byte, 4)
	binary.BigEndian.PutUint32(bs4, 2500000)
	respCh := newChunk(TYPE_ID_WINDOW_ACK_SIZE, CSID_AUTO, 0, bs4)
	if err = conn.writeChunk(respCh); err != nil {
		return
	}

	bs5 := make([]byte, 5)
	binary.BigEndian.PutUint32(bs5[:4], 2500000)
	bs5[4] = 2
	respCh = newChunk(TYPE_ID_SET_PEER_BANDWIDTH, CSID_AUTO, 0, bs5)
	if err = conn.writeChunk(respCh); err != nil {
		return
	}

	bs4 = make([]byte, 4)
	targetChunkSize := uint32(1024)
	binary.BigEndian.PutUint32(bs4, targetChunkSize)
	respCh = newChunk(TYPE_ID_SET_CHUNK_SIZE, CSID_AUTO, 0, bs4)
	if err = conn.writeChunk(respCh); err != nil {
		return
	}
	conn.chunkSize = targetChunkSize

	resp := make(map[string]interface{})
	resp["fmsVer"] = "FMS/3,0,1,123"
	resp["capabilities"] = 31

	event := make(map[string]interface{})
	event["level"] = "status"
	event["code"] = "NetConnection.Connect.Success"
	event["description"] = "Connection succeeded."
	if enc, ok := conn.connInfo["objectEncoding"]; ok {
		event["objectEncoding"] = enc
	} else {
		event["objectEncoding"] = 0
	}

	err = conn.writeAmfMsg(TYPE_ID_CMD_MSG_AMF0, CSID_CMD, ch.streamId, "_result", conn.transactionId, resp, event)
	return
}

func cmdCreateStreamHandler(conn *connection, ch *chunk, datas []interface{}) (err error) {
	for _, v := range datas {
		switch v.(type) {
		case float64:
			conn.transactionId = int(v.(float64))
		}
	}
	err = conn.writeAmfMsg(TYPE_ID_CMD_MSG_AMF0, CSID_CMD, ch.streamId, "_result", conn.transactionId, nil, 1)
	return
}

func cmdPublishHandler(conn *connection, ch *chunk, datas []interface{}) (err error) {
	for k, v := range datas {
		switch v.(type) {
		case float64:
			conn.transactionId = int(v.(float64))
		case string:
			switch k {
			case 2:
				conn.publishInfo.name = v.(string)
			case 3:
				conn.publishInfo.publishType = v.(string)
			}
		}
	}

	event := utils.AMFObj{
		"level":       "status",
		"code":        "NetStream.Publish.Start",
		"description": "Start publishing",
	}
	err = conn.writeAmfMsg(TYPE_ID_CMD_MSG_AMF0, CSID_CMD, ch.streamId, "onStatus", 0, nil, event)
	if err != nil {
		return
	}

	pubName := conn.getPublisherName()
	pub := newPublisher(pubName)
	publisherMap.Set(pubName, pub)
	conn.done = true
	return
}
