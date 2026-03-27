package clawbot

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"
)

// ===========================
//  消息类型与状态常量
// ===========================

// MessageType 表示消息的发送者类型，对应 WeixinMessage.MessageType。
type MessageType int

const (
	MessageTypeUser MessageType = 1 // 用户发送的消息
	MessageTypeBot  MessageType = 2 // Bot 发送的消息
)

// MessageItemType 表示消息内容项的类型，对应 MessageItem.Type。
type MessageItemType int

const (
	MessageItemTypeText  MessageItemType = 1 // 纯文本
	MessageItemTypeImage MessageItemType = 2 // 图片
	MessageItemTypeVoice MessageItemType = 3 // 语音
	MessageItemTypeFile  MessageItemType = 4 // 文件附件
	MessageItemTypeVideo MessageItemType = 5 // 视频
)

// MessageState 表示消息的生成状态，对应 WeixinMessage.MessageState。
type MessageState int

const (
	MessageStateNew        MessageState = 0 // 新建（未开始生成）
	MessageStateGenerating MessageState = 1 // 正在生成中（流式场景）
	MessageStateFinish     MessageState = 2 // 生成完毕
)

// DefaultBaseURL 是 iLink API 的默认网关地址。
// 首次登录时使用此地址；登录成功后微信可能返回不同的 BaseURL（如区域节点），
// 后续 API 调用和重新登录应优先使用 Credentials.BaseURL。
const DefaultBaseURL = "https://ilinkai.weixin.qq.com"

// DefaultCDNBaseURL 是微信 CDN 的默认地址，用于媒体文件下载。
const DefaultCDNBaseURL = "https://novac2c.cdn.weixin.qq.com/c2c"

// sessionExpiredErrCode 是 iLink API 返回"会话过期"时的 errcode 值。
// 当 getUpdates 返回此错误码时，Bot 应暂停请求并考虑重新登录。
const sessionExpiredErrCode = -14

// channelVersion 标识客户端版本，附加在每个 API 请求的 base_info 中。
const channelVersion = "2.0.1-go"

// ===========================
//  凭据
// ===========================

// Credentials 是 QR 登录成功后从微信获得的 Bot 凭据。
//
// 四个字段全部来自微信 get_qrcode_status API 的响应，不是调用方生成的。
// 调用方拿到后自行决定如何持久化（文件、数据库、环境变量等）。
type Credentials struct {
	// BotToken 是 Bearer 鉴权令牌，用于所有后续 API 请求的 Authorization 头。
	// 来自微信返回的 bot_token 字段。
	BotToken string `json:"bot_token"`

	// BaseURL 是 API 根地址（如 "https://ilinkai.weixin.qq.com"）。
	// 微信在登录确认时可能返回一个与登录发起时不同的 BaseURL，以此为准。
	BaseURL string `json:"base_url"`

	// BotID 是 Bot 的唯一标识，来自微信返回的 ilink_bot_id 字段。
	// 已经过归一化处理：@ → -, . → -（如 "a1b2@im.bot" → "a1b2-im-bot"）。
	BotID string `json:"bot_id"`

	// UserID 是扫码微信用户的 ID，来自微信返回的 ilink_user_id 字段。
	// 格式如 "wxid_abc@im.wechat"，同一微信用户始终不变。
	UserID string `json:"user_id"`
}

// ===========================
//  登录相关类型
// ===========================

// QRCodeResult 是 FetchQRCode 的返回结果。
type QRCodeResult struct {
	// QRCode 是二维码的唯一 key，用于后续调用 WaitForLogin 轮询扫码状态。
	// 此值不需要展示给用户，仅作为内部标识。
	QRCode string `json:"qrcode"`

	// QRCodeImgContent 是二维码图片的 URL，供展示给用户扫描。
	// 来自微信返回的 qrcode_img_content 字段。
	QRCodeImgContent string `json:"qrcode_img_content"`
}

// LoginCallbacks 是登录等待过程中的可选事件回调。
// 所有字段均可为 nil；传 nil 的 LoginCallbacks 指针也是安全的。
type LoginCallbacks struct {
	// OnScanned 在用户已扫码但尚未在微信端点击确认时触发。
	OnScanned func()

	// OnQRRefreshed 在二维码过期后自动刷新成功时触发，参数为新的二维码图片 URL。
	OnQRRefreshed func(newQRCodeURL string)
}

