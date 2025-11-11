// Package trace 基于 context 封装一个 trace ID 功能
package trace

import (
	"context"

	objectid "github.com/Andrew-M-C/go.objectid"
	"golang.org/x/exp/slices"
)

// TraceIDContextKey 返回 trace ID 的 context key, 用于一些外部需要复制 context 的
// key / value 的场景
func TraceIDContextKey() any {
	return traceIDKey{}
}

// 保存 ctx 中的 trace ID 字段
type traceIDKey struct{}

type traceIDStackyValue []string

func (t traceIDStackyValue) id() string {
	if len(t) == 0 {
		return ""
	}
	return t[len(t)-1]
}

// WithTraceID 更新 trace ID
func WithTraceID(ctx context.Context, traceID string) context.Context {
	if traceID == "" {
		return ctx
	}
	stack := traceIDStack(ctx)
	if stack.id() == traceID {
		return ctx
	}
	stack = slices.Clone(stack)
	stack = append(stack, traceID)
	return context.WithValue(ctx, traceIDKey{}, stack)
}

// WithTraceIDStack 完全替换整个 trace ID 栈
func WithTraceIDStack(ctx context.Context, traceIDStack []string) context.Context {
	if len(traceIDStack) == 0 {
		return ctx
	}
	stack := traceIDStackyValue(slices.Clone(traceIDStack))
	return context.WithValue(ctx, traceIDKey{}, stack)
}

func traceIDStack(ctx context.Context) traceIDStackyValue {
	v := ctx.Value(traceIDKey{})
	if v == nil {
		return traceIDStackyValue{}
	}
	st, _ := v.(traceIDStackyValue)
	return st
}

// TraceID 从 context 中读取 trace ID
func TraceID(ctx context.Context) string {
	v := ctx.Value(traceIDKey{})
	if v == nil {
		return ""
	}
	s, _ := v.(traceIDStackyValue)
	return s.id()
}

// TraceIDStack 从 context 中读取历史 trace ID 栈
func TraceIDStack(ctx context.Context) []string {
	v := ctx.Value(traceIDKey{})
	if v == nil {
		return nil
	}
	s, _ := v.(traceIDStackyValue)
	return slices.Clone(s)
}

// EnsureTraceID 确保 context 中有一个 trace ID
func EnsureTraceID(ctx context.Context) context.Context {
	traceID := TraceID(ctx)
	if traceID == "" {
		return WithTraceID(ctx, generateTraceID())
	}
	return ctx
}

func generateTraceID() string {
	return objectid.New16().String()
}
