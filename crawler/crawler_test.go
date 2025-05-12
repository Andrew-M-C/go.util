package crawler_test

import (
	"os"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

var (
	cv = convey.Convey
	so = convey.So
	gt = convey.ShouldBeGreaterThan

	isNil = convey.ShouldBeNil
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
