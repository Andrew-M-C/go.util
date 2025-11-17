// Package crawler 用于实现一些简单的爬虫工具
package crawler

import (
	"net/http"
	"time"

	"github.com/chromedp/cdproto/network"
)

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

	// 额外的 HTTP header 配置
	headers http.Header
}

func mergeOptions(opts ...Option) *options {
	opt := &options{
		num:      10,
		language: "",
		debug:    func(string, ...any) {},
		headers:  http.Header{},
	}
	for _, o := range opts {
		o(opt)
	}
	return opt
}

// WithHeader 设置额外的 HTTP header 配置
func WithHeader(h http.Header) Option {
	return func(o *options) {
		hCopy := http.Header{}
		for k, values := range h {
			for _, v := range values {
				hCopy.Add(k, v)
			}
		}
		// 然后使用 set 的方式添加
		for k, v := range hCopy {
			o.headers[k] = v
		}
	}
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

// ChromeCookiesToStandard 将 chromedp 的 cookie 类型转为 Go 标准库的类型
func ChromeCookiesToStandard(cookies []*network.Cookie) []*http.Cookie {
	res := make([]*http.Cookie, 0, len(cookies))
	for _, c := range cookies {
		if c == nil {
			continue
		}

		httpCookie := &http.Cookie{
			Name:     c.Name,
			Value:    c.Value,
			Path:     c.Path,
			Domain:   c.Domain,
			Expires:  time.Unix(int64(c.Expires), 0),
			Secure:   c.Secure,
			HttpOnly: c.HTTPOnly,
			SameSite: convertSameSite(c.SameSite),
		}

		res = append(res, httpCookie)
	}

	return res
}

func convertSameSite(sameSite network.CookieSameSite) http.SameSite {
	switch sameSite {
	case network.CookieSameSiteStrict:
		return http.SameSiteStrictMode
	case network.CookieSameSiteLax:
		return http.SameSiteLaxMode
	case network.CookieSameSiteNone:
		return http.SameSiteNoneMode
	default:
		return http.SameSiteDefaultMode
	}
}
