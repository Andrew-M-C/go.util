package math

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

var (
	cv = convey.Convey
	so = convey.So

	eq = convey.ShouldEqual
)

func TestMath(t *testing.T) {
	cv("test math_int_sqrt", t, func() { testIntSqrt(t) })
	cv("test math_linear", t, func() { testLinear(t) })
}
