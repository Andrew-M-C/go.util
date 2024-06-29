package expire_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/Andrew-M-C/go.util/localcache/expire"
	"github.com/smartystreets/goconvey/convey"
)

func milli(i int) time.Duration {
	return time.Duration(i) * time.Millisecond
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

var (
	cv = convey.Convey
	so = convey.So
	eq = convey.ShouldEqual

	isNil  = convey.ShouldBeNil
	notNil = convey.ShouldNotBeNil
)

func TestExpireMap(t *testing.T) {
	cases := [][2]any{
		{"基础功能", testExpireMapBasic},
	}
	for _, c := range cases {
		name, _ := c[0].(string)
		tester, _ := c[1].(func(*testing.T))
		cv(name, t, func() { tester(t) })
	}
}

func testExpireMapBasic(t *testing.T) {
	m := expire.NewMap[int, string](
		expire.WithTimeout(milli(100)),
		expire.WithDebugger(t.Logf),
		expire.WithNewer(func(key int) *string {
			s := fmt.Sprintf("{int-%d}", key)
			return &s
		}),
	)

	v, exist := m.Load(1)
	so(exist, eq, false)
	so(v, isNil)

	v, exist = m.LoadOrNew(1)
	so(exist, eq, false)
	so(v, notNil)
	so(*v, eq, "{int-1}")

	_, exist = m.Load(1)
	so(exist, eq, true)

	v, exist = m.LoadOrNew(1, expire.WithNewer(func(key int) *string {
		s := fmt.Sprint(key)
		return &s
	}))
	so(exist, eq, true)
	so(*v, eq, "{int-1}")

	s := "111"
	v, exist = m.Swap(1, &s)
	so(exist, eq, true)
	so(*v, eq, "{int-1}")

	v, exist = m.Load(1)
	so(exist, eq, true)
	so(*v, eq, "111")

	time.Sleep(milli(100))
	v, exist = m.Load(1)
	so(exist, eq, false)
	so(v, isNil)
}
