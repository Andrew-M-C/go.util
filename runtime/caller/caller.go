// Package caller provides packaging for system runtime.Caller
package caller

import (
	"fmt"
	"runtime"
)

// Caller identifies a caller
type Caller struct {
	File File     `json:"file"`
	Func Function `json:"function"`
	Line int      `json:"line"`
}

// String implements fmt.Stringer
func (c Caller) String() string {
	return fmt.Sprintf("%s, %s(), Line %d", c.File, c.Func, c.Line)
}

// GetCaller get last caller. If skip is set to 0, will get yourself.
func GetCaller(skip int) Caller {
	pcs := make([]uintptr, 128)
	depth := runtime.Callers(skip+2, pcs)
	ca := runtime.CallersFrames(pcs[:depth])
	fr, _ := ca.Next()
	return Caller{
		File: File(fr.File),
		Func: Function(fr.Function),
		Line: fr.Line,
	}
}

// GetAllCallers get all caller infos
func GetAllCallers() (callers []Caller) {
	pcs := make([]uintptr, 128)
	depth := runtime.Callers(1, pcs)

	ca := runtime.CallersFrames(pcs[:depth])

	for {
		fr, more := ca.Next()
		callers = append(callers, Caller{
			File: File(fr.File),
			Func: Function(fr.Function),
			Line: fr.Line,
		})

		if !more {
			break
		}
	}

	return
}
