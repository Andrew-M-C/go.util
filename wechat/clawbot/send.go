package clawbot

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/md5" //nolint:gosec
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	defaultAPITimeout       = 15 * time.Second
	typingKeepaliveInterval = 5 * time.Second
	defaultCharThreshold    = 200
	defaultIdleTimeout      = 3 * time.Second
)

// uploadedFileInfo 保存 CDN 上传完成后的信息，用于构造 sendMessage 的媒体条目。
type uploadedFileInfo struct {
	downloadParam  string // CDN 响应头 x-encrypted-param
	aesKeyHex      string // 32 字符 hex 编码的 AES 密钥
	fileSizePlain  int    // 明文字节数
	fileSizeCipher int    // 密文字节数
	fileName       string // 文件基础名（如 "photo.jpg"）
}

// ===========================
//  内部 HTTP 工具
// ===========================

// apiPost 向 iLink API 发送 POST 请求，返回原始响应体。
// 所有 send/config/typing/upload API 共用此方法。
func apiPost(
	ctx context.Context, creds Credentials, endpoint string, reqBody any,
) ([]byte, error) {
	base := ensureTrailingSlash(creds.BaseURL)
	fullURL := base + endpoint

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal %s request: %w", endpoint, err)
	}

	reqCtx, cancel := context.WithTimeout(ctx, defaultAPITimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(
		reqCtx, http.MethodPost, fullURL, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("create %s request: %w", endpoint, err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("AuthorizationType", "ilink_bot_token")
	if token := strings.TrimSpace(creds.BotToken); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req.Header.Set("X-WECHAT-UIN", randomWechatUIN())
	req.Header.Set("iLink-App-ClientVersion", "1")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, fmt.Errorf("%s HTTP error: %w", endpoint, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read %s response: %w", endpoint, err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("%s HTTP %d: %s", endpoint, resp.StatusCode, body)
	}
	return body, nil
}

// generateClientID 生成一个 32 字符的随机 hex 字符串，作为消息的 client_id。
func generateClientID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// ===========================
//  方案 A: 文本 + Typing
// ===========================

// SendText 发送一条纯文本消息。
//
// 调用: POST {creds.BaseURL}/ilink/bot/sendmessage
//
// 文本为空时，会发送一条没有 item_list 的空消息（微信侧显示为空气泡）；
// 如果不希望发送空消息，调用方应自行检查。
func SendText(
	ctx context.Context, creds Credentials, target ReplyTarget, text string,
) error {
	var items []*MessageItem
	if text != "" {
		items = []*MessageItem{{
			Type:     MessageItemTypeText,
			TextItem: &TextItem{Text: text},
		}}
	}
	req := sendMessageReq{
		Msg: &WeixinMessage{
			ToUserID:     target.ToUserID,
			ClientID:     generateClientID(),
			MessageType:  MessageTypeBot,
			MessageState: MessageStateFinish,
			ItemList:     items,
			ContextToken: target.ContextToken,
		},
		BaseInfo: newBaseInfo(),
	}
	_, err := apiPost(ctx, creds, "ilink/bot/sendmessage", req)
	return err
}

// GetTypingTicket 通过 getConfig API 获取 typing 凭证。
//
// 调用: POST {creds.BaseURL}/ilink/bot/getconfig
//
// typing ticket 是一个临时令牌，用于后续调用 SendTyping。
// 通常在消息处理开始时获取一次即可。
func GetTypingTicket(
	ctx context.Context, creds Credentials, target ReplyTarget,
) (string, error) {
	req := getConfigReq{
		ILinkUserID:  target.ToUserID,
		ContextToken: target.ContextToken,
		BaseInfo:     newBaseInfo(),
	}
	body, err := apiPost(ctx, creds, "ilink/bot/getconfig", req)
	if err != nil {
		return "", err
	}
	var resp getConfigResp
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", fmt.Errorf("decode getConfig response: %w", err)
	}
	if resp.Ret != nil && *resp.Ret != 0 {
		return "", fmt.Errorf("getConfig error: ret=%d errmsg=%s",
			*resp.Ret, resp.ErrMsg)
	}
	return resp.TypingTicket, nil
}

// SendTyping 发送"正在输入…"指示器。
//
// 调用: POST {creds.BaseURL}/ilink/bot/sendtyping
//
// 需要先通过 GetTypingTicket 获取凭证。指示器会在微信侧显示几秒钟后自动消失，
// 如需持续显示应定期重发（每 5-8 秒一次）。
func SendTyping(
	ctx context.Context, creds Credentials, action TypingAction,
) error {
	return sendTypingWithStatus(
		ctx, creds, action.ToUserID, action.TypingTicket, typingStatusTyping)
}

// sendTypingWithStatus 是 SendTyping 的内部实现，支持指定 status
// （typingStatusTyping=1 开始 / typingStatusCancel=2 停止）。
func sendTypingWithStatus(
	ctx context.Context, creds Credentials,
	toUserID, typingTicket string, status int,
) error {
	req := sendTypingReq{
		ILinkUserID:  toUserID,
		TypingTicket: typingTicket,
		Status:       status,
		BaseInfo:     newBaseInfo(),
	}
	_, err := apiPost(ctx, creds, "ilink/bot/sendtyping", req)
	return err
}

// ===========================
//  方案 B: 媒体发送
// ===========================

// SendImage 发送图片消息。
//
// 内部流程: 读文件 → AES-128-ECB 加密 → getUploadUrl → POST 到 CDN → sendMessage
//
// media.CDNBaseURL 为空时使用 DefaultCDNBaseURL。
func SendImage(
	ctx context.Context, creds Credentials,
	target ReplyTarget, media MediaOpts,
) error {
	return uploadAndSendMedia(
		ctx, creds, target, media, uploadMediaTypeImage, buildImageItem)
}

// SendVideo 发送视频消息。流程同 SendImage。
func SendVideo(
	ctx context.Context, creds Credentials,
	target ReplyTarget, media MediaOpts,
) error {
	return uploadAndSendMedia(
		ctx, creds, target, media, uploadMediaTypeVideo, buildVideoItem)
}

// SendFile 发送文件附件。流程同 SendImage。
func SendFile(
	ctx context.Context, creds Credentials,
	target ReplyTarget, media MediaOpts,
) error {
	return uploadAndSendMedia(
		ctx, creds, target, media, uploadMediaTypeFile, buildFileItem)
}

// SendMediaByPath 根据文件扩展名自动判断媒体类型，路由到
// SendImage / SendVideo / SendFile。
func SendMediaByPath(
	ctx context.Context, creds Credentials,
	target ReplyTarget, media MediaOpts,
) error {
	mt := detectUploadMediaType(media.FilePath)
	switch mt {
	case uploadMediaTypeImage:
		return SendImage(ctx, creds, target, media)
	case uploadMediaTypeVideo:
		return SendVideo(ctx, creds, target, media)
	default:
		return SendFile(ctx, creds, target, media)
	}
}

// mediaItemBuilder 根据上传结果构造一个 MessageItem 指针。
type mediaItemBuilder func(info uploadedFileInfo) *MessageItem

// uploadAndSendMedia 是 SendImage/SendVideo/SendFile 的公共实现：
// 上传文件到 CDN → 构造媒体条目 → 通过 sendMessage 发送。
// 如果 media.Caption 非空，会先发一条文本消息，再发媒体消息。
func uploadAndSendMedia(
	ctx context.Context, creds Credentials,
	target ReplyTarget, media MediaOpts,
	mediaType int, buildItem mediaItemBuilder,
) error {
	cdnBase := media.CDNBaseURL
	if cdnBase == "" {
		cdnBase = DefaultCDNBaseURL
	}

	info, err := uploadFileToCDN(ctx, creds, target.ToUserID, media.FilePath, cdnBase, mediaType)
	if err != nil {
		return fmt.Errorf("upload %s: %w", filepath.Base(media.FilePath), err)
	}

	// 如果有文字说明，先发一条文本
	if media.Caption != "" {
		if err := SendText(ctx, creds, target, media.Caption); err != nil {
			return fmt.Errorf("send caption: %w", err)
		}
	}

	// 发送媒体消息
	item := buildItem(info)
	return sendSingleItem(ctx, creds, target, item)
}

// sendSingleItem 将一个 MessageItem 作为独立消息发送。
func sendSingleItem(
	ctx context.Context, creds Credentials,
	target ReplyTarget, item *MessageItem,
) error {
	req := sendMessageReq{
		Msg: &WeixinMessage{
			ToUserID:     target.ToUserID,
			ClientID:     generateClientID(),
			MessageType:  MessageTypeBot,
			MessageState: MessageStateFinish,
			ItemList:     []*MessageItem{item},
			ContextToken: target.ContextToken,
		},
		BaseInfo: newBaseInfo(),
	}
	_, err := apiPost(ctx, creds, "ilink/bot/sendmessage", req)
	return err
}

// ---------- CDN 上传管线 ----------

// uploadFileToCDN 读取本地文件 → 加密 → 获取上传 URL → POST 到 CDN。
func uploadFileToCDN(
	ctx context.Context, creds Credentials,
	toUserID, filePath, cdnBaseURL string, mediaType int,
) (uploadedFileInfo, error) {
	// 1. 读取文件
	plaintext, err := os.ReadFile(filePath)
	if err != nil {
		return uploadedFileInfo{}, fmt.Errorf("read file: %w", err)
	}

	// 2. 生成随机 filekey（16 bytes → 32 hex chars）和 AES 密钥（16 bytes）
	fileKeyBytes := make([]byte, 16)
	aesKey := make([]byte, 16)
	if _, err := rand.Read(fileKeyBytes); err != nil {
		return uploadedFileInfo{}, fmt.Errorf("generate filekey: %w", err)
	}
	if _, err := rand.Read(aesKey); err != nil {
		return uploadedFileInfo{}, fmt.Errorf("generate aes key: %w", err)
	}
	fileKey := hex.EncodeToString(fileKeyBytes)
	aesKeyHex := hex.EncodeToString(aesKey)

	// 3. 计算明文 MD5 和 AES-ECB 加密
	rawMD5 := fmt.Sprintf("%x", md5.Sum(plaintext))
	ciphertext, err := encryptAESECB(plaintext, aesKey)
	if err != nil {
		return uploadedFileInfo{}, err
	}

	// 4. 请求上传 URL
	uploadParam, err := requestUploadParam(
		ctx, creds, toUserID, fileKey, aesKeyHex, rawMD5,
		len(plaintext), len(ciphertext), mediaType)
	if err != nil {
		return uploadedFileInfo{}, err
	}

	// 5. POST 密文到 CDN
	downloadParam, err := postToCDN(cdnBaseURL, uploadParam, fileKey, ciphertext)
	if err != nil {
		return uploadedFileInfo{}, err
	}

	return uploadedFileInfo{
		downloadParam:  downloadParam,
		aesKeyHex:      aesKeyHex,
		fileSizePlain:  len(plaintext),
		fileSizeCipher: len(ciphertext),
		fileName:       filepath.Base(filePath),
	}, nil
}

// requestUploadParam 调用 getUploadUrl API 获取 CDN 上传凭证。
func requestUploadParam(
	ctx context.Context, creds Credentials,
	toUserID, fileKey, aesKeyHex, rawMD5 string,
	rawSize, cipherSize, mediaType int,
) (string, error) {
	req := getUploadURLReq{
		FileKey:     fileKey,
		MediaType:   mediaType,
		ToUserID:    toUserID,
		RawSize:     rawSize,
		RawFileMD5:  rawMD5,
		FileSize:    cipherSize,
		NoNeedThumb: true,
		AESKey:      aesKeyHex,
		BaseInfo:    newBaseInfo(),
	}
	body, err := apiPost(ctx, creds, "ilink/bot/getuploadurl", req)
	if err != nil {
		return "", err
	}
	var resp getUploadURLResp
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", fmt.Errorf("decode getUploadUrl response: %w", err)
	}
	if resp.UploadParam == "" {
		return "", fmt.Errorf("getUploadUrl: empty upload_param")
	}
	return resp.UploadParam, nil
}

