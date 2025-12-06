// Package csv 定义一些基于 csv 的工具函数
package csv

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"slices"
)

// WriteCSVStringMaps 将 map[LINE]map[COL]V 写入为 CSV 字节流。
// 输出的 CSV 格式与 ReadCSVStringMaps 兼容，即第一行为列标题（第一个单元格为空），
// 后续每行第一列为行键，其余列为对应的值。
//
// 参数说明：
//   - data: 要写入的数据，格式为 map[行键]map[列键]值
//   - columnSequences: 指定列的优先输出顺序。
//     如果为 nil，则所有列按字母顺序排序输出；
//     如果指定了列顺序，则优先按指定顺序输出这些列，数据中存在但未在 columnSequences 中指定的列
//     会按字母顺序排在后面输出。columnSequences 中不存在于数据中的列会被忽略。
//
// 行的输出顺序始终按字母排序以保证确定性输出。
func WriteCSVStringMaps[LINE ~string, COL ~string, V ~string](
	data map[LINE]map[COL]V, columnSequences []COL,
) ([]byte, error) {
	if len(data) == 0 {
		return nil, errors.New("数据为空")
	}

	buff := &bytes.Buffer{}

	// 写入 BOM 头, UTF-8 BOM (0xEF 0xBB 0xBF)
	// 使用 UTF-8 BOM 而非 UTF-16，因为 Go 的 csv 包默认输出 UTF-8 编码，
	// 且 UTF-8 BOM 在大多数程序（包括 Excel）中都能正确识别。
	buff.Write([]byte{0xEF, 0xBB, 0xBF})

	// 收集所有列名
	columnSet := make(map[COL]struct{})
	for _, row := range data {
		for col := range row {
			columnSet[col] = struct{}{}
		}
	}

	if len(columnSet) == 0 {
		return nil, errors.New("数据中没有有效的列")
	}

	// 确定最终的列顺序
	var columns []COL
	usedColumns := make(map[COL]struct{}) // 记录已经添加的列，避免重复

	if len(columnSequences) > 0 {
		// 1. 先按指定顺序添加存在于数据中的列
		for _, col := range columnSequences {
			if _, exists := columnSet[col]; exists {
				if _, used := usedColumns[col]; !used {
					columns = append(columns, col)
					usedColumns[col] = struct{}{}
				}
			}
		}
	}

	// 2. 收集未在 columnSequences 中指定的列，按字母序排序后追加
	var remainingColumns []COL
	for col := range columnSet {
		if _, used := usedColumns[col]; !used {
			remainingColumns = append(remainingColumns, col)
		}
	}
	slices.SortFunc(remainingColumns, func(a, b COL) int {
		if string(a) < string(b) {
			return -1
		} else if string(a) > string(b) {
			return 1
		}
		return 0
	})
	columns = append(columns, remainingColumns...)

	// 收集所有行名并排序（保证确定性输出）
	lines := make([]LINE, 0, len(data))
	for line := range data {
		lines = append(lines, line)
	}
	slices.SortFunc(lines, func(a, b LINE) int {
		if string(a) < string(b) {
			return -1
		} else if string(a) > string(b) {
			return 1
		}
		return 0
	})

	// 使用 csv.Writer 写入数据
	writer := csv.NewWriter(buff)

	// 写入标题行（第一个单元格为空，作为占位符）
	header := make([]string, 0, len(columns)+1)
	header = append(header, "") // 第一个单元格为空（与 ReadCSVStringMaps 格式对应）
	for _, col := range columns {
		header = append(header, string(col))
	}
	if err := writer.Write(header); err != nil {
		return nil, fmt.Errorf("写入标题行失败: %w", err)
	}

	// 写入数据行
	for _, line := range lines {
		row := make([]string, 0, len(columns)+1)
		row = append(row, string(line)) // 第一列为行键

		rowData := data[line]
		for _, col := range columns {
			if val, exists := rowData[col]; exists {
				row = append(row, string(val))
			} else {
				row = append(row, "") // 空值
			}
		}

		if err := writer.Write(row); err != nil {
			return nil, fmt.Errorf("写入数据行 %s 失败: %w", string(line), err)
		}
	}

	// 刷新缓冲区
	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("写入 CSV 数据失败: %w", err)
	}

	return buff.Bytes(), nil
}

