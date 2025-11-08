package reflect

import (
	"reflect"
	"unsafe"
)

// FieldDesc 字段描述
type FieldDesc struct {
	// TypeDesc 字段类型描述
	TypeDesc TypeDesc

	// StructFieldName struct 类型的字段名称
	StructFieldName string
	// StructTag 字段标签
	StructTag reflect.StructTag
	// AnonymousStructField 是否为 struct 类型的匿名字段
	AnonymousStructField bool

	// MapKey map 类型的键值
	MapKey any
	// MapValue map 类型的值
	MapValue any
}

// ValueReader 字段值读取器。其中第一个的 i 参数只在 struct、slice 和 array 有效,
// 其他情况都为 -1
type ValueReader interface {
	ReadNil(i int, f FieldDesc)
	ReadBool(i int, f FieldDesc, v bool)
	ReadInt64(i int, f FieldDesc, v int64)
	ReadUint64(i int, f FieldDesc, v uint64)
	ReadUintptr(i int, f FieldDesc, v uintptr)
	ReadFloat32(i int, f FieldDesc, v float32)
	ReadFloat64(i int, f FieldDesc, v float64)
	ReadComplex128(i int, f FieldDesc, v complex128)
	ReadArray(i int, f FieldDesc, v any)
	ReadChan(i int, f FieldDesc, v any)
	ReadFunc(i int, f FieldDesc, v any)
	ReadInterface(i int, f FieldDesc, v any)
	ReadMap(i int, f FieldDesc, v any)
	ReadSlice(i int, f FieldDesc, v any)
	ReadString(i int, f FieldDesc, v string)
	ReadStruct(i int, f FieldDesc, v any)
	ReadUnsafePointer(i int, f FieldDesc, v unsafe.Pointer)
}

// NewSimpleValueReader new(SimpleValueReader)
func NewSimpleValueReader() *SimpleValueReader {
	return &SimpleValueReader{}
}

// SimpleValueReader 最简单化的字段值读取器, 按顺序简单地罗列获得的值
type SimpleValueReader struct {
	Results []SimpleValueReadResultItem
}

func (r *SimpleValueReader) Clear() {
	r.Results = nil
}

func (r *SimpleValueReader) ReadNil(i int, f FieldDesc) {
	item := SimpleValueReadResultItem{
		Index: i,
		Desc:  f,
	}
	item.Read.Nil = true
	r.Results = append(r.Results, item)
}

func (r *SimpleValueReader) ReadBool(i int, f FieldDesc, v bool) {
	item := SimpleValueReadResultItem{
		Index: i,
		Desc:  f,
	}
	item.Read.Bool = true
	item.Value.Bool = v
	r.Results = append(r.Results, item)
}

func (r *SimpleValueReader) ReadInt64(i int, f FieldDesc, v int64) {
	item := SimpleValueReadResultItem{
		Index: i,
		Desc:  f,
	}
	item.Read.Int64 = true
	item.Value.Int64 = v
	r.Results = append(r.Results, item)
}

func (r *SimpleValueReader) ReadUint64(i int, f FieldDesc, v uint64) {
	item := SimpleValueReadResultItem{
		Index: i,
		Desc:  f,
	}
	item.Read.Uint64 = true
	item.Value.Uint64 = v
	r.Results = append(r.Results, item)
}

func (r *SimpleValueReader) ReadUintptr(i int, f FieldDesc, v uintptr) {
	item := SimpleValueReadResultItem{
		Index: i,
		Desc:  f,
	}
	item.Read.Uintptr = true
	item.Value.Uintptr = v
	r.Results = append(r.Results, item)
}

func (r *SimpleValueReader) ReadFloat32(i int, f FieldDesc, v float32) {
	item := SimpleValueReadResultItem{
		Index: i,
		Desc:  f,
	}
	item.Read.Float32 = true
	item.Value.Float32 = v
	r.Results = append(r.Results, item)
}

