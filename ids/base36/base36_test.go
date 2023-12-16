package base36_test

import (
	"testing"

	"github.com/Andrew-M-C/go.util/ids/base36"
	"github.com/smartystreets/goconvey/convey"
)

var (
	cv = convey.Convey
	so = convey.So
	eq = convey.ShouldEqual

	isNil = convey.ShouldBeNil
)

func TestAtoi(t *testing.T) {
	cv("Atoi", t, func() {
		s := "zzzzzzzzzz"
		i, err := base36.Atoi[int64](s)
		so(err, isNil)
		so(i, eq, base36.MaxID)
	})

	cv("浮点数精度支持", t, func() {
		s := "ZZZZZZZZZZ"
		f, err := base36.Atoi[float64](s)
		so(err, isNil)
		so(f, eq, base36.MaxID)
		so(uint64(f), eq, base36.MaxID)
		so(int64(f), eq, base36.MaxID)
	})
}

func TestItoa(t *testing.T) {
	cv("Itoa", t, func() {
		i := int64(1234567890)
		s := base36.Itoa(i)
		ii, err := base36.Atoi[int64](s)
		so(err, isNil)
		so(ii, eq, i)
	})
}