// ===========================
//  回复相关类型
// ===========================

// ReplyTarget 标识一次回复的目标：回复给谁、在哪个对话上下文中。
//
// 通常直接从入站消息构造：
//
//	target := utils.ReplyTarget{
//	    ToUserID:     msg.FromUserID,
//	    ContextToken: msg.ContextToken,
//	}
type ReplyTarget struct {
	// ToUserID 是回复目标用户的 ID（即入站消息的 FromUserID）。
	ToUserID string `json:"to_user_id"`

	// ContextToken 是对话上下文令牌（即入站消息的 ContextToken）。
	// 回复时原样传回，微信用它关联上下文。
	ContextToken string `json:"context_token,omitempty"`
}

// TypingAction 是 SendTyping 的参数，包含目标用户和 typing 凭证。
//
// TypingTicket 通过 GetTypingTicket 获取。
type TypingAction struct {
	// ToUserID 是 typing 指示器的目标用户 ID。
	ToUserID string `json:"to_user_id"`

	// TypingTicket 是通过 GetTypingTicket 获取的临时凭证。
	TypingTicket string `json:"typing_ticket"`
}

// ===========================
//  流式发送相关类型
// ===========================

// StreamSender 将 AI 的增量输出聚合为块，自动分次发送到微信。
//
// 微信的每次 sendMessage 产生一个独立气泡，StreamSender 将连续的小块文本
// 聚合到一定长度后再发送，避免产生过多碎片气泡。
//
// 通过 NewStreamSender 获取实例；使用完毕后必须调用 Close。
type StreamSender interface {
	// WriteChunk 写入一段增量文本（如 AI SSE 返回的一个 token）。
	// 累积文本超过 CharThreshold 时自动触发 Flush。线程安全。
	WriteChunk(text string)

	// Flush 立即将当前缓冲区的文本作为一条消息发送。
	Flush() error

	// Close 发送剩余缓冲区内容，停止 typing 指示器，释放资源。
	// 必须调用，建议 defer sender.Close()。
	Close() error
}

// StreamSenderOpts 是 NewStreamSender 的配置。
type StreamSenderOpts struct {
	Creds  Credentials `json:"creds"`
	Target ReplyTarget `json:"target"`

	// TypingTicket 可选；非空时自动维护 typing 指示器的 start/keepalive/stop。
	TypingTicket string `json:"typing_ticket,omitempty"`

	// CharThreshold 是缓冲区字符数阈值，达到后自动 flush。默认 200。
	CharThreshold int `json:"char_threshold,omitempty"`

	// IdleTimeout 是最后一次 WriteChunk 后多久自动 flush。默认 3s。
	IdleTimeout time.Duration `json:"idle_timeout,omitempty"`
}

// ===========================
//  媒体发送相关类型
// ===========================

// MediaOpts 是 SendImage / SendVideo / SendFile / SendMediaByPath 的公共媒体参数。
type MediaOpts struct {
	// FilePath 是本地文件路径（如 "/tmp/photo.jpg"）。
	FilePath string `json:"file_path"`

	// Caption 是随媒体附带的文字说明。为空则不附带文字。
	Caption string `json:"caption,omitempty"`

	// CDNBaseURL 是 CDN 上传地址。为空时自动使用 DefaultCDNBaseURL。
	CDNBaseURL string `json:"cdn_base_url,omitempty"`
}

// ===========================
//  轮询相关类型
// ===========================

// PollResult 是 Poll 函数的返回结果。
type PollResult struct {
	// Messages 是本次轮询收到的消息列表。
	// 可能为空切片——表示本次长轮询超时，没有新消息。
	Messages []*WeixinMessage `json:"messages,omitempty"`

	// GetUpdatesBuf 是下一次调用 Poll 时应传入的游标（翻页令牌）。
	// 首次调用 Poll 时传空字符串 ""，之后始终使用上一次 PollResult 返回的值。
	GetUpdatesBuf string `json:"get_updates_buf"`
}

// ===========================
//  错误类型
// ===========================

// APIError 表示 iLink API 返回的业务层错误。
// HTTP 状态码为 200 但响应体中的 ret 或 errcode 非零时会产生此错误。
type APIError struct {
	Ret     int    // 响应中的 ret 字段
	ErrCode int    // 响应中的 errcode 字段
	ErrMsg  string // 响应中的 errmsg 字段
}

