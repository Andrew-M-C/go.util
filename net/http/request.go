package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// JSON 发起一个 JSON 请求
func JSON[T any](ctx context.Context, targetURL string, opts ...RequestOption) (rsp T, err error) {
	o := mergeOptions(opts)

	var reqBody io.Reader
	if o.body != nil {
		b, e := json.Marshal(o.body)
		if e != nil {
			err = fmt.Errorf("Marshal request error (%w)", err)
			return
		}
		reqBody = bytes.NewBuffer(b)
	}

	u, err := url.Parse(targetURL)
	if err != nil {
		err = fmt.Errorf("illegal target URL (%w)", err)
		return
	}
	o.mergeQuery(u.Query())
	u.RawQuery = o.query.Encode()

	httpReq, err := http.NewRequestWithContext(ctx, o.method, u.String(), reqBody)
	if err != nil {
		err = fmt.Errorf("http.NewRequest error (%w)", err)
		return
	}

	o.header.Set("Content-Type", "application/json")
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
	if err = json.Unmarshal(b, &rsp); err != nil {
		err = fmt.Errorf("json.Unmarshal error (%w)", err)
		return
	}

	return rsp, err
}
