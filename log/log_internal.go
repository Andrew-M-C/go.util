package log

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/fatih/color"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
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
	}

	caller struct {
		skip int
	}

	// rebuildMu 序列化 rebuildLoggers 调用，避免并发重建时 lumberjack 被重复关闭
	rebuildMu sync.Mutex

	levelToString []string
	debugf        func(f string, a ...any)
}{}

// loggers 通过 atomic.Pointer 存储，使热路径 (log 调用) 完全无锁
var (
	atomicConsoleLogger atomic.Pointer[zap.Logger]
	atomicFileLogger    atomic.Pointer[zap.Logger]
	atomicJack          atomic.Pointer[lumberjack.Logger]
)

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
	internal.file.size = 500 * 1000 * 1000 // 500 MB (in bytes)
	internal.file.name = &log

	internal.debugf = func(string, ...any) {}

	rebuildLoggers()
}

// rebuildLoggers 根据当前配置重建 zap logger 实例。
// 每次调用都会关闭旧的 lumberjack 并创建新的，调用应尽量少（配置变更时触发）。
func rebuildLoggers() {
	internal.rebuildMu.Lock()
	defer internal.rebuildMu.Unlock()

	callerSkip := internal.caller.skip + 3 // 3 = logf + doNonCtxLogf + public func

	opts := []zap.Option{
		zap.AddCaller(),
		zap.AddCallerSkip(callerSkip),
		// 不让 zap 在 Fatal 时调用 os.Exit，由 log.go 中的 os.Exit(-1) 负责
		zap.WithFatalHook(zapcore.WriteThenNoop),
	}

	// console logger
	consoleCore := zapcore.NewCore(
		zapcore.NewConsoleEncoder(consoleEncoderConfig()),
		zapcore.AddSync(os.Stdout),
		zapcore.DebugLevel, // 级别过滤由 getNonCtxLoggers / getCtxLoggers 负责
	)
	atomicConsoleLogger.Store(zap.New(consoleCore, opts...))

	// file logger (lumberjack)
	maxSizeMB := int(internal.file.size / (1000 * 1000))
	if maxSizeMB < 1 {
		maxSizeMB = 1
	}
	if prev := atomicJack.Load(); prev != nil {
		_ = prev.Close()
	}
	jack := &lumberjack.Logger{
		Filename:   *internal.file.name,
		MaxSize:    maxSizeMB,
		MaxBackups: 10,
	}
	atomicJack.Store(jack)

	fileCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(fileEncoderConfig()),
		zapcore.AddSync(jack),
		zapcore.DebugLevel,
	)
	atomicFileLogger.Store(zap.New(fileCore, opts...))

	internal.debugf("rebuilt loggers: callerSkip=%d, file=%s, maxSizeMB=%d",
		callerSkip, *internal.file.name, maxSizeMB)
}

// consoleEncoderConfig 控制台输出格式配置，按级别着色
func consoleEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		TimeKey:        "T",
		LevelKey:       "L",
		CallerKey:      "C",
		MessageKey:     "M",
		EncodeTime:     zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000"),
		EncodeLevel:    colorLevelEncoder,
		EncodeCaller:   customCallerEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
	}
}

// fileEncoderConfig 文件输出 JSON 格式配置，字段名与旧实现保持一致
func fileEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		CallerKey:      "location",
		MessageKey:     "content",
		EncodeTime:     zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000"),
		EncodeLevel:    uppercaseLevelEncoder,
		EncodeCaller:   customCallerEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
	}
}

// colorLevelEncoder 为控制台输出的 level 字段着色（与原 fatih/color 行为一致）
func colorLevelEncoder(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	switch l {
	case zapcore.FatalLevel:
		enc.AppendString(color.HiRedString("FATAL"))
	case zapcore.ErrorLevel:
		enc.AppendString(color.RedString("ERROR"))
	case zapcore.WarnLevel:
		enc.AppendString(color.YellowString("WARN"))
	case zapcore.DebugLevel:
		enc.AppendString(color.BlueString("DEBUG"))
	default:
		enc.AppendString(l.CapitalString())
	}
}

// uppercaseLevelEncoder 用于文件 JSON 输出的大写 level 字段
func uppercaseLevelEncoder(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(l.CapitalString())
}

// customCallerEncoder 复现原 callerDesc 格式：
// "github.com/user/pkg/log.go, Line 42, pkg.FuncName()"
func customCallerEncoder(c zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
	if !c.Defined {
		enc.AppendString("unknown")
		return
	}
	// c.Function 形如 "github.com/Andrew-M-C/go.util/log.Infof"
	funcBase := filepath.Base(c.Function)             // "log.Infof"
	prefix := strings.TrimRight(c.Function, funcBase) // "github.com/Andrew-M-C/go.util/"
	file := filepath.Base(c.File)                     // "log.go"
	enc.AppendString(fmt.Sprintf("%s%s, Line %d, %s()", prefix, file, c.Line, funcBase))
}

// toZapLevel 将本包的 Level 映射到 zapcore.Level。
// TraceLevel 映射到 DebugLevel（不再支持 Trace）。
func toZapLevel(l Level) zapcore.Level {
	switch l {
	case TraceLevel, DebugLevel:
		return zapcore.DebugLevel
	case InfoLevel:
		return zapcore.InfoLevel
	case WarnLevel:
		return zapcore.WarnLevel
	case ErrorLevel:
		return zapcore.ErrorLevel
	case FatalLevel:
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}
