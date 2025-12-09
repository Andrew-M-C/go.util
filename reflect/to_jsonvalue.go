package reflect

import (
	"fmt"
	"reflect"
	"unsafe"

	jsonvalue "github.com/Andrew-M-C/go.jsonvalue"
)

// ReadAnyToJsonvalue 读取一个基于 struct tag 的值到 *jsonvalue.V 类型
//
// WARN: beta feature
func ReadAnyToJsonvalue(v any, tag string) (*jsonvalue.V, error) {
	// 基于 ReadAny 能力, 读取 any 类型的数据, 然后只处理 JSON 支持的那些类型,
	// 然后基于指定的 tag, 将对应的 value 转为 *jsonvalue.V 类型的 key 导出

	if v == nil {
		return jsonvalue.NewNull(), nil
	}

	// 获取实际的值类型
	val := reflect.ValueOf(v)
	val = digPointer(val)

	switch val.Kind() {
	case reflect.Struct:
		return readStructToJsonvalue(v, tag)
	case reflect.Map:
		return readMapToJsonvalue(v, tag)
	case reflect.Slice, reflect.Array:
		return readSliceToJsonvalue(v, tag)
	default:
		// 基本类型直接转换
		return convertValueToJsonvalue(val)
	}
}

// readStructToJsonvalue 将 struct 转换为 jsonvalue 对象
func readStructToJsonvalue(v any, tag string) (*jsonvalue.V, error) {
	obj := jsonvalue.NewObject()
	reader := &jsonvalueStructReader{
		obj: obj,
		tag: tag,
	}
	ReadStruct(v, reader)
	return obj, reader.err
}

// readMapToJsonvalue 将 map 转换为 jsonvalue 对象
func readMapToJsonvalue(v any, tag string) (*jsonvalue.V, error) {
	obj := jsonvalue.NewObject()
	reader := &jsonvalueMapReader{
		obj: obj,
		tag: tag,
	}
	ReadMap(v, reader)
	return obj, reader.err
}

// readSliceToJsonvalue 将 slice/array 转换为 jsonvalue 数组
func readSliceToJsonvalue(v any, tag string) (*jsonvalue.V, error) {
	arr := jsonvalue.NewArray()
	reader := &jsonvalueArrayReader{
		arr: arr,
		tag: tag,
	}
	val := digPointer(reflect.ValueOf(v))
	if val.Kind() == reflect.Slice {
		ReadSlice(v, reader)
	} else {
		ReadArray(v, reader)
	}
	return arr, reader.err
}

// convertValueToJsonvalue 将基本类型转换为 jsonvalue
func convertValueToJsonvalue(val reflect.Value) (*jsonvalue.V, error) {
	switch val.Kind() {
	case reflect.Bool:
		return jsonvalue.NewBool(val.Bool()), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return jsonvalue.NewInt64(val.Int()), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return jsonvalue.NewUint64(val.Uint()), nil
	case reflect.Float32, reflect.Float64:
		return jsonvalue.NewFloat64(val.Float()), nil
	case reflect.String:
		return jsonvalue.NewString(val.String()), nil
	case reflect.Interface, reflect.Pointer:
		if val.IsNil() {
			return jsonvalue.NewNull(), nil
		}
		return convertValueToJsonvalue(val.Elem())
	default:
		return nil, fmt.Errorf("unsupported type for JSON: %v", val.Kind())
	}
}

// jsonvalueStructReader 用于将 struct 字段读取到 jsonvalue 对象
type jsonvalueStructReader struct {
	obj *jsonvalue.V
	tag string
	err error
}

func (r *jsonvalueStructReader) getKey(f FieldDesc) string {
	if r.tag != "" {
		if tagValue := f.StructTag.Get(r.tag); tagValue != "" {
			return tagValue
		}
	}
	return f.StructFieldName
}

func (r *jsonvalueStructReader) ReadNil(i int, f FieldDesc) {
	key := r.getKey(f)
	r.obj.MustSet(jsonvalue.NewNull()).At(key)
}

func (r *jsonvalueStructReader) ReadBool(i int, f FieldDesc, v bool) {
	key := r.getKey(f)
	r.obj.MustSet(jsonvalue.NewBool(v)).At(key)
}

