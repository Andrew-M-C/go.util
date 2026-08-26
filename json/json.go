package json

import (
	"errors"
	"fmt"
	"os"

	jsonvalue "github.com/Andrew-M-C/go.jsonvalue"
)

// GetJSONFromFile 从文件中获取 JSON 数据。参数 args 可以是 string 或者 int,
// 一级级递归查找指定的值。
func GetJSONFromFile(fileName string, args ...any) (*jsonvalue.V, error) {
	// 从文件中杰仔
	b, err := os.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("read file error (%w)", err)
	}
	v, err := jsonvalue.Unmarshal(b)
	if err != nil {
		return nil, fmt.Errorf("unmarshal json error (%w)", err)
	}
	// 允许获取完整的值本身
	if len(args) == 0 {
		return v, nil
	}
	// 也允许获取子值
	res, err := v.Get(args)
	if err != nil {
		return nil, fmt.Errorf("get json value error (%w)", err)
	}
	return res, nil
}

// MustGetJSONFromFile 从文件中获取 JSON 数据, 如果失败或没有相应的值, 则返回 JSON null
// 数值。返回值保证非 nil。
func MustGetJSONFromFile(fileName string, args ...any) *jsonvalue.V {
	v, err := GetJSONFromFile(fileName, args...)
	if err != nil {
		return jsonvalue.NewNull()
	}
	return v
}

// SetJSONToFile 将 JSON 数据写入文件。但需要注意的是, 第一级 key 必须是 string, 否则会报错。
func SetJSONToFile(fileName string, v *jsonvalue.V, args ...any) error {
	if v == nil {
		return errors.New("JSON value is nil")
	}

	// 首先尝试读取原始文件
	whole := MustGetJSONFromFile(fileName)
	if whole.IsNull() {
		// 如果读取失败, 那么直接初始化一个空的 object 来承载
		whole = jsonvalue.NewObject()
	}

	// 然后将 value 设置进去
	if len(args) == 0 {
		// 如果命令就是覆盖整个文件, 那么直接写入
		whole = v
	} else {
		// 如果是带参数的, 那么 set
		whole.MustSet(v).At(args)
	}

	// 写入文件
	opts := []jsonvalue.Option{
		jsonvalue.OptSetSequence(),
		jsonvalue.OptUTF8(),
	}
	return os.WriteFile(fileName, whole.MustMarshal(opts...), 0600)
}
