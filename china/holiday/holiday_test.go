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

	doYear := func(year int) {
		for i := 1; i <= 12; i++ {
			m := newMonthCalendar(year, i)
			t.Logf("<< %v >>\n%v", time.Month(i), m)
		}
	}

	y := time.Now().Year()
	for i := 2024; i <= y; i++ {
		cv(fmt.Sprintf("%d", i), t, func() {
			doYear(i)
		})
	}

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

// TestIsRestDay 测试 IsRestDay 方法，覆盖所有类型的休息日
func TestIsRestDay(t *testing.T) {
	cv("IsRestDay 覆盖所有休息日类型", t, func() {
		// 测试 Weekend（普通周末）
		cv("测试 Weekend 类型", func() {
			// 2024年1月6日是周六
			weekend := holiday.Date(2024, 1, 6)
			so(weekend.Type(), eq, holiday.Weekend)
			so(weekend.IsRestDay(), eq, true)
			t.Logf("Weekend: %s (%s) - IsRestDay: %v", weekend, weekend.Description(), weekend.IsRestDay())

			// 2024年1月7日是周日
			sunday := holiday.Date(2024, 1, 7)
			so(sunday.Type(), eq, holiday.Weekend)
			so(sunday.IsRestDay(), eq, true)
			t.Logf("Weekend: %s (%s) - IsRestDay: %v", sunday, sunday.Description(), sunday.IsRestDay())
		})

		// 测试 Holiday（法定节日当天）
		cv("测试 Holiday 类型", func() {
			// 2024年1月1日是元旦
			newYear := holiday.Date(2024, 1, 1)
			so(newYear.Type(), eq, holiday.Holiday)
			so(newYear.IsRestDay(), eq, true)
			t.Logf("Holiday: %s (%s) - IsRestDay: %v", newYear, newYear.Description(), newYear.IsRestDay())

			// 2024年5月1日是劳动节
			laborDay := holiday.Date(2024, 5, 1)
			so(laborDay.Type(), eq, holiday.Holiday)
			so(laborDay.IsRestDay(), eq, true)
			t.Logf("Holiday: %s (%s) - IsRestDay: %v", laborDay, laborDay.Description(), laborDay.IsRestDay())

			// 2024年10月1日是国庆节
			nationalDay := holiday.Date(2024, 10, 1)
			so(nationalDay.Type(), eq, holiday.Holiday)
			so(nationalDay.IsRestDay(), eq, true)
			t.Logf("Holiday: %s (%s) - IsRestDay: %v", nationalDay, nationalDay.Description(), nationalDay.IsRestDay())
		})

		// 测试 HolidayPeriod（法定节日假期但不是当天）
		cv("测试 HolidayPeriod 类型", func() {
			// 2024年2月11日是春节假期
			springFestivalPeriod := holiday.Date(2024, 2, 11)
			so(springFestivalPeriod.Type(), eq, holiday.HolidayPeriod)
			so(springFestivalPeriod.IsRestDay(), eq, true)
			t.Logf("HolidayPeriod: %s (%s) - IsRestDay: %v",
				springFestivalPeriod, springFestivalPeriod.Description(), springFestivalPeriod.IsRestDay())

			// 2024年10月2日是国庆假期
			nationalDayPeriod := holiday.Date(2024, 10, 2)
			so(nationalDayPeriod.Type(), eq, holiday.HolidayPeriod)
			so(nationalDayPeriod.IsRestDay(), eq, true)
			t.Logf("HolidayPeriod: %s (%s) - IsRestDay: %v",
				nationalDayPeriod, nationalDayPeriod.Description(), nationalDayPeriod.IsRestDay())

			// 2025年5月2日是劳动节假期
			laborDayPeriod := holiday.Date(2025, 5, 2)
			so(laborDayPeriod.Type(), eq, holiday.HolidayPeriod)
			so(laborDayPeriod.IsRestDay(), eq, true)
			t.Logf("HolidayPeriod: %s (%s) - IsRestDay: %v",
				laborDayPeriod, laborDayPeriod.Description(), laborDayPeriod.IsRestDay())
		})

		// 测试 ShiftedDayOff（调休假日）
		cv("测试 ShiftedDayOff 类型", func() {
			// 2024年2月13日是春节调休假日（周二）
			shiftedDayOff1 := holiday.Date(2024, 2, 13)
			so(shiftedDayOff1.Type(), eq, holiday.ShiftedDayOff)
			so(shiftedDayOff1.IsRestDay(), eq, true)
			t.Logf("ShiftedDayOff: %s (%s) - IsRestDay: %v",
				shiftedDayOff1, shiftedDayOff1.Description(), shiftedDayOff1.IsRestDay())

			// 2024年5月2日是劳动节调休假日
			shiftedDayOff2 := holiday.Date(2024, 5, 2)
			so(shiftedDayOff2.Type(), eq, holiday.ShiftedDayOff)
			so(shiftedDayOff2.IsRestDay(), eq, true)
			t.Logf("ShiftedDayOff: %s (%s) - IsRestDay: %v",
				shiftedDayOff2, shiftedDayOff2.Description(), shiftedDayOff2.IsRestDay())

			// 2025年10月7日是国庆节调休假日
			shiftedDayOff3 := holiday.Date(2025, 10, 7)
			so(shiftedDayOff3.Type(), eq, holiday.ShiftedDayOff)
			so(shiftedDayOff3.IsRestDay(), eq, true)
			t.Logf("ShiftedDayOff: %s (%s) - IsRestDay: %v",
				shiftedDayOff3, shiftedDayOff3.Description(), shiftedDayOff3.IsRestDay())
		})

		// 测试非休息日类型
		cv("测试非休息日类型", func() {
			// Workday 普通工作日
			workday := holiday.Date(2024, 1, 2) // 周二
			so(workday.Type(), eq, holiday.Workday)
			so(workday.IsRestDay(), eq, false)
			t.Logf("Workday: %s (%s) - IsRestDay: %v", workday, workday.Description(), workday.IsRestDay())

			// ShiftedWorkday 调休上班日
			shiftedWorkday := holiday.Date(2024, 2, 4) // 春节调休上班
			so(shiftedWorkday.Type(), eq, holiday.ShiftedWorkday)
			so(shiftedWorkday.IsRestDay(), eq, false)
			t.Logf("ShiftedWorkday: %s (%s) - IsRestDay: %v",
				shiftedWorkday, shiftedWorkday.Description(), shiftedWorkday.IsRestDay())
		})
	})
}

