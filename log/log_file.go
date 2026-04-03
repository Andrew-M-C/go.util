package log

import (
	"context"
	"fmt"
	"strings"

	"github.com/Andrew-M-C/go.util/log/trace"
	"go.uber.org/zap"
)

type fileLog Level

func (l fileLog) logf(f string, a ...any) {
	atomicFileLogger.Load().Log(toZapLevel(Level(l)), fmt.Sprintf(f, a...))
}

func (l fileLog) log(a ...any) {
	atomicFileLogger.Load().Log(toZapLevel(Level(l)), fmt.Sprint(a...))
}

func (l fileLog) logCtxf(ctx context.Context, f string, a ...any) {
	logger := atomicFileLogger.Load()
	if id := trace.TraceID(ctx); id != "" {
		logger = logger.With(zap.String("trace_id", id))
	}
	logger.Log(toZapLevel(Level(l)), fmt.Sprintf(f, a...))
}

func (l fileLog) logCtx(ctx context.Context, a ...any) {
	logger := atomicFileLogger.Load()
	if id := trace.TraceID(ctx); id != "" {
		logger = logger.With(zap.String("trace_id", id))
	}
	logger.Log(toZapLevel(Level(l)), fmt.Sprint(a...))
}

// SetFileName 设置日志文件名
func SetFileName(name string) {
	if name == "" {
		return
	}
	name = strings.Clone(name)
	internal.file.name = &name
	rebuildLoggers()
}

// SetFileSize 设置滚动日志文件大小（字节），最小 10 KB。
// 注意：底层由 lumberjack 管理，实际最小滚动粒度为 1 MB。
func SetFileSize(size int64) {
	if size < 10*1000 {
		size = 10 * 1000
	}
	internal.file.size = size
	rebuildLoggers()
}