// postToCDN 将加密后的文件内容 POST 到 CDN，返回 download 用的 encrypted_query_param。
func postToCDN(
	cdnBaseURL, uploadParam, fileKey string, ciphertext []byte,
) (string, error) {
	cdnURL := buildCDNUploadURL(cdnBaseURL, uploadParam, fileKey)

	resp, err := http.Post(cdnURL, "application/octet-stream",
		bytes.NewReader(ciphertext))
	if err != nil {
		return "", fmt.Errorf("CDN upload HTTP error: %w", err)
	}
	defer resp.Body.Close()
	_, _ = io.ReadAll(resp.Body) // drain body

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		errMsg := resp.Header.Get("x-error-message")
		return "", fmt.Errorf("CDN upload HTTP %d: %s", resp.StatusCode, errMsg)
	}

	downloadParam := resp.Header.Get("x-encrypted-param")
	if downloadParam == "" {
		return "", fmt.Errorf("CDN upload: missing x-encrypted-param header")
	}
	return downloadParam, nil
}

// buildCDNUploadURL 拼接 CDN 上传的完整 URL。
func buildCDNUploadURL(cdnBaseURL, uploadParam, fileKey string) string {
	return cdnBaseURL + "/upload?encrypted_query_param=" +
		url.QueryEscape(uploadParam) + "&filekey=" + url.QueryEscape(fileKey)
}

