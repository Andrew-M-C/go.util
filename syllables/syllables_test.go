package syllables_test

import (
	"testing"
	"time"

	"github.com/Andrew-M-C/go.util/log"
	"github.com/Andrew-M-C/go.util/syllables"
	"github.com/smartystreets/goconvey/convey"
)

var (
	cv = convey.Convey
	so = convey.So
	eq = convey.ShouldEqual
)

func TestSyllables(t *testing.T) {
	cv("英文", t, func() { testEnglish(t) })
	cv("英文 + 数字", t, func() { testEnglishAndNumbers(t) })
	cv("中文", t, func() { testChinese(t) })
	cv("一些奇怪的 case", t, func() { testStrangeCases(t) })
}

func testEnglish(t *testing.T) {
	const in = `This is Golang, not Java.`

	defer func(start time.Time) {
		t.Logf("elapsed: %v", time.Since(start))
	}(time.Now())

	total, w := syllables.SplitAndCount(in)
	t.Log(in)
	t.Log(log.ToJSON(w))
	so(total, eq, 7)
	so(len(w), eq, 11)

	so(w[0].Word, eq, "This")
	so(w[0].SyllableCount, eq, 1)

	so(w[1].Word, eq, " ")
	so(w[1].SyllableCount, eq, 0)

	so(w[2].Word, eq, "is")
	so(w[2].SyllableCount, eq, 1)

	so(w[3].Word, eq, " ")
	so(w[3].SyllableCount, eq, 0)

	so(w[4].Word, eq, "Golang")
	so(w[4].SyllableCount, eq, 2)

	so(w[5].Word, eq, ",")
	so(w[5].SyllableCount, eq, 0)

	so(w[6].Word, eq, " ")
	so(w[6].SyllableCount, eq, 0)

	so(w[7].Word, eq, "not")
	so(w[7].SyllableCount, eq, 1)

	so(w[8].Word, eq, " ")
	so(w[8].SyllableCount, eq, 0)

	so(w[9].Word, eq, "Java")
	so(w[9].SyllableCount, eq, 2)

	so(w[10].Word, eq, ".")
	so(w[10].SyllableCount, eq, 0)
}

func testEnglishAndNumbers(t *testing.T) {
	const in = `Now is Year 2023, not 2003 anymore.`

	defer func(start time.Time) {
		t.Logf("elapsed: %v", time.Since(start))
	}(time.Now())

	total, w := syllables.SplitAndCount(in)
	t.Log(in)
	t.Log(log.ToJSON(w))
	so(total, eq, 15) // anymore 三个音节, 每个阿拉伯数字视为一个字节
	so(len(w), eq, 21)
}

func testChinese(t *testing.T) {
	const in = `各位观众晚上好，欢迎收看新闻联播`

	total, w := syllables.SplitAndCount(in)
	t.Log(in)
	t.Log(log.ToJSON(w))
	so(total, eq, 15)
}

func testStrangeCases(t *testing.T) {
	cv("急急急急……", func() {
		const in = `急急急急急急急急急急急急急急急急急急急急急急急急急急急急急急急急急急急急急急急急急急急急急急急急急急急急急急急急急急急急急急急急急急急急急急急急急急急急急急`

		total, w := syllables.SplitAndCount(in)
		t.Log(in)
		t.Log(log.ToJSON(w))
		so(total, eq, utf8Len(in))
		so(len(w), eq, (utf8Len(in)+1)/2) // "急急" 表示一个词
	})

	cv("圆周率 - 阿拉伯数字", func() {
		const in = `3.1415926535897935384626`

		total, w := syllables.SplitAndCount(in)
		t.Log(in)
		t.Log(log.ToJSON(w))
		so(total, eq, len(in)-1)
		so(len(w), eq, len(in))
	})
}

func utf8Len(s string) (length int) {
	for range s {
		length++
	}
	return
}
