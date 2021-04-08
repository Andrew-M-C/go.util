package context

import (
	"context"
	"errors"
	"testing"
	"time"
)

func Test_Cancel(t *testing.T) {
	ctx, cancel := Cancel()

	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	err := HandleContext(ctx, func() error {
		time.Sleep(time.Second)
		return nil
	})

	if !errors.Is(err, context.Canceled) {
		t.Errorf("returned error is not 'Calceled'")
		return
	}
}

func Test_Deadline(t *testing.T) {
	ctx, cancel := Deadline(time.Now().Add(100 * time.Millisecond))
	defer cancel()

	err := HandleContext(ctx, func() error {
		time.Sleep(time.Second)
		return nil
	})

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("returned error is not 'DeadlineExceeded'")
		return
	}
}

func Test_Timeout(t *testing.T) {
	ctx, cancel := Timeout(100 * time.Millisecond)
	defer cancel()

	err := HandleContext(ctx, func() error {
		time.Sleep(time.Second)
		return nil
	})

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("returned error is not 'DeadlineExceeded'")
		return
	}
}

func Test_HandleContext(t *testing.T) {
	ctx := context.Background()

	err := HandleContext(ctx, func() error {
		time.Sleep(100 * time.Millisecond)
		return nil
	})

	if err != nil {
		t.Errorf("error is not expected: %v", err)
		return
	}
}
