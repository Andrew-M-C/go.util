package log

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Andrew-M-C/go.util/log/trace"
	"github.com/Andrew-M-C/go.util/runtime/caller"
)

// SetFileName 设置文件名
func SetFileName(name string) {
	bdr := strings.Builder{}
	bdr.WriteString(name)
	name = bdr.String() // copy

	if name != "" {
		internal.file.name = &name
	}
}

// SetFileSize 设置滚动日志文件大小, 最小 10 KB
func SetFileSize(size int64) {
	if size < 10*1000 {
		size = 10 * 1000
	}
	internal.file.size = size
}

type fileLog Level

func (l fileLog) logf(f string, a ...any) {
	ca := caller.GetCaller(callerSkip)
	f = fmt.Sprintf("%s - %s - %s - %s", timeDesc(), Level(l).String(), callerDesc(ca), f)
	s := fmt.Sprintf(f, a...)
	l.add(s)
}

func (l fileLog) log(a ...any) {
	ca := caller.GetCaller(callerSkip)
	f := fmt.Sprintf("%s - %s - %s", timeDesc(), Level(l).String(), callerDesc(ca))
	s := fmt.Sprint(a...)
	s = fmt.Sprintf("%s - %s", f, s)
	l.add(s)
}

func (l fileLog) logCtxf(ctx context.Context, f string, a ...any) {
	id := trace.GetTraceID(ctx)
	ca := caller.GetCaller(callerSkip)

	if id == "" {
		f = fmt.Sprintf("%s - %s - %s - %s", timeDesc(), Level(l).String(), callerDesc(ca), f)
	} else {
		f = fmt.Sprintf(
			"%s - %s - %s - %s (trace ID: %s)",
			timeDesc(), Level(l).String(), callerDesc(ca), f, id,
		)
	}

	s := fmt.Sprintf(f, a...)
	l.add(s)
}

func (l fileLog) logCtx(ctx context.Context, a ...any) {
	id := trace.GetTraceID(ctx)
	ca := caller.GetCaller(callerSkip)
	f := fmt.Sprintf("%s - %s - %s", timeDesc(), Level(l).String(), callerDesc(ca))
	s := fmt.Sprint(a...)
	s = fmt.Sprintf("%s - %s", f, s)

	if id == "" {
		l.add(s)
	} else {
		l.add(fmt.Sprint(s, fmt.Sprintf("- %s", id)))
	}
}

func (l fileLog) add(s string) {
	internal.file.lock.Lock()
	defer internal.file.lock.Unlock()

	internal.file.logs = append(internal.file.logs, s)
}

func fileLogRoutine() {
	defer func() {
		if e := recover(); e != nil {
			consoleLog(ErrorLevel).logf("panic, error: %v", e)
		}
		go fileLogRoutine()
	}()

	var fd *os.File
	var name string
	var err error

	newLine := []byte{'\n'}
	prevBuffer := make([]string, 0, 1000)

	iterate := func() {
		fd, name, err = renewFileHandle(name, fd)
		if err != nil {
			func() { consoleLog(ErrorLevel).logf("renew file log error: %v", err) }()
			return
		}

		// 切换一下
		internal.file.lock.Lock()
		prevBuffer, internal.file.logs = internal.file.logs, prevBuffer
		internal.file.lock.Unlock()

		for _, s := range prevBuffer {
			fd.WriteString(s)
			fd.Write(newLine)
		}
		fd.Sync()
		// func() { consoleLog(DebugLevel).logf("written %d lines to file %v", len(prevBuffer), name) }()
		prevBuffer = prevBuffer[:0]
	}

	for {
		time.Sleep(time.Second)
		iterate()
	}
}

func renewFileHandle(prevName string, prevFd *os.File) (fd *os.File, name string, err error) {
	name = *internal.file.name
	if name != prevName || prevFd == nil {
		if prevFd != nil {
			prevFd.Close()
		}
		// 直接打开新文件
		f, err := os.OpenFile(name, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
		if f == nil {
			return prevFd, name, fmt.Errorf("open '%s' error: %w", name, err)
		}
		return f, name, nil
	}

	// 检查一下文件大小是不是已经满了?
	st, err := os.Stat(prevName)
	if err != nil {
		// 那就不用重命名了
		return prevFd, name, nil
	}
	if st.Size() < internal.file.size {
		return prevFd, name, nil
	}

	// 需要重命名文件
	prevFd.Close()

	newFileName := func() string {
		now := time.Now().In(internal.Beijing).Format("2006-01-02-15:04:05")
		ext := filepath.Ext(name)
		if ext == "" {
			return fmt.Sprintf("%s_%s", name, now)
		}
		return fmt.Sprintf("%s_%s%s", strings.TrimSuffix(name, ext), now, ext)
	}()
	os.Rename(name, newFileName)

	// 新建一个文件返回
	f, err := os.OpenFile(prevName, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		return prevFd, name, fmt.Errorf("open new file '%s' error: %w", name, err)
	}
	return f, name, err
}
