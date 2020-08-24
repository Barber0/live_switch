package utils

import "math/rand"

func RandBytes(bsLen int) []byte {
	bs := make([]byte, bsLen)
	for i := 0; i < bsLen; i++ {
		bs[i] = byte(rand.Intn(255))
	}
	return bs
}
