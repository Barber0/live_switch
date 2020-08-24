package video

type DataType byte

const (
	DATA_TYPE_VIDEO DataType = iota
	DATA_TYPE_AUDIO
	DATA_TYPE_META
)

type Packet struct {
	DataType
	StreamId  uint32
	Data      []byte
	Timestamp uint32
}
