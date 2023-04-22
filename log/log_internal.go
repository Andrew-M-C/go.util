package log

import (
	"sync"
)

const callerSkip = 3

var internal = struct {
	level struct {
		file    Level
		console Level
	}

	file struct {
		name *string
		size int64
		lock sync.Mutex // TODO: 以后再用更高性能的方案代替, 暂时先实现功能
		logs []string
	}

	levelToString []string
	debugf        func(f string, a ...any)
}{}

func init() {
	internal.level.file = InfoLevel
	internal.level.console = ErrorLevel

	internal.levelToString = []string{
		"DEBUG",
		"INFO ",
		"WARN ",
		"ERROR",
		"FATAL",
		"",
	}

	log := "./log.log"
	internal.file.size = 500 * 1000 * 1000 // 500 MB
	internal.file.name = &log
	go fileLogRoutine()

	internal.debugf = func(string, ...any) {}
}
