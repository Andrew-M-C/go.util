package crawler

import (
	"context"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// BingSearch 执行必应搜索
func BingSearch(
	ctx context.Context, keywords string, opts ...Option,
) ([]HyperLink, error) {
	o := mergeBingSearchOptions(opts...)
	body, err := getBingBody(ctx, keywords, o)
	if err != nil {
		return nil, err
	}
	o.debug("%s", body)
	res, err := parseGoogleBody(body)
	if err != nil {
		return nil, err
	}
	return deDuplicateHyperLinks(res), nil
}

// ref:
//
//   - [搜索引擎Bing必应高级搜索语法](https://blog.csdn.net/hansel/article/details/53886828)
func mergeBingSearchOptions(opts ...Option) *options {
	opt := mergeOptions(opts...)

	// 特殊参数翻译
	opt.language = googleLangs[opt.language]
	return opt
}

func getBingBody(ctx context.Context, keywords string, o *options) (string, error) {
	keywords = strings.TrimSpace(keywords)
	const targetURL = "https://cn.bing.com/search"

	u, _ := url.Parse(targetURL)
	q := url.Values{}

	if o.language != "" {
		q.Add("setLang", o.language)
	}
	if o.num > 0 {
		q.Add("count", strconv.Itoa(o.num))
	}

	q.Add("q", keywords)
	u.RawQuery = q.Encode()
	finalTargetURL := u.String()
	o.debug("target URL: '%s'", finalTargetURL)

	start := time.Now()
	body, err := GetHTML(ctx, finalTargetURL)
	ela := time.Since(start)
	if err != nil {
		return "", err
	}

	o.debug("搜索 '%s' 耗时 %v", keywords, ela)
	return body, nil
}
