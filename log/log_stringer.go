package log

import (
	"encoding/json"
	"fmt"
	"unsafe"
)

// ToJSON 输出一个 stringer 用于在 log 中打印 JSON 数据
func ToJSON(v any) fmt.Stringer {
	return jsonWrapper{
		v: v,
	}
}

type jsonWrapper struct {
	v any
}

func (j jsonWrapper) String() string {
	b, _ := json.Marshal(j.v)
	return bToS(b)
}

func bToS(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// ToHex 返回一个 stringer 用来在 log 中打印字节流的十六进制值
func ToHex(b []byte) fmt.Stringer {
	return hexWrapper{
		b: b,
	}
}

type hexWrapper struct {
	b []byte
}

func (h hexWrapper) String() string {
	return fmt.Sprintf("%X", h.b)
}
