package errors

import (
	"crypto/md5"
	"encoding/binary"
	"strconv"
	"strings"
)

var (
	replace = strings.NewReplacer(
		" ", "0",
		"O", "0",
		"I", "1",
	)

	hashFunc = hash
)

// ErrorToCode 用 error 生成一个字符串 code
func ErrorToCode(err error) (code string) {
	if err == nil {
		return ""
	}

	_, code = genCode(err.Error())
	return code
}

// HashFunc 表示用于生成 error code 的哈希函数，返回的数值不得大于 0xFFFFF
type HashFunc func(string) uint64

// SetHashFunc 设置哈希函数
func SetHashFunc(f HashFunc) {
	if f != nil {
		hashFunc = f
	}
}

func genCode(s string) (uint64, string) {
	u64 := hashFunc(s) & 0xFFFFF
	codeStr := encode(u64)
	u64, _ = decode(codeStr)
	return u64, codeStr
}

func hash(s string) uint64 {
	h := md5.Sum([]byte(s))
	u := binary.BigEndian.Uint32(h[0:16])
	return uint64(u)
}

func encode(code uint64) string {
	s := strconv.FormatUint(code, 36)
	s = strings.ToUpper(s)
	return replace.Replace(s)
}

func decode(s string) (uint64, bool) {
	if len(s) != 4 {
		return 0, false
	}
	s = strings.Replace(s, "l", "1", -1)
	s = strings.ToUpper(s)
	s = replace.Replace(s)
	code, _ := strconv.ParseUint(s, 36, 64)
	return code, code > 0
}
