package log

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/Andrew-M-C/go-bytesize"
	"github.com/Andrew-M-C/go.util/log/trace"
	"github.com/Andrew-M-C/go.util/runtime/caller"
	timeutil "github.com/Andrew-M-C/go.util/time"
)

type logItem struct {
	Time     string
	Location string
	Level    Level
	Content  string
	TraceID  string
}

func (l *logItem) marshalJSONWithBuffer(by []byte) ([]byte, error) {
	buff := bytes.NewBuffer(by[:0])
	buff.WriteByte('{')
	buff.WriteString(`"time":"`)
	buff.WriteString(l.Time)

	buff.WriteString(`","level":"`)
	buff.WriteString(l.Level.String())

	buff.WriteString(`","location":`)
	b, _ := json.Marshal(l.Location)
	buff.Write(b)

	buff.WriteString(`,"content":`)
	b, _ = json.Marshal(l.Content)
	buff.Write(b)

	if l.TraceID != "" {
		buff.WriteString(`,"trace_id":`)
		b, _ := json.Marshal(l.TraceID)
		buff.Write(b)
	}

	buff.WriteByte('}')
	return buff.Bytes(), nil
}

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
	ca := caller.GetCaller(internalGetCallerSkip())
	item := &logItem{
		Time:     timeDesc(),
		Level:    Level(l),
		Location: callerDesc(ca),
		Content:  fmt.Sprintf(f, a...),
	}

	l.add(item)
}

func (l fileLog) log(a ...any) {
	ca := caller.GetCaller(internalGetCallerSkip())
	item := &logItem{
		Time:     timeDesc(),
		Level:    Level(l),
		Location: callerDesc(ca),
		Content:  fmt.Sprint(a...),
	}
	l.add(item)
}

func (l fileLog) logCtxf(ctx context.Context, f string, a ...any) {
	id := trace.GetTraceID(ctx)
	ca := caller.GetCaller(internalGetCallerSkip())
	item := &logItem{
		Time:     timeDesc(),
		Level:    Level(l),
		Location: callerDesc(ca),
		Content:  fmt.Sprintf(f, a...),
		TraceID:  id,
	}
	l.add(item)
}

func (l fileLog) logCtx(ctx context.Context, a ...any) {
	id := trace.GetTraceID(ctx)
	ca := caller.GetCaller(internalGetCallerSkip())
	item := &logItem{
		Time:     timeDesc(),
		Level:    Level(l),
		Location: callerDesc(ca),
		Content:  fmt.Sprint(a...),
		TraceID:  id,
	}
	l.add(item)
}

func (l fileLog) add(item *logItem) {
	internal.file.lock.Lock()
	defer internal.file.lock.Unlock()

	internal.file.logs = append(internal.file.logs, item)
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
	prevBuffer := make([]*logItem, 0, 1000)

	iterate := func() {
		// 如果没有日志请求, 那么啥都不用做
		internal.file.lock.Lock()
		if len(internal.file.logs) == 0 {
			internal.file.lock.Unlock()
			return
		}
		internal.file.lock.Unlock()

		fd, name, err = renewFileHandle(name, fd)
		if err != nil {
			func() { consoleLog(ErrorLevel).logf("renew file log error: %v", err) }()
			return
		}

		// 切换一下
		internal.file.lock.Lock()
		prevBuffer, internal.file.logs = internal.file.logs, prevBuffer
		internal.file.lock.Unlock()

		writtenBytes := 0
		buff := make([]byte, 4096)
		for _, item := range prevBuffer {
			b, _ := item.marshalJSONWithBuffer(buff)
			n, _ := fd.Write(b)
			writtenBytes += n
			n, _ = fd.Write(newLine)
			writtenBytes += n
		}
		internal.debugf("写入 %d 行日志, %v, 文件: %v", len(prevBuffer), bytesize.Base10(writtenBytes), name)
		_ = fd.Sync()
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
		internal.debugf("日志文件未满 %s", prevName)
		// 那就不用重命名了
		return prevFd, name, nil
	}
	if st.Size() < internal.file.size {
		return prevFd, name, nil
	}

	// 需要重命名文件
	prevFd.Close()

	newFileName := func() string {
		now := time.Now().In(timeutil.Beijing).Format("2006-01-02-15:04:05")
		ext := filepath.Ext(name)
		if ext == "" {
			return fmt.Sprintf("%s_%s", name, now)
		}
		return fmt.Sprintf("%s_%s%s", strings.TrimSuffix(name, ext), now, ext)
	}()
	_ = os.Rename(name, newFileName)

	// 新建一个文件返回
	f, err := os.OpenFile(prevName, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		return prevFd, name, fmt.Errorf("open new file '%s' error: %w", name, err)
	}

	if err = clearOldLogFiles(name); err != nil {
		return f, name, fmt.Errorf("rename older files error: %w", err)
	}
	return f, name, err
}

// 清除旧日志文件
func clearOldLogFiles(name string) error {
	const maxFileNum = 10
	dir := filepath.Dir(name)

	// 首先检查日志下面的所有文件
	files, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	suffix := filepath.Ext(name)
	prefix := strings.TrimSuffix(filepath.Base(name), suffix)

	// 把文件名掐头去尾, 剩下的部分就理应是日期格式
	r := regexp.MustCompile(`_20\d\d-[01]\d-\d\d-[012]\d:\d\d:\d\d`)
	filesToHandle := make([]fs.DirEntry, 0, len(files))
	fileInfos := make(map[string]os.FileInfo, len(files))

	for _, f := range files {
		if f.IsDir() {
			internal.debugf("忽略目录 %s", f.Name())
			continue
		}
		base := strings.TrimPrefix(f.Name(), prefix)
		base = strings.TrimSuffix(base, suffix)

		// 剩余部分如果 match 的话, 那么就是需要处理的文件
		if !r.MatchString(base) {
			internal.debugf("忽略文件 %s (base: %v)", f.Name(), base)
			continue
		}

		s, err := os.Stat(f.Name())
		if err != nil {
			internal.debugf("stat file %v error:%v", f.Name(), err)
			continue
		}
		filesToHandle = append(filesToHandle, f)
		fileInfos[f.Name()] = s
	}

	total := len(filesToHandle)
	if total < maxFileNum {
		internal.debugf("no need to delete log file, cnt %d", total)
		return nil
	}

	// 排序, 从旧到新
	sort.Slice(filesToHandle, func(i, j int) bool {
		fi, fj := filesToHandle[i], filesToHandle[j]
		modTimeI := fileInfos[fi.Name()].ModTime()
		modTimeJ := fileInfos[fj.Name()].ModTime()
		return modTimeI.Before(modTimeJ)
	})

	remains := total
	for _, f := range filesToHandle {
		if err := os.Remove(f.Name()); err != nil {
			internal.debugf("removed file %v error: %v", f.Name(), err)
			continue
		}
		internal.debugf("removed log file %v", f.Name())
		remains--
		if remains < maxFileNum {
			break
		}
	}

	return nil
}
