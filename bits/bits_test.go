package bits_test

import (
	"testing"

	"github.com/Andrew-M-C/go.util/bits"
	"github.com/smartystreets/goconvey/convey"
)

var (
	cv = convey.Convey
	so = convey.So
	eq = convey.ShouldEqual
)

func TestBits(t *testing.T) {
	cv("Set 和 Clear 函数", t, func() { testSetClear(t) })
	cv("HasAny 和 HasAll 函数", t, func() { testHasXxx(t) })
}

func testSetClear(t *testing.T) {
	b := uint64(0)

	b = bits.Set(b, 1)
	so(b, eq, 0x02)

	b = bits.Set(b, 0, 1)
	so(b, eq, 0x03)

	b = bits.Set(b, 0, 63)
	so(b, eq, uint64(0x8000000000000003))

	b = bits.Clear(b, 62)
	so(b, eq, uint64(0x8000000000000003))

	b = bits.Clear(b, 1, 63)
	so(b, eq, 0x01)

	b = bits.Clear(b, 0)
	so(b, eq, 0x0)
}

func testHasXxx(t *testing.T) {
	b := bits.New64(0, 1, 32, 63)

	so(bits.HasAll(b, 0, 1, 32, 63), eq, true)
	so(bits.HasAll(b, 0, 63), eq, true)
	so(bits.HasAll(b, 2, 63), eq, false)

	so(bits.HasAny(b, 0, 1, 32, 63), eq, true)
	so(bits.HasAny(b, 0, 63), eq, true)
	so(bits.HasAny(b, 2, 63), eq, true)
}
