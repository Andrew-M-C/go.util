// Package csv 定义一些基于 csv 的工具函数
package csv

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
)

// ReadCSVStringMaps 读取 CSV 表格数据并转换为 KKV 格式, 其中每一行的第一列为最外层的 map
// key, 第一行的每一列作为内层 map 的 key, 其余的列为 value。最左上角的单元格无意义
func ReadCSVStringMaps[LINE ~string, COL ~string, V ~string](data []byte) (map[LINE]map[COL]V, error) {
	if bytes.HasPrefix(data, []byte{0xFE, 0xFF}) ||
		bytes.HasPrefix(data, []byte{0xFF, 0xFE}) {
		data = data[2:]
	}

	// 首先解析出 csv 表格
	reader := csv.NewReader(bytes.NewReader(data))
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("解析 CSV 数据失败: %w", err)
	}

	if len(records) == 0 {
		return nil, errors.New("CSV 数据为空")
	}

	// 检查第一行是否有足够的列
	if len(records[0]) < 2 {
		return nil, errors.New("CSV 数据格式不正确：第一行至少需要两列")
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

	return result, nil
}
