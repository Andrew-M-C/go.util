package time

import (
	"strconv"
	"strings"
	"time"
)

func init() {
	s := startTime.String()
	// 查找类似于 m=+0.000846169 的部分
	idx := strings.Index(s, "m=+")
	if idx < 0 {
		return
	}
	f, err := strconv.ParseFloat(s[idx+3:], 64)
	if err != nil {
		return
	}
	d := time.Duration(f * float64(time.Second))
	startTime = startTime.Add(-d)
}

// UpTime 返回程序启动了多久。单调时钟
func UpTime() time.Duration {
	return time.Since(startTime)
}
