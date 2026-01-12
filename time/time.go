package time

import (
	"time"
)

var (
	// Beijing 表示北京时间
	Beijing = time.FixedZone("Asia/Beijing", 8*60*60)
)

var (
	startTime = time.Now()
)

// Days 用天数生成 time.Duration
func Days[T Integer](days T) time.Duration {
	return time.Duration(days) * 24 * time.Hour
}

// Hour 用小时数生成 time.Duration
func Hour[T Integer](hours T) time.Duration {
	return time.Duration(hours) * time.Hour
}

// Min 用分钟数生成 time.Duration
func Min[T Integer](mins T) time.Duration {
	return time.Duration(mins) * time.Minute
}

// Sec 用秒数生成 time.Duration
func Sec[T Integer](secs T) time.Duration {
	return time.Duration(secs) * time.Second
}

// Milli 用毫秒数生成 time.Duration
func Milli[T Integer](msecs T) time.Duration {
	return time.Duration(msecs) * time.Millisecond
}

// UnixFloat 返回个位数为秒、但包含了小数的 UTC 时间戳值
func UnixFloat(t time.Time) float64 {
	micro := t.UnixMicro()
	return float64(micro) / float64(1000000)
}

// Reference:
//   - [Golang中实现禁止拷贝](https://jiajunhuang.com/articles/2018_11_12-golang_nocopy.md.html)
//   - [runtime: add NoCopy documentation struct type?](https://github.com/golang/go/issues/8005)
type noCopy struct{}

// Lock is a no-op used by -copylocks checker from `go vet`.
func (*noCopy) Lock()   {}
func (*noCopy) Unlock() {}
