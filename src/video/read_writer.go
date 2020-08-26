package video

type Writer interface {
	Write(*Packet) error
}

type Reader interface {
	Read(*Packet) error
}
