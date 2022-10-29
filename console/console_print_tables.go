package console

import (
	"os"

	"github.com/Andrew-M-C/go.util/unicode"
)

// PrintTables 格式化地打印表格
func PrintTables(tab [][]string, opts ...Option) error {
	o := mergeOptions(opts)
	if !o.align.enabled {
		return printTablesWithoutAlign(tab, o)
	}

	return printTablesWithAlign(tab, o)
}

func printTablesWithoutAlign(tab [][]string, o *options) error {
	for _, line := range tab {
		for i, s := range line {
			if i > 0 {
				os.Stdout.WriteString(o.separator)
			}
			if _, err := os.Stdout.WriteString(s); err != nil {
				return err
			}
		}
		os.Stdout.WriteString("\n")
	}
	return nil
}

func printTablesWithAlign(tab [][]string, o *options) error {
	maxWidths := []int{}
	getMaxWidth := func(j int) int {
		if j >= len(maxWidths) {
			return -1
		}
		return maxWidths[j]
	}
	setMaxWidth := func(j int, w int) {
		if j >= len(maxWidths) {
			maxWidths = append(maxWidths, w)
		} else {
			maxWidths[j] = w
		}
	}

	for _, line := range tab {
		for j, col := range line {
			w := unicode.EastAsianDisplayWidth(col, unicode.WithTabWidth(2))
			if w > getMaxWidth(j) {
				setMaxWidth(j, w)
			}
		}
	}

	for _, line := range tab {
		for j, col := range line {
			if j > 0 {
				os.Stdout.WriteString(o.separator)
			}
			w := getMaxWidth(j)
			s := unicode.EastAsianStringer(col, w, unicode.WithAlign(o.getAlignAtIndex(j)))
			if _, err := os.Stdout.WriteString(s.String()); err != nil {
				return err
			}
			// TODO:
		}
		os.Stdout.WriteString("\n")
	}
	return nil
}
