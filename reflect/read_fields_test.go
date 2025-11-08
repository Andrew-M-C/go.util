package reflect_test

import (
	"testing"

	"github.com/Andrew-M-C/go.util/reflect"
)

func testReadAny(*testing.T) {
	cv("定点数", func() {
		r := reflect.NewSimpleValueReader()
		reflect.ReadAny(int32(-4321), r)
		so(len(r.Results), eq, 1)
		res := r.Results[0]
		so(res.Index, eq, -1)
		so(res.Desc.TypeDesc.TypeName, eq, "int32")
		so(res.Read.Int64, eq, true)
		so(res.Value.Int64, eq, int64(-4321))

		r.Clear()
		i := int32(-9876)
		ptr := &i
		reflect.ReadAny(ptr, r)
		so(len(r.Results), eq, 1)
		res = r.Results[0]
		so(res.Index, eq, -1)
		so(res.Desc.TypeDesc.TypeName, eq, "int32")
		so(res.Read.Int64, eq, true)
		so(res.Value.Int64, eq, int64(-9876))
		so(res.Desc.TypeDesc.PointerLevels, eq, 1)

		r.Clear()
		reflect.ReadAny(&ptr, r)
		so(len(r.Results), eq, 1)
		res = r.Results[0]
		so(res.Index, eq, -1)
		so(res.Desc.TypeDesc.TypeName, eq, "int32")
		so(res.Read.Int64, eq, true)
		so(res.Value.Int64, eq, int64(-9876))
		so(res.Desc.TypeDesc.PointerLevels, eq, 2)
	})
}

