package reflect

import (
	"errors"
	"fmt"
	"reflect"

	jsonvalue "github.com/Andrew-M-C/go.jsonvalue"
)

// FromJsonvalueToAny 将 *jsonvalue.V 类型转换为 any 类型的值, 这是 ReadAnyToJsonvalue
// 的反操作
//
// WARN: beta feature
func FromJsonvalueToAny(target any, jv *jsonvalue.V, tag string) error {
	// 首先检查 target, 必须是一个 pointer, 否则无法赋值
	if target == nil {
		return errors.New("target is nil")
	}

	val := reflect.ValueOf(target)
	if val.Kind() != reflect.Pointer {
		return fmt.Errorf("target must be a pointer, but got %v", val.Kind())
	}

	if val.IsNil() {
		return errors.New("target pointer is nil")
	}

	if jv == nil {
		return errors.New("jsonvalue is nil")
	}

	// 然后执行 ReadAnyToJsonvalue 的反操作, 使用 reflect 生成 target
	return setValueFromJsonvalue(val.Elem(), jv, tag)
}

// setValueFromJsonvalue 将 jsonvalue 的值设置到 reflect.Value 中
func setValueFromJsonvalue(val reflect.Value, jv *jsonvalue.V, tag string) error {
	// 处理 nil 值
	if jv.IsNull() {
		// 如果目标是指针、接口、切片、map 或 channel，则设置为 nil
		switch val.Kind() {
		case reflect.Pointer, reflect.Interface, reflect.Slice, reflect.Map, reflect.Chan:
			val.Set(reflect.Zero(val.Type()))
		}
		return nil
	}

	// 根据目标类型和 jsonvalue 类型进行转换
	switch val.Kind() {
	case reflect.Bool:
		if !jv.IsBoolean() {
			return fmt.Errorf("cannot convert %v to bool", jv.ValueType())
		}
		val.SetBool(jv.Bool())

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if !jv.IsNumber() {
			return fmt.Errorf("cannot convert %v to int", jv.ValueType())
		}
		val.SetInt(jv.Int64())

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		if !jv.IsNumber() {
			return fmt.Errorf("cannot convert %v to uint", jv.ValueType())
		}
		val.SetUint(jv.Uint64())

	case reflect.Float32, reflect.Float64:
		if !jv.IsNumber() {
			return fmt.Errorf("cannot convert %v to float", jv.ValueType())
		}
		val.SetFloat(jv.Float64())

	case reflect.String:
		if !jv.IsString() {
			return fmt.Errorf("cannot convert %v to string", jv.ValueType())
		}
		val.SetString(jv.String())

	case reflect.Slice:
		return setSliceFromJsonvalue(val, jv, tag)

	case reflect.Array:
		return setArrayFromJsonvalue(val, jv, tag)

	case reflect.Map:
		return setMapFromJsonvalue(val, jv, tag)

	case reflect.Struct:
		return setStructFromJsonvalue(val, jv, tag)

	case reflect.Pointer:
		return setPointerFromJsonvalue(val, jv, tag)

	case reflect.Interface:
		return setInterfaceFromJsonvalue(val, jv, tag)

	default:
		return fmt.Errorf("unsupported target type: %v", val.Kind())
	}

	return nil
}

// setStructFromJsonvalue 将 jsonvalue 对象设置到 struct 中
func setStructFromJsonvalue(val reflect.Value, jv *jsonvalue.V, tag string) error {
	if !jv.IsObject() {
		return fmt.Errorf("cannot convert %v to struct", jv.ValueType())
	}

	typ := val.Type()
	numField := typ.NumField()

	for i := 0; i < numField; i++ {
		field := typ.Field(i)
		fieldVal := val.Field(i)

		// 跳过不可设置的字段
		if !fieldVal.CanSet() {
			continue
		}

		// 处理匿名嵌入字段（展平的 JSON 结构）
		if field.Anonymous {
			if err := setAnonymousFieldFromJsonvalue(fieldVal, jv, tag); err != nil {
				return fmt.Errorf("failed to set anonymous field %s: %w", field.Name, err)
			}
			continue
		}

		// 获取字段对应的 JSON key
		key := field.Name
		if tag != "" {
			if tagValue := field.Tag.Get(tag); tagValue != "" {
				key = tagValue
			}
		}

		// 从 jsonvalue 中获取对应的值
		childJv, err := jv.Get(key)
		if err != nil {
			// 字段不存在，跳过
			continue
		}

		// 递归设置字段值
		if err := setValueFromJsonvalue(fieldVal, childJv, tag); err != nil {
			return fmt.Errorf("failed to set field %s: %w", field.Name, err)
		}
	}

	return nil
}

