// Package deepcopy 提供一个生成对象深复制的功能
package deepcopy

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/Andrew-M-C/go.util/ternary"
	"github.com/iancoleman/strcase"
)

type kind int

const (
	kindIllegal kind = iota
	kindBasic        // 可以直接赋值的
	kindStruct       // 结构体
	kindSlice
	kindArray
	kindMap
)

func (k kind) String() string {
	switch k {
	default:
		return "不支持的类型"
	case kindBasic:
		return "基本类型"
	case kindStruct:
		return "结构体"
	case kindSlice:
		return "切片"
	case kindArray:
		return "数组"
	case kindMap:
		return "字典"
	}
}

type typeDetail struct {
	Type reflect.Type
	Elem reflect.Type
	Kind kind

	IsPointer bool

	overriddenPackageName string
}

func (t *typeDetail) SelfName() string {
	return strings.ToLower(t.TypeName()[:1])
}

func (t *typeDetail) TypeName() string {
	if t.Kind == kindArray {
		return t.Type.Name()
	}
	return t.Elem.Name()
}

func (t *typeDetail) SetPackageName(n string) {
	if n != "" {
		t.overriddenPackageName = n
	}
}

func (t *typeDetail) PackageName() string {
	if t.overriddenPackageName != "" {
		return t.overriddenPackageName
	}

	s := t.Elem.String()
	if t.Kind == kindArray {
		s = t.Type.String()
	}
	if strings.Contains(s, ".") {
		return strings.TrimSuffix(s, "."+t.TypeName())
	}
	return ""
}

func (t *typeDetail) TypeReferenceName() string {
	packageName := t.PackageName()
	if packageName == "" {
		if t.IsPointer {
			return "*" + t.TypeName()
		}
		return t.TypeName()
	}

	if t.IsPointer {
		return fmt.Sprintf("*%v.%v", packageName, t.TypeName())
	}
	return fmt.Sprintf("%v.%v", packageName, t.TypeName())
}

func (t *typeDetail) FunctionComment() string {
	s := fmt.Sprintf(
		"makes a deep copy of %s%s.%s",
		ternary.If(t.IsPointer, "*", ""), t.PackageName(), t.TypeName(),
	)
	return s
}

func (t *typeDetail) PackagePath() string {
	if p := t.Type.PkgPath(); p != "" {
		return p
	}
	return t.Elem.PkgPath()
}

func (t *typeDetail) CopyFuncName(exportable bool) string {
	packagePart := strcase.ToCamel(t.PackageName())
	typePart := strcase.ToCamel(t.TypeName())

	ptrDesc := ternary.If(t.IsPointer, "Ptr", "")

	if exportable {
		return fmt.Sprintf("DeepCopy%s%s%s", packagePart, typePart, ptrDesc)
	}
	return fmt.Sprintf("deepCopy%s%s%s", packagePart, typePart, ptrDesc)
}

// ================

type structTypeDetail struct {
	typeDetail
	Field reflect.StructField
}
