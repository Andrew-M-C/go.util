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
	cv("test int_sqrt", t, func() { testIntSqrt(t) })
}
