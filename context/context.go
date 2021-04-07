package context

import (
	"context"
	"errors"
	"time"
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
