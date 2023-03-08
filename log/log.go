// Package log 封装一些通用的日志功能。底层实现可能调整, 但是对外暴露的接口是保持不变的
package log

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/Andrew-M-C/go.util/runtime/caller"
)

// -------- log without context --------

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
	if level >= internal.level.console {
		loggers = append(loggers, consoleLog(level))
	}

	// TODO: file logger

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

func callerDesc(ca caller.Caller) string {
	return fmt.Sprintf("%s, Line %d", ca.Func, ca.Line)
}

func timeDesc() string {
	return time.Now().In(internal.Beijing).Format("2006-01-02 15:04:05.000")
}

// -------- log with context --------

// DebugContextf 调试日志
func DebugContextf(ctx context.Context, f string, a ...any) {}

// DebugContext 调试日志
func DebugContext(ctx context.Context, a ...any) {}

// InfoContextf 信息日志
func InfoContextf(ctx context.Context, f string, a ...any) {}

// InfoContext 信息日志
func InfoContext(ctx context.Context, a ...any) {}

// WarnContextf 警告日志
func WarnContextf(ctx context.Context, f string, a ...any) {}

// WarnContext 警告日志
func WarnContext(ctx context.Context, a ...any) {}

// ErrorContextf 错误日志
func ErrorContextf(ctx context.Context, f string, a ...any) {}

// ErrorContext 错误日志
func ErrorContext(ctx context.Context, a ...any) {}

// FatalContextf 崩溃日志
func FatalContextf(ctx context.Context, f string, a ...any) {}

// FatalContext 崩溃日志
func FatalContext(ctx context.Context, a ...any) {}
