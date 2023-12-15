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
		i, err := base36.Atoi[int](s)
		so(err, isNil)
		so(i, eq, base36.MaxID)
	})
}

func TestItoa(t *testing.T) {
	cv("Itoa", t, func() {
		i := int32(1234567890)
		s := base36.Itoa(i)
		ii, err := base36.Atoi[int](s)
		so(err, isNil)
		so(ii, eq, i)
	})
}
