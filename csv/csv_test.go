package csv_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Andrew-M-C/go.util/csv"
	"github.com/smartystreets/goconvey/convey"
)

var (
	cv = convey.Convey
	so = convey.So
	eq = convey.ShouldEqual

	isNil   = convey.ShouldBeNil
	notNil  = convey.ShouldNotBeNil
	isFalse = convey.ShouldBeFalse
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

// readTestData 读取测试数据文件
func readTestData(filename string) ([]byte, error) {
	return os.ReadFile(filepath.Join("testdata", filename))
}

// ========== TestReadCSVStringMaps 测试 ReadCSVStringMaps 函数 ==========

func TestReadCSVStringMaps_Normal(t *testing.T) {
	cv("正常读取 CSV 文件", t, func() {
		data, err := readTestData("normal.csv")
		so(err, isNil)

		result, _, err := csv.ReadCSVStringMaps[string, string, string](data)
		so(err, isNil)
		so(result, notNil)

		// 验证数据结构
		so(len(result), eq, 3) // 3 行数据

		// 验证 user1
		so(result["user1"], notNil)
		so(result["user1"]["name"], eq, "Alice")
		so(result["user1"]["age"], eq, "25")
		so(result["user1"]["city"], eq, "Beijing")

		// 验证 user2
		so(result["user2"], notNil)
		so(result["user2"]["name"], eq, "Bob")
		so(result["user2"]["age"], eq, "30")
		so(result["user2"]["city"], eq, "Shanghai")

		// 验证 user3
		so(result["user3"], notNil)
		so(result["user3"]["name"], eq, "Charlie")
		so(result["user3"]["age"], eq, "35")
		so(result["user3"]["city"], eq, "Guangzhou")
	})
}

func TestReadCSVStringMaps_WithEmptyValues(t *testing.T) {
	cv("读取带空值的 CSV 文件", t, func() {
		data, err := readTestData("with_empty_values.csv")
		so(err, isNil)

		result, _, err := csv.ReadCSVStringMaps[string, string, string](data)
		so(err, isNil)
		so(result, notNil)

		// row1: value1,,value3 - col2 为空应被跳过
		so(result["row1"], notNil)
		so(result["row1"]["col1"], eq, "value1")
		_, hasCol2 := result["row1"]["col2"]
		so(hasCol2, isFalse) // 空值应该被跳过
		so(result["row1"]["col3"], eq, "value3")

		// row2: ,value2, - col1 和 col3 为空应被跳过
		so(result["row2"], notNil)
		_, hasCol1 := result["row2"]["col1"]
		so(hasCol1, isFalse)
		so(result["row2"]["col2"], eq, "value2")
		_, hasCol3 := result["row2"]["col3"]
		so(hasCol3, isFalse)

		// row3: a,b,c - 全部有值
		so(result["row3"], notNil)
		so(result["row3"]["col1"], eq, "a")
		so(result["row3"]["col2"], eq, "b")
		so(result["row3"]["col3"], eq, "c")
	})
}

func TestReadCSVStringMaps_DuplicateKeys(t *testing.T) {
	cv("读取带重复行键的 CSV 文件", t, func() {
		data, err := readTestData("duplicate_keys.csv")
		so(err, isNil)

		result, _, err := csv.ReadCSVStringMaps[string, string, string](data)
		so(err, isNil)
		so(result, notNil)

		// rowA 出现两次，后面的应该覆盖前面的
		so(result["rowA"], notNil)
		so(result["rowA"]["key1"], eq, "second1") // 应该是第二次的值
		so(result["rowA"]["key2"], eq, "second2")

		// rowB 正常
		so(result["rowB"], notNil)
		so(result["rowB"]["key1"], eq, "valueB1")
		so(result["rowB"]["key2"], eq, "valueB2")
	})
}

func TestReadCSVStringMaps_ValidRows(t *testing.T) {
	cv("读取完整行的 CSV 文件", t, func() {
		data, err := readTestData("incomplete_rows.csv")
		so(err, isNil)

		result, _, err := csv.ReadCSVStringMaps[string, string, string](data)
		so(err, isNil)
		so(result, notNil)

		// row1 完整
		so(result["row1"], notNil)
		so(result["row1"]["col1"], eq, "v1")
		so(result["row1"]["col2"], eq, "v2")
		so(result["row1"]["col3"], eq, "v3")

		// row3 完整
		so(result["row3"], notNil)
		so(result["row3"]["col1"], eq, "v31")
		so(result["row3"]["col2"], eq, "v32")
		so(result["row3"]["col3"], eq, "v33")
	})
}

func TestReadCSVStringMaps_IncompleteRowsError(t *testing.T) {
	cv("列数不一致的行会导致解析错误", t, func() {
		// Go 标准库的 csv 解析器默认要求每行列数一致
		// 如果有不完整的行，会返回错误
		data := []byte("ignore,col1,col2,col3\nrow1,v1,v2,v3\nshort\nrow3,v31,v32,v33\n")

		result, _, err := csv.ReadCSVStringMaps[string, string, string](data)
		so(err, notNil) // 应该返回解析错误
		so(result, isNil)
	})
}

func TestReadCSVStringMaps_SingleColumn(t *testing.T) {
	cv("读取单列的 CSV 文件应返回错误", t, func() {
		data, err := readTestData("single_column.csv")
		so(err, isNil)

		result, _, err := csv.ReadCSVStringMaps[string, string, string](data)
		so(err, notNil) // 应该返回错误
		so(result, isNil)
		so(err.Error(), eq, "CSV 数据格式不正确：第一行至少需要两列")
	})
}

func TestReadCSVStringMaps_Empty(t *testing.T) {
	cv("读取空的 CSV 文件应返回错误", t, func() {
		data, err := readTestData("empty.csv")
		so(err, isNil)

		result, _, err := csv.ReadCSVStringMaps[string, string, string](data)
		so(err, notNil) // 应该返回错误
		so(result, isNil)
		so(err.Error(), eq, "CSV 数据为空")
	})
}

func TestReadCSVStringMaps_Minimal(t *testing.T) {
	cv("读取最小有效的 CSV 文件", t, func() {
		data, err := readTestData("minimal.csv")
		so(err, isNil)

		result, _, err := csv.ReadCSVStringMaps[string, string, string](data)
		so(err, isNil)
		so(result, notNil)

		so(len(result), eq, 1)
		so(result["row1"], notNil)
		so(result["row1"]["col1"], eq, "value1")
	})
}

func TestReadCSVStringMaps_Chinese(t *testing.T) {
	cv("读取中文内容的 CSV 文件", t, func() {
		data, err := readTestData("chinese.csv")
		so(err, isNil)

		result, _, err := csv.ReadCSVStringMaps[string, string, string](data)
		so(err, isNil)
		so(result, notNil)

		so(len(result), eq, 3)

		// 验证中文内容
		so(result["用户1"], notNil)
		so(result["用户1"]["姓名"], eq, "张三")
		so(result["用户1"]["年龄"], eq, "25")
		so(result["用户1"]["城市"], eq, "北京")

		so(result["用户2"], notNil)
		so(result["用户2"]["姓名"], eq, "李四")
		so(result["用户2"]["年龄"], eq, "30")
		so(result["用户2"]["城市"], eq, "上海")

		so(result["用户3"], notNil)
		so(result["用户3"]["姓名"], eq, "王五")
		so(result["用户3"]["年龄"], eq, "35")
		so(result["用户3"]["城市"], eq, "广州")
	})
}

func TestReadCSVStringMaps_WithBOM(t *testing.T) {
	cv("读取带 BOM 头的 CSV 数据", t, func() {
		// 准备正常的 CSV 数据
		normalCSV := []byte("ignore,col1\nrow1,value1\n")

		// 测试 UTF-16 BE BOM (0xFE 0xFF)
		convey.Convey("测试 UTF-16 BE BOM", func() {
			dataWithBOM := append([]byte{0xFE, 0xFF}, normalCSV...)
			result, _, err := csv.ReadCSVStringMaps[string, string, string](dataWithBOM)
			so(err, isNil)
			so(result, notNil)
			so(result["row1"]["col1"], eq, "value1")
		})

		// 测试 UTF-16 LE BOM (0xFF 0xFE)
		convey.Convey("测试 UTF-16 LE BOM", func() {
			dataWithBOM := append([]byte{0xFF, 0xFE}, normalCSV...)
			result, _, err := csv.ReadCSVStringMaps[string, string, string](dataWithBOM)
			so(err, isNil)
			so(result, notNil)
			so(result["row1"]["col1"], eq, "value1")
		})
	})
}

// UserID 自定义用户 ID 类型
type UserID string

// ColumnName 自定义列名类型
type ColumnName string

// CellValue 自定义单元格值类型
type CellValue string

func TestReadCSVStringMaps_CustomTypes(t *testing.T) {
	cv("使用自定义类型读取 CSV", t, func() {
		data, err := readTestData("normal.csv")
		so(err, isNil)

		// 使用自定义类型
		result, _, err := csv.ReadCSVStringMaps[UserID, ColumnName, CellValue](data)
		so(err, isNil)
		so(result, notNil)

		// 验证使用自定义类型可以正常访问
		so(result[UserID("user1")], notNil)
		so(result[UserID("user1")][ColumnName("name")], eq, CellValue("Alice"))
		so(result[UserID("user1")][ColumnName("age")], eq, CellValue("25"))
		so(result[UserID("user1")][ColumnName("city")], eq, CellValue("Beijing"))

		// 验证类型转换
		var userName CellValue = result[UserID("user2")][ColumnName("name")]
		so(string(userName), eq, "Bob")
	})
}

func TestReadCSVStringMaps_MalformedCSV(t *testing.T) {
	cv("测试格式错误的 CSV 数据", t, func() {
		// 测试带有未闭合引号的 CSV 数据
		malformedData := []byte(`ignore,col1
row1,"unclosed quote`)

		result, _, err := csv.ReadCSVStringMaps[string, string, string](malformedData)
		so(err, notNil) // 应该返回解析错误
		so(result, isNil)
	})
}

func TestReadCSVStringMaps_HeaderOnlyFile(t *testing.T) {
	cv("测试只有标题行的 CSV 文件", t, func() {
		// 只有标题行，没有数据行
		headerOnlyData := []byte("ignore,col1,col2\n")

		result, _, err := csv.ReadCSVStringMaps[string, string, string](headerOnlyData)
		so(err, isNil)
		so(result, notNil)
		so(len(result), eq, 0) // 应该返回空 map
	})
}

func TestReadCSVStringMaps_ColumnCountMismatch(t *testing.T) {
	cv("测试数据行与标题行列数不一致的情况", t, func() {
		// Go 标准库 csv.Reader 默认要求所有行列数一致
		// 如果数据行列数与标题行不一致，会返回解析错误

		convey.Convey("数据行比标题行多列应返回错误", func() {
			data := []byte("ignore,col1,col2\nrow1,v1,v2,v3,v4\n")
			result, _, err := csv.ReadCSVStringMaps[string, string, string](data)
			so(err, notNil) // 应该返回解析错误
			so(result, isNil)
		})

		convey.Convey("数据行比标题行少列应返回错误", func() {
			data := []byte("ignore,col1,col2,col3\nrow1,v1\n")
			result, _, err := csv.ReadCSVStringMaps[string, string, string](data)
			so(err, notNil) // 应该返回解析错误
			so(result, isNil)
		})
	})
}

func TestReadCSVStringMaps_SpecialCharacters(t *testing.T) {
	cv("测试包含特殊字符的 CSV 数据", t, func() {
		convey.Convey("包含逗号的引号字段", func() {
			data := []byte(`ignore,col1,col2
row1,"hello, world",value2
`)
			result, _, err := csv.ReadCSVStringMaps[string, string, string](data)
			so(err, isNil)
			so(result["row1"]["col1"], eq, "hello, world")
		})

		convey.Convey("包含换行的引号字段", func() {
			data := []byte(`ignore,col1,col2
row1,"line1
line2",value2
`)
			result, _, err := csv.ReadCSVStringMaps[string, string, string](data)
			so(err, isNil)
			so(result["row1"]["col1"], eq, "line1\nline2")
		})

		convey.Convey("包含双引号的引号字段", func() {
			data := []byte(`ignore,col1,col2
row1,"say ""hello""",value2
`)
			result, _, err := csv.ReadCSVStringMaps[string, string, string](data)
			so(err, isNil)
			so(result["row1"]["col1"], eq, `say "hello"`)
		})
	})
}

func TestReadCSVStringMaps_EmptyRowKey(t *testing.T) {
	cv("测试行键为空字符串的情况", t, func() {
		data := []byte("ignore,col1,col2\n,v1,v2\nrow2,v21,v22\n")

		result, _, err := csv.ReadCSVStringMaps[string, string, string](data)
		so(err, isNil)
		so(result, notNil)

		// 空字符串作为 key 也是有效的
		so(result[""], notNil)
		so(result[""]["col1"], eq, "v1")
		so(result[""]["col2"], eq, "v2")

		so(result["row2"]["col1"], eq, "v21")
		so(result["row2"]["col2"], eq, "v22")
	})
}

func TestReadCSVStringMaps_WindowsLineEndings(t *testing.T) {
	cv("测试 Windows 风格换行符 (CRLF)", t, func() {
		// Windows 风格: \r\n
		data := []byte("ignore,col1,col2\r\nrow1,v1,v2\r\nrow2,v21,v22\r\n")

		result, _, err := csv.ReadCSVStringMaps[string, string, string](data)
		so(err, isNil)
		so(result, notNil)
		so(len(result), eq, 2)
		so(result["row1"]["col1"], eq, "v1")
		so(result["row2"]["col2"], eq, "v22")
	})
}

func TestReadCSVStringMaps_MultipleDataRows(t *testing.T) {
	cv("测试多行数据的情况", t, func() {
		// 测试较多行数据
		data := []byte(`ignore,col1,col2,col3
row1,a1,b1,c1
row2,a2,b2,c2
row3,a3,b3,c3
row4,a4,b4,c4
row5,a5,b5,c5
`)
		result, _, err := csv.ReadCSVStringMaps[string, string, string](data)
		so(err, isNil)
		so(result, notNil)
		so(len(result), eq, 5)

		// 验证每行数据
		for i := 1; i <= 5; i++ {
			rowKey := "row" + string(rune('0'+i))
			so(result[rowKey], notNil)
			so(result[rowKey]["col1"], eq, "a"+string(rune('0'+i)))
			so(result[rowKey]["col2"], eq, "b"+string(rune('0'+i)))
			so(result[rowKey]["col3"], eq, "c"+string(rune('0'+i)))
		}
	})
}

func TestReadCSVStringMaps_UTF8BOM(t *testing.T) {
	cv("测试 UTF-8 BOM 的 CSV 数据", t, func() {
		convey.Convey("使用内存中构造的带 UTF-8 BOM 数据", func() {
			// UTF-8 BOM: 0xEF 0xBB 0xBF
			normalCSV := []byte("ignore,col1\nrow1,value1\n")
			dataWithBOM := append([]byte{0xEF, 0xBB, 0xBF}, normalCSV...)

			result, _, err := csv.ReadCSVStringMaps[string, string, string](dataWithBOM)
			so(err, isNil)
			so(result, notNil)
			// UTF-8 BOM 应该被正确处理，列名不应该带有 BOM 前缀
			so(result["row1"]["col1"], eq, "value1")
		})

		convey.Convey("使用 WriteCSVStringMaps 生成的带 BOM 中文 CSV 文件", func() {
			// 读取由 WriteCSVStringMaps 函数生成的带 UTF-8 BOM 的中文 CSV 文件
			data, err := readTestData("test_write_csv_string_maps_chinese.csv")
			so(err, isNil)

			// 验证文件确实包含 UTF-8 BOM
			so(len(data) >= 3, eq, true)
			so(data[0], eq, byte(0xEF))
			so(data[1], eq, byte(0xBB))
			so(data[2], eq, byte(0xBF))

			// 解析 CSV 数据
			result, columnOrder, err := csv.ReadCSVStringMaps[string, string, string](data)
			so(err, isNil)
			so(result, notNil)
			so(len(result), eq, 2) // 用户1, 用户2

			// 验证列顺序 (按字母序排序: 城市, 姓名, 年龄)
			so(len(columnOrder), eq, 3)
			so(columnOrder[0], eq, "城市")
			so(columnOrder[1], eq, "姓名")
			so(columnOrder[2], eq, "年龄")

			// 验证中文数据被正确解析
			so(result["用户1"], notNil)
			so(result["用户1"]["姓名"], eq, "张三")
			so(result["用户1"]["年龄"], eq, "25")
			so(result["用户1"]["城市"], eq, "北京")

			so(result["用户2"], notNil)
			so(result["用户2"]["姓名"], eq, "李四")
			so(result["用户2"]["年龄"], eq, "30")
			so(result["用户2"]["城市"], eq, "上海")
		})

		convey.Convey("验证 Write 和 Read 的往返一致性", func() {
			// 原始数据
			original := map[string]map[string]string{
				"用户1": {"姓名": "张三", "年龄": "25", "城市": "北京"},
				"用户2": {"姓名": "李四", "年龄": "30", "城市": "上海"},
			}

			// 写入
			written, err := csv.WriteCSVStringMaps(original, nil)
			so(err, isNil)

			// 验证包含 UTF-8 BOM
			so(written[0], eq, byte(0xEF))
			so(written[1], eq, byte(0xBB))
			so(written[2], eq, byte(0xBF))

			// 读取
			readBack, _, err := csv.ReadCSVStringMaps[string, string, string](written)
			so(err, isNil)

			// 验证数据一致性
			so(len(readBack), eq, len(original))
			for lineKey, lineData := range original {
				so(readBack[lineKey], notNil)
				for colKey, value := range lineData {
					so(readBack[lineKey][colKey], eq, value)
				}
			}
		})
	})
}

func TestReadCSVStringMaps_EmptyColumnHeader(t *testing.T) {
	cv("测试列头为空字符串的情况", t, func() {
		// 列头为空字符串
		data := []byte("ignore,,col2\nrow1,v1,v2\n")

		result, _, err := csv.ReadCSVStringMaps[string, string, string](data)
		so(err, isNil)
		so(result, notNil)

		// 空字符串作为列头也是有效的
		so(result["row1"], notNil)
		so(result["row1"][""], eq, "v1")
		so(result["row1"]["col2"], eq, "v2")
	})
}

func TestReadCSVStringMaps_ColumnOrder(t *testing.T) {
	cv("测试返回的列顺序", t, func() {
		// 测试列顺序是否正确返回
		data := []byte("ignore,col_z,col_a,col_m,col_b\nrow1,v1,v2,v3,v4\n")

		result, columnOrder, err := csv.ReadCSVStringMaps[string, string, string](data)
		so(err, isNil)
		so(result, notNil)
		so(columnOrder, notNil)

		// 验证列顺序与 CSV 中定义的顺序一致（不是按字母排序）
		so(len(columnOrder), eq, 4)
		so(columnOrder[0], eq, "col_z")
		so(columnOrder[1], eq, "col_a")
		so(columnOrder[2], eq, "col_m")
		so(columnOrder[3], eq, "col_b")

		// 验证数据也能正常读取
		so(result["row1"]["col_z"], eq, "v1")
		so(result["row1"]["col_a"], eq, "v2")
		so(result["row1"]["col_m"], eq, "v3")
		so(result["row1"]["col_b"], eq, "v4")
	})
}

// ========== TestWriteCSVStringMaps 测试 WriteCSVStringMaps 函数 ==========

func TestWriteCSVStringMaps_Normal(t *testing.T) {
	cv("正常写入 CSV", t, func() {
		data := map[string]map[string]string{
			"user1": {"name": "Alice", "age": "25", "city": "Beijing"},
			"user2": {"name": "Bob", "age": "30", "city": "Shanghai"},
			"user3": {"name": "Charlie", "age": "35", "city": "Guangzhou"},
		}

		result, err := csv.WriteCSVStringMaps(data, nil)
		so(err, isNil)
		so(result, notNil)

		// 验证输出包含 UTF-8 BOM
		so(len(result) >= 3, eq, true)
		so(result[0], eq, byte(0xEF))
		so(result[1], eq, byte(0xBB))
		so(result[2], eq, byte(0xBF))

		// 验证可以被 ReadCSVStringMaps 正确读取
		readBack, _, err := csv.ReadCSVStringMaps[string, string, string](result)
		so(err, isNil)
		so(readBack, notNil)
		so(len(readBack), eq, 3)

		// 验证数据一致性
		so(readBack["user1"]["name"], eq, "Alice")
		so(readBack["user1"]["age"], eq, "25")
		so(readBack["user1"]["city"], eq, "Beijing")
		so(readBack["user2"]["name"], eq, "Bob")
		so(readBack["user3"]["name"], eq, "Charlie")

		// 写入文件备查
		err = os.WriteFile("./testdata/test_write_csv_string_maps_normal.csv", result, 0644)
		so(err, isNil)
	})
}

func TestWriteCSVStringMaps_Empty(t *testing.T) {
	cv("写入空数据应返回错误", t, func() {
		data := map[string]map[string]string{}

		result, err := csv.WriteCSVStringMaps(data, nil)
		so(err, notNil)
		so(result, isNil)
		so(err.Error(), eq, "数据为空")
	})
}

func TestWriteCSVStringMaps_EmptyRows(t *testing.T) {
	cv("写入全部为空行的数据应返回错误", t, func() {
		data := map[string]map[string]string{
			"row1": {},
			"row2": {},
		}

		result, err := csv.WriteCSVStringMaps(data, nil)
		so(err, notNil)
		so(result, isNil)
		so(err.Error(), eq, "数据中没有有效的列")
	})
}

func TestWriteCSVStringMaps_Chinese(t *testing.T) {
	cv("写入中文内容", t, func() {
		data := map[string]map[string]string{
			"用户1": {"姓名": "张三", "年龄": "25", "城市": "北京"},
			"用户2": {"姓名": "李四", "年龄": "30", "城市": "上海"},
		}

		result, err := csv.WriteCSVStringMaps(data, nil)
		so(err, isNil)
		so(result, notNil)

		// 验证可以被正确读取
		readBack, _, err := csv.ReadCSVStringMaps[string, string, string](result)
		so(err, isNil)
		so(readBack["用户1"]["姓名"], eq, "张三")
		so(readBack["用户2"]["城市"], eq, "上海")

		// 写入文件备查
		err = os.WriteFile("./testdata/test_write_csv_string_maps_chinese.csv", result, 0644)
		so(err, isNil)
	})
}

func TestWriteCSVStringMaps_SparseData(t *testing.T) {
	cv("写入稀疏数据（不同行有不同的列）", t, func() {
		data := map[string]map[string]string{
			"row1": {"col1": "v1", "col2": "v2"},
			"row2": {"col2": "v22", "col3": "v23"},
			"row3": {"col1": "v31", "col3": "v33"},
		}

		result, err := csv.WriteCSVStringMaps(data, nil)
		so(err, isNil)
		so(result, notNil)

		// 验证可以被正确读取
		readBack, columnOrder, err := csv.ReadCSVStringMaps[string, string, string](result)
		so(err, isNil)
		so(readBack, notNil)

		// 验证所有列都被收集
		so(len(columnOrder), eq, 3) // col1, col2, col3

		// 验证数据一致性
		so(readBack["row1"]["col1"], eq, "v1")
		so(readBack["row1"]["col2"], eq, "v2")
		_, hasCol3 := readBack["row1"]["col3"]
		so(hasCol3, isFalse) // row1 没有 col3

		so(readBack["row2"]["col2"], eq, "v22")
		so(readBack["row2"]["col3"], eq, "v23")

		so(readBack["row3"]["col1"], eq, "v31")
		so(readBack["row3"]["col3"], eq, "v33")
	})
}

func TestWriteCSVStringMaps_Deterministic(t *testing.T) {
	cv("多次写入应产生相同的输出", t, func() {
		data := map[string]map[string]string{
			"z_row": {"z_col": "v1", "a_col": "v2"},
			"a_row": {"m_col": "v3", "b_col": "v4"},
			"m_row": {"a_col": "v5", "z_col": "v6"},
		}

		// 多次调用，结果应该完全一致
		result1, err := csv.WriteCSVStringMaps(data, nil)
		so(err, isNil)

		result2, err := csv.WriteCSVStringMaps(data, nil)
		so(err, isNil)

		result3, err := csv.WriteCSVStringMaps(data, nil)
		so(err, isNil)

		so(string(result1), eq, string(result2))
		so(string(result2), eq, string(result3))
	})
}

func TestWriteCSVStringMaps_CustomTypes(t *testing.T) {
	cv("使用自定义类型写入 CSV", t, func() {
		data := map[UserID]map[ColumnName]CellValue{
			UserID("user1"): {ColumnName("name"): CellValue("Alice")},
			UserID("user2"): {ColumnName("name"): CellValue("Bob")},
		}

		result, err := csv.WriteCSVStringMaps(data, nil)
		so(err, isNil)
		so(result, notNil)

		// 验证可以被正确读取（使用相同的自定义类型）
		readBack, _, err := csv.ReadCSVStringMaps[UserID, ColumnName, CellValue](result)
		so(err, isNil)
		so(readBack[UserID("user1")][ColumnName("name")], eq, CellValue("Alice"))
		so(readBack[UserID("user2")][ColumnName("name")], eq, CellValue("Bob"))
	})
}

func TestWriteCSVStringMaps_SpecialCharacters(t *testing.T) {
	cv("写入包含特殊字符的数据", t, func() {
		convey.Convey("包含逗号", func() {
			data := map[string]map[string]string{
				"row1": {"col1": "hello, world"},
			}

			result, err := csv.WriteCSVStringMaps(data, nil)
			so(err, isNil)

			readBack, _, err := csv.ReadCSVStringMaps[string, string, string](result)
			so(err, isNil)
			so(readBack["row1"]["col1"], eq, "hello, world")
		})

		convey.Convey("包含换行", func() {
			data := map[string]map[string]string{
				"row1": {"col1": "line1\nline2"},
			}

			result, err := csv.WriteCSVStringMaps(data, nil)
			so(err, isNil)

			readBack, _, err := csv.ReadCSVStringMaps[string, string, string](result)
			so(err, isNil)
			so(readBack["row1"]["col1"], eq, "line1\nline2")
		})

		convey.Convey("包含双引号", func() {
			data := map[string]map[string]string{
				"row1": {"col1": `say "hello"`},
			}

			result, err := csv.WriteCSVStringMaps(data, nil)
			so(err, isNil)

			readBack, _, err := csv.ReadCSVStringMaps[string, string, string](result)
			so(err, isNil)
			so(readBack["row1"]["col1"], eq, `say "hello"`)
		})
	})
}

func TestWriteCSVStringMaps_RoundTrip(t *testing.T) {
	cv("读取-写入-读取往返测试", t, func() {
		// 先读取一个标准文件
		data, err := readTestData("normal.csv")
		so(err, isNil)

		original, columnOrder, err := csv.ReadCSVStringMaps[string, string, string](data)
		so(err, isNil)
		so(len(columnOrder), eq, 3) // name, age, city

		// 写入
		written, err := csv.WriteCSVStringMaps(original, nil)
		so(err, isNil)

		// 再读取
		readBack, _, err := csv.ReadCSVStringMaps[string, string, string](written)
		so(err, isNil)

		// 验证数据一致性
		so(len(readBack), eq, len(original))
		for lineKey, lineData := range original {
			so(readBack[lineKey], notNil)
			for colKey, value := range lineData {
				so(readBack[lineKey][colKey], eq, value)
			}
		}
	})
}

// ========== TestWriteCSVStringMaps_ColumnSequences 测试列顺序功能 ==========

func TestWriteCSVStringMaps_ColumnSequences_Default(t *testing.T) {
	cv("传 nil columnSequences 时按字母排序", t, func() {
		data := map[string]map[string]string{
			"row1": {"z_col": "v1", "a_col": "v2", "m_col": "v3"},
		}

		// 传 nil，应该按字母排序
		result, err := csv.WriteCSVStringMaps(data, nil)
		so(err, isNil)
		so(result, notNil)

		// 读取并验证列顺序
		readBack, columnOrder, err := csv.ReadCSVStringMaps[string, string, string](result)
		so(err, isNil)
		so(readBack, notNil)

		// 列应该按字母排序: a_col, m_col, z_col
		so(len(columnOrder), eq, 3)
		so(columnOrder[0], eq, "a_col")
		so(columnOrder[1], eq, "m_col")
		so(columnOrder[2], eq, "z_col")
	})
}

func TestWriteCSVStringMaps_ColumnSequences_CustomOrder(t *testing.T) {
	cv("传入 columnSequences 时按指定顺序输出", t, func() {
		data := map[string]map[string]string{
			"row1": {"col_a": "v1", "col_b": "v2", "col_c": "v3"},
			"row2": {"col_a": "v4", "col_b": "v5", "col_c": "v6"},
		}

		// 指定列顺序: col_c, col_a, col_b
		result, err := csv.WriteCSVStringMaps(data, []string{"col_c", "col_a", "col_b"})
		so(err, isNil)
		so(result, notNil)

		// 读取并验证列顺序
		readBack, columnOrder, err := csv.ReadCSVStringMaps[string, string, string](result)
		so(err, isNil)
		so(readBack, notNil)

		// 列应该按指定顺序: col_c, col_a, col_b
		so(len(columnOrder), eq, 3)
		so(columnOrder[0], eq, "col_c")
		so(columnOrder[1], eq, "col_a")
		so(columnOrder[2], eq, "col_b")

		// 验证数据完整性
		so(readBack["row1"]["col_a"], eq, "v1")
		so(readBack["row1"]["col_b"], eq, "v2")
		so(readBack["row1"]["col_c"], eq, "v3")
		so(readBack["row2"]["col_a"], eq, "v4")
		so(readBack["row2"]["col_b"], eq, "v5")
		so(readBack["row2"]["col_c"], eq, "v6")
	})
}

func TestWriteCSVStringMaps_ColumnSequences_PartialColumns(t *testing.T) {
	cv("指定部分列时，指定的列优先，其余列按字母序排在后面", t, func() {
		data := map[string]map[string]string{
			"row1": {"name": "Alice", "age": "25", "city": "Beijing", "country": "China"},
			"row2": {"name": "Bob", "age": "30", "city": "Shanghai", "country": "China"},
		}

		// 指定 name 和 city 优先
		result, err := csv.WriteCSVStringMaps(data, []string{"name", "city"})
		so(err, isNil)
		so(result, notNil)

		// 读取并验证
		readBack, columnOrder, err := csv.ReadCSVStringMaps[string, string, string](result)
		so(err, isNil)
		so(readBack, notNil)

		// 应该有 4 列: name, city（指定的优先）, age, country（未指定的按字母序）
		so(len(columnOrder), eq, 4)
		so(columnOrder[0], eq, "name")
		so(columnOrder[1], eq, "city")
		so(columnOrder[2], eq, "age")     // 未指定的按字母序
		so(columnOrder[3], eq, "country") // 未指定的按字母序

		// 验证数据
		so(readBack["row1"]["name"], eq, "Alice")
		so(readBack["row1"]["city"], eq, "Beijing")
		so(readBack["row1"]["age"], eq, "25")
		so(readBack["row1"]["country"], eq, "China")

		so(readBack["row2"]["name"], eq, "Bob")
		so(readBack["row2"]["city"], eq, "Shanghai")
	})
}

func TestWriteCSVStringMaps_ColumnSequences_WithNonExistentColumns(t *testing.T) {
	cv("指定的列包含不存在的列名", t, func() {
		data := map[string]map[string]string{
			"row1": {"col_a": "v1", "col_b": "v2"},
		}

		convey.Convey("混合存在和不存在的列名", func() {
			// 指定 col_b（存在）, col_x（不存在）, col_a（存在）
			result, err := csv.WriteCSVStringMaps(data, []string{"col_b", "col_x", "col_a"})
			so(err, isNil)
			so(result, notNil)

			// 读取并验证：指定的存在列优先，不存在的列被忽略
			readBack, columnOrder, err := csv.ReadCSVStringMaps[string, string, string](result)
			so(err, isNil)

			// 应该有 2 列: col_b, col_a（按指定顺序，跳过不存在的 col_x，无其他未指定列）
			so(len(columnOrder), eq, 2)
			so(columnOrder[0], eq, "col_b")
			so(columnOrder[1], eq, "col_a")

			so(readBack["row1"]["col_a"], eq, "v1")
			so(readBack["row1"]["col_b"], eq, "v2")
		})

		convey.Convey("全部都是不存在的列名时，所有列按字母序输出", func() {
			// 如果指定的列全部不存在，则所有实际列按字母序输出
			result, err := csv.WriteCSVStringMaps(data, []string{"col_x", "col_y", "col_z"})
			so(err, isNil)
			so(result, notNil)

			readBack, columnOrder, err := csv.ReadCSVStringMaps[string, string, string](result)
			so(err, isNil)

			// 应该有 2 列，按字母序: col_a, col_b
			so(len(columnOrder), eq, 2)
			so(columnOrder[0], eq, "col_a")
			so(columnOrder[1], eq, "col_b")

			so(readBack["row1"]["col_a"], eq, "v1")
			so(readBack["row1"]["col_b"], eq, "v2")
		})
	})
}

func TestWriteCSVStringMaps_ColumnSequences_ChineseColumns(t *testing.T) {
	cv("使用中文列名并指定顺序", t, func() {
		data := map[string]map[string]string{
			"用户1": {"姓名": "张三", "年龄": "25", "城市": "北京"},
			"用户2": {"姓名": "李四", "年龄": "30", "城市": "上海"},
		}

		// 指定中文列顺序: 城市, 年龄, 姓名（与字母序不同）
		result, err := csv.WriteCSVStringMaps(data, []string{"城市", "年龄", "姓名"})
		so(err, isNil)
		so(result, notNil)

		// 读取并验证列顺序
		readBack, columnOrder, err := csv.ReadCSVStringMaps[string, string, string](result)
		so(err, isNil)
		so(readBack, notNil)

		// 列应该按指定顺序
		so(len(columnOrder), eq, 3)
		so(columnOrder[0], eq, "城市")
		so(columnOrder[1], eq, "年龄")
		so(columnOrder[2], eq, "姓名")

		// 验证数据完整性
		so(readBack["用户1"]["姓名"], eq, "张三")
		so(readBack["用户1"]["年龄"], eq, "25")
		so(readBack["用户1"]["城市"], eq, "北京")
	})
}

func TestWriteCSVStringMaps_ColumnSequences_DuplicateColumns(t *testing.T) {
	cv("指定重复的列名", t, func() {
		data := map[string]map[string]string{
			"row1": {"col_a": "v1", "col_b": "v2", "col_c": "v3"},
		}

		// 指定重复的列名: col_a, col_b, col_a
		// 实现中会去重，重复的列只输出一次（第一次出现的位置）
		// 未指定的 col_c 会排在后面
		result, err := csv.WriteCSVStringMaps(data, []string{"col_a", "col_b", "col_a"})
		so(err, isNil)
		so(result, notNil)

		// 读取并验证
		readBack, columnOrder, err := csv.ReadCSVStringMaps[string, string, string](result)
		so(err, isNil)

		// 应该有 3 列: col_a, col_b（去重后的指定列）, col_c（未指定的列按字母序）
		so(len(columnOrder), eq, 3)
		so(columnOrder[0], eq, "col_a")
		so(columnOrder[1], eq, "col_b")
		so(columnOrder[2], eq, "col_c")

		so(readBack["row1"]["col_a"], eq, "v1")
		so(readBack["row1"]["col_b"], eq, "v2")
		so(readBack["row1"]["col_c"], eq, "v3")
	})
}

func TestWriteCSVStringMaps_ColumnSequences_SparseData(t *testing.T) {
	cv("稀疏数据按指定列顺序输出", t, func() {
		// 不同行有不同的列
		data := map[string]map[string]string{
			"row1": {"col_a": "v1", "col_c": "v3"},   // 没有 col_b
			"row2": {"col_b": "v22", "col_c": "v23"}, // 没有 col_a
			"row3": {"col_a": "v31", "col_b": "v32"}, // 没有 col_c
		}

		// 指定列顺序: col_c, col_b, col_a（正好覆盖所有列）
		result, err := csv.WriteCSVStringMaps(data, []string{"col_c", "col_b", "col_a"})
		so(err, isNil)
		so(result, notNil)

		// 读取并验证
		readBack, columnOrder, err := csv.ReadCSVStringMaps[string, string, string](result)
		so(err, isNil)

		// 验证列顺序
		so(len(columnOrder), eq, 3)
		so(columnOrder[0], eq, "col_c")
		so(columnOrder[1], eq, "col_b")
		so(columnOrder[2], eq, "col_a")

		// 验证数据（空值不应该存在于结果中）
		so(readBack["row1"]["col_a"], eq, "v1")
		so(readBack["row1"]["col_c"], eq, "v3")
		_, hasColB := readBack["row1"]["col_b"]
		so(hasColB, isFalse)

		so(readBack["row2"]["col_b"], eq, "v22")
		so(readBack["row2"]["col_c"], eq, "v23")

		so(readBack["row3"]["col_a"], eq, "v31")
		so(readBack["row3"]["col_b"], eq, "v32")
	})
}

func TestWriteCSVStringMaps_ColumnSequences_Deterministic(t *testing.T) {
	cv("指定列顺序后多次写入应产生相同的输出", t, func() {
		data := map[string]map[string]string{
			"z_row": {"z_col": "v1", "a_col": "v2", "m_col": "v3"},
			"a_row": {"z_col": "v4", "a_col": "v5", "m_col": "v6"},
			"m_row": {"z_col": "v7", "a_col": "v8", "m_col": "v9"},
		}

		// 指定固定的列顺序
		columnOrder := []string{"m_col", "z_col", "a_col"}

		// 多次调用，结果应该完全一致
		result1, err := csv.WriteCSVStringMaps(data, columnOrder)
		so(err, isNil)

		result2, err := csv.WriteCSVStringMaps(data, columnOrder)
		so(err, isNil)

		result3, err := csv.WriteCSVStringMaps(data, columnOrder)
		so(err, isNil)

		so(string(result1), eq, string(result2))
		so(string(result2), eq, string(result3))
	})
}

func TestWriteCSVStringMaps_ColumnSequences_CustomTypes(t *testing.T) {
	cv("使用自定义类型并指定列顺序", t, func() {
		data := map[UserID]map[ColumnName]CellValue{
			UserID("user1"): {
				ColumnName("name"): CellValue("Alice"),
				ColumnName("age"):  CellValue("25"),
				ColumnName("city"): CellValue("Beijing"),
			},
			UserID("user2"): {
				ColumnName("name"): CellValue("Bob"),
				ColumnName("age"):  CellValue("30"),
				ColumnName("city"): CellValue("Shanghai"),
			},
		}

		// 指定列顺序: city, name 优先，age 会排在后面
		result, err := csv.WriteCSVStringMaps(data, []ColumnName{ColumnName("city"), ColumnName("name")})
		so(err, isNil)
		so(result, notNil)

		// 读取并验证
		readBack, columnOrder, err := csv.ReadCSVStringMaps[UserID, ColumnName, CellValue](result)
		so(err, isNil)

		// 验证列顺序: city, name（指定的）, age（未指定的按字母序）
		so(len(columnOrder), eq, 3)
		so(columnOrder[0], eq, ColumnName("city"))
		so(columnOrder[1], eq, ColumnName("name"))
		so(columnOrder[2], eq, ColumnName("age"))

		// 验证数据
		so(readBack[UserID("user1")][ColumnName("name")], eq, CellValue("Alice"))
		so(readBack[UserID("user1")][ColumnName("city")], eq, CellValue("Beijing"))
		so(readBack[UserID("user1")][ColumnName("age")], eq, CellValue("25"))
	})
}

func TestWriteCSVStringMaps_ColumnSequences_SingleColumn(t *testing.T) {
	cv("指定单列优先", t, func() {
		data := map[string]map[string]string{
			"row1": {"col_a": "v1", "col_b": "v2", "col_c": "v3"},
			"row2": {"col_a": "v4", "col_b": "v5", "col_c": "v6"},
		}

		// 指定 col_b 优先，其余列按字母序排在后面
		result, err := csv.WriteCSVStringMaps(data, []string{"col_b"})
		so(err, isNil)
		so(result, notNil)

		// 读取并验证
		readBack, columnOrder, err := csv.ReadCSVStringMaps[string, string, string](result)
		so(err, isNil)

		// 应该有 3 列: col_b（指定的优先）, col_a, col_c（未指定的按字母序）
		so(len(columnOrder), eq, 3)
		so(columnOrder[0], eq, "col_b")
		so(columnOrder[1], eq, "col_a")
		so(columnOrder[2], eq, "col_c")

		// 验证数据
		so(readBack["row1"]["col_a"], eq, "v1")
		so(readBack["row1"]["col_b"], eq, "v2")
		so(readBack["row1"]["col_c"], eq, "v3")
		so(readBack["row2"]["col_b"], eq, "v5")
	})
}

// ========== Benchmark 测试 ==========

func BenchmarkReadCSVStringMaps(b *testing.B) {
	data, err := readTestData("normal.csv")
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = csv.ReadCSVStringMaps[string, string, string](data)
	}
}

func BenchmarkReadCSVStringMaps_Large(b *testing.B) {
	// 构造一个较大的 CSV 数据
	var largeData []byte
	largeData = append(largeData, []byte("ignore,col1,col2,col3,col4,col5\n")...)
	for i := 0; i < 1000; i++ {
		line := []byte("row" + string(rune('0'+i%10)) + ",v1,v2,v3,v4,v5\n")
		largeData = append(largeData, line...)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = csv.ReadCSVStringMaps[string, string, string](largeData)
	}
}