// ---------- 媒体 MessageItem 构造 ----------

// buildCDNMedia 根据上传结果构造 CDNMedia 引用。
//
// aes_key 字段的编码规则：将 AES 密钥的 hex 字符串（32 个 ASCII 字符）
// 直接作为 UTF-8 字节进行 Base64 编码——而非先 hex-decode 成 16 字节再 Base64。
func buildCDNMedia(info uploadedFileInfo) *CDNMedia {
	return &CDNMedia{
		EncryptQueryParam: info.downloadParam,
		AESKey:            base64.StdEncoding.EncodeToString([]byte(info.aesKeyHex)),
		EncryptType:       1,
	}
}

func buildImageItem(info uploadedFileInfo) *MessageItem {
	return &MessageItem{
		Type: MessageItemTypeImage,
		ImageItem: &ImageItem{
			Media:   buildCDNMedia(info),
			MidSize: info.fileSizeCipher,
		},
	}
}

func buildVideoItem(info uploadedFileInfo) *MessageItem {
	return &MessageItem{
		Type: MessageItemTypeVideo,
		VideoItem: &VideoItem{
			Media:     buildCDNMedia(info),
			VideoSize: info.fileSizeCipher,
		},
	}
}

func buildFileItem(info uploadedFileInfo) *MessageItem {
	return &MessageItem{
		Type: MessageItemTypeFile,
		FileItem: &FileItem{
			Media:    buildCDNMedia(info),
			FileName: info.fileName,
			Len:      fmt.Sprintf("%d", info.fileSizePlain),
		},
	}
}

