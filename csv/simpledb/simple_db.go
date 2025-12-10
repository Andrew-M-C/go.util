// Package simpledb 实现一个基于 CSV 的极简 KV 数据库
package simpledb

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/Andrew-M-C/go.util/csv"
)

// DB 实现基于 CSV 的极简 KV 数据库
type DB[LINE ~string, COL ~string, V ~string] struct {
	*options

	lock     sync.RWMutex
	data     map[LINE]map[COL]V
	filePath string

	uniqueColumns map[COL]map[V]LINE
	columnSeqs    []COL

	// 是否正在等待存入文件
	waitingSavingFile bool
}

// NewDB 新建一个 *simpledb.DB
func NewDB[LINE ~string, COL ~string, V ~string](filePath string, opts ...Option) (*DB[LINE, COL, V], error) {
	db := &DB[LINE, COL, V]{
		options:  mergeOptions(opts),
		data:     map[LINE]map[COL]V{},
		filePath: filePath,
	}
	// 初始化 unique columns
	db.uniqueColumns = make(map[COL]map[V]LINE, len(db.options.uniqueColumns))
	for col := range db.options.uniqueColumns {
		db.uniqueColumns[COL(col)] = make(map[V]LINE)
	}

	if err := db.ensureDir(); err != nil {
		return nil, err
	}

	// 读取已有文件数据
	if err := db.readFile(); err != nil {
		return nil, err
	}

	// 初始化唯一索引
	db.initUniqueIndex()

	return db, nil
}

// Load 加载一个数据
func (db *DB[LINE, COL, V]) Load(key LINE) (map[COL]V, bool) {
	db.lock.RLock()
	defer db.lock.RUnlock()

	row, exist := db.data[key]
	return row, exist
}

// LoadWithUniqueColumn 按照唯一键加载数据
func (db *DB[LINE, COL, V]) LoadWithUniqueColumn(column COL, value V) (LINE, map[COL]V, bool) {
	db.lock.RLock()
	defer db.lock.RUnlock()

	row, exist := db.uniqueColumns[column][value]
	if !exist {
		return "", nil, false
	}

	res, exist := db.data[row]
	return row, res, exist
}

// LoadWithColumn 按照指定列和值查找所有匹配的行
func (db *DB[LINE, COL, V]) LoadWithColumn(column COL, value V) map[LINE]map[COL]V {
	db.lock.RLock()
	defer db.lock.RUnlock()

	res := map[LINE]map[COL]V{}

	for line, row := range db.data {
		if v, exist := row[column]; exist && v == value {
			res[line] = row
		}
	}
	return res
}

// Store 存储一个数据
func (db *DB[LINE, COL, V]) Store(key LINE, value map[COL]V) error {
	if key == "" {
		return ErrEmptyLineKey
	}

	db.lock.Lock()
	defer db.lock.Unlock()

	if err := db.storeColumnsLocked(key, value); err != nil {
		return err
	}

	db.writeToFileOrWait()
	return nil
}

// StoreColumns 存储一组列, 与 Store 的区别是只更新指定行的列，而不是整行替换
func (db *DB[LINE, COL, V]) StoreColumns(key LINE, columns map[COL]V) error {
	if key == "" {
		return ErrEmptyLineKey
	}
	if len(columns) == 0 {
		return nil // 没有需要更新的列，直接返回
	}

	db.lock.Lock()
	defer db.lock.Unlock()

	if err := db.storeColumnsLocked(key, columns); err != nil {
		return err
	}

	db.writeToFileOrWait()
	return nil
}