// TestDayFromTime 测试使用 DayFromTime 转换 time.Time 后判断是否是休息日
func TestDayFromTime(t *testing.T) {
	cv("使用 DayFromTime 测试 IsRestDay", t, func() {
		// 测试 Weekend（普通周末）
		cv("测试 Weekend 类型", func() {
			// 2024年1月6日是周六
			weekend := time.Date(2024, 1, 6, 15, 30, 0, 0, holiday.BeijingZone())
			day := holiday.DayFromTime(weekend)
			so(day.IsRestDay(), eq, true)
			t.Logf("Weekend: %s -> Day: %s (%s) - IsRestDay: %v",
				weekend.Format(time.DateTime), day, day.Description(), day.IsRestDay())

			// 2024年1月7日是周日
			sunday := time.Date(2024, 1, 7, 10, 0, 0, 0, holiday.BeijingZone())
			day = holiday.DayFromTime(sunday)
			so(day.IsRestDay(), eq, true)
			t.Logf("Weekend: %s -> Day: %s (%s) - IsRestDay: %v",
				sunday.Format(time.DateTime), day, day.Description(), day.IsRestDay())
		})

		// 测试 Holiday（法定节日当天）
		cv("测试 Holiday 类型", func() {
			// 2024年1月1日是元旦
			newYear := time.Date(2024, 1, 1, 0, 0, 0, 0, holiday.BeijingZone())
			day := holiday.DayFromTime(newYear)
			so(day.IsRestDay(), eq, true)
			t.Logf("Holiday: %s -> Day: %s (%s) - IsRestDay: %v",
				newYear.Format(time.DateTime), day, day.Description(), day.IsRestDay())

			// 2024年5月1日是劳动节
			laborDay := time.Date(2024, 5, 1, 12, 0, 0, 0, holiday.BeijingZone())
			day = holiday.DayFromTime(laborDay)
			so(day.IsRestDay(), eq, true)
			t.Logf("Holiday: %s -> Day: %s (%s) - IsRestDay: %v",
				laborDay.Format(time.DateTime), day, day.Description(), day.IsRestDay())
		})

		// 测试 HolidayPeriod（法定节日假期但不是当天）
		cv("测试 HolidayPeriod 类型", func() {
			// 2024年2月11日是春节假期
			springFestivalPeriod := time.Date(2024, 2, 11, 8, 30, 0, 0, holiday.BeijingZone())
			day := holiday.DayFromTime(springFestivalPeriod)
			so(day.IsRestDay(), eq, true)
			t.Logf("HolidayPeriod: %s -> Day: %s (%s) - IsRestDay: %v",
				springFestivalPeriod.Format(time.DateTime), day, day.Description(), day.IsRestDay())
		})

		// 测试 ShiftedDayOff（调休假日）
		cv("测试 ShiftedDayOff 类型", func() {
			// 2024年2月13日是春节调休假日（周二）
			shiftedDayOff := time.Date(2024, 2, 13, 18, 0, 0, 0, holiday.BeijingZone())
			day := holiday.DayFromTime(shiftedDayOff)
			so(day.IsRestDay(), eq, true)
			t.Logf("ShiftedDayOff: %s -> Day: %s (%s) - IsRestDay: %v",
				shiftedDayOff.Format(time.DateTime), day, day.Description(), day.IsRestDay())
		})

		// 测试非休息日类型
		cv("测试非休息日类型", func() {
			// Workday 普通工作日
			workday := time.Date(2024, 1, 2, 9, 0, 0, 0, holiday.BeijingZone())
			day := holiday.DayFromTime(workday)
			so(day.IsRestDay(), eq, false)
			t.Logf("Workday: %s -> Day: %s (%s) - IsRestDay: %v",
				workday.Format(time.DateTime), day, day.Description(), day.IsRestDay())

			// ShiftedWorkday 调休上班日
			shiftedWorkday := time.Date(2024, 2, 4, 14, 0, 0, 0, holiday.BeijingZone())
			day = holiday.DayFromTime(shiftedWorkday)
			so(day.IsRestDay(), eq, false)
			t.Logf("ShiftedWorkday: %s -> Day: %s (%s) - IsRestDay: %v",
				shiftedWorkday.Format(time.DateTime), day, day.Description(), day.IsRestDay())
		})

		// 测试不同时区的时间
		cv("测试不同时区", func() {
			// 使用 UTC 时区创建时间，但日期应该根据北京时间判断
			// UTC 2024-01-05 23:00:00 = 北京时间 2024-01-06 07:00:00（周六，休息日）
			utcTime := time.Date(2024, 1, 5, 23, 0, 0, 0, time.UTC)
			day := holiday.DayFromTime(utcTime)
			so(day.IsRestDay(), eq, true)
			t.Logf("UTC Time: %s (北京时间: %s) -> Day: %s (%s) - IsRestDay: %v",
				utcTime.Format(time.DateTime),
				utcTime.In(holiday.BeijingZone()).Format(time.DateTime),
				day, day.Description(), day.IsRestDay())

			// UTC 2024-01-01 20:00:00 = 北京时间 2024-01-02 04:00:00（工作日）
			utcTime2 := time.Date(2024, 1, 1, 20, 0, 0, 0, time.UTC)
			day = holiday.DayFromTime(utcTime2)
			so(day.IsRestDay(), eq, false)
			t.Logf("UTC Time: %s (北京时间: %s) -> Day: %s (%s) - IsRestDay: %v",
				utcTime2.Format(time.DateTime),
				utcTime2.In(holiday.BeijingZone()).Format(time.DateTime),
				day, day.Description(), day.IsRestDay())
		})
	})
}

