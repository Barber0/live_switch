package video

type TagParser struct {
}

func (tp *TagParser) Parse(p *Packet) error {
	var tag Tag
	err := tag.ParsePacketHeader(p.Data, p.DataType)
	if err != nil {
		return err
	}
	p.Header = &tag
	return nil
}
