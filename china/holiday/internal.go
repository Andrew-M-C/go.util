package holiday

import (
	"time"

	syncutil "github.com/Andrew-M-C/go.util/sync"
)

var (
	beijing = time.FixedZone("Asia/Beijing", 8*60*60)
)

type date struct {
	year  uint16
	month uint8
	day   uint8

	typ  DayType
	desc string
}

func (d date) key() uint32 {
	return (uint32(d.year) << 16) + (uint32(d.month) << 8) + uint32(d.day)
}

func newDate(year int, month time.Month, day int) date {
	d := date{}

	d.year = uint16(year)
	d.month = uint8(month)
	d.day = uint8(day)

	return d
}

func (d date) withType(typ DayType) date {
	d.typ = typ
	return d
}

func (d date) withName(name string) date {
	switch d.typ {
	case Holiday:
		d.desc = name
	case HolidayPeriod:
		d.desc = name + "假期"
	case ShiftedWorkday:
		d.desc = name + "调班"
	case ShiftedDayOff:
		d.desc = name + "调休"
	default:
		d.desc = name
	}
	return d
}

func (d date) add() {
	internal.specialDates.Store(d.key(), d)
}

var internal = struct {
	specialDates syncutil.Map[uint32, date]
	dayTypeDesc  map[DayType]string
}{
	specialDates: syncutil.NewMap[uint32, date](),
	dayTypeDesc: map[DayType]string{
		UnknownType:    "未知类型",
		Workday:        "工作日",
		Weekend:        "周末",
		Holiday:        "节日",
		HolidayPeriod:  "节日假期",
		ShiftedDayOff:  "调休休息",
		ShiftedWorkday: "调休上班",
	},
}
