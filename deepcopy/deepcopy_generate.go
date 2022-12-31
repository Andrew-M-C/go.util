// Package deepcopy 用于分析并输出一个 struct 的深复制代码
package deepcopy

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/iancoleman/strcase"
)

type generator struct {
	typ        reflect.Type
	typPkgName string
	typSelf    string
	isPtr      bool
	ptrChar    string

	funcOrMethodName string

	log   func(string, ...interface{})
	debug bool

	packageName string

	importLines []string
	codeLines   []string

	subTypesToDo map[reflect.Type]struct{}
	subTypesDone map[reflect.Type]struct{}

	deferFunctions []func()
}

func (g *generator) infof(f string, a ...interface{}) {
	g.log(f, a...)
}

func (g *generator) debugf(f string, a ...interface{}) {
	if g.debug {
		g.log("DEBUG: "+f, a...)
	}
}

func (g *generator) addImportLine(f string, a ...interface{}) {
	g.importLines = append(g.importLines, fmt.Sprintf("\t"+f, a...))
}

func (g *generator) addCodeLine(f string, a ...interface{}) {
	g.codeLines = append(g.codeLines, fmt.Sprintf(f, a...))
}

func newGenerator(prototype interface{}) *generator {
	g := &generator{}
	g.typ = reflect.TypeOf(prototype)
	g.log = func(string, ...interface{}) {}
	g.packageName = "deepcopy"
	return g
}

func (g *generator) do() (err error) {
	defer func() {
		if err != nil {
			return
		}
		for i := len(g.deferFunctions) - 1; i >= 0; i-- {
			f := g.deferFunctions[i]
			f()
		}
	}()

	pipeline := []func() error{
		g.readTypeNames,
		g.printTypeNames,
		g.genFunctionLineFunctionMode,
		g.handleTargetType,
		// TODO:
	}
	for _, p := range pipeline {
		if err = p(); err != nil {
			return
		}
	}

	return nil
}

// readTypeNames 读取各种与类型相关的名称
func (g *generator) readTypeNames() error {
	switch g.typ.Kind() {
	default:
		return fmt.Errorf("unsupported type: '%v'", g.typ)
	case reflect.Pointer:
		if g.isPtr {
			return fmt.Errorf("pointer to pointer type (*%v) is not supported", g.typ)
		}
		g.typ = g.typ.Elem()
		g.isPtr = true
		g.ptrChar = "*"
		return g.readTypeNames()
	case reflect.Struct, reflect.Map, reflect.Slice, reflect.Array:
		// go on
	}

	name := g.typ.String()
	g.typPkgName = strings.TrimSuffix(name, fmt.Sprintf(".%s", g.typ.Name()))
	g.typSelf = strings.ToLower(g.typ.Name()[:1])

	return nil
}

// printTypeNames 普通地输出结果
func (g *generator) printTypeNames() error {
	if g.isPtr {
		g.debugf("full type path: %v.*%v", g.typ.PkgPath(), g.typ.Name())
	} else {
		g.debugf("full type path: %v.%v", g.typ.PkgPath(), g.typ.Name())
	}

	g.debugf("type package name: %v", g.typPkgName)
	return nil
}

func (g *generator) genFunctionLineFunctionMode() error {
	if g.funcOrMethodName == "" {
		g.funcOrMethodName = fmt.Sprintf("DeepCopy%s%s", strcase.ToCamel(g.typPkgName), g.typ.Name())
	}

	g.addCodeLine(`// %s make a deep copy of %s.%s%s`, g.funcOrMethodName, g.typ.PkgPath(), g.ptrChar, g.typ.Name())

	g.addCodeLine(
		`func %s(%s %s%s) %s%s {`,
		g.funcOrMethodName, g.typSelf, g.ptrChar, g.typ.Name(), g.ptrChar, g.typ.Name(),
	)

	if g.isPtr {
		g.addCodeLine(`	cpy := &%s{}`, g.typ.Name())
	} else {
		g.addCodeLine(`	cpy := %s{}`, g.typ.Name())
	}

	g.deferFunctions = append(g.deferFunctions, func() {
		g.addCodeLine(`	return cpy`)
		g.addCodeLine(`}`)
	})

	// 函数模式, 需要加 import 行
	g.addImportLine(`%s "%s"`, g.typPkgName, g.typ.PkgPath())

	return nil
}

func (g *generator) handleTargetType() error {
	switch g.typ.Kind() {
	default:
		return fmt.Errorf("unsupported type %s.%s%s", g.typ.PkgPath(), g.ptrChar, g.typ.Name())
	case reflect.Struct:
		return g.handleTargetKindStruct()
		// TODO:
	}
}

func (g *generator) markTypeGenerated(typ reflect.Type) {
	if g.subTypesDone == nil {
		g.subTypesDone = make(map[reflect.Type]struct{}, 1)
	}
	if g.subTypesToDo != nil {
		delete(g.subTypesToDo, typ)
	}
	g.subTypesDone[typ] = struct{}{}
}

func (g *generator) requestGenerateType(typ reflect.Type) {
	if _, exist := g.subTypesDone[typ]; exist {
		return
	}
	if g.subTypesToDo == nil {
		g.subTypesToDo = make(map[reflect.Type]struct{}, 1)
	}
	g.subTypesToDo[typ] = struct{}{}
}

func (g *generator) packageResult() string {
	bdr := strings.Builder{}

	// package 行
	bdr.WriteString(fmt.Sprintf("package %s\n\n", g.packageName))

	// import 行
	if len(g.importLines) > 0 {
		bdr.WriteString("import (\n")
		for _, line := range g.importLines {
			bdr.WriteString(line)
			bdr.WriteByte('\n')
		}
		bdr.WriteString(")\n")
	}

	// 代码行
	bdr.WriteByte('\n')
	for _, line := range g.codeLines {
		bdr.WriteString(line)
		bdr.WriteByte('\n')
	}

	return bdr.String()
}

func (g *generator) findFinalElementForPointerKind(
	ptr reflect.Type,
) (elem reflect.Type, level int, isBasicType, copyable bool) {
	elem = ptr
	for {
		elem = elem.Elem()
		level++

		switch elem.Kind() {
		default:
			// basic type
			copyable = true
			return

		case reflect.Invalid, reflect.Chan, reflect.UnsafePointer, reflect.Uintptr:
			return

		case reflect.Pointer:
			// continue
		}
	}
}

func subTypeCopyFunctionName(root, sub reflect.Type) string {
	rootTypeName := strcase.ToLowerCamel(root.Name())
	subTypeName := sub.Name()

	subTypePath := sub.PkgPath()
	subTypePath = strings.ReplaceAll(subTypePath, "/", "_")
	subTypePath = strings.ReplaceAll(subTypePath, ".", "_")
	subTypePath = strcase.ToCamel(subTypePath)

	return fmt.Sprintf("internalDeepCopy_%s_%s_%s", rootTypeName, subTypePath, subTypeName)
}