// ---------- 扩展名 → 媒体类型映射 ----------

func detectUploadMediaType(filePath string) int {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp":
		return uploadMediaTypeImage
	case ".mp4", ".mov", ".avi", ".mkv", ".webm":
		return uploadMediaTypeVideo
	default:
		return uploadMediaTypeFile
	}
}

// ---------- AES-128-ECB 加密 ----------

// encryptAESECB 使用 AES-128-ECB 模式 + PKCS7 填充对明文加密。
//
// ECB 模式逐块独立加密（每块 16 字节），Go 标准库不直接提供 ECB，
// 因此手动循环调用 block.Encrypt。
func encryptAESECB(plaintext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("aes-ecb: new cipher: %w", err)
	}
	bs := block.BlockSize()
	padded := pkcs7Pad(plaintext, bs)
	ciphertext := make([]byte, len(padded))
	for i := 0; i < len(padded); i += bs {
		block.Encrypt(ciphertext[i:i+bs], padded[i:i+bs])
	}
	return ciphertext, nil
}

// pkcs7Pad 对数据进行 PKCS#7 填充，使总长度为 blockSize 的整数倍。
// 即使数据已对齐，也会追加一整个块的填充。
func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	pad := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, pad...)
}

// ===========================
//  方案 C: 流式发送
// ===========================

// streamSender 是 StreamSender 的内部实现。
//
// 聚合策略：
//   - WriteChunk 将文本追加到内部缓冲区
//   - 缓冲区字符数 ≥ charThreshold 时自动 flush
//   - 距上次 WriteChunk 超过 idleTimeout 未写入时自动 flush
//   - 可选: 自动维护 typing 指示器的 start / keepalive / stop
type streamSender struct {
	mu sync.Mutex

	ctx    context.Context
	creds  Credentials
	target ReplyTarget

	typingTicket  string
	buf           strings.Builder
	charThreshold int
	idleTimeout   time.Duration

	idleTimer *time.Timer // 空闲超时自动 flush
	doneCh    chan struct{}
	closed    bool
	lastErr   error
}

