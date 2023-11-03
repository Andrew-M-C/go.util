//go:build go1.20
// +build go1.20

package unsafe

import (
	"unsafe"
)

// BtoS []byte to string
func BtoS(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// StoB string to []byte
func StoB(s string) []byte {
	// Reference:
	//  - [与日俱进，在 Go 1.20 中这种高效转换的方式又变了](https://colobu.com/2022/09/06/string-byte-convertion/)
	//  - [非类型安全指针](https://gfw.go101.org/article/unsafe.html)
	if s == "" {
		return nil
	}
	return unsafe.Slice(unsafe.StringData(s), len(s))
}
