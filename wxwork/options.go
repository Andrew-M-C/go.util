package wxwork

type options struct {
	debugf func(string, ...any)
}

// Option 表示额外选项
type Option func(*options)

// WithDebugger 传入调试函数
func WithDebugger(f func(string, ...any)) Option {
	return func(o *options) {
		if f != nil {
			o.debugf = f
		}
	}
}

func mergeOptions(opts []Option) *options {
	o := &options{
		debugf: func(s string, a ...any) {},
	}

	for _, fu := range opts {
		if fu != nil {
			fu(o)
		}
	}
	return o
}