func (e *APIError) Error() string {
	return fmt.Sprintf("iLink API error: ret=%d errcode=%d errmsg=%s",
		e.Ret, e.ErrCode, e.ErrMsg)
}

// ErrSessionExpired 表示微信会话已过期（errcode=-14）。
//
// 收到此错误后，调用方应暂停所有对该账号的 API 请求（建议至少 1 小时），
// 并考虑重新执行 QR 登录流程。
//
// 判断方式: errors.Is(err, ErrSessionExpired)
var ErrSessionExpired = errors.New("weixin session expired (errcode -14)")

// ===========================
//  消息协议类型
// ===========================

// WeixinMessage 是 getUpdates 返回的单条微信消息。
// 同一结构体也用于 sendMessage 的请求体（出站消息）。
//
// 入站消息中:
//   - FromUserID = 发消息的微信用户
//   - ToUserID   = 接收消息的 Bot（即 Credentials.BotID 对应的原始值）
//   - ContextToken 应缓存下来，回复时原样传回
//
// 出站消息中:
//   - FromUserID 留空
//   - ToUserID   = 要回复的目标用户（即入站消息的 FromUserID）
type WeixinMessage struct {
	Seq          int64          `json:"seq,omitempty"`            // 消息序号
	MessageID    int64          `json:"message_id,omitempty"`     // 消息唯一 ID
	FromUserID   string         `json:"from_user_id,omitempty"`   // 发送者 ID
	ToUserID     string         `json:"to_user_id,omitempty"`     // 接收者 ID
	ClientID     string         `json:"client_id,omitempty"`      // 客户端生成的消息 ID（出站时使用）
	CreateTimeMs int64          `json:"create_time_ms,omitempty"` // 创建时间（毫秒时间戳）
	UpdateTimeMs int64          `json:"update_time_ms,omitempty"` // 更新时间
	DeleteTimeMs int64          `json:"delete_time_ms,omitempty"` // 删除时间
	SessionID    string         `json:"session_id,omitempty"`     // 会话 ID
	GroupID      string         `json:"group_id,omitempty"`       // 群组 ID
	MessageType  MessageType    `json:"message_type,omitempty"`   // MessageTypeUser / MessageTypeBot
	MessageState MessageState   `json:"message_state,omitempty"`  // MessageState* 常量
	ItemList     []*MessageItem `json:"item_list,omitempty"`      // 消息内容项列表
	ContextToken string         `json:"context_token,omitempty"`  // 对话上下文令牌
}

// GetTextBody 从消息的 ItemList 中提取第一个纯文本内容。
// 如果没有文本类型的 item，返回空字符串。
func (m *WeixinMessage) GetTextBody() string {
	for _, item := range m.ItemList {
		if item.Type == MessageItemTypeText && item.TextItem != nil {
			return item.TextItem.Text
		}
	}
	return ""
}

// HasMedia 判断消息是否包含媒体内容（图片、语音、文件、视频）。
func (m *WeixinMessage) HasMedia() bool {
	for _, item := range m.ItemList {
		switch item.Type {
		case MessageItemTypeImage, MessageItemTypeVoice,
			MessageItemTypeFile, MessageItemTypeVideo:
			return true
		}
	}
	return false
}

// MessageItem 是消息中的一个内容项。
// Type 字段决定了哪个 *Item 子字段有值（互斥）。
type MessageItem struct {
	Type         MessageItemType `json:"type,omitempty"` // MessageItemType* 常量
	CreateTimeMs int64           `json:"create_time_ms,omitempty"`
	UpdateTimeMs int64           `json:"update_time_ms,omitempty"`
	IsCompleted  bool            `json:"is_completed,omitempty"`
	MsgID        string          `json:"msg_id,omitempty"`
	RefMsg       *RefMessage     `json:"ref_msg,omitempty"`    // 引用/回复的原消息（如有）
	TextItem     *TextItem       `json:"text_item,omitempty"`  // Type=1 时有值
	ImageItem    *ImageItem      `json:"image_item,omitempty"` // Type=2 时有值
	VoiceItem    *VoiceItem      `json:"voice_item,omitempty"` // Type=3 时有值
	FileItem     *FileItem       `json:"file_item,omitempty"`  // Type=4 时有值
	VideoItem    *VideoItem      `json:"video_item,omitempty"` // Type=5 时有值
}

