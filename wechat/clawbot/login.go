package clawbot

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// ===========================
//  登录流程相关常量
// ===========================

const (
	// defaultQRPollTimeout 是单次 QR 状态轮询的 HTTP 超时。
	// iLink 服务端对此请求做长轮询，服务端无消息时约 30s 后返回。
	// 客户端超时设得略长于服务端超时，避免误断连接。
	defaultQRPollTimeout = 35 * time.Second

	// maxQRRefreshCount 是二维码过期后自动刷新的最大次数。
	// 超过此次数后 WaitForLogin 将放弃并返回错误。
	maxQRRefreshCount = 3

	// defaultLoginTimeout 是整体登录流程的默认超时。
	// 若调用方的 ctx 未设置 deadline，WaitForLogin 自动应用此超时。
	defaultLoginTimeout = 8 * time.Minute

	// ilinkBotType 是 iLink Bot 的固定类型标识，当前微信侧仅支持 "3"。
	ilinkBotType = "3"

	// qrPollInterval 是两次 QR 状态轮询之间的间隔。
	// 由于服务端已做长轮询（≈30s），此间隔无需太长，仅作为请求间缓冲。
	qrPollInterval = 1 * time.Second
)

// ===========================
//  FetchQRCode — 获取二维码
// ===========================

// FetchQRCode 从 iLink API 获取一个新的登录二维码。
//
// 调用 API:
//
//	GET {baseURL}/ilink/bot/get_bot_qrcode?bot_type=3
//
// baseURL 为空时自动使用 DefaultBaseURL（首次登录的常见场景）。
// 重新登录时可传入之前 Credentials.BaseURL 保存的地址。
//
// 返回值:
//   - QRCodeResult.QRCodeURL: 二维码图片地址，展示给用户扫描
//   - QRCodeResult.QRCode: 二维码唯一 key，传给 WaitForLogin 使用
func FetchQRCode(ctx context.Context, baseURL string) (QRCodeResult, error) {
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	base := ensureTrailingSlash(baseURL)
	endpoint := base + "ilink/bot/get_bot_qrcode?bot_type=" +
		url.QueryEscape(ilinkBotType)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return QRCodeResult{}, fmt.Errorf("create QR request: %w", err)
	}
	req.Header.Set("iLink-App-ClientVersion", "1")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return QRCodeResult{}, fmt.Errorf("fetch QR code: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return QRCodeResult{}, fmt.Errorf(
			"QR code API HTTP %d: %s", resp.StatusCode, string(body))
	}

	var result qrCodeResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return QRCodeResult{}, fmt.Errorf("decode QR response: %w", err)
	}

	return QRCodeResult{
		QRCode:    result.QRCode,
		QRCodeURL: result.QRCodeImgContent,
	}, nil
}

// ===========================
//  WaitForLogin — 等待扫码确认
// ===========================

// WaitForLogin 阻塞等待用户扫码并确认登录，返回 Bot 凭据。
//
// 内部行为:
//  1. 循环调用 get_qrcode_status 长轮询（每次最多 35s）
//  2. 用户扫码但未确认时，通过 cb.OnScanned 通知
//  3. 二维码过期时自动刷新（最多 3 次），通过 cb.OnQRRefreshed 通知新 URL
//  4. 整体默认 8 分钟超时；调用方可通过 ctx 提前取消
//
// 成功时返回的 Credentials 中四个字段均来自微信响应，可直接用于后续 API 调用。
// baseURL 为空时自动使用 DefaultBaseURL，与 FetchQRCode 保持一致。
func WaitForLogin(
	ctx context.Context, baseURL string, qr QRCodeResult, cb LoginCallbacks,
) (Credentials, error) {
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	// 如果调用方未在 ctx 上设置截止时间，应用默认的 8 分钟超时
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, defaultLoginTimeout)
		defer cancel()
	}

	currentQR := qr
	refreshCount := 0
	scannedNotified := false // 避免重复触发 OnScanned

	for {
		// 检查上下文是否已被取消或超时
		if err := ctx.Err(); err != nil {
			return Credentials{}, fmt.Errorf(
				"login cancelled or timed out: %w", err)
		}

		status, err := pollQRStatus(ctx, baseURL, currentQR.QRCode)
		if err != nil {
			return Credentials{}, err
		}

		switch status.Status {
		case "wait":
			// 用户尚未扫码，继续等待

		case "scaned":
			// 用户已扫码但未确认
			if !scannedNotified && cb.OnScanned != nil {
				cb.OnScanned()
				scannedNotified = true
			}

		case "expired":
			// 二维码过期，尝试刷新
			newQR, refreshErr := refreshQRCode(ctx, baseURL, &refreshCount, cb)
			if refreshErr != nil {
				return Credentials{}, refreshErr
			}
			currentQR = newQR
			scannedNotified = false // 新二维码需要重新扫码

		case "confirmed":
			// 登录成功
			return buildCredentials(status)

		default:
			return Credentials{}, fmt.Errorf(
				"unexpected QR status: %q", status.Status)
		}

		// 短暂等待后进入下一轮轮询
		select {
		case <-ctx.Done():
			return Credentials{}, fmt.Errorf(
				"login cancelled or timed out: %w", ctx.Err())
		case <-time.After(qrPollInterval):
		}
	}
}

