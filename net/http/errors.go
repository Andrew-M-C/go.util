package http

import (
	"errors"
	"net/http"
)

// Error 表示本 http 工具返回的错误
type Error struct {
	wrapped error

	detail ErrorDetail
}

func (e Error) Error() string {
	if e.wrapped == nil {
		return ""
	}
	return e.wrapped.Error()
}

func (e Error) Unwrap() error {
	return e.wrapped
}

// ErrorDetail 错误详情
func (e Error) Detail() ErrorDetail {
	return e.detail
}

// ErrorDetail 错误详情
type ErrorDetail struct {
	// From	http.Response
	Status     string // e.g. "200 OK"
	StatusCode int    // e.g. 200
	Proto      string // e.g. "HTTP/1.0"
	ProtoMajor int    // e.g. 1
	ProtoMinor int    // e.g. 0
	Header     http.Header

	Body []byte
}

func packError(res *http.Response, body []byte, err error) *Error {
	e := &Error{
		wrapped: err,
	}
	if res != nil {
		e.detail = ErrorDetail{
			Status:     res.Status,
			StatusCode: res.StatusCode,
			Proto:      res.Proto,
			ProtoMajor: res.ProtoMajor,
			ProtoMinor: res.ProtoMinor,
			Header:     res.Header,
			Body:       body,
		}
	}
	return e
}

// UnwrapError 尝试将 error 转换为 Error 类型
func UnwrapError(err error) (Error, bool) {
	if err == nil {
		return Error{}, false
	}

	var e *Error
	if errors.As(err, &e) {
		return *e, true
	}

	return Error{}, false
}
