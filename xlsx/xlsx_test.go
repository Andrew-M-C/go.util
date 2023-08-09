package xlsx_test

import (
	"testing"

	"github.com/Andrew-M-C/go.util/xlsx"
	"github.com/smartystreets/goconvey/convey"
)

var (
	cv = convey.Convey
	so = convey.So
	eq = convey.ShouldEqual
)

func TestXlsx(t *testing.T) {
	cv("测试 col 名称转换", t, func() { testCellName(t) })
}

func testCellName(t *testing.T) {
	n := xlsx.CellName(0, 0)
	so(n, eq, "A1")

	n = xlsx.CellName(3, 1)
	so(n, eq, "B4")

	n = xlsx.CellName(0, 26)
	so(n, eq, "AA1")

	n = xlsx.CellName(0, 52)
	so(n, eq, "BA1")
}