func (r *jsonvalueStructReader) ReadInt64(i int, f FieldDesc, v int64) {
	key := r.getKey(f)
	r.obj.MustSet(jsonvalue.NewInt64(v)).At(key)
}

func (r *jsonvalueStructReader) ReadUint64(i int, f FieldDesc, v uint64) {
	key := r.getKey(f)
	r.obj.MustSet(jsonvalue.NewUint64(v)).At(key)
}

func (r *jsonvalueStructReader) ReadUintptr(i int, f FieldDesc, v uintptr) {
	key := r.getKey(f)
	r.obj.MustSet(jsonvalue.NewUint64(uint64(v))).At(key)
}

func (r *jsonvalueStructReader) ReadFloat32(i int, f FieldDesc, v float32) {
	key := r.getKey(f)
	r.obj.MustSet(jsonvalue.NewFloat64(float64(v))).At(key)
}

func (r *jsonvalueStructReader) ReadFloat64(i int, f FieldDesc, v float64) {
	key := r.getKey(f)
	r.obj.MustSet(jsonvalue.NewFloat64(v)).At(key)
}

func (r *jsonvalueStructReader) ReadComplex128(i int, f FieldDesc, v complex128) {
	// Complex 类型不被 JSON 支持, 转为字符串
	key := r.getKey(f)
	r.obj.MustSet(jsonvalue.NewString(fmt.Sprintf("%v", v))).At(key)
}

func (r *jsonvalueStructReader) ReadArray(i int, f FieldDesc, v any) {
	key := r.getKey(f)
	jv, err := readSliceToJsonvalue(v, r.tag)
	if err != nil && r.err == nil {
		r.err = err
		return
	}
	r.obj.MustSet(jv).At(key)
}

func (r *jsonvalueStructReader) ReadChan(i int, f FieldDesc, v any) {
	// Chan 类型不被 JSON 支持, 忽略
}

func (r *jsonvalueStructReader) ReadFunc(i int, f FieldDesc, v any) {
	// Func 类型不被 JSON 支持, 忽略
}

func (r *jsonvalueStructReader) ReadInterface(i int, f FieldDesc, v any) {
	key := r.getKey(f)
	jv, err := ReadAnyToJsonvalue(v, r.tag)
	if err != nil && r.err == nil {
		r.err = err
		return
	}
	r.obj.MustSet(jv).At(key)
}

func (r *jsonvalueStructReader) ReadMap(i int, f FieldDesc, v any) {
	key := r.getKey(f)
	jv, err := readMapToJsonvalue(v, r.tag)
	if err != nil && r.err == nil {
		r.err = err
		return
	}
	r.obj.MustSet(jv).At(key)
}

func (r *jsonvalueStructReader) ReadSlice(i int, f FieldDesc, v any) {
	key := r.getKey(f)
	jv, err := readSliceToJsonvalue(v, r.tag)
	if err != nil && r.err == nil {
		r.err = err
		return
	}
	r.obj.MustSet(jv).At(key)
}

func (r *jsonvalueStructReader) ReadString(i int, f FieldDesc, v string) {
	key := r.getKey(f)
	r.obj.MustSet(jsonvalue.NewString(v)).At(key)
}

func (r *jsonvalueStructReader) ReadStruct(i int, f FieldDesc, v any) {
	key := r.getKey(f)
	jv, err := readStructToJsonvalue(v, r.tag)
	if err != nil && r.err == nil {
		r.err = err
		return
	}
	r.obj.MustSet(jv).At(key)
}

func (r *jsonvalueStructReader) ReadUnsafePointer(i int, f FieldDesc, v unsafe.Pointer) {
	// UnsafePointer 类型不被 JSON 支持, 忽略
}

// jsonvalueArrayReader 用于将数组/切片元素读取到 jsonvalue 数组
type jsonvalueArrayReader struct {
	arr *jsonvalue.V
	tag string
	err error
}

func (r *jsonvalueArrayReader) ReadNil(i int, f FieldDesc) {
	r.arr.MustAppend(jsonvalue.NewNull()).InTheEnd()
}

func (r *jsonvalueArrayReader) ReadBool(i int, f FieldDesc, v bool) {
	r.arr.MustAppend(jsonvalue.NewBool(v)).InTheEnd()
}

