package freshness_test

import (
	"fmt"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Andrew-M-C/go.util/maps/freshness"
	tmutil "github.com/Andrew-M-C/go.util/time"
	"github.com/smartystreets/goconvey/convey"
)

var (
	cv = convey.Convey
	so = convey.So
	eq = convey.ShouldEqual
	ne = convey.ShouldNotEqual

	isNil  = convey.ShouldBeNil
	notNil = convey.ShouldNotBeNil
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func TestMap(t *testing.T) {
	cv("没有 newer 时的默认逻辑", t, func() { testMapNoNewer(t) })
	cv("测试基础逻辑", t, func() { testMapGeneral(t) })
}

func testMapNoNewer(*testing.T) {
	cv("整数", func() {
		m := freshness.NewMap[uint, int]()
		v, exist, err := m.GetOrNew(1)
		so(exist, eq, false)
		so(err, isNil)
		so(v, eq, 0)

		v, exist, err = m.GetOrNew(1)
		so(exist, eq, true)
		so(err, isNil)
		so(v, eq, 0)

		exist = m.Set(1, -1)
		so(exist, eq, true)

		v, exist, err = m.GetOrNew(1)
		so(exist, eq, true)
		so(err, isNil)
		so(v, eq, -1)
	})

	cv("指针", func() {
		type data struct {
			P string
			N string
		}

		m := freshness.NewMap[uint, *data]()
		v, exist, err := m.GetOrNew(1)
		so(exist, eq, false)
		so(err, isNil)
		so(v, notNil)
		so(v.P, eq, "")
		so(v.N, eq, "")
		v.P = "1"
		v.N = "-1"

		v, exist, err = m.GetOrNew(1)
		so(exist, eq, true)
		so(err, isNil)
		so(v, notNil)
		so(v.P, ne, "")
		so(v.N, ne, "")
	})
}

func testMapGeneral(t *testing.T) {
	cv("general", func() {
		expCount := uint32(0)
		expCalback := func(string, string) {
			_ = atomic.AddUint32(&expCount, 1)
		}
		newer := func(key string) (string, error) {
			return fmt.Sprintf("%X", []byte(key)), nil
		}

		m := freshness.NewMap[string, string](
			freshness.WithDebug(t.Logf),
			freshness.WithTimeout(time.Second),
			freshness.WithExpireCallback[string, string](expCalback),
			freshness.WithNewer[string, string](newer),
		)

		exist := m.Set("12345", "56789")
		so(exist, eq, false)

		v, exist := m.Get("12345")
		so(exist, eq, true)
		so(v, eq, "56789")

		tmutil.Sleep(0.55)

		v, exist, err := m.GetOrNew("0")
		so(err, isNil)
		so(exist, eq, false)
		so(v, eq, "30")

		tmutil.Sleep(0.55)

		v, exist = m.Get("12345")
		so(exist, eq, false)
		so(v, eq, "")

		v, exist = m.Get("0")
		so(exist, eq, true)
		so(v, eq, "30")

		tmutil.Sleep(0.55)

		v, exist = m.Get("0")
		so(exist, eq, true)
		so(v, eq, "30")

		tmutil.Sleep(1.1)

		v, exist = m.Get("")
		so(exist, eq, false)
		so(v, eq, "")

		so(expCount, eq, 2)
	})
}
