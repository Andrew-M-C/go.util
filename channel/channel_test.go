package channel_test

import (
	"testing"

	"github.com/Andrew-M-C/go.util/channel"
	"github.com/smartystreets/goconvey/convey"
)

var (
	cv = convey.Convey
	so = convey.So
	eq = convey.ShouldEqual

	isTrue  = convey.ShouldBeTrue
	isFalse = convey.ShouldBeFalse
)

func TestChannel(t *testing.T) {
	cv("测试 WriteNonBlocked 和 ReadNonBlocked", t, func() { testWriteNonBlockedReadNonBlocked(t) })
}

func testWriteNonBlockedReadNonBlocked(t *testing.T) {
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
