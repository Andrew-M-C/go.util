package sync

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

var (
	cv = convey.Convey
	so = convey.So
	eq = convey.ShouldEqual

	isTrue  = convey.ShouldBeTrue
	isFalse = convey.ShouldBeFalse

	notPanic = convey.ShouldNotPanic
)

func TestContext(t *testing.T) {
	cv("测试 spinlock", t, func() { testSpinLock(t) })
	cv("测试可重入的锁", t, func() { testReentrantLock(t) })
	cv("测试 Map", t, func() { testMap(t) })
	cv("测试 sync.Pool", t, func() { testPool(t) })
}
