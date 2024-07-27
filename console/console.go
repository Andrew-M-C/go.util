// Package console 提供了一些命令行工具的简单封装
//
// Deprecated: 请使用 log 包替代 (github.com/Andrew-M-C/go.util/log)
//
// Reference:
//   - [How can I print to Stderr in Go without using log](https://stackoverflow.com/questions/29721449/how-can-i-print-to-stderr-in-go-without-using-log)
package console

import (
	"fmt"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
)

var (
	debugOn bool
)

// SetDebug 开启/关闭调试
func SetDebug(on bool) {
	debugOn = on
}

// Debugf 在 stdout 打印调试信息
func Debugf(f string, a ...interface{}) (int, error) {
	if !debugOn {
		return 0, nil
	}
	return Printf(f, a...)
}

// Printf 打普通信息, 自动在末尾换行。
func Printf(f string, a ...interface{}) (int, error) {
	s := packLine(f, a...)
	if s == "" {
		return 0, nil
	}

	return os.Stdout.WriteString(s)
}

// Errorf 打错误信息,
func Errorf(f string, a ...interface{}) (int, error) {
	s := packLine(f, a...)
	if s == "" {
		return 0, nil
	}

	s = color.RedString("%s", s)
	return os.Stderr.WriteString(s)
}

func packLine(f string, a ...interface{}) string {
	s := f
	if len(a) > 0 {
		s = fmt.Sprintf(f, a...)
	}

	size := len(s)
	if size == 0 {
		return ""
	}

	if s[size-1] != '\n' {
		s += "\n"
	}

	return s
}

// ReadOptions 给几个选项给用户并返回用户的选择
func ReadOptions(description string, options []string) (selection map[int]string, err error) {
	if len(options) == 0 {
		return map[int]string{}, nil
	}

	optionsWithIndex := make(map[string]int, len(options))
	for i, o := range options {
		if _, exist := optionsWithIndex[o]; exist {
			return nil, fmt.Errorf("option '%s' is set more than once", o)
		}
		optionsWithIndex[o] = i
	}

	qs := []*survey.Question{
		{
			Name: "Options",
			Prompt: &survey.MultiSelect{
				Message: "Choose some colors:",
				Options: []string{"red", "blue", "green"},
			},
		},
	}

	answers := struct {
		Options []string
	}{}

	if err := survey.Ask(qs, &answers); err != nil {
		return nil, err
	}
	selection = make(map[int]string, len(answers.Options))
	for _, ans := range answers.Options {
		if index, exist := optionsWithIndex[ans]; exist {
			selection[index] = ans
		}
	}
	return selection, nil
}
