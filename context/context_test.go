package context

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/smartystreets/goconvey/convey"
)

var (
	cv = convey.Convey
	so = convey.So

	isTrue  = convey.ShouldBeTrue
	isNil   = convey.ShouldBeNil
	isEmpty = convey.ShouldBeEmpty

	notEmpty = convey.ShouldNotBeEmpty
)

func TestContext(t *testing.T) {
	cv("测试 Cancel 函数", t, func() { testCancel(t) })
	cv("测试 Deadline 函数", t, func() { testDeadline(t) })
	cv("测试 Timeout 函数", t, func() { testTimeout(t) })
	cv("测试 HandleTimeout 函数", t, func() { testHandleContext(t) })
	cv("测试 WithUniqID 和 UniqueID 函数", t, func() { testWithUniqueID(t) })
}

func testCancel(t *testing.T) {
	ctx, cancel := Cancel()

	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	err := HandleContext(ctx, func() error {
		time.Sleep(time.Second)
		return nil
	})
	so(errors.Is(err, context.Canceled), isTrue)
}

func testDeadline(t *testing.T) {
	ctx, cancel := Deadline(time.Now().Add(100 * time.Millisecond))
	defer cancel()

	err := HandleContext(ctx, func() error {
		time.Sleep(time.Second)
		return nil
	})

	so(errors.Is(err, context.DeadlineExceeded), isTrue)
}

func testTimeout(t *testing.T) {
	ctx, cancel := Timeout(100 * time.Millisecond)
	defer cancel()

	err := HandleContext(ctx, func() error {
		time.Sleep(time.Second)
		return nil
	})

	so(errors.Is(err, context.DeadlineExceeded), isTrue)
}

func testHandleContext(t *testing.T) {
	ctx := context.Background()

	err := HandleContext(ctx, func() error {
		time.Sleep(100 * time.Millisecond)
		return nil
	})
	so(err, isNil)
}

func testWithUniqueID(t *testing.T) {
	ctx := context.Background()
	uid := UniqueID(ctx)
	so(uid, isEmpty)

	ctx, uid = WithUniqueID(ctx)
	so(uid, notEmpty)

	uid = UniqueID(ctx)
	so(uid, notEmpty)
}