func testReadStruct(*testing.T) {
	type st struct {
		Int   int
		Str   string
		Sub   *st
		Float float64
		Bool  bool
		Uint  uint32
		IPtr  *int
	}

	cv("基础功能", func() {
		intVal := 555
		s := st{
			Int: 123,
			Str: "456",
			Sub: &st{
				Int: 789,
			},
			Float: 3.14,
			Bool:  true,
			Uint:  999,
			IPtr:  &intVal,
		}
		r := reflect.NewSimpleValueReader()
		reflect.ReadStruct(s, r)
		so(len(r.Results), eq, 7)
		so(r.Results[0].Index, eq, 0)
		so(r.Results[0].Desc.TypeDesc.TypeName, eq, "int")
		so(r.Results[0].Read.Int64, eq, true)
		so(r.Results[0].Value.Int64, eq, int64(123))
		so(r.Results[1].Index, eq, 1)
		so(r.Results[1].Desc.TypeDesc.TypeName, eq, "string")
		so(r.Results[1].Read.String, eq, true)
		so(r.Results[1].Value.String, eq, "456")
		so(r.Results[2].Index, eq, 2)
		so(r.Results[2].Desc.TypeDesc.TypeName, eq, "st")
		so(r.Results[2].Desc.TypeDesc.PointerLevels, eq, 1)
		so(r.Results[3].Index, eq, 3)
		so(r.Results[3].Desc.TypeDesc.TypeName, eq, "float64")
		so(r.Results[3].Read.Float64, eq, true)
		so(r.Results[3].Value.Float64, eq, 3.14)
		so(r.Results[4].Index, eq, 4)
		so(r.Results[4].Desc.TypeDesc.TypeName, eq, "bool")
		so(r.Results[4].Read.Bool, eq, true)
		so(r.Results[4].Value.Bool, eq, true)
		so(r.Results[5].Index, eq, 5)
		so(r.Results[5].Desc.TypeDesc.TypeName, eq, "uint32")
		so(r.Results[5].Read.Uint64, eq, true)
		so(r.Results[5].Value.Uint64, eq, uint64(999))
		so(r.Results[6].Index, eq, 6)
		so(r.Results[6].Desc.TypeDesc.TypeName, eq, "int")
		so(r.Results[6].Desc.TypeDesc.PointerLevels, eq, 1)
		so(r.Results[6].Read.Int64, eq, true)
		so(r.Results[6].Value.Int64, eq, int64(555))
	})

	cv("空结构体", func() {
		type empty struct{}
		e := empty{}
		r := reflect.NewSimpleValueReader()
		reflect.ReadStruct(e, r)
		so(len(r.Results), eq, 0)
	})

	cv("嵌套结构体", func() {
		type inner struct {
			Name string
		}
		type outer struct {
			ID    int
			Inner inner
		}
		o := outer{
			ID: 1,
			Inner: inner{
				Name: "test",
			},
		}
		r := reflect.NewSimpleValueReader()
		reflect.ReadStruct(o, r)
		so(len(r.Results), eq, 2)
		so(r.Results[0].Index, eq, 0)
		so(r.Results[0].Desc.TypeDesc.TypeName, eq, "int")
		so(r.Results[0].Value.Int64, eq, int64(1))
		so(r.Results[1].Index, eq, 1)
		so(r.Results[1].Desc.TypeDesc.TypeName, eq, "inner")
		so(r.Results[1].Read.Struct, eq, true)
	})

	cv("多级指针字段", func() {
		type ptrSt struct {
			Num int
		}
		n := 456
		p1 := &n
		p2 := &p1
		s := struct {
			Ptr **int
		}{
			Ptr: p2,
		}
		r := reflect.NewSimpleValueReader()
		reflect.ReadStruct(s, r)
		so(len(r.Results), eq, 1)
		so(r.Results[0].Desc.TypeDesc.TypeName, eq, "int")
		so(r.Results[0].Desc.TypeDesc.PointerLevels, eq, 2)
		so(r.Results[0].Value.Int64, eq, int64(456))
	})

	cv("传入非结构体类型", func() {
		r := reflect.NewSimpleValueReader()
		reflect.ReadStruct(123, r)
		so(len(r.Results), eq, 0)

		r.Clear()
		reflect.ReadStruct("string", r)
		so(len(r.Results), eq, 0)

		r.Clear()
		reflect.ReadStruct([]int{1, 2, 3}, r)
		so(len(r.Results), eq, 0)
	})

	cv("传入nil", func() {
		r := reflect.NewSimpleValueReader()
		reflect.ReadStruct(nil, r)
		so(len(r.Results), eq, 0)
	})

	cv("nil指针字段", func() {
		type withNilPtr struct {
			Name   string
			Ptr    *int
			PtrStr *string
		}
		s := withNilPtr{
			Name:   "test",
			Ptr:    nil,
			PtrStr: nil,
		}
		r := reflect.NewSimpleValueReader()
		reflect.ReadStruct(s, r)
		so(len(r.Results), eq, 3)
		so(r.Results[0].Desc.TypeDesc.TypeName, eq, "string")
		so(r.Results[0].Value.String, eq, "test")
		// nil 指针字段会触发 ReadNil
		so(r.Results[1].Desc.TypeDesc.TypeName, eq, "int")
		so(r.Results[1].Desc.TypeDesc.PointerLevels, eq, 1)
		so(r.Results[1].Read.Nil, eq, true)
		so(r.Results[2].Desc.TypeDesc.TypeName, eq, "string")
		so(r.Results[2].Desc.TypeDesc.PointerLevels, eq, 1)
		so(r.Results[2].Read.Nil, eq, true)
	})

	cv("字段标签", func() {
		type tagged struct {
			Field string `json:"field_name" db:"field_db"`
		}
		t := tagged{Field: "value"}
		r := reflect.NewSimpleValueReader()
		reflect.ReadStruct(t, r)
		so(len(r.Results), eq, 1)
		so(r.Results[0].Desc.StructFieldName, eq, "Field")
		so(r.Results[0].Desc.StructTag.Get("json"), eq, "field_name")
		so(r.Results[0].Desc.StructTag.Get("db"), eq, "field_db")
	})

	cv("any类型字段", func() {
		type withAny struct {
			IntField    any
			StringField any
			FloatField  any
			BoolField   any
			NilField    any
			StructField any
		}
		s := withAny{
			IntField:    42,
			StringField: "test",
			FloatField:  1.23,
			BoolField:   true,
			NilField:    nil,
			StructField: struct{ X int }{X: 99},
		}
		r := reflect.NewSimpleValueReader()
		reflect.ReadStruct(s, r)
		so(len(r.Results), eq, 6)

		// 第一个字段: int 类型
		so(r.Results[0].Index, eq, 0)
		so(r.Results[0].Desc.StructFieldName, eq, "IntField")
		so(r.Results[0].Desc.TypeDesc.TypeName, eq, "int")
		so(r.Results[0].Read.Int64, eq, true)
		so(r.Results[0].Value.Int64, eq, int64(42))

		// 第二个字段: string 类型
		so(r.Results[1].Index, eq, 1)
		so(r.Results[1].Desc.StructFieldName, eq, "StringField")
		so(r.Results[1].Desc.TypeDesc.TypeName, eq, "string")
		so(r.Results[1].Read.String, eq, true)
		so(r.Results[1].Value.String, eq, "test")

		// 第三个字段: float64 类型
		so(r.Results[2].Index, eq, 2)
		so(r.Results[2].Desc.StructFieldName, eq, "FloatField")
		so(r.Results[2].Desc.TypeDesc.TypeName, eq, "float64")
		so(r.Results[2].Read.Float64, eq, true)
		so(r.Results[2].Value.Float64, eq, 1.23)

		// 第四个字段: bool 类型
		so(r.Results[3].Index, eq, 3)
		so(r.Results[3].Desc.StructFieldName, eq, "BoolField")
		so(r.Results[3].Desc.TypeDesc.TypeName, eq, "bool")
		so(r.Results[3].Read.Bool, eq, true)
		so(r.Results[3].Value.Bool, eq, true)

		// 第五个字段: nil
		so(r.Results[4].Index, eq, 4)
		so(r.Results[4].Desc.StructFieldName, eq, "NilField")
		so(r.Results[4].Read.Nil, eq, true)

		// 第六个字段: 匿名结构体
		so(r.Results[5].Index, eq, 5)
		so(r.Results[5].Desc.StructFieldName, eq, "StructField")
		so(r.Results[5].Read.Struct, eq, true)
	})
}

