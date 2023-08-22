package errors

import "errors"

// Unwrap 按类型将 wrapped 的错误类型提取出来。如果提取成功, 则返回对应的类型, 并且 ok == true
func Unwrap[T any](err error) (res T, ok bool) {
	if err == nil {
		return
	}

	res, ok = err.(T)
	if ok {
		return res, true
	}

	unwrapped := errors.Unwrap(err)
	if unwrapped == nil {
		// 没有 wrapped 的错误类型
		return
	}

	res, ok = unwrapped.(T)
	if !ok {
		return Unwrap[T](unwrapped)
	}
	return res, true
}
