package log

import (
	"context"
	"os"
)

// Logger 表示一个日志器
type Logger interface {
	Tracef(f string, a ...any)
	Trace(a ...any)
	Debugf(f string, a ...any)
	Debug(a ...any)
	Infof(f string, a ...any)
	Info(a ...any)
	Warnf(f string, a ...any)
	Warn(a ...any)
	Errorf(f string, a ...any)
	Error(a ...any)
	Fatalf(f string, a ...any)
	Fatal(a ...any)
}

type nonCtxLogger interface {
	logf(f string, a ...any)
	log(a ...any)
}

type ctxLogger interface {
	logCtxf(ctx context.Context, f string, a ...any)
	logCtx(ctx context.Context, a ...any)
}

// NewLogger 返回一个日志器
func NewLogger(ctx ...context.Context) Logger {
	if len(ctx) == 0 || ctx[0] == nil {
		return loggerImplWithoutCtx{}
	}
	return &loggerImplWithCtx{ctx: ctx[0]}
}

// -------- logger without context --------

type loggerImplWithoutCtx struct{}

// Tracef 调试日志
func (loggerImplWithoutCtx) Tracef(f string, a ...any) {
	l := getNonCtxLoggers(TraceLevel)
	doNonCtxLogf(l, f, a...)
}

// Trace 调试日志
func (loggerImplWithoutCtx) Trace(a ...any) {
	l := getNonCtxLoggers(TraceLevel)
	doNonCtxLog(l, a...)
}

// Debugf 调试日志
func (loggerImplWithoutCtx) Debugf(f string, a ...any) {
	l := getNonCtxLoggers(DebugLevel)
	doNonCtxLogf(l, f, a...)
}

// Debug 调试日志
func (loggerImplWithoutCtx) Debug(a ...any) {
	l := getNonCtxLoggers(DebugLevel)
	doNonCtxLog(l, a...)
}

// Infof 信息日志
func (loggerImplWithoutCtx) Infof(f string, a ...any) {
	l := getNonCtxLoggers(InfoLevel)
	doNonCtxLogf(l, f, a...)
}

// Info 信息日志
func (loggerImplWithoutCtx) Info(a ...any) {
	l := getNonCtxLoggers(InfoLevel)
	doNonCtxLog(l, a...)
}

// Warnf 警告日志
func (loggerImplWithoutCtx) Warnf(f string, a ...any) {
	l := getNonCtxLoggers(WarnLevel)
	doNonCtxLogf(l, f, a...)
}

// Warn 警告日志
func (loggerImplWithoutCtx) Warn(a ...any) {
	l := getNonCtxLoggers(WarnLevel)
	doNonCtxLog(l, a...)
}

// Errorf 错误日志
func (loggerImplWithoutCtx) Errorf(f string, a ...any) {
	l := getNonCtxLoggers(ErrorLevel)
	doNonCtxLogf(l, f, a...)
}

// Error 错误日志
func (loggerImplWithoutCtx) Error(a ...any) {
	l := getNonCtxLoggers(ErrorLevel)
	doNonCtxLog(l, a...)
}

// Fatalf 崩溃日志
func (loggerImplWithoutCtx) Fatalf(f string, a ...any) {
	l := getNonCtxLoggers(FatalLevel)
	doNonCtxLogf(l, f, a...)
	os.Exit(-1)
}

// Fatal 崩溃日志
func (loggerImplWithoutCtx) Fatal(a ...any) {
	l := getNonCtxLoggers(FatalLevel)
	doNonCtxLog(l, a...)
	os.Exit(-1)
}

// -------- logger without context --------

type loggerImplWithCtx struct {
	ctx context.Context
}

// Tracef 底层跟踪日志
func (c *loggerImplWithCtx) Tracef(f string, a ...any) {
	l := getCtxLoggers(c.ctx, TraceLevel)
	doCtxLogf(c.ctx, l, f, a...)
}

// Trace 底层跟踪日志
func (c *loggerImplWithCtx) Trace(a ...any) {
	l := getCtxLoggers(c.ctx, TraceLevel)
	doCtxLog(c.ctx, l, a...)
}

// Debugf 调试日志
func (c *loggerImplWithCtx) Debugf(f string, a ...any) {
	l := getCtxLoggers(c.ctx, DebugLevel)
	doCtxLogf(c.ctx, l, f, a...)
}

// Debug 调试日志
func (c *loggerImplWithCtx) Debug(a ...any) {
	l := getCtxLoggers(c.ctx, DebugLevel)
	doCtxLog(c.ctx, l, a...)
}

// Infof 信息日志
func (c *loggerImplWithCtx) Infof(f string, a ...any) {
	l := getCtxLoggers(c.ctx, InfoLevel)
	doCtxLogf(c.ctx, l, f, a...)
}

// Info 信息日志
func (c *loggerImplWithCtx) Info(a ...any) {
	l := getCtxLoggers(c.ctx, InfoLevel)
	doCtxLog(c.ctx, l, a...)
}

// Warnf 警告日志
func (c *loggerImplWithCtx) Warnf(f string, a ...any) {
	l := getCtxLoggers(c.ctx, WarnLevel)
	doCtxLogf(c.ctx, l, f, a...)
}

// WarnContext 警告日志
func (c *loggerImplWithCtx) Warn(a ...any) {
	l := getCtxLoggers(c.ctx, WarnLevel)
	doCtxLog(c.ctx, l, a...)
}

// Errorf 错误日志
func (c *loggerImplWithCtx) Errorf(f string, a ...any) {
	l := getCtxLoggers(c.ctx, ErrorLevel)
	doCtxLogf(c.ctx, l, f, a...)
}

// Error 错误日志
func (c *loggerImplWithCtx) Error(a ...any) {
	l := getCtxLoggers(c.ctx, ErrorLevel)
	doCtxLog(c.ctx, l, a...)
}

// Fatalf 崩溃日志
func (c *loggerImplWithCtx) Fatalf(f string, a ...any) {
	l := getCtxLoggers(c.ctx, FatalLevel)
	doCtxLogf(c.ctx, l, f, a...)
	os.Exit(-1)
}

// Fatal 崩溃日志
func (c *loggerImplWithCtx) Fatal(a ...any) {
	l := getCtxLoggers(c.ctx, FatalLevel)
	doCtxLog(c.ctx, l, a...)
	os.Exit(-1)
}
