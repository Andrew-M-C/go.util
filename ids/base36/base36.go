// Package base36
package base36

import (
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/exp/constraints"
)

// FullPrecisionSize uint64 可支持的完整 string 长度
// 考虑后决定不使用这个值
// const FullPrecisionSize = 12

// Float64PrecisionSize IEEE 754 双精度浮点数不会丢失的精度。推荐使用这个值
const Float64PrecisionSize = 10

// MaxID ID 最大值
const MaxID = 3656158440062975 // ZZZZZZZZZZ

// Itoa 将一个 id 转为 string 类型
func Itoa[N constraints.Float | constraints.Integer](id N) string {
	s := strconv.FormatUint(uint64(id), 36)
	return strings.ToUpper(s)
}

// Atoi 将一个 string 转为数字值
func Atoi[N constraints.Float | constraints.Integer](s string) (N, error) {
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
