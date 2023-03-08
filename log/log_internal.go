package log

import "time"

const callerSkip = 3

var internal = struct {
	level struct {
		file    Level
		console Level
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
}
