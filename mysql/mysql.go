// Package mysql 提供一些使用 Go 操作 MySQL 时的帮助工具
package mysql

import (
	"time"
)

// TimezoneCorrector 根据解析后的 DB 数据, 修正一个时间所包含的时区信息
//
// WARNING: beta use ONLY
type TimezoneCorrector interface {
	CorrectTimezoneFromDB(in time.Time) time.Time
	CorrectTimezoneToDB(in time.Time) time.Time
}
