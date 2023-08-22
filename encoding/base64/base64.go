// Package base64 提供原生包 base64 的简单扩展封装
package base64

import (
	"bytes"
	"encoding/base64"
	"unsafe"
)

// StdEncoding 大致等同于 encoding/base64.StdEncoder, 不同的是在 Decode 的时候, 会自动补齐 padding
var StdEncoding Encoding = stdEncodeDecoder{}

// Encoding 表示 base64 Encoder
type Encoding interface {
	Decode(dst, src []byte) (n int, err error)
	DecodeString(s string) ([]byte, error)
	DecodedLen(n int) int
	Encode(dst, src []byte)
	EncodeToString(src []byte) string
	EncodedLen(n int) int
}

type stdEncodeDecoder struct{}

func fillBytesPadding(src []byte) []byte {
	remain := len(src) % 3
	if remain == 0 {
		return src
	}

	res := make([]byte, len(src)+remain)
	res[len(res)-1] = byte(base64.StdPadding)
	res[len(res)-2] = byte(base64.StdPadding)
	copy(res, src)
	return res
}

func fillStringPadding(s string) string {
	remain := 4 - len(s)%4
	if remain == 4 {
		return s
	}

	buff := bytes.NewBuffer(make([]byte, 0, len(s)+remain))
	buff.WriteString(s)
	for i := 0; i < remain; i++ {
		buff.WriteRune(base64.StdPadding)
	}
	return bToS(buff.Bytes())
}

func bToS(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func (stdEncodeDecoder) Decode(dst, src []byte) (n int, err error) {
	return base64.StdEncoding.Decode(dst, fillBytesPadding(src))
}

func (stdEncodeDecoder) DecodeString(s string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(fillStringPadding(s))
}

func (stdEncodeDecoder) DecodedLen(n int) int {
	return base64.StdEncoding.DecodedLen(n)
}

func (stdEncodeDecoder) Encode(dst, src []byte) {
	base64.StdEncoding.Encode(dst, src)
}

func (stdEncodeDecoder) EncodeToString(src []byte) string {
	return base64.StdEncoding.EncodeToString(src)
}

func (stdEncodeDecoder) EncodedLen(n int) int {
	return base64.StdEncoding.EncodedLen(n)
}
