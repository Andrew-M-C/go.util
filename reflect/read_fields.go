package reflect

import (
	"reflect"
)

// ReadAny 读取任意类型的值
func ReadAny(v any, reader ValueReader) {
	doReadAny(-1, v, reader, 0)
}

func doReadAny(index int, v any, reader ValueReader, pointerLevels int) {
	desc := FieldDesc{
		TypeDesc: describeType(reflect.TypeOf(v)),
	}
	if v == nil {
		reader.ReadNil(index, desc)
		return
	}

	val := reflect.ValueOf(v)
	// 使用 describeType 识别出的指针层级，而不是传入的 pointerLevels 参数
	actualPointerLevels := desc.TypeDesc.PointerLevels
	if actualPointerLevels > 0 {
		for i := 0; i < actualPointerLevels; i++ {
			val = val.Elem()
		}
	}

	callReader(index, val.Kind(), desc, val, reader)
}

// 这个函数需要确保 kind 不为 Pointer
func callReader(index int, kind reflect.Kind, desc FieldDesc, val reflect.Value, reader ValueReader) {
	switch kind {
	case reflect.Bool:
		reader.ReadBool(index, desc, val.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		reader.ReadInt64(index, desc, val.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		reader.ReadUint64(index, desc, val.Uint())
	case reflect.Float32:
		reader.ReadFloat32(index, desc, float32(val.Float()))
	case reflect.Float64:
		reader.ReadFloat64(index, desc, val.Float())
	case reflect.Complex64, reflect.Complex128:
		reader.ReadComplex128(index, desc, val.Complex())
	case reflect.Array:
		reader.ReadArray(index, desc, val.Interface())
	case reflect.Chan:
		reader.ReadChan(index, desc, val.Interface())
	case reflect.Func:
		reader.ReadFunc(index, desc, val.Interface())
	case reflect.Interface:
		reader.ReadInterface(index, desc, val.Interface())
	case reflect.Map:
		reader.ReadMap(index, desc, val.Interface())
	case reflect.Slice:
		reader.ReadSlice(index, desc, val.Interface())
	case reflect.String:
		reader.ReadString(index, desc, val.String())
	case reflect.Struct:
		reader.ReadStruct(index, desc, val.Interface())
	case reflect.UnsafePointer:
		reader.ReadUnsafePointer(index, desc, val.UnsafePointer())
	default:
		reader.ReadNil(index, desc)
	}
}

// ReadStruct 读取结构体类型的值。注意, 仅 reflect.Struct 类型有效, 其他类型什么都不做
func ReadStruct(v any, reader ValueReader) {
	if v == nil {
		return
	}
	val := digPointer(reflect.ValueOf(v))
	if val.Kind() != reflect.Struct {
		return
	}
	numField := val.NumField()
	for i := 0; i < numField; i++ {
		fieldType := val.Type().Field(i)
		subVal := val.Field(i)

		desc := FieldDesc{
			TypeDesc:             DescribeType(subVal.Interface()),
			StructFieldName:      fieldType.Name,
			StructTag:            fieldType.Tag,
			AnonymousStructField: fieldType.Anonymous,
		}

		// 检查是否为 nil interface 或 nil pointer
		if (subVal.Kind() == reflect.Interface || subVal.Kind() == reflect.Pointer) && subVal.IsNil() {
			reader.ReadNil(i, desc)
			continue
		}

		subVal = digPointer(subVal)
		callReader(i, subVal.Kind(), desc, subVal, reader)
	}
}

func digPointer(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Pointer || v.Kind() == reflect.Interface {
		if v.IsNil() {
			break
		}
		v = v.Elem()
	}
	return v
}

// ReadSlice 读取切片类型的值。除了 reflect.Slice 类型之外其他什么类型都不做
func ReadSlice(v any, reader ValueReader) {
	if v == nil {
		return
	}
	val := digPointer(reflect.ValueOf(v))
	if val.Kind() != reflect.Slice {
		return
	}
	readSliceOrArray(val, reader)
}

// ReadArray 读取数组类型的值。除了 reflect.Array 类型之外什么事情都不做
func ReadArray(v any, reader ValueReader) {
	if v == nil {
		return
	}
	val := digPointer(reflect.ValueOf(v))
	if val.Kind() != reflect.Array {
		return
	}
	readSliceOrArray(val, reader)
}

func readSliceOrArray(val reflect.Value, reader ValueReader) {
	for i := 0; i < val.Len(); i++ {
		subVal := val.Index(i)
		desc := FieldDesc{
			TypeDesc: DescribeType(subVal.Interface()),
		}

		// 检查是否为 nil interface 或 nil pointer
		if (subVal.Kind() == reflect.Interface || subVal.Kind() == reflect.Pointer) && subVal.IsNil() {
			reader.ReadNil(i, desc)
			continue
		}

		subVal = digPointer(subVal)
		callReader(i, subVal.Kind(), desc, subVal, reader)
	}
}

// ReadMap 读取 map 类型的值。除了 reflect.Map 类型之外什么事情都不做
func ReadMap(v any, reader ValueReader) {
	if v == nil {
		return
	}
	val := digPointer(reflect.ValueOf(v))
	if val.Kind() != reflect.Map {
		return
	}
	keys := val.MapKeys()
	for _, key := range keys {
		subVal := val.MapIndex(key)
		desc := FieldDesc{
			TypeDesc: DescribeType(subVal.Interface()),
			MapKey:   key.Interface(),
			MapValue: subVal.Interface(),
		}

		// 检查是否为 nil interface 或 nil pointer
		if (subVal.Kind() == reflect.Interface || subVal.Kind() == reflect.Pointer) && subVal.IsNil() {
			reader.ReadNil(-1, desc)
			continue
		}

		subVal = digPointer(subVal)
		callReader(-1, subVal.Kind(), desc, subVal, reader)
	}
}
