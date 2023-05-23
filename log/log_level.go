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

// SetLevel 设置日志级别
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

// SetFileLevel 设置日志文件级别
func SetFileLevel(lv Level) {
	console := internal.level.console
	SetLevel(lv, console)
}

// SetConsoleLevel 设置 console 日志级别
func SetConsoleLevel(lv Level) {
	file := internal.level.file
	SetLevel(file, lv)
}

// SetSkipCaller 当外部封装本 logger 时, 可以设置该值, 那么 logger 在输出调用信息的时候
// 可以跳过指定的层数。
func SetSkipCaller(skip int) {
	if skip >= 0 {
		internal.caller.skip = skip
	}
}