// setAnonymousFieldFromJsonvalue 处理匿名嵌入字段
// 匿名嵌入字段在 JSON 中可能是：
// 1. 展平的 - 嵌入结构体的字段直接出现在外层 JSON 对象中（标准 encoding/json 行为）
// 2. 嵌套的 - 使用类型名作为 key 的嵌套对象（ReadAnyToJsonvalue 当前的行为）
func setAnonymousFieldFromJsonvalue(fieldVal reflect.Value, jv *jsonvalue.V, tag string) error {
	fieldType := fieldVal.Type()
	fieldName := fieldType.Name()

	// 处理指针类型的匿名嵌入
	if fieldType.Kind() == reflect.Pointer {
		elemType := fieldType.Elem()
		// 只处理指向 struct 的指针
		if elemType.Kind() != reflect.Struct {
			return nil
		}
		fieldName = elemType.Name()

		// 首先尝试嵌套格式（使用类型名作为 key）
		if nestedJv, err := jv.Get(fieldName); err == nil && nestedJv.IsObject() {
			newVal := reflect.New(elemType)
			if err := setStructFromJsonvalue(newVal.Elem(), nestedJv, tag); err != nil {
				return err
			}
			fieldVal.Set(newVal)
			return nil
		}

		// 然后尝试展平格式
		hasAnyField := checkStructHasAnyFieldInJsonvalue(elemType, jv, tag)
		if !hasAnyField {
			// 没有任何字段，保持 nil
			return nil
		}

		// 创建新的结构体实例
		newVal := reflect.New(elemType)
		if err := setStructFromJsonvalue(newVal.Elem(), jv, tag); err != nil {
			return err
		}
		fieldVal.Set(newVal)
		return nil
	}

	// 处理非指针类型的匿名嵌入（必须是 struct）
	if fieldType.Kind() != reflect.Struct {
		return nil
	}

	// 首先尝试嵌套格式（使用类型名作为 key）
	if nestedJv, err := jv.Get(fieldName); err == nil && nestedJv.IsObject() {
		return setStructFromJsonvalue(fieldVal, nestedJv, tag)
	}

	// 然后尝试展平格式 - 直接使用当前的 jv 对象递归设置嵌入结构体的字段
	return setStructFromJsonvalue(fieldVal, jv, tag)
}

// checkStructHasAnyFieldInJsonvalue 检查 JSON 对象中是否包含结构体的任何字段
func checkStructHasAnyFieldInJsonvalue(structType reflect.Type, jv *jsonvalue.V, tag string) bool {
	numField := structType.NumField()
	for i := 0; i < numField; i++ {
		field := structType.Field(i)

		// 处理嵌套的匿名字段
		if field.Anonymous {
			fieldType := field.Type
			if fieldType.Kind() == reflect.Pointer {
				fieldType = fieldType.Elem()
			}
			if fieldType.Kind() == reflect.Struct {
				if checkStructHasAnyFieldInJsonvalue(fieldType, jv, tag) {
					return true
				}
			}
			continue
		}

		// 获取字段对应的 JSON key
		key := field.Name
		if tag != "" {
			if tagValue := field.Tag.Get(tag); tagValue != "" {
				key = tagValue
			}
		}

		// 检查该 key 是否存在于 JSON 中
		if _, err := jv.Get(key); err == nil {
			return true
		}
	}
	return false
}

// setSliceFromJsonvalue 将 jsonvalue 数组设置到 slice 中
func setSliceFromJsonvalue(val reflect.Value, jv *jsonvalue.V, tag string) error {
	if !jv.IsArray() {
		return fmt.Errorf("cannot convert %v to slice", jv.ValueType())
	}

	length := jv.Len()
	slice := reflect.MakeSlice(val.Type(), length, length)

	idx := 0
	jv.RangeArray(func(i int, childJv *jsonvalue.V) bool {
		elemVal := slice.Index(idx)
		if err := setValueFromJsonvalue(elemVal, childJv, tag); err != nil {
			// 出错时停止遍历
			return false
		}
		idx++
		return true
	})

	val.Set(slice)
	return nil
}