// TextItem 是纯文本消息内容。
type TextItem struct {
	Text string `json:"text,omitempty"`
}

// CDNMedia 是 CDN 上的加密媒体文件引用。
// 下载后需要用 AESKey（base64 解码后）进行 AES-128-ECB 解密。
type CDNMedia struct {
	// EncryptQueryParam 是 CDN 下载 URL 的加密路径参数。
	EncryptQueryParam string `json:"encrypt_query_param,omitempty"`
	// AESKey 是 base64 编码的 AES-128 解密密钥（16 字节明文）。
	AESKey string `json:"aes_key,omitempty"`
	// EncryptType 加密类型，通常为 1。
	EncryptType int `json:"encrypt_type,omitempty"`
}

// ImageItem 是图片消息内容。
type ImageItem struct {
	Media       *CDNMedia `json:"media,omitempty"`
	ThumbMedia  *CDNMedia `json:"thumb_media,omitempty"`
	AESKey      string    `json:"aeskey,omitempty"`
	URL         string    `json:"url,omitempty"`
	MidSize     int       `json:"mid_size,omitempty"`
	ThumbSize   int       `json:"thumb_size,omitempty"`
	ThumbHeight int       `json:"thumb_height,omitempty"`
	ThumbWidth  int       `json:"thumb_width,omitempty"`
	HDSize      int       `json:"hd_size,omitempty"`
}

// VoiceItem 是语音消息内容。
type VoiceItem struct {
	Media         *CDNMedia `json:"media,omitempty"`
	EncodeType    int       `json:"encode_type,omitempty"`
	BitsPerSample int       `json:"bits_per_sample,omitempty"`
	SampleRate    int       `json:"sample_rate,omitempty"`
	PlayTime      int       `json:"playtime,omitempty"`
	Text          string    `json:"text,omitempty"` // 语音转文字结果（如有）
}

// FileItem 是文件附件消息内容。
type FileItem struct {
	Media    *CDNMedia `json:"media,omitempty"`
	FileName string    `json:"file_name,omitempty"`
	MD5      string    `json:"md5,omitempty"`
	Len      string    `json:"len,omitempty"` // 文件大小（字符串形式的字节数）
}

// VideoItem 是视频消息内容。
type VideoItem struct {
	Media       *CDNMedia `json:"media,omitempty"`
	VideoSize   int       `json:"video_size,omitempty"`
	PlayLength  int       `json:"play_length,omitempty"`
	VideoMD5    string    `json:"video_md5,omitempty"`
	ThumbMedia  *CDNMedia `json:"thumb_media,omitempty"`
	ThumbSize   int       `json:"thumb_size,omitempty"`
	ThumbHeight int       `json:"thumb_height,omitempty"`
	ThumbWidth  int       `json:"thumb_width,omitempty"`
}

// RefMessage 是被引用/回复的原消息。
type RefMessage struct {
	MessageItem *MessageItem `json:"message_item,omitempty"`
	Title       string       `json:"title,omitempty"`
}

// ===========================
//  内部 API 类型
//  仅供 login.go / poll.go / send.go 使用，不对外暴露。
// ===========================

// baseInfo 是所有 iLink API 请求中的公共元数据字段。
type baseInfo struct {
	ChannelVersion string `json:"channel_version,omitempty"`
}

// qrCodeResponse 是 GET /ilink/bot/get_bot_qrcode 的响应。
type qrCodeResponse struct {
	QRCode           string `json:"qrcode"`
	QRCodeImgContent string `json:"qrcode_img_content"`
}

// statusResponse 是 GET /ilink/bot/get_qrcode_status 的响应。
type statusResponse struct {
	Status      string `json:"status"`                 // "wait" / "scaned" / "expired" / "confirmed"
	BotToken    string `json:"bot_token,omitempty"`    // 仅 confirmed 时有值
	ILinkBotID  string `json:"ilink_bot_id,omitempty"` // 仅 confirmed 时有值
	BaseURL     string `json:"baseurl,omitempty"`      // 仅 confirmed 时有值
	ILinkUserID string `json:"ilink_user_id,omitempty"`
}

// getUpdatesReq 是 POST /ilink/bot/getupdates 的请求体。
type getUpdatesReq struct {
	GetUpdatesBuf string    `json:"get_updates_buf"`
	BaseInfo      *baseInfo `json:"base_info,omitempty"`
}

