// Package crawler 用于实现一些简单的爬虫工具
package crawler

// HyperLink 表示一个超链接
type HyperLink struct {
	Title string `json:"title" yaml:"title"`
	URL   string `json:"url"   yaml:"url"`
}

// Debugger 调试日志器
type Debugger func(string, ...any)

// Option 表示选项
type Option func(*options)

// Reference:
//
//   - [最详细的GOOGLE搜索指令大全]
//   - [语言值 - Programmable Search Engine]
//
// [最详细的GOOGLE搜索指令大全]: https://zhuanlan.zhihu.com/p/136076792
// [语言值 - Programmable Search Engine]: https://developers.google.com/custom-search/docs/ref_languages?hl=zh-cn
type options struct {
	num      int
	language string
	debug    Debugger
}

func mergeOptions(opts ...Option) *options {
	opt := &options{
		num:      10,
		language: "",
		debug:    func(string, ...any) {},
	}
	for _, o := range opts {
		o(opt)
	}
	return opt
}

// WithDebugger 设置调试日志器
func WithDebugger(d Debugger) Option {
	return func(o *options) {
		if d != nil {
			o.debug = d
		}
	}
}

// WithNum 设置搜索结果数量
func WithNum(n int) Option {
	return func(o *options) {
		if n > 0 {
			o.num = n
		}
	}
}

// WithLanguage 设置搜索语言
func WithLanguage(l string) Option {
	return func(o *options) {
		if l != "" {
			o.language = l
		}
	}
}
