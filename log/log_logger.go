package log

import "context"

// Logger 表示一个日志器
// type Logger interface {
// 	Debugf(f string, a ...any)
// 	Debug(a ...any)
// 	Infof(f string, a ...any)
// 	Info(a ...any)
// 	Warnf(f string, a ...any)
// 	Warn(a ...any)
// 	Errorf(f string, a ...any)
// 	Error(a ...any)
// 	Fatalf(f string, a ...any)
// 	Fatal(a ...any)
// 	DebugContextf(ctx context.Context, f string, a ...any)
// 	DebugContext(ctx context.Context, a ...any)
// 	InfoContextf(ctx context.Context, f string, a ...any)
// 	InfoContext(ctx context.Context, a ...any)
// 	WarnContextf(ctx context.Context, f string, a ...any)
// 	WarnContext(ctx context.Context, a ...any)
// 	ErrorContextf(ctx context.Context, f string, a ...any)
// 	ErrorContext(ctx context.Context, a ...any)
// 	FatalContextf(ctx context.Context, f string, a ...any)
// 	FatalContext(ctx context.Context, a ...any)
// }

type nonCtxLogger interface {
	logf(f string, a ...any)
	log(a ...any)
}

type ctxLogger interface {
	logCtxf(ctx context.Context, f string, a ...any)
	logCtx(ctx context.Context, a ...any)
}
