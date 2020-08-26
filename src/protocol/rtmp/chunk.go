package rtmp

type chunk struct {
	format byte
	basicHeader
	messageHeader
	data []byte

	finished bool
	index    int
	remain   uint32
}

type basicHeader struct {
	csid   uint32
	tmpFmt byte
}

type messageHeader struct {
	timestamp uint32
	timeIncr  uint32
	length    uint32
	typeId    byte
	streamId  uint32
	extTs     uint32
	hasExtTs  bool
}

func newChunk(typeId byte, csid, streamId uint32, data []byte) *chunk {
	ch := &chunk{
		format: 0,
		messageHeader: messageHeader{
			streamId: streamId,
			typeId:   typeId,
			length:   uint32(len(data)),
		},
		data: data,
	}
	if csid == CSID_AUTO {
		switch typeId {
		case TYPE_ID_SET_CHUNK_SIZE,
			TYPE_ID_ABORT_MSG,
			TYPE_ID_ACK,
			TYPE_ID_WINDOW_ACK_SIZE,
			TYPE_ID_SET_PEER_BANDWIDTH:
			ch.basicHeader.csid = CSID_PRO_CTRL
		case TYPE_ID_CMD_MSG_AMF0, TYPE_ID_CMD_MSG_AMF3:
			ch.basicHeader.csid = CSID_CMD
		case TYPE_ID_DATA_MSG_AMF0, TYPE_ID_DATA_MSG_AMF3, TYPE_ID_VIDEO_MSG:
			ch.basicHeader.csid = CSID_VIDEO
		case TYPE_ID_AUDIO_MSG:
			ch.basicHeader.csid = CSID_AUDIO
		}
	} else {
		ch.basicHeader.csid = csid
	}

	return ch
}

func (ch *chunk) newData() {
	ch.finished=false
	ch.index = 0
	ch.remain = ch.length
	ch.data = make([]byte, ch.length)
}
