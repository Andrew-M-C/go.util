package deepcopy

// Option 表示额外选项
type Option func(*generator)

// WithLogFunc 指定日志函数
func WithLogFunc(f func(string, ...interface{})) Option {
	return func(g *generator) {
		if f != nil {
			g.log = f
		}
	}
}

// WithDebug 开/关调试日志
func WithDebug(on bool) Option {
	return func(g *generator) {
		g.debug = on
	}
}

// WithPackageName 指定 package 名称
func WithPackageName(n string) Option {
	return func(g *generator) {
		if n != "" {
			g.packageName = n
		}
	}
}

// WithFunctionMethodName 指定深复制的函数/方法名。不指定的话默认是 DeepCopy() 或者是 DeepCopyXxxxx
func WithFunctionMethodName(n string) Option {
	return func(g *generator) {
		if n != "" {
			g.funcOrMethodName = n
		}
	}
}
