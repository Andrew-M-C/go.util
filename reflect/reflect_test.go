package reflect_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	jsonvalue "github.com/Andrew-M-C/go.jsonvalue"
	"github.com/Andrew-M-C/go.util/reflect"
	"github.com/smartystreets/goconvey/convey"
)

var (
	cv = convey.Convey
	so = convey.So
	eq = convey.ShouldEqual
)

func TestReflect(t *testing.T) {
	cv("DescribeType", t, func() { testDescribeType(t) })
	cv("ReadAny", t, func() { testReadAny(t) })
	cv("testReadStruct", t, func() { testReadStruct(t) })
	cv("testReadSlice", t, func() { testReadSlice(t) })
	cv("testReadArray", t, func() { testReadArray(t) })
	cv("testReadMap", t, func() { testReadMap(t) })
}

func testDescribeType(t *testing.T) {
	cv("基本类型", func() {
		f := 1.5
		desc := reflect.DescribeType(f)
		so(desc.TypeName, eq, "float64")
		so(desc.PackageName, eq, "")
		so(desc.PointerLevels, eq, 0)
		so(desc.Path.Prefix, eq, "")
		so(desc.Path.Full, eq, "")
	})

	cv("基本类型的指针", func() {
		f := 1.5
		ft := &f
		desc := reflect.DescribeType(ft)
		so(desc.TypeName, eq, "float64")
		so(desc.PackageName, eq, "")
		so(desc.PointerLevels, eq, 1)
		so(desc.Path.Prefix, eq, "")
		so(desc.Path.Full, eq, "")

		ftt := &ft
		desc = reflect.DescribeType(ftt)
		so(desc.TypeName, eq, "float64")
		so(desc.PackageName, eq, "")
		so(desc.PointerLevels, eq, 2)
		so(desc.Path.Prefix, eq, "")
		so(desc.Path.Full, eq, "")
	})

	cv("官方包中的类型", func() {
		tm := time.Time{}
		desc := reflect.DescribeType(tm)
		so(desc.TypeName, eq, "Time")
		so(desc.PackageName, eq, "time")
		so(desc.PointerLevels, eq, 0)
		so(desc.Path.Prefix, eq, "")
		so(desc.Path.Full, eq, "time")

		desc = reflect.DescribeType(&tm)
		so(desc.TypeName, eq, "Time")
		so(desc.PackageName, eq, "time")
		so(desc.PointerLevels, eq, 1)
		so(desc.Path.Prefix, eq, "")
		so(desc.Path.Full, eq, "time")
	})

	cv("原生 interface 类型", func() {
		var err error
		desc := reflect.DescribeType(err)
		so(desc.TypeName, eq, "nil")
		so(desc.PackageName, eq, "")
		so(desc.PointerLevels, eq, 0)
		so(desc.Path.Prefix, eq, "")
		so(desc.Path.Full, eq, "")

		desc = reflect.DescribeType(&err)
		so(desc.TypeName, eq, "error")
		so(desc.PackageName, eq, "")
		so(desc.PointerLevels, eq, 1)
		so(desc.Path.Prefix, eq, "")
		so(desc.Path.Full, eq, "")
	})

	cv("第三方包类型", func() {
		m := convey.StackMode("")
		desc := reflect.DescribeType(m)
		so(desc.TypeName, eq, "StackMode")
		so(desc.PackageName, eq, "convey")
		so(desc.PointerLevels, eq, 0)
		so(desc.Path.Prefix, eq, "github.com/smartystreets/goconvey")
		so(desc.Path.Full, eq, "github.com/smartystreets/goconvey/convey")

		j := jsonvalue.NewObject()
		desc = reflect.DescribeType(j)
		so(desc.TypeName, eq, "V")
		so(desc.PackageName, eq, "jsonvalue")
		so(desc.PointerLevels, eq, 1)
		so(desc.Path.Prefix, eq, "github.com/Andrew-M-C")
		so(desc.Path.Full, eq, "github.com/Andrew-M-C/go.jsonvalue")
	})

	cv("第三方接口类型", func() {
		var intf jsonvalue.Caseless
		desc := reflect.DescribeType(&intf)
		so(desc.TypeName, eq, "Caseless")
		so(desc.PackageName, eq, "jsonvalue")
		so(desc.PointerLevels, eq, 1)
		so(desc.Path.Prefix, eq, "github.com/Andrew-M-C")
		so(desc.Path.Full, eq, "github.com/Andrew-M-C/go.jsonvalue")

		desc = reflect.DescribeType(intf)
		so(desc.TypeName, eq, "nil")
		so(desc.PackageName, eq, "")
		so(desc.PointerLevels, eq, 0)
		so(desc.Path.Prefix, eq, "")
		so(desc.Path.Full, eq, "")
	})

	cv("函数类型", func() {
		_, cancel := context.WithCancel(context.Background())
		desc := reflect.DescribeType(cancel)
		so(desc.TypeName, eq, "CancelFunc")
		so(desc.PackageName, eq, "context")
		so(desc.PointerLevels, eq, 0)
		so(desc.Path.Prefix, eq, "")
		so(desc.Path.Full, eq, "context")

		var f1 func() error
		desc = reflect.DescribeType(f1)
		so(desc.TypeName, eq, "func() error")
		so(desc.PackageName, eq, "")
		so(desc.PointerLevels, eq, 0)
		so(desc.Path.Prefix, eq, "")
		so(desc.Path.Full, eq, "")

		var f2 func(*jsonvalue.V) (any, error)
		desc = reflect.DescribeType(f2)
		so(desc.TypeName, eq, "func(*jsonvalue.V) (interface {}, error)")
		so(desc.PackageName, eq, "")
		so(desc.PointerLevels, eq, 0)
		so(desc.Path.Prefix, eq, "")
		so(desc.Path.Full, eq, "")

		b, _ := json.Marshal(desc)
		t.Log(string(b))
	})
}
