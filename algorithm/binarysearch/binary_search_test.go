package binarysearch_test

import (
	"testing"

	"github.com/Andrew-M-C/go.util/algorithm/binarysearch"
	"github.com/smartystreets/goconvey/convey"
)

var (
	cv = convey.Convey
	so = convey.So
	eq = convey.ShouldEqual
)

func TestBinarySearch(t *testing.T) {
	cv("SearchOneByFunc", t, func() { testSearchOneByFunc(t) })
	cv("SearchFloor, SearchCeiling", t, func() { testSearchFloorAndCeiling(t) })
}

func testSearchOneByFunc(t *testing.T) {
	compare := func(a, b uint) int {
		switch {
		case a < b:
			return -1
		case a > b:
			return 1
		default:
			return 0
		}
	}

	tables := [][3]any{
		// input, target, return
		{[]uint{1, 2, 3, 4, 5, 6, 7, 8, 9}, 2, 1},
		{[]uint{1, 2, 3, 4, 5, 6, 7, 8, 9}, 1, 0},
		{[]uint{1, 2, 3, 4, 5, 6, 7, 8, 9}, 0, -1},
		{[]uint{1, 2, 3, 4, 5, 6, 7, 8, 9}, 10, -1},
		{[]uint{1, 2, 3, 4, 5, 6, 7, 8, 9}, 9, 8},
		{[]uint{1, 2, 3, 4, 5, 6, 7, 8, 9}, 6, 5},
		{[]uint{1, 2, 3, 4, 6, 7, 8, 9}, 5, -1},
		{[]uint{1, 2}, 1, 0},
		{[]uint{1, 2}, 2, 1},
		{[]uint{1, 2}, 0, -1},
		{[]uint{1, 2}, 3, -1},
		{[]uint{1, 2, 3}, 1, 0},
		{[]uint{1, 2, 3}, 2, 1},
		{[]uint{1, 2, 3}, 3, 2},
		{[]uint{1, 2, 3}, 0, -1},
		{[]uint{1, 2, 3}, 4, -1},
		{[]uint{}, 1, -1},
		{[]uint{1}, 2, -1},
		{[]uint{1}, 1, 0},
	}

	for _, tbl := range tables {
		arr, _ := tbl[0].([]uint)
		tgt := uint(tbl[1].(int))
		res, _ := tbl[2].(int)

		got := binarysearch.SearchOneByFunc(arr, tgt, compare)
		so(got, eq, res)
		got = binarysearch.SearchOne(arr, tgt)
		so(got, eq, res)
	}
}

func testSearchFloorAndCeiling(t *testing.T) {
	cases := [][4]any{
		// input, target, floor, ceiling
		{[]uint{1, 1, 1, 1, 2, 2, 2, 2, 4, 4, 6, 7, 7, 8}, 2, 4, 7},
		{[]uint{1, 1, 1, 1, 2, 2, 2, 2, 4, 4, 6, 7, 7, 8}, 6, 10, 10},
		{[]uint{1, 1, 1, 1, 2, 2, 2, 2, 4, 4, 6, 7, 7, 8}, 7, 11, 12},
		{[]uint{1, 1, 1, 1, 2, 2, 2, 2, 4, 4, 6, 7, 7, 8}, 1, 0, 3},
		{[]uint{1, 2, 2, 2, 2, 2, 2, 2, 4, 4, 6, 7, 7, 8}, 1, 0, 0},
		{[]uint{1, 2, 2, 2, 2, 2, 2, 2, 4, 4, 6, 7, 7, 8}, 2, 1, 7},
		{[]uint{1, 2, 2, 2, 2, 2, 2, 2, 4, 4, 6, 7, 7, 8}, 9, -1, -1},
		{[]uint{1, 2, 2, 2, 2, 2, 2, 2, 4, 4, 6, 7, 7, 8}, 0, -1, -1},
		{[]uint{1, 1, 1}, 1, 0, 2},
		{[]uint{1, 2, 2}, 2, 1, 2},
		{[]uint{1, 2, 3}, 3, 2, 2},
		{[]uint{1, 2, 3}, 0, -1, -1},
		{[]uint{1, 2, 3}, 4, -1, -1},
		{[]uint{}, 1, -1, -1},
		{[]uint{1}, 2, -1, -1},
		{[]uint{1}, 1, 0, 0},
		{make([]uint, 100000), 1, -1, -1},
		{make([]uint, 1000000), 0, 0, 999999},
	}

	for _, c := range cases {
		arr, _ := c[0].([]uint)
		tgt := uint(c[1].(int))
		floor, _ := c[2].(int)
		ceiling, _ := c[3].(int)

		got := binarysearch.SearchFloor(arr, tgt)
		so(got, eq, floor)
		got = binarysearch.SearchCeiling(arr, tgt)
		so(got, eq, ceiling)
	}
}
