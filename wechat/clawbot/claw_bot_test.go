package clawbot_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Andrew-M-C/go.util/wechat/clawbot"
	qrcode "github.com/yeqown/go-qrcode/v2"
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

// pi100 是圆周率小数点后 100 位。
const pi100 = "3." +
	"14159265358979323846264338327950288419716939937510" +
	"58209749445923078164062862089986280348253421170679"

// pollFirstUserMessage 循环调用 Poll 直到收到一条用户消息（忽略 Bot 自身的消息）。
//
// 刚登录成功后服务端可能尚未完全就绪，首次 getUpdates 会返回 errcode=-14
// （形式上等同 session expired）。这不是真正的会话失效，等几秒重试即可恢复。
// 因此对 session expired 错误允许最多重试 5 次。
func pollFirstUserMessage(
	t *testing.T, ctx context.Context, creds clawbot.Credentials, buf string,
) (*clawbot.WeixinMessage, string) {
	t.Helper()
	sessionExpiredRetries := 0
	const maxSessionExpiredRetries = 5

	for {
		result, err := clawbot.Poll(ctx, creds, buf)
		if err != nil {
			if clawbot.IsSessionExpired(err) {
				sessionExpiredRetries++
				if sessionExpiredRetries > maxSessionExpiredRetries {
					t.Fatalf("连续 %d 次会话过期，放弃: %v",
						maxSessionExpiredRetries, err)
				}
				t.Logf("会话尚未就绪（第 %d/%d 次），%d 秒后重试: %v",
					sessionExpiredRetries, maxSessionExpiredRetries,
					sessionExpiredRetries*3, err)
				time.Sleep(time.Duration(sessionExpiredRetries*3) * time.Second)
				continue
			}
			t.Logf("poll 出错（将重试）: %v", err)
			time.Sleep(2 * time.Second)
			continue
		}
		sessionExpiredRetries = 0
		buf = result.GetUpdatesBuf
		for _, msg := range result.Messages {
			if msg.MessageType == clawbot.MessageTypeUser {
				return msg, buf
			}
		}
	}
}

