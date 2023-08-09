// Package xlsx 提供 excel 工具封装, 仅实现简单填值。更复杂的还是要使用 excelize
package xlsx

import (
	"fmt"
	"os"

	"github.com/Andrew-M-C/go.util/maps"
	"github.com/Andrew-M-C/go.util/slice"
	"github.com/Andrew-M-C/go.util/unsafe"
	"github.com/xuri/excelize/v2"
)

const defaultSheet = "Sheet1"

// Xlsx 提供 Excel 简易操作
type Xlsx struct {
	excel  *excelize.File
	sheets maps.Set[string]
}

// New 新建一个 Xlsx 工具
func New() *Xlsx {
	x := &Xlsx{}
	x.lazyInit()
	return x
}

func (x *Xlsx) lazyInit() {
	if x.excel != nil {
		// 已经初始化, OK
		return
	}

	x.sheets = maps.NewSet[string]()
	x.excel = excelize.NewFile()
}

// Excelize 获取内置的 *excelize.File 对象
func (x *Xlsx) Excelize() *excelize.File {
	x.lazyInit()
	return x.excel
}

// Set 填充值
func (x *Xlsx) Set(sheet string, row, col int, content any) {
	if sheet == "" {
		sheet = defaultSheet
	}

	x.lazyInit()

	if !x.sheets.Has(sheet) {
		_, _ = x.excel.NewSheet(sheet)
		x.sheets.Add(sheet)
	}

	s, ok := content.(string)
	if !ok {
		s = fmt.Sprint(content)
	}

	cell := CellName(row, col)
	_ = x.excel.SetCellStr(sheet, cell, s)
}

// Save 保存至文件
func (x *Xlsx) Save(filePath string) error {
	x.lazyInit()
	if len(x.sheets) > 0 {
		if !x.sheets.Has(defaultSheet) {
			_ = x.excel.DeleteSheet(defaultSheet)
		}
	}

	_ = os.Remove(filePath)
	return x.excel.SaveAs(filePath)
}

// CellName 按行、列（均从 0 开始）对应的单元格名称
func CellName(row, col int) string {
	return fmt.Sprintf("%s%d", formatCol(col), row+1)
}

func formatCol(c int) string {
	if c < 26 {
		return string('A' + byte(c))
	}

	res := []byte{}
	for c >= 26 {
		remain := c % 26
		res = append(res, 'A'+byte(remain))
		c /= 26
	}

	res = append(res, 'A'+byte(c)-1)
	slice.Reverse(res)

	return unsafe.BtoS(res)
}
