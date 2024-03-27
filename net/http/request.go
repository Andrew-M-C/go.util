package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"net/http"
)

// JSON 发起一个 JSON 请求
func JSON[T any](
	ctx context.Context, method, url string, req any, header http.Header,
) (rsp T, err error) {
	var reqBody io.Reader
	if req != nil {
		b, _ := json.Marshal(req)
		reqBody = bytes.NewBuffer(b)
	}

	httpReq, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		err = fmt.Errorf("http.NewRequest error (%w)", err)
		return
	}

	header = maps.Clone(header)
	if header == nil {
		header = http.Header{}
	}
	header.Set("Content-Type", "application/json")
	httpReq.Header = header

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
