package time

import (
	"testing"
	"time"
)

func testAge(t *testing.T) {
	cv("正常逻辑", func() {
		age := CalculateAge(date(2001, 6, 15), date(2001, 7, 25))
		t.Log(age)
		so(age.Years, eq, 0)
		so(age.Months, eq, 1)
		so(age.Days, eq, 10)
	})

	cv("天数借位", func() {
		age := CalculateAge(date(2001, 6, 15), date(2001, 7, 10))
		t.Log(age)
		so(age.Years, eq, 0)
		so(age.Months, eq, 0)
		so(age.Days, eq, 25)
	})

	cv("月数借位", func() {
		age := CalculateAge(date(2001, 6, 15), date(2002, 6, 14))
		t.Log(age)
		so(age.Years, eq, 0)
		so(age.Months, eq, 11)
		so(age.Days, eq, 30)
	})

	cv("月为空", func() {
		age := CalculateAge(date(2001, 6, 15), date(2002, 7, 14))
		t.Log(age)
		so(age.Years, eq, 1)
		so(age.Months, eq, 0)
		so(age.Days, eq, 29)
	})

	cv("天为空", func() {
		age := CalculateAge(date(2001, 6, 15), date(2002, 7, 15))
		t.Log(age)
		so(age.Years, eq, 1)
		so(age.Months, eq, 1)
		so(age.Days, eq, 0)
	})

	cv("未来", func() {
		age := CalculateAge(date(2001, 7, 16), date(2001, 7, 15))
		t.Log(age)
		so(age.Years, eq, 0)
		so(age.Months, eq, 0)
		so(age.Days, eq, 0)
	})
}

func date(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, Beijing)
}
