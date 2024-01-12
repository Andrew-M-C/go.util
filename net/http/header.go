// Package http 提供 net/http 包的一些工具
package http

import (
	"net/http"
)

// NormalizeHeader 返回统一 mine key 格式的 header
func NormalizeHeader(h http.Header) http.Header {
	out := make(http.Header, len(h))
	for k, vList := range h {
		for _, v := range vList {
			out.Add(k, v)
		}
	}
	return out
}
