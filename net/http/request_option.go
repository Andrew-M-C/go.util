package http

import (
	"bytes"
	"encoding/json"
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

// WithRequestCookies 设定请求头中的 Cookie
func WithRequestCookies(cookies []*http.Cookie) RequestOption {
	return func(ro *requestOption) {
		buff := bytes.Buffer{}
		for _, cookie := range cookies {
			if cookie == nil {
				continue
			}
			if buff.Len() > 0 {
				buff.WriteRune(';')
			}
			buff.WriteString(cookie.Name)
			buff.WriteByte('=')
			buff.WriteString(cookie.Value)
		}
		if buff.Len() == 0 {
			return
		}
		ro.header.Set("Cookie", buff.String())
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

// WithProgressCallback 指定请求回调, 一般用在预期 body 很大的场景, 比如下载大文件
func WithProgressCallback(cb func(*RequestProgress)) RequestOption {
	return func(ro *requestOption) {
		ro.progressCB = cb
	}
}

// WithSSEUnmarshalErrorCallback 指定反序列化错误回调, 用在 SSE 场景。当反序列化失败时,
// 调用该回调, 而不返回错误 (相当于出错但 continue)
func WithSSEUnmarshalErrorCallback(cb func(error, string)) RequestOption {
	return func(ro *requestOption) {
		ro.sseUnmarshalErrorCB = cb
	}
}

type requestOption struct {
	method string
	header http.Header
	body   any
	query  url.Values
	debugf func(string, ...any)

	rspCharset  encoding.Encoding
	marshaler   marshalerType
	unmarshaler unmarshalerType

	progressCB func(*RequestProgress)
	progress   *requestProgressWriter

	sseUnmarshalErrorCB func(error, string)
}

type marshalerType func(any) ([]byte, error)
type unmarshalerType func([]byte, any) error

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
		return nil, fmt.Errorf("marshal request error (%w)", e)
	}
	o.debugf("request body '%s'", b)
	return bytes.NewBuffer(b), nil
}

func mergeOptions(opts []RequestOption, marshaler marshalerType) *requestOption {
	o := &requestOption{
		method:      "GET",
		header:      http.Header{},
		body:        nil,
		query:       url.Values{},
		debugf:      func(string, ...any) {},
		marshaler:   marshaler,
		unmarshaler: json.Unmarshal,
	}
	for _, f := range opts {
		if f != nil {
			f(o)
		}
	}

	// request progress
	if o.progressCB == nil {
		return o
	}
	o.progress = &requestProgressWriter{
		RequestProgress: &RequestProgress{},
		callback:        o.progressCB,
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
