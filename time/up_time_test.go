package time

import (
	"testing"
	"time"
)

func testUpTime(t *testing.T) {
	tm := UpTime()
	t.Logf("现在时间 %v, 启动耗时: %v, 启动时间: %v", time.Now(), tm, startTime)
}
