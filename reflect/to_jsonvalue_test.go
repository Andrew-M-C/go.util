package reflect

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

var (
	cv = convey.Convey
	so = convey.So
	eq = convey.ShouldEqual
)

func TestReadAnyToJsonvalue(t *testing.T) {
	cv("ReadAnyToJsonvalue", t, func() { testReadAnyToJsonvalue(t) })
}

func testReadAnyToJsonvalue(t *testing.T) {
	cv("基本类型", func() {
		// 测试 nil
		jv, err := ReadAnyToJsonvalue(nil, "")
		so(err, eq, nil)
		so(jv.IsNull(), eq, true)

		// 测试 bool
		jv, err = ReadAnyToJsonvalue(true, "")
		so(err, eq, nil)
		so(jv.IsBoolean(), eq, true)
		so(jv.Bool(), eq, true)

		// 测试 false
		jv, err = ReadAnyToJsonvalue(false, "")
		so(err, eq, nil)
		so(jv.IsBoolean(), eq, true)
		so(jv.Bool(), eq, false)

		// 测试 int
		jv, err = ReadAnyToJsonvalue(int64(42), "")
		so(err, eq, nil)
		so(jv.IsNumber(), eq, true)
		so(jv.Int(), eq, int(42))

		// 测试 uint
		jv, err = ReadAnyToJsonvalue(uint64(100), "")
		so(err, eq, nil)
		so(jv.IsNumber(), eq, true)
		so(jv.Uint64(), eq, uint64(100))

		// 测试 float
		jv, err = ReadAnyToJsonvalue(float64(3.14), "")
		so(err, eq, nil)
		so(jv.IsNumber(), eq, true)
		so(jv.Float64(), eq, float64(3.14))

		// 测试 string
		jv, err = ReadAnyToJsonvalue("hello", "")
		so(err, eq, nil)
		so(jv.IsString(), eq, true)
		so(jv.String(), eq, "hello")
	})

	cv("struct 基础功能", func() {
		type Person struct {
			Name string `test:"name"`
			Age  int    `test:"age"`
		}
		person := Person{Name: "Alice", Age: 30}

		// 使用 test tag
		jv, err := ReadAnyToJsonvalue(person, "test")
		so(err, eq, nil)
		so(jv.IsObject(), eq, true)

		name, _ := jv.Get("name")
		so(name.IsString(), eq, true)
		so(name.String(), eq, "Alice")

		age, _ := jv.Get("age")
		so(age.IsNumber(), eq, true)
		so(age.Int(), eq, 30)
	})

	cv("struct 不使用 tag", func() {
		type Person struct {
			Name string
			Age  int
		}
		person := Person{Name: "Bob", Age: 25}

		// 不使用 tag，应该使用字段名
		jv, err := ReadAnyToJsonvalue(person, "")
		so(err, eq, nil)
		so(jv.IsObject(), eq, true)

		name, _ := jv.Get("Name")
		so(name.IsString(), eq, true)
		so(name.String(), eq, "Bob")

		age, _ := jv.Get("Age")
		so(age.IsNumber(), eq, true)
		so(age.Int(), eq, 25)
	})

	cv("struct 指针", func() {
		type Person struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		}
		person := &Person{Name: "Charlie", Age: 35}

		jv, err := ReadAnyToJsonvalue(person, "json")
		so(err, eq, nil)
		so(jv.IsObject(), eq, true)

		name, _ := jv.Get("name")
		so(name.String(), eq, "Charlie")
	})

	cv("slice", func() {
		slice := []int{1, 2, 3, 4, 5}
		jv, err := ReadAnyToJsonvalue(slice, "")
		so(err, eq, nil)
		so(jv.IsArray(), eq, true)
		so(jv.Len(), eq, 5)

		// 验证数组元素
		v0, _ := jv.Get(0)
		so(v0.Int(), eq, 1)
		v4, _ := jv.Get(4)
		so(v4.Int(), eq, 5)
	})

	cv("slice of strings", func() {
		slice := []string{"apple", "banana", "cherry"}
		jv, err := ReadAnyToJsonvalue(slice, "")
		so(err, eq, nil)
		so(jv.IsArray(), eq, true)
		so(jv.Len(), eq, 3)

		v0, _ := jv.Get(0)
		so(v0.String(), eq, "apple")
		v2, _ := jv.Get(2)
		so(v2.String(), eq, "cherry")
	})

	cv("空 slice", func() {
		var slice []int
		jv, err := ReadAnyToJsonvalue(slice, "")
		so(err, eq, nil)
		so(jv.IsArray(), eq, true)
		so(jv.Len(), eq, 0)
	})

	cv("array", func() {
		arr := [3]int{10, 20, 30}
		jv, err := ReadAnyToJsonvalue(arr, "")
		so(err, eq, nil)
		so(jv.IsArray(), eq, true)
		so(jv.Len(), eq, 3)

		v1, _ := jv.Get(1)
		so(v1.Int(), eq, 20)
	})

	cv("map[string]int", func() {
		m := map[string]int{
			"one":   1,
			"two":   2,
			"three": 3,
		}
		jv, err := ReadAnyToJsonvalue(m, "")
		so(err, eq, nil)
		so(jv.IsObject(), eq, true)

		one, _ := jv.Get("one")
		so(one.IsNumber(), eq, true)
		so(one.Int(), eq, 1)

		two, _ := jv.Get("two")
		so(two.Int(), eq, 2)

		three, _ := jv.Get("three")
		so(three.Int(), eq, 3)
	})

	cv("map[int]string", func() {
		m := map[int]string{
			1: "one",
			2: "two",
		}
		jv, err := ReadAnyToJsonvalue(m, "")
		so(err, eq, nil)
		so(jv.IsObject(), eq, true)

		// map 的 key 会被转为字符串
		one, _ := jv.Get("1")
		so(one.String(), eq, "one")

		two, _ := jv.Get("2")
		so(two.String(), eq, "two")
	})

	cv("空 map", func() {
		m := map[string]int{}
		jv, err := ReadAnyToJsonvalue(m, "")
		so(err, eq, nil)
		so(jv.IsObject(), eq, true)
		so(jv.Len(), eq, 0)
	})

	cv("嵌套结构", func() {
		type Address struct {
			City    string `json:"city"`
			Country string `json:"country"`
		}
		type Person struct {
			Name    string  `json:"name"`
			Age     int     `json:"age"`
			Address Address `json:"address"`
		}
		person := Person{
			Name: "Bob",
			Age:  25,
			Address: Address{
				City:    "New York",
				Country: "USA",
			},
		}

		jv, err := ReadAnyToJsonvalue(person, "json")
		so(err, eq, nil)
		so(jv.IsObject(), eq, true)

		name, _ := jv.Get("name")
		so(name.String(), eq, "Bob")

		age, _ := jv.Get("age")
		so(age.Int(), eq, 25)

		address, _ := jv.Get("address")
		so(address.IsObject(), eq, true)

		city, _ := address.Get("city")
		so(city.String(), eq, "New York")

		country, _ := address.Get("country")
		so(country.String(), eq, "USA")
	})

	cv("slice of structs", func() {
		type Item struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		}
		items := []Item{
			{ID: 1, Name: "First"},
			{ID: 2, Name: "Second"},
		}

		jv, err := ReadAnyToJsonvalue(items, "json")
		so(err, eq, nil)
		so(jv.IsArray(), eq, true)
		so(jv.Len(), eq, 2)

		item0, _ := jv.Get(0)
		so(item0.IsObject(), eq, true)
		id0, _ := item0.Get("id")
		so(id0.Int(), eq, 1)
		name0, _ := item0.Get("name")
		so(name0.String(), eq, "First")

		item1, _ := jv.Get(1)
		id1, _ := item1.Get("id")
		so(id1.Int(), eq, 2)
	})

	cv("map with struct values", func() {
		type User struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		}
		m := map[string]User{
			"user1": {Name: "Alice", Age: 30},
			"user2": {Name: "Bob", Age: 25},
		}

		jv, err := ReadAnyToJsonvalue(m, "json")
		so(err, eq, nil)
		so(jv.IsObject(), eq, true)

		user1, _ := jv.Get("user1")
		so(user1.IsObject(), eq, true)
		name1, _ := user1.Get("name")
		so(name1.String(), eq, "Alice")
	})

	cv("nil 字段", func() {
		type Person struct {
			Name    string  `json:"name"`
			Address *string `json:"address"`
		}
		person := Person{Name: "David", Address: nil}

		jv, err := ReadAnyToJsonvalue(person, "json")
		so(err, eq, nil)

		address, _ := jv.Get("address")
		so(address.IsNull(), eq, true)
	})

	cv("指针字段", func() {
		type Person struct {
			Name string `json:"name"`
			Age  *int   `json:"age"`
		}
		age := 28
		person := Person{Name: "Eve", Age: &age}

		jv, err := ReadAnyToJsonvalue(person, "json")
		so(err, eq, nil)

		ageVal, _ := jv.Get("age")
		so(ageVal.IsNumber(), eq, true)
		so(ageVal.Int(), eq, 28)
	})

	cv("interface{} 字段", func() {
		type Container struct {
			Value interface{} `json:"value"`
		}

		// 字符串
		c1 := Container{Value: "hello"}
		jv, err := ReadAnyToJsonvalue(c1, "json")
		so(err, eq, nil)
		val1, _ := jv.Get("value")
		so(val1.String(), eq, "hello")

		// 数字
		c2 := Container{Value: 42}
		jv, err = ReadAnyToJsonvalue(c2, "json")
		so(err, eq, nil)
		val2, _ := jv.Get("value")
		so(val2.Int(), eq, 42)

		// nil
		c3 := Container{Value: nil}
		jv, err = ReadAnyToJsonvalue(c3, "json")
		so(err, eq, nil)
		val3, _ := jv.Get("value")
		so(val3.IsNull(), eq, true)
	})

	cv("嵌套 slice", func() {
		nested := [][]int{
			{1, 2, 3},
			{4, 5},
			{6, 7, 8, 9},
		}

		jv, err := ReadAnyToJsonvalue(nested, "")
		so(err, eq, nil)
		so(jv.IsArray(), eq, true)
		so(jv.Len(), eq, 3)

		arr0, _ := jv.Get(0)
		so(arr0.IsArray(), eq, true)
		so(arr0.Len(), eq, 3)

		arr1, _ := jv.Get(1)
		so(arr1.Len(), eq, 2)
	})

	cv("嵌套 map", func() {
		nested := map[string]map[string]int{
			"group1": {"a": 1, "b": 2},
			"group2": {"x": 10, "y": 20},
		}

		jv, err := ReadAnyToJsonvalue(nested, "")
		so(err, eq, nil)
		so(jv.IsObject(), eq, true)

		group1, _ := jv.Get("group1")
		so(group1.IsObject(), eq, true)
		a, _ := group1.Get("a")
		so(a.Int(), eq, 1)
	})

	cv("复杂混合结构", func() {
		type Meta struct {
			Version string `json:"version"`
			Author  string `json:"author"`
		}
		type Item struct {
			ID    int      `json:"id"`
			Tags  []string `json:"tags"`
			Extra Meta     `json:"extra"`
		}
		type Response struct {
			Status string         `json:"status"`
			Items  []Item         `json:"items"`
			Meta   map[string]int `json:"meta"`
		}

		resp := Response{
			Status: "ok",
			Items: []Item{
				{
					ID:   1,
					Tags: []string{"tag1", "tag2"},
					Extra: Meta{
						Version: "1.0",
						Author:  "Alice",
					},
				},
			},
			Meta: map[string]int{
				"count": 1,
				"total": 100,
			},
		}

		jv, err := ReadAnyToJsonvalue(resp, "json")
		so(err, eq, nil)
		so(jv.IsObject(), eq, true)

		status, _ := jv.Get("status")
		so(status.String(), eq, "ok")

		items, _ := jv.Get("items")
		so(items.IsArray(), eq, true)
		so(items.Len(), eq, 1)

		item0, _ := items.Get(0)
		id, _ := item0.Get("id")
		so(id.Int(), eq, 1)

		tags, _ := item0.Get("tags")
		so(tags.Len(), eq, 2)
		tag0, _ := tags.Get(0)
		so(tag0.String(), eq, "tag1")

		extra, _ := item0.Get("extra")
		version, _ := extra.Get("version")
		so(version.String(), eq, "1.0")

		meta, _ := jv.Get("meta")
		count, _ := meta.Get("count")
		so(count.Int(), eq, 1)
	})
}
