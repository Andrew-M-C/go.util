package recovery_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/Andrew-M-C/go.util/recovery"
	"github.com/Andrew-M-C/go.util/runtime/caller"
	"github.com/smartystreets/goconvey/convey"
)

var (
	cv = convey.Convey
	so = convey.So
	eq = convey.ShouldEqual

	notPanic = convey.ShouldNotPanic
)

func TestRecovery(t *testing.T) {
	cv("测试 recovery", t, func() { testRecovery(t) })
}

func testRecovery(t *testing.T) {
	cv("通用逻辑", func() {
		f := func() {
			defer recovery.CatchPanic()

			panic("错误")
		}

		so(f, notPanic)
	})

	cv("打日志", func() {
		f := func() {
			ctx := context.Background()

			defer recovery.CatchPanic(
				recovery.WithContext(ctx),
				recovery.WithErrorLog(),
			)

			s := []int{}
			_ = s[2]
		}

		so(f, notPanic)
	})

	cv("打日志 No 2", func() {
		f := func() (info string) {
			defer recovery.CatchPanic(
				recovery.WithCallback(func(e any, stack []caller.Caller) {
					t.Logf("Got panic position: %v", stack[0])
					info = fmt.Sprint(e)
				}),
			)

			panic("主动触发")
		}

		so(f(), eq, "主动触发")
	})
}
