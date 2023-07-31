package time

import "time"

// UpTime 返回程序启动了多久。Wall time
func UpTime() time.Duration {
	return time.Since(startTime)
}
