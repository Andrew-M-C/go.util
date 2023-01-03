package deepcopy

import (
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/Andrew-M-C/go.util/maps"
)

func (d *DeepCopyBuilder) do() (string, error) {
	procedure := []func() error{
		d.analyzePrototypes,
		d.iterateEachTypes,
		d.iterateUnExportableTypes,
		d.packCodes,
	}

	for _, p := range procedure {
		if err := p(); err != nil {
			return "", err
		}
	}

	return d.result.String(), nil
}

// 分析需要处理的原型
func (d *DeepCopyBuilder) analyzePrototypes() error {
	if len(d.prototypes) == 0 {
		return errors.New("no prototypes given")
	}

	for _, prototype := range d.prototypes {
		detail := d.analyzeType(reflect.TypeOf(prototype), true)
		if detail.Kind == kindIllegal {
			return fmt.Errorf("unsupported type: %v", detail.Type)
		}

		d.typesToParseAndExport = append(d.typesToParseAndExport, &detail)
		d.logf(
			"待处理类型 %v, 分类: %v, 类型引用名称: %v, 包路径: %v",
			detail.Type, detail.Kind, detail.TypeReferenceName(), detail.PackagePath(),
		)
	}

	return nil
}

// 迭代处理每一个类型
func (d *DeepCopyBuilder) iterateEachTypes() error {
	for _, detail := range d.typesToParseAndExport {
		var err error

		switch detail.Kind {
		default:
			continue

		case kindBasic:
			err = fmt.Errorf("%v is a basic type which does not need copying", detail.TypeReferenceName())

		case kindStruct:
			err = d.handleStructKind(detail)
			// TODO:

		case kindSlice:
			// TODO:

		case kindMap:
			// TODO:

		case kindArray:
			// TODO:
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func (d *DeepCopyBuilder) handleStructKind(detail *typeDetail) error {
	if detail.PackagePath() != "" {
		n := d.addImportLine(detail.PackageName(), detail.PackagePath())
		if n != detail.PackageName() {
			detail.SetPackageName(n)
		}
	}

	funcName := detail.CopyFuncName(true)
	d.addCodeLine("")
	d.addCodeLine(`// %s %s`, funcName, detail.FunctionComment())
	d.addCodeLine(
		`func %s(%s %s) %s {`,
		funcName, detail.SelfName(), detail.TypeReferenceName(), detail.TypeReferenceName(),
	)

	if detail.IsPointer {
		d.addCodeLine(`	cpy := %s.%s{}`, detail.PackageName(), detail.TypeName())
		d.addCodeLine(`	cpy = *%s`, detail.SelfName())
	} else {
		d.addCodeLine(`	cpy = %s`, detail.SelfName())
	}

	fields := d.readStructFields(detail)
	for _, f := range fields {
		switch f.Kind {
		default:
			d.logf("跳过不支持复制的字段 '%s' (%v)", f.Field.Name, f.Kind)
			continue

		case kindBasic:
			d.debugf("基本类型无需显式复制 (%s)", f.Field.Name)
			continue
			// d.addCodeLine(`	cpy.%s = %s.%s`, f.Field.Name, detail.SelfName(), f.Field.Name)

		case kindStruct:
			d.typesToParseInternally[f.Type] = &f.typeDetail
			subFuncName := f.CopyFuncName(false)
			d.addCodeLine(
				`	cpy.%s = %s(%s.%s)`,
				f.Field.Name, subFuncName, detail.SelfName(), f.Field.Name,
			)

		case kindSlice:
			elemDetail := d.analyzeType(f.Elem, true)
			switch elemDetail.Kind {
			default:
				d.logf("跳过不支持复制的字段 '%s' (%v)", f.Field.Name, elemDetail.Kind)
				continue

			case kindBasic:
				d.addCodeLine(
					`	cpy.%s = append(nil, %s.%s...)`,
					f.Field.Name, detail.SelfName(), f.Field.Name,
				)

			case kindStruct, kindSlice, kindArray, kindMap:
				d.typesToParseInternally[elemDetail.Type] = &elemDetail
				d.addCodeLine(
					`	cpy.%s = make([]%s, len(%s.%s))`,
					f.Field.Name, f.TypeReferenceName(), detail.SelfName(), f.Field.Name,
				)
				d.addCodeLine(
					`	for idx, %s := range %s.%s {`,
					elemDetail.SelfName(), detail.SelfName(), f.Field.Name,
				)
				subFuncName := elemDetail.CopyFuncName(false)
				d.addCodeLine(
					`		cpy.%s[idx] = %s(%s)`,
					f.Field.Name, subFuncName, elemDetail.SelfName(),
				)
				d.addCodeLine(`	}`)
			}

		case kindArray:
			elemDetail := d.analyzeType(f.Elem, true)
			switch elemDetail.Kind {
			default:
				d.logf("跳过不支持复制的字段 '%s' (%v)", f.Field.Name, elemDetail.Kind)
				continue

			case kindBasic:
				d.addCodeLine(
					`	cpy.%s = %s{}`, f.Field.Name, f.TypeReferenceName(),
				)
				d.addCodeLine(
					`	for idx, %s := range %s.%s {`,
					elemDetail.SelfName(), detail.SelfName(), f.Field.Name,
				)
				d.addCodeLine(
					`		cpy.%s[idx] = %s`,
					f.Field.Name, elemDetail.SelfName(),
				)
				d.addCodeLine(`	}`)

			case kindStruct, kindSlice, kindArray, kindMap:
				d.typesToParseInternally[elemDetail.Type] = &elemDetail
				d.addCodeLine(
					`	cpy.%s = %s{}`, f.Field.Name, f.TypeReferenceName(),
				)
				d.addCodeLine(
					`	for idx, %s := range %s.%s {`,
					elemDetail.SelfName(), detail.SelfName(), f.Field.Name,
				)
				subFuncName := elemDetail.CopyFuncName(false)
				d.addCodeLine(
					`		cpy.%s[idx] = %s(%s)`,
					f.Field.Name, subFuncName, elemDetail.SelfName(),
				)
				d.addCodeLine(`	}`)
			}

		case kindMap:
			// TODO:
		}
	}

	if detail.IsPointer {
		d.addCodeLine(`	return &cpy`)
	} else {
		d.addCodeLine(`	return cpy`)
	}

	d.addCodeLine("}")
	return nil
}

// 处理不导出的内部类型
func (d *DeepCopyBuilder) iterateUnExportableTypes() error {
	// TODO:
	return nil
}

// packCodes 打包返回值
func (d *DeepCopyBuilder) packCodes() error {
	// package 行
	d.result.WriteString(fmt.Sprintf("package %s\n\n", d.filePackage))

	// import 部分
	isGoPkg := func(p string) bool {
		parts := strings.Split(p, "/")
		return len(parts) <= 2
	}
	keys := maps.Keys(d.importLines)
	sort.Slice(keys, func(i, j int) bool {
		vi := d.importLines[keys[i]]
		vj := d.importLines[keys[j]]
		if isGoPkg(vi) {
			if !isGoPkg(vj) {
				return true
			}
		} else if isGoPkg(vj) {
			return false
		}
		return vi < vj
	})
	if len(keys) > 0 {
		prev := ""
		d.result.WriteString("import (\n")
		for _, k := range keys {
			path := d.importLines[k]
			if prev != "" && isGoPkg(prev) && !isGoPkg(path) {
				d.result.WriteByte('\n')
			}
			prev = path

			pathParts := strings.Split(path, "/")
			var line string
			if k == pathParts[len(pathParts)-1] {
				line = fmt.Sprintf(`	"%s"`, path)
			} else {
				line = fmt.Sprintf(`	%s "%s"`, k, path)
			}
			d.result.WriteString(line)
			d.result.WriteByte('\n')
		}
		d.result.WriteString(")\n\n")
	}

	// 代码部分
	for _, line := range d.codeLines {
		d.result.WriteString(line)
		d.result.WriteByte('\n')
	}

	return nil
}
