package errors

import (
	"errors"
	"strings"
	"testing"

	"github.com/martinlindhe/base36"
	. "github.com/smartystreets/goconvey/convey"
)

func TestErrors(t *testing.T) {
	Convey("测试 error_code 逻辑", t, func() { testErrorCode(t) })
}

func testErrorCode(t *testing.T) {
	Convey("易认错的字符", func() {
		c, ok := decode("IiII")
		So(ok, ShouldBeTrue)
		So(c, ShouldEqual, (36*36*36)+(36*36)+(36)+1)

		c, ok = decode("ilI1")
		So(ok, ShouldBeTrue)
		So(c, ShouldEqual, (36*36*36)+(36*36)+(36)+1)

		code := ErrorToCode(nil)
		So(code, ShouldBeEmpty)
	})

	Convey("正式逻辑", func() {
		iterCount := 0
		zeroCount := 0
		values := map[string]struct{}{}

		gothrough := func(from, to rune) {
			for r := from; r <= to; r++ {
				e := errors.New(string(r))
				c := ErrorToCode(e)
				So(strings.ContainsAny(c, "Il O"), ShouldBeFalse)
				i := base36.Decode(c)
				So(i, ShouldBeLessThanOrEqualTo, 0xFFFFF)
				So(i, ShouldBeGreaterThanOrEqualTo, 0)

				if i == 0 {
					zeroCount++
				}
				values[c] = struct{}{}
				iterCount++
			}
		}

		gothrough('a', 'z')
		gothrough('A', 'Z')
		gothrough('0', '9')

		So(zeroCount, ShouldBeLessThanOrEqualTo, 2)
		So(iterCount, ShouldEqual, len(values))
		t.Logf("zero count: %d, all value count: %d, iter count: %d", zeroCount, len(values), iterCount)
	})
}
