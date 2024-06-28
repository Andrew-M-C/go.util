package localcache

import (
	"os"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

var (
	cv = convey.Convey
	so = convey.So
	eq = convey.ShouldEqual

	isErr = convey.ShouldBeError
	isNil = convey.ShouldBeNil
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
