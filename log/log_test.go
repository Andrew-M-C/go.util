package log_test

import (
	"testing"

	"github.com/Andrew-M-C/go.util/log"
	"github.com/smartystreets/goconvey/convey"
)

var (
	cv = convey.Convey
)

func TestLog(t *testing.T) {
	cv("调试", t, func() { testDebugging(t) })
}

func testDebugging(t *testing.T) {
	log.SetLevel(log.NoLog, log.DebugLevel)
	log.Debugf("Hello,", "debug")
	log.Warnf("Hello, %s!", "warning")
	log.Error("Hello", "error")

	log.Infof("以下应该没有日志")
	log.SetLevel(log.NoLog, log.NoLog)
	log.Error("Hello", "no error")
}
