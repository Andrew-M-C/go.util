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
	le = convey.ShouldBeLessThanOrEqualTo

	isTrue  = convey.ShouldBeTrue
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
	internal.debugf = t.Logf

	test(t, "slice.go", testSlice)
	test(t, "CombineEvenly", testCombineEvenly)
	test(t, "LCS", testLCS)
	test(t, "binary search", testBinarySearch)
	test(t, "List type", testList)
}

func testSlice(t *testing.T) {
	cv("Equal", func() {
		a := []int{1, 2, 3, 4}
		b := []int{-1, 1, 2, 3}
		so(Equal(a, b), eq, false)

		b = append(b, 4)
		so(Equal(a, b), eq, false)

		b = b[1:]
		so(Equal(a, b), eq, true)
	})

	cv("HaveEqualValues", func() {
		a := []int{0, 1, 0}
		b := []int{0, 1, 0, 1, -1}
		so(HaveEqualValues(a, b), eq, false)
		so(Equal(a, b), eq, false)

		a = append(a, -1)
		so(HaveEqualValues(a, b), eq, true)
		so(Equal(a, b), eq, false)
	})

	cv("Element", func() {
		a := []int{10, 20, 30}

		n, ok := Element(a, 1)
		so(ok, eq, true)
		so(n, eq, a[1])

		n, ok = Element(a, -1)
		so(ok, eq, true)
		so(n, eq, a[2])

		n, ok = Element(a, 3)
		so(ok, eq, false)
		so(n, eq, 0)

		n, ok = Element(a, -3)
		so(ok, eq, true)
		so(n, eq, a[0])

		n, ok = Element(a, -4)
		so(ok, eq, false)
		so(n, eq, 0)
	})

	cv("SetElement", func() {
		a := []int{10, 20, 30}

		ok := SetElement(a, 1, 200)
		so(ok, eq, true)
		so(a[1], eq, 200)

		ok = SetElement(a, -1, -300)
		so(ok, eq, true)
		so(a[2], eq, -300)

		ok = SetElement(a, 3, 333)
		so(ok, eq, false)

		ok = SetElement(a, -3, -333)
		so(ok, eq, true)
		so(a[0], eq, -333)

		ok = SetElement(a, -4, -444)
		so(ok, eq, false)
	})
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
}

var testSliceNum = 0

type testCombineEvenlyType struct {
	ID      string
	Display rune
}

func testCombineEvenlyProcess(t *testing.T, numA, numB int) {
	testSliceNum++
	t.Logf("\n======== Test No.%02d ========", testSliceNum)

	a := testCombineEvenlyTypeSlice('|', numA)
	b := testCombineEvenlyTypeSlice('.', numB)

	printTestCombineEvenlyType(t, a)
	printTestCombineEvenlyType(t, b)

	res := CombineEvenly(a, b)
	printTestCombineEvenlyType(t, res)

	counts := map[rune]int{}
	existedID := map[string]struct{}{}

	for _, r := range res {
		_, exist := existedID[r.ID]
		so(exist, eq, false)

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
