package dyeing_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/Andrew-M-C/go.util/log/dyeing"
	"github.com/smartystreets/goconvey/convey"
)

var (
	cv = convey.Convey
	so = convey.So
	eq = convey.ShouldEqual
)

func TestDyeing(t *testing.T) {
	cv("WithDyeing 函数", t, func() { testWithDyeing(t) })
}

func testWithDyeing(t *testing.T) {
	cv("基本逻辑", func() {
		ctx := context.Background()
		so(dyeing.Dyeing(ctx), eq, false)

		ctx = dyeing.WithDyeing(ctx, false)
		so(dyeing.Dyeing(ctx), eq, false)

		ctx = dyeing.WithDyeing(ctx, true)
		so(dyeing.Dyeing(ctx), eq, true)

		ctx = dyeing.WithDyeing(ctx, false)
		so(dyeing.Dyeing(ctx), eq, false)
	})

	cv("dyeing 状态相同时不新建 context", func() {
		ctx := context.Background()
		ctx1 := dyeing.WithDyeing(ctx, false)
		so(dyeing.Dyeing(ctx), eq, false)
		so(dyeing.Dyeing(ctx1), eq, false)

		addr := fmt.Sprintf("%p", ctx)
		addr1 := fmt.Sprintf("%p", ctx1)
		so(addr, eq, addr1)
	})
}
