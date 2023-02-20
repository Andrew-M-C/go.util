package slice

import (
	"testing"

	"github.com/Andrew-M-C/go.util/maps"
)

func testList(t *testing.T) {
	cv("Shuffle", func() { testListShuffle(t) })
	cv("KeySet", func() { testListKeySet(t) })
}

func testListShuffle(t *testing.T) {
	orig := []int{1, 2, 3, 4, 5, 6, 7}
	equalCnt := 0

	const repeat = 1000
	const equalLimit = 2

	for i := 0; i < repeat; i++ {
		lst := List[int](orig).Copy()
		lst.Shuffle()

		if lst.Equal(orig) {
			equalCnt++
		}
	}

	t.Logf("Equal count: %d", equalCnt)
	so(equalCnt, le, equalLimit)
}

func testListKeySet(t *testing.T) {
	lst := []int{1, 1, 1, 1}
	set := ToList(lst).KeySet()
	so(len(set), eq, 1)
	so(set.Equal(maps.Set[int]{1: struct{}{}}), eq, true)

	lst = append(lst, -1)
	set = ToList(lst).KeySet()
	so(len(set), eq, 2)
	so(set.Has(1), eq, true)
	so(set.Has(-1), eq, true)
}
