package errors

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

var (
	cv = convey.Convey
	so = convey.So

	eq = convey.ShouldEqual
	ne = convey.ShouldNotEqual
	le = convey.ShouldBeLessThanOrEqualTo
	ge = convey.ShouldBeGreaterThanOrEqualTo

	isTrue  = convey.ShouldBeTrue
	isFalse = convey.ShouldBeFalse

	isEmpty = convey.ShouldBeEmpty
	isNil   = convey.ShouldBeNil
	notNil  = convey.ShouldNotBeNil
)

func TestErrors(t *testing.T) {
	cv("测试 error_code 逻辑", t, func() { testErrorCode(t) })
	cv("测试 Unwrap", t, func() { testUnwrap(t) })
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

func testUnwrap(t *testing.T) {
	cv("没有 wrapping", func() {
		err := errors.New("some error")
		e := errors.Unwrap(err)
		so(e, isNil)
		ea, ok := Unwrap[errTypeA](err)
		so(ok, isFalse)
		so(ea.Error(), eq, "")
	})

	cv("类型本身", func() {
		var err error = errTypeA("A")

		e := errors.Unwrap(err)
		so(e, isNil)

		ea, ok := Unwrap[errTypeA](err)
		so(ok, isTrue)
		so(ea.Error(), eq, err.Error())

		eb, ok := Unwrap[errTypeB](err)
		so(ok, isFalse)
		so(eb.Error(), eq, "")
	})

	cv("一层 wrapping, 被 wrapped", func() {
		var a error = errTypeA("A")
		err := fmt.Errorf("一层: %w", a)

		e := errors.Unwrap(err)
		so(e, notNil)
		so(e.Error(), eq, a.Error())

		ea, ok := Unwrap[errTypeA](err)
		so(string(ea), ne, "")
		so(ok, isTrue)
		so(ea.Error(), eq, a.Error())

		eb, ok := Unwrap[errTypeB](err)
		so(ok, isFalse)
		so(string(eb), eq, "")
	})

	cv("两层 wrapping", func() {
		var a error = errTypeA("A")
		var errWithA error = fmt.Errorf("E -> %w", a)
		var errWithErrWithA error = fmt.Errorf("E -> %w", errWithA)

		var b error = errTypeB("B")
		var bWithErrWithErrWithA error = fmt.Errorf("%w -> %v", b, errWithErrWithA)

		e := errors.Unwrap(errWithA)
		so(e, notNil)
		so(e.Error(), eq, a.Error())

		e = errors.Unwrap(errWithErrWithA)
		so(e, notNil)
		so(e.Error(), eq, errWithA.Error())

		ea, ok := Unwrap[errTypeA](errWithErrWithA)
		so(ok, isTrue)
		so(ea.Error(), eq, a.Error())

		_, ok = Unwrap[errTypeA](bWithErrWithErrWithA)
		so(ok, isFalse)

		eb, ok := Unwrap[errTypeB](bWithErrWithErrWithA)
		so(ok, isTrue)
		so(eb.Error(), eq, b.Error())
	})
}

type errTypeA string

func (e errTypeA) Error() string {
	return string(e)
}

type errTypeB string

func (e errTypeB) Error() string {
	return string(e)
}
