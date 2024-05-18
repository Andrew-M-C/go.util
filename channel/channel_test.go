package channel_test

import (
	"testing"
	"time"

	"github.com/Andrew-M-C/go.util/channel"
	"github.com/smartystreets/goconvey/convey"
)

var (
	cv = convey.Convey
	so = convey.So
	eq = convey.ShouldEqual
	lt = convey.ShouldBeLessThan
	gt = convey.ShouldBeGreaterThan

	isTrue  = convey.ShouldBeTrue
	isFalse = convey.ShouldBeFalse
)

func TestChannel(t *testing.T) {
	cv("测试 WriteNonBlocked 和 ReadNonBlocked", t, func() { testWriteNonBlockedReadNonBlocked(t) })
	cv("测试 WriteWithTimeout 和 ReadWithTimeout", t, func() { testWriteWithTimeoutReadWithTimeout(t) })
	cv("测试 ReadManyInTime", t, func() { testReadManyInTime(t) })
}

func testWriteNonBlockedReadNonBlocked(*testing.T) {
	cv("没有缓冲区的 chan", func() {
		ch := make(chan struct{})

		full, closed := channel.WriteNonBlocked(ch, struct{}{})
		so(full, isTrue)
		so(closed, isFalse)

		_, empty, closed := channel.ReadNonBlocked(ch)
		so(empty, isTrue)
		so(closed, isFalse)

		close(ch)
		full, closed = channel.WriteNonBlocked(ch, struct{}{})
		so(full, isFalse)
		so(closed, isTrue)

		_, empty, closed = channel.ReadNonBlocked(ch)
		so(empty, isTrue)
		so(closed, isTrue)
	})

	cv("有缓冲区的 chan", func() {
		ch := make(chan int, 2)

		v, empty, closed := channel.ReadNonBlocked(ch)
		so(v, eq, 0)
		so(empty, eq, true)
		so(closed, eq, false)

		full, closed := channel.WriteNonBlocked(ch, 10)
		so(full, eq, false)
		so(closed, eq, false)

		v, empty, closed = channel.ReadNonBlocked(ch)
		so(v, eq, 10)
		so(empty, eq, false)
		so(closed, eq, false)

		full, closed = channel.WriteNonBlocked(ch, 20)
		so(full, eq, false)
		so(closed, eq, false)

		full, closed = channel.WriteNonBlocked(ch, 30)
		so(full, eq, false)
		so(closed, eq, false)

		full, closed = channel.WriteNonBlocked(ch, 40)
		so(full, eq, true)
		so(closed, eq, false)

		v, empty, closed = channel.ReadNonBlocked(ch)
		so(v, eq, 20)
		so(empty, eq, false)
		so(closed, eq, false)

		v, empty, closed = channel.ReadNonBlocked(ch)
		so(v, eq, 30)
		so(empty, eq, false)
		so(closed, eq, false)

		v, empty, closed = channel.ReadNonBlocked(ch)
		so(v, eq, 0)
		so(empty, eq, true)
		so(closed, eq, false)

		full, closed = channel.WriteNonBlocked(ch, 100)
		so(full, eq, false)
		so(closed, eq, false)

		close(ch)
		full, closed = channel.WriteNonBlocked(ch, 200)
		so(full, eq, false)
		so(closed, eq, true)

		v, empty, closed = channel.ReadNonBlocked(ch)
		so(v, eq, 100)
		so(empty, eq, false)
		so(closed, eq, false)

		v, empty, closed = channel.ReadNonBlocked(ch)
		so(v, eq, 0)
		so(empty, eq, true)
		so(closed, eq, true)
	})
}

func testWriteWithTimeoutReadWithTimeout(*testing.T) {
	cv("没有缓冲区的 chan", func() {
		ch := make(chan int)
		res := int(0)

		readInMilli(ch, 100, &res)
		full, closed := channel.WriteWithTimeout(ch, 1, ms(150))
		so(full, eq, false)
		so(closed, eq, false)
		so(res, eq, 1)

		readInMilli(ch, 100, &res)
		full, closed = channel.WriteWithTimeout(ch, 2, ms(50))
		so(full, eq, true)
		so(closed, eq, false)

		// 让上一个 readInMilli 消费完
		ch <- 3
		time.Sleep(ms(100))
		so(res, eq, 3)

		// 测试读超时的情况
		start := time.Now()
		writeInMilli(ch, 4, 100)
		res = <-ch
		so(time.Since(start), gt, ms(100))
		so(res, eq, 4)

		start = time.Now()
		writeInMilli(ch, 5, 50)
		writeInMilli(ch, 6, 60)
		writeInMilli(ch, 7, 70)
		res = <-ch
		so(res, eq, 5)
		so(time.Since(start), gt, ms(50))
		res = <-ch
		so(res, eq, 6)
		so(time.Since(start), gt, ms(60))
		res = <-ch
		so(res, eq, 7)
		so(time.Since(start), gt, ms(70))

		// 关闭
		close(ch)
		start = time.Now()
		full, closed = channel.WriteWithTimeout(ch, 2, ms(1000))
		so(time.Since(start), lt, ms(100))
		so(full, eq, false)
		so(closed, eq, true)
	})
}

func ms(msec int) time.Duration {
	return time.Duration(msec) * time.Millisecond
}

func writeInMilli[T any](ch chan T, v T, msec int) {
	go func() {
		time.Sleep(time.Duration(msec) * time.Millisecond)
		ch <- v
	}()
}

func readInMilli[T any](ch chan T, msec int, result ...*T) {
	go func() {
		time.Sleep(time.Duration(msec) * time.Millisecond)
		v := <-ch
		if len(result) > 0 && result[0] != nil {
			*result[0] = v
		}
	}()
}

func testReadManyInTime(t *testing.T) {
	cv("带缓冲区的 chan", func() {
		ch := make(chan int, 100)

		repeat(50, func(i int) {
			ch <- i + 1
		})

		t.Log("写入 50, 读取 50")
		start := time.Now()
		res, emptyAndTimeout, emptyAndClosed := channel.ReadManyInTime(ch, -1, ms(50))
		so(time.Since(start), gt, ms(50))
		so(time.Since(start), lt, ms(100))
		so(emptyAndTimeout, eq, true)
		so(emptyAndClosed, eq, false)
		so(len(res), eq, 50)
		repeat(50, func(i int) {
			so(res[i], eq, i+1)
		})

		t.Log("写入 50, 读取 10")
		repeat(50, func(i int) {
			ch <- i * 10
		})
		res, emptyAndTimeout, emptyAndClosed = channel.ReadManyInTime(ch, 10, 0)
		so(emptyAndTimeout, eq, false)
		so(emptyAndClosed, eq, false)
		so(len(res), eq, 10)
		repeat(len(res), func(i int) {
			so(res[i], eq, i*10)
		})

		t.Log("读取 40 (全部)")
		res, emptyAndTimeout, emptyAndClosed = channel.ReadManyInTime(ch, 50, 0)
		so(emptyAndTimeout, eq, true)
		so(emptyAndClosed, eq, false)
		so(len(res), eq, 40)
		repeat(len(res), func(i int) {
			so(res[i], eq, (i+10)*10)
		})
	})
}

func repeat(count int, fu func(i int)) {
	for i := 0; i < count; i++ {
		fu(i)
	}
}
