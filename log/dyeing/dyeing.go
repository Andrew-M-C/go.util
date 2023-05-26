// Package dyeing 在 context 中设置一个染色标记
package dyeing

import "context"

// Dyeing 返回一个 context 是否标记了染色
func Dyeing(ctx context.Context) bool {
	v := ctx.Value(key{})
	return v != nil
}

// WithDyeing 添加染色标记
func WithDyeing(ctx context.Context, b bool) context.Context {
	if Dyeing(ctx) == b {
		return ctx // 染色状态与原来相同, 不需要新建 context
	}
	if b {
		return context.WithValue(ctx, key{}, struct{}{})
	}
	return context.WithValue(ctx, key{}, nil)
}

type key struct{}
