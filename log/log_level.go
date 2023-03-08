package log

// Level 表示日志级别
type Level uint8

const (
	DebugLevel Level = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
	NoLog
)

func (l Level) String() string {
	if l >= NoLog {
		l = NoLog
	}
	return internal.levelToString[l]
}

// SetLevel 设置
func SetLevel(file, console Level) {
	if file >= NoLog {
		file = NoLog + 1
	}
	if console >= NoLog {
		console = NoLog + 1
	}
	internal.level.file = file
	internal.level.console = console
}
