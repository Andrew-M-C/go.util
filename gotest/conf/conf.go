package conf

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
)

// Option 表示配置选项，只能内部生成
type Option interface {
	mergeTo(*option)
}

// option 内部使用的 option
type option struct {
	filename string
	format   string
	create   bool
}

func (opt *option) mergeTo(tgt *option) {
	if opt.filename != "" {
		tgt.filename = opt.filename
	}
	if opt.format != "" {
		tgt.format = opt.format
	}
}

func (opt *option) validate() error {
	if opt.filename == "" {
		return fmt.Errorf("%w: missing filename", ErrParseOptions)
	}
	if opt.format == "" {
		return fmt.Errorf("%w: missing file format", ErrParseOptions)
	}
	if opt.format != "JSON" {
		return fmt.Errorf("%w: unsupported file format '%s'", ErrParseOptions, opt.format)
	}
	return nil
}

// optionGenFile 配置是否要生成一个空文件
type optionGenFile bool

func (opt optionGenFile) mergeTo(tgt *option) {
	tgt.create = bool(opt)
}

func defaultOption() *option {
	return &option{
		filename: "./.testconf.json",
		format:   "JSON",
		create:   true,
	}
}

// OptFileName 返回一个可选的配置文件名选项
func OptFileName(filename string) Option {
	return &option{
		filename: filename,
	}
}

// OptFormat 返回一个可选的文件格式。目前只支持 JSON
func OptFormat(format string) Option {
	return &option{
		format: strings.ToUpper(format),
	}
}

// OptCreateIfNotExist 当配置文件不存在时，是否需要生成默认配置
func OptCreateIfNotExist(b bool) Option {
	return optionGenFile(b)
}

// Load 加载配置
func Load(tgt interface{}, opts ...Option) error {
	// 合并和检查参数
	opt := defaultOption()
	for _, o := range opts {
		o.mergeTo(opt)
	}
	if err := opt.validate(); err != nil {
		return err
	}

	// 读取文件
	b, err := os.ReadFile(opt.filename)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("%w: %v", ErrReadFile, err)
		}

		// 文件不存在
		if !opt.create {
			return ErrFileNotExist
		}
		return createFile(tgt, opt)
	}

	if err := json.Unmarshal(b, tgt); err != nil {
		return fmt.Errorf("%w: %v", ErrUnmarshal, err)
	}

	return nil
}

func createFile(prototype interface{}, opt *option) error {
	return createJSONFile(prototype, opt)
}

func createJSONFile(prototype interface{}, opt *option) error {
	typ := reflect.TypeOf(prototype)
	if typ.Kind() != reflect.Ptr {
		return fmt.Errorf("%w, should be a pointer pointing to a struct", ErrTarget)
	}

	typ = typ.Elem()
	var v interface{}

	switch typ.Kind() {
	default:
		return fmt.Errorf("%w, should be a pointer pointing to a struct", ErrTarget)
	case reflect.Struct:
		v = reflect.New(typ).Interface()
	}

	b, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		return fmt.Errorf("marshal JSON error: %w", err)
	}

	err = os.WriteFile(opt.filename, b, 0644)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrWriteToFile, err)
	}

	return ErrFileNotExist
}
