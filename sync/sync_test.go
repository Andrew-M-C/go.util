// Package sync 提供一些额外的、非常规的 sync 功能
package sync

import (
	"math/rand"
	"testing"
	"time"

	"github.com/smartystreets/goconvey/convey"
)

var (
	cv = convey.Convey
	so = convey.So

	isTrue  = convey.ShouldBeTrue
	isFalse = convey.ShouldBeFalse

	notPanic = convey.ShouldNotPanic
)

func init() {
	rand.Seed(time.Now().UnixMicro())
}

func TestContext(t *testing.T) {
	cv("测试 spinlock", t, func() { testSpinLock(t) })
	cv("测试可重入的锁", t, func() { testReentrantLock(t) })
}
