package log

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/Andrew-M-C/go.util/log/trace"
	"github.com/smartystreets/goconvey/convey"
)

var (
	cv = convey.Convey
	so = convey.So
	eq = convey.ShouldEqual
)

func TestLog(t *testing.T) {
	cv("调试", t, func() { testDebugging(t) })
}

func testDebugging(t *testing.T) {
	SetFileName("./test.log")
	SetLevel(DebugLevel, InfoLevel)

	Debug("Hello,", "debug")
	Warnf("Hello, %s!", "warning")
	Error("Hello", "error")

	ctx := context.Background()
	InfoContextf(ctx, "Hello, %s context", "info")

	ctx = trace.EnsureTraceID(ctx)
	InfoContextf(ctx, "Hello, %s trace context", "info")

	time.Sleep(1 * time.Second)

	// 更新文件大小并快速创建多个日志
	SetLevel(DebugLevel, NoLog)
	SetFileSize(10 * 1000)

	logMany := func() {
		for i := 0; i < 100000; i++ {
			Warnf("填充日志, 第 %d 条", i+1)
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
