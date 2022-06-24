package context

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

// Cancel is equavalent to context.WithCancel(context.Background())
func Cancel() (context.Context, context.CancelFunc) {
	return context.WithCancel(context.Background())
}

// Deadline is equavalent to context.WithDeadline(context.Background(), d)
func Deadline(d time.Time) (context.Context, context.CancelFunc) {
	return context.WithDeadline(context.Background(), d)
}

// Timeout is equavalent to context.WithTimeout(context.Background(), timeout)
func Timeout(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout)
}

// HandleContext embeds a function without context, but can easily implement cancel and timeout logic without complex
// detecting in codes.
func HandleContext(ctx context.Context, f func() error) error {
	if f == nil {
		return errors.New("missing function")
	}

	ch := make(chan error)

	go func() {
		err := f()
		ch <- err
		close(ch)
	}()

	select {
	case err := <-ch:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

// uniqIDKeyType 定义 context 中的 uniq ID key
type uniqIDKeyType string

const (
	uniqIDKey = uniqIDKeyType("uid")
)

// WithUniqueID 返回一个新的、内置了 uniqID 的 context, 便于新建 context 的时候区分
// goroutine。如果不指定 uniqID 或者指定了空字符串, 则使用 Google 的 uuid算法生成 uid。
//
// 返回新的 context 以及 unique ID。
func WithUniqueID(parent context.Context, uniqID ...string) (context.Context, string) {
	uid := ""
	if len(uniqID) > 0 {
		uid = uniqID[0]
	}
	if uid == "" {
		uid = uuid.New().String()
	}
	return context.WithValue(parent, uniqIDKey, uid), uid
}

// UniqueID 返回保存在 context 中的 unique ID。如果不存在则返回空
func UniqueID(ctx context.Context) (uniqID string) {
	v := ctx.Value(uniqIDKey)
	if v == nil {
		return ""
	}
	id, _ := v.(string)
	return id
}