// TestAddWorkday 测试 AddWorkday 方法，覆盖各种休息日类型的跳过
func TestAddWorkday(t *testing.T) {
	cv("AddWorkday 测试", t, func() {
		// 测试从普通工作日开始，往后增加工作日
		cv("从工作日往后增加工作日", func() {
			// 2024年1月2日是周二，往后加1个工作日应该是1月3日（周三）
			start := holiday.Date(2024, 1, 2)
			result := start.AddWorkday(1)
			expected := holiday.Date(2024, 1, 3)
			so(result.String(), eq, expected.String())
			t.Logf("从 %s (%s) 往后1个工作日 -> %s (%s)",
				start, start.Description(), result, result.Description())

			// 2024年1月5日是周五，往后加1个工作日应该跳过周末，到1月8日（周一）
			start = holiday.Date(2024, 1, 5)
			result = start.AddWorkday(1)
			expected = holiday.Date(2024, 1, 8)
			so(result.String(), eq, expected.String())
			t.Logf("从 %s (%s) 往后1个工作日（跳过周末）-> %s (%s)",
				start, start.Description(), result, result.Description())
		})

		// 测试从普通工作日开始，往前减少工作日
		cv("从工作日往前减少工作日", func() {
			// 2024年1月3日是周三，往前减1个工作日应该是1月2日（周二）
			start := holiday.Date(2024, 1, 3)
			result := start.AddWorkday(-1)
			expected := holiday.Date(2024, 1, 2)
			so(result.String(), eq, expected.String())
			t.Logf("从 %s (%s) 往前1个工作日 -> %s (%s)",
				start, start.Description(), result, result.Description())

			// 2024年1月8日是周一，往前减1个工作日应该跳过周末，到1月5日（周五）
			start = holiday.Date(2024, 1, 8)
			result = start.AddWorkday(-1)
			expected = holiday.Date(2024, 1, 5)
			so(result.String(), eq, expected.String())
			t.Logf("从 %s (%s) 往前1个工作日（跳过周末）-> %s (%s)",
				start, start.Description(), result, result.Description())
		})

		// 测试从 Weekend 开始
		cv("从 Weekend 开始添加工作日", func() {
			// 2024年1月6日是周六，往后加1个工作日应该先到周一1月8日，然后再加1天到1月9日
			start := holiday.Date(2024, 1, 6)
			so(start.Type(), eq, holiday.Weekend)
			result := start.AddWorkday(1)
			expected := holiday.Date(2024, 1, 9)
			so(result.String(), eq, expected.String())
			t.Logf("从 %s (%s) 往后1个工作日 -> %s (%s)",
				start, start.Description(), result, result.Description())

			// 2024年1月7日是周日，往前减1个工作日应该先到周五1月5日
			start = holiday.Date(2024, 1, 7)
			so(start.Type(), eq, holiday.Weekend)
			result = start.AddWorkday(-1)
			expected = holiday.Date(2024, 1, 4)
			so(result.String(), eq, expected.String())
			t.Logf("从 %s (%s) 往前1个工作日 -> %s (%s)",
				start, start.Description(), result, result.Description())
		})

		// 测试跨越 Holiday（法定节日）
		cv("跨越 Holiday 添加工作日", func() {
			// 2023年12月29日（周五），往后加1个工作日，应该跳过周末和2024年1月1日元旦，到1月2日
			start := holiday.Date(2023, 12, 29)
			result := start.AddWorkday(1)
			expected := holiday.Date(2024, 1, 2)
			so(result.String(), eq, expected.String())
			t.Logf("从 %s (%s) 往后1个工作日（跳过元旦）-> %s (%s)",
				start, start.Description(), result, result.Description())

			// 2024年1月2日（周二），往前减1个工作日，应该跳过元旦和周末，到2023年12月29日
			start = holiday.Date(2024, 1, 2)
			result = start.AddWorkday(-1)
			expected = holiday.Date(2023, 12, 29)
			so(result.String(), eq, expected.String())
			t.Logf("从 %s (%s) 往前1个工作日（跳过元旦）-> %s (%s)",
				start, start.Description(), result, result.Description())
		})

		// 测试跨越 HolidayPeriod（法定节日假期）
		cv("跨越 HolidayPeriod 添加工作日", func() {
			// 2024年2月9日（周五，春节前最后一个工作日），往后加1个工作日
			// 应该跳过春节长假（2月10-17日），到2月18日（周日但是调班日）
			start := holiday.Date(2024, 2, 9)
			result := start.AddWorkday(1)
			expected := holiday.Date(2024, 2, 18)
			so(result.String(), eq, expected.String())
			t.Logf("从 %s (%s) 往后1个工作日（跨越春节长假）-> %s (%s)",
				start, start.Description(), result, result.Description())

			// 从春节假期中间开始
			start = holiday.Date(2024, 2, 12) // 春节假期
			so(start.Type(), eq, holiday.HolidayPeriod)
			result = start.AddWorkday(1)
			expected = holiday.Date(2024, 2, 19) // 跳过假期后，到2月18日（调班日），再加1个工作日到2月19日
			so(result.String(), eq, expected.String())
			t.Logf("从 %s (%s) 往后1个工作日 -> %s (%s)",
				start, start.Description(), result, result.Description())
		})

		// 测试跨越 ShiftedDayOff（调休假日）
		cv("跨越 ShiftedDayOff 添加工作日", func() {
			// 2024年5月1日是劳动节，2日和3日是调休假日
			// 从4月30日（周二）往后加1个工作日，应该跳过5月1-3日，到5月6日（周一）
			start := holiday.Date(2024, 4, 30)
			result := start.AddWorkday(1)
			expected := holiday.Date(2024, 5, 6)
			so(result.String(), eq, expected.String())
			t.Logf("从 %s (%s) 往后1个工作日（跳过劳动节调休）-> %s (%s)",
				start, start.Description(), result, result.Description())

			// 从调休假日本身开始
			start = holiday.Date(2024, 5, 2) // 劳动节调休假日
			so(start.Type(), eq, holiday.ShiftedDayOff)
			result = start.AddWorkday(1)
			expected = holiday.Date(2024, 5, 7) // 跳到工作日后再加1天
			so(result.String(), eq, expected.String())
			t.Logf("从 %s (%s) 往后1个工作日 -> %s (%s)",
				start, start.Description(), result, result.Description())
		})

		// 测试跨越多种休息日类型的复合场景
		cv("复合场景：跨越多种休息日类型", func() {
			// 2024年9月27日（周五），往后加1个工作日
			// 需要跳过周末（28-29日）和国庆长假（9月29日是调班日，但后面是国庆假期）
			start := holiday.Date(2024, 9, 27)
			result := start.AddWorkday(1)
			// 9月28日是周六，9月29日是周日但是调班，所以下一个工作日是9月29日
			expected := holiday.Date(2024, 9, 29)
			so(result.String(), eq, expected.String())
			t.Logf("从 %s (%s) 往后1个工作日 -> %s (%s)",
				start, start.Description(), result, result.Description())

			// 从9月30日（周一，普通工作日）往后加1个工作日
			// 应该跳过国庆长假（10月1-7日），到10月8日
			start = holiday.Date(2024, 9, 30)
			result = start.AddWorkday(1)
			expected = holiday.Date(2024, 10, 8)
			so(result.String(), eq, expected.String())
			t.Logf("从 %s (%s) 往后1个工作日（跳过国庆长假）-> %s (%s)",
				start, start.Description(), result, result.Description())
		})

		// 测试连续添加多个工作日
		cv("连续添加多个工作日", func() {
			// 2024年1月2日，往后加5个工作日
			start := holiday.Date(2024, 1, 2)
			result := start.AddWorkday(5)
			expected := holiday.Date(2024, 1, 9) // 跳过周末
			so(result.String(), eq, expected.String())
			t.Logf("从 %s (%s) 往后5个工作日 -> %s (%s)",
				start, start.Description(), result, result.Description())

			// 2024年1月9日，往前减5个工作日
			start = holiday.Date(2024, 1, 9)
			result = start.AddWorkday(-5)
			expected = holiday.Date(2024, 1, 2) // 跳过周末
			so(result.String(), eq, expected.String())
			t.Logf("从 %s (%s) 往前5个工作日 -> %s (%s)",
				start, start.Description(), result, result.Description())
		})

		// 测试加0个工作日
		cv("添加0个工作日", func() {
			// 从工作日开始，加0应该返回自己
			start := holiday.Date(2024, 1, 2)
			result := start.AddWorkday(0)
			so(result.String(), eq, start.String())
			t.Logf("从 %s (%s) 加0个工作日 -> %s (%s)",
				start, start.Description(), result, result.Description())

			// 从休息日开始，加0应该顺延到下一个工作日
			start = holiday.Date(2024, 1, 6) // 周六
			result = start.AddWorkday(0)
			expected := holiday.Date(2024, 1, 8) // 周一
			so(result.String(), eq, expected.String())
			t.Logf("从 %s (%s) 加0个工作日（顺延到工作日）-> %s (%s)",
				start, start.Description(), result, result.Description())
		})
	})
}
