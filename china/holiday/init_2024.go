package holiday

// reference: https://www.gov.cn/zhengce/content/202310/content_6911527.htm
func init() {
	// 元旦
	newDate(2024, 1, 1).withType(Holiday).withName("元旦节").add()
	// 春节
	newDate(2024, 2, 4).withType(ShiftedWorkday).withName("春节").add()
	newDate(2024, 2, 10).withType(Holiday).withName("春节").add()
	newDate(2024, 2, 11).withType(HolidayPeriod).withName("春节").add()
	newDate(2024, 2, 12).withType(HolidayPeriod).withName("春节").add()
	newDate(2024, 2, 13).withType(ShiftedDayOff).withName("春节").add()
	newDate(2024, 2, 14).withType(ShiftedDayOff).withName("春节").add()
	newDate(2024, 2, 15).withType(ShiftedDayOff).withName("春节").add()
	newDate(2024, 2, 16).withType(ShiftedDayOff).withName("春节").add()
	newDate(2024, 2, 18).withType(ShiftedWorkday).withName("春节").add()
	// 清明节
	newDate(2024, 4, 4).withType(Holiday).withName("清明节").add()
	newDate(2024, 4, 5).withType(ShiftedDayOff).withName("清明节").add()
	newDate(2024, 4, 7).withType(ShiftedWorkday).withName("清明节").add()
	// 劳动节
	newDate(2024, 4, 28).withType(ShiftedWorkday).withName("劳动节").add()
	newDate(2024, 5, 1).withType(Holiday).withName("劳动节").add()
	newDate(2024, 5, 2).withType(ShiftedDayOff).withName("劳动节").add()
	newDate(2024, 5, 3).withType(ShiftedDayOff).withName("劳动节").add()
	newDate(2024, 5, 11).withType(ShiftedWorkday).withName("劳动节").add()
	// 端午节
	newDate(2024, 6, 10).withType(Holiday).withName("端午节").add()
	// 中秋节
	newDate(2024, 9, 14).withType(ShiftedWorkday).withName("端午节").add()
	newDate(2024, 9, 16).withType(ShiftedDayOff).withName("端午节").add()
	newDate(2024, 9, 17).withType(Holiday).withName("端午节").add()
	// 国庆节
	newDate(2024, 9, 29).withType(ShiftedWorkday).withName("国庆节").add()
	newDate(2024, 10, 1).withType(Holiday).withName("国庆节").add()
	newDate(2024, 10, 2).withType(HolidayPeriod).withName("国庆节").add()
	newDate(2024, 10, 3).withType(HolidayPeriod).withName("国庆节").add()
	newDate(2024, 10, 4).withType(ShiftedDayOff).withName("国庆节").add()
	newDate(2024, 10, 7).withType(ShiftedDayOff).withName("国庆节").add()
	newDate(2024, 10, 12).withType(ShiftedWorkday).withName("国庆节").add()
}
