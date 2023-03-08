package log

import (
	"fmt"

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
