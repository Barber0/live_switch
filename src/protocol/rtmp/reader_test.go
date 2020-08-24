package rtmp_test

import (
	"encoding/binary"
	"fmt"
	"testing"
)

func TestNewVideoReader(t *testing.T) {
	a := []byte{1, 2, 0, 0}
	fmt.Println(binary.LittleEndian.Uint32(a))
}
