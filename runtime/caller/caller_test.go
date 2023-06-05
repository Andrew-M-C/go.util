// Package caller provides packaging for system runtime.Caller
package caller

import (
	"encoding/json"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

var (
	cv = convey.Convey
	so = convey.So

	eq = convey.ShouldEqual
)

func testCallerGeneral(t *testing.T) {

	c := GetCaller(0)
	t.Logf("Got: %v", c)
	so(c.File.Base(), eq, "caller_test.go")
	so(c.Func.Name(), eq, "testCallerGeneral")
	so(c.Func.Package(), eq, "caller")
	// so(c.Func.ReceiverType(), eq, "")
	so(c.Line, eq, 20)

	c = GetCaller(1)
	t.Logf("Got: %v", c)
	so(c.File.Base(), eq, "caller_test.go")
	so(c.Func.Name(), eq, "func1")
	so(c.Func.Package(), eq, "caller")
	so(c.Func.Base(), eq, "caller.TestCaller.func1")
	so(c.Line, eq, 88)
}

func testCallerMethod(t *testing.T) {
	d := dummy{}
	c := d.getCaller()
	t.Logf("Got: %v", c)
	so(c.File.Base(), eq, "caller_test.go")
	so(c.Func.Name(), eq, "getCaller")
	so(c.Func.Package(), eq, "caller")
	so(c.Func.Base(), eq, "caller.dummy.getCaller")
	so(c.Line, eq, 63)

}

func testCallerInClosure(t *testing.T) {
	d := dummy{}
	c := d.getCallerByClosure()
	t.Logf("Got: %v", c)
	so(c.File.Base(), eq, "caller_test.go")
	so(c.Func.Name(), eq, "func1")
	so(c.Func.Package(), eq, "caller")
	so(c.Func.Base(), eq, "caller.dummy.getCallerByClosure.func1")
	so(c.Line, eq, 68)
}

type dummy struct{}

func (dummy) getCaller() Caller {
	return GetCaller(0)
}

func (dummy) getCallerByClosure() Caller {
	c := func() Caller {
		return GetCaller(0)
	}()
	return c
}

func testGetAllCallers(t *testing.T) {
	callers := GetAllCallers()

	b, _ := json.Marshal(callers)
	t.Logf("Got callers: %s", b)
}

func testCallerOutOfRange(t *testing.T) {
	c := func() Caller {
		return GetCaller(100)
	}()
	so(c.Line, eq, -1)
}

func TestCaller(t *testing.T) {
	cv("基础逻辑", t, func() { testCallerGeneral(t) })
	cv("在方法内调用", t, func() { testCallerMethod(t) })
	cv("在闭包内调用", t, func() { testCallerInClosure(t) })
	cv("测试 GetAllCallers", t, func() { testGetAllCallers(t) })
	cv("测试 skip 参数超出范围的情况", t, func() { testCallerOutOfRange(t) })
}