// NewStreamSender 创建流式发送器。
//
// 如果 opts.TypingTicket 非空，会立即发送 typing 指示器并启动
// 后台 keepalive（每 5 秒重发一次）。调用 Close 时自动停止。
//
// 使用完毕后必须调用 Close，建议 defer sender.Close()。
func NewStreamSender(ctx context.Context, opts StreamSenderOpts) StreamSender {
	charThreshold := opts.CharThreshold
	if charThreshold <= 0 {
		charThreshold = defaultCharThreshold
	}
	idleTimeout := opts.IdleTimeout
	if idleTimeout <= 0 {
		idleTimeout = defaultIdleTimeout
	}

	s := &streamSender{
		ctx:           ctx,
		creds:         opts.Creds,
		target:        opts.Target,
		typingTicket:  opts.TypingTicket,
		charThreshold: charThreshold,
		idleTimeout:   idleTimeout,
		doneCh:        make(chan struct{}),
	}

	s.startTypingLoop()
	return s
}

// WriteChunk 写入一段增量文本。
// 累积字符数达到 charThreshold 时自动触发 flush。线程安全。
func (s *streamSender) WriteChunk(text string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return
	}
	s.buf.WriteString(text)
	s.resetIdleTimerLocked()

	if s.buf.Len() >= s.charThreshold {
		_ = s.flushLocked()
	}
}

// Flush 立即将缓冲区内容作为一条消息发送。
func (s *streamSender) Flush() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.flushLocked()
}

// Close 发送剩余缓冲区内容、停止 typing 指示器、释放资源。
func (s *streamSender) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return s.lastErr
	}
	s.closed = true

	// 停止空闲定时器
	s.stopIdleTimerLocked()

	// 通知 typing keepalive 退出
	close(s.doneCh)

	// flush 残余
	flushErr := s.flushLocked()

	// 停止 typing 指示器（best-effort）
	s.stopTypingIndicator()

	if flushErr != nil {
		return flushErr
	}
	return s.lastErr
}

// flushLocked 在持有 mu 的情况下发送缓冲区内容。
func (s *streamSender) flushLocked() error {
	if s.buf.Len() == 0 {
		return s.lastErr
	}
	text := s.buf.String()
	s.buf.Reset()

	err := SendText(s.ctx, s.creds, s.target, text)
	if err != nil {
		s.lastErr = err
	}
	return err
}

// ---------- 空闲定时器 ----------

// resetIdleTimerLocked 重置空闲 flush 定时器。须在持有 mu 时调用。
func (s *streamSender) resetIdleTimerLocked() {
	if s.idleTimer != nil {
		s.idleTimer.Stop()
	}
	s.idleTimer = time.AfterFunc(s.idleTimeout, func() {
		_ = s.Flush()
	})
}

// stopIdleTimerLocked 停止空闲定时器。须在持有 mu 时调用。
func (s *streamSender) stopIdleTimerLocked() {
	if s.idleTimer != nil {
		s.idleTimer.Stop()
		s.idleTimer = nil
	}
}

// ---------- Typing 指示器生命周期 ----------

// startTypingLoop 发送初始 typing 指示器并启动后台 keepalive goroutine。
func (s *streamSender) startTypingLoop() {
	if s.typingTicket == "" {
		return
	}
	_ = sendTypingWithStatus(
		s.ctx, s.creds, s.target.ToUserID, s.typingTicket, typingStatusTyping)

	go func() {
		ticker := time.NewTicker(typingKeepaliveInterval)
		defer ticker.Stop()

		for {
			select {
			case <-s.ctx.Done():
				return
			case <-s.doneCh:
				return
			case <-ticker.C:
				_ = sendTypingWithStatus(
					s.ctx, s.creds, s.target.ToUserID,
					s.typingTicket, typingStatusTyping)
			}
		}
	}()
}

// stopTypingIndicator 发送 typing cancel（best-effort）。
// 使用独立短超时 context 以确保即使原 ctx 已取消也能发出。
func (s *streamSender) stopTypingIndicator() {
	if s.typingTicket == "" {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_ = sendTypingWithStatus(
		ctx, s.creds, s.target.ToUserID, s.typingTicket, typingStatusCancel)
}
