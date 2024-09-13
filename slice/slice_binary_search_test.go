package slice

import (
	"testing"
)

func testBinarySearch(t *testing.T) {
	cv("测试基本的命中情况", func() { testBinarySearchBasicHit(t) })
	cv("测试命中多于一个数据的情况", func() { testBinarySearchMultipleHit(t) })
	cv("测试空 slice", func() { testBinarySearchEmptySlice(t) })
	cv("测试 BinarySearchOne", func() { testBinarySearchOne(t) })
}

func testBinarySearchBasicHit(_ *testing.T) {
	cv("奇数长", func() {
		sli := []int{-20, -10, 10, 20, 30}

		from, to, hit := BinarySearch(sli, 20)
		so(hit, isTrue)
		so(from, eq, 3)
		so(to, eq, 4)

		from, to, hit = BinarySearch(sli, 15)
		so(hit, isFalse)
		so(from, eq, 2)
		so(to, eq, 3)

		from, to, hit = BinarySearch(sli, 0)
		so(hit, isFalse)
		so(from, eq, 1)
		so(to, eq, 2)

		from, to, hit = BinarySearch(sli, -100)
		so(hit, isFalse)
		so(from, eq, 0)
		so(to, eq, 0)

		from, to, hit = BinarySearch(sli, 100)
		so(hit, isFalse)
		so(from, eq, 5)
		so(to, eq, 5)
		so(len(sli[from:to]), eq, 0)
	})

	cv("偶数长", func() {
		sli := []int{-20, -10, 20, 30}

		from, to, hit := BinarySearch(sli, 20)
		so(hit, isTrue)
		so(from, eq, 2)
		so(to, eq, 3)

		from, to, hit = BinarySearch(sli, 25)
		so(hit, isFalse)
		so(from, eq, 2)
		so(to, eq, 3)

		from, to, hit = BinarySearch(sli, 0)
		so(hit, isFalse)
		so(from, eq, 1)
		so(to, eq, 2)

		from, to, hit = BinarySearch(sli, 100)
		so(hit, isFalse)
		so(from, eq, 4)
		so(to, eq, 4)
		so(len(sli[from:to]), eq, 0)
	})
}

func testBinarySearchMultipleHit(_ *testing.T) {
	sli := []int{-1, -1, 0, 0, 1, 2, 5, 5, 5, 6, 8, 10, 25, 25, 100}

	from, to, hit := BinarySearch(sli, 5)
	so(hit, isTrue)
	so(from, eq, 6)
	so(to, eq, 9)

	from, to, hit = BinarySearch(sli, 25)
	so(hit, isTrue)
	so(from, eq, 12)
	so(to, eq, 14)

	from, to, hit = BinarySearch(sli, 100)
	so(hit, isTrue)
	so(from, eq, 14)
	so(to, eq, 15)

	from, to, hit = BinarySearch(sli, 101)
	so(hit, isFalse)
	so(from, eq, 15)
	so(to, eq, 15)
	so(len(sli[from:to]), eq, 0)

	from, to, hit = BinarySearch(sli, -1)
	so(hit, isTrue)
	so(from, eq, 0)
	so(to, eq, 2)

	from, to, hit = BinarySearch(sli, -2)
	so(hit, isFalse)
	so(from, eq, 0)
	so(to, eq, 0)
	so(len(sli[from:to]), eq, 0)
}

func testBinarySearchEmptySlice(_ *testing.T) {
	var sli []int32

	from, to, hit := BinarySearch(sli, 100)
	so(hit, isFalse)
	so(from, eq, 0)
	so(to, eq, 0)
	so(len(sli[from:to]), eq, 0)

	sli = make([]int32, 0)
	from, to, hit = BinarySearch(sli, 100)
	so(hit, isFalse)
	so(from, eq, 0)
	so(to, eq, 0)
	so(len(sli[from:to]), eq, 0)
}

func testBinarySearchOne(_ *testing.T) {
	cv("基础逻辑 - 偶数个", func() {
		sli := []int{1, 3, 5, 7, 9, 11}
		comp := func(a, b int) int {
			if a < b {
				return -1
			}
			if a > b {
				return 1
			}
			return 0
		}
		i := BinarySearchOne(sli, 3, comp)
		so(i, eq, 1)
		i = BinarySearchOne(sli, 4, comp)
		so(i, eq, -1)
		i = BinarySearchOne(sli, 1, comp)
		so(i, eq, 0)
		i = BinarySearchOne(sli, 11, comp)
		so(i, eq, 5)
		i = BinarySearchOne(sli, 0, comp)
		so(i, eq, -1)
		i = BinarySearchOne(sli, 12, comp)
		so(i, eq, -1)
		i = BinarySearchOne(sli, 5, comp)
		so(i, eq, 2)
		i = BinarySearchOne(sli, 7, comp)
		so(i, eq, 3)
	})

	cv("基础逻辑 - 奇数个", func() {
		sli := []int{1, 3, 5, 7, 9}
		comp := func(a, b int) int {
			if a < b {
				return -1
			}
			if a > b {
				return 1
			}
			return 0
		}
		i := BinarySearchOne(sli, 3, comp)
		so(i, eq, 1)
		i = BinarySearchOne(sli, 4, comp)
		so(i, eq, -1)
		i = BinarySearchOne(sli, 1, comp)
		so(i, eq, 0)
		i = BinarySearchOne(sli, 9, comp)
		so(i, eq, 4)
		i = BinarySearchOne(sli, 0, comp)
		so(i, eq, -1)
		i = BinarySearchOne(sli, 12, comp)
		so(i, eq, -1)
		i = BinarySearchOne(sli, 5, comp)
		so(i, eq, 2)
	})
}
