package json

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	jsonvalue "github.com/Andrew-M-C/go.jsonvalue"
	"golang.org/x/exp/slices"
)

// ProtocolAdapter JSON 协议转换器
type ProtocolAdapter interface {
	ConvertJSON(from, to any) error
}

// FieldType 表示字段类型
type FieldType string

const (
	Integer FieldType = "integer"
	Int     FieldType = "int"
	String  FieldType = "string"
	Boolean FieldType = "boolean"
	Bool    FieldType = "bool"
	Float   FieldType = "float"
)

// ProtocolAdapterMapping 表示 ProtocolAdapter 的一个字段映射配置
type ProtocolAdapterMapping struct {
	From     string    `json:"from"`
	To       string    `json:"to"`
	Required bool      `json:"required,omitempty"`
	Type     FieldType `json:"type,omitempty"`
}

type adapter struct {
	fields []*fieldAdapter
}

// NewProtocolAdapterByFieldsConfig 使用字段映射配置初始化一个解析器
func NewProtocolAdapterByFieldsConfig(mappings []ProtocolAdapterMapping) (ProtocolAdapter, error) {
	a := &adapter{}

	for i, m := range mappings {
		field, err := parseFieldAdapter(m)
		if err != nil {
			return nil, fmt.Errorf("parse field at Index %d error: %w", i, err)
		}
		a.fields = append(a.fields, field)
	}

	return a, nil
}

func (a *adapter) ConvertJSON(from, to any) error {
	fromV, err := jsonvalue.Import(from)
	if err != nil {
		return fmt.Errorf("reading input error: %w", err)
	}

	var toV *jsonvalue.V
	for i, f := range a.fields {
		toV, err = f.convert(fromV, toV)
		if err != nil {
			return fmt.Errorf("parsing error at field with Index %d, error: %w", i, err)
		}
	}

	if err := toV.Export(to); err != nil {
		return fmt.Errorf("export to 'to' error: %w", err)
	}
	return nil
}

type fieldAdapter struct {
	from struct {
		path         []any
		required     bool
		iterateIndex int
	}
	to struct {
		path         []any
		iterateIndex int
		typ          FieldType
	}
	raw ProtocolAdapterMapping
}

func parseFieldAdapter(conf ProtocolAdapterMapping) (*fieldAdapter, error) {
	f := &fieldAdapter{}

	path, iterateIndex, err := parseFieldAdapterField(conf.From)
	if err != nil {
		return nil, err
	}
	f.from.path = path
	f.from.iterateIndex = iterateIndex
	f.from.required = conf.Required

	path, iterateIndex, err = parseFieldAdapterField(conf.To)
	if err != nil {
		return nil, err
	}
	f.to.path = path
	f.to.iterateIndex = iterateIndex
	f.to.typ = conf.Type
	f.raw = conf

	return f, nil
}

func parseFieldAdapterField(s string) (path []any, iterateIndex int, err error) {
	iterateIndex = -1

	for i, part := range strings.Split(s, ".") {
		if part == "[n]" {
			iterateIndex = i
			continue
		}
		if strings.HasPrefix(part, "[") && strings.HasSuffix(part, "]") {
			n, e := strconv.Atoi(part[1 : len(part)-1])
			if e != nil {
				err = fmt.Errorf("illegal field part '%s'", part)
				return
			}
			path = append(path, n)
			continue
		}
		path = append(path, part)
	}

	return
}

func (f *fieldAdapter) convert(from *jsonvalue.V, to *jsonvalue.V) (*jsonvalue.V, error) {
	if f.from.iterateIndex < 0 {
		return f.singleConvert(from, to)
	}
	return f.arrayIterationConvert(from, to)
}

func (f *fieldAdapter) singleConvert(from *jsonvalue.V, to *jsonvalue.V) (*jsonvalue.V, error) {
	field, err := from.Get(f.from.path[0], f.from.path[1:]...)
	if err != nil {
		if errors.Is(err, jsonvalue.ErrNotFound) {
			if f.from.required {
				return to, fmt.Errorf("field %s not found", f.raw.From)
			}
			return to, nil
		}
	}

	out := getValueByType(field, f.raw.Type)
	if to == nil {
		if _, ok := f.to.path[0].(int); ok {
			to = jsonvalue.NewArray()
		} else {
			to = jsonvalue.NewObject()
		}
	}

	_, err = to.Set(out).At(f.to.path[0], f.to.path[1:]...)
	if err != nil {
		return to, fmt.Errorf("set field %s error: %w", f.raw.To, err)
	}
	return to, nil
}

func (f *fieldAdapter) arrayIterationConvert(from *jsonvalue.V, to *jsonvalue.V) (*jsonvalue.V, error) {
	for i := 0; true; i++ {
		// 是不是已经拿不到对应的 array 了?
		first, remaining := generatePath(i, f.from.path[:f.from.iterateIndex], f.from.iterateIndex)
		if _, err := from.Get(first, remaining...); err != nil {
			break
		}

		first, remaining = generatePath(i, f.from.path, f.from.iterateIndex)
		field, err := from.Get(first, remaining...)
		if err != nil {
			// 为了保持 array 数目, 用 null 占一个坑
			field = jsonvalue.NewNull()
		}

		out := getValueByType(field, f.raw.Type)
		first, remaining = generatePath(i, f.to.path, f.to.iterateIndex)
		if to == nil {
			if _, ok := first.(int); ok {
				to = jsonvalue.NewArray()
			} else {
				to = jsonvalue.NewObject()
			}
		}

		_, err = to.Set(out).At(first, remaining...)
		if err != nil {
			return to, fmt.Errorf("set field %s error: %w", f.raw.To, err)
		}
	}

	return to, nil
}

func generatePath(i int, path []any, iterateIndex int) (first any, remaining []any) {
	if iterateIndex == 0 {
		return i, path
	}

	remaining = slices.Clone(path[1:iterateIndex])
	remaining = append(remaining, i)
	remaining = append(remaining, path[iterateIndex:]...)
	return path[0], remaining
}

func getValueByType(v *jsonvalue.V, typ FieldType) any {
	switch FieldType(strings.ToLower(string(typ))) {
	default:
		return v
	case String:
		return v.String()
	case Integer, Int:
		return v.Int64()
	case Float:
		return v.Float64()
	case Boolean, Bool:
		return v.Bool()
	}
}
