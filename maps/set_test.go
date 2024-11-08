package maps_test

import (
	"fmt"
	"testing"

	"github.com/Andrew-M-C/go.util/maps"
)

func testSet(*testing.T) {
	cv("基本操作", func() {
		const testSize = 1000
		s := maps.NewSet[int]()
		for i := 0; i < testSize; i++ {
			s.Add(i)
		}
		c := s.Clone()
		so(fmt.Sprintf("%p", s), ne, fmt.Sprintf("%p", c))

		for i := 0; i < testSize; i++ {
			so(c.Has(i), eq, true)
		}

		so(s.Equal(c), eq, true)

		c.Del(1)
		so(c.Len(), eq, testSize-1)
		so(c.Has(1), eq, false)
		so(s.Has(1), eq, true)

		so(s.Equal(c), eq, false)
		so(c.Difference(s).Has(1), eq, false)
		so(s.Difference(c).Has(1), eq, true)
		so(s.Difference(c).Has(2), eq, false)
		so(s.Union(c).Len(), eq, testSize)
		so(s.Intersection(c).Len(), eq, testSize-1)
		so(s.Intersection(c).Has(1), eq, false)
		so(s.Intersection(c).Has(2), eq, true)
		so(s.Intersection(c).Has(testSize-1), eq, true)
		so(s.SymmetricDifference(c).Len(), eq, 1)
		so(s.SymmetricDifference(c).Has(1), eq, true)
	})
}
