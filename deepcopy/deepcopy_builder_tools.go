package deepcopy

import (
	"fmt"
	"reflect"
)

func (d *DeepCopyBuilder) addCodeLine(f string, a ...any) {
	s := fmt.Sprintf(f, a...)
	d.codeLines = append(d.codeLines, s)
}

func (d *DeepCopyBuilder) addImportLine(packageName, packagePath string) (packageNameUsed string) {
	if d.importLines == nil {
		d.importLines = make(map[string]string)
	}

	i := 1
	packageNameUsed = packageName
	for {
		s, exist := d.importLines[packageNameUsed]
		if !exist {
			break
		}
		if s == packagePath {
			break
		}

		i++
		packageNameUsed = fmt.Sprintf("%s%d", packageName, i)
	}

	d.importLines[packageNameUsed] = packagePath
	return packageNameUsed
}

// 解析一个原型
func (d *DeepCopyBuilder) analyzeType(
	typ reflect.Type, allowPointer bool,
) (res typeDetail) {

	res.Type = typ
	res.Kind = kindIllegal
	res.Elem = typ

	switch typ.Kind() {
	default:
		d.debugf("不支持的类型: %v (%v)", typ, typ.Kind())
		return

	case reflect.Bool:
		fallthrough
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		fallthrough
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		fallthrough
	case reflect.Float32, reflect.Float64:
		fallthrough
	case reflect.Complex128, reflect.Complex64:
		fallthrough
	case reflect.Func, reflect.String, reflect.Interface:
		res.Kind = kindBasic
		return

	case reflect.Struct:
		res.Kind = kindStruct
		return

	case reflect.Pointer:
		res.Elem = typ.Elem()
		elemDetail := d.analyzeType(res.Elem, false)
		if elemDetail.Kind == kindIllegal {
			res.Kind = kindIllegal
			return
		}
		res.Kind = elemDetail.Kind
		res.IsPointer = true
		return

	case reflect.Array:
		res.Kind = kindArray

	case reflect.Slice:
		res.Kind = kindSlice

	case reflect.Map:
		res.Kind = kindMap
	}

	res.Elem = typ.Elem()
	elemDetail := d.analyzeType(res.Elem, true)
	if elemDetail.Kind == kindIllegal {
		res.Kind = kindIllegal
		return
	}
	res.Elem = elemDetail.Elem
	res.IsPointer = elemDetail.IsPointer
	return
}

func (d *DeepCopyBuilder) debugf(f string, a ...any) {
	if d.debug {
		d.logf("[DEBUG] "+f, a...)
	}
}

func (d *DeepCopyBuilder) readStructFields(detail *typeDetail) (fields []*structTypeDetail) {
	num := detail.Elem.NumField()
	for i := 0; i < num; i++ {
		ft := detail.Elem.Field(i)
		if !ft.IsExported() {
			d.debugf("skip un-exported field '%s'", ft.Name)
			continue
		}

		fieldDetail := d.analyzeType(ft.Type, true)
		if ft.Anonymous {
			if ft.Type.Kind() == reflect.Struct {
				d.debugf("find anonymous field: '%s'", ft.Type)
				fieldDetail := d.analyzeType(ft.Type, false)
				fields = append(fields, d.readStructFields(&fieldDetail)...)
				continue
			}
			ft.Name = fieldDetail.TypeName()
		}

		stDetail := &structTypeDetail{
			typeDetail: fieldDetail,
			Field:      ft,
		}
		fields = append(fields, stDetail)
	}

	return
}
