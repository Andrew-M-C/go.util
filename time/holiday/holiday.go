package holiday

import (
	_ "embed"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

// NextHoliday 返回下一个假日在几天后, 以及假日的名称。如果当天是假日, 则返回 0
func NextHoliday() (int, string) {
	return NextHolidayForDate(today())
}

// NextWorkday 返回下一个工作日在几天后
func NextWorkday() int {
	return NextWorkdayForDate(today())
}

// NextHolidayForDate 返回指定日期的下一个假日在几天后, 以及假日的名称。如果当天是假日, 则返回 0
func NextHolidayForDate(date time.Time) (int, string) {
	year := date.Year()
	yday := date.YearDay() - 1
	lastday := Day(year, 12, 31).YearDay() - 1

	// 当年假日
	days := readYearData(year)
	for i := yday; i <= lastday; i++ {
		desc := days[i]
		if !isWorkday(desc) {
			return i - yday, desc
		}
	}

	// 跨年假日
	days = readYearData(year + 1)
	for i, desc := range days {
		if !isWorkday(desc) {
			return lastday - yday + 1 + i, desc
		}
	}
	return 365, "" // 不可能执行到这里
}

// IsWorkday 返回指定时间是否工作日
func IsWorkday(tm time.Time) bool {
	yesterday := tm.AddDate(0, 0, -1)
	return NextWorkdayForDate(yesterday) == 1
}

// NextWorkdayForDate 返回指定日期的下一个工作日在几天后, 如果今天是假日, 则返回 1
func NextWorkdayForDate(date time.Time) int {
	year := date.Year()
	yday := date.YearDay() - 1
	lastday := Day(year, 12, 31).YearDay() - 1

	// 当年工作日
	days := readYearData(year)
	for i := yday; i <= lastday; i++ {
		desc := days[i]
		if isWorkday(desc) {
			return i - yday
		}
	}

	// 跨年假日
	days = readYearData(year + 1)
	for i, desc := range days {
		if isWorkday(desc) {
			return lastday - yday + 1 + i
		}
	}
	return 365 // 不可能执行到这里
}

// Day 构建某一天的 0:00:00 时刻
func Day(year, month, day int) time.Time {
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, internal.china)
}

// Override 覆盖某一天的属性, holidayName 如果传空或 "工作日" 则表示设置为工作日,
// 否则设置为对应名称的休息日
func Override(day time.Time, holidayName string) {
	days := readYearData(day.Year())
	yday := day.YearDay() - 1

	if isWorkday(holidayName) {
		days[yday] = ""
	} else {
		days[yday] = holidayName
	}
}

// AddYearYAML 添加 (如果已存在则覆盖) 某年的一个配置, 配置只支持 MM-DD 格式, yaml 格式
func AddYearYAML(year int, data []byte) error {
	m, err := unmarshalYAML(data)
	if err != nil {
		return err
	}
	days, err := parseYear(year, m)
	if err != nil {
		return err
	}
	addYearData(year, days)
	return nil
}

//go:embed config.yaml
var configBytes []byte

func init() {
	var m map[string]map[string]string
	yaml.Unmarshal(configBytes, &m)

	for k, v := range m {
		year, _ := strconv.ParseInt(k, 10, 32)
		days, _ := parseYear(int(year), v)
		addYearData(int(year), days)
	}
}

func today() time.Time {
	return time.Now().In(internal.china)
}
