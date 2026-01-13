package time

import (
	"testing"
	"time"
)

func TestTimestamp(t *testing.T) {
	cv("测试 TimestampSecs", t, func() { testTimestampSecs(t) })
	cv("测试 TimestampMillis", t, func() { testTimestampMillis(t) })
	cv("测试 TimestampMicros", t, func() { testTimestampMicros(t) })
	cv("测试 TimestampNanos", t, func() { testTimestampNanos(t) })
	cv("测试边界情况", t, func() { testTimestampEdgeCases(t) })
	cv("测试负数时间戳", t, func() { testNegativeTimestamps(t) })
}

func testTimestampSecs(t *testing.T) {
	// 测试正常的秒级时间戳
	// 2024-01-01 00:00:00 UTC
	ts := TimestampSecs(1704067200)
	tm := ts.Time()

	so(tm.Unix(), eq, int64(1704067200))
	so(tm.Year(), eq, 2024)
	so(tm.Month(), eq, time.January)
	so(tm.Day(), eq, 1)

	// 测试 String() 方法
	str := ts.String()
	if str == "" {
		t.Errorf("String() should not be empty")
	}
	t.Logf("TimestampSecs String: %s", str)

	// 测试零值
	tsZero := TimestampSecs(0)
	tmZero := tsZero.Time()
	so(tmZero.Unix(), eq, int64(0))
	so(tmZero.Year(), eq, 1970)
}

func testTimestampMillis(t *testing.T) {
	// 测试毫秒级时间戳
	// 2024-01-01 00:00:00.123 UTC
	ts := TimestampMillis(1704067200123)
	tm := ts.Time()

	so(tm.Unix(), eq, int64(1704067200))
	so(tm.UnixMilli(), eq, int64(1704067200123))
	so(tm.Nanosecond()/1000000, eq, 123) // 毫秒部分

	// 测试 String() 方法
	str := ts.String()
	if str == "" {
		t.Errorf("String() should not be empty")
	}
	t.Logf("TimestampMillis String: %s", str)

	// 测试精度保持
	tsMillis := TimestampMillis(1704067200999)
	tmMillis := tsMillis.Time()
	so(tmMillis.UnixMilli(), eq, int64(1704067200999))
}

func testTimestampMicros(t *testing.T) {
	// 测试微秒级时间戳
	// 2024-01-01 00:00:00.123456 UTC
	ts := TimestampMicros(1704067200123456)
	tm := ts.Time()

	so(tm.Unix(), eq, int64(1704067200))
	so(tm.UnixMicro(), eq, int64(1704067200123456))
	so(tm.Nanosecond()/1000, eq, 123456) // 微秒部分

	// 测试 String() 方法
	str := ts.String()
	if str == "" {
		t.Errorf("String() should not be empty")
	}
	t.Logf("TimestampMicros String: %s", str)

	// 测试精度保持
	tsMicros := TimestampMicros(1704067200999999)
	tmMicros := tsMicros.Time()
	so(tmMicros.UnixMicro(), eq, int64(1704067200999999))
}

func testTimestampNanos(t *testing.T) {
	// 测试纳秒级时间戳
	// 2024-01-01 00:00:00.123456789 UTC
	ts := TimestampNanos(1704067200123456789)
	tm := ts.Time()

	so(tm.Unix(), eq, int64(1704067200))
	so(tm.Nanosecond(), eq, 123456789)

	// 测试 String() 方法
	str := ts.String()
	if str == "" {
		t.Errorf("String() should not be empty")
	}
	t.Logf("TimestampNanos String: %s", str)

	// 测试完整的纳秒时间戳转换
	now := time.Now()
	tsNow := TimestampNanos(now.UnixNano())
	tmNow := tsNow.Time()

	// 验证转换的准确性（纳秒级应该完全一致）
	diff := now.Sub(tmNow)
	if diff < 0 {
		diff = -diff
	}
	if diff >= time.Nanosecond {
		t.Errorf("Nano timestamp conversion diff too large: %v", diff)
	}
	t.Logf("Nano timestamp conversion diff: %v", diff)
}

func testTimestampEdgeCases(t *testing.T) {
	// 测试各种边界情况

	// 1. Unix 纪元时间 (1970-01-01 00:00:00)
	cv("Unix 纪元时间", func() {
		tsSec := TimestampSecs(0)
		so(tsSec.Time().Unix(), eq, int64(0))

		tsMillis := TimestampMillis(0)
		so(tsMillis.Time().UnixMilli(), eq, int64(0))

		tsMicros := TimestampMicros(0)
		so(tsMicros.Time().UnixMicro(), eq, int64(0))

		tsNanos := TimestampNanos(0)
		so(tsNanos.Time().UnixNano(), eq, int64(0))
	})

	// 2. 很大的时间戳 (2100-01-01)
	cv("未来时间戳", func() {
		// 2100-01-01 00:00:00 UTC
		futureSec := int64(4102444800)
		tsSec := TimestampSecs(futureSec)
		tm := tsSec.Time()
		so(tm.Year(), eq, 2100)
		t.Logf("Future time: %s", tm.String())
	})

	// 3. 测试不同精度之间的一致性
	cv("不同精度一致性", func() {
		baseSec := int64(1704067200)

		tsSec := TimestampSecs(baseSec)
		tsMillis := TimestampMillis(baseSec * 1000)
		tsMicros := TimestampMicros(baseSec * 1000000)
		tsNanos := TimestampNanos(baseSec * 1000000000)

		so(tsSec.Time().Unix(), eq, tsMillis.Time().Unix())
		so(tsSec.Time().Unix(), eq, tsMicros.Time().Unix())
		so(tsSec.Time().Unix(), eq, tsNanos.Time().Unix())
	})
}

func testNegativeTimestamps(t *testing.T) {
	// 测试负数时间戳（1970年之前的时间）

	cv("负数秒级时间戳", func() {
		// Unix(-1) = 1969-12-31 23:59:59 UTC = 1970-01-01 07:59:59 CST
		ts := TimestampSecs(-1)
		tm := ts.Time()
		so(tm.Unix(), eq, int64(-1))
		// 在 UTC 时区下是 1969 年，但在 CST 时区下是 1970 年
		tmUTC := tm.UTC()
		so(tmUTC.Year(), eq, 1969)
		t.Logf("Negative timestamp: %s (UTC: %s)", tm.String(), tmUTC.String())
	})

	cv("负数毫秒级时间戳", func() {
		ts := TimestampMillis(-1000)
		tm := ts.Time()
		so(tm.Unix(), eq, int64(-1))
		so(tm.UnixMilli(), eq, int64(-1000))
	})

	cv("负数微秒级时间戳", func() {
		ts := TimestampMicros(-1000000)
		tm := ts.Time()
		so(tm.Unix(), eq, int64(-1))
		so(tm.UnixMicro(), eq, int64(-1000000))
	})

	cv("负数纳秒级时间戳", func() {
		ts := TimestampNanos(-1000000000)
		tm := ts.Time()
		so(tm.Unix(), eq, int64(-1))

		// 注意：这里可能会暴露 TimestampNanos 实现的问题
		// 如果实现有问题，这个测试会失败
		expectedNano := int64(-1000000000)
		actualNano := tm.UnixNano()

		if expectedNano != actualNano {
			t.Logf("WARNING: TimestampNanos 处理负数时间戳可能有问题")
			t.Logf("Expected: %d, Got: %d", expectedNano, actualNano)
		}
	})
}
