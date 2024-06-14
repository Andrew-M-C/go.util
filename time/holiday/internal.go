package holiday

import (
	"sync"
	"time"
)

var internal = struct {
	lck        sync.RWMutex
	china      *time.Location
	daysByYear map[int][]dayType // 1st key: 年份; 2nd key: day of year;
}{
	china:      time.FixedZone("China", 8*60*60),
	daysByYear: make(map[int][]dayType, 10),
}

func addYearData(year int, days []dayType) {
	internal.lck.Lock()
	defer internal.lck.Unlock()

	internal.daysByYear[year] = days
}

func readYearData(year int) []dayType {
	internal.lck.RLock()
	res, exist := internal.daysByYear[year]
	internal.lck.RUnlock()
	if exist {
		return res
	}

	internal.lck.Lock()
	defer internal.lck.Unlock()

	res, exist = internal.daysByYear[year]
	if exist {
		return res
	}
	res = defaultConfForYear(year)
	internal.daysByYear[year] = res
	return res
}