func testReadSlice(*testing.T) {
	cv("基础切片", func() {
		slice := []int{1, 2, 3, 4, 5}
		r := reflect.NewSimpleValueReader()
		reflect.ReadSlice(slice, r)
		so(len(r.Results), eq, 5)
		for i := 0; i < 5; i++ {
			so(r.Results[i].Index, eq, i)
			so(r.Results[i].Desc.TypeDesc.TypeName, eq, "int")
			so(r.Results[i].Read.Int64, eq, true)
			so(r.Results[i].Value.Int64, eq, int64(i+1))
		}
	})

	cv("字符串切片", func() {
		slice := []string{"hello", "world", "test"}
		r := reflect.NewSimpleValueReader()
		reflect.ReadSlice(slice, r)
		so(len(r.Results), eq, 3)
		so(r.Results[0].Value.String, eq, "hello")
		so(r.Results[1].Value.String, eq, "world")
		so(r.Results[2].Value.String, eq, "test")
	})

	cv("指针切片", func() {
		a, b, c := 10, 20, 30
		slice := []*int{&a, &b, &c}
		r := reflect.NewSimpleValueReader()
		reflect.ReadSlice(slice, r)
		so(len(r.Results), eq, 3)
		so(r.Results[0].Desc.TypeDesc.TypeName, eq, "int")
		so(r.Results[0].Desc.TypeDesc.PointerLevels, eq, 1)
		so(r.Results[0].Value.Int64, eq, int64(10))
		so(r.Results[1].Value.Int64, eq, int64(20))
		so(r.Results[2].Value.Int64, eq, int64(30))
	})

	cv("结构体切片", func() {
		type item struct {
			ID   int
			Name string
		}
		slice := []item{
			{ID: 1, Name: "first"},
			{ID: 2, Name: "second"},
		}
		r := reflect.NewSimpleValueReader()
		reflect.ReadSlice(slice, r)
		so(len(r.Results), eq, 2)
		so(r.Results[0].Index, eq, 0)
		so(r.Results[0].Desc.TypeDesc.TypeName, eq, "item")
		so(r.Results[0].Read.Struct, eq, true)
		so(r.Results[1].Index, eq, 1)
		so(r.Results[1].Desc.TypeDesc.TypeName, eq, "item")
	})

	cv("空切片", func() {
		slice := []int{}
		r := reflect.NewSimpleValueReader()
		reflect.ReadSlice(slice, r)
		so(len(r.Results), eq, 0)
	})

	cv("nil切片", func() {
		var slice []int
		r := reflect.NewSimpleValueReader()
		reflect.ReadSlice(slice, r)
		so(len(r.Results), eq, 0)
	})

	cv("传入非切片类型", func() {
		r := reflect.NewSimpleValueReader()
		reflect.ReadSlice(123, r)
		so(len(r.Results), eq, 0)

		r.Clear()
		reflect.ReadSlice("string", r)
		so(len(r.Results), eq, 0)

		r.Clear()
		type st struct{ A int }
		reflect.ReadSlice(st{A: 1}, r)
		so(len(r.Results), eq, 0)
	})

	cv("传入nil", func() {
		r := reflect.NewSimpleValueReader()
		reflect.ReadSlice(nil, r)
		so(len(r.Results), eq, 0)
	})

	cv("浮点数切片", func() {
		slice := []float64{1.1, 2.2, 3.3}
		r := reflect.NewSimpleValueReader()
		reflect.ReadSlice(slice, r)
		so(len(r.Results), eq, 3)
		so(r.Results[0].Value.Float64, eq, 1.1)
		so(r.Results[1].Value.Float64, eq, 2.2)
		so(r.Results[2].Value.Float64, eq, 3.3)
	})

	cv("布尔切片", func() {
		slice := []bool{true, false, true}
		r := reflect.NewSimpleValueReader()
		reflect.ReadSlice(slice, r)
		so(len(r.Results), eq, 3)
		so(r.Results[0].Value.Bool, eq, true)
		so(r.Results[1].Value.Bool, eq, false)
		so(r.Results[2].Value.Bool, eq, true)
	})

	cv("any切片", func() {
		slice := []any{123, "hello", 3.14, true, nil}
		r := reflect.NewSimpleValueReader()
		reflect.ReadSlice(slice, r)
		so(len(r.Results), eq, 5)

		// 第一个元素: int
		so(r.Results[0].Index, eq, 0)
		so(r.Results[0].Desc.TypeDesc.TypeName, eq, "int")
		so(r.Results[0].Read.Int64, eq, true)
		so(r.Results[0].Value.Int64, eq, int64(123))

		// 第二个元素: string
		so(r.Results[1].Index, eq, 1)
		so(r.Results[1].Desc.TypeDesc.TypeName, eq, "string")
		so(r.Results[1].Read.String, eq, true)
		so(r.Results[1].Value.String, eq, "hello")

		// 第三个元素: float64
		so(r.Results[2].Index, eq, 2)
		so(r.Results[2].Desc.TypeDesc.TypeName, eq, "float64")
		so(r.Results[2].Read.Float64, eq, true)
		so(r.Results[2].Value.Float64, eq, 3.14)

		// 第四个元素: bool
		so(r.Results[3].Index, eq, 3)
		so(r.Results[3].Desc.TypeDesc.TypeName, eq, "bool")
		so(r.Results[3].Read.Bool, eq, true)
		so(r.Results[3].Value.Bool, eq, true)

		// 第五个元素: nil
		so(r.Results[4].Index, eq, 4)
		so(r.Results[4].Read.Nil, eq, true)
	})
}