// ===========================
//  内部辅助函数
// ===========================

// refreshQRCode 处理二维码过期：获取新的 QR 码并通知调用方。
// 累计超过 maxQRRefreshCount 次刷新后放弃并返回错误。
func refreshQRCode(
	ctx context.Context, baseURL string, refreshCount *int, cb LoginCallbacks,
) (QRCodeResult, error) {
	*refreshCount++
	if *refreshCount > maxQRRefreshCount {
		return QRCodeResult{}, fmt.Errorf(
			"QR code expired %d times, giving up", maxQRRefreshCount)
	}

	newQR, err := FetchQRCode(ctx, baseURL)
	if err != nil {
		return QRCodeResult{}, fmt.Errorf("refresh QR code: %w", err)
	}

	// 通知调用方新的二维码 URL，以便更新 UI
	if cb.OnQRRefreshed != nil {
		cb.OnQRRefreshed(newQR.QRCodeURL)
	}
	return newQR, nil
}

// pollQRStatus 执行一次 QR 码状态长轮询。
//
// 调用 API:
//
//	GET {baseURL}/ilink/bot/get_qrcode_status?qrcode=<qrcode>
//
// 此接口在服务端做长轮询（≈30s 无变化才返回 "wait"）。
// 客户端超时（35s）到达时视为 status="wait"，而非错误。
func pollQRStatus(
	ctx context.Context, baseURL string, qrcode string,
) (statusResponse, error) {
	base := ensureTrailingSlash(baseURL)
	endpoint := base + "ilink/bot/get_qrcode_status?qrcode=" +
		url.QueryEscape(qrcode)

	// 每次轮询有独立的 35s 超时，防止无限挂起
	pollCtx, cancel := context.WithTimeout(ctx, defaultQRPollTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(
		pollCtx, http.MethodGet, endpoint, nil)
	if err != nil {
		return statusResponse{}, fmt.Errorf(
			"create QR poll request: %w", err)
	}
	req.Header.Set("iLink-App-ClientVersion", "1")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		// 如果是本次轮询自身超时，返回 "wait" 让外层继续循环
		if pollCtx.Err() == context.DeadlineExceeded && ctx.Err() == nil {
			return statusResponse{Status: "wait"}, nil
		}
		return statusResponse{}, fmt.Errorf("poll QR status: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return statusResponse{}, fmt.Errorf(
			"QR status API HTTP %d: %s", resp.StatusCode, string(body))
	}

	var result statusResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return statusResponse{}, fmt.Errorf(
			"decode QR status: %w", err)
	}
	return result, nil
}

// buildCredentials 从登录确认的 statusResponse 构造 Credentials。
func buildCredentials(status statusResponse) (Credentials, error) {
	if status.BotToken == "" {
		return Credentials{}, fmt.Errorf(
			"login confirmed but bot_token is empty in response")
	}
	if status.ILinkBotID == "" {
		return Credentials{}, fmt.Errorf(
			"login confirmed but ilink_bot_id is empty in response")
	}

	return Credentials{
		Token:     status.BotToken,
		BaseURL:   status.BaseURL,
		AccountID: normalizeAccountID(status.ILinkBotID),
		UserID:    status.ILinkUserID,
	}, nil
}
