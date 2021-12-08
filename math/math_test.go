package math

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestMath(t *testing.T) {
	Convey("test int_sqrt", t, func() { testIntSqrt(t) })
}
