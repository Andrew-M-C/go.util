package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"
)

type mysqlTiming struct {
	OneDayTx int64 `db:"one_day_ts"`
}

func (t mysqlTiming) toTzCorrector() (tz tzCorrector) {
	const halfHour = 30 * 60
	const oneDay = 24 * 60 * 60
	tz = tzCorrector(t.OneDayTx)
	defer func() {
		fmt.Println("DB 的时间偏移是: ", tz)
	}()
	switch {
	default:
		return tz
	case tz > 0:
		return (tz+1)/halfHour*halfHour - oneDay
	case tz < 0:
		return (tz-1)/halfHour*halfHour - oneDay
	}
}

type tzCorrector int64

func (tz tzCorrector) CorrectTimezoneFromDB(in time.Time) time.Time {
	usec := in.UnixMicro()

	// 这个时候直接转出来, 就是 DB 存储的 DATETIME 字面值, 但是我们还没计算数据库中的时区
	// fmt.Println("从 DB 中拿到的时间", in)
	usec += int64(tz * 1000 * 1000)
	return time.UnixMicro(usec).Local()
}

func (tz tzCorrector) CorrectTimezoneToDB(in time.Time) time.Time {
	name := strconv.FormatInt(int64(-tz), 10)
	zone := time.FixedZone(name, int(-tz))
	return in.In(zone)
}

// 通过这条指令, 可以得知 UNIX 开始时间的第二天, DB 实际对应的时间戳是什么。使用第二天避免了
// UNIX_TIMESTAMP 返回 0 的情况。
const mysqlTimingStatement = "SELECT UNIX_TIMESTAMP('1970-01-02 00:00:00') AS `one_day_ts`"

func NewTimezoneCorrectorBySelector(ctx context.Context, selector Selector) (TimezoneCorrector, error) {
	var res []mysqlTiming
	if err := selector.SelectContext(ctx, &res, mysqlTimingStatement); err != nil {
		return nil, fmt.Errorf("reading MySQL timing error: '%w'", err)
	}
	if len(res) == 0 {
		return nil, errors.New("cannot read MySQL timing which returns empty")
	}

	return res[0].toTzCorrector(), nil
}

type Selector interface {
	SelectContext(ctx context.Context, dest any, query string, args ...any) error
}

func NewTimezoneCorrectorByQueryer(ctx context.Context, queryer Queryer) (TimezoneCorrector, error) {
	rows, err := queryer.QueryContext(ctx, mysqlTimingStatement)
	if err != nil {
		return nil, fmt.Errorf("reading MySQL timing error: '%w'", err)
	}
	if !rows.Next() {
		return nil, errors.New("cannot read MySQL timing which returns empty")
	}
	tm := mysqlTiming{}
	if err := rows.Scan(&tm.OneDayTx); err != nil {
		return nil, fmt.Errorf("scaning MySQL timing result error: '%w'", err)
	}
	return tm.toTzCorrector(), nil
}

type Queryer interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}
