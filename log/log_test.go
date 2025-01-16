package log

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Andrew-M-C/go.util/log/dyeing"
	"github.com/Andrew-M-C/go.util/log/trace"
	"github.com/smartystreets/goconvey/convey"
)

var (
	cv = convey.Convey
	so = convey.So
	eq = convey.ShouldEqual
)

func TestLog(t *testing.T) {
	testInit(t)
	cv("调试", t, func() { testDebugging(t) })
	cv("测试 stringer 逻辑", t, func() { testStringer(t) })
	cv("测试自动删除", t, func() { testAutoRemove(t) })
	cv("测试 SetSkipCaller", t, func() { testSetSkipCaller(t) })
	cv("测试染色日志", t, func() { testDyeing(t) })

	t.Logf("等待文件写入")
	time.Sleep(4 * time.Second)
}

func testInit(t *testing.T) {
	// 打开调试信息
	internal.debugf = t.Logf
}

func testDebugging(t *testing.T) {
	SetFileName("./test.log")
	SetLevel(TraceLevel, InfoLevel)

	Trace("Hello,", "trace")
	Debug("Hello,", "debug")
	Warnf("Hello, %s!", "warning")
	Error("Hello", "error")

	l := NewLogger()
	l.Debug("Hello,", "debug", "logger")
	l.Warnf("Hello, %s logger!", "warning")
	l.Error("Hello", "error", "logger")

	ctx := context.Background()
	InfoContext(ctx, "Hello, info context")
	InfoContextf(ctx, "Hello, %s context", "infof")

	ctx = trace.EnsureTraceID(ctx)
	InfoContext(ctx, "Hello, info context")
	InfoContextf(ctx, "Hello, %s trace context", "infof")

	time.Sleep(1 * time.Second)

	// 更新文件大小并快速创建多个日志
	SetLevel(DebugLevel, NoLog)
	SetFileSize(1000) // 1MB

	logMany := func() {
		for i := 0; i < 100000; i++ {
			Warn("填充日志, 第", i+1, "条")
		}
	}
	logMany()
	time.Sleep(time.Second)

	logMany()
	time.Sleep(time.Second)

	// 尝试读取日志文件
	file, err := os.ReadFile(*internal.file.name)
	so(err, eq, nil)

	t.Logf("files size: %d", len(file))

	Infof("以下应该没有日志")
	SetLevel(NoLog, NoLog)
	Error("Hello", "no error")
}

func testAutoRemove(t *testing.T) {
	ctx := trace.WithTraceID(context.Background(), "refill_test")

	logMany := func() {
		for i := 0; i < 100000; i++ {
			WarnContext(ctx, "再次填充日志, 第", i+1, "条")
		}
	}

	SetLevel(DebugLevel, NoLog)
	start := time.Now()

	for i := 0; i < 15; i++ {
		logMany()
		time.Sleep(500 * time.Millisecond)

		files, _ := os.ReadDir(".")
		cnt := 0
		for _, f := range files {
			if strings.HasSuffix(f.Name(), ".log") {
				cnt++
				t.Log(f.Name())
			}
		}
		t.Logf("%v - 共有 %d 个日志文件", time.Since(start), cnt)
	}
}

func testSetSkipCaller(*testing.T) {
	ctx := trace.WithTraceID(context.Background(), "testSetSkipCaller")

	SetFileLevel(NoLog)
	SetConsoleLevel(DebugLevel)
	SetSkipCaller(1)

	InfoContext(ctx, "这一个日志应该是包含了 TestLog() 函数的信息")

	time.Sleep(time.Second)
}

func testDyeing(*testing.T) {
	SetSkipCaller(0)
	SetLevel(ErrorLevel, NoLog)            // 暂时关闭日志
	SetDyeingLevel(DebugLevel, DebugLevel) // 文件和日志都给调试级别的染色日志

	ctx := context.Background()
	ctx = trace.WithTraceID(ctx, "dyeing-test")
	ErrorContext(ctx, "这句日志不应该出现在命令行")

	ctx = dyeing.WithDyeing(ctx, true)
	DebugContext(ctx, "这句日志因为染色了, 应该出现在文件和命令行")

	ctx = dyeing.WithDyeing(ctx, false)
	ErrorContext(ctx, "这句日志取消染色了, 不应该出现在命令行")
}

func testStringer(*testing.T) {
	cv("JSON", func() {
		type testType struct {
			String string `json:"string"`
		}
		data := testType{
			String: `"logger"`,
		}
		s := fmt.Sprint(ToJSON(data))
		so(s, eq, `{"string":"\"logger\""}`)
	})

	cv("hex", func() {
		b := []byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef}
		s := fmt.Sprint(ToHex(b))
		so(s, eq, "0123456789ABCDEF")
	})
}
