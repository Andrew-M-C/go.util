// Package channel 提供一些方便的 chan 操作封装
package channel

// WriteNonBlocked 非阻塞地向一个 chan 中写入数据, 当写入失败时返回
func WriteNonBlocked[T any](ch chan<- T, v T) (full, closed bool) {
	defer func() {
		if e := recover(); e != nil {
			closed = true
		}
	}()

	select {
	default:
		full = true
	case ch <- v:
		// OK
	}

	return
}

// ReadNonBlocked 非阻塞地从一个 chan 中读出数据。当读取失败时返回
func ReadNonBlocked[T any](ch <-chan T) (v T, empty, emptyAndClosed bool) {
	chanOpened := true

	select {
	default:
		empty = true
	case v, chanOpened = <-ch:
		// OK
	}

	if !chanOpened {
		return v, true, true
	}
	return v, empty, false
}
