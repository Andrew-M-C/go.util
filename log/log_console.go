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
	ca := caller.GetCaller(internalGetCallerSkip())
	f = fmt.Sprintf("%s - %s - %s - %s", timeDesc(), Level(l).String(), callerDesc(ca), f)
	s := fu(f, a...)
	fmt.Println(s)
}

func (l consoleLog) log(a ...any) {
	fu := l.getLogger()
	ca := caller.GetCaller(internalGetCallerSkip())
	f := fmt.Sprintf("%s - %s - %s", timeDesc(), Level(l).String(), callerDesc(ca))
	s := fmt.Sprintln(a...)
	s = fu("%s - %s", f, s)
	fmt.Print(s)
}

func (l consoleLog) logCtxf(ctx context.Context, f string, a ...any) {
	id := trace.GetTraceID(ctx)
	ca := caller.GetCaller(internalGetCallerSkip())

	if id == "" {
		f = fmt.Sprintf(
			"%s - %s - %s - %s",
			timeDesc(), Level(l).String(), callerDesc(ca), f,
		)
	} else {
		f = fmt.Sprintf(
			"%s - %s - %s - %s {\"trace_id\":\"%s\")",
			timeDesc(), Level(l).String(), callerDesc(ca), f, id,
		)
	}

	s := fmt.Sprintf(f, a...)
	fmt.Println(s)
}

func (l consoleLog) logCtx(ctx context.Context, a ...any) {
	id := trace.GetTraceID(ctx)
	ca := caller.GetCaller(internalGetCallerSkip())
	f := fmt.Sprintf("%s - %s - %s", timeDesc(), Level(l).String(), callerDesc(ca))
	s := fmt.Sprint(a...)
	s = fmt.Sprintf("%s - %s", f, s)

	if id == "" {
		fmt.Println(s)
	} else {
		s = fmt.Sprint(s, fmt.Sprintf(` {"trace_id":"%s"}`, id))
		fmt.Println(s)
	}
}
