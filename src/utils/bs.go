package utils

import "fmt"

func Eq(a, b []byte) bool {
	if len(a) != len(b) {
		fmt.Println("len not match")
		return false
	}
	match := true
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			fmt.Printf("i: %d\ta: %d\tb: %d\n", i, a[i], b[i])
			match = false
		}
	}
	return match
}
