// Package url 提供 net/url 的一些工具和替代逻辑
package url

import (
	"net/url"
	"strings"
)

type URL struct {
	url.URL
}

// NewURLByOfficial 使用官方 *url.URL 返回 URL 类型
func NewURLByOfficial(u *url.URL) *URL {
	return &URL{
		URL: *u,
	}
}

func (u URL) String() string {
	s := (&u.URL).String()

	if !strings.Contains(s, "%23") {
		return s
	}

	parts := strings.SplitN(s, "?", 2)
	parts[0] = strings.ReplaceAll(parts[0], "%23", "#")
	return strings.Join(parts, "?")
}
