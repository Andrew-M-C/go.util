package unicode

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"

	rbtree "github.com/emirpasic/gods/trees/redblacktree"
)

func testEastAsianWidth(t *testing.T) {
	cv("简单打印数据情况", func() { testEastAsianWidthPrintInternalData(t) })
	cv("基本功能", func() { testEastAsianWidthBasicFunction(t) })
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
	ioutil.WriteFile(outfile, buff.Bytes(), 0644)
}

func testEastAsianWidthBasicFunction(t *testing.T) {
	lines := []string{
		"0123456789",
		"一二三四五",
		//lint:ignore ST1018 intend to do this to check emoji
		"👦👧👨👩👨‍👩‍👧‍👧",
	}

	t.Log("lines in console:")
	for _, line := range lines {
		t.Log(line)
	}

	for _, line := range lines {
		t.Logf("|%v|", ActualEastAsianWidth(line, 30, WithAlign(AlignLeft), WithBlank("-")))
	}
	for _, line := range lines {
		t.Logf("|%v|", ActualEastAsianWidth(line, 30, WithAlign(AlignCenter), WithBlank("二")))
	}
	for _, line := range lines {
		t.Logf("|%v|", ActualEastAsianWidth(line, 30, WithAlign(AlignRight), WithBlank("=")))
	}
}