func (r *SimpleValueReader) ReadFloat64(i int, f FieldDesc, v float64) {
	item := SimpleValueReadResultItem{
		Index: i,
		Desc:  f,
	}
	item.Read.Float64 = true
	item.Value.Float64 = v
	r.Results = append(r.Results, item)
}

func (r *SimpleValueReader) ReadComplex128(i int, f FieldDesc, v complex128) {
	item := SimpleValueReadResultItem{
		Index: i,
		Desc:  f,
	}
	item.Read.Complex128 = true
	item.Value.Complex128 = v
	r.Results = append(r.Results, item)
}

func (r *SimpleValueReader) ReadArray(i int, f FieldDesc, v any) {
	item := SimpleValueReadResultItem{
		Index: i,
		Desc:  f,
	}
	item.Read.Array = true
	item.Value.Array = v
	r.Results = append(r.Results, item)
}

func (r *SimpleValueReader) ReadChan(i int, f FieldDesc, v any) {
	item := SimpleValueReadResultItem{
		Index: i,
		Desc:  f,
	}
	item.Read.Chan = true
	item.Value.Chan = v
	r.Results = append(r.Results, item)
}

func (r *SimpleValueReader) ReadFunc(i int, f FieldDesc, v any) {
	item := SimpleValueReadResultItem{
		Index: i,
		Desc:  f,
	}
	item.Read.Func = true
	item.Value.Func = v
	r.Results = append(r.Results, item)
}

func (r *SimpleValueReader) ReadInterface(i int, f FieldDesc, v any) {
	item := SimpleValueReadResultItem{
		Index: i,
		Desc:  f,
	}
	item.Read.Interface = true
	item.Value.Interface = v
	r.Results = append(r.Results, item)
}

func (r *SimpleValueReader) ReadMap(i int, f FieldDesc, v any) {
	item := SimpleValueReadResultItem{
		Index: i,
		Desc:  f,
	}
	item.Read.Map = true
	item.Value.Map = v
	r.Results = append(r.Results, item)
}

func (r *SimpleValueReader) ReadSlice(i int, f FieldDesc, v any) {
	item := SimpleValueReadResultItem{
		Index: i,
		Desc:  f,
	}
	item.Read.Slice = true
	item.Value.Slice = v
	r.Results = append(r.Results, item)
}

func (r *SimpleValueReader) ReadString(i int, f FieldDesc, v string) {
	item := SimpleValueReadResultItem{
		Index: i,
		Desc:  f,
	}
	item.Read.String = true
	item.Value.String = v
	r.Results = append(r.Results, item)
}

func (r *SimpleValueReader) ReadStruct(i int, f FieldDesc, v any) {
	item := SimpleValueReadResultItem{
		Index: i,
		Desc:  f,
	}
	item.Read.Struct = true
	item.Value.Struct = v
	r.Results = append(r.Results, item)
}

func (r *SimpleValueReader) ReadUnsafePointer(i int, f FieldDesc, v unsafe.Pointer) {
	item := SimpleValueReadResultItem{
		Index: i,
		Desc:  f,
	}
	item.Read.UnsafePointer = true
	item.Value.UnsafePointer = v
	r.Results = append(r.Results, item)
}

type SimpleValueReadResultItem struct {
	Index int
	Desc  FieldDesc
	Read  struct {
		Nil           bool
		Bool          bool
		Int64         bool
		Uint64        bool
		Uintptr       bool
		Float32       bool
		Float64       bool
		Complex128    bool
		Array         bool
		Chan          bool
		Func          bool
		Interface     bool
		Map           bool
		Slice         bool
		String        bool
		Struct        bool
		UnsafePointer bool
	}
	Value struct {
		Bool          bool
		Int64         int64
		Uint64        uint64
		Uintptr       uintptr
		Float32       float32
		Float64       float64
		Complex128    complex128
		Array         any
		Chan          any
		Func          any
		Interface     any
		Map           any
		Slice         any
		String        string
		Struct        any
		UnsafePointer unsafe.Pointer
	}
}
