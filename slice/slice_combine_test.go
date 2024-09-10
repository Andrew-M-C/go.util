package slice

import (
	"bytes"
	"fmt"
	"strconv"
	"testing"

	"golang.org/x/exp/constraints"
)

func testCombineEvenly(t *testing.T) {
	testCombineEvenlyProcess(t, 10, 7)
	testCombineEvenlyProcess(t, 15, 5)

	testCombineEvenlyProcess(t, 0, 0)
	testCombineEvenlyProcess(t, 1, 0)
	testCombineEvenlyProcess(t, 1, 1)
	testCombineEvenlyProcess(t, 30, 0)
	testCombineEvenlyProcess(t, 30, 1)
	testCombineEvenlyProcess(t, 14, 12)
	testCombineEvenlyProcess(t, 14, 13)

	testCombineEvenlyProcess(t, 30, 30)

	testCombineEvenlyProcess(t, 100, 77)

	testCombineEvenlyProcess(t, 100, 31)
}

var testSliceNum = 0

type testCombineEvenlyType struct {
	ID      string
	Display rune
}

func testCombineEvenlyProcess(t *testing.T, numA, numB int) {
	testSliceNum++
	t.Logf("\n======== Test CombineEvenly No.%02d, %d, %d ========", testSliceNum, numA, numB)

	a := testCombineEvenlyTypeSlice('|', numA)
	b := testCombineEvenlyTypeSlice('.', numB)

	printTestCombineEvenlyType(t, a)
	printTestCombineEvenlyType(t, b)

	res := CombineEvenly(a, b)
	printTestCombineEvenlyType(t, res)

	lessIntervals := map[int]struct{}{}
	lastLessIndex := -1
	lessValue := func() rune {
		if numA >= numB {
			return '.'
		}
		return '|'
	}()

	firstLessIndex := -1

	counts := map[rune]int{}
	existedID := map[string]struct{}{}

	for i, r := range res {
		_, exist := existedID[r.ID]
		so(exist, eq, false)

		counts[r.Display] = counts[r.Display] + 1
		existedID[r.ID] = struct{}{}

		if r.Display == lessValue {
			if lastLessIndex < 0 {
				firstLessIndex = i
			} else {
				lessIntervals[i-lastLessIndex] = struct{}{}
				// t.Log(i, "-", lastLessIndex, "=", i-lastLessIndex)
			}
			lastLessIndex = i
		}
	}

	so(counts['|'], eq, numA)
	so(counts['.'], eq, numB)

	t.Log("lessIntervals", lessIntervals)
	so(len(lessIntervals), le, 2)

	if len(lessIntervals) == 2 {
		indexes := make([]int, 0, 2)
		for i := range lessIntervals {
			indexes = append(indexes, i)
		}
		diff := indexes[0] - indexes[1]
		t.Logf("indexes: %v", indexes)
		so(abs(diff), eq, 1)
	}

	if firstLessIndex >= 0 {
		lastDiff := len(res) - 1 - lastLessIndex
		so(abs(lastDiff-firstLessIndex), le, 1)
	}
}

func abs[T constraints.Signed](n T) T {
	if n >= 0 {
		return n
	}
	return -n
}

func testCombineEvenlyTypeSlice(r rune, repeat int) []*testCombineEvenlyType {
	res := make([]*testCombineEvenlyType, repeat)
	for i := range res {
		res[i] = &testCombineEvenlyType{
			ID:      fmt.Sprintf("%s-%d", string(r), i),
			Display: r,
		}
	}
	return res
}

func printTestCombineEvenlyType(t *testing.T, slice []*testCombineEvenlyType) {
	buf := bytes.Buffer{}
	buf.WriteString("[ ")

	for _, item := range slice {
		buf.WriteRune(item.Display)
	}

	buf.WriteString(" ] len ")
	buf.WriteString(strconv.FormatInt(int64(len(slice)), 10))

	t.Log(buf.String())
}
