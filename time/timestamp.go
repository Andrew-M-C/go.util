package time

import "time"

// TimestampSecs 表示秒级时间戳
type TimestampSecs int64

func (ts TimestampSecs) Time() time.Time {
	return time.Unix(int64(ts), 0)
}

func (ts TimestampSecs) String() string {
	return ts.Time().String()
}

// TimestampMillis 表示毫秒级时间戳
type TimestampMillis int64

func (ts TimestampMillis) Time() time.Time {
	return time.UnixMilli(int64(ts))
}

func (ts TimestampMillis) String() string {
	return ts.Time().String()
}

// TimestampMicros 表示微秒级时间戳
type TimestampMicros int64

func (ts TimestampMicros) Time() time.Time {
	return time.UnixMicro(int64(ts))
}

func (ts TimestampMicros) String() string {
	return ts.Time().String()
}

// TimestampNanos 表示纳秒级时间戳
type TimestampNanos int64

func (ts TimestampNanos) Time() time.Time {
	return time.Unix(0, int64(ts))
}

func (ts TimestampNanos) String() string {
	return ts.Time().String()
}
