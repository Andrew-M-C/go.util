package sync

import "testing"

func testMap(*testing.T) {
	type value struct {
		Count int
	}

	cv("读写", func() {
		m := NewMap[string, *value]()
		v, exist := m.LoadOrStore("key", new(value))
		so(exist, isFalse)

		v.Count = 2
		v, exist = m.LoadOrStore("key", new(value))
		so(exist, isTrue)
		so(v.Count, eq, 2)
	})
}
