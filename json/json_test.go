package json_test

import (
	"os"
	"path/filepath"
	"testing"

	jsonvalue "github.com/Andrew-M-C/go.jsonvalue"
	"github.com/Andrew-M-C/go.util/json"
	"github.com/smartystreets/goconvey/convey"
)

var (
	cv = convey.Convey
	so = convey.So
	eq = convey.ShouldEqual

	isNil    = convey.ShouldBeNil
	notNil   = convey.ShouldNotBeNil
	isTrue   = convey.ShouldBeTrue
	isFalse  = convey.ShouldBeFalse
	contains = convey.ShouldContainSubstring
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func TestGetJSONFromFile(t *testing.T) {
	cv("GetJSONFromFile", t, func() {
		cv("文件不存在时返回读文件错误", func() {
			v, err := json.GetJSONFromFile(filepath.Join(t.TempDir(), "not-exist.json"))
			so(v, isNil)
			so(err, notNil)
			so(err.Error(), contains, "read file error")
		})

		cv("文件内容不是合法 JSON 时返回反序列化错误", func() {
			p := writeTempFile(t, "invalid.json", "this is not json")
			v, err := json.GetJSONFromFile(p)
			so(v, isNil)
			so(err, notNil)
			so(err.Error(), contains, "unmarshal json error")
		})

		cv("不带路径参数时返回完整 JSON", func() {
			p := writeTempFile(t, "full.json", `{"user":{"name":"Alice","age":18},"tags":["a","b"]}`)
			v, err := json.GetJSONFromFile(p)
			so(err, isNil)
			so(v, notNil)
			so(v.IsObject(), isTrue)
			so(v.MustGet("user", "name").String(), eq, "Alice")
			so(v.MustGet("user", "age").Int(), eq, 18)
			so(v.MustGet("tags", 1).String(), eq, "b")
		})

		cv("带路径参数时返回对应子值", func() {
			p := writeTempFile(t, "child.json", `{
				"user": {"name": "Bob", "pets": ["cat", "dog"]},
				"ok": true
			}`)

			name, err := json.GetJSONFromFile(p, "user", "name")
			so(err, isNil)
			so(name.IsString(), isTrue)
			so(name.String(), eq, "Bob")

			pet, err := json.GetJSONFromFile(p, "user", "pets", 0)
			so(err, isNil)
			so(pet.String(), eq, "cat")

			ok, err := json.GetJSONFromFile(p, "ok")
			so(err, isNil)
			so(ok.Bool(), isTrue)
		})

		cv("路径不存在时返回取值错误", func() {
			p := writeTempFile(t, "missing-key.json", `{"user":{"name":"Carol"}}`)
			v, err := json.GetJSONFromFile(p, "user", "age")
			so(v, isNil)
			so(err, notNil)
			so(err.Error(), contains, "get json value error")
		})
	})
}

func TestMustGetJSONFromFile(t *testing.T) {
	cv("MustGetJSONFromFile", t, func() {
		cv("失败时返回 JSON null, 且保证非 nil", func() {
			v := json.MustGetJSONFromFile(filepath.Join(t.TempDir(), "not-exist.json"))
			so(v, notNil)
			so(v.IsNull(), isTrue)

			p := writeTempFile(t, "invalid.json", "{")
			v = json.MustGetJSONFromFile(p)
			so(v, notNil)
			so(v.IsNull(), isTrue)

			p = writeTempFile(t, "ok.json", `{"k":1}`)
			v = json.MustGetJSONFromFile(p, "missing")
			so(v, notNil)
			so(v.IsNull(), isTrue)
		})

		cv("成功时返回对应值", func() {
			p := writeTempFile(t, "ok.json", `{"msg":"hello","n":7}`)
			whole := json.MustGetJSONFromFile(p)
			so(whole, notNil)
			so(whole.IsNull(), isFalse)
			so(whole.IsObject(), isTrue)
			so(whole.MustGet("msg").String(), eq, "hello")

			n := json.MustGetJSONFromFile(p, "n")
			so(n.IsNull(), isFalse)
			so(n.Int(), eq, 7)
		})
	})
}

func TestSetJSONToFile(t *testing.T) {
	cv("SetJSONToFile", t, func() {
		cv("值为 nil 时直接报错, 不写文件", func() {
			p := filepath.Join(t.TempDir(), "nil.json")
			err := json.SetJSONToFile(p, nil)
			so(err, notNil)
			so(err.Error(), eq, "JSON value is nil")
			_, statErr := os.Stat(p)
			so(os.IsNotExist(statErr), isTrue)
		})

		cv("目标文件不存在时, 不带路径则将 object 字段写入新文件", func() {
			p := filepath.Join(t.TempDir(), "new.json")
			src := jsonvalue.MustUnmarshalString(`{"hello":"世界","n":1}`)
			err := json.SetJSONToFile(p, src)
			so(err, isNil)

			got, err := json.GetJSONFromFile(p)
			so(err, isNil)
			so(got.IsObject(), isTrue)
			so(got.MustGet("hello").String(), eq, "世界")
			so(got.MustGet("n").Int(), eq, 1)

			// OptUTF8: 中文不应被转成 \uXXXX
			raw, err := os.ReadFile(p)
			so(err, isNil)
			so(string(raw), contains, "世界")
			so(string(raw), convey.ShouldNotContainSubstring, `\u`)
		})

		cv("目标文件不存在时, 按路径写入并自动创建 object", func() {
			p := filepath.Join(t.TempDir(), "nested.json")
			err := json.SetJSONToFile(p, jsonvalue.NewString("Alice"), "user", "name")
			so(err, isNil)

			name, err := json.GetJSONFromFile(p, "user", "name")
			so(err, isNil)
			so(name.String(), eq, "Alice")
		})

		cv("目标已存在时, 按路径合并写入, 保留其它字段", func() {
			p := writeTempFile(t, "merge.json", `{"user":{"name":"Bob","age":20},"keep":true}`)
			err := json.SetJSONToFile(p, jsonvalue.NewInt(21), "user", "age")
			so(err, isNil)

			got, err := json.GetJSONFromFile(p)
			so(err, isNil)
			so(got.MustGet("user", "name").String(), eq, "Bob")
			so(got.MustGet("user", "age").Int(), eq, 21)
			so(got.MustGet("keep").Bool(), isTrue)
		})

		cv("目标已存在时, 不带路径则按第一层 key 浅合并", func() {
			p := writeTempFile(t, "merge-root.json", `{
				"keep": true,
				"n": 1,
				"user": {"name": "Bob", "age": 20},
				"tags": ["a", "b"]
			}`)
			incoming := jsonvalue.MustUnmarshalString(`{
				"n": 2,
				"user": {"name": "Alice"},
				"tags": ["c"],
				"extra": "added"
			}`)
			err := json.SetJSONToFile(p, incoming)
			so(err, isNil)

			got, err := json.GetJSONFromFile(p)
			so(err, isNil)
			so(got.IsObject(), isTrue)

			// 未出现在 incoming 中的第一层字段应保留
			so(got.MustGet("keep").Bool(), isTrue)

			// 同名第一层标量 / 数组被整体覆盖
			so(got.MustGet("n").Int(), eq, 2)
			so(got.MustGet("tags").IsArray(), isTrue)
			so(got.MustGet("tags").Len(), eq, 1)
			so(got.MustGet("tags", 0).String(), eq, "c")

			// 同名第一层 object 是浅合并: 整个子对象被替换, 而不是递归深合并
			so(got.MustGet("user", "name").String(), eq, "Alice")
			_, err = got.Get("user", "age")
			so(err, notNil)

			// incoming 中的新 key 应写入
			so(got.MustGet("extra").String(), eq, "added")
		})

		cv("目标已存在时, 不带路径多次合并会累加字段", func() {
			p := writeTempFile(t, "merge-twice.json", `{"a":1}`)
			err := json.SetJSONToFile(p, jsonvalue.MustUnmarshalString(`{"b":2}`))
			so(err, isNil)
			err = json.SetJSONToFile(p, jsonvalue.MustUnmarshalString(`{"c":3,"a":10}`))
			so(err, isNil)

			got, err := json.GetJSONFromFile(p)
			so(err, isNil)
			so(got.MustGet("a").Int(), eq, 10)
			so(got.MustGet("b").Int(), eq, 2)
			so(got.MustGet("c").Int(), eq, 3)
			so(got.Len(), eq, 3)
		})

		cv("不带路径时若传入空 object, 则已有字段保持不变", func() {
			p := writeTempFile(t, "merge-empty.json", `{"old":1,"keep":"yes"}`)
			err := json.SetJSONToFile(p, jsonvalue.NewObject())
			so(err, isNil)

			got, err := json.GetJSONFromFile(p)
			so(err, isNil)
			so(got.MustGet("old").Int(), eq, 1)
			so(got.MustGet("keep").String(), eq, "yes")
			so(got.Len(), eq, 2)
		})

		cv("不带路径时若传入非 object, 则不改变已有内容", func() {
			p := writeTempFile(t, "merge-non-object.json", `{"old":1}`)
			err := json.SetJSONToFile(p, jsonvalue.MustUnmarshalString(`["x","y"]`))
			so(err, isNil)

			got, err := json.GetJSONFromFile(p)
			so(err, isNil)
			so(got.IsObject(), isTrue)
			so(got.MustGet("old").Int(), eq, 1)
			so(got.Len(), eq, 1)

			err = json.SetJSONToFile(p, jsonvalue.NewString("ignored"))
			so(err, isNil)
			got, err = json.GetJSONFromFile(p)
			so(err, isNil)
			so(got.MustGet("old").Int(), eq, 1)
			so(got.Len(), eq, 1)
		})

		cv("原文件是非法 JSON 时视为空 object 再写入", func() {
			p := writeTempFile(t, "broken.json", "not-json")
			err := json.SetJSONToFile(p, jsonvalue.NewString("ok"), "status")
			so(err, isNil)

			got, err := json.GetJSONFromFile(p, "status")
			so(err, isNil)
			so(got.String(), eq, "ok")
		})

		cv("原文件是 JSON null 时视为空 object 再写入", func() {
			p := writeTempFile(t, "null.json", "null")
			err := json.SetJSONToFile(p, jsonvalue.NewBool(true), "ok")
			so(err, isNil)
			got, err := json.GetJSONFromFile(p, "ok")
			so(err, isNil)
			so(got.Bool(), isTrue)
		})

		cv("写入路径的父目录不存在时返回写文件错误", func() {
			p := filepath.Join(t.TempDir(), "no-such-dir", "a.json")
			err := json.SetJSONToFile(p, jsonvalue.NewObject())
			so(err, notNil)
		})

		cv("目标路径是目录时返回写文件错误", func() {
			dir := t.TempDir()
			err := json.SetJSONToFile(dir, jsonvalue.NewString("x"))
			so(err, notNil)
		})
	})
}

func writeTempFile(t *testing.T, name, content string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(p, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	return p
}