func (r *jsonvalueArrayReader) ReadInt64(i int, f FieldDesc, v int64) {
	r.arr.MustAppend(jsonvalue.NewInt64(v)).InTheEnd()
}

func (r *jsonvalueArrayReader) ReadUint64(i int, f FieldDesc, v uint64) {
	r.arr.MustAppend(jsonvalue.NewUint64(v)).InTheEnd()
}

func (r *jsonvalueArrayReader) ReadUintptr(i int, f FieldDesc, v uintptr) {
	r.arr.MustAppend(jsonvalue.NewUint64(uint64(v))).InTheEnd()
}

func (r *jsonvalueArrayReader) ReadFloat32(i int, f FieldDesc, v float32) {
	r.arr.MustAppend(jsonvalue.NewFloat64(float64(v))).InTheEnd()
}

func (r *jsonvalueArrayReader) ReadFloat64(i int, f FieldDesc, v float64) {
	r.arr.MustAppend(jsonvalue.NewFloat64(v)).InTheEnd()
}

func (r *jsonvalueArrayReader) ReadComplex128(i int, f FieldDesc, v complex128) {
	r.arr.MustAppend(jsonvalue.NewString(fmt.Sprintf("%v", v))).InTheEnd()
}

func (r *jsonvalueArrayReader) ReadArray(i int, f FieldDesc, v any) {
	jv, err := readSliceToJsonvalue(v, r.tag)
	if err != nil && r.err == nil {
		r.err = err
		return
	}
	r.arr.MustAppend(jv).InTheEnd()
}

func (r *jsonvalueArrayReader) ReadChan(i int, f FieldDesc, v any) {
	// Chan 类型不被 JSON 支持, 追加 null
	r.arr.MustAppend(jsonvalue.NewNull()).InTheEnd()
}

func (r *jsonvalueArrayReader) ReadFunc(i int, f FieldDesc, v any) {
	// Func 类型不被 JSON 支持, 追加 null
	r.arr.MustAppend(jsonvalue.NewNull()).InTheEnd()
}

func (r *jsonvalueArrayReader) ReadInterface(i int, f FieldDesc, v any) {
	jv, err := ReadAnyToJsonvalue(v, r.tag)
	if err != nil && r.err == nil {
		r.err = err
		return
	}
	r.arr.MustAppend(jv).InTheEnd()
}

func (r *jsonvalueArrayReader) ReadMap(i int, f FieldDesc, v any) {
	jv, err := readMapToJsonvalue(v, r.tag)
	if err != nil && r.err == nil {
		r.err = err
		return
	}
	r.arr.MustAppend(jv).InTheEnd()
}

func (r *jsonvalueArrayReader) ReadSlice(i int, f FieldDesc, v any) {
	jv, err := readSliceToJsonvalue(v, r.tag)
	if err != nil && r.err == nil {
		r.err = err
		return
	}
	r.arr.MustAppend(jv).InTheEnd()
}

func (r *jsonvalueArrayReader) ReadString(i int, f FieldDesc, v string) {
	r.arr.MustAppend(jsonvalue.NewString(v)).InTheEnd()
}

func (r *jsonvalueArrayReader) ReadStruct(i int, f FieldDesc, v any) {
	jv, err := readStructToJsonvalue(v, r.tag)
	if err != nil && r.err == nil {
		r.err = err
		return
	}
	r.arr.MustAppend(jv).InTheEnd()
}

func (r *jsonvalueArrayReader) ReadUnsafePointer(i int, f FieldDesc, v unsafe.Pointer) {
	// UnsafePointer 类型不被 JSON 支持, 追加 null
	r.arr.MustAppend(jsonvalue.NewNull()).InTheEnd()
}

// jsonvalueMapReader 用于将 map 读取到 jsonvalue 对象
type jsonvalueMapReader struct {
	obj *jsonvalue.V
	tag string
	err error
}

func (r *jsonvalueMapReader) ReadNil(i int, f FieldDesc) {
	key := fmt.Sprintf("%v", f.MapKey)
	r.obj.MustSet(jsonvalue.NewNull()).At(key)
}

