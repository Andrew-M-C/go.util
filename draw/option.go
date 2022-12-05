package draw

import "image/color"

type option struct {
	div          float64
	defaultColor color.Color
	drawColor    color.Color
}

func defaultOption() *option {
	o := &option{}
	o.div = 1
	o.defaultColor = color.Black
	return o
}

func mergeOptions(opts []Option) *option {
	opt := defaultOption()
	for _, o := range opts {
		o(opt)
	}
	return opt
}

// Option 表示一个可选参数
type Option func(*option)

// WithZoomOutFactor 缩小倍数, 暂不支持放大
func WithZoomOutFactor(f float64) Option {
	return func(o *option) {
		if f <= 1 {
			f = 1
		}
		o.div = f
	}
}

// WithDefaultDrawColor 设置默认绘图颜色
func WithDefaultDrawColor(c color.Color) Option {
	return func(o *option) {
		if c != nil {
			o.defaultColor = c
		}
	}
}

// WithColor 设置当前绘图颜色
func WithColor(c color.Color) Option {
	return func(o *option) {
		if c != nil {
			o.drawColor = c
		}
	}
}
