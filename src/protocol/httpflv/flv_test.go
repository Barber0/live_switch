package httpflv

import (
	"encoding/binary"
	"fmt"
	"testing"
)

func TestAlpha(t *testing.T) {
	a := []byte{0, 1, 2, 3}
	fmt.Println(binary.BigEndian.Uint32(a))
}
