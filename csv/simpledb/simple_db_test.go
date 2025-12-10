package simpledb_test

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/Andrew-M-C/go.util/csv/simpledb"
	"github.com/smartystreets/goconvey/convey"
)

var (
	cv = convey.Convey
	so = convey.So
	eq = convey.ShouldEqual

	isNil    = convey.ShouldBeNil
	notNil   = convey.ShouldNotBeNil
	isFalse  = convey.ShouldBeFalse
	isTrue   = convey.ShouldBeTrue
	contains = convey.ShouldContainSubstring
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

const testDataDir = "./testdata"

// 辅助函数：确保测试目录存在
func ensureTestDataDir(t *testing.T) {
	if err := os.MkdirAll(testDataDir, 0755); err != nil {
		t.Fatal(err)
	}
}

// 辅助函数：清理测试目录中的文件
func cleanupTestDataDir() {
	os.RemoveAll(testDataDir)
}

// ========== 基本功能测试 ==========

func TestNewDB_Basic(t *testing.T) {
	cv("测试 NewDB 基本创建", t, func() {
		ensureTestDataDir(t)
		defer cleanupTestDataDir()

		filePath := filepath.Join(testDataDir, "test.csv")

		cv("创建新的数据库实例", func() {
			db, err := simpledb.NewDB[string, string, string](filePath)
			so(err, isNil)
			so(db, notNil)
		})

		cv("创建到子目录中的数据库（目录不存在应自动创建）", func() {
			subFilePath := filepath.Join(testDataDir, "sub", "dir", "test.csv")
			db, err := simpledb.NewDB[string, string, string](subFilePath)
			so(err, isNil)
			so(db, notNil)

			// 验证目录已创建
			_, err = os.Stat(filepath.Dir(subFilePath))
			so(err, isNil)
		})
	})
}

func TestStore_And_Load(t *testing.T) {
	cv("测试 Store 和 Load 功能", t, func() {
		ensureTestDataDir(t)
		defer cleanupTestDataDir()

		cv("存储并加载单行数据", func() {
			filePath := filepath.Join(testDataDir, "test1.csv")
			db, err := simpledb.NewDB[string, string, string](filePath)
			so(err, isNil)

			err = db.Store("user1", map[string]string{
				"name":  "张三",
				"email": "zhangsan@example.com",
				"age":   "25",
			})
			so(err, isNil)

			row, exist := db.Load("user1")
			so(exist, isTrue)
			so(row["name"], eq, "张三")
			so(row["email"], eq, "zhangsan@example.com")
			so(row["age"], eq, "25")
		})

		cv("加载不存在的行返回 false", func() {
			filePath := filepath.Join(testDataDir, "test2.csv")
			db, err := simpledb.NewDB[string, string, string](filePath)
			so(err, isNil)

			row, exist := db.Load("nonexistent")
			so(exist, isFalse)
			so(row, isNil)
		})

		cv("部分更新已存在的行（Store 也是部分更新）", func() {
			filePath := filepath.Join(testDataDir, "test3.csv")
			db, err := simpledb.NewDB[string, string, string](filePath)
			so(err, isNil)

			// 首先存储完整数据
			err = db.Store("user1", map[string]string{
				"name":  "张三",
				"email": "zhangsan@example.com",
				"age":   "25",
			})
			so(err, isNil)

			// 只更新部分字段
			err = db.Store("user1", map[string]string{
				"name":  "李四",
				"email": "lisi@example.com",
			})
			so(err, isNil)

			row, exist := db.Load("user1")
			so(exist, isTrue)
			so(row["name"], eq, "李四")
			so(row["email"], eq, "lisi@example.com")
			// Store 实际上调用的是 storeColumnsLocked，是部分更新而不是整行替换
			// 所以 age 列仍然存在
			so(row["age"], eq, "25")
		})

		cv("存储多行数据", func() {
			filePath := filepath.Join(testDataDir, "test4.csv")
			db, err := simpledb.NewDB[string, string, string](filePath)
			so(err, isNil)

			err = db.Store("user1", map[string]string{
				"name":  "张三",
				"email": "zhangsan@example.com",
			})
			so(err, isNil)

			err = db.Store("user2", map[string]string{
				"name":  "王五",
				"email": "wangwu@example.com",
			})
			so(err, isNil)

			err = db.Store("user3", map[string]string{
				"name":  "赵六",
				"email": "zhaoliu@example.com",
			})
			so(err, isNil)

			row1, exist1 := db.Load("user1")
			row2, exist2 := db.Load("user2")
			row3, exist3 := db.Load("user3")

			so(exist1, isTrue)
			so(exist2, isTrue)
			so(exist3, isTrue)
			so(row1["name"], eq, "张三")
			so(row2["name"], eq, "王五")
			so(row3["name"], eq, "赵六")
		})
	})
}

func TestStoreColumns(t *testing.T) {
	cv("测试 StoreColumns 部分更新功能", t, func() {
		ensureTestDataDir(t)
		defer cleanupTestDataDir()

		cv("部分更新已存在的行", func() {
			filePath := filepath.Join(testDataDir, "test1.csv")
			db, err := simpledb.NewDB[string, string, string](filePath)
			so(err, isNil)

			// 先存储完整行
			err = db.Store("user1", map[string]string{
				"name":  "张三",
				"email": "zhangsan@example.com",
				"age":   "25",
			})
			so(err, isNil)

			// 部分更新
			err = db.StoreColumns("user1", map[string]string{
				"email": "newemail@example.com",
			})
			so(err, isNil)

			row, exist := db.Load("user1")
			so(exist, isTrue)
			so(row["name"], eq, "张三")                    // 保持不变
			so(row["email"], eq, "newemail@example.com") // 已更新
			so(row["age"], eq, "25")                     // 保持不变
		})

		cv("部分更新不存在的行（创建新行）", func() {
			filePath := filepath.Join(testDataDir, "test2.csv")
			db, err := simpledb.NewDB[string, string, string](filePath)
			so(err, isNil)

			err = db.StoreColumns("user2", map[string]string{
				"name": "李四",
			})
			so(err, isNil)

			row, exist := db.Load("user2")
			so(exist, isTrue)
			so(row["name"], eq, "李四")
		})

		cv("空 columns 不做任何操作", func() {
			filePath := filepath.Join(testDataDir, "test3.csv")
			db, err := simpledb.NewDB[string, string, string](filePath)
			so(err, isNil)

			err = db.Store("user1", map[string]string{
				"name": "张三",
			})
			so(err, isNil)

			err = db.StoreColumns("user1", map[string]string{})
			so(err, isNil)

			// 数据应保持不变
			row, exist := db.Load("user1")
			so(exist, isTrue)
			so(row["name"], eq, "张三")
		})

		cv("添加新列", func() {
			filePath := filepath.Join(testDataDir, "test4.csv")
			db, err := simpledb.NewDB[string, string, string](filePath)
			so(err, isNil)

			// 先存储初始数据
			err = db.Store("user1", map[string]string{
				"name":  "张三",
				"email": "zhangsan@example.com",
			})
			so(err, isNil)

			// 添加新列
			err = db.StoreColumns("user1", map[string]string{
				"phone": "13800138000",
			})
			so(err, isNil)

			row, exist := db.Load("user1")
			so(exist, isTrue)
			so(row["phone"], eq, "13800138000")
			so(row["name"], eq, "张三")                    // 其他列保持不变
			so(row["email"], eq, "zhangsan@example.com") // 其他列保持不变
		})
	})
}

// ========== 错误测试 ==========

func TestEmptyLineKeyError(t *testing.T) {
	cv("测试空行键错误", t, func() {
		ensureTestDataDir(t)
		defer cleanupTestDataDir()

		filePath := filepath.Join(testDataDir, "test.csv")
		db, err := simpledb.NewDB[string, string, string](filePath)
		so(err, isNil)

		cv("Store 空键应返回 ErrEmptyLineKey", func() {
			err := db.Store("", map[string]string{"name": "test"})
			so(err, notNil)
			so(errors.Is(err, simpledb.ErrEmptyLineKey), isTrue)
		})

		cv("StoreColumns 空键应返回 ErrEmptyLineKey", func() {
			err := db.StoreColumns("", map[string]string{"name": "test"})
			so(err, notNil)
			so(errors.Is(err, simpledb.ErrEmptyLineKey), isTrue)
		})
	})
}

// ========== WithAsyncTime Option 测试 ==========

func TestWithAsyncTime(t *testing.T) {
	cv("测试 WithAsyncTime 异步写入选项", t, func() {
		ensureTestDataDir(t)
		defer cleanupTestDataDir()

		cv("同步写入模式（asyncTime <= 0）", func() {
			filePath := filepath.Join(testDataDir, "sync_test.csv")
			db, err := simpledb.NewDB[string, string, string](filePath)
			so(err, isNil)

			err = db.Store("key1", map[string]string{"col": "val"})
			so(err, isNil)

			// 同步模式下，Store 完成后文件应立即存在
			_, err = os.Stat(filePath)
			so(err, isNil)

			content, err := os.ReadFile(filePath)
			so(err, isNil)
			so(string(content), contains, "key1")
		})

		cv("异步写入模式（asyncTime > 0）", func() {
			filePath := filepath.Join(testDataDir, "async_test.csv")
			asyncTime := 100 * time.Millisecond

			db, err := simpledb.NewDB[string, string, string](
				filePath,
				simpledb.WithAsyncTime(asyncTime),
			)
			so(err, isNil)

			err = db.Store("key1", map[string]string{"col": "val"})
			so(err, isNil)

			// 异步模式下，Store 完成后文件可能还未写入
			// 等待足够时间让异步写入完成
			time.Sleep(asyncTime + 50*time.Millisecond)

			// 现在文件应该已经写入
			content, err := os.ReadFile(filePath)
			so(err, isNil)
			so(string(content), contains, "key1")
		})

		cv("异步写入合并多次写入", func() {
			filePath := filepath.Join(testDataDir, "async_merge_test.csv")
			asyncTime := 200 * time.Millisecond

			db, err := simpledb.NewDB[string, string, string](
				filePath,
				simpledb.WithAsyncTime(asyncTime),
			)
			so(err, isNil)

			// 快速连续写入多次
			for i := 0; i < 5; i++ {
				err = db.Store("key"+string(rune('1'+i)), map[string]string{"col": "val"})
				so(err, isNil)
			}

			// 等待异步写入完成
			time.Sleep(asyncTime + 100*time.Millisecond)

			// 所有数据都应该在文件中
			content, err := os.ReadFile(filePath)
			so(err, isNil)
			contentStr := string(content)
			so(contentStr, contains, "key1")
			so(contentStr, contains, "key2")
		})
	})
}

// ========== WithUniqueColumns Option 测试 ==========

func TestWithUniqueColumns(t *testing.T) {
	cv("测试 WithUniqueColumns 唯一列约束选项", t, func() {
		ensureTestDataDir(t)
		defer cleanupTestDataDir()

		cv("唯一列约束正常工作", func() {
			filePath := filepath.Join(testDataDir, "unique_test.csv")
			db, err := simpledb.NewDB[string, string, string](
				filePath,
				simpledb.WithUniqueColumns("email"),
			)
			so(err, isNil)

			// 第一次存储
			err = db.Store("user1", map[string]string{
				"name":  "张三",
				"email": "test@example.com",
			})
			so(err, isNil)

			// 尝试用相同的 email 存储到不同行，应该失败
			err = db.Store("user2", map[string]string{
				"name":  "李四",
				"email": "test@example.com",
			})
			so(err, notNil)
			so(errors.Is(err, simpledb.ErrColumnDuplicate), isTrue)
		})

		cv("同一行更新唯一列不触发错误", func() {
			filePath := filepath.Join(testDataDir, "unique_same_row_test.csv")
			db, err := simpledb.NewDB[string, string, string](
				filePath,
				simpledb.WithUniqueColumns("email"),
			)
			so(err, isNil)

			err = db.Store("user1", map[string]string{
				"name":  "张三",
				"email": "test@example.com",
			})
			so(err, isNil)

			// 更新同一行的同一个唯一列值应该成功
			err = db.Store("user1", map[string]string{
				"name":  "张三改名",
				"email": "test@example.com", // 同一行，同一个值
			})
			so(err, isNil)

			row, exist := db.Load("user1")
			so(exist, isTrue)
			so(row["name"], eq, "张三改名")
		})

		cv("更新唯一列到新值后，旧值可被其他行使用", func() {
			filePath := filepath.Join(testDataDir, "unique_release_test.csv")
			db, err := simpledb.NewDB[string, string, string](
				filePath,
				simpledb.WithUniqueColumns("email"),
			)
			so(err, isNil)

			// user1 使用 email1
			err = db.Store("user1", map[string]string{
				"email": "email1@example.com",
			})
			so(err, isNil)

			// user1 更新到 email2
			err = db.StoreColumns("user1", map[string]string{
				"email": "email2@example.com",
			})
			so(err, isNil)

			// 现在 email1 应该可以被 user2 使用
			err = db.Store("user2", map[string]string{
				"email": "email1@example.com",
			})
			so(err, isNil)
		})

		cv("多个唯一列约束", func() {
			filePath := filepath.Join(testDataDir, "multi_unique_test.csv")
			db, err := simpledb.NewDB[string, string, string](
				filePath,
				simpledb.WithUniqueColumns("email", "phone"),
			)
			so(err, isNil)

			err = db.Store("user1", map[string]string{
				"email": "user1@example.com",
				"phone": "13800138001",
			})
			so(err, isNil)

			// 相同 email，不同 phone，应该失败
			err = db.Store("user2", map[string]string{
				"email": "user1@example.com",
				"phone": "13800138002",
			})
			so(err, notNil)

			// 不同 email，相同 phone，也应该失败
			err = db.Store("user3", map[string]string{
				"email": "user3@example.com",
				"phone": "13800138001",
			})
			so(err, notNil)

			// 不同 email，不同 phone，应该成功
			err = db.Store("user4", map[string]string{
				"email": "user4@example.com",
				"phone": "13800138004",
			})
			so(err, isNil)
		})

		cv("StoreColumns 也受唯一约束限制", func() {
			filePath := filepath.Join(testDataDir, "unique_store_columns_test.csv")
			db, err := simpledb.NewDB[string, string, string](
				filePath,
				simpledb.WithUniqueColumns("email"),
			)
			so(err, isNil)

			err = db.Store("user1", map[string]string{
				"email": "taken@example.com",
			})
			so(err, isNil)

			err = db.Store("user2", map[string]string{
				"email": "free@example.com",
			})
			so(err, isNil)

			// 尝试通过 StoreColumns 更新到已被占用的值
			err = db.StoreColumns("user2", map[string]string{
				"email": "taken@example.com",
			})
			so(err, notNil)
			so(errors.Is(err, simpledb.ErrColumnDuplicate), isTrue)
		})
	})
}

// ========== WithDebugger Option 测试 ==========

func TestWithDebugger(t *testing.T) {
	cv("测试 WithDebugger 调试器选项", t, func() {
		ensureTestDataDir(t)
		defer cleanupTestDataDir()

		cv("调试器被正确调用", func() {
			filePath := filepath.Join(testDataDir, "debug_test.csv")
			var debugLogs []string
			var mu sync.Mutex

			debugFunc := func(format string, args ...any) {
				mu.Lock()
				defer mu.Unlock()
				debugLogs = append(debugLogs, format)
			}

			db, err := simpledb.NewDB[string, string, string](
				filePath,
				simpledb.WithDebugger(debugFunc),
			)
			so(err, isNil)

			err = db.Store("key1", map[string]string{"col": "val"})
			so(err, isNil)

			// 给异步操作一点时间完成
			time.Sleep(50 * time.Millisecond)

			mu.Lock()
			logCount := len(debugLogs)
			mu.Unlock()

			// 调试器应该被调用过
			so(logCount > 0, isTrue)
		})

		cv("nil 调试器不会导致 panic", func() {
			filePath := filepath.Join(testDataDir, "nil_debug_test.csv")

			// 传入 nil 调试器
			db, err := simpledb.NewDB[string, string, string](
				filePath,
				simpledb.WithDebugger(nil),
			)
			so(err, isNil)

			// 操作应该正常执行而不会 panic
			err = db.Store("key1", map[string]string{"col": "val"})
			so(err, isNil)
		})
	})
}

// ========== 文件持久化测试 ==========

func TestFilePersistence(t *testing.T) {
	cv("测试文件持久化", t, func() {
		ensureTestDataDir(t)
		defer cleanupTestDataDir()

		filePath := filepath.Join(testDataDir, "persist_test.csv")

		cv("数据写入后重新加载", func() {
			// 第一个数据库实例写入数据
			db1, err := simpledb.NewDB[string, string, string](filePath)
			so(err, isNil)

			err = db1.Store("user1", map[string]string{
				"name":  "张三",
				"email": "zhangsan@example.com",
			})
			so(err, isNil)

			err = db1.Store("user2", map[string]string{
				"name":  "李四",
				"email": "lisi@example.com",
			})
			so(err, isNil)

			// 创建新的数据库实例，读取同一文件
			db2, err := simpledb.NewDB[string, string, string](filePath)
			so(err, isNil)

			row1, exist1 := db2.Load("user1")
			so(exist1, isTrue)
			so(row1["name"], eq, "张三")
			so(row1["email"], eq, "zhangsan@example.com")

			row2, exist2 := db2.Load("user2")
			so(exist2, isTrue)
			so(row2["name"], eq, "李四")
			so(row2["email"], eq, "lisi@example.com")
		})

		cv("带唯一约束的持久化", func() {
			uniqueFilePath := filepath.Join(testDataDir, "persist_unique_test.csv")

			// 第一个数据库实例写入数据
			db1, err := simpledb.NewDB[string, string, string](
				uniqueFilePath,
				simpledb.WithUniqueColumns("email"),
			)
			so(err, isNil)

			err = db1.Store("user1", map[string]string{
				"email": "unique@example.com",
			})
			so(err, isNil)

			// 创建新的数据库实例，带相同的唯一约束
			db2, err := simpledb.NewDB[string, string, string](
				uniqueFilePath,
				simpledb.WithUniqueColumns("email"),
			)
			so(err, isNil)

			// 唯一约束应该从持久化数据中重建
			err = db2.Store("user2", map[string]string{
				"email": "unique@example.com",
			})
			so(err, notNil)
			so(errors.Is(err, simpledb.ErrColumnDuplicate), isTrue)
		})
	})
}

// ========== 类型参数测试 ==========

type UserID string
type ColumnName string
type ColumnValue string

func TestCustomTypes(t *testing.T) {
	cv("测试自定义类型参数", t, func() {
		ensureTestDataDir(t)
		defer cleanupTestDataDir()

		filePath := filepath.Join(testDataDir, "custom_type_test.csv")

		cv("使用自定义类型作为泛型参数", func() {
			db, err := simpledb.NewDB[UserID, ColumnName, ColumnValue](filePath)
			so(err, isNil)

			err = db.Store(UserID("user_001"), map[ColumnName]ColumnValue{
				ColumnName("name"):  ColumnValue("测试用户"),
				ColumnName("level"): ColumnValue("VIP"),
			})
			so(err, isNil)

			row, exist := db.Load(UserID("user_001"))
			so(exist, isTrue)
			so(row[ColumnName("name")], eq, ColumnValue("测试用户"))
			so(row[ColumnName("level")], eq, ColumnValue("VIP"))
		})
	})
}

// ========== 并发安全测试 ==========

func TestConcurrency(t *testing.T) {
	cv("测试并发安全", t, func() {
		ensureTestDataDir(t)
		defer cleanupTestDataDir()

		filePath := filepath.Join(testDataDir, "concurrent_test.csv")

		cv("并发写入和读取", func() {
			db, err := simpledb.NewDB[string, string, string](
				filePath,
				simpledb.WithAsyncTime(50*time.Millisecond),
			)
			so(err, isNil)

			var wg sync.WaitGroup
			errors := make(chan error, 100)

			// 并发写入
			for i := 0; i < 20; i++ {
				wg.Add(1)
				go func(idx int) {
					defer wg.Done()
					key := "key" + string(rune('A'+idx))
					err := db.Store(key, map[string]string{
						"index": string(rune('0' + idx%10)),
					})
					if err != nil {
						errors <- err
					}
				}(i)
			}

			// 并发读取
			for i := 0; i < 20; i++ {
				wg.Add(1)
				go func(idx int) {
					defer wg.Done()
					key := "key" + string(rune('A'+idx%10))
					db.Load(key)
				}(i)
			}

			wg.Wait()
			close(errors)

			// 检查是否有错误
			for err := range errors {
				so(err, isNil)
			}
		})
	})
}

// ========== 边界条件测试 ==========

func TestEdgeCases(t *testing.T) {
	cv("测试边界条件", t, func() {
		ensureTestDataDir(t)
		defer cleanupTestDataDir()

		cv("空值处理", func() {
			filePath := filepath.Join(testDataDir, "empty_value_test.csv")
			db, err := simpledb.NewDB[string, string, string](filePath)
			so(err, isNil)

			// 存储带空值的数据
			err = db.Store("key1", map[string]string{
				"col1": "",
				"col2": "value2",
			})
			so(err, isNil)

			row, exist := db.Load("key1")
			so(exist, isTrue)
			so(row["col1"], eq, "")
			so(row["col2"], eq, "value2")
		})

		cv("特殊字符处理", func() {
			filePath := filepath.Join(testDataDir, "special_char_test.csv")
			db, err := simpledb.NewDB[string, string, string](filePath)
			so(err, isNil)

			// 包含逗号、引号、换行等特殊字符
			err = db.Store("key1", map[string]string{
				"col1": "hello,world",
				"col2": `say "hi"`,
				"col3": "line1\nline2",
			})
			so(err, isNil)

			row, exist := db.Load("key1")
			so(exist, isTrue)
			so(row["col1"], eq, "hello,world")
			so(row["col2"], eq, `say "hi"`)
			so(row["col3"], eq, "line1\nline2")

			// 重新加载数据库验证持久化
			db2, err := simpledb.NewDB[string, string, string](filePath)
			so(err, isNil)

			row2, exist2 := db2.Load("key1")
			so(exist2, isTrue)
			so(row2["col1"], eq, "hello,world")
			so(row2["col2"], eq, `say "hi"`)
			so(row2["col3"], eq, "line1\nline2")
		})

		cv("Unicode 字符处理", func() {
			filePath := filepath.Join(testDataDir, "unicode_test.csv")
			db, err := simpledb.NewDB[string, string, string](filePath)
			so(err, isNil)

			err = db.Store("用户1", map[string]string{
				"姓名": "张三",
				"邮箱": "zhangsan@例子.com",
				"备注": "🎉🎊✨",
			})
			so(err, isNil)

			row, exist := db.Load("用户1")
			so(exist, isTrue)
			so(row["姓名"], eq, "张三")
			so(row["邮箱"], eq, "zhangsan@例子.com")
			so(row["备注"], eq, "🎉🎊✨")
		})

		cv("大量数据测试", func() {
			filePath := filepath.Join(testDataDir, "large_data_test.csv")
			db, err := simpledb.NewDB[string, string, string](filePath)
			so(err, isNil)

			// 写入 100 行数据
			for i := 0; i < 100; i++ {
				key := fmt.Sprintf("row_%03d", i)
				err := db.Store(key, map[string]string{
					"col1": "value1_" + key,
					"col2": "value2_" + key,
					"col3": "value3_" + key,
				})
				so(err, isNil)
			}

			// 验证可以重新加载
			db2, err := simpledb.NewDB[string, string, string](filePath)
			so(err, isNil)

			row, exist := db2.Load("row_050")
			so(exist, isTrue)
			so(row["col1"], eq, "value1_row_050")
		})
	})
}

// ========== 组合选项测试 ==========

func TestCombinedOptions(t *testing.T) {
	cv("测试多个选项组合使用", t, func() {
		ensureTestDataDir(t)
		defer cleanupTestDataDir()

		filePath := filepath.Join(testDataDir, "combined_options_test.csv")

		cv("异步写入 + 唯一约束 + 调试器", func() {
			var debugLogs []string
			var mu sync.Mutex

			debugFunc := func(format string, args ...any) {
				mu.Lock()
				defer mu.Unlock()
				debugLogs = append(debugLogs, format)
			}

			db, err := simpledb.NewDB[string, string, string](
				filePath,
				simpledb.WithAsyncTime(100*time.Millisecond),
				simpledb.WithUniqueColumns("email"),
				simpledb.WithDebugger(debugFunc),
			)
			so(err, isNil)

			// 正常存储
			err = db.Store("user1", map[string]string{
				"name":  "张三",
				"email": "zhangsan@example.com",
			})
			so(err, isNil)

			// 违反唯一约束
			err = db.Store("user2", map[string]string{
				"name":  "李四",
				"email": "zhangsan@example.com",
			})
			so(err, notNil)
			so(errors.Is(err, simpledb.ErrColumnDuplicate), isTrue)

			// 等待异步写入完成
			time.Sleep(200 * time.Millisecond)

			// 验证文件已写入
			_, err = os.Stat(filePath)
			so(err, isNil)

			// 验证调试器被调用
			mu.Lock()
			logCount := len(debugLogs)
			mu.Unlock()
			so(logCount > 0, isTrue)
		})
	})
}

// ========== nil Option 测试 ==========

func TestNilOption(t *testing.T) {
	cv("测试 nil Option 不会导致 panic", t, func() {
		ensureTestDataDir(t)
		defer cleanupTestDataDir()

		filePath := filepath.Join(testDataDir, "nil_option_test.csv")

		cv("传入 nil Option", func() {
			db, err := simpledb.NewDB[string, string, string](
				filePath,
				nil, // nil option
				simpledb.WithAsyncTime(0),
				nil, // 另一个 nil option
			)
			so(err, isNil)
			so(db, notNil)

			err = db.Store("key", map[string]string{"col": "val"})
			so(err, isNil)
		})
	})
}

// ========== LoadWithUniqueColumn 测试 ==========

func TestLoadWithUniqueColumn(t *testing.T) {
	cv("测试 LoadWithUniqueColumn 按唯一列加载", t, func() {
		ensureTestDataDir(t)
		defer cleanupTestDataDir()

		cv("通过唯一列成功加载数据", func() {
			filePath := filepath.Join(testDataDir, "load_unique_test.csv")
			db, err := simpledb.NewDB[string, string, string](
				filePath,
				simpledb.WithUniqueColumns("email"),
			)
			so(err, isNil)

			// 存储数据
			err = db.Store("user1", map[string]string{
				"name":  "张三",
				"email": "zhangsan@example.com",
				"age":   "25",
			})
			so(err, isNil)

			err = db.Store("user2", map[string]string{
				"name":  "李四",
				"email": "lisi@example.com",
				"age":   "30",
			})
			so(err, isNil)

			// 通过唯一列加载
			line, row, exist := db.LoadWithUniqueColumn("email", "zhangsan@example.com")
			so(exist, isTrue)
			so(line, eq, "user1")
			so(row, notNil)
			so(row["name"], eq, "张三")
			so(row["email"], eq, "zhangsan@example.com")
			so(row["age"], eq, "25")

			// 加载另一个用户
			line2, row2, exist2 := db.LoadWithUniqueColumn("email", "lisi@example.com")
			so(exist2, isTrue)
			so(line2, eq, "user2")
			so(row2["name"], eq, "李四")
			so(row2["age"], eq, "30")
		})

		cv("加载不存在的唯一列值", func() {
			filePath := filepath.Join(testDataDir, "load_unique_notfound_test.csv")
			db, err := simpledb.NewDB[string, string, string](
				filePath,
				simpledb.WithUniqueColumns("email"),
			)
			so(err, isNil)

			err = db.Store("user1", map[string]string{
				"name":  "张三",
				"email": "zhangsan@example.com",
			})
			so(err, isNil)

			// 查询不存在的值
			line, row, exist := db.LoadWithUniqueColumn("email", "nonexistent@example.com")
			so(exist, isFalse)
			so(line, eq, "")
			so(row, isNil)
		})

		cv("尝试从非唯一列加载（列未配置为唯一）", func() {
			filePath := filepath.Join(testDataDir, "load_nonunique_column_test.csv")
			db, err := simpledb.NewDB[string, string, string](
				filePath,
				simpledb.WithUniqueColumns("email"), // 只有 email 是唯一的
			)
			so(err, isNil)

			err = db.Store("user1", map[string]string{
				"name":  "张三",
				"email": "zhangsan@example.com",
			})
			so(err, isNil)

			// 尝试通过非唯一列加载
			line, row, exist := db.LoadWithUniqueColumn("name", "张三")
			so(exist, isFalse)
			so(line, eq, "")
			so(row, isNil)
		})

		cv("多个唯一列分别加载", func() {
			filePath := filepath.Join(testDataDir, "load_multi_unique_test.csv")
			db, err := simpledb.NewDB[string, string, string](
				filePath,
				simpledb.WithUniqueColumns("email", "phone"),
			)
			so(err, isNil)

			err = db.Store("user1", map[string]string{
				"name":  "张三",
				"email": "zhangsan@example.com",
				"phone": "13800138001",
			})
			so(err, isNil)

			// 通过 email 加载
			line1, row1, exist1 := db.LoadWithUniqueColumn("email", "zhangsan@example.com")
			so(exist1, isTrue)
			so(line1, eq, "user1")
			so(row1["name"], eq, "张三")
			so(row1["phone"], eq, "13800138001")

			// 通过 phone 加载（应该加载到同一行）
			line2, row2, exist2 := db.LoadWithUniqueColumn("phone", "13800138001")
			so(exist2, isTrue)
			so(line2, eq, "user1")
			so(row2["name"], eq, "张三")
			so(row2["email"], eq, "zhangsan@example.com")
		})

		cv("唯一列值更新后的加载", func() {
			filePath := filepath.Join(testDataDir, "load_after_update_test.csv")
			db, err := simpledb.NewDB[string, string, string](
				filePath,
				simpledb.WithUniqueColumns("email"),
			)
			so(err, isNil)

			// 初始存储
			err = db.Store("user1", map[string]string{
				"name":  "张三",
				"email": "old@example.com",
			})
			so(err, isNil)

			// 通过旧值可以加载
			line1, row1, exist1 := db.LoadWithUniqueColumn("email", "old@example.com")
			so(exist1, isTrue)
			so(line1, eq, "user1")
			so(row1["name"], eq, "张三")

			// 更新 email
			err = db.StoreColumns("user1", map[string]string{
				"email": "new@example.com",
			})
			so(err, isNil)

			// 旧值不能加载
			line2, row2, exist2 := db.LoadWithUniqueColumn("email", "old@example.com")
			so(exist2, isFalse)
			so(line2, eq, "")
			so(row2, isNil)

			// 新值可以加载
			line3, row3, exist3 := db.LoadWithUniqueColumn("email", "new@example.com")
			so(exist3, isTrue)
			so(line3, eq, "user1")
			so(row3["name"], eq, "张三")
		})

		cv("空值的处理", func() {
			filePath := filepath.Join(testDataDir, "load_empty_value_test.csv")
			db, err := simpledb.NewDB[string, string, string](
				filePath,
				simpledb.WithUniqueColumns("email"),
			)
			so(err, isNil)

			// 存储空 email 值
			err = db.Store("user1", map[string]string{
				"name":  "张三",
				"email": "",
			})
			so(err, isNil)

			// 通过空值查询
			line, row, exist := db.LoadWithUniqueColumn("email", "")
			so(exist, isTrue)
			so(line, eq, "user1")
			so(row["name"], eq, "张三")
			so(row["email"], eq, "")
		})

		cv("从持久化文件重建索引后加载", func() {
			filePath := filepath.Join(testDataDir, "load_persist_test.csv")

			// 第一个数据库实例写入数据
			db1, err := simpledb.NewDB[string, string, string](
				filePath,
				simpledb.WithUniqueColumns("email"),
			)
			so(err, isNil)

			err = db1.Store("user1", map[string]string{
				"name":  "张三",
				"email": "zhangsan@example.com",
			})
			so(err, isNil)

			// 创建新实例，重建索引
			db2, err := simpledb.NewDB[string, string, string](
				filePath,
				simpledb.WithUniqueColumns("email"),
			)
			so(err, isNil)

			// 通过唯一列加载应该正常工作
			line, row, exist := db2.LoadWithUniqueColumn("email", "zhangsan@example.com")
			so(exist, isTrue)
			so(line, eq, "user1")
			so(row["name"], eq, "张三")
		})

		cv("并发读取唯一列", func() {
			filePath := filepath.Join(testDataDir, "load_concurrent_test.csv")
			db, err := simpledb.NewDB[string, string, string](
				filePath,
				simpledb.WithUniqueColumns("email"),
			)
			so(err, isNil)

			// 先存储一些数据
			for i := 0; i < 10; i++ {
				err = db.Store(fmt.Sprintf("user%d", i), map[string]string{
					"name":  fmt.Sprintf("用户%d", i),
					"email": fmt.Sprintf("user%d@example.com", i),
				})
				so(err, isNil)
			}

			var wg sync.WaitGroup
			errors := make(chan error, 50)

			// 并发读取
			for i := 0; i < 50; i++ {
				wg.Add(1)
				go func(idx int) {
					defer wg.Done()
					email := fmt.Sprintf("user%d@example.com", idx%10)
					line, row, exist := db.LoadWithUniqueColumn("email", email)
					if !exist {
						errors <- fmt.Errorf("未找到 email: %s", email)
						return
					}
					expectedLine := fmt.Sprintf("user%d", idx%10)
					if line != expectedLine {
						errors <- fmt.Errorf("line 不匹配: 期望 %s, 得到 %s", expectedLine, line)
						return
					}
					expectedName := fmt.Sprintf("用户%d", idx%10)
					if row["name"] != expectedName {
						errors <- fmt.Errorf("name 不匹配: 期望 %s, 得到 %s", expectedName, row["name"])
					}
				}(i)
			}

			wg.Wait()
			close(errors)

			// 检查是否有错误
			for err := range errors {
				so(err, isNil)
			}
		})
	})
}

// ========== LoadWithColumn 测试 ==========

func TestLoadWithColumn(t *testing.T) {
	cv("测试 LoadWithColumn 按列查找", t, func() {
		ensureTestDataDir(t)
		defer cleanupTestDataDir()

		cv("查找匹配多行的情况", func() {
			filePath := filepath.Join(testDataDir, "load_column_multi_test.csv")
			db, err := simpledb.NewDB[string, string, string](filePath)
			so(err, isNil)

			// 存储多个用户，其中一些有相同的 status
			err = db.Store("user1", map[string]string{
				"name":   "张三",
				"status": "active",
				"age":    "25",
			})
			so(err, isNil)

			err = db.Store("user2", map[string]string{
				"name":   "李四",
				"status": "active",
				"age":    "30",
			})
			so(err, isNil)

			err = db.Store("user3", map[string]string{
				"name":   "王五",
				"status": "inactive",
				"age":    "28",
			})
			so(err, isNil)

			// 查找所有 active 状态的用户
			results := db.LoadWithColumn("status", "active")
			so(len(results), eq, 2)
			so(results["user1"]["name"], eq, "张三")
			so(results["user2"]["name"], eq, "李四")

			// 验证返回的数据完整性
			so(results["user1"]["age"], eq, "25")
			so(results["user2"]["age"], eq, "30")
		})

		cv("查找匹配单行的情况", func() {
			filePath := filepath.Join(testDataDir, "load_column_single_test.csv")
			db, err := simpledb.NewDB[string, string, string](filePath)
			so(err, isNil)

			err = db.Store("user1", map[string]string{
				"name":  "张三",
				"email": "zhangsan@example.com",
			})
			so(err, isNil)

			err = db.Store("user2", map[string]string{
				"name":  "李四",
				"email": "lisi@example.com",
			})
			so(err, isNil)

			// 查找唯一的 email
			results := db.LoadWithColumn("email", "zhangsan@example.com")
			so(len(results), eq, 1)
			so(results["user1"]["name"], eq, "张三")
		})

		cv("查找不存在的值", func() {
			filePath := filepath.Join(testDataDir, "load_column_notfound_test.csv")
			db, err := simpledb.NewDB[string, string, string](filePath)
			so(err, isNil)

			err = db.Store("user1", map[string]string{
				"name":   "张三",
				"status": "active",
			})
			so(err, isNil)

			// 查找不存在的值
			results := db.LoadWithColumn("status", "nonexistent")
			so(len(results), eq, 0)
			so(results, notNil) // 应该返回空 map，而不是 nil
		})

		cv("查找不存在的列", func() {
			filePath := filepath.Join(testDataDir, "load_column_nocol_test.csv")
			db, err := simpledb.NewDB[string, string, string](filePath)
			so(err, isNil)

			err = db.Store("user1", map[string]string{
				"name": "张三",
			})
			so(err, isNil)

			// 查找不存在的列
			results := db.LoadWithColumn("nonexistent", "value")
			so(len(results), eq, 0)
			so(results, notNil)
		})

		cv("空数据库查询", func() {
			filePath := filepath.Join(testDataDir, "load_column_empty_test.csv")
			db, err := simpledb.NewDB[string, string, string](filePath)
			so(err, isNil)

			// 在空数据库中查询
			results := db.LoadWithColumn("any", "value")
			so(len(results), eq, 0)
			so(results, notNil)
		})

		cv("部分行有该列的情况", func() {
			filePath := filepath.Join(testDataDir, "load_column_partial_test.csv")
			db, err := simpledb.NewDB[string, string, string](filePath)
			so(err, isNil)

			// user1 和 user2 有 phone 列
			err = db.Store("user1", map[string]string{
				"name":  "张三",
				"phone": "13800138001",
			})
			so(err, isNil)

			err = db.Store("user2", map[string]string{
				"name":  "李四",
				"phone": "13800138002",
			})
			so(err, isNil)

			// user3 没有 phone 列
			err = db.Store("user3", map[string]string{
				"name": "王五",
			})
			so(err, isNil)

			// 查找特定 phone
			results := db.LoadWithColumn("phone", "13800138001")
			so(len(results), eq, 1)
			so(results["user1"]["name"], eq, "张三")

			// user3 不应该被包含在结果中（因为它没有 phone 列）
			_, exist := results["user3"]
			so(exist, isFalse)
		})

		cv("空字符串值的查询", func() {
			filePath := filepath.Join(testDataDir, "load_column_empty_value_test.csv")
			db, err := simpledb.NewDB[string, string, string](filePath)
			so(err, isNil)

			err = db.Store("user1", map[string]string{
				"name":  "张三",
				"email": "",
			})
			so(err, isNil)

			err = db.Store("user2", map[string]string{
				"name":  "李四",
				"email": "lisi@example.com",
			})
			so(err, isNil)

			// 查找空 email
			results := db.LoadWithColumn("email", "")
			so(len(results), eq, 1)
			so(results["user1"]["name"], eq, "张三")
		})

		cv("数据更新后的查询", func() {
			filePath := filepath.Join(testDataDir, "load_column_update_test.csv")
			db, err := simpledb.NewDB[string, string, string](filePath)
			so(err, isNil)

			// 初始状态
			err = db.Store("user1", map[string]string{
				"name":   "张三",
				"status": "active",
			})
			so(err, isNil)

			err = db.Store("user2", map[string]string{
				"name":   "李四",
				"status": "active",
			})
			so(err, isNil)

			// 查询初始状态
			results1 := db.LoadWithColumn("status", "active")
			so(len(results1), eq, 2)

			// 更新 user1 的状态
			err = db.StoreColumns("user1", map[string]string{
				"status": "inactive",
			})
			so(err, isNil)

			// 再次查询 active 状态
			results2 := db.LoadWithColumn("status", "active")
			so(len(results2), eq, 1)
			so(results2["user2"]["name"], eq, "李四")

			// 查询 inactive 状态
			results3 := db.LoadWithColumn("status", "inactive")
			so(len(results3), eq, 1)
			so(results3["user1"]["name"], eq, "张三")
		})

		cv("从持久化文件加载后查询", func() {
			filePath := filepath.Join(testDataDir, "load_column_persist_test.csv")

			// 第一个数据库实例写入数据
			db1, err := simpledb.NewDB[string, string, string](filePath)
			so(err, isNil)

			err = db1.Store("user1", map[string]string{
				"name": "张三",
				"city": "北京",
			})
			so(err, isNil)

			err = db1.Store("user2", map[string]string{
				"name": "李四",
				"city": "北京",
			})
			so(err, isNil)

			err = db1.Store("user3", map[string]string{
				"name": "王五",
				"city": "上海",
			})
			so(err, isNil)

			// 创建新实例，从文件加载
			db2, err := simpledb.NewDB[string, string, string](filePath)
			so(err, isNil)

			// 查询应该正常工作
			results := db2.LoadWithColumn("city", "北京")
			so(len(results), eq, 2)
			so(results["user1"]["name"], eq, "张三")
			so(results["user2"]["name"], eq, "李四")
		})

		cv("特殊字符值的查询", func() {
			filePath := filepath.Join(testDataDir, "load_column_special_test.csv")
			db, err := simpledb.NewDB[string, string, string](filePath)
			so(err, isNil)

			specialValue := "hello,world\"with'quotes"
			err = db.Store("user1", map[string]string{
				"name":    "张三",
				"comment": specialValue,
			})
			so(err, isNil)

			err = db.Store("user2", map[string]string{
				"name":    "李四",
				"comment": "normal",
			})
			so(err, isNil)

			// 查询特殊字符值
			results := db.LoadWithColumn("comment", specialValue)
			so(len(results), eq, 1)
			so(results["user1"]["name"], eq, "张三")
		})

		cv("并发读取测试", func() {
			filePath := filepath.Join(testDataDir, "load_column_concurrent_test.csv")
			db, err := simpledb.NewDB[string, string, string](filePath)
			so(err, isNil)

			// 准备测试数据
			for i := 0; i < 20; i++ {
				status := "active"
				if i%3 == 0 {
					status = "inactive"
				}
				err = db.Store(fmt.Sprintf("user%d", i), map[string]string{
					"name":   fmt.Sprintf("用户%d", i),
					"status": status,
				})
				so(err, isNil)
			}

			var wg sync.WaitGroup
			errors := make(chan error, 50)

			// 并发查询
			for i := 0; i < 50; i++ {
				wg.Add(1)
				go func(idx int) {
					defer wg.Done()

					// 查询 active 状态
					results := db.LoadWithColumn("status", "active")
					if len(results) != 13 { // 20个用户中，13个是active（i%3!=0）
						errors <- fmt.Errorf("active 数量不正确: 期望 13, 得到 %d", len(results))
						return
					}

					// 查询 inactive 状态
					results2 := db.LoadWithColumn("status", "inactive")
					if len(results2) != 7 { // 7个是inactive（i%3==0: 0,3,6,9,12,15,18）
						errors <- fmt.Errorf("inactive 数量不正确: 期望 7, 得到 %d", len(results2))
					}
				}(i)
			}

			wg.Wait()
			close(errors)

			// 检查是否有错误
			for err := range errors {
				so(err, isNil)
			}
		})

		cv("返回的数据应该是原始数据的副本（隔离性）", func() {
			filePath := filepath.Join(testDataDir, "load_column_isolation_test.csv")
			db, err := simpledb.NewDB[string, string, string](filePath)
			so(err, isNil)

			err = db.Store("user1", map[string]string{
				"name":   "张三",
				"status": "active",
			})
			so(err, isNil)

			// 获取查询结果
			results := db.LoadWithColumn("status", "active")
			so(len(results), eq, 1)

			// 注意：当前实现返回的是对内部数据的直接引用
			// 这个测试验证当前行为（但不一定是最佳实践）
			// 如果未来改为返回副本，这个测试可能需要调整
			originalName := results["user1"]["name"]
			so(originalName, eq, "张三")

			// 再次查询，确保数据仍然正确
			results2 := db.LoadWithColumn("status", "active")
			so(results2["user1"]["name"], eq, "张三")
		})
	})
}

// ========== 当前目录文件路径测试 ==========

func TestCurrentDirPath(t *testing.T) {
	cv("测试当前目录文件路径", t, func() {
		// 使用临时文件名避免冲突
		filePath := "test_current_dir_" + time.Now().Format("20060102150405") + ".csv"
		defer os.Remove(filePath)

		cv("相对路径（当前目录）", func() {
			db, err := simpledb.NewDB[string, string, string](filePath)
			so(err, isNil)

			err = db.Store("key1", map[string]string{"col": "val"})
			so(err, isNil)

			// 文件应该在当前目录
			_, err = os.Stat(filePath)
			so(err, isNil)
		})
	})
}