// getUpdatesResp 是 POST /ilink/bot/getupdates 的响应体。
type getUpdatesResp struct {
	Ret                  *int             `json:"ret,omitempty"`
	ErrCode              *int             `json:"errcode,omitempty"`
	ErrMsg               string           `json:"errmsg,omitempty"`
	Msgs                 []*WeixinMessage `json:"msgs,omitempty"`
	GetUpdatesBuf        string           `json:"get_updates_buf,omitempty"`
	LongPollingTimeoutMs *int             `json:"longpolling_timeout_ms,omitempty"`
}

// sendMessageReq 是 POST /ilink/bot/sendmessage 的请求体。
type sendMessageReq struct {
	Msg      *WeixinMessage `json:"msg,omitempty"`
	BaseInfo *baseInfo      `json:"base_info,omitempty"`
}

// getConfigReq 是 POST /ilink/bot/getconfig 的请求体。
type getConfigReq struct {
	ILinkUserID  string    `json:"ilink_user_id,omitempty"`
	ContextToken string    `json:"context_token,omitempty"`
	BaseInfo     *baseInfo `json:"base_info,omitempty"`
}

// getConfigResp 是 POST /ilink/bot/getconfig 的响应体。
type getConfigResp struct {
	Ret          *int   `json:"ret,omitempty"`
	ErrMsg       string `json:"errmsg,omitempty"`
	TypingTicket string `json:"typing_ticket,omitempty"`
}

// sendTypingReq 是 POST /ilink/bot/sendtyping 的请求体。
type sendTypingReq struct {
	ILinkUserID  string    `json:"ilink_user_id,omitempty"`
	TypingTicket string    `json:"typing_ticket,omitempty"`
	Status       int       `json:"status,omitempty"` // typingStatusTyping=1 / typingStatusCancel=2
	BaseInfo     *baseInfo `json:"base_info,omitempty"`
}

// typing 状态常量。
const (
	typingStatusTyping = 1 // 显示"对方正在输入..."
	typingStatusCancel = 2 // 停止显示
)

// getUploadURLReq 是 POST /ilink/bot/getuploadurl 的请求体。
type getUploadURLReq struct {
	FileKey     string    `json:"filekey,omitempty"`
	MediaType   int       `json:"media_type,omitempty"`
	ToUserID    string    `json:"to_user_id,omitempty"`
	RawSize     int       `json:"rawsize,omitempty"`
	RawFileMD5  string    `json:"rawfilemd5,omitempty"`
	FileSize    int       `json:"filesize,omitempty"`
	NoNeedThumb bool      `json:"no_need_thumb,omitempty"`
	AESKey      string    `json:"aeskey,omitempty"`
	BaseInfo    *baseInfo `json:"base_info,omitempty"`
}

// getUploadURLResp 是 POST /ilink/bot/getuploadurl 的响应体。
type getUploadURLResp struct {
	UploadParam string `json:"upload_param,omitempty"`
}

// 上传媒体类型常量，对应 getUploadURLReq.MediaType。
const (
	uploadMediaTypeImage = 1
	uploadMediaTypeVideo = 2
	uploadMediaTypeFile  = 3
)

// ===========================
//  内部工具函数
// ===========================

// ensureTrailingSlash 确保 URL 以 "/" 结尾。
func ensureTrailingSlash(u string) string {
	if strings.HasSuffix(u, "/") {
		return u
	}
	return u + "/"
}

// randomWechatUIN 生成一个随机的 9 位数字字符串，用于 X-WECHAT-UIN 请求头。
// 微信 API 要求此头部存在，但不校验其具体值。
func randomWechatUIN() string {
	n, err := rand.Int(rand.Reader, big.NewInt(900_000_000))
	if err != nil {
		return "100000000"
	}
	return fmt.Sprintf("%d", n.Int64()+100_000_000)
}

// newBaseInfo 构造 API 请求中的 base_info 字段。
func newBaseInfo() *baseInfo {
	return &baseInfo{ChannelVersion: channelVersion}
}

// normalizeBotID 将微信原始 ilink_bot_id（如 "hex@im.bot"）
// 转换为文件系统和 URL 安全的格式（如 "hex-im-bot"）。
func normalizeBotID(raw string) string {
	s := strings.TrimSpace(raw)
	s = strings.ReplaceAll(s, "@", "-")
	s = strings.ReplaceAll(s, ".", "-")
	return s
}
