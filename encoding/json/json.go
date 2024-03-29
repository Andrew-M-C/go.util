package json

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Andrew-M-C/go.util/unsafe"
)

// MarshalToString 序列化为 string 类型
func MarshalToString(v any) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return unsafe.BtoS(b), err
}

// UnmarshalFromString 从 string 反序列化
func UnmarshalFromString(s string, v any) error {
	b := unsafe.StoB(s)
	return json.Unmarshal(b, v)
}

// MarshalToFile 序列化到文件中
func MarshalToFile(fileName string, v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshal JSON error (%w)", err)
	}
	if err := os.WriteFile(fileName, b, 0644); err != nil {
		return fmt.Errorf("write to file error (%w)", err)
	}
	return nil
}

// UnmarshalFromFile 从文件中反序列化出来
func UnmarshalFromFile(fileName string, v any) error {
	b, err := os.ReadFile(fileName)
	if err != nil {
		return fmt.Errorf("read from file error (%w)", err)
	}
	if err := json.Unmarshal(b, v); err != nil {
		return fmt.Errorf("unmarshal JSON error (%w)", err)
	}
	return nil
}
