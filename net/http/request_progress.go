package http

import (
	"net/http"
	"strconv"
	"sync/atomic"
)

// RequestState 表示请求中间阶段
type RequestState int

const (
	// RequestInitialized 表示请求准备好, 准备发出
	RequestInitialized RequestState = iota
	// ResponseReceived 表示已经接收到响应 header, 但未读取 body
	ResponseReceived
	// ReceivingBody 表示正在读取响应 body
	ReceivingBody
)

// RequestProgress 表示请求进度, 用于回调
type RequestProgress struct {
	state         RequestState
	contentLength int64
	readLength    int64

	rsp *http.Response
}

// RequestState 获取请求状态
func (p *RequestProgress) RequestState() RequestState {
	return p.state
}

// ContentLength 从 response header 中判断文件大小。-1 表示未知
func (p *RequestProgress) ContentLength() int64 {
	if l := atomic.LoadInt64(&p.contentLength); l != 0 {
		return l
	}
	rsp := p.rsp
	if rsp == nil {
		return -1
	}

	v := rsp.Header.Get("Content-Length")
	if v == "" {
		return -1
	}
	i, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		i = -1
	}
	atomic.StoreInt64(&p.contentLength, i)
	return i
}

// ReadLength 表示已读取的 body 大小
func (p *RequestProgress) ReadLength() int64 {
	return atomic.LoadInt64(&p.readLength)
}

type requestProgressWriter struct {
	*RequestProgress

	callback func(*RequestProgress)
}

func (p *requestProgressWriter) Write(b []byte) (n int, err error) {
	le := len(b)
	_ = atomic.AddInt64(&p.readLength, int64(le))
	p.invokeIfNotNil(ReceivingBody)
	return le, nil
}

func (p *requestProgressWriter) invokeIfNotNil(s RequestState) {
	if p == nil {
		return
	}
	p.state = s

	if cb := p.callback; cb != nil {
		cb(p.RequestProgress)
	}
}
