package http

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"golang.org/x/text/encoding"
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

// WithDebugger 添加调试函数
func WithDebugger(f func(string, ...any)) RequestOption {
	return func(ro *requestOption) {
		if f != nil {
			ro.debugf = f
		}
	}
}

// WithResponseCharset 指定响应 charset
func WithResponseCharset(c encoding.Encoding) RequestOption {
	return func(ro *requestOption) {
		ro.rspCharset = c
	}
}

type requestOption struct {
	method string
	header http.Header
	body   any
	query  url.Values
	debugf func(string, ...any)

	rspCharset encoding.Encoding
	marshaler  marshalerType
}

type marshalerType func(any) ([]byte, error)

func (o *requestOption) getBody() (io.Reader, error) {
	if o.body == nil {
		return nil, nil
	}
	if b, ok := o.body.([]byte); ok {
		o.debugf("request body '%s'", b)
		return bytes.NewBuffer(b), nil
	}
	b, e := o.marshaler(o.body)
	if e != nil {
		return nil, fmt.Errorf("Marshal request error (%w)", e)
	}
	o.debugf("request body '%s'", b)
	return bytes.NewBuffer(b), nil
}

func mergeOptions(opts []RequestOption, marshaler marshalerType) *requestOption {
	o := &requestOption{
		method:    "GET",
		header:    http.Header{},
		body:      nil,
		query:     url.Values{},
		debugf:    func(s string, a ...any) {},
		marshaler: marshaler,
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