// setArrayFromJsonvalue 将 jsonvalue 数组设置到 array 中
func setArrayFromJsonvalue(val reflect.Value, jv *jsonvalue.V, tag string) error {
	if !jv.IsArray() {
		return fmt.Errorf("cannot convert %v to array", jv.ValueType())
	}

	length := val.Len()
	jvLength := jv.Len()

	if jvLength != length {
		return fmt.Errorf("array length mismatch: target=%d, jsonvalue=%d", length, jvLength)
	}

	idx := 0
	jv.RangeArray(func(i int, childJv *jsonvalue.V) bool {
		elemVal := val.Index(idx)
		if err := setValueFromJsonvalue(elemVal, childJv, tag); err != nil {
			// 出错时停止遍历
			return false
		}
		idx++
		return true
	})

	return nil
}

// setMapFromJsonvalue 将 jsonvalue 对象设置到 map 中
func setMapFromJsonvalue(val reflect.Value, jv *jsonvalue.V, tag string) error {
	if !jv.IsObject() {
		return fmt.Errorf("cannot convert %v to map", jv.ValueType())
	}

	typ := val.Type()
	mapVal := reflect.MakeMap(typ)

	keyType := typ.Key()
	valueType := typ.Elem()

	jv.RangeObjects(func(k string, childJv *jsonvalue.V) bool {
		// 将字符串 key 转换为目标 key 类型
		keyVal := reflect.New(keyType).Elem()
		if keyType.Kind() == reflect.String {
			keyVal.SetString(k)
		} else {
			// 如果 key 不是 string 类型，尝试其他转换
			// 这里简化处理，只支持 string key
			return true
		}

		// 创建 value 并设置
		valueVal := reflect.New(valueType).Elem()
		if err := setValueFromJsonvalue(valueVal, childJv, tag); err != nil {
			// 出错时停止遍历
			return false
		}

		mapVal.SetMapIndex(keyVal, valueVal)
		return true
	})

	val.Set(mapVal)
	return nil
}

// setPointerFromJsonvalue 将 jsonvalue 设置到 pointer 中
func setPointerFromJsonvalue(val reflect.Value, jv *jsonvalue.V, tag string) error {
	if jv.IsNull() {
		val.Set(reflect.Zero(val.Type()))
		return nil
	}

	// 创建新的指针指向的值
	elemType := val.Type().Elem()
	elemVal := reflect.New(elemType).Elem()

	if err := setValueFromJsonvalue(elemVal, jv, tag); err != nil {
		return err
	}

	// 创建指针并设置
	ptrVal := reflect.New(elemType)
	ptrVal.Elem().Set(elemVal)
	val.Set(ptrVal)

	return nil
}

// setInterfaceFromJsonvalue 将 jsonvalue 设置到 interface{} 中
func setInterfaceFromJsonvalue(val reflect.Value, jv *jsonvalue.V, tag string) error {
	// 根据 jsonvalue 的类型，创建对应的 Go 值
	var result any

	switch {
	case jv.IsNull():
		result = nil
	case jv.IsBoolean():
		result = jv.Bool()
	case jv.IsInteger():
		result = jv.Int64()
	case jv.IsFloat():
		result = jv.Float64()
	case jv.IsString():
		result = jv.String()
	case jv.IsArray():
		// 创建 []interface{} 类型
		length := jv.Len()
		arr := make([]any, length)
		idx := 0
		jv.RangeArray(func(i int, childJv *jsonvalue.V) bool {
			var elem any
			elemVal := reflect.ValueOf(&elem).Elem()
			if err := setInterfaceFromJsonvalue(elemVal, childJv, tag); err != nil {
				return false
			}
			arr[idx] = elem
			idx++
			return true
		})
		result = arr
	case jv.IsObject():
		// 创建 map[string]interface{} 类型
		obj := make(map[string]any)
		jv.RangeObjects(func(k string, childJv *jsonvalue.V) bool {
			var elem any
			elemVal := reflect.ValueOf(&elem).Elem()
			if err := setInterfaceFromJsonvalue(elemVal, childJv, tag); err != nil {
				return false
			}
			obj[k] = elem
			return true
		})
		result = obj
	default:
		return fmt.Errorf("unsupported jsonvalue type: %v", jv.ValueType())
	}

	val.Set(reflect.ValueOf(result))
	return nil
}