func testReadArray(*testing.T) {
	cv("基础数组", func() {
		arr := [5]int{1, 2, 3, 4, 5}
		r := reflect.NewSimpleValueReader()
		reflect.ReadArray(arr, r)
		so(len(r.Results), eq, 5)
		for i := 0; i < 5; i++ {
			so(r.Results[i].Index, eq, i)
			so(r.Results[i].Desc.TypeDesc.TypeName, eq, "int")
			so(r.Results[i].Read.Int64, eq, true)
			so(r.Results[i].Value.Int64, eq, int64(i+1))
		}
	})

	cv("字符串数组", func() {
		arr := [3]string{"hello", "world", "test"}
		r := reflect.NewSimpleValueReader()
		reflect.ReadArray(arr, r)
		so(len(r.Results), eq, 3)
		so(r.Results[0].Value.String, eq, "hello")
		so(r.Results[1].Value.String, eq, "world")
		so(r.Results[2].Value.String, eq, "test")
	})

	cv("指针数组", func() {
		a, b, c := 100, 200, 300
		arr := [3]*int{&a, &b, &c}
		r := reflect.NewSimpleValueReader()
		reflect.ReadArray(arr, r)
		so(len(r.Results), eq, 3)
		so(r.Results[0].Desc.TypeDesc.TypeName, eq, "int")
		so(r.Results[0].Desc.TypeDesc.PointerLevels, eq, 1)
		so(r.Results[0].Value.Int64, eq, int64(100))
		so(r.Results[1].Value.Int64, eq, int64(200))
		so(r.Results[2].Value.Int64, eq, int64(300))
	})

	cv("结构体数组", func() {
		type item struct {
			ID   int
			Name string
		}
		arr := [2]item{
			{ID: 1, Name: "first"},
			{ID: 2, Name: "second"},
		}
		r := reflect.NewSimpleValueReader()
		reflect.ReadArray(arr, r)
		so(len(r.Results), eq, 2)
		so(r.Results[0].Index, eq, 0)
		so(r.Results[0].Desc.TypeDesc.TypeName, eq, "item")
		so(r.Results[0].Read.Struct, eq, true)
		so(r.Results[1].Index, eq, 1)
		so(r.Results[1].Desc.TypeDesc.TypeName, eq, "item")
	})

	cv("零长度数组", func() {
		arr := [0]int{}
		r := reflect.NewSimpleValueReader()
		reflect.ReadArray(arr, r)
		so(len(r.Results), eq, 0)
	})

	cv("传入非数组类型", func() {
		r := reflect.NewSimpleValueReader()
		reflect.ReadArray(123, r)
		so(len(r.Results), eq, 0)

		r.Clear()
		reflect.ReadArray("string", r)
		so(len(r.Results), eq, 0)

		r.Clear()
		reflect.ReadArray([]int{1, 2, 3}, r)
		so(len(r.Results), eq, 0)
	})

	cv("传入nil", func() {
		r := reflect.NewSimpleValueReader()
		reflect.ReadArray(nil, r)
		so(len(r.Results), eq, 0)
	})

	cv("浮点数数组", func() {
		arr := [3]float32{1.1, 2.2, 3.3}
		r := reflect.NewSimpleValueReader()
		reflect.ReadArray(arr, r)
		so(len(r.Results), eq, 3)
		so(r.Results[0].Read.Float32, eq, true)
		so(r.Results[0].Value.Float32, eq, float32(1.1))
		so(r.Results[1].Value.Float32, eq, float32(2.2))
		so(r.Results[2].Value.Float32, eq, float32(3.3))
	})

	cv("布尔数组", func() {
		arr := [4]bool{true, false, true, false}
		r := reflect.NewSimpleValueReader()
		reflect.ReadArray(arr, r)
		so(len(r.Results), eq, 4)
		so(r.Results[0].Value.Bool, eq, true)
		so(r.Results[1].Value.Bool, eq, false)
		so(r.Results[2].Value.Bool, eq, true)
		so(r.Results[3].Value.Bool, eq, false)
	})

	cv("uint类型数组", func() {
		arr := [3]uint64{100, 200, 300}
		r := reflect.NewSimpleValueReader()
		reflect.ReadArray(arr, r)
		so(len(r.Results), eq, 3)
		so(r.Results[0].Read.Uint64, eq, true)
		so(r.Results[0].Value.Uint64, eq, uint64(100))
		so(r.Results[1].Value.Uint64, eq, uint64(200))
		so(r.Results[2].Value.Uint64, eq, uint64(300))
	})

	cv("any数组", func() {
		arr := [5]any{999, "world", 2.718, false, struct{ X int }{X: 42}}
		r := reflect.NewSimpleValueReader()
		reflect.ReadArray(arr, r)
		so(len(r.Results), eq, 5)

		// 第一个元素: int
		so(r.Results[0].Index, eq, 0)
		so(r.Results[0].Desc.TypeDesc.TypeName, eq, "int")
		so(r.Results[0].Read.Int64, eq, true)
		so(r.Results[0].Value.Int64, eq, int64(999))

		// 第二个元素: string
		so(r.Results[1].Index, eq, 1)
		so(r.Results[1].Desc.TypeDesc.TypeName, eq, "string")
		so(r.Results[1].Read.String, eq, true)
		so(r.Results[1].Value.String, eq, "world")

		// 第三个元素: float64
		so(r.Results[2].Index, eq, 2)
		so(r.Results[2].Desc.TypeDesc.TypeName, eq, "float64")
		so(r.Results[2].Read.Float64, eq, true)
		so(r.Results[2].Value.Float64, eq, 2.718)

		// 第四个元素: bool
		so(r.Results[3].Index, eq, 3)
		so(r.Results[3].Desc.TypeDesc.TypeName, eq, "bool")
		so(r.Results[3].Read.Bool, eq, true)
		so(r.Results[3].Value.Bool, eq, false)

		// 第五个元素: 匿名结构体
		so(r.Results[4].Index, eq, 4)
		so(r.Results[4].Read.Struct, eq, true)
	})
}

