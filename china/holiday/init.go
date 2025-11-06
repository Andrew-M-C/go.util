package holiday

func init() {
	init2026()
	init2025()
	init2024()
}

// reference: https://www.gov.cn/zhengce/zhengceku/202511/content_7047091.htm
func init2026() {
	// 元旦：1月1日（周四）至3日（周六）放假调休，共3天。1月4日（周日）上班。
	newDate(2026, 1, 1).withType(Holiday).withName("元旦节").add()
	newDate(2026, 1, 2).withType(ShiftedDayOff).withName("元旦节").add()
	newDate(2026, 1, 4).withType(ShiftedWorkday).withName("元旦节").add()

	// 春节：2月15日（农历腊月二十八、周日）至23日（农历正月初七、周一）放假调休，共9天。2月14日（周六）、2月28日（周六）上班。
	newDate(2026, 2, 14).withType(ShiftedWorkday).withName("春节").add()
	// newDate(2026, 2, 15).withType(HolidayPeriod).withName("春节").add() // 周日
	newDate(2026, 2, 16).withType(HolidayPeriod).withName("春节").add()
	newDate(2026, 2, 17).withType(Holiday).withName("春节").add()
	newDate(2026, 2, 18).withType(HolidayPeriod).withName("春节").add()
	newDate(2026, 2, 19).withType(HolidayPeriod).withName("春节").add()
	newDate(2026, 2, 20).withType(ShiftedDayOff).withName("春节").add()
	newDate(2026, 2, 23).withType(ShiftedDayOff).withName("春节").add()
	newDate(2026, 2, 28).withType(ShiftedWorkday).withName("春节").add()

	// 清明节：4月4日（周六）至6日（周一）放假，共3天。
	// newDate(2026, 4, 4).withType(HolidayPeriod).withName("清明节").add() // 周六
	newDate(2026, 4, 5).withType(Holiday).withName("清明节").add()       // 周日
	newDate(2026, 4, 6).withType(ShiftedDayOff).withName("清明节").add() // 周一

	// 劳动节：5月1日（周五）至5日（周二）放假调休，共5天。5月9日（周六）上班。
	newDate(2026, 5, 1).withType(Holiday).withName("劳动节").add()
	// newDate(2026, 5, 4).withType(ShiftedDayOff).withName("劳动节").add() // 周六
	// newDate(2026, 5, 5).withType(ShiftedDayOff).withName("劳动节").add() // 周日
	newDate(2026, 5, 9).withType(ShiftedWorkday).withName("劳动节").add()

	// 端午节：6月19日（周五）至21日（周日）放假，共3天。
	newDate(2026, 6, 19).withType(Holiday).withName("端午节").add()

	// 中秋节：9月25日（周五）至27日（周日）放假，共3天。
	newDate(2026, 9, 25).withType(Holiday).withName("中秋节").add()

	// 国庆节：10月1日（周四）至7日（周三）放假调休，共7天。9月20日（周日）、10月10日（周六）上班。
	newDate(2026, 9, 20).withType(ShiftedWorkday).withName("国庆节").add()
	newDate(2026, 10, 1).withType(Holiday).withName("国庆节").add()
	newDate(2026, 10, 2).withType(HolidayPeriod).withName("国庆节").add()
	newDate(2026, 10, 3).withType(HolidayPeriod).withName("国庆节").add() // 周六
	newDate(2026, 10, 5).withType(ShiftedDayOff).withName("国庆节").add() // 周一
	newDate(2026, 10, 6).withType(ShiftedDayOff).withName("国庆节").add() // 周二
	newDate(2026, 10, 7).withType(ShiftedDayOff).withName("国庆节").add() // 周三
	newDate(2026, 10, 10).withType(ShiftedWorkday).withName("国庆节").add()
}

// reference: https://www.gov.cn/zhengce/content/202310/content_6911527.htm
func init2024() {
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

// reference: https://www.gov.cn/zhengce/zhengceku/202411/content_6986383.htm
func init2025() {
	// 元旦
	newDate(2025, 1, 1).withType(Holiday).withName("元旦节").add()
	// 春节
	newDate(2025, 1, 26).withType(ShiftedWorkday).withName("春节").add()
	newDate(2025, 1, 28).withType(HolidayPeriod).withName("春节").add()
	newDate(2025, 1, 29).withType(Holiday).withName("春节").add()
	newDate(2025, 1, 30).withType(HolidayPeriod).withName("春节").add()
	newDate(2025, 1, 31).withType(HolidayPeriod).withName("春节").add()
	newDate(2025, 2, 3).withType(ShiftedDayOff).withName("春节").add()
	newDate(2025, 2, 4).withType(ShiftedDayOff).withName("春节").add()
	newDate(2025, 2, 8).withType(ShiftedWorkday).withName("春节").add()
	// 清明节
	newDate(2025, 4, 4).withType(Holiday).withName("清明节").add()
	// 劳动节
	newDate(2025, 4, 27).withType(ShiftedWorkday).withName("劳动节").add()
	newDate(2025, 5, 1).withType(Holiday).withName("劳动节").add()
	newDate(2025, 5, 2).withType(HolidayPeriod).withName("劳动节").add()
	newDate(2025, 5, 5).withType(ShiftedDayOff).withName("劳动节").add()
	// 端午节
	newDate(2025, 5, 31).withType(Holiday).withName("端午节").add()
	newDate(2025, 6, 2).withType(ShiftedDayOff).withName("端午节").add()
	// 国庆节、中秋节
	newDate(2025, 9, 28).withType(ShiftedWorkday).withName("国庆节").add()
	newDate(2025, 10, 1).withType(Holiday).withName("国庆节").add()
	newDate(2025, 10, 2).withType(HolidayPeriod).withName("国庆节").add()
	newDate(2025, 10, 3).withType(HolidayPeriod).withName("国庆节").add()
	newDate(2025, 10, 6).withType(Holiday).withName("中秋节").add()
	newDate(2025, 10, 7).withType(ShiftedDayOff).withName("国庆节").add()
	newDate(2025, 10, 8).withType(ShiftedDayOff).withName("国庆节").add()
	newDate(2025, 10, 11).withType(ShiftedWorkday).withName("国庆节").add()
}
