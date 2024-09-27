package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/Andrew-M-C/go.util/log"
	"github.com/Andrew-M-C/go.util/unsafe"
)

// Raw 发起一个请求, 但是返回 []byte
func Raw(ctx context.Context, targetURL string, opts ...RequestOption) (rsp []byte, err error) {
	o := mergeOptions(opts)

	reqBody, err := o.getBody()
	if err != nil {
		return rsp, err
	}
	u, err := url.Parse(targetURL)
	if err != nil {
		err = fmt.Errorf("illegal target URL (%w)", err)
		return
	}
	o.mergeQuery(u.Query())
	u.RawQuery = o.query.Encode()

	fullURL := u.String()
	o.debugf("request URL: %s", fullURL)

	httpReq, err := http.NewRequestWithContext(ctx, o.method, fullURL, reqBody)
	if err != nil {
		err = fmt.Errorf("http.NewRequest error (%w)", err)
		return
	}
	httpReq.Header = o.header

	cli := http.Client{
		Transport: http.DefaultTransport,
	}
	httpRsp, err := cli.Do(httpReq)
	if err != nil {
		err = fmt.Errorf("cli.Do error (%w)", err)
		return
	}
	if httpRsp.StatusCode != 200 {
		err = errors.New(httpRsp.Status)
		return
	}
	defer httpRsp.Body.Close()
	b, err := io.ReadAll(httpRsp.Body)
	if err != nil {
		err = fmt.Errorf("io.ReadAll error (%w)", err)
		return
	}
	return b, nil
}

// JSON 发起一个 JSON 请求
func JSON[T any](ctx context.Context, targetURL string, opts ...RequestOption) (*T, error) {
	o := mergeOptions(opts)

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

	if reqBody != nil {
		o.header.Set("Content-Type", "application/json")
	}
	httpReq.Header = o.header

	cli := http.Client{
		Transport: http.DefaultTransport,
	}
	httpRsp, err := cli.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("cli.Do error (%w)", err)
	}
	if httpRsp.StatusCode != 200 {
		return nil, errors.New(httpRsp.Status)
	}
	defer httpRsp.Body.Close()
	b, err := io.ReadAll(httpRsp.Body)
	if err != nil {
		return nil, fmt.Errorf("io.ReadAll error (%w)", err)
	}
	if len(b) == 0 {
		return nil, errors.New("empty body from remote server")
	}
	o.debugf("response: '%s'", bytesStringer(b))
	rsp := new(T)
	if err := json.Unmarshal(b, rsp); err != nil {
		return nil, fmt.Errorf("json.Unmarshal error (%w)", err)
	}

	return rsp, nil
}

func bytesStringer(b []byte) any {
	s := unsafe.BtoS(b)
	if !strings.Contains(s, "\n") {
		return s
	}
	return log.ToJSON(s)
}
