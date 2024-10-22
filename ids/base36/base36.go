// Package base36 实现混合字母和数字的数字型 ID 值, 这主要是为了便于人为设定一个可读友好的字符串,
// 同时又能转换为数字的 ID 存储。不区分大小写, 最多 10 个字符
package base36

import (
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"
)

// FullPrecisionSize uint64 可支持的完整 string 长度
// 考虑后决定不使用这个值
// const FullPrecisionSize = 12

// Float64PrecisionSize IEEE 754 双精度浮点数不会丢失的精度。推荐使用这个值
const Float64PrecisionSize = 10

// MaxID ID 最大值
const MaxID = 3656158440062975 // ZZZZZZZZZZ

// Double 表示支持的数据类型
type Double interface {
	~float64 | ~int64 | ~uint64
}

// Itoa 将一个 id 转为 string 类型
func Itoa[N Double](id N) string {
	s := strconv.FormatUint(uint64(id), 36)
	return strings.ToUpper(s)
}

// Atoi 将一个 string 转为数字值
func Atoi[N Double](s string) (N, error) {
	if len(s) > Float64PrecisionSize {
		return 0, fmt.Errorf("string length should not be more than %d", Float64PrecisionSize)
	}
	u, err := strconv.ParseUint(s, 36, 64)
	if err != nil {
		return 0, err
	}
	if u > MaxID {
		return 0, fmt.Errorf("id should not be greater than %d", MaxID)
	}
	return N(u), nil
}

// RevEndianItoa32 为了避免暴露出自增 id 的本质, 将一个 uint32 类型值颠倒大小端之后再进行
// base36 转为 string。当然缺点是数字越大, 字符串反而越短
func RevEndianItoa32(id uint32) string {
	id = reverseUint32(id)
	return strconv.FormatUint(uint64(id), 36)
}

// RevEndianAtoi32 是 RevEndianItoa32 的逆操作
func RevEndianAtoi32(id string) (uint32, error) {
	u64, err := strconv.ParseUint(id, 36, 32)
	if err != nil {
		return 0, err
	}
	u32 := uint32(u64 & 0xFFFFFFFF)
	return reverseUint32(u32), nil
}

func reverseUint32(u uint32) uint32 {
	buff := make([]byte, 4)
	binary.BigEndian.PutUint32(buff, u)
	return binary.LittleEndian.Uint32(buff)
}

// RevEndianItoa64 为了避免暴露出自增 id 的本质, 将一个 uint64 类型值颠倒大小端之后再进行
// base36 转为 string。当然缺点是数字越大, 字符串反而越短
func RevEndianItoa64(id uint64) string {
	id = reverseUint64(id)
	return strconv.FormatUint(uint64(id), 36)
}

// RevEndianAtoi64 是 RevEndianItoa64 的逆操作
func RevEndianAtoi64(id string) (uint64, error) {
	u64, err := strconv.ParseUint(id, 36, 64)
	if err != nil {
		return 0, err
	}
	return reverseUint64(u64), nil
}

func reverseUint64(u uint64) uint64 {
	buff := make([]byte, 8)
	binary.BigEndian.PutUint64(buff, u)
	return binary.LittleEndian.Uint64(buff)
}