func (r *jsonvalueMapReader) ReadBool(i int, f FieldDesc, v bool) {
	key := fmt.Sprintf("%v", f.MapKey)
	r.obj.MustSet(jsonvalue.NewBool(v)).At(key)
}

func (r *jsonvalueMapReader) ReadInt64(i int, f FieldDesc, v int64) {
	key := fmt.Sprintf("%v", f.MapKey)
	r.obj.MustSet(jsonvalue.NewInt64(v)).At(key)
}

func (r *jsonvalueMapReader) ReadUint64(i int, f FieldDesc, v uint64) {
	key := fmt.Sprintf("%v", f.MapKey)
	r.obj.MustSet(jsonvalue.NewUint64(v)).At(key)
}

func (r *jsonvalueMapReader) ReadUintptr(i int, f FieldDesc, v uintptr) {
	key := fmt.Sprintf("%v", f.MapKey)
	r.obj.MustSet(jsonvalue.NewUint64(uint64(v))).At(key)
}

func (r *jsonvalueMapReader) ReadFloat32(i int, f FieldDesc, v float32) {
	key := fmt.Sprintf("%v", f.MapKey)
	r.obj.MustSet(jsonvalue.NewFloat64(float64(v))).At(key)
}

func (r *jsonvalueMapReader) ReadFloat64(i int, f FieldDesc, v float64) {
	key := fmt.Sprintf("%v", f.MapKey)
	r.obj.MustSet(jsonvalue.NewFloat64(v)).At(key)
}

func (r *jsonvalueMapReader) ReadComplex128(i int, f FieldDesc, v complex128) {
	key := fmt.Sprintf("%v", f.MapKey)
	r.obj.MustSet(jsonvalue.NewString(fmt.Sprintf("%v", v))).At(key)
}

func (r *jsonvalueMapReader) ReadArray(i int, f FieldDesc, v any) {
	key := fmt.Sprintf("%v", f.MapKey)
	jv, err := readSliceToJsonvalue(v, r.tag)
	if err != nil && r.err == nil {
		r.err = err
		return
	}
	r.obj.MustSet(jv).At(key)
}

func (r *jsonvalueMapReader) ReadChan(i int, f FieldDesc, v any) {
	// Chan 类型不被 JSON 支持, 忽略
}

func (r *jsonvalueMapReader) ReadFunc(i int, f FieldDesc, v any) {
	// Func 类型不被 JSON 支持, 忽略
}

func (r *jsonvalueMapReader) ReadInterface(i int, f FieldDesc, v any) {
	key := fmt.Sprintf("%v", f.MapKey)
	jv, err := ReadAnyToJsonvalue(v, r.tag)
	if err != nil && r.err == nil {
		r.err = err
		return
	}
	r.obj.MustSet(jv).At(key)
}

func (r *jsonvalueMapReader) ReadMap(i int, f FieldDesc, v any) {
	key := fmt.Sprintf("%v", f.MapKey)
	jv, err := readMapToJsonvalue(v, r.tag)
	if err != nil && r.err == nil {
		r.err = err
		return
	}
	r.obj.MustSet(jv).At(key)
}

func (r *jsonvalueMapReader) ReadSlice(i int, f FieldDesc, v any) {
	key := fmt.Sprintf("%v", f.MapKey)
	jv, err := readSliceToJsonvalue(v, r.tag)
	if err != nil && r.err == nil {
		r.err = err
		return
	}
	r.obj.MustSet(jv).At(key)
}

func (r *jsonvalueMapReader) ReadString(i int, f FieldDesc, v string) {
	key := fmt.Sprintf("%v", f.MapKey)
	r.obj.MustSet(jsonvalue.NewString(v)).At(key)
}

func (r *jsonvalueMapReader) ReadStruct(i int, f FieldDesc, v any) {
	key := fmt.Sprintf("%v", f.MapKey)
	jv, err := readStructToJsonvalue(v, r.tag)
	if err != nil && r.err == nil {
		r.err = err
		return
	}
	r.obj.MustSet(jv).At(key)
}

func (r *jsonvalueMapReader) ReadUnsafePointer(i int, f FieldDesc, v unsafe.Pointer) {
	// UnsafePointer 类型不被 JSON 支持, 忽略
}
