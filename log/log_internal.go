package log

import (
	"sync"
	"time"
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

	Beijing *time.Location
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

	internal.Beijing = time.FixedZone("Asia/Beijing", 8*60*60)

	log := "./log.log"
	internal.file.size = 500 * 1000 * 1000 // 500 MB
	internal.file.name = &log
	go fileLogRoutine()
}
