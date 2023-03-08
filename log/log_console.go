package log

import (
	"context"
	"fmt"

	"github.com/Andrew-M-C/go.util/log/trace"
	"github.com/Andrew-M-C/go.util/runtime/caller"
	"github.com/fatih/color"
)

type consoleLog Level

func (l consoleLog) getLogger() func(string, ...any) string {
	var fu func(string, ...any) string
	switch Level(l) {
	default:
		fu = fmt.Sprintf
	case FatalLevel:
		fu = color.HiRedString
	case ErrorLevel:
		fu = color.RedString
	case WarnLevel:
		fu = color.YellowString
	}
	return fu
}

func (l consoleLog) logf(f string, a ...any) {
	fu := l.getLogger()
	ca := caller.GetCaller(callerSkip)
	f = fmt.Sprintf("%s - %s - %s - %s", timeDesc(), Level(l).String(), callerDesc(ca), f)
	s := fu(f, a...)
	fmt.Println(s)
}

func (l consoleLog) log(a ...any) {
	fu := l.getLogger()
	ca := caller.GetCaller(callerSkip)
	f := fmt.Sprintf("%s - %s - %s", timeDesc(), Level(l).String(), callerDesc(ca))
	s := fmt.Sprint(a...)
	s = fu("%s - %s", f, s)
	fmt.Println(s)
}

func (l consoleLog) logCtxf(ctx context.Context, f string, a ...any) {
	id := trace.GetTraceID(ctx)
	fu := l.getLogger()
	ca := caller.GetCaller(callerSkip)

	f = fmt.Sprintf("%s - %s - %s - %s", timeDesc(), Level(l).String(), callerDesc(ca), f)
	s := fu(f, a...)

	if id == "" {
		fmt.Println(s)
	} else {
		fmt.Println(s, fmt.Sprintf("(trace ID: %s)", id))
	}
}

func (l consoleLog) logCtx(ctx context.Context, a ...any) {
	id := trace.GetTraceID(ctx)
	fu := l.getLogger()
	ca := caller.GetCaller(callerSkip)
	f := fmt.Sprintf("%s - %s - %s", timeDesc(), Level(l).String(), callerDesc(ca))
	s := fmt.Sprint(a...)
	s = fu("%s - %s", f, s)

	if id == "" {
		fmt.Println(s)
	} else {
		fmt.Println(s, fmt.Sprintf("- %s", id))
	}
}
