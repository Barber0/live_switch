package rtmp

type handshake struct {
	version       byte
	timeC1        uint32
	randomBytesC1 []byte
	timeC2        uint32
	time2C2       uint32
	randomBytesC2 []byte
}
