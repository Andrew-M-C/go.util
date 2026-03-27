package clawbot

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// ===========================
//  Poll — 消息轮询
// ===========================

// Poll 执行一次 getUpdates 长轮询，阻塞直到收到消息或服务端超时。
//
// 调用 API:
//
//	POST {creds.BaseURL}/ilink/bot/getupdates
//	Body: {"get_updates_buf": "<cursor>", "base_info": {"channel_version": "2.0.1-go"}}
//	Headers: Authorization: Bearer <token>, X-WECHAT-UIN: <random>, ...
//
// 参数:
//   - creds: 登录成功后获得的凭据（BotToken + BaseURL）
//   - getUpdatesBuf: 上次 Poll 返回的游标；首次调用传空字符串 ""
//
// 返回:
//   - 正常: PollResult 包含消息列表（可能为空）和下次轮询使用的新游标
//   - 会话过期: 返回包装了 ErrSessionExpired 的 error，可通过 errors.Is 判断
//   - 其他 API 错误: 返回 *APIError
//
// 单次调用最多阻塞约 35s（服务端长轮询超时 + 网络余量）。
// 调用方应在循环中反复调用，用上次返回的 GetUpdatesBuf 作为下次的入参:
//
//	buf := ""
//	for {
//	    result, err := utils.Poll(ctx, creds, buf)
//	    if errors.Is(err, utils.ErrSessionExpired) {
//	        break // 会话过期，需要重新登录
//	    }
//	    if err != nil {
//	        log.Println("poll error:", err)
//	        time.Sleep(3 * time.Second)
//	        continue
//	    }
//	    buf = result.GetUpdatesBuf
//	    for _, msg := range result.Messages {
//	        handleMessage(msg)
//	    }
//	}
func Poll(
	ctx context.Context, creds Credentials, getUpdatesBuf string,
) (PollResult, error) {
	rawBody, err := doGetUpdates(ctx, creds, getUpdatesBuf)
	if err != nil {
		return PollResult{}, err
	}

	var resp getUpdatesResp
	if err := json.Unmarshal(rawBody, &resp); err != nil {
		return PollResult{}, fmt.Errorf(
			"decode getUpdates response: %w", err)
	}

	// 检查 API 业务层错误（含会话过期判断）
	if apiErr := checkGetUpdatesError(resp); apiErr != nil {
		return PollResult{}, apiErr
	}

	return PollResult{
		Messages:      resp.Msgs,
		GetUpdatesBuf: resp.GetUpdatesBuf,
	}, nil
}

// ===========================
//  内部实现
// ===========================

// doGetUpdates 构造并执行 getUpdates HTTP 请求，返回原始响应体。
func doGetUpdates(
	ctx context.Context, creds Credentials, getUpdatesBuf string,
) ([]byte, error) {
	// 1. 构造请求 URL
	base := ensureTrailingSlash(creds.BaseURL)
	endpoint := base + "ilink/bot/getupdates"

	// 2. 构造 JSON 请求体
	reqBody := getUpdatesReq{
		GetUpdatesBuf: getUpdatesBuf,
		BaseInfo:      newBaseInfo(),
	}
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal getUpdates request: %w", err)
	}

	// 3. 创建 HTTP 请求

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("create getUpdates request: %w", err)
	}

	// 4. 设置必要的请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("AuthorizationType", "ilink_bot_token")
	req.Header.Set("X-WECHAT-UIN", randomWechatUIN())
	req.Header.Set("iLink-App-ClientVersion", "1")

	token := strings.TrimSpace(creds.BotToken)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	// 5. 执行请求
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		// 区分"调用方取消"和"本次请求自身超时"
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, fmt.Errorf("getUpdates HTTP error: %w", err)
	}
	defer resp.Body.Close()

	// 6. 读取并校验 HTTP 状态码
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read getUpdates response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf(
			"getUpdates HTTP %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// checkGetUpdatesError 检查 getUpdates 响应体中的业务层错误。
//
// iLink API 的错误以两种方式表达:
//   - ret 字段非零: 通用错误
//   - errcode 字段非零: 特定错误码（如 -14 = 会话过期）
//
// 当两者均为 0 时表示成功。
func checkGetUpdatesError(resp getUpdatesResp) error {
	ret := 0
	if resp.Ret != nil {
		ret = *resp.Ret
	}
	errCode := 0
	if resp.ErrCode != nil {
		errCode = *resp.ErrCode
	}

	// ret=0 且 errcode=0 表示正常
	if ret == 0 && errCode == 0 {
		return nil
	}

	// errcode=-14 或 ret=-14 表示会话过期
	if errCode == sessionExpiredErrCode || ret == sessionExpiredErrCode {
		return fmt.Errorf(
			"%w: ret=%d errcode=%d errmsg=%s",
			ErrSessionExpired, ret, errCode, resp.ErrMsg)
	}

	// 其他业务错误，包装为 *APIError 供调用方类型断言
	return &APIError{
		Ret:     ret,
		ErrCode: errCode,
		ErrMsg:  resp.ErrMsg,
	}
}

// IsSessionExpired 是 errors.Is(err, ErrSessionExpired) 的便捷别名。
func IsSessionExpired(err error) bool {
	return errors.Is(err, ErrSessionExpired) ||
		errors.Is(err, context.DeadlineExceeded)
}
