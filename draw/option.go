package draw

import "image/color"

// MergedOptions 表示合并后的参数。调用方无需使用，这是给实现用的
type MergedOptions interface {
	Color() color.Color
	FontSize() int
	Rotate() float64
}

// Option 表示额外参数
type Option func(o *optImpl)

func defaultOption(c Canvas) *optImpl {
	o := &optImpl{
		color: c.CurrentDrawColor(),
	}
	_, height := c.Size()
	o.font.size = int(height) / 10
	if o.font.size <= 0 {
		o.font.size = 1
	}
	return o
}

// MergeOptions 给实现方使用, 实现各种 option
func MergeOptions(c Canvas, opts []Option) MergedOptions {
	o := defaultOption(c)
	for _, opt := range opts {
		if opt != nil {
			opt(o)
		}
	}
	return o
}

type optImpl struct {
	color color.Color
	font  struct {
		size int
	}
	section struct {
		rotate float64
	}
}

func (o *optImpl) Color() color.Color {
	return o.color
}

func (o *optImpl) FontSize() int {
	return o.font.size
}

func (o *optImpl) Rotate() float64 {
	return o.section.rotate
}

// -------- 具体 option 函数 --------

// OptColor 指定此次绘制颜色
func OptColor(clr color.Color) Option {
	return func(o *optImpl) {
		if clr != nil {
			o.color = clr
		}
	}
}

// OptFontSize 指定字号
func OptFontSize[T Number](size T) Option {
	return func(o *optImpl) {
		s := int(size)
		if s < 0 {
			s = -s
		}
		o.font.size = s
	}
}

// OptRotate 指定旋转角度
func OptRotate[T Number](angle T) Option {
	return func(o *optImpl) {
		o.section.rotate = float64(angle)
	}
}
