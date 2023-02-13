package cpu

import (
	"runtime"

	"go.uber.org/automaxprocs/maxprocs"
)

var internal = struct {
	cpuNum int
}{}

func init() {
	undo, _ := maxprocs.Set()
	internal.cpuNum = runtime.GOMAXPROCS(0)
	undo()
}