func TestClawBotSequential(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// -------- 步骤 1: 获取二维码 --------
	t.Log("===== 步骤 1: 获取登录二维码 =====")
	qr, err := clawbot.FetchQRCode(ctx, "")
	if err != nil {
		t.Fatalf("获取二维码失败: %v", err)
	}
	t.Logf("二维码 URL: %s", qr.QRCodeImgContent)
	if qr.QRCodeImgContent == "" {
		t.Errorf("二维码 URL 为空")
		return
	}
	printQRCodeToConsole(t, qr.QRCodeImgContent)

	t.Logf("请在 30 秒内打开上述 URL 并扫码")

	// -------- 步骤 2: 等待扫码登录 --------
	t.Log("===== 步骤 2: 等待扫码登录 =====")
	creds, err := clawbot.WaitForLogin(ctx, "", qr, clawbot.LoginCallbacks{
		OnScanned: func() {
			t.Log("已扫码，等待确认...")
		},
		OnQRRefreshed: func(newURL string) {
			t.Logf("二维码已刷新: %s", newURL)
		},
	})
	if err != nil {
		t.Fatalf("登录失败: %v", err)
	}
	t.Logf("登录成功！微信返回参数:")
	t.Logf("  BotToken: %s", creds.BotToken[:min(len(creds.BotToken), 20)]+"...")
	t.Logf("  BaseURL:  %s", creds.BaseURL)
	t.Logf("  BotID:    %s", creds.BotID)
	t.Logf("  UserID:   %s", creds.UserID)

	buf := ""

	// -------- 步骤 2.5: 主动发送（无 contextToken）--------
	t.Log("===== 步骤 2.5: 主动发送消息（无 contextToken）=====")
	proactiveTarget := clawbot.ReplyTarget{
		ToUserID: creds.UserID, // 发给扫码登录的用户自己
	}
	if err := clawbot.SendText(ctx, creds, proactiveTarget, "Andrew at your service!"); err != nil {
		t.Fatalf("主动发送失败: %v", err)
	}
	t.Log("主动发送成功！但是按照目前微信的设计来看, 这是无法发送到的, " +
		"似乎微信的设计是在用户发出一条消息后的 24 小时内, 最多允许主动发送 10 条消息, 包括回复的消息本身")

	// -------- 步骤 3: 等待第一条消息 --------
	t.Log("===== 步骤 3: 等待接收消息 =====")
	t.Log("现在请你随意输入一个消息")

	msg1, buf := pollFirstUserMessage(t, ctx, creds, buf)
	t.Logf("收到消息: from=%s text=%q", msg1.FromUserID, msg1.GetTextBody())

	// -------- 步骤 4: 回复 "歪比巴卜" + HH:MM:SS --------
	t.Log("===== 步骤 4: 回复歪比巴卜 =====")
	target := clawbot.ReplyTarget{
		ToUserID:     msg1.FromUserID,
		ContextToken: msg1.ContextToken,
	}
	reply := fmt.Sprintf("歪比巴卜 %s", time.Now().Format("15:04:05"))
	if err := clawbot.SendText(ctx, creds, target, reply); err != nil {
		t.Fatalf("SendText 失败: %v", err)
	}
	t.Logf("已回复: %s", reply)

	// -------- 步骤 5: 等待第二条消息 --------
	t.Log("===== 步骤 5: 等待接收第二条消息 =====")
	t.Log("现在请你再随意输入一个消息")

	msg2, buf := pollFirstUserMessage(t, ctx, creds, buf)
	t.Logf("收到消息: from=%s text=%q", msg2.FromUserID, msg2.GetTextBody())

	// -------- 步骤 6: 流式逐字输出圆周率 --------
	t.Log("===== 步骤 6: 流式输出圆周率（小数点后 100 位）=====")
	target2 := clawbot.ReplyTarget{
		ToUserID:     msg2.FromUserID,
		ContextToken: msg2.ContextToken,
	}

	ticket, err := clawbot.GetTypingTicket(ctx, creds, target2)
	if err != nil {
		t.Logf("获取 typingTicket 失败（非致命）: %v", err)
	} else {
		t.Logf("获取 typingTicket 成功: %s", ticket[:min(len(ticket), 20)]+"...")
		if err := clawbot.SendTyping(ctx, creds, clawbot.TypingAction{
			ToUserID: target2.ToUserID, TypingTicket: ticket,
		}); err != nil {
			t.Logf("SendTyping 失败（非致命）: %v", err)
		} else {
			t.Log("已发送 typing 指示器")
		}
	}

	sender := clawbot.NewStreamSender(ctx, clawbot.StreamSenderOpts{
		Creds:  creds,
		Target: target2,
		// CharThreshold: 999, // 禁止自动 flush，完全手动控制
		IdleTimeout: time.Hour,
	})
	defer sender.Close()

	for _, ch := range pi100 {
		sender.WriteChunk(string(ch))
		// if err := sender.Flush(); err != nil {
		// 	t.Fatalf("Flush 第 %d 个字符失败: %v", i, err)
		// }
		time.Sleep(100 * time.Millisecond)
	}

	if err := sender.Close(); err != nil {
		t.Fatalf("StreamSender Close 失败: %v", err)
	}
	t.Log("StreamSender 圆周率输出完毕！")

	// -------- 步骤 7: 用 SendText 逐字输出圆周率 --------
	t.Log("===== 步骤 7: 等待接收第三条消息 =====")
	t.Log("现在请你再随意输入一个消息")

	msg3, _ := pollFirstUserMessage(t, ctx, creds, buf)
	t.Logf("收到消息: from=%s text=%q", msg3.FromUserID, msg3.GetTextBody())

	t.Log("===== 步骤 7: 使用 SendText 逐字输出圆周率 =====")
	target3 := clawbot.ReplyTarget{
		ToUserID:     msg3.FromUserID,
		ContextToken: msg3.ContextToken,
	}
	for i, ch := range pi100 {
		if err := clawbot.SendText(ctx, creds, target3, string(ch)); err != nil {
			t.Fatalf("SendText 第 %d 个字符失败: %v", i, err)
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Log("SendText 圆周率输出完毕！测试通过。")
}

// printQRCodeToConsole 将 content 编码为二维码并打印到控制台。
// 深色模块用 █，浅色模块用全角空格　，最外围包一圈 █ 边框。
func printQRCodeToConsole(t *testing.T, content string) {
	t.Helper()

	q, err := qrcode.NewWith(
		content,
		qrcode.WithErrorCorrectionLevel(qrcode.ErrorCorrectionHighest),
	)
	if err != nil {
		t.Logf("生成二维码失败（ASCII-art 展示跳过）: %v", err)
		return
	}

	if err := q.Save(&qrTermWriter{}); err != nil {
		t.Logf("向终端输出二维码失败（ASCII-art 展示跳过）: %v", err)
	}
}

// qrTermWriter 实现 qrcode.Writer，将 QR 矩阵以字符画形式输出到 stdout。
type qrTermWriter struct{}

func (qrTermWriter) Write(mat qrcode.Matrix) error {
	const dark = "⬛"
	const light = "⬜"
	const border = light

	w := mat.Width()
	h := mat.Height()

	// 逐行收集格子状态：IsSet()==true 为深色模块。
	rows := make([][]bool, h)
	for i := range rows {
		rows[i] = make([]bool, w)
	}
	mat.Iterate(qrcode.IterDirection_ROW, func(x, y int, v qrcode.QRValue) {
		rows[y][x] = v.IsSet()
	})

	// 上边框：宽度 = 内容列数 + 左右各 1 格边框
	borderWidth := w + 2
	borderLine := strings.Repeat(border, borderWidth)
	fmt.Println(borderLine)

	for _, row := range rows {
		var sb strings.Builder
		sb.WriteString(border)
		for _, cell := range row {
			if cell {
				sb.WriteString(dark)
			} else {
				sb.WriteString(light)
			}
		}
		sb.WriteString(border)
		fmt.Println(sb.String())
	}

	// 下边框
	fmt.Println(borderLine)
	return nil
}

func (qrTermWriter) Close() error { return nil }
