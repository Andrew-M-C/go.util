package log

import (
	"context"
	"fmt"

	"github.com/Andrew-M-C/go.util/log/trace"
	"go.uber.org/zap"
)

type consoleLog Level

func (l consoleLog) logf(f string, a ...any) {
	atomicConsoleLogger.Load().Log(toZapLevel(Level(l)), fmt.Sprintf(f, a...))
}

func (l consoleLog) log(a ...any) {
	atomicConsoleLogger.Load().Log(toZapLevel(Level(l)), fmt.Sprint(a...))
}

func (l consoleLog) logCtxf(ctx context.Context, f string, a ...any) {
	logger := atomicConsoleLogger.Load()
	if id := trace.TraceID(ctx); id != "" {
		logger = logger.With(zap.String("trace_id", id))
	}
	logger.Log(toZapLevel(Level(l)), fmt.Sprintf(f, a...))
}

func (l consoleLog) logCtx(ctx context.Context, a ...any) {
	logger := atomicConsoleLogger.Load()
	if id := trace.TraceID(ctx); id != "" {
		logger = logger.With(zap.String("trace_id", id))
	}
	logger.Log(toZapLevel(Level(l)), fmt.Sprint(a...))
}
