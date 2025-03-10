package holiday_test

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Andrew-M-C/go.util/china/holiday"
	"github.com/fatih/color"
	"github.com/smartystreets/goconvey/convey"
)

var (
	cv = convey.Convey
	so = convey.So
	eq = convey.ShouldEqual
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func Test2024(t *testing.T) {
	cv("2024", t, func() {
		for i := 1; i <= 12; i++ {
			m := newMonthCalendar(2024, i)
			t.Logf("<< %v >>\n%v", time.Month(i), m)
		}
	})

	cv("2025", t, func() {
		for i := 1; i <= 12; i++ {
			m := newMonthCalendar(2025, i)
			t.Logf("<< %v >>\n%v", time.Month(i), m)
		}
	})

	cv("自定义类型", t, func() {
		const vocation = holiday.DayType(holiday.ShiftedWorkday + 1)
		holiday.AddNewDayType(vocation, "请假")

		// 首先除夕不是假期
		d := holiday.Date(2024, 2, 9)
		so(d.Type(), eq, holiday.Workday)

		// 然后请假, 就标记上了
		holiday.AddSpecialDay(d, vocation, "年休假")
		so(d.Type(), eq, vocation)
		so(d.Description(), eq, "年休假")
	})
}

type monthCalendar struct {
	days [6][7]string
	desc [6][7]string
}

func newMonthCalendar(year, month int) *monthCalendar {
	c := new(monthCalendar)
	line, col := 0, 0

	// 第一行和最后一行都先填充
	fillLine := func(line int) {
		for i := range c.desc[line] {
			c.days[line][i] = "  "
		}
	}
	fillLine(0)
	fillLine(4)
	fillLine(5)

	// 首先找出当月第一天是星期几
	firstDay := holiday.Date(year, month, 1)
	col = int(firstDay.Weekday())

	// 填充
	for d := firstDay; d.Month() == firstDay.Month(); d = d.AddDate(0, 0, 1) {
		var formatter func(string, ...any) string
		switch d.Type() {
		case holiday.Workday:
			formatter = fmt.Sprintf
		case holiday.ShiftedWorkday:
			formatter = fmt.Sprintf
		default:
			formatter = color.RedString
		}
		c.days[line][col] = formatter("%02d", d.Day())
		c.desc[line][col] = formatter("%s", d.Description())

		col++
		if col >= 7 {
			line++
			col = 0
		}
	}

	return c
}

func (c *monthCalendar) String() string {
	lines := []string{
		fmt.Sprintf("|-%s-|-MO-|-TU-|-WE-|-TH-|-FR-|-%s-|",
			color.RedString("SU"), color.RedString("SA"),
		),
	}
	for i, line := range c.days {
		if i > 0 {
			if strings.TrimSpace(line[0]) == "" {
				continue
			}
		}
		lines = append(lines, "| "+strings.Join(line[:], " | ")+" |")
	}
	return strings.Join(lines, "\n")
}
