package http

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Andrew-M-C/go.util/log"
	"github.com/Andrew-M-C/go.util/unsafe"
	"golang.org/x/net/html/charset"
	"golang.org/x/text/encoding"
	"golang.org/x/text/transform"
)

// Raw 发起一个请求, 但是返回 []byte
func Raw(ctx context.Context, targetURL string, opts ...RequestOption) (rsp []byte, err error) {
	o := mergeOptions(opts, json.Marshal)
	httpRsp, err := raw(ctx, targetURL, o)
	if err != nil {
		return nil, err
	}
	defer httpRsp.Body.Close()

	return readBody(o, httpRsp.Body)
}

func raw(ctx context.Context, targetURL string, o *requestOption) (*http.Response, error) {
	reqBody, err := o.getBody()
	if err != nil {
		return nil, err
	}
	u, err := url.Parse(targetURL)
	if err != nil {
		return nil, fmt.Errorf("illegal target URL (%w)", err)
	}
	o.mergeQuery(u.Query())
	u.RawQuery = o.query.Encode()

	fullURL := u.String()
	o.debugf("request URL: %s", fullURL)

	httpReq, err := http.NewRequestWithContext(ctx, o.method, fullURL, reqBody)
	if err != nil {
		return nil, fmt.Errorf("http.NewRequest error (%w)", err)
	}
	httpReq.Header = o.header
	o.progress.invokeIfNotNil(ResponseReceived)

	cli := http.Client{Transport: http.DefaultTransport}
	start := time.Now()
	httpRsp, err := cli.Do(httpReq)
	ela := time.Since(start)
	if err != nil {
		return nil, fmt.Errorf("cli.Do error (%w)", err)
	}

	if o.progress != nil {
		o.progress.rsp = httpRsp
	}
	o.progress.invokeIfNotNil(ResponseReceived)
	o.debugf("request done, ela %v, status %v", ela, httpRsp.Status)

	if httpRsp.StatusCode != http.StatusOK {
		return nil, errors.New(httpRsp.Status)
	}
	return httpRsp, nil
}

func readBody(o *requestOption, body io.ReadCloser) ([]byte, error) {
	if o.progressCB == nil {
		b, err := io.ReadAll(body)
		if err != nil {
			return nil, fmt.Errorf("io.ReadAll error (%w)", err)
		}
		return b, nil
	}

	buff := &bytes.Buffer{}
	w := io.MultiWriter(buff, o.progress)
	if _, err := io.Copy(w, body); err != nil {
		return nil, fmt.Errorf("io.ReadAll error (%w)", err)
	}

	return buff.Bytes(), nil
}

// rawAndRead raw 请求并 io.ReadAll, 拿到的 response 无需 close
func rawAndRead(ctx context.Context, targetURL string, o *requestOption) (*http.Response, []byte, error) {
	rsp, err := raw(ctx, targetURL, o)
	if err != nil {
		return rsp, nil, err
	}

	defer rsp.Body.Close()

	b, err := readBody(o, rsp.Body)
	if err != nil {
		return rsp, nil, err
	}
	return rsp, b, nil
}

// JSON 发起一个 JSON 请求
func JSON[T any](ctx context.Context, targetURL string, opts ...RequestOption) (*T, error) {
	o := mergeOptions(opts, json.Marshal)
	if o.body != nil {
		o.header.Set("Content-Type", "application/json")
	}
	httpRsp, b, err := rawAndRead(ctx, targetURL, o)
	if err != nil {
		return nil, err
	}
	if len(b) == 0 {
		return nil, errors.New("empty body from remote server")
	}

	o.debugf("response: '%s'", bytesStringer(b))
	o.debugf("response header: %+v", httpRsp.Header)

	b = decodeIfNecessary(o, b, httpRsp)
	rsp := new(T)
	if err := json.Unmarshal(b, rsp); err != nil {
		return nil, fmt.Errorf("json.Unmarshal error (%w)", err)
	}
	return rsp, nil
}

// XMLGetRspBody 发起一个 XML 请求并返回包体字节
func XMLGetRspBody(ctx context.Context, targetURL string, opts ...RequestOption) ([]byte, error) {
	o := mergeOptions(opts, xml.Marshal)
	if o.body != nil {
		o.header.Set("Content-Type", "application/xml")
	}
	httpRsp, b, err := rawAndRead(ctx, targetURL, o)
	if err != nil {
		return nil, err
	}
	if len(b) == 0 {
		return nil, errors.New("empty body from remote server")
	}

	o.debugf("response: '%s'", bytesStringer(b))
	o.debugf("response header: %+v", httpRsp.Header)

	b = decodeIfNecessary(o, b, httpRsp)
	return b, nil
}

// XML 发起一个 XML 请求
//
// WARNING: 未测试, 请注意
func XML[T any](ctx context.Context, targetURL string, opts ...RequestOption) (*T, error) {
	b, err := XMLGetRspBody(ctx, targetURL, opts...)
	if err != nil {
		return nil, err
	}
	rsp := new(T)
	if err := xml.Unmarshal(b, &rsp); err != nil {
		return nil, fmt.Errorf("xml.Unmarshal error (%w)", err)
	}
	return rsp, nil
}

func decodeIfNecessary(o *requestOption, b []byte, httpRsp *http.Response) []byte {
	dec := o.rspCharset
	if dec != nil {
		o.debugf("指定使用 charset %v", dec)
	} else {
		contentType := httpRsp.Header.Get("Content-Type")
		charset, name, certain := charset.DetermineEncoding(b, contentType)
		o.debugf("响应编码为 '%v', 是否确定 %v, 采用编码器 '%v'", name, certain, charset)
		if charset == nil {
			return b
		}
		dec = charset
	}
	if dec == encoding.Nop {
		o.debugf("指定 charset 为 nop, 则默认为 UTF-8")
		return b
	}

	utf8Reader := transform.NewReader(bytes.NewReader(b), dec.NewDecoder())
	utf8Byte, err := io.ReadAll(utf8Reader)
	if err != nil {
		o.debugf("解码失败, 预测编码为 %v, 错误 %v", dec, err)
		return b
	}

	return utf8Byte
}

func bytesStringer(b []byte) any {
	s := unsafe.BtoS(b)
	if !strings.Contains(s, "\n") {
		return s
	}
	return log.ToJSON(s)
}
