package console

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

var (
	cv = convey.Convey
)

func TestConsole(t *testing.T) {
	cv("测试 PrintTables", t, func() { testPrintTables(t) })
}
