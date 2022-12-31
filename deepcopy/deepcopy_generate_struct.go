package deepcopy

import (
	"fmt"
	"reflect"
	"strings"
)

func (g *generator) handleTargetKindStruct() error {
	var iterField func(middle string, ft reflect.StructField)
	iterField = func(middle string, ft reflect.StructField) {
		switch ft.Type.Kind() {
		default:
			// basic copyable type
			g.addCodeLine(`	cpy%s.%s = %s%s.%s`, middle, ft.Name, g.typSelf, middle, ft.Name)

		case reflect.Invalid, reflect.Chan, reflect.UnsafePointer, reflect.Uintptr:
			g.infof("skip un-copyable field %s.%s.%s, type '%v'", g.typ.PkgPath(), g.typ.Name(), ft.Name, ft.Type)

		case reflect.Array:
			// TODO: 也要检查这个 array 的内容物是不是合法的
		case reflect.Slice:
			elem, level, isBasic, copyable := g.findFinalElementForPointerKind(ft.Type)
			if !copyable {
				g.infof("skip un-copyable field %s.%s.%s, type '%v'", g.typ.PkgPath(), g.typ.Name(), ft.Name, ft.Type)
				return
			}
			level--
			if level > 1 {
				g.infof("skip pointer to pointer field %s.%s.%s, type '%v'", g.typ.PkgPath(), g.typ.Name(), ft.Name, ft.Type)
				return
			}

			elemPkgNameSect := ""
			if elem.PkgPath() != g.typ.PkgPath() {
				elemPkgName := elem.String()
				elemPkgName = strings.TrimSuffix(elemPkgName, fmt.Sprintf(".%s", elem.Name()))
				g.addImportLine(`%s "%s"`, elemPkgName, elem.PkgPath())
				elemPkgNameSect = elemPkgName + "."
			}

			g.addCodeLine(
				`	cpy%s.%s = make([]%s%s%s, 0, len(%s%s.%s))`,
				middle, ft.Name,
				strings.Repeat("*", level), elemPkgNameSect, elem.Name(),
				g.typSelf, middle, ft.Name,
			)

			_ = isBasic
			// TODO:

		case reflect.Map:
			// TODO: 也要检查这个 map 的内容物是不是合法的

		case reflect.Pointer:
			elem, level, isBasic, copyable := g.findFinalElementForPointerKind(ft.Type)
			if !copyable {
				g.infof("skip un-copyable field %s.%s.%s, type '%v'", g.typ.PkgPath(), g.typ.Name(), ft.Name, ft.Type)
				return
			}

			ifLine := strings.Builder{}
			ifLine.WriteString(`	if`)
			for i := 0; i < level; i++ {
				if i > 0 {
					ifLine.WriteString(" &&")
				}
				ifLine.WriteString(fmt.Sprintf(" %s%s.%s != nil", strings.Repeat("*", i+1), g.typSelf, ft.Name))
			}
			ifLine.WriteString(" {")

			g.addCodeLine(ifLine.String())
			if isBasic {
				g.addCodeLine(`		tmp%s := %s%s%s.%s`, ft.Name, strings.Repeat("*", level), g.typSelf, middle, ft.Name)
			} else {
				g.requestGenerateType(elem)
				copyFuncName := subTypeCopyFunctionName(g.typ, elem)
				g.addCodeLine(`		tmp%s := %s(%s%s%s.%s)`, ft.Name, copyFuncName, strings.Repeat("*", level), g.typSelf, middle, ft.Name)
			}

			g.addCodeLine(`		cpy%s.%s = %stmp%s`, middle, ft.Name, strings.Repeat("&", level), ft.Name)
			g.addCodeLine(`	}`)

			// TODO:

		case reflect.Struct:
			fields := g.readStructFields(ft.Type)
			for _, fft := range fields {
				iterField(middle+"."+ft.Name, fft)
			}
		}

		// TODO:
	}

	// 分析各字段
	for _, ft := range g.readStructFields(g.typ) {
		iterField("", ft)
	}

	// TODO:
	g.markTypeGenerated(g.typ)
	return nil
}

// readStructFields 读取一个 type 的所有同级 fields, 主要是削平匿名字段
func (g *generator) readStructFields(typ reflect.Type) []reflect.StructField {
	res := make([]reflect.StructField, 0, typ.NumField())
	for i := 0; i < typ.NumField(); i++ {
		ft := typ.Field(i)

		if !ft.IsExported() {
			g.debugf("skip un exported field %v", ft)
			continue
		}
		if ft.Anonymous {
			fields := g.readStructFields(ft.Type)
			res = append(res, fields...)
			continue
		}

		res = append(res, ft)
	}
	return res
}
