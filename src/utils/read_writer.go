package utils

import (
	"bufio"
	"encoding/binary"
	"io"
)

const (
	BUF_SIZE = 1 << 12
)

type ReadWriter struct {
	*bufio.ReadWriter
}

func NewReadWriter(rw io.ReadWriter) *ReadWriter {
	return &ReadWriter{
		bufio.NewReadWriter(
			bufio.NewReaderSize(rw, BUF_SIZE),
			bufio.NewWriterSize(rw, BUF_SIZE)),
	}
}

func (rw *ReadWriter) ReadUint32BE(size int) (uint32, error) {
	bs4 := make([]byte, 4)
	_, err := rw.Read(bs4[4-size:])
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint32(bs4), nil
}

func (rw *ReadWriter) ReadUint32LE(size int) (uint32, error) {
	bs4 := make([]byte, 4)
	_, err := rw.Read(bs4[:size])
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(bs4), nil
}
