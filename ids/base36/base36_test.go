package base36_test

import (
	"strconv"
	"testing"

	"github.com/Andrew-M-C/go.util/ids/base36"
	"github.com/smartystreets/goconvey/convey"
)

var (
	cv = convey.Convey
	so = convey.So
	eq = convey.ShouldEqual

	isNil = convey.ShouldBeNil
)

func TestAtoi(t *testing.T) {
	cv("Atoi", t, func() {
		s := "zzzzzzzzzz"
		i, err := base36.Atoi[int64](s)
		so(err, isNil)
		so(i, eq, base36.MaxID)
	})

	cv("浮点数精度支持", t, func() {
		s := "ZZZZZZZZZZ"
		f, err := base36.Atoi[float64](s)
		so(err, isNil)
		so(f, eq, base36.MaxID)
		so(uint64(f), eq, base36.MaxID)
		so(int64(f), eq, base36.MaxID)
	})
}

func TestItoa(t *testing.T) {
	cv("Itoa", t, func() {
		i := int64(1234567890)
		s := base36.Itoa(i)
		ii, err := base36.Atoi[int64](s)
		so(err, isNil)
		so(ii, eq, i)
	})
}

func TestRevEndian(t *testing.T) {
	cv("RevEndianItoa32 and RevEndianAtoi32", t, func() {
		id := uint32(0x1278abef)
		s := base36.RevEndianItoa32(id)
		t.Log("reversed string ID:", s)

		reversedID, err := strconv.ParseUint(s, 36, 32)
		so(err, isNil)
		so(reversedID, eq, 0xefab7812)

		parsedID, err := base36.RevEndianAtoi32(s)
		so(err, isNil)
		so(parsedID, eq, id)

		s = base36.RevEndianItoa32(1)
		t.Log("reversed string ID for 1:", s)

		s = base36.RevEndianItoa32(0x1000000)
		t.Log("reversed string ID for 0x1000000:", s)
	})

	cv("RevEndianItoa64 and RevEndianAtoi64", t, func() {
		id := uint64(0x123456789abcdef0)
		s := base36.RevEndianItoa64(id)
		t.Log("reversed string ID:", s)

		reversedID, err := strconv.ParseUint(s, 36, 64)
		so(err, isNil)
		so(reversedID, eq, uint64(0xf0debc9a78563412))

		parsedID, err := base36.RevEndianAtoi64(s)
		so(err, isNil)
		so(parsedID, eq, id)

		s = base36.RevEndianItoa64(1)
		t.Log("reversed string ID for 1:", s)
	})
}

func TestQuirky(t *testing.T) {
	cv("QuirkyItoa32 and QuirkyAtoi32", t, func() {
		id := uint32(0x1278abef)
		s := base36.QuirkyItoa32(id)
		t.Log("quirky string ID:", s)

		reversedID, err := strconv.ParseUint(s, 36, 64)
		so(err, isNil)
		so(reversedID, eq, 0x1efab7812)

		parsedID, err := base36.QuirkyAtoi32(s)
		so(err, isNil)
		so(parsedID, eq, id)

		s = base36.QuirkyItoa32(1)
		t.Log("quirky string ID for 1:", s)

		s = base36.QuirkyItoa32(0) // 最小值
		t.Log("quirky string ID for 0:", s)

		s = base36.QuirkyItoa32(0xFFFFFFFF) // 最大值
		t.Log("quirky string ID for 0xFFFFFFFF:", s)
	})
}
