// Package log 封装一些通用的日志功能。底层实现可能调整, 但是对外暴露的接口是保持不变的
package log

import (
	"context"
	"os"

	"github.com/Andrew-M-C/go.util/log/dyeing"
)

// -------- log without context --------

// Tracef 底层跟踪日志（内部映射到 Debug 级别）
func Tracef(f string, a ...any) {
	l := getNonCtxLoggers(TraceLevel)
	doNonCtxLogf(l, f, a...)
}

// Trace 底层跟踪日志（内部映射到 Debug 级别）
func Trace(a ...any) {
	l := getNonCtxLoggers(TraceLevel)
	doNonCtxLog(l, a...)
}

// Debugf 调试日志
func Debugf(f string, a ...any) {
	l := getNonCtxLoggers(DebugLevel)
	doNonCtxLogf(l, f, a...)
}

// Debug 调试日志
func Debug(a ...any) {
	l := getNonCtxLoggers(DebugLevel)
	doNonCtxLog(l, a...)
}

// Infof 信息日志
func Infof(f string, a ...any) {
	l := getNonCtxLoggers(InfoLevel)
	doNonCtxLogf(l, f, a...)
}

// Info 信息日志
func Info(a ...any) {
	l := getNonCtxLoggers(InfoLevel)
	doNonCtxLog(l, a...)
}

// Warnf 警告日志
func Warnf(f string, a ...any) {
	l := getNonCtxLoggers(WarnLevel)
	doNonCtxLogf(l, f, a...)
}

// Warn 警告日志
func Warn(a ...any) {
	l := getNonCtxLoggers(WarnLevel)
	doNonCtxLog(l, a...)
}

// Errorf 错误日志
func Errorf(f string, a ...any) {
	l := getNonCtxLoggers(ErrorLevel)
	doNonCtxLogf(l, f, a...)
}

// Error 错误日志
func Error(a ...any) {
	l := getNonCtxLoggers(ErrorLevel)
	doNonCtxLog(l, a...)
}

// Fatalf 崩溃日志
func Fatalf(f string, a ...any) {
	l := getNonCtxLoggers(FatalLevel)
	doNonCtxLogf(l, f, a...)
	os.Exit(-1)
}

// Fatal 崩溃日志
func Fatal(a ...any) {
	l := getNonCtxLoggers(FatalLevel)
	doNonCtxLog(l, a...)
	os.Exit(-1)
}

func getNonCtxLoggers(level Level) (loggers []nonCtxLogger) {
	// console
	if level >= internal.level.normal.console {
		loggers = append(loggers, consoleLog(level))
	}

	// file logger
	if level >= internal.level.normal.file {
		loggers = append(loggers, fileLog(level))
	}

	return
}

func doNonCtxLogf(loggers []nonCtxLogger, f string, a ...any) {
	for _, l := range loggers {
		l.logf(f, a...)
	}
}

func doNonCtxLog(loggers []nonCtxLogger, a ...any) {
	for _, l := range loggers {
		l.log(a...)
	}
}

// -------- log with context --------

// Deprecated: TraceContextf 底层跟踪日志（内部映射到 Debug 级别）
func TraceContextf(ctx context.Context, f string, a ...any) {
	l := getCtxLoggers(ctx, TraceLevel)
	doCtxLogf(ctx, l, f, a...)
}

// Deprecated: TraceContext 底层跟踪日志（内部映射到 Debug 级别）
func TraceContext(ctx context.Context, a ...any) {
	l := getCtxLoggers(ctx, TraceLevel)
	doCtxLog(ctx, l, a...)
}

// DebugContextf 调试日志
func DebugContextf(ctx context.Context, f string, a ...any) {
	l := getCtxLoggers(ctx, DebugLevel)
	doCtxLogf(ctx, l, f, a...)
}

// DebugContext 调试日志
func DebugContext(ctx context.Context, a ...any) {
	l := getCtxLoggers(ctx, DebugLevel)
	doCtxLog(ctx, l, a...)
}

// InfoContextf 信息日志
func InfoContextf(ctx context.Context, f string, a ...any) {
	l := getCtxLoggers(ctx, InfoLevel)
	doCtxLogf(ctx, l, f, a...)
}

// InfoContext 信息日志
func InfoContext(ctx context.Context, a ...any) {
	l := getCtxLoggers(ctx, InfoLevel)
	doCtxLog(ctx, l, a...)
}

// WarnContextf 警告日志
func WarnContextf(ctx context.Context, f string, a ...any) {
	l := getCtxLoggers(ctx, WarnLevel)
	doCtxLogf(ctx, l, f, a...)
}

// WarnContext 警告日志
func WarnContext(ctx context.Context, a ...any) {
	l := getCtxLoggers(ctx, WarnLevel)
	doCtxLog(ctx, l, a...)
}

// ErrorContextf 错误日志
func ErrorContextf(ctx context.Context, f string, a ...any) {
	l := getCtxLoggers(ctx, ErrorLevel)
	doCtxLogf(ctx, l, f, a...)
}

// ErrorContext 错误日志
func ErrorContext(ctx context.Context, a ...any) {
	l := getCtxLoggers(ctx, ErrorLevel)
	doCtxLog(ctx, l, a...)
}

// FatalContextf 崩溃日志
func FatalContextf(ctx context.Context, f string, a ...any) {
	l := getCtxLoggers(ctx, FatalLevel)
	doCtxLogf(ctx, l, f, a...)
	os.Exit(-1)
}

// FatalContext 崩溃日志
func FatalContext(ctx context.Context, a ...any) {
	l := getCtxLoggers(ctx, FatalLevel)
	doCtxLog(ctx, l, a...)
	os.Exit(-1)
}

func getCtxLoggers(ctx context.Context, level Level) (loggers []ctxLogger) {
	isDyeing := dyeing.Dyeing(ctx)

	// console
	if level >= internal.level.normal.console {
		loggers = append(loggers, consoleLog(level))
	} else if isDyeing && level >= internal.level.dyeing.console {
		internal.debugf("dyeing with console")
		loggers = append(loggers, consoleLog(level))
	}

	// file logger
	if level >= internal.level.normal.file {
		loggers = append(loggers, fileLog(level))
	} else if isDyeing && level >= internal.level.dyeing.file {
		internal.debugf("dyeing with file")
		loggers = append(loggers, fileLog(level))
	}

	return
}

func doCtxLogf(ctx context.Context, loggers []ctxLogger, f string, a ...any) {
	for _, l := range loggers {
		l.logCtxf(ctx, f, a...)
	}
}

func doCtxLog(ctx context.Context, loggers []ctxLogger, a ...any) {
	for _, l := range loggers {
		l.logCtx(ctx, a...)
	}
}
