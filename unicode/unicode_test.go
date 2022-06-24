package unicode

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

var (
	cv = convey.Convey
	so = convey.So

	eq = convey.ShouldEqual
)

func TestUnicode(t *testing.T) {
	cv("测试 east_asian_width 逻辑", t, func() { testEastAsianWidth(t) })
}
