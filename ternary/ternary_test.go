package ternary_test

import (
	"testing"

	"github.com/Andrew-M-C/go.util/ternary"
	"github.com/smartystreets/goconvey/convey"
)

var (
	cv = convey.Convey
	so = convey.So
	eq = convey.ShouldEqual
)

func TestTernary(t *testing.T) {
	cv("check()", t, func() { testCheck(t) })
}

func testCheck(t *testing.T) {
	a, b := 1, 2
	res := ternary.Check(a >= b, a, b)
	so(res, eq, b)

	res = ternary.Check(a <= b, a, b)
	so(res, eq, a)
}
