// Package sync 提供一些额外的、非常规的 sync 功能
package sync

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

var (
	cv = convey.Convey
	so = convey.So

	isTrue  = convey.ShouldBeTrue
	isFalse = convey.ShouldBeFalse
)

func TestContext(t *testing.T) {
	cv("测试 spinlock", t, func() { testSpinLock(t) })
}
