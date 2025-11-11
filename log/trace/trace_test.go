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
	cv("TraceIDContextKey", t, func() { testTraceIDContextKey(t) })
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

func testTraceIDContextKey(*testing.T) {
	cv("返回一致的 key", func() {
		// TraceIDContextKey() 应该每次返回相同的 key
		key1 := trace.TraceIDContextKey()
		key2 := trace.TraceIDContextKey()
		so(key1, eq, key2)
	})

	cv("可以用于 context 操作", func() {
		// 使用 TraceIDContextKey 获取的 key 应该能够正确访问 trace ID
		ctx := context.Background()
		traceID := "test_trace_id_123"
		ctx = trace.WithTraceID(ctx, traceID)

		// 通过 TraceIDContextKey 获取存储的值
		key := trace.TraceIDContextKey()
		value := ctx.Value(key)
		so(value, ne, nil)
	})

	cv("用于复制 context key/value", func() {
		// 模拟外部需要复制 context 的场景
		ctx := context.Background()
		traceID := "original_trace_id"
		ctx = trace.WithTraceID(ctx, traceID)

		// 获取 key 和 value
		key := trace.TraceIDContextKey()
		value := ctx.Value(key)

		// 创建新的 context 并复制这个 key/value
		newCtx := context.Background()
		newCtx = context.WithValue(newCtx, key, value)

		// 验证新 context 中也能读取到相同的 trace ID
		so(trace.TraceID(newCtx), eq, traceID)
	})

	cv("与 TraceIDStack 配合使用", func() {
		// 测试多层 trace ID 栈的场景
		ctx := context.Background()
		ctx = trace.WithTraceID(ctx, "trace_1")
		ctx = trace.WithTraceID(ctx, "trace_2")
		ctx = trace.WithTraceID(ctx, "trace_3")

		// 获取当前的栈
		stack := trace.TraceIDStack(ctx)
		so(len(stack), eq, 3)
		so(stack[0], eq, "trace_1")
		so(stack[1], eq, "trace_2")
		so(stack[2], eq, "trace_3")

		// 通过 TraceIDContextKey 复制到新 context
		key := trace.TraceIDContextKey()
		value := ctx.Value(key)
		newCtx := context.WithValue(context.Background(), key, value)

		// 验证栈也被正确复制
		newStack := trace.TraceIDStack(newCtx)
		so(len(newStack), eq, 3)
		so(newStack[0], eq, "trace_1")
		so(newStack[1], eq, "trace_2")
		so(newStack[2], eq, "trace_3")
		so(trace.TraceID(newCtx), eq, "trace_3")
	})
}