func testReadMap(*testing.T) {
	cv("基础map - string到int", func() {
		m := map[string]int{
			"one":   1,
			"two":   2,
			"three": 3,
		}
		r := reflect.NewSimpleValueReader()
		reflect.ReadMap(m, r)
		so(len(r.Results), eq, 3)

		// 验证每个结果的基本属性
		for _, res := range r.Results {
			so(res.Index, eq, -1) // map 的 index 应该为 -1
			so(res.Desc.TypeDesc.TypeName, eq, "int")
			so(res.Read.Int64, eq, true)

			// 验证 MapKey 和 MapValue
			key, ok := res.Desc.MapKey.(string)
			so(ok, eq, true)
			value, ok := res.Desc.MapValue.(int)
			so(ok, eq, true)

			// 验证键值对应关系
			switch key {
			case "one":
				so(res.Value.Int64, eq, int64(1))
				so(value, eq, 1)
			case "two":
				so(res.Value.Int64, eq, int64(2))
				so(value, eq, 2)
			case "three":
				so(res.Value.Int64, eq, int64(3))
				so(value, eq, 3)
			}
		}
	})

	cv("int到string的map", func() {
		m := map[int]string{
			1: "one",
			2: "two",
			3: "three",
		}
		r := reflect.NewSimpleValueReader()
		reflect.ReadMap(m, r)
		so(len(r.Results), eq, 3)

		for _, res := range r.Results {
			so(res.Index, eq, -1)
			so(res.Desc.TypeDesc.TypeName, eq, "string")
			so(res.Read.String, eq, true)

			key, ok := res.Desc.MapKey.(int)
			so(ok, eq, true)
			value, ok := res.Desc.MapValue.(string)
			so(ok, eq, true)

			switch key {
			case 1:
				so(res.Value.String, eq, "one")
				so(value, eq, "one")
			case 2:
				so(res.Value.String, eq, "two")
				so(value, eq, "two")
			case 3:
				so(res.Value.String, eq, "three")
				so(value, eq, "three")
			}
		}
	})

	cv("指针值的map", func() {
		v1, v2, v3 := 100, 200, 300
		m := map[string]*int{
			"a": &v1,
			"b": &v2,
			"c": &v3,
		}
		r := reflect.NewSimpleValueReader()
		reflect.ReadMap(m, r)
		so(len(r.Results), eq, 3)

		for _, res := range r.Results {
			so(res.Index, eq, -1)
			so(res.Desc.TypeDesc.TypeName, eq, "int")
			so(res.Desc.TypeDesc.PointerLevels, eq, 1)
			so(res.Read.Int64, eq, true)

			key, ok := res.Desc.MapKey.(string)
			so(ok, eq, true)

			switch key {
			case "a":
				so(res.Value.Int64, eq, int64(100))
			case "b":
				so(res.Value.Int64, eq, int64(200))
			case "c":
				so(res.Value.Int64, eq, int64(300))
			}
		}
	})

	cv("结构体值的map", func() {
		type person struct {
			Name string
			Age  int
		}
		m := map[string]person{
			"alice": {Name: "Alice", Age: 30},
			"bob":   {Name: "Bob", Age: 25},
		}
		r := reflect.NewSimpleValueReader()
		reflect.ReadMap(m, r)
		so(len(r.Results), eq, 2)

		for _, res := range r.Results {
			so(res.Index, eq, -1)
			so(res.Desc.TypeDesc.TypeName, eq, "person")
			so(res.Read.Struct, eq, true)

			key, ok := res.Desc.MapKey.(string)
			so(ok, eq, true)
			value, ok := res.Desc.MapValue.(person)
			so(ok, eq, true)

			switch key {
			case "alice":
				so(value.Name, eq, "Alice")
				so(value.Age, eq, 30)
			case "bob":
				so(value.Name, eq, "Bob")
				so(value.Age, eq, 25)
			}
		}
	})

	cv("空map", func() {
		m := map[string]int{}
		r := reflect.NewSimpleValueReader()
		reflect.ReadMap(m, r)
		so(len(r.Results), eq, 0)
	})

	cv("nil map", func() {
		var m map[string]int
		r := reflect.NewSimpleValueReader()
		reflect.ReadMap(m, r)
		so(len(r.Results), eq, 0)
	})

	cv("传入nil", func() {
		r := reflect.NewSimpleValueReader()
		reflect.ReadMap(nil, r)
		so(len(r.Results), eq, 0)
	})

	cv("传入非map类型", func() {
		r := reflect.NewSimpleValueReader()
		reflect.ReadMap(123, r)
		so(len(r.Results), eq, 0)

		r.Clear()
		reflect.ReadMap("string", r)
		so(len(r.Results), eq, 0)

		r.Clear()
		reflect.ReadMap([]int{1, 2, 3}, r)
		so(len(r.Results), eq, 0)

		r.Clear()
		type st struct{ A int }
		reflect.ReadMap(st{A: 1}, r)
		so(len(r.Results), eq, 0)
	})

	cv("指针指向map", func() {
		m := map[string]int{
			"x": 10,
			"y": 20,
		}
		ptr := &m
		r := reflect.NewSimpleValueReader()
		reflect.ReadMap(ptr, r)
		so(len(r.Results), eq, 2)

		for _, res := range r.Results {
			so(res.Index, eq, -1)
			so(res.Desc.TypeDesc.TypeName, eq, "int")
			so(res.Read.Int64, eq, true)

			key, ok := res.Desc.MapKey.(string)
			so(ok, eq, true)

			switch key {
			case "x":
				so(res.Value.Int64, eq, int64(10))
			case "y":
				so(res.Value.Int64, eq, int64(20))
			}
		}
	})

	cv("浮点数值的map", func() {
		m := map[string]float64{
			"pi":  3.14,
			"e":   2.71,
			"phi": 1.618,
		}
		r := reflect.NewSimpleValueReader()
		reflect.ReadMap(m, r)
		so(len(r.Results), eq, 3)

		for _, res := range r.Results {
			so(res.Index, eq, -1)
			so(res.Desc.TypeDesc.TypeName, eq, "float64")
			so(res.Read.Float64, eq, true)

			key, ok := res.Desc.MapKey.(string)
			so(ok, eq, true)

			switch key {
			case "pi":
				so(res.Value.Float64, eq, 3.14)
			case "e":
				so(res.Value.Float64, eq, 2.71)
			case "phi":
				so(res.Value.Float64, eq, 1.618)
			}
		}
	})

	cv("布尔值的map", func() {
		m := map[string]bool{
			"enabled":  true,
			"disabled": false,
			"active":   true,
		}
		r := reflect.NewSimpleValueReader()
		reflect.ReadMap(m, r)
		so(len(r.Results), eq, 3)

		for _, res := range r.Results {
			so(res.Index, eq, -1)
			so(res.Desc.TypeDesc.TypeName, eq, "bool")
			so(res.Read.Bool, eq, true)

			key, ok := res.Desc.MapKey.(string)
			so(ok, eq, true)

			switch key {
			case "enabled":
				so(res.Value.Bool, eq, true)
			case "disabled":
				so(res.Value.Bool, eq, false)
			case "active":
				so(res.Value.Bool, eq, true)
			}
		}
	})

	cv("uint类型值的map", func() {
		m := map[string]uint64{
			"small":  100,
			"medium": 1000,
			"large":  10000,
		}
		r := reflect.NewSimpleValueReader()
		reflect.ReadMap(m, r)
		so(len(r.Results), eq, 3)

		for _, res := range r.Results {
			so(res.Index, eq, -1)
			so(res.Desc.TypeDesc.TypeName, eq, "uint64")
			so(res.Read.Uint64, eq, true)

			key, ok := res.Desc.MapKey.(string)
			so(ok, eq, true)

			switch key {
			case "small":
				so(res.Value.Uint64, eq, uint64(100))
			case "medium":
				so(res.Value.Uint64, eq, uint64(1000))
			case "large":
				so(res.Value.Uint64, eq, uint64(10000))
			}
		}
	})

	cv("切片值的map", func() {
		m := map[string][]int{
			"a": {1, 2, 3},
			"b": {4, 5, 6},
		}
		r := reflect.NewSimpleValueReader()
		reflect.ReadMap(m, r)
		so(len(r.Results), eq, 2)

		for _, res := range r.Results {
			so(res.Index, eq, -1)
			so(res.Desc.TypeDesc.TypeName, eq, "[]int")
			so(res.Read.Slice, eq, true)

			key, ok := res.Desc.MapKey.(string)
			so(ok, eq, true)
			value, ok := res.Desc.MapValue.([]int)
			so(ok, eq, true)

			switch key {
			case "a":
				so(len(value), eq, 3)
				so(value[0], eq, 1)
				so(value[1], eq, 2)
				so(value[2], eq, 3)
			case "b":
				so(len(value), eq, 3)
				so(value[0], eq, 4)
				so(value[1], eq, 5)
				so(value[2], eq, 6)
			}
		}
	})

	cv("嵌套map", func() {
		m := map[string]map[string]int{
			"group1": {"a": 1, "b": 2},
			"group2": {"c": 3, "d": 4},
		}
		r := reflect.NewSimpleValueReader()
		reflect.ReadMap(m, r)
		so(len(r.Results), eq, 2)

		for _, res := range r.Results {
			so(res.Index, eq, -1)
			so(res.Desc.TypeDesc.TypeName, eq, "map[string]int")
			so(res.Read.Map, eq, true)

			key, ok := res.Desc.MapKey.(string)
			so(ok, eq, true)
			value, ok := res.Desc.MapValue.(map[string]int)
			so(ok, eq, true)

			switch key {
			case "group1":
				so(len(value), eq, 2)
				so(value["a"], eq, 1)
				so(value["b"], eq, 2)
			case "group2":
				so(len(value), eq, 2)
				so(value["c"], eq, 3)
				so(value["d"], eq, 4)
			}
		}
	})

	cv("int类型作为键", func() {
		m := map[int]string{
			100: "hundred",
			200: "two hundred",
			300: "three hundred",
		}
		r := reflect.NewSimpleValueReader()
		reflect.ReadMap(m, r)
		so(len(r.Results), eq, 3)

		for _, res := range r.Results {
			so(res.Index, eq, -1)
			so(res.Desc.TypeDesc.TypeName, eq, "string")
			so(res.Read.String, eq, true)

			key, ok := res.Desc.MapKey.(int)
			so(ok, eq, true)

			switch key {
			case 100:
				so(res.Value.String, eq, "hundred")
			case 200:
				so(res.Value.String, eq, "two hundred")
			case 300:
				so(res.Value.String, eq, "three hundred")
			}
		}
	})

	cv("复杂类型组合", func() {
		type item struct {
			ID   int
			Tags []string
		}
		m := map[int]item{
			1: {ID: 1, Tags: []string{"tag1", "tag2"}},
			2: {ID: 2, Tags: []string{"tag3", "tag4"}},
		}
		r := reflect.NewSimpleValueReader()
		reflect.ReadMap(m, r)
		so(len(r.Results), eq, 2)

		for _, res := range r.Results {
			so(res.Index, eq, -1)
			so(res.Desc.TypeDesc.TypeName, eq, "item")
			so(res.Read.Struct, eq, true)

			key, ok := res.Desc.MapKey.(int)
			so(ok, eq, true)
			value, ok := res.Desc.MapValue.(item)
			so(ok, eq, true)

			switch key {
			case 1:
				so(value.ID, eq, 1)
				so(len(value.Tags), eq, 2)
				so(value.Tags[0], eq, "tag1")
				so(value.Tags[1], eq, "tag2")
			case 2:
				so(value.ID, eq, 2)
				so(len(value.Tags), eq, 2)
				so(value.Tags[0], eq, "tag3")
				so(value.Tags[1], eq, "tag4")
			}
		}
	})

	cv("map[string]any - 混合类型值", func() {
		m := map[string]any{
			"int":    123,
			"string": "hello",
			"float":  3.14,
			"bool":   true,
			"nil":    nil,
			"struct": struct{ Name string }{Name: "test"},
		}
		r := reflect.NewSimpleValueReader()
		reflect.ReadMap(m, r)
		so(len(r.Results), eq, 6)

		// 验证每个结果
		for _, res := range r.Results {
			so(res.Index, eq, -1)

			key, ok := res.Desc.MapKey.(string)
			so(ok, eq, true)

			switch key {
			case "int":
				so(res.Desc.TypeDesc.TypeName, eq, "int")
				so(res.Read.Int64, eq, true)
				so(res.Value.Int64, eq, int64(123))
			case "string":
				so(res.Desc.TypeDesc.TypeName, eq, "string")
				so(res.Read.String, eq, true)
				so(res.Value.String, eq, "hello")
			case "float":
				so(res.Desc.TypeDesc.TypeName, eq, "float64")
				so(res.Read.Float64, eq, true)
				so(res.Value.Float64, eq, 3.14)
			case "bool":
				so(res.Desc.TypeDesc.TypeName, eq, "bool")
				so(res.Read.Bool, eq, true)
				so(res.Value.Bool, eq, true)
			case "nil":
				so(res.Read.Nil, eq, true)
			case "struct":
				so(res.Read.Struct, eq, true)
			}
		}
	})

	cv("map[int]any - 包含nil值", func() {
		m := map[int]any{
			1: 100,
			2: "text",
			3: nil,
			4: 2.5,
		}
		r := reflect.NewSimpleValueReader()
		reflect.ReadMap(m, r)
		so(len(r.Results), eq, 4)

		nilFound := false
		for _, res := range r.Results {
			so(res.Index, eq, -1)
			key, ok := res.Desc.MapKey.(int)
			so(ok, eq, true)

			if key == 3 {
				so(res.Read.Nil, eq, true)
				nilFound = true
			}
		}
		so(nilFound, eq, true)
	})

	cv("map[any]any - nil 作为 key", func() {
		m := map[any]any{
			nil:      "value for nil key",
			"string": "string value",
			42:       "int value",
		}
		r := reflect.NewSimpleValueReader()
		reflect.ReadMap(m, r)
		so(len(r.Results), eq, 3)

		// 查找 nil key 对应的结果
		nilKeyFound := false
		for _, res := range r.Results {
			so(res.Index, eq, -1)

			// 检查是否是 nil key
			if res.Desc.MapKey == nil {
				nilKeyFound = true
				so(res.Desc.TypeDesc.TypeName, eq, "string")
				so(res.Read.String, eq, true)
				so(res.Value.String, eq, "value for nil key")
			}
		}
		so(nilKeyFound, eq, true)
	})

	cv("map[any]any - nil pointer 作为 key", func() {
		var nilPtr *int
		m := map[any]any{
			nilPtr: "value for nil pointer key",
			"key":  "normal value",
		}
		r := reflect.NewSimpleValueReader()
		reflect.ReadMap(m, r)
		so(len(r.Results), eq, 2)

		// 验证 nil pointer key 的处理
		nilPtrKeyFound := false
		for _, res := range r.Results {
			so(res.Index, eq, -1)

			key := res.Desc.MapKey
			if key != nil {
				if ptr, ok := key.(*int); ok && ptr == nil {
					nilPtrKeyFound = true
					so(res.Desc.TypeDesc.TypeName, eq, "string")
					so(res.Read.String, eq, true)
					so(res.Value.String, eq, "value for nil pointer key")
				}
			}
		}
		so(nilPtrKeyFound, eq, true)
	})
}
