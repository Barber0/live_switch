package video

import "fmt"

type Tag struct {
	soundFmt        uint8
	soundRate       uint8
	soundSize       uint8
	soundType       uint8
	aacPktType      uint8
	avcPktType      uint8
	frameType       uint8
	codecID         uint8
	compositionTime uint32
}

func (t *Tag) ParsePacketHeader(data []byte, dtype DataType) error {
	switch dtype {
	case DATA_TYPE_VIDEO:
		return t.parseVideoTag(data)
	default:
		return t.parseAudioTag(data)
	}
}

func (t *Tag) parseVideoTag(data []byte) error {
	if len(data) < 5 {
		return fmt.Errorf("invalid video packet len[%d]", len(data))
	}
	flags := data[0]
	t.frameType = flags >> 4
	t.codecID = flags & 0xf

	switch t.frameType {
	case FRAME_TYPE_INTER, FRAME_TYPE_KEY:
		t.avcPktType = data[1]
		for i := 0; i < 3; i++ {
			t.compositionTime = t.compositionTime<<8 + uint32(data[2+i])
		}
	}
	return nil
}

func (t *Tag) parseAudioTag(data []byte) error {
	if len(data) < 1 {
		return fmt.Errorf("invalid audio packet len")
	}
	flags := data[0]
	t.soundFmt = flags >> 4
	t.soundRate = (flags >> 2) & 3
	t.soundSize = (flags >> 1) & 1
	t.soundType = flags & 1
	switch t.soundFmt {
	case SOUND_FMT_AAC:
		t.aacPktType = data[1]
	}
	return nil
}

func (t *Tag) IsKeyFrame() bool {
	return t.frameType == FRAME_TYPE_KEY
}

func (t *Tag) IsSeq() bool {
	return t.IsKeyFrame() && t.avcPktType == AVC_PKT_TYPE_SEQHDR
}

func (t *Tag) SoundFmt() uint8 {
	return t.soundFmt
}

func (t *Tag) AACPacketType() uint8 {
	return t.aacPktType
}
