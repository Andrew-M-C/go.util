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
	v, _, empty, emptyAndClosed = readNonBlocked(ch)
	return v, empty, emptyAndClosed
}

func readNonBlocked[T any](ch <-chan T) (v T, drained, empty, emptyAndClosed bool) {
	chanOpened := true

	select {
	default:
		empty = true
	case v, chanOpened = <-ch:
		drained = true
		// OK
	}

	if !chanOpened {
		return v, drained, true, true
	}
	return v, drained, empty, false
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
		emptyAndTimeout = true
		emptyAndClosed = true
	}
	return v, emptyAndTimeout, false
}

// ReadManyInTime 从一个 chan 中尽可能读出数据, 当读取数据超时，或者达到 limit 值时返回。
// 参数 limit <= 0 时表示无限制。参数 withIn <= 0 则表示将当前 channel 中已有的数据尽数读出
// (limit 参数逻辑仍在), 不考虑超时。
func ReadManyInTime[T any](
	ch <-chan T, limit int, withIn time.Duration,
) (res []T, emptyAndTimeout, emptyAndClosed bool) {
	// 不带超时, 那就尽量读
	if withIn <= 0 {
		return readManyAtOnce(ch, limit)
	}

	// 检查是否已关闭
	defer func() {
		if e := recover(); e != nil {
			emptyAndClosed = true
		}
	}()

	t := time.NewTimer(withIn)
	defer func() {
		if !t.Stop() {
			_, _, _ = ReadNonBlocked(t.C)
		}
	}()

	// 读取
	var v T
	chanOpened := true

	for {
		shouldBreak := false
		select {
		case <-t.C:
			emptyAndTimeout = true
			shouldBreak = true

		case v, chanOpened = <-ch:
			res = append(res, v)
			if !chanOpened {
				emptyAndTimeout = true
				emptyAndClosed = true
				shouldBreak = true
			} else if limit > 0 && len(res) >= limit {
				shouldBreak = true
			}
		}

		if shouldBreak {
			break
		}
	}

	if !chanOpened {
		emptyAndTimeout = true
		emptyAndClosed = true
	}
	return res, emptyAndTimeout, false
}

func readManyAtOnce[T any](ch <-chan T, limit int) (res []T, empty, emptyAndClosed bool) {
	var v T
	var drained bool
	for {
		v, drained, empty, emptyAndClosed = readNonBlocked(ch)
		if !drained {
			return
		}
		res = append(res, v)
		if limit > 0 && len(res) >= limit {
			return
		}
	}
}