// ReadCSVStringMaps 读取 CSV 表格数据并转换为 KKV 格式, 其中每一行的第一列为最外层的 map
// key, 第一行的每一列作为内层 map 的 key, 其余的列为 value。最左上角的单元格无意义
func ReadCSVStringMaps[LINE ~string, COL ~string, V ~string](
	data []byte,
) (map[LINE]map[COL]V, []COL, error) {
	// 处理 BOM (Byte Order Mark, 字节顺序标记)
	// BOM 是 Unicode 文本文件开头的特殊字节序列, 用于标识文件的编码格式和字节顺序。
	// 许多文本编辑器 (如 Windows 记事本、Excel 导出的 CSV) 会自动添加 BOM 头。
	//
	// UTF-8 BOM (3 字节):
	//   - 0xEF, 0xBB, 0xBF: UTF-8 编码标识
	//   - UTF-8 本身不需要 BOM (因为没有字节顺序问题), 但某些软件 (如 Excel) 使用它来识别 UTF-8 编码
	//   - Go 的 csv.Reader 可以正常处理 UTF-8 编码, 只需跳过 BOM 即可
	//
	// UTF-16 BOM (2 字节):
	//   - 0xFE, 0xFF: UTF-16 Big-Endian (大端序), 高位字节在前
	//   - 0xFF, 0xFE: UTF-16 Little-Endian (小端序), 低位字节在前
	//   - 注意: 由于 Go 的 csv.Reader 期望 UTF-8 编码, 如果原始文件确实是 UTF-16 编码,
	//     仅跳过 BOM 可能不足以正确解析, 可能需要进行编码转换。
	//
	// BOM 处理优先级: UTF-8 BOM 优先 (3 字节), 然后是 UTF-16 BOM (2 字节)
	if bytes.HasPrefix(data, []byte{0xEF, 0xBB, 0xBF}) {
		// UTF-8 BOM
		data = data[3:]
	} else if bytes.HasPrefix(data, []byte{0xFE, 0xFF}) ||
		bytes.HasPrefix(data, []byte{0xFF, 0xFE}) {
		// UTF-16 BOM
		data = data[2:]
	}

	// 首先解析出 csv 表格
	reader := csv.NewReader(bytes.NewReader(data))
	records, err := reader.ReadAll()
	if err != nil {
		return nil, nil, fmt.Errorf("解析 CSV 数据失败: %w", err)
	}

	if len(records) == 0 {
		return nil, nil, errors.New("CSV 数据为空")
	}

	// 检查第一行是否有足够的列
	if len(records[0]) < 2 {
		return nil, nil, errors.New("CSV 数据格式不正确：第一行至少需要两列")
	}

	// 然后解析出两层 key
	// 第一行的列标题（跳过第一个单元格）
	columnKeys := make([]COL, 0, len(records[0])-1)
	for i := 1; i < len(records[0]); i++ {
		columnKeys = append(columnKeys, COL(records[0][i]))
	}

	// 最后整合到 map 中
	result := make(map[LINE]map[COL]V)

	// 遍历数据行（跳过第一行标题行）
	for i := 1; i < len(records); i++ {
		record := records[i]

		// 检查行是否有足够的列
		if len(record) < 2 {
			continue // 跳过不完整的行
		}

		// 获取行键（第一列）
		lineKey := LINE(record[0])

		// 初始化内层 map
		if result[lineKey] == nil {
			result[lineKey] = make(map[COL]V)
		}

		// 填充数据（从第二列开始）
		for j := 1; j < len(record) && j-1 < len(columnKeys); j++ {
			columnKey := columnKeys[j-1]
			value := V(record[j])
			if value == "" {
				continue
			}
			result[lineKey][columnKey] = value
		}
	}

	return result, columnKeys, nil
}
