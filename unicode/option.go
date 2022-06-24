package unicode

import "strings"

type option struct {
	align Align
	blank string
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

func WithBlank(blank string) Option {
	return func(o *option) {
		o.blank = blank
	}
}
