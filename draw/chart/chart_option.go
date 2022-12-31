package chart

import (
	"strconv"

	"golang.org/x/exp/constraints"
)

type Option func(*opt)

func OptXString(f func(x float64) string) Option {
	return func(o *opt) {
		if f != nil {
			o.xNameFunc = f
		}
	}
}

func OptXScale(scale float64) Option {
	return func(o *opt) {
		if scale > 0 {
			o.xScale = scale
		}
	}
}

func OptYScale(scale float64) Option {
	return func(o *opt) {
		if scale > 0 {
			o.yScale = scale
		}
	}
}

func OptFontSize[T constraints.Integer](size T) Option {
	return func(o *opt) {
		if size > 0 {
			o.fontSize = int(size)
		}
	}
}

type opt struct {
	xNameFunc func(x float64) string
	xScale    float64
	yScale    float64
	fontSize  int
}

func ftos(f float64) string {
	return strconv.FormatFloat(f, 'g', -1, 64)
}

func defaultOpt() *opt {
	o := &opt{}
	o.xNameFunc = ftos
	o.fontSize = 7
	return o
}

func mergeOpts(opts []Option) *opt {
	o := defaultOpt()
	for _, opt := range opts {
		if opt != nil {
			opt(o)
		}
	}
	return o
}
