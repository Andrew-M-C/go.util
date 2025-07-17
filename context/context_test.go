package context

import (
	"context"
	"errors"
	"sync"
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
	cv("测试 HandleContextT 函数", t, func() { testHandleContextT(t) })
	cv("测试 WithUniqID 和 UniqueID 函数", t, func() { testWithUniqueID(t) })

	// time.Sleep(time.Second)
}

func testCancel(t *testing.T) {
	ctx, cancel := Cancel()

	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	wg := sync.WaitGroup{}
	wg.Add(1)

	err := HandleContext(ctx, func() error {
		time.Sleep(time.Second)
		t.Logf("done")
		wg.Done()
		// panic("done")
		return nil
	})
	so(errors.Is(err, context.Canceled), isTrue)

	wg.Wait()
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

func testHandleContextT(t *testing.T) {
	cv("测试正常执行返回值", func() {
		ctx := context.Background()

		result, err := HandleContextT(ctx, func() (string, error) {
			time.Sleep(50 * time.Millisecond)
			return "test result", nil
		})

		so(err, isNil)
		so(result, convey.ShouldEqual, "test result")
	})

	cv("测试函数返回错误", func() {
		ctx := context.Background()
		testErr := errors.New("test error")

		result, err := HandleContextT(ctx, func() (int, error) {
			return 42, testErr
		})

		so(err, convey.ShouldEqual, testErr)
		so(result, convey.ShouldEqual, 42)
	})

	cv("测试 context 取消", func() {
		ctx, cancel := Cancel()

		go func() {
			time.Sleep(50 * time.Millisecond)
			cancel()
		}()

		result, err := HandleContextT(ctx, func() (string, error) {
			time.Sleep(200 * time.Millisecond)
			return "should not reach here", nil
		})

		so(errors.Is(err, context.Canceled), isTrue)
		so(result, convey.ShouldEqual, "")
	})

	cv("测试 context 超时", func() {
		ctx, cancel := Timeout(50 * time.Millisecond)
		defer cancel()

		result, err := HandleContextT(ctx, func() (int, error) {
			time.Sleep(200 * time.Millisecond)
			return 100, nil
		})

		so(errors.Is(err, context.DeadlineExceeded), isTrue)
		so(result, convey.ShouldEqual, 0)
	})

	cv("测试传入 nil 函数", func() {
		ctx := context.Background()

		result, err := HandleContextT[string](ctx, nil)

		so(err, convey.ShouldNotBeNil)
		so(err.Error(), convey.ShouldEqual, "missing function")
		so(result, convey.ShouldEqual, "")
	})

	cv("测试不同类型的返回值", func() {
		ctx := context.Background()

		// 测试整数类型
		intResult, err := HandleContextT(ctx, func() (int, error) {
			return 42, nil
		})
		so(err, isNil)
		so(intResult, convey.ShouldEqual, 42)

		// 测试布尔类型
		boolResult, err := HandleContextT(ctx, func() (bool, error) {
			return true, nil
		})
		so(err, isNil)
		so(boolResult, convey.ShouldEqual, true)

		// 测试结构体类型
		type testStruct struct {
			Name  string
			Value int
		}
		structResult, err := HandleContextT(ctx, func() (testStruct, error) {
			return testStruct{Name: "test", Value: 123}, nil
		})
		so(err, isNil)
		so(structResult.Name, convey.ShouldEqual, "test")
		so(structResult.Value, convey.ShouldEqual, 123)
	})

	cv("测试 deadline 场景", func() {
		deadline := time.Now().Add(100 * time.Millisecond)
		ctx, cancel := Deadline(deadline)
		defer cancel()

		result, err := HandleContextT(ctx, func() (string, error) {
			time.Sleep(200 * time.Millisecond)
			return "timeout test", nil
		})

		so(errors.Is(err, context.DeadlineExceeded), isTrue)
		so(result, convey.ShouldEqual, "")
	})
}
