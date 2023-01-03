// Package deepcopy 提供一个生成对象深复制的功能
package deepcopy

import (
	"reflect"
	"strings"
)

type DeepCopyBuilder struct {
	filePackage string
	prototypes  []any

	typesToParseAndExport  []*typeDetail
	typesToParseInternally map[reflect.Type]*typeDetail

	debug bool
	logf  func(string, ...any)

	importLines map[string]string // key: package name; value: package path
	codeLines   []string

	result strings.Builder
}

// BuildDeepCopy 构建一个 deepcopy 的构建器
func BuildDeepCopy(prototypes ...any) *DeepCopyBuilder {
	d := &DeepCopyBuilder{
		filePackage: "deepcopy",
		prototypes:  prototypes,
		logf:        emptyLog,
	}
	d.typesToParseInternally = make(map[reflect.Type]*typeDetail)
	d.importLines = make(map[string]string)
	return d
}

// Do 执行构建并返回文件
func (d *DeepCopyBuilder) Do() (string, error) {
	return d.do()
}

func (d *DeepCopyBuilder) PackageName(n string) *DeepCopyBuilder {
	if n != "" {
		d.filePackage = n
	}
	return d
}

// WithLogFunc 指定日志函数
func (d *DeepCopyBuilder) WithLogFunc(f func(string, ...any)) *DeepCopyBuilder {
	if f == nil {
		return d
	}
	d.logf = f
	return d
}

// EnableDebug 开启调试日志
func (d *DeepCopyBuilder) EnableDebug() *DeepCopyBuilder {
	d.debug = true
	return d
}

func emptyLog(string, ...any) {
	// do nothing
}
