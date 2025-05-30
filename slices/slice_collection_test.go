package slices

import (
	"fmt"
	"testing"
)

func testCollection(*testing.T) {
	a := []int{4, 6, 6, 9, 1, 1, 3}
	b := []int{4, 9, 9, 3, 3}
	c := []int{4, 9, 3, 6, 1}
	d := []int{4, 9, 9, 3, 3, 10}

	so(CollectionEqual(a, b), eq, false)
	so(CollectionEqual(a, c), eq, true) // 长度不同, 但是元素相同

	so(fmt.Sprint(CollectionDifference(a, b)), eq, "[6 1]")
	so(fmt.Sprint(CollectionDifference(b, a)), eq, "[]")

	so(fmt.Sprint(CollectionUnion(a, b)), eq, "[4 6 9 1 3]")
	so(fmt.Sprint(CollectionUnion(b, a)), eq, "[4 9 3 6 1]")

	so(fmt.Sprint(CollectionIntersection(a, b)), eq, "[4 9 3]")

	so(fmt.Sprint(CollectionSymmetricDifference(a, b)), eq, "[6 1]")
	so(fmt.Sprint(CollectionSymmetricDifference(a, d)), eq, "[6 1 10]")
}
