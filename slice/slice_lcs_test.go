package slice

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/fatih/color"
	. "github.com/smartystreets/goconvey/convey"
)

func testLCS(t *testing.T) {
	var x, y []byte

	equal := func(i, j int) bool {
		return x[i] == y[j]
	}

	test := func(x, y []byte, maxLen int) {
		t.Logf("")
		t.Logf("X: %s", x)
		t.Logf("Y: %s", y)

		res := LCS(len(x), len(y), equal)
		r := res.GetRoute()

		desc := func(indexes []int, str []byte) string {
			m := posSliceToMap(indexes)
			buf := bytes.Buffer{}

			for i, b := range str {
				if _, exist := m[i]; exist {
					buf.WriteString(color.GreenString("%c", b))
				} else {
					buf.WriteString(color.RedString("%c", b))
				}
			}
			return buf.String()
		}

		t.Logf("XIndexes: %v", r.XIndexes)
		t.Logf("YIndexes: %v", r.YIndexes)

		t.Logf("X: %s", desc(r.XIndexes, x))
		t.Logf("Y: %s", desc(r.YIndexes, y))

		printRes(t, res)

		so(res.MaxSubLen(), eq, maxLen)
		so(res.MaxSubLen(), ShouldNotBeZeroValue)

		so(len(r.XIndexes), eq, res.MaxSubLen())
		so(len(r.YIndexes), eq, res.MaxSubLen())
	}

	x = []byte("abcbdab")
	y = []byte("bdcaba")
	test(x, y, 4)

	x = []byte("Hello, world!")
	y = []byte("Hello")
	test(x, y, 5)

	x = []byte("acdacacvfsnackxjnhvbdhabxacacakcjdacacadxasxcascadacasadasxsaxasdadxdacaxcac")
	y = []byte("xacacnaiceaacaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa1")
	test(x, y, 28)
}

func posSliceToMap(sl []int) map[int]struct{} {
	ret := map[int]struct{}{}
	for _, i := range sl {
		ret[i] = struct{}{}
	}
	return ret
}

func printRes(t *testing.T, m *LCSMap) {
	buf := bytes.Buffer{}

	for i, line := range m.M {
		// 数字行
		buf.WriteString("\n|   ")
		for j, v := range line {
			buf.WriteString(fmt.Sprintf("%2d", v))

			if m.L[i][j].Right {
				buf.WriteString(" →")
			} else {
				buf.WriteString("  ")
			}
		}
		buf.WriteString("|")

		// 箭头行
		buf.WriteString("\n|   ")
		for _, l := range m.L[i] {
			if l.Down {
				buf.WriteString(" ↓")
			} else {
				buf.WriteString("  ")
			}
			if l.LowerRight {
				buf.WriteString(" ↘")
			} else {
				buf.WriteString("  ")
			}
		}
		buf.WriteString("|")
	}

	t.Logf("\n%s", buf.String())
}
