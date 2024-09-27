// Package holiday 实现中国假期和调休统计逻辑。仅准确支持 2024 年及以后。
package holiday

import "time"

// MARK: 公开定义

// Day 表示一天
type Day struct {
	time.Time
}

// DayType 表示这一天所属的类型
type DayType int

const (
	UnknownType DayType = iota
	// 普通的工作日
	Workday
	// 普通的周末, 正常放假
	Weekend
	// 法定节日当天
	Holiday
	// 法定节日假期, 但不是法定节日当天
	HolidayPeriod
	// 某一天按照日历是工作日, 但因为调休, 变成了休息日
	ShiftedDayOff
	// 某一天按照日历是周末, 但因为调休, 变成了工作日
	ShiftedWorkday
)

const (
	// Weekday 等同于 Workday
	Weekday = Workday
)

func (t DayType) String() string {
	switch t {
	default:
		fallthrough
	case UnknownType:
		return "未知日期类型"
	case Workday:
		return "工作日"
	case Weekend:
		return "周末"
	case Holiday:
		return "节日"
	case HolidayPeriod:
		return "节日假期"
	case ShiftedDayOff:
		return "调休休息"
	case ShiftedWorkday:
		return "调休上班"
	}
}

// Today 表示今天
func Today() Day {
	tm := time.Now().In(beijing)
	return Day{Time: tm}
}

// 从一个时间提取出日期。
func DayOfTime(tm time.Time) Day {
	return Day{Time: tm.In(beijing)}
}

// Integer 表示任意整数
type Integer interface {
	~int | ~uint |
		~int8 | ~int16 | ~int32 | ~int64 |
		~uint8 | ~uint16 | ~uint32 | ~uint64
}

// Date 表示某一天
func Date[T Integer](year int, month T, day int) Day {
	tm := time.Date(year, time.Month(month), day, 0, 0, 0, 0, beijing)
	return Day{Time: tm}
}

func (d Day) String() string {
	return d.In(beijing).Format(time.DateOnly)
}

// Type 返回这一天的类型
func (d Day) Type() DayType {
	// 如果是特殊日子
	if da, exist := internal.specialDates[d.key()]; exist {
		return da.typ
	}
	// 不是特殊日子的话, 那就看是周中还是周末
	switch d.Time.Weekday() {
	case time.Saturday, time.Sunday:
		return Weekend
	default:
		return Workday
	}
}

// Description 描述, 比如: 工作日 / 周末 / 国庆调休放假 / 国庆调休上班
func (d Day) Description() string {
	// 如果今天是特殊日子
	if da, exist := internal.specialDates[d.key()]; exist {
		return da.desc
	}
	return d.Type().String()
}

// MARK: 继承 time.Time 的方法

// AddDate 重载 time.Time 的 AddDate 方法, 但返回 holiday.Day 类型
func (d Day) AddDate(years, months, days int) Day {
	tm := d.Time.AddDate(years, months, days)
	return Day{Time: tm}
}

// MARK: 内部方法

func (d Day) key() uint32 {
	tm := d.In(beijing)
	y, m, da := tm.Year(), int(tm.Month()), tm.Day()
	k := (y << 16) + (m << 8) + da
	return uint32(k)
}
