package maps_test

import (
	"testing"

	"github.com/Andrew-M-C/go.util/maps"
)

func TestOrderedMap(t *testing.T) {
	cv("OrderedMap 迭代器测试", t, func() {
		testOrderedMapAll(t)
		testOrderedMapKeys(t)
		testOrderedMapEarlyBreak(t)
		testOrderedMapEmpty(t)
	})
}

func testOrderedMapAll(t *testing.T) {
	cv("All() 迭代器", func() {
		m := maps.NewOrderedMap[int, string]()
		m.Set(3, "three")
		m.Set(1, "one")
		m.Set(2, "two")
		m.Set(5, "five")
		m.Set(4, "four")

		// 验证按照 key 升序遍历
		expectedKeys := []int{1, 2, 3, 4, 5}
		expectedValues := []string{"one", "two", "three", "four", "five"}

		i := 0
		for k, v := range m.All() {
			so(k, eq, expectedKeys[i])
			so(v, eq, expectedValues[i])
			i++
		}
		so(i, eq, 5)
	})
}

func testOrderedMapKeys(t *testing.T) {
	cv("Keys() 迭代器", func() {
		m := maps.NewOrderedMap[int, string]()
		m.Set(3, "three")
		m.Set(1, "one")
		m.Set(2, "two")

		// 验证按照 key 升序遍历
		expectedKeys := []int{1, 2, 3}

		i := 0
		for k := range m.Keys() {
			so(k, eq, expectedKeys[i])
			i++
		}
		so(i, eq, 3)
	})
}

func testOrderedMapEarlyBreak(t *testing.T) {
	cv("提前 break 的情况", func() {
		m := maps.NewOrderedMap[int, string]()
		m.Set(1, "one")
		m.Set(2, "two")
		m.Set(3, "three")
		m.Set(4, "four")
		m.Set(5, "five")

		// 测试提前 break
		count := 0
		for k, v := range m.All() {
			count++
			if k == 3 {
				so(v, eq, "three")
				break
			}
		}
		so(count, eq, 3)

		// 测试 Keys 提前 break
		count = 0
		for k := range m.Keys() {
			count++
			if k == 2 {
				break
			}
		}
		so(count, eq, 2)
	})
}

func testOrderedMapEmpty(t *testing.T) {
	cv("空 map 的情况", func() {
		m := maps.NewOrderedMap[int, string]()

		// All() 不应该执行任何迭代
		count := 0
		for range m.All() {
			count++
		}
		so(count, eq, 0)

		// Keys() 不应该执行任何迭代
		count = 0
		for range m.Keys() {
			count++
		}
		so(count, eq, 0)
	})
}
