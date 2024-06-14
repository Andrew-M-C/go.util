package holiday

import (
	"testing"
	"time"

	"github.com/smartystreets/goconvey/convey"
)

var (
	cv = convey.Convey
	so = convey.So
	eq = convey.ShouldEqual
	ne = convey.ShouldNotEqual
)

var (
	wdays = []string{
		"星期日",
		"星期一",
		"星期二",
		"星期三",
		"星期四",
		"星期五",
		"星期六",
	}
)

func TestHoliday(t *testing.T) {
	cv("测试 2022 年节假日", t, func() { test2022(t) })
}

func test2022(t *testing.T) {
	cv("今天", func() {
		d := today()
		now := time.Now().In(internal.china)
		i, typ := NextHoliday()
		if i == 0 {
			t.Logf("今天是: %v, %v, %v", d.Format("2006-01-02"), wdays[d.Weekday()], typ)
		} else {
			t.Logf("今天是: %v, %v, 工作日", d.Format("2006-01-02"), wdays[d.Weekday()])
		}

		so(d.Year(), eq, now.Year())
		so(d.Month(), eq, now.Month())
		so(d.Day(), eq, now.Day())
	})

	// cv("全量打印", func() {
	// 	days := readYearData(2022)
	// 	for i, d := range days {
	// 		t.Logf("%03d - %s", i, d)
	// 	}
	// })

	cv("假日列表", func() {
		firstDay := Day(2022, 1, 1)

		for i := 0; i < 365+7; i++ {
			day := firstDay.AddDate(0, 0, i)
			off, name := NextHolidayForDate(day)
			dayDesc := day.Format("01月02日")
			if off == 0 {
				t.Logf("%v, %v, %v", dayDesc, wdays[day.Weekday()], name)
				off := NextWorkdayForDate(day)
				so(off, ne, 0)
			} else {
				t.Logf("%v, %v, 上班, %d天后是%v", dayDesc, wdays[day.Weekday()], off, name)
				off := NextWorkdayForDate(day)
				so(off, eq, 0)
			}
		}
	})

	cv("Override", func() {
		so(NextWorkdayForDate(Day(2022, 1, 30)), eq, 0)
		so(NextWorkdayForDate(Day(2022, 2, 7)), eq, 0)

		Override(Day(2022, 1, 30), "春节假期")
		Override(Day(2022, 2, 7), "春节假期")

		day := Day(2022, 1, 30)
		for i := 0; i < 9; i++ {
			i, desc := NextHolidayForDate(day)
			so(i, eq, 0)
			so(desc, eq, "春节假期")
			day = day.AddDate(0, 0, 1)
		}
	})
}
