// Package caller provides packaging for system runtime.Caller
package caller

import (
	"fmt"
	"runtime"
)

// Caller identifies a caller
type Caller struct {
	File File
	Func Function
	Line int
}

// String implemets fmt.Stringer
func (c Caller) String() string {
	return fmt.Sprintf("%s, %s(), Line %d", c.File, c.Func, c.Line)
}

// GetCaller get last caller. If skip is set to 0, will get yourself.
func GetCaller(skip int) Caller {
	pc, _, _, ok := runtime.Caller(skip + 1)
	if !ok {
		return Caller{
			Func: "(unknown)",
			File: "(unknown)",
			Line: 0,
		}
	}

	ca := runtime.CallersFrames([]uintptr{pc})
	fr, _ := ca.Next()

	return Caller{
		File: File(fr.File),
		Func: Function(fr.Function),
		Line: fr.Line,
	}
}
