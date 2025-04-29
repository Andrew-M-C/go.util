package http

import (
	"context"
	"fmt"
	"mime"
	"net/url"
	"path"
	"strconv"

	"github.com/Andrew-M-C/go-bytesize"
)

// DownloadFile 下载文件
func DownloadFile(
	ctx context.Context, targetURL string, opts ...RequestOption,
) (fileName string, content []byte, err error) {
	o := mergeOptions(opts, nil)
	httpRsp, err := raw(ctx, targetURL, o)
	if err != nil {
		return "", nil, err
	}
	defer httpRsp.Body.Close()

	if v := httpRsp.Header.Get("Content-Length"); v != "" {
		if u, _ := strconv.ParseUint(v, 10, 64); u > 0 {
			o.debugf("content length: %v", bytesize.Base10(u))
		}
	}

	// 尝试用 mime 中读取文件名
	if c := httpRsp.Header.Get("Content-Disposition"); c != "" {
		_, params, err := mime.ParseMediaType(c)
		if err == nil { // 注意, 不是 !=
			fileName = params["filename"]
		}
	}

	// 尝试从路径中读取文件名
	if fileName == "" {
		if u, _ := url.Parse(targetURL); u != nil {
			fileName = path.Base(u.Path)
		}
	}

	// 读取正文
	content, err = readBody(o, httpRsp.Header.Get("Content-Encoding"), httpRsp.Body)
	if err != nil {
		return fileName, nil, fmt.Errorf("read body error (%w)", err)
	}

	return fileName, content, nil
}
