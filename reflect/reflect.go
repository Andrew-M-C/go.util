// Package reflect 基于 reflect 库实现一些方便的功能
package reflect

import (
	"path"
	"reflect"
	"strings"

	jsonvalue "github.com/Andrew-M-C/go.jsonvalue"
)

// TypeDesc 描述一个类型的各种信息
type TypeDesc struct {
	// 类型名称
	TypeName string `json:"type_name"`
	// package 名称
	PackageName string `json:"package_name"`
	// 指针层级, 也就是 * 的个数
	PointerLevels int `json:"pointer_levels"`
	// 包路径信息
	Path struct {
		Prefix string `json:"prefix,omitempty"` // 相当于 dir 部分
		Full   string `json:"full,omitempty"`   // 完整的 path + package 路径名
	} `json:"path"`

	kind reflect.Kind
}

// Kind 返回类型的 Kind
func (t TypeDesc) Kind() reflect.Kind {
	return t.kind
}

func (t TypeDesc) MarshalJSON() ([]byte, error) {
	j := jsonvalue.NewObject()
	j.MustSetString(t.TypeName).At("type_name")
	j.MustSetString(t.PackageName).At("package_name")
	j.MustSetInt(t.PointerLevels).At("pointer_levels")

	if t.Path.Prefix != "" {
		j.MustSetString(t.Path.Prefix).At("path", "prefix")
	}
	if t.Path.Full != "" {
		j.MustSetString(t.Path.Prefix).At("path", "full")
	}

	j.MustSetString(t.kind.String()).At("kind")

	return j.Marshal(jsonvalue.OptSetSequence())
}

// DescribeType 描述一个类型
func DescribeType(v any) TypeDesc {
	if v == nil {
		return describeNilType()
	}
	typ := reflect.TypeOf(v)
	res := describeType(typ)
	res.kind = typ.Kind()
	return res
}

func describeNilType() (desc TypeDesc) {
	desc.TypeName = "nil"
	return
}

func describeType(typ reflect.Type) (desc TypeDesc) {
	if typ.Kind() == reflect.Pointer {
		desc = describeType(typ.Elem())
		desc.PointerLevels++
		return
	}

	s := typ.String()

	if typ.Name() != "" {
		parts := strings.Split(s, ".")
		if len(parts) == 1 {
			// 原生类型
			desc.TypeName = s
		} else {
			desc.PackageName = parts[0]
			desc.TypeName = parts[1]
		}
	}

	desc.TypeName = typ.Name()
	if desc.TypeName == "" {
		desc.TypeName = s
	}

	desc.Path.Full = typ.PkgPath()
	desc.Path.Prefix = path.Dir(typ.PkgPath())

	if desc.Path.Prefix == "." {
		desc.Path.Prefix = ""
	}

	return
}
