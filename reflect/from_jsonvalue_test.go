package reflect

import (
	"testing"

	jsonvalue "github.com/Andrew-M-C/go.jsonvalue"
)

func TestFromJsonvalueToAny(t *testing.T) {
	cv("FromJsonvalueToAny", t, func() { testFromJsonvalueToAny(t) })
}

func testFromJsonvalueToAny(t *testing.T) {
	cv("基本类型", func() {
		// 测试 bool
		jv := jsonvalue.NewBool(true)
		var result bool
		err := FromJsonvalueToAny(&result, jv, "")
		so(err, eq, nil)
		so(result, eq, true)

		// 测试 false
		jv = jsonvalue.NewBool(false)
		err = FromJsonvalueToAny(&result, jv, "")
		so(err, eq, nil)
		so(result, eq, false)

		// 测试 int
		jv = jsonvalue.NewInt64(42)
		var intResult int
		err = FromJsonvalueToAny(&intResult, jv, "")
		so(err, eq, nil)
		so(intResult, eq, 42)

		// 测试 uint
		jv = jsonvalue.NewUint64(100)
		var uintResult uint
		err = FromJsonvalueToAny(&uintResult, jv, "")
		so(err, eq, nil)
		so(uintResult, eq, uint(100))

		// 测试 float
		jv = jsonvalue.NewFloat64(3.14)
		var floatResult float64
		err = FromJsonvalueToAny(&floatResult, jv, "")
		so(err, eq, nil)
		so(floatResult, eq, 3.14)

		// 测试 string
		jv = jsonvalue.NewString("hello")
		var strResult string
		err = FromJsonvalueToAny(&strResult, jv, "")
		so(err, eq, nil)
		so(strResult, eq, "hello")
	})

	cv("slice 类型", func() {
		arr := jsonvalue.NewArray()
		arr.MustAppend(jsonvalue.NewInt64(1)).InTheEnd()
		arr.MustAppend(jsonvalue.NewInt64(2)).InTheEnd()
		arr.MustAppend(jsonvalue.NewInt64(3)).InTheEnd()

		var result []int
		err := FromJsonvalueToAny(&result, arr, "")
		so(err, eq, nil)
		so(len(result), eq, 3)
		so(result[0], eq, 1)
		so(result[1], eq, 2)
		so(result[2], eq, 3)
	})

	cv("array 类型", func() {
		arr := jsonvalue.NewArray()
		arr.MustAppend(jsonvalue.NewInt64(10)).InTheEnd()
		arr.MustAppend(jsonvalue.NewInt64(20)).InTheEnd()
		arr.MustAppend(jsonvalue.NewInt64(30)).InTheEnd()

		var result [3]int
		err := FromJsonvalueToAny(&result, arr, "")
		so(err, eq, nil)
		so(result[0], eq, 10)
		so(result[1], eq, 20)
		so(result[2], eq, 30)
	})

	cv("map 类型", func() {
		obj := jsonvalue.NewObject()
		obj.MustSet(jsonvalue.NewString("Alice")).At("name")
		obj.MustSet(jsonvalue.NewInt64(25)).At("age")

		var result map[string]any
		err := FromJsonvalueToAny(&result, obj, "")
		so(err, eq, nil)
		so(result["name"], eq, "Alice")
		so(result["age"], eq, int64(25))
	})

	cv("struct 基础功能", func() {
		type Person struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		}

		obj := jsonvalue.NewObject()
		obj.MustSet(jsonvalue.NewString("Bob")).At("name")
		obj.MustSet(jsonvalue.NewInt64(30)).At("age")

		var result Person
		err := FromJsonvalueToAny(&result, obj, "json")
		so(err, eq, nil)
		so(result.Name, eq, "Bob")
		so(result.Age, eq, 30)
	})

	cv("struct 不使用 tag", func() {
		type Person struct {
			Name string
			Age  int
		}

		obj := jsonvalue.NewObject()
		obj.MustSet(jsonvalue.NewString("Charlie")).At("Name")
		obj.MustSet(jsonvalue.NewInt64(28)).At("Age")

		var result Person
		err := FromJsonvalueToAny(&result, obj, "")
		so(err, eq, nil)
		so(result.Name, eq, "Charlie")
		so(result.Age, eq, 28)
	})

	cv("嵌套 struct", func() {
		type Address struct {
			City    string `json:"city"`
			Country string `json:"country"`
		}

		type Person struct {
			Name    string  `json:"name"`
			Age     int     `json:"age"`
			Address Address `json:"address"`
		}

		obj := jsonvalue.NewObject()
		obj.MustSet(jsonvalue.NewString("David")).At("name")
		obj.MustSet(jsonvalue.NewInt64(35)).At("age")

		addrObj := jsonvalue.NewObject()
		addrObj.MustSet(jsonvalue.NewString("New York")).At("city")
		addrObj.MustSet(jsonvalue.NewString("USA")).At("country")
		obj.MustSet(addrObj).At("address")

		var result Person
		err := FromJsonvalueToAny(&result, obj, "json")
		so(err, eq, nil)
		so(result.Name, eq, "David")
		so(result.Age, eq, 35)
		so(result.Address.City, eq, "New York")
		so(result.Address.Country, eq, "USA")
	})

	cv("pointer 类型", func() {
		jv := jsonvalue.NewInt64(100)
		var result *int
		err := FromJsonvalueToAny(&result, jv, "")
		so(err, eq, nil)
		so(result != nil, eq, true)
		so(*result, eq, 100)
	})

	cv("null 值", func() {
		jv := jsonvalue.NewNull()

		// pointer
		var ptrResult *int
		err := FromJsonvalueToAny(&ptrResult, jv, "")
		so(err, eq, nil)
		so(ptrResult == nil, eq, true)

		// slice
		var sliceResult []int
		err = FromJsonvalueToAny(&sliceResult, jv, "")
		so(err, eq, nil)
		so(sliceResult == nil, eq, true)

		// map
		var mapResult map[string]int
		err = FromJsonvalueToAny(&mapResult, jv, "")
		so(err, eq, nil)
		so(mapResult == nil, eq, true)
	})

	cv("interface{} 类型", func() {
		// 测试 object
		obj := jsonvalue.NewObject()
		obj.MustSet(jsonvalue.NewString("test")).At("key")
		obj.MustSet(jsonvalue.NewInt64(123)).At("value")

		var result any
		err := FromJsonvalueToAny(&result, obj, "")
		so(err, eq, nil)

		m, ok := result.(map[string]any)
		so(ok, eq, true)
		so(m["key"], eq, "test")
		so(m["value"], eq, int64(123))

		// 测试 array
		arr := jsonvalue.NewArray()
		arr.MustAppend(jsonvalue.NewInt64(1)).InTheEnd()
		arr.MustAppend(jsonvalue.NewString("two")).InTheEnd()
		arr.MustAppend(jsonvalue.NewBool(true)).InTheEnd()

		var arrResult any
		err = FromJsonvalueToAny(&arrResult, arr, "")
		so(err, eq, nil)

		arrVal, ok := arrResult.([]any)
		so(ok, eq, true)
		so(len(arrVal), eq, 3)
		so(arrVal[0], eq, int64(1))
		so(arrVal[1], eq, "two")
		so(arrVal[2], eq, true)
	})

	cv("错误处理", func() {
		// nil target
		jv := jsonvalue.NewInt64(42)
		err := FromJsonvalueToAny(nil, jv, "")
		so(err != nil, eq, true)

		// non-pointer target
		var result int
		err = FromJsonvalueToAny(result, jv, "")
		so(err != nil, eq, true)

		// nil jsonvalue
		err = FromJsonvalueToAny(&result, nil, "")
		so(err != nil, eq, true)

		// 类型不匹配
		strJv := jsonvalue.NewString("not a number")
		var numResult int
		err = FromJsonvalueToAny(&numResult, strJv, "")
		so(err != nil, eq, true)
	})

	cv("往返转换测试", func() {
		type TestStruct struct {
			Name    string   `json:"name"`
			Age     int      `json:"age"`
			Email   string   `json:"email"`
			Tags    []string `json:"tags"`
			Enabled bool     `json:"enabled"`
		}

		original := TestStruct{
			Name:    "Test User",
			Age:     25,
			Email:   "test@example.com",
			Tags:    []string{"tag1", "tag2", "tag3"},
			Enabled: true,
		}

		// Convert to jsonvalue
		jv, err := ReadAnyToJsonvalue(original, "json")
		so(err, eq, nil)

		// Convert back
		var result TestStruct
		err = FromJsonvalueToAny(&result, jv, "json")
		so(err, eq, nil)

		// Compare
		so(result.Name, eq, original.Name)
		so(result.Age, eq, original.Age)
		so(result.Email, eq, original.Email)
		so(len(result.Tags), eq, len(original.Tags))
		for i := range original.Tags {
			so(result.Tags[i], eq, original.Tags[i])
		}
		so(result.Enabled, eq, original.Enabled)
	})

	cv("复杂嵌套往返转换", func() {
		type Inner struct {
			Value int `test:"val"`
		}
		type Middle struct {
			Inner  Inner    `test:"inner"`
			Values []string `test:"values"`
		}
		type Outer struct {
			Middle Middle         `test:"middle"`
			Map    map[string]int `test:"map"`
		}

		original := Outer{
			Middle: Middle{
				Inner:  Inner{Value: 42},
				Values: []string{"a", "b", "c"},
			},
			Map: map[string]int{"x": 1, "y": 2},
		}

		// To jsonvalue
		jv, err := ReadAnyToJsonvalue(original, "test")
		so(err, eq, nil)

		// Back to struct
		var result Outer
		err = FromJsonvalueToAny(&result, jv, "test")
		so(err, eq, nil)

		so(result.Middle.Inner.Value, eq, original.Middle.Inner.Value)
		so(len(result.Middle.Values), eq, len(original.Middle.Values))
		so(result.Middle.Values[0], eq, "a")
		so(result.Middle.Values[1], eq, "b")
		so(result.Middle.Values[2], eq, "c")
		// Note: map 的遍历顺序不确定，只检查值是否存在
		so(result.Map["x"], eq, 1)
		so(result.Map["y"], eq, 2)
	})

	cv("匿名嵌入 struct - 简单场景", func() {
		// 定义一个嵌入的基础结构体
		type Address struct {
			City    string `json:"city"`
			Country string `json:"country"`
		}

		// Person 匿名嵌入 Address，这样 Person 就可以直接访问 City 和 Country
		type Person struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
			Address
		}

		// JSON 结构应该是展平的:
		// {"name": "Alice", "age": 30, "city": "Beijing", "country": "China"}
		obj := jsonvalue.NewObject()
		obj.MustSet(jsonvalue.NewString("Alice")).At("name")
		obj.MustSet(jsonvalue.NewInt64(30)).At("age")
		obj.MustSet(jsonvalue.NewString("Beijing")).At("city")
		obj.MustSet(jsonvalue.NewString("China")).At("country")

		var result Person
		err := FromJsonvalueToAny(&result, obj, "json")
		so(err, eq, nil)
		so(result.Name, eq, "Alice")
		so(result.Age, eq, 30)
		so(result.City, eq, "Beijing")
		so(result.Country, eq, "China")
	})

	cv("匿名嵌入 struct - 多个嵌入", func() {
		type ContactInfo struct {
			Email string `json:"email"`
			Phone string `json:"phone"`
		}

		type SocialInfo struct {
			Twitter  string `json:"twitter"`
			LinkedIn string `json:"linkedin"`
		}

		type User struct {
			Name string `json:"name"`
			ContactInfo
			SocialInfo
		}

		// 展平的 JSON 结构
		obj := jsonvalue.NewObject()
		obj.MustSet(jsonvalue.NewString("Bob")).At("name")
		obj.MustSet(jsonvalue.NewString("bob@example.com")).At("email")
		obj.MustSet(jsonvalue.NewString("123-456-7890")).At("phone")
		obj.MustSet(jsonvalue.NewString("@bob")).At("twitter")
		obj.MustSet(jsonvalue.NewString("bob-linkedin")).At("linkedin")

		var result User
		err := FromJsonvalueToAny(&result, obj, "json")
		so(err, eq, nil)
		so(result.Name, eq, "Bob")
		so(result.Email, eq, "bob@example.com")
		so(result.Phone, eq, "123-456-7890")
		so(result.Twitter, eq, "@bob")
		so(result.LinkedIn, eq, "bob-linkedin")
	})

	cv("匿名嵌入 struct - 指针嵌入", func() {
		type Metadata struct {
			CreatedAt string `json:"created_at"`
			UpdatedAt string `json:"updated_at"`
		}

		type Document struct {
			Title string `json:"title"`
			*Metadata
		}

		// 展平的 JSON 结构
		obj := jsonvalue.NewObject()
		obj.MustSet(jsonvalue.NewString("My Document")).At("title")
		obj.MustSet(jsonvalue.NewString("2024-01-01")).At("created_at")
		obj.MustSet(jsonvalue.NewString("2024-06-15")).At("updated_at")

		var result Document
		err := FromJsonvalueToAny(&result, obj, "json")
		so(err, eq, nil)
		so(result.Title, eq, "My Document")
		so(result.Metadata != nil, eq, true)
		so(result.CreatedAt, eq, "2024-01-01")
		so(result.UpdatedAt, eq, "2024-06-15")
	})

	cv("匿名嵌入 struct - 嵌套匿名嵌入", func() {
		type Level3 struct {
			L3Field string `json:"l3_field"`
		}

		type Level2 struct {
			L2Field string `json:"l2_field"`
			Level3
		}

		type Level1 struct {
			L1Field string `json:"l1_field"`
			Level2
		}

		// 展平的 JSON 结构，所有字段都在顶层
		obj := jsonvalue.NewObject()
		obj.MustSet(jsonvalue.NewString("value1")).At("l1_field")
		obj.MustSet(jsonvalue.NewString("value2")).At("l2_field")
		obj.MustSet(jsonvalue.NewString("value3")).At("l3_field")

		var result Level1
		err := FromJsonvalueToAny(&result, obj, "json")
		so(err, eq, nil)
		so(result.L1Field, eq, "value1")
		so(result.L2Field, eq, "value2")
		so(result.L3Field, eq, "value3")
	})

	cv("匿名嵌入 struct - 字段名冲突（外层优先）", func() {
		type Base struct {
			Name  string `json:"name"`
			Value int    `json:"value"`
		}

		type Extended struct {
			Name string `json:"name"` // 与 Base.Name 冲突，外层优先
			Base
		}

		obj := jsonvalue.NewObject()
		obj.MustSet(jsonvalue.NewString("ExtendedName")).At("name")
		obj.MustSet(jsonvalue.NewInt64(100)).At("value")

		var result Extended
		err := FromJsonvalueToAny(&result, obj, "json")
		so(err, eq, nil)
		so(result.Name, eq, "ExtendedName") // 外层的 Name
		so(result.Value, eq, 100)           // 来自 Base
		// Base.Name 应该也被设置（因为 JSON key 相同）
		so(result.Base.Name, eq, "ExtendedName")
	})

	cv("匿名嵌入 struct - 往返转换", func() {
		type Embedded struct {
			EmbeddedField1 string `json:"embedded_field1"`
			EmbeddedField2 int    `json:"embedded_field2"`
		}

		type Container struct {
			ContainerField string `json:"container_field"`
			Embedded
		}

		original := Container{
			ContainerField: "container_value",
			Embedded: Embedded{
				EmbeddedField1: "embedded_value",
				EmbeddedField2: 42,
			},
		}

		// Convert to jsonvalue
		jv, err := ReadAnyToJsonvalue(original, "json")
		so(err, eq, nil)

		// Convert back
		var result Container
		err = FromJsonvalueToAny(&result, jv, "json")
		so(err, eq, nil)

		// Compare
		so(result.ContainerField, eq, original.ContainerField)
		so(result.EmbeddedField1, eq, original.EmbeddedField1)
		so(result.EmbeddedField2, eq, original.EmbeddedField2)
	})

	cv("匿名嵌入 struct - 不使用 tag", func() {
		type Base struct {
			BaseField string
		}

		type Derived struct {
			DerivedField int
			Base
		}

		obj := jsonvalue.NewObject()
		obj.MustSet(jsonvalue.NewInt64(999)).At("DerivedField")
		obj.MustSet(jsonvalue.NewString("base_value")).At("BaseField")

		var result Derived
		err := FromJsonvalueToAny(&result, obj, "")
		so(err, eq, nil)
		so(result.DerivedField, eq, 999)
		so(result.BaseField, eq, "base_value")
	})
}
