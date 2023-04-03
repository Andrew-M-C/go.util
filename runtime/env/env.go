// Package env 提供一些环境变量相关的工具
package env

import (
	"os"
	"strconv"

	"golang.org/x/exp/constraints"
)

// GetString 获取 string 类型环境变量, 如果环境变量不存在则返回默认值
func GetString(key string, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}

// GetInt 获取 int 类型环境变量, 如果环境变量不存在或非法则返回默认值
func GetInt[T constraints.Integer](key string, fallback T) T {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	i, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return fallback
	}
	return T(i)
}

// GetUint 获取 uint 类型环境变量, 如果环境变量不存在或非法则返回默认值
func GetUint[T constraints.Unsigned](key string, fallback T) T {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	i, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		return fallback
	}
	return T(i)
}

// GetBool 获取 bool 类型环境变量。大于零的值均会被视为 true, 字符串 true 和 TRUE 也会视为 true。
// 剩余情况 (包括不存在) 则返回 false
func GetBool(key string) bool {
	v := os.Getenv(key)
	if v == "" {
		return false
	}

	switch v {
	case "TRUE", "true", "1":
		return true
	default:
		i, _ := strconv.ParseInt(v, 10, 64)
		return i > 0
	}
}
