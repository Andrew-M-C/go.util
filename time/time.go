package time

import "time"

var (
	// Beijing 表示北京时间
	Beijing = time.FixedZone("Asia/Beijing", 8*60*60)
)

var (
	startTime = time.Now()
)

// Reference:
//   - [Golang中实现禁止拷贝](https://jiajunhuang.com/articles/2018_11_12-golang_nocopy.md.html)
//   - [runtime: add NoCopy documentation struct type?](https://github.com/golang/go/issues/8005)
type noCopy struct{}

// Lock is a no-op used by -copylocks checker from `go vet`.
func (*noCopy) Lock()   {}
func (*noCopy) Unlock() {}
