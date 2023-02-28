package slice

import (
	"testing"
)

func testList(t *testing.T) {
	cv("Shuffle", func() { testListShuffle(t) })
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
