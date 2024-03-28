package http

import (
	"net/http"
	"net/url"
)

// 发起请求的额外选项
type RequestOption func(*requestOption)

// WithMethod 指定 HTTP 调用方法
func WithMethod(method string) RequestOption {
	return func(ro *requestOption) {
		if method != "" {
			ro.method = method
		}
	}
}

// WithRequestHeader 指定请求头参数
func WithRequestHeader(h http.Header) RequestOption {
	return func(ro *requestOption) {
		for k, values := range h {
			for _, v := range values {
				ro.header.Add(k, v)
			}
		}
	}
}

// WithRequestBody 请求正文
func WithRequestBody(req any) RequestOption {
	return func(ro *requestOption) {
		ro.body = req
	}
}

// WithQuery query 参数
func WithQuery(q url.Values) RequestOption {
	return func(ro *requestOption) {
		for k, values := range q {
			for _, v := range values {
				ro.query.Add(k, v)
			}
		}
	}
}

type requestOption struct {
	method string
	header http.Header
	body   any
	query  url.Values
}

func mergeOptions(opts []RequestOption) *requestOption {
	o := &requestOption{
		method: "GET",
		header: http.Header{},
		body:   nil,
		query:  url.Values{},
	}
	for _, f := range opts {
		if f != nil {
			f(o)
		}
	}
	return o
}

func (o *requestOption) mergeQuery(q url.Values) {
	for k, values := range q {
		for _, v := range values {
			o.query.Add(k, v)
		}
	}
}
