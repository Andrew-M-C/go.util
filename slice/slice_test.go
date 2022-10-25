package slice

import (
	"bytes"
	"fmt"
	"strconv"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

var (
	cv = convey.Convey
	so = convey.So

	eq = convey.ShouldEqual

	isErr   = convey.ShouldBeError
	isNil   = convey.ShouldBeNil
	isFalse = convey.ShouldBeFalse
)

func test(t *testing.T, scene string, f func(*testing.T)) {
	if t.Failed() {
		return
	}
	cv(scene, t, func() {
		f(t)
	})
}

func TestSlice(t *testing.T) {
	test(t, "CombineEvenly", testCombineEvenly)
	test(t, "LCS", testLCS)
}

func testCombineEvenly(t *testing.T) {
	testCombineEvenlyProcess(t, 10, 7)
	testCombineEvenlyProcess(t, 0, 0)
	testCombineEvenlyProcess(t, 1, 0)
	testCombineEvenlyProcess(t, 1, 1)
	testCombineEvenlyProcess(t, 30, 0)
	testCombineEvenlyProcess(t, 30, 1)
	testCombineEvenlyProcess(t, 14, 12)

	testCombineEvenlyProcess(t, 30, 30)

	testCombineEvenlyProcess(t, 100, 77)

	testCombineEvenlyProcess(t, 100, 31)

	testCombineEvenlyErrors(t)
}

var testSliceNum = 0

type testCombineEvenlyType struct {
	ID      string
	Display rune
}

func testCombineEvenlyProcess(t *testing.T, numA, numB int) {
	t.Logf("\n======== Test No.%02d ========", testSliceNum)
	testSliceNum++

	a := testCombineEvenlyTypeSlice('|', numA)
	b := testCombineEvenlyTypeSlice('.', numB)

	printTestCombineEvenlyType(t, a)
	printTestCombineEvenlyType(t, b)

	slice, err := CombineEvenly(a, b)
	so(err, isNil)

	res := slice.([]*testCombineEvenlyType)
	printTestCombineEvenlyType(t, res)

	counts := map[rune]int{}
	existedID := map[string]struct{}{}

	for _, r := range res {
		_, exist := existedID[r.ID]
		so(exist, isFalse)

		counts[r.Display] = counts[r.Display] + 1
		existedID[r.ID] = struct{}{}
	}

	so(counts['|'], eq, numA)
	so(counts['.'], eq, numB)
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

func testCombineEvenlyErrors(t *testing.T) {
	cv("mismatch element types", func() {
		type integer int
		s1 := []int{1, 2, 3, 4}
		s2 := []integer{1, 2, 3, 4}
		_, err := CombineEvenly(s1, s2)
		so(err, isErr)
	})

	cv("matched underlying types", func() {
		type ints []int
		s1 := []int{1, 2, 3, 4}
		s2 := ints{5, 6, 7, 8}
		_, err := CombineEvenly(s1, s2)
		so(err, isNil)
	})

	cv("not array types", func() {
		_, err := CombineEvenly(12, []int{})
		so(err, isErr)
		_, err = CombineEvenly([]int{}, 34)
		so(err, isErr)
	})
}