// storeColumnsLocked 内部方法，在已持有锁的情况下更新列数据
// 调用者需要确保已持有 db.lock
func (db *DB[LINE, COL, V]) storeColumnsLocked(key LINE, columns map[COL]V) error {
	// 检查 unique
	if len(db.uniqueColumns) > 0 {
		for col, val := range columns {
			uniques, exist := db.uniqueColumns[col]
			if !exist {
				db.debugf("列 '%v' 无需检查唯一键, 跳过", col)
				continue
			}
			prevLine := uniques[val]
			if prevLine != "" && prevLine != key {
				return fmt.Errorf("%w: 列 '%v' = '%v' 已属于 '%v', 无法重复赋值",
					ErrColumnDuplicate, col, val, prevLine,
				)
			}
			db.debugf("列 '%s' 值 '%v' 依然属于 '%v', OK", col, val, key)
		}
	}

	// 获取或创建该行的数据
	row, exist := db.data[key]
	if !exist {
		row = make(map[COL]V)
		db.data[key] = row
	}

	// 更新列数据，并维护唯一索引
	for col, val := range columns {
		// 处理唯一索引
		if uniques, isUnique := db.uniqueColumns[col]; isUnique {
			// 先删除旧值的索引
			if oldVal, hasOld := row[col]; hasOld && oldVal != val {
				delete(uniques, oldVal)
			}
			// 添加新值的索引
			uniques[val] = key
		}

		// 更新数据
		row[col] = val

		// 检查是否需要添加新列到 columnSeqs
		db.addColumnIfNotExist(col)
	}

	return nil
}

// -------- 写入文件 --------

func (db *DB[LINE, COL, V]) writeToFileOrWait() {
	if db.asyncTime <= 0 {
		db.writeToFileSync()
		return
	}
	if db.waitingSavingFile {
		return // 已经有 goroutine 在等待写入文件了
	}

	db.waitingSavingFile = true

	go func() {
		time.Sleep(db.asyncTime)

		db.lock.Lock()
		defer db.lock.Unlock()

		db.writeToFileSync()
		db.waitingSavingFile = false
	}()
}

func (db *DB[LINE, COL, V]) writeToFileSync() {
	if len(db.data) == 0 {
		db.debugf("数据为空, 跳过写入文件")
		return
	}

	b, err := csv.WriteCSVStringMaps(db.data, db.columnSeqs)
	if err != nil {
		db.debugf("序列化 CSV 数据失败: %v", err)
		return
	}

	if err := os.WriteFile(db.filePath, b, 0644); err != nil {
		db.debugf("写入文件 '%s' 失败: %v", db.filePath, err)
		return
	}

	db.debugf("成功写入文件 '%s', 共 %d 行", db.filePath, len(db.data))
}

// -------- 初始化方法 --------

func (db *DB[LINE, COL, V]) ensureDir() error {
	dir := filepath.Dir(db.filePath)
	if dir == "" || dir == "." {
		return nil
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("检查目录失败 (%w)", err)
	}
	return nil
}

func (db *DB[LINE, COL, V]) readFile() error {
	b, err := os.ReadFile(db.filePath)
	if err != nil {
		if isFileNotExistErr(err) {
			db.debugf("读取文件 '%s' 失败 (%v), 视为初始化", db.filePath, err)
			return nil // 前面已经初始化了数据了
		}
		return fmt.Errorf("读取文件失败 (%w)", err)
	}

	db.data, db.columnSeqs, err = csv.ReadCSVStringMaps[LINE, COL, V](b)
	if err != nil {
		return fmt.Errorf("解析文件失败 (%w)", err)
	}
	db.debugf("成功读取文件 '%s', 共 %d 行, %d 列", db.filePath, len(db.data), len(db.columnSeqs))
	return nil
}

// initUniqueIndex 初始化唯一索引
func (db *DB[LINE, COL, V]) initUniqueIndex() {
	for lineKey, row := range db.data {
		for col, val := range row {
			if uniques, isUnique := db.uniqueColumns[col]; isUnique {
				uniques[val] = lineKey
			}
		}
	}
}

// addColumnIfNotExist 添加列到 columnSeqs（如果不存在）
func (db *DB[LINE, COL, V]) addColumnIfNotExist(col COL) {
	for _, c := range db.columnSeqs {
		if c == col {
			return // 已存在
		}
	}
	db.columnSeqs = append(db.columnSeqs, col)
}

func isFileNotExistErr(err error) bool {
	if os.IsNotExist(err) {
		return true
	}
	return strings.Contains(err.Error(), "exist")
}
