package slice

import "testing"

func testBinarySearch(t *testing.T) {
	cv("测试基本的命中情况", func() { testBinarySearchBasicHit(t) })
	cv("测试命中多于一个数据的情况", func() { testBinarySearchMultipleHit(t) })
	cv("测试空 slice", func() { testBinarySearchEmptySlice(t) })
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
