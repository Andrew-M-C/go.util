// Package wxwork 实现一些在阅读代码时顺便实现的企业微信接口小逻辑
package wxwork

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/Andrew-M-C/go.util/log"
	hutil "github.com/Andrew-M-C/go.util/net/http"
	"github.com/Andrew-M-C/go.util/recovery"
	"github.com/Andrew-M-C/go.util/runtime/caller"
)

func newAccessTokenGetter(corpID, corpSecret string, opts ...Option) (*accessTokenImpl, error) {
	if corpID == "" || corpSecret == "" {
		return nil, errors.New("缺少有效的 corpid 或 corpsecret 参数")
	}

	impl := &accessTokenImpl{
		corpID: corpID,
		secret: corpSecret,
		opts:   mergeOptions(opts),
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	accessToken, err := impl.getNewAccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取 access_token 失败 (%w)", err)
	}
	impl.token = accessToken
	go impl.doRefresh()

	return impl, nil
}

type accessTokenImpl struct {
	// 参数
	corpID, secret string
	// 各种选项
	opts *options
	// 当前可用的 access_token
	token      *accessToken
	refreshErr error
}

func (impl *accessTokenImpl) GetAccessToken(context.Context) (string, error) {
	token := impl.token
	if token.expired() {
		return "", impl.refreshErr
	}
	return token.token, nil
}

func (impl *accessTokenImpl) getNewAccessToken(ctx context.Context) (*accessToken, error) {
	q := url.Values{}
	q.Add("corpid", impl.corpID)
	q.Add("corpsecret", impl.secret)

	const target = "https://qyapi.weixin.qq.com/cgi-bin/gettoken"
	start := time.Now()
	rsp, err := hutil.JSON[getTokenRsp](
		ctx, target,
		hutil.WithMethod("GET"), hutil.WithQuery(q),
		hutil.WithDebugger(impl.opts.debugf),
	)
	if err != nil {
		return nil, err
	}
	if rsp.ErrCode != 0 {
		return nil, fmt.Errorf("[%d] %v", rsp.ErrCode, rsp.ErrMsg)
	}
	until := start.Add(time.Duration(rsp.TTLSec) * time.Second)
	until = until.Add(-time.Minute) // 提前一分钟刷新 access_token, 避免边界条件
	impl.opts.debugf("刷新 access_token: %s, TTL %v, 预计刷新时间 %v", rsp.AccessToken, rsp.TTLSec, until)

	token := &accessToken{
		token: rsp.AccessToken,
		until: until,
	}
	return token, nil
}

func (impl *accessTokenImpl) doRefresh() {
	defer recovery.CatchPanic(
		recovery.WithCallback(func(info any, stack []caller.Caller) {
			impl.opts.debugf("routine panic, info '%v', stack %v", info, log.ToJSON(stack))
			go impl.doRefresh()
		}),
	)
	iterate := func() {
		next := impl.token.until
		time.Sleep(time.Until(next))

		ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
		defer cancel()

		token, err := impl.getNewAccessToken(ctx)
		if err != nil {
			impl.opts.debugf("请求刷新 access_token 失败: '%v'", err)
			impl.refreshErr = err
			return
		}

		impl.token = token
		impl.refreshErr = nil
	}
	for {
		iterate()
		time.Sleep(defaultTimeout)
	}
}

type accessToken struct {
	token string
	until time.Time
}

func (t *accessToken) expired() bool {
	return time.Now().After(t.until)
}

type getTokenRsp struct {
	ErrCode     int    `json:"errcode"`
	ErrMsg      string `json:"errmsg"`
	AccessToken string `json:"access_token"`
	TTLSec      int64  `json:"expires_in"`
}
