package unicode

import "strings"

type option struct {
	align Align
	blank string
	tab   string
}

func defaultOption() *option {
	return &option{
		align: AlignRight,
		blank: " ",
	}
}

func (o *option) blanks(repeat int) string {
	le := eastAsianStringWidth(o.blank)
	if repeat <= le {
		return o.blank
	}

	actualRepeat := repeat / le
	repeatRemain := repeat % le
	return strings.Repeat(o.blank, actualRepeat) + strings.Repeat(" ", repeatRemain)
}

type Option func(*option)

// WithAlign 返回关于对齐方式的选项
func WithAlign(align Align) Option {
	return func(o *option) {
		o.align = align
	}
}

// WithBlank 表示空格填充符号
func WithBlank(blank string) Option {
	return func(o *option) {
		o.blank = blank
	}
}

// WithTabWidth 表示将 tab 替换成的空格数。如果不指定则视为一个字符, 显示效果取决于终端
func WithTabWidth(width int) Option {
	if width <= 0 {
		width = 2
	}
	return func(o *option) {
		o.tab = strings.Repeat(" ", width)
	}
}
