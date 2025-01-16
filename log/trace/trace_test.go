package trace_test

import (
	"context"
	"testing"

	"github.com/Andrew-M-C/go.util/log/trace"
	"github.com/smartystreets/goconvey/convey"
)

// go test -v -failfast -cover -coverprofile cover.out && go tool cover -html cover.out -o ~/Desktop/cover.html

var (
	cv = convey.Convey
	so = convey.So
	eq = convey.ShouldEqual
	ne = convey.ShouldNotEqual
)

func TestTrace(t *testing.T) {
	cv("trace.go", t, func() { testTraceGo(t) })
}

func testTraceGo(*testing.T) {
	cv("基本逻辑", func() {
		ctx := context.Background()
		so(trace.TraceID(ctx), eq, "")

		traceID := "test_trace_id"
		ctx = trace.WithTraceID(context.Background(), traceID)
		so(trace.TraceID(ctx), eq, traceID)

		ctx = trace.EnsureTraceID(ctx)
		so(trace.TraceID(ctx), eq, traceID)
	})

	cv("EnsureTraceID", func() {
		ctx := trace.EnsureTraceID(context.Background())
		tid := trace.TraceID(ctx)

		ctx = trace.EnsureTraceID(ctx)
		so(trace.TraceID(ctx), eq, tid)

		ctx = trace.WithTraceID(ctx, "12345")
		so(trace.TraceID(ctx), ne, tid)
		so(trace.TraceID(ctx), eq, "12345")
	})
}
