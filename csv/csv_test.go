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

		result, err := csv.ReadCSVStringMaps[string, string, string](data)
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

		result, err := csv.ReadCSVStringMaps[string, string, string](data)
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

		result, err := csv.ReadCSVStringMaps[string, string, string](data)
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

		result, err := csv.ReadCSVStringMaps[string, string, string](data)
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

		result, err := csv.ReadCSVStringMaps[string, string, string](data)
		so(err, notNil) // 应该返回解析错误
		so(result, isNil)
	})
}

func TestReadCSVStringMaps_SingleColumn(t *testing.T) {
	cv("读取单列的 CSV 文件应返回错误", t, func() {
		data, err := readTestData("single_column.csv")
		so(err, isNil)

		result, err := csv.ReadCSVStringMaps[string, string, string](data)
		so(err, notNil) // 应该返回错误
		so(result, isNil)
		so(err.Error(), eq, "CSV 数据格式不正确：第一行至少需要两列")
	})
}

func TestReadCSVStringMaps_Empty(t *testing.T) {
	cv("读取空的 CSV 文件应返回错误", t, func() {
		data, err := readTestData("empty.csv")
		so(err, isNil)

		result, err := csv.ReadCSVStringMaps[string, string, string](data)
		so(err, notNil) // 应该返回错误
		so(result, isNil)
		so(err.Error(), eq, "CSV 数据为空")
	})
}

func TestReadCSVStringMaps_Minimal(t *testing.T) {
	cv("读取最小有效的 CSV 文件", t, func() {
		data, err := readTestData("minimal.csv")
		so(err, isNil)

		result, err := csv.ReadCSVStringMaps[string, string, string](data)
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

		result, err := csv.ReadCSVStringMaps[string, string, string](data)
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
			result, err := csv.ReadCSVStringMaps[string, string, string](dataWithBOM)
			so(err, isNil)
			so(result, notNil)
			so(result["row1"]["col1"], eq, "value1")
		})

		// 测试 UTF-16 LE BOM (0xFF 0xFE)
		convey.Convey("测试 UTF-16 LE BOM", func() {
			dataWithBOM := append([]byte{0xFF, 0xFE}, normalCSV...)
			result, err := csv.ReadCSVStringMaps[string, string, string](dataWithBOM)
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
		result, err := csv.ReadCSVStringMaps[UserID, ColumnName, CellValue](data)
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

		result, err := csv.ReadCSVStringMaps[string, string, string](malformedData)
		so(err, notNil) // 应该返回解析错误
		so(result, isNil)
	})
}

func TestReadCSVStringMaps_HeaderOnlyFile(t *testing.T) {
	cv("测试只有标题行的 CSV 文件", t, func() {
		// 只有标题行，没有数据行
		headerOnlyData := []byte("ignore,col1,col2\n")

		result, err := csv.ReadCSVStringMaps[string, string, string](headerOnlyData)
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
			result, err := csv.ReadCSVStringMaps[string, string, string](data)
			so(err, notNil) // 应该返回解析错误
			so(result, isNil)
		})

		convey.Convey("数据行比标题行少列应返回错误", func() {
			data := []byte("ignore,col1,col2,col3\nrow1,v1\n")
			result, err := csv.ReadCSVStringMaps[string, string, string](data)
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
			result, err := csv.ReadCSVStringMaps[string, string, string](data)
			so(err, isNil)
			so(result["row1"]["col1"], eq, "hello, world")
		})

		convey.Convey("包含换行的引号字段", func() {
			data := []byte(`ignore,col1,col2
row1,"line1
line2",value2
`)
			result, err := csv.ReadCSVStringMaps[string, string, string](data)
			so(err, isNil)
			so(result["row1"]["col1"], eq, "line1\nline2")
		})

		convey.Convey("包含双引号的引号字段", func() {
			data := []byte(`ignore,col1,col2
row1,"say ""hello""",value2
`)
			result, err := csv.ReadCSVStringMaps[string, string, string](data)
			so(err, isNil)
			so(result["row1"]["col1"], eq, `say "hello"`)
		})
	})
}

func TestReadCSVStringMaps_EmptyRowKey(t *testing.T) {
	cv("测试行键为空字符串的情况", t, func() {
		data := []byte("ignore,col1,col2\n,v1,v2\nrow2,v21,v22\n")

		result, err := csv.ReadCSVStringMaps[string, string, string](data)
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

		result, err := csv.ReadCSVStringMaps[string, string, string](data)
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
		result, err := csv.ReadCSVStringMaps[string, string, string](data)
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
		// UTF-8 BOM: 0xEF 0xBB 0xBF
		// 注意：当前实现不处理 UTF-8 BOM，所以第一个列名会带有 BOM 前缀
		// 这个测试是为了记录当前行为
		normalCSV := []byte("ignore,col1\nrow1,value1\n")
		dataWithBOM := append([]byte{0xEF, 0xBB, 0xBF}, normalCSV...)

		result, err := csv.ReadCSVStringMaps[string, string, string](dataWithBOM)
		so(err, isNil)
		so(result, notNil)
		// 由于 UTF-8 BOM 未被处理，第一列键名会带有 BOM 前缀
		so(result["row1"]["col1"], eq, "value1")
	})
}

func TestReadCSVStringMaps_EmptyColumnHeader(t *testing.T) {
	cv("测试列头为空字符串的情况", t, func() {
		// 列头为空字符串
		data := []byte("ignore,,col2\nrow1,v1,v2\n")

		result, err := csv.ReadCSVStringMaps[string, string, string](data)
		so(err, isNil)
		so(result, notNil)

		// 空字符串作为列头也是有效的
		so(result["row1"], notNil)
		so(result["row1"][""], eq, "v1")
		so(result["row1"]["col2"], eq, "v2")
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
		_, _ = csv.ReadCSVStringMaps[string, string, string](data)
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
		_, _ = csv.ReadCSVStringMaps[string, string, string](largeData)
	}
}
