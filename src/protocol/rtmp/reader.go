package rtmp

import (
	"live/src/video"
)

type streamReader interface {
	Read(*video.Packet) error
	Close()
}

var _ streamReader = &videoReader{}

type videoReader struct {
	conn *connection
}

func NewVideoReader(conn *connection) *videoReader {
	r := &videoReader{
		conn: conn,
	}
	return r
}

func (v *videoReader) Close() {
	v.conn.Close()
}

func (v *videoReader) Read(p *video.Packet) (err error) {
	var ch *chunk
GET_CH:
	for {
		if ch, err = v.conn.readMsg(); err != nil {
			return err
		}
		switch ch.typeId {
		case TYPE_ID_VIDEO_MSG:
			p.DataType = video.DATA_TYPE_VIDEO
			break GET_CH
		case TYPE_ID_AUDIO_MSG:
			p.DataType = video.DATA_TYPE_AUDIO
			break GET_CH
		case TYPE_ID_DATA_MSG_AMF0, TYPE_ID_DATA_MSG_AMF3:
			p.DataType = video.DATA_TYPE_META
			break GET_CH
			//default:
			//	return fmt.Errorf("invalid chunk type: %d", ch.typeId)
		}
	}
	p.StreamId = ch.streamId
	p.Data = ch.data
	p.Timestamp = ch.timestamp
	return nil
}
