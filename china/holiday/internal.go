package holiday

import "time"

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
		d.desc = name + "调休上班"
	case ShiftedDayOff:
		d.desc = name + "调休休息"
	default:
		d.desc = name
	}
	return d
}

func (d date) add() {
	internal.specialDates[d.key()] = d
}

var internal = struct {
	specialDates map[uint32]date
}{
	specialDates: map[uint32]date{},
}
