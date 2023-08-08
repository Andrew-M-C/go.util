package log

import (
	"sync"
)

var internal = struct {
	level struct {
		normal struct {
			file    Level
			console Level
		}
		dyeing struct {
			file    Level
			console Level
		}
	}

	file struct {
		name *string
		size int64
		lock sync.Mutex // TODO: 以后再用更高性能的方案代替, 暂时先实现功能
		logs []*logItem
	}

	caller struct {
		skip int
	}

	levelToString []string
	debugf        func(f string, a ...any)
}{}

func internalGetCallerSkip() int {
	return internal.caller.skip + 3
}

func init() {
	internal.level.normal.file = NoLog
	internal.level.normal.console = InfoLevel
	internal.level.dyeing.file = NoLog
	internal.level.dyeing.console = InfoLevel

	internal.levelToString = []string{
		"TRACE",
		"DEBUG",
		"INFO",
		"WARN",
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
