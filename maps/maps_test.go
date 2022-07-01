// Package sync 提供一些额外的、非常规的 sync 功能
package maps

import (
	"math/rand"
	"testing"
	"time"

	"github.com/smartystreets/goconvey/convey"
)

var (
	cv = convey.Convey
	so = convey.So
	eq = convey.ShouldEqual
)

func init() {
	rand.Seed(time.Now().UnixMicro())
}

func TestContext(t *testing.T) {
	cv("测试 StringKeys", t, func() { testStringKeys(t) })
	cv("测试 IntKeys", t, func() { testIntKeys(t) })
	cv("测试 UintKeys", t, func() { testUintKeys(t) })
}

func testStringKeys(t *testing.T) {
	cv("基本逻辑", func() {
		const repeat = 10000
		m := map[string]int{
			"one": 1,
			"two": 2,
		}

		keys := StringKeys(m)
		t.Log(keys)
		so(len(keys), eq, len(m))

		for i := 0; i < repeat; i++ {
			keys := StringKeysSorted(m)
			// keys := StringKeys(m)
			so(keys[0], eq, "one")
			so(keys[1], eq, "two")
		}
	})
}

func testIntKeys(t *testing.T) {
	cv("基本逻辑", func() {
		const repeat = 10000
		m := map[int]int{
			-10000: -1,
			10000:  1,
		}

		keys := IntKeys(m)
		t.Log(keys)
		so(len(keys), eq, len(m))

		for i := 0; i < repeat; i++ {
			keys := IntKeysSorted(m)
			// keys := StringKeys(m)
			so(keys[0], eq, -10000)
			so(keys[1], eq, 10000)
		}
	})
}

func testUintKeys(t *testing.T) {
	cv("基本逻辑", func() {
		const repeat = 10000
		m := map[uint64]bool{
			1:     true,
			10000: false,
		}

		keys := UintKeys(m)
		t.Log(keys)
		so(len(keys), eq, len(m))

		for i := 0; i < repeat; i++ {
			keys := UintKeysSorted(m)
			// keys := StringKeys(m)
			so(keys[0], eq, 1)
			so(keys[1], eq, 10000)
		}
	})
}
