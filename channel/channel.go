// Package channel 提供一些方便的 chan 操作封装
package channel

import "time"

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

// WriteWithTimeout 带超时向一个 chan 中写入数据, 当时间到了的时候依然写入失败, 说明
// channel 是满的, 直接返回
func WriteWithTimeout[T any](ch chan<- T, v T, timeout time.Duration) (full, closed bool) {
	if timeout <= 0 {
		return WriteNonBlocked(ch, v)
	}

	// 检查是否已关闭
	defer func() {
		if e := recover(); e != nil {
			closed = true
		}
	}()

	// 创建 timer
	t := time.NewTimer(timeout)
	defer func() {
		if !t.Stop() {
			// ref: [使用 Golang Timer 的正确方式](http://russellluo.com/2018/09/the-correct-way-to-use-timer-in-golang.html)
			_, _, _ = ReadNonBlocked(t.C)
		}
	}()

	select {
	case <-t.C:
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

// ReadWithTimeout 从一个 chan 中读出数据。当读取超时时返回
func ReadWithTimeout[T any](ch <-chan T, timeout time.Duration) (v T, emptyAndTimeout, emptyAndClosed bool) {
	if timeout <= 0 {
		return ReadNonBlocked(ch)
	}

	// 检查是否已关闭
	defer func() {
		if e := recover(); e != nil {
			emptyAndClosed = true
		}
	}()

	t := time.NewTimer(timeout)
	defer func() {
		if !t.Stop() {
			_, _, _ = ReadNonBlocked(t.C)
		}
	}()

	// 读取
	chanOpened := true

	select {
	case <-t.C:
		emptyAndTimeout = true
	case v, chanOpened = <-ch:
		// OK
	}

	if !chanOpened {
		return v, true, true
	}
	return v, emptyAndTimeout, false
}
