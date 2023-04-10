// Package sync 提供一些额外的、非常规的 sync 功能
package maps

import (
	"fmt"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

var (
	cv = convey.Convey
	so = convey.So
	eq = convey.ShouldEqual
	ne = convey.ShouldNotEqual
)

func TestMaps(t *testing.T) {
	cv("测试 StringKeys", t, func() { testStringKeys(t) })
	cv("测试 IntKeys", t, func() { testIntKeys(t) })
	cv("测试 UintKeys", t, func() { testUintKeys(t) })
	cv("测试 Equal 和 KeysEqual", t, func() { testEqual(t) })
	cv("测试 GetOrDefault 和 GetStringOrFormat", t, func() { testGetFunctions(t) })
	cv("测试 maps_kv_pair.go", t, func() { testKVPairs(t) })
}

func testStringKeys(t *testing.T) {
	cv("基本逻辑", func() {
		const repeat = 10000
		m := map[string]int{
			"one": 1,
			"two": 2,
		}

		keys := Keys(m)
		t.Log(keys)
		so(len(keys), eq, len(m))

		for i := 0; i < repeat; i++ {
			keys.SortAsc()
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

		keys := Keys(m)
		t.Log(keys)
		so(len(keys), eq, len(m))

		for i := 0; i < repeat; i++ {
			keys.SortAsc()
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

		keys := Keys(m)
		t.Log(keys)
		so(len(keys), eq, len(m))

		for i := 0; i < repeat; i++ {
			keys.SortAsc()
			// keys := StringKeys(m)
			so(keys[0], eq, 1)
			so(keys[1], eq, 10000)
		}
	})
}

func testEqual(t *testing.T) {
	cv("Equal()", func() {
		a := map[int]int{
			1: -1,
			2: -22,
		}
		b := map[int]int{
			1: -1,
			2: -2,
		}
		so(Equal(a, b), eq, false)

		a[2] = -2
		so(Equal(a, b), eq, true)
	})

	cv("KeysEqual", func() {
		a := map[int]struct{}{
			1: {},
			2: {},
		}
		b := map[int]int{
			1: 1,
			2: 22,
		}
		so(KeysEqual(a, b), eq, true)
	})
}

func testGetFunctions(t *testing.T) {
	cv("GetOrDefault", func() {
		m := map[int]string{
			1: "one",
			2: "two",
		}

		v := GetOrDefault(m, 1, "1")
		so(v, eq, "one")

		v = GetOrDefault(m, 3, "三")
		so(v, eq, "三")
	})

	cv("GetStringOrFormat", func() {
		m := map[int]string{
			1: "one",
			2: "two",
		}

		format := "unrecognized key %d"
		v := GetStringOrFormat(m, 1, format)
		so(v, eq, m[1])

		v = GetStringOrFormat(m, 3, format)
		so(v, ne, m[3])
		so(v, eq, fmt.Sprintf(format, 3))
	})
}
