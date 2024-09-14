package time

import (
	"bytes"
	"fmt"
	"time"
)

// Age 表示年龄
type Age struct {
	Years  int
	Months int
	Days   int
}

func (a Age) String() string {
	return a.DetailedDesc()
}

// DetailedDesc 详细描述
func (a Age) DetailedDesc() string {
	if a.IsZero() {
		return "未出生"
	}
	buff := bytes.Buffer{}
	if a.Years > 0 {
		buff.WriteString(fmt.Sprintf("%d岁", a.Years))
	}
	if a.Months > 0 {
		buff.WriteString(fmt.Sprintf("%d个月", a.Months))
	}
	if a.Days > 0 {
		if a.Years > 0 && a.Months <= 0 {
			buff.WriteString(fmt.Sprintf("零%d天", a.Days))
		} else {
			buff.WriteString(fmt.Sprintf("%d天", a.Days))
		}
	}
	return buff.String()
}

func (a Age) IsZero() bool {
	return a.Years <= 0 && a.Months <= 0 && a.Days <= 0
}

// CalculateAge 计算年龄。如果目标时间未到, 则统一返回 0
func CalculateAge(birthday time.Time, to ...time.Time) Age {
	until := time.Now()
	if len(to) > 0 {
		until = to[0]
	}
	if until.Before(birthday) {
		return Age{}
	}

	a := Age{}

	// 计算日, 如果日小于零则需要借位
	a.Days = until.Day() - birthday.Day()
	if a.Days < 0 {
		d := daysOfMonth(until)
		a.Days += d
		until = until.AddDate(0, 0, -d)
	}

	// 计算月差, 如果月小于零则需要借位
	a.Months = int(until.Month() - birthday.Month())
	if a.Months < 0 {
		a.Months += 12
		dYear := daysIfYear(until)
		until = until.AddDate(-0, 0, -dYear)
	}

	a.Years = until.Year() - birthday.Year()
	return a
}

func daysOfMonth(tm time.Time) int {
	lastDayOfMonth := time.Date(tm.Year(), tm.Month(), 1, 0, 0, 0, 0, Beijing).
		AddDate(0, 0, -1)
	return lastDayOfMonth.Day()
}

func daysIfYear(tm time.Time) int {
	targetYear := tm
	if tm.Month() <= 2 {
		// 计算上一年的平闰
		targetYear = tm.AddDate(-1, 0, 0)
	}
	daysfFebruary := time.Date(targetYear.Year(), 3, 1, 0, 0, 0, 0, Beijing).
		AddDate(0, 0, -1).
		Day()
	if daysfFebruary == 29 {
		return 366
	}
	return 365
}
