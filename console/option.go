package console

import "github.com/Andrew-M-C/go.util/unicode"

type options struct {
	separator string

	align struct {
		enabled bool
		unify   unicode.Align
		byCols  []unicode.Align
	}
}

func (o *options) getAlignAtIndex(i int) unicode.Align {
	if i < len(o.align.byCols) {
		return o.align.byCols[i]
	}
	return o.align.unify
}

func defaultOpt() *options {
	o := &options{
		separator: " ",
	}
	o.align.unify = unicode.AlignRight
	return o
}

func mergeOptions(opts []Option) *options {
	o := defaultOpt()
	for _, opt := range opts {
		opt(o)
	}
	return o
}

// Option 表示 console 包的额外参数
type Option func(*options)

// WithSeparator 指定分隔符
func WithSeparator(sep string) Option {
	if sep == "" {
		sep = " "
	}
	return func(o *options) {
		o.separator = sep
	}
}

// WithUnifyAlign 所有列统一采用对齐方式
func WithUnifyAlign(a unicode.Align) Option {
	return func(o *options) {
		o.align.enabled = true
		o.align.unify = a
	}
}

// WithAlignByCols 按列决定对齐方式
func WithAlignByCols(a ...unicode.Align) Option {
	return func(o *options) {
		o.align.enabled = true
		o.align.byCols = a
	}
}
