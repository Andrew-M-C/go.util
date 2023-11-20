package snowflake_test

import (
	"testing"

	"github.com/Andrew-M-C/go.util/ids/snowflake"
	"github.com/smartystreets/goconvey/convey"
)

var (
	cv = convey.Convey
)

func TestSnowflake(t *testing.T) {
	cv("测试 New 函数", t, func() { testNew(t) })
}

func testNew(t *testing.T) {
	source := snowflake.Source()
	t.Log("Source:", source)
	id := snowflake.New()
	t.Log("ID:", id)
}
