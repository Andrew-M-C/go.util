// Package trace 基于 context 封装一个 trace ID 功能
package trace

import (
	"context"

	"github.com/google/uuid"
)

type traceIDKey struct{}

// GetTraceID 读取 trace ID
func GetTraceID(ctx context.Context) string {
	if v := ctx.Value(traceIDKey{}); v != nil {
		id, _ := v.(string)
		return id
	}
	return ""
}

// EnsureTraceID 如果没有 trace id 的话, 那就填充一个并返回
func EnsureTraceID(ctx context.Context) context.Context {
	if id := GetTraceID(ctx); id != "" {
		return ctx
	}
	return SetTraceID(ctx)
}

// SetTraceID 设置一个 trace ID, 如果之前已经设置过, 则会覆盖
func SetTraceID(ctx context.Context, traceID ...string) context.Context {
	id := ""
	if len(traceID) > 0 && traceID[0] != "" {
		id = traceID[0]
	} else {
		id = uuid.NewString()
	}
	return context.WithValue(ctx, traceIDKey{}, id)
}

// WithTraceID 等价于 SetTraceID
func WithTraceID(ctx context.Context, traceID ...string) context.Context {
	return SetTraceID(ctx, traceID...)
}
