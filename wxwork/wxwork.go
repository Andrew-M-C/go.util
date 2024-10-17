// Package wxwork 实现一些在阅读代码时顺便实现的企业微信接口小逻辑
package wxwork

import (
	"context"
	"time"
)

const defaultTimeout = 5 * time.Second

// AccessTokenGetter 实现 access_token 自动刷新获取的逻辑
//
// Reference:
//   - [获取access_token - 接口文档 - 企业微信开发者中心](https://developer.work.weixin.qq.com/document/path/91039)
//   - [关于企业微信对接内部应用开发，access_token的管理机制和业务接口调用项目实战的八个要点](https://blog.csdn.net/privateHiroki/article/details/110819337)
type AccessTokenGetter interface {
	GetAccessToken(context.Context) (string, error)
}

// NewAccessTokenGetter 新建一个 access_token 获取器。
//
// 该逻辑目前暂时不支持 close, 一旦创建了就占用内存, 请注意。
func NewAccessTokenGetter(corpID, corpSecret string, opts ...Option) (AccessTokenGetter, error) {
	return newAccessTokenGetter(corpID, corpSecret, opts...)
}
