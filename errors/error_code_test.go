package errors

import (
	"errors"
	"strconv"
	"strings"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

var (
	cv = convey.Convey
	so = convey.So

	eq = convey.ShouldEqual
	le = convey.ShouldBeLessThanOrEqualTo
	ge = convey.ShouldBeGreaterThanOrEqualTo

	isTrue  = convey.ShouldBeTrue
	isFalse = convey.ShouldBeFalse

	isEmpty = convey.ShouldBeEmpty
	isNil   = convey.ShouldBeNil
)

func TestErrors(t *testing.T) {
	cv("测试 error_code 逻辑", t, func() { testErrorCode(t) })
}

func testErrorCode(t *testing.T) {
	cv("易认错的字符", func() {
		c, ok := decode("IiII")
		so(ok, isTrue)
		so(c, eq, (36*36*36)+(36*36)+(36)+1)

		c, ok = decode("ilI1")
		so(ok, isTrue)
		so(c, eq, (36*36*36)+(36*36)+(36)+1)

		code := ErrorToCode(nil)
		so(code, isEmpty)
	})

	cv("正式逻辑", func() {
		iterCount := 0
		zeroCount := 0
		values := map[string]struct{}{}

		goThrough := func(from, to rune) {
			for r := from; r <= to; r++ {
				e := errors.New(string(r))
				c := ErrorToCode(e)
				// t.Logf("code '%s', error '%v'", c, e)
				so(strings.ContainsAny(c, "Il O"), isFalse)
				i, err := strconv.ParseUint(c, 36, 64)
				so(err, isNil)
				so(i, le, 0xFFFFF)
				so(i, ge, 0)

				if i == 0 {
					zeroCount++
				}
				values[c] = struct{}{}
				iterCount++
			}
		}

		goThrough('a', 'z')
		goThrough('A', 'Z')
		goThrough('0', '9')

		so(zeroCount, le, 2)
		so(iterCount, eq, len(values))
		t.Logf("zero count: %d, all value count: %d, iter count: %d", zeroCount, len(values), iterCount)
	})
}
