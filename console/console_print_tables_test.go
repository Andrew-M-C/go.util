package console

import (
	"testing"

	"github.com/Andrew-M-C/go.util/unicode"
)

func testPrintTables(t *testing.T) {
	tab := [][]string{
		{"你好", "世界"},
		{"1", "2", "3", "4", "5", "6"},
		{"https", "pkg", "go", "dev"},
	}

	cv("不带按列对齐", func() {
		t.Log("")
		PrintTables(tab, WithSeparator(" / "))
	})

	cv("按列对齐", func() {
		t.Log("")
		PrintTables(
			tab, WithSeparator("--"),
			WithAlignByCols(unicode.AlignRight, unicode.AlignLeft),
			WithUnifyAlign(unicode.AlignCenter),
		)
	})
}
