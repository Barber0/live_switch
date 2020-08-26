package video

type DataType byte

const (
	DATA_TYPE_VIDEO DataType = iota
	DATA_TYPE_AUDIO
	DATA_TYPE_META
)

const (
	SOUND_FMT_AAC = 10
)

const (
	_ = iota
	FRAME_TYPE_KEY
	FRAME_TYPE_INTER
)

const (
	AAC_PKT_TYPE_SEQHDR = iota
	AAC_PKT_TYPE_RAW
)

const (
	AVC_PKT_TYPE_SEQHDR = iota
	AVC_PKT_TYPE_NALU
	AVC_PKT_TYPE_EOS
)

type Packet struct {
	DataType
	Header    PacketHeader
	StreamId  uint32
	Data      []byte
	Timestamp uint32
}

type PacketHeader interface{}

type AudioPacketHeader interface {
	PacketHeader
	SoundFmt() uint8
	AACPacketType() uint8
}

type VideoPacketHeader interface {
	PacketHeader
	IsKeyFrame() bool
	IsSeq() bool
}
