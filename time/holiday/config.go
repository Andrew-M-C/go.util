package holiday

import (
	"errors"
	"fmt"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	workday = "工作日" // 空字符串也是
	weekend = "周末"
	// 其他的就是具体的假期原因
)

type dayType = string

func isWorkday(s string) bool {
	if s == "" {
		return true
	}
	return s == workday
}

// func unmarshalJSON(b []byte) (map[string]string, error) {
// 	var data map[string]string
// 	if err := json.Unmarshal(b, &data); err != nil {
// 		return nil, err
// 	}
// 	return data, nil
// }

func unmarshalYAML(b []byte) (map[string]string, error) {
	var data map[string]string
	if err := yaml.Unmarshal(b, &data); err != nil {
		return nil, err
	}
	return data, nil
}

// parseYear 解析某年的一个配置, 配置只支持 MM-DD 格式
func parseYear(year int, data map[string]string) (days []dayType, err error) {
	if year < 0 {
		return nil, errors.New("不支持公元前")
	}

	days = defaultConfForYear(year)
	for s, v := range data {
		str := fmt.Sprintf("%d-%s", year, s)
		d, err := parseDay(str)
		if err != nil {
			return nil, fmt.Errorf("illegal day: '%s'", s)
		}

		yday := d.YearDay() - 1
		if isWorkday(v) {
			days[yday] = ""
		} else {
			days[yday] = v
		}
	}
	return days, nil
}

func parseDay(s string) (time.Time, error) {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return t, err
	}
	return Day(t.Year(), int(t.Month()), t.Day()), nil
}

// defaultConfForYear 返回某年的配置
func defaultConfForYear(year int) []dayType {
	res := make([]dayType, 366)
	d := Day(year, 1, 1)
	for i := range res {
		switch d.Weekday() {
		case time.Sunday, time.Saturday:
			res[i] = weekend
		default:
			// nothing needs to be done
		}
		d = d.AddDate(0, 0, 1)
	}
	return res
}
