package ring_test

import (
	"fmt"
	"testing"

	"github.com/Andrew-M-C/go.util/runtime/cpu/internal/ring"
	"github.com/smartystreets/goconvey/convey"
)

var (
	cv = convey.Convey
	so = convey.So
	eq = convey.ShouldEqual
)

func TestRing(t *testing.T) {
	cv("测试 *RingQueue 类型", t, func() { testRingQueue(t) })
}

func testRingQueue(t *testing.T) {
	q := ring.NewRingQueue[int](5)
	so(q.Len(), eq, 0)

	q.Push(1)
	q.Push(2)
	q.Push(3)
	so(q.Len(), eq, 3)
	so(fmt.Sprint(q.GetAllValues()), eq, "[3 2 1]")

	q.Push(4)
	q.Push(5)
	so(q.Len(), eq, 5)
	so(fmt.Sprint(q.GetAllValues()), eq, "[5 4 3 2 1]")

	q.Push(6)
	so(q.Len(), eq, 5)
	so(fmt.Sprint(q.GetAllValues()), eq, "[6 5 4 3 2]")

	q.Push(7)
	q.Push(8)
	so(q.Len(), eq, 5)
	so(fmt.Sprint(q.GetAllValues()), eq, "[8 7 6 5 4]")

	q.Push(9)
	q.Push(10)
	q.Push(11)
	q.Push(12)
	so(q.Len(), eq, 5)
	so(fmt.Sprint(q.GetAllValues()), eq, "[12 11 10 9 8]")

	q.Clear()
	so(q.Len(), eq, 0)
	so(len(q.GetAllValues()), eq, 0)
}
