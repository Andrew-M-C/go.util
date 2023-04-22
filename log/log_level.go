package log

// Level 表示日志级别
type Level uint8

const (
	// DebugLevel 调试级别日志
	DebugLevel Level = iota
	// InfoLevel 信息级别日志
	InfoLevel
	// WarnLevel 警告级别日志
	WarnLevel
	// ErrorLevel 错误级别日志
	ErrorLevel
	// FatalLevel 崩溃日志
	FatalLevel
	// NoLog 不输出任何日志
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
