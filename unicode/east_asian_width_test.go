package unicode

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	rbtree "github.com/emirpasic/gods/trees/redblacktree"
)

func testEastAsianWidth(t *testing.T) {
	cv("简单打印数据情况", func() { testEastAsianWidthPrintInternalData(t) })
	cv("基本功能", func() { testEastAsianWidthBasicFunction(t) })
	cv("测试 EastAsianDisplayWidth", func() { testEastAsianDisplayWidth(t) })
	cv("测试 CutSetWithMaxDisplayWidth", func() { testCutSetWithMaxDisplayWidth(t) })
}

func testEastAsianWidthPrintInternalData(t *testing.T) {
	chars := rbtree.NewWith(func(a, b interface{}) int {
		aa, _ := a.(rune)
		bb, _ := b.(rune)
		if aa < bb {
			return -1
		} else if aa > bb {
			return 1
		}
		return 0
	})

	halfCnt, fullCnt := 0, 0
	for r, w := range internal.eastAsianWidth {
		switch w {
		default:
			t.Errorf("unrecognized witdh: %v", w)
		case 1:
			halfCnt++
		case 2:
			fullCnt++
		}

		chars.Put(r, nil)
	}

	so(halfCnt+fullCnt, eq, len(internal.eastAsianWidth))
	t.Logf("全角字符 %d 个, 半角字符 %d 个, 总共 %d 个字符", fullCnt, halfCnt, len(internal.eastAsianWidth))

	buff := bytes.NewBuffer(make([]byte, 0, len(internal.eastAsianWidth)*20))
	for it := chars.Iterator(); it.Next(); {
		key := it.Key()
		r, _ := key.(rune)
		if r <= 127 {
			continue
		}

		var line string
		if internal.eastAsianWidth[r] == 2 {
			line = fmt.Sprintf("%06x - %c\n", r, r)
		} else {
			line = fmt.Sprintf("%06x -  %c\n", r, r)
		}
		buff.WriteString(line)
	}

	// t.Logf("完整的汉字列表:\n%s", buff.String())
	const outfile = "./.all_runes.txt"
	_ = os.WriteFile(outfile, buff.Bytes(), 0644)
}

func testEastAsianWidthBasicFunction(t *testing.T) {
	lines := []string{
		"0123456789",
		"一二三四五",
		"👦👧👨👩👨‍👩‍👧‍👧",
	}

	t.Log("lines in console:")
	for _, line := range lines {
		t.Log(line)
	}

	for _, line := range lines {
		t.Logf("|%v|", EastAsianStringer(line, 30, WithAlign(AlignLeft), WithBlank("-")))
	}
	for _, line := range lines {
		t.Logf("|%v|", EastAsianStringer(line, 30, WithAlign(AlignCenter), WithBlank("二")))
	}
	for _, line := range lines {
		t.Logf("|%v|", EastAsianStringer(line, 30, WithAlign(AlignRight), WithBlank("=")))
	}
}

func testEastAsianDisplayWidth(*testing.T) {
	s := "一二三四五"
	w := EastAsianDisplayWidth(s)
	so(w, eq, 10)

	s = "\t一二三四五"
	w = EastAsianDisplayWidth(s, WithTabWidth(4))
	so(w, eq, 14)
}

func testCutSetWithMaxDisplayWidth(*testing.T) {
	s := "一二三四五六七八九十😊"

	res := CutSetWithMaxDisplayWidth(s, 8)
	so(res, eq, "一二三四")

	res = CutSetWithMaxDisplayWidth(s, 9)
	so(res, eq, "一二三四")

	res = CutSetWithMaxDisplayWidth(s, 10)
	so(res, eq, "一二三四五")

	res = CutSetWithMaxDisplayWidth(s, 20)
	so(res, eq, "一二三四五六七八九十")

	res = CutSetWithMaxDisplayWidth(s, 21)
	so(res, eq, "一二三四五六七八九十")

	res = CutSetWithMaxDisplayWidth(s, 22)
	so(res, eq, "一二三四五六七八九十😊")
}
