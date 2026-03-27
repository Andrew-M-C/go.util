# clawbot — 微信 iLink Bot 工具库设计

- [clawbot — 微信 iLink Bot 工具库设计](#clawbot--微信-ilink-bot-工具库设计)
  - [定位](#定位)
  - [核心概念](#核心概念)
    - [`botID` 与 `userID` — 两个身份标识](#botid-与-userid--两个身份标识)
    - [`contextToken` — 对话上下文令牌](#contexttoken--对话上下文令牌)
    - [`typingTicket` — 打字指示器凭证](#typingticket--打字指示器凭证)
    - [概念关系总览](#概念关系总览)
  - [1. `entity.go` — 数据结构与常量](#1-entitygo--数据结构与常量)
  - [2. `login.go` — QR 扫码登录](#2-logingo--qr-扫码登录)
  - [3. `poll.go` — 消息轮询](#3-pollgo--消息轮询)
  - [4. `send.go` — 消息回复](#4-sendgo--消息回复)


## 定位

`clawbot` 是一个**无状态、零持久化**的工具库，封装微信 iLink Bot API 的核心交互操作。调用方自行负责凭据存储、消息路由、AI 接入等上层逻辑。

---

## 核心概念

在阅读后续 API 设计之前，需要先理解 iLink 协议中的几个关键概念。

### `botID` 与 `userID` — 两个身份标识

iLink 协议中有两个身份：**Bot 自己**和**跟 Bot 聊天的微信用户**。

| | `botID` | `userID` |
|---|---|---|
| 代表谁 | Bot 自身 | 微信用户 |
| 来源 | 登录成功后微信返回的 `ilink_bot_id` | 登录时微信返回的 `ilink_user_id`，以及每条入站消息的 `from_user_id` |
| 典型格式 | `a1b2c3d4@im.bot` → 归一化为 `a1b2c3d4-im-bot` | `wxid_abc@im.wechat` |
| 生命周期 | 每次 QR 登录微信分配一个新的，重新登录会变 | 固定不变，同一微信用户始终相同 |
| 在消息中的角色 | 入站消息的 `to_user_id` | 入站消息的 `from_user_id` |

**重要澄清：`botID` 是微信返回的，不是 Bot 自行生成的。** 四个凭据字段（BotToken、BaseURL、BotID、UserID）全部来自 QR 登录成功后 `get_qrcode_status` API 的响应：

```text
GET /ilink/bot/get_qrcode_status?qrcode=xxx

← 200 OK
{
  "status":        "confirmed",
  "bot_token":     "eyJhbG...",                      → Credentials.BotToken
  "ilink_bot_id":  "a1b2c3d4@im.bot",               → Credentials.BotID
  "baseurl":       "https://ilinkai.weixin.qq.com",  → Credentials.BaseURL
  "ilink_user_id": "wxid_abc@im.wechat"              → Credentials.UserID
}
```

在消息收发中的对应关系：

```text
用户给 Bot 发消息（入站）:
  { "from_user_id": "wxid_abc@im.wechat",  ← userID（用户）
    "to_user_id":   "a1b2c3d4@im.bot" }    ← botID（Bot）

Bot 回复用户（出站）:
  { "from_user_id": "",                     ← Bot 发送时留空
    "to_user_id":   "wxid_abc@im.wechat" }  ← userID（用户）
```

对 clawbot 包的影响：`botID` 在登录时获得一次，已包含在 `Credentials` 中，后续不需要单独传递。调用方真正高频使用的是每条入站消息的 `FromUserID`——它就是回复目标。

### `contextToken` — 对话上下文令牌

`contextToken` 是微信用来关联一组对话消息上下文的令牌（类似 session ID）。

**数据流：**

```text
用户给 Bot 发消息
  → getUpdates 返回的 WeixinMessage 中携带 context_token
  → Bot 缓存它（按 userID 索引）
  → Bot 回复时在 sendMessage 请求体中原样传回
```

**关键特性：**

- **强烈建议传递**——不传也能发消息，但微信侧可能无法正确关联上下文（引用回复等功能会失效）
- **会变化**——每条入站消息可能携带新的 `context_token`，应始终使用最新的
- 被列为敏感字段，日志中应脱敏处理

**对 clawbot 包的影响：** `Poll()` 返回的 `WeixinMessage` 中已包含 `ContextToken`。`response.go` 的发送函数需要接受它作为参数。调用方负责维护"哪个 userID 对应哪个最新 contextToken"的映射。

### `typingTicket` — 打字指示器凭证

`typingTicket` 是调用 `sendTyping` API 所需的临时授权凭证。微信用它确认你和该用户有活跃对话。

**获取方式：**

```text
POST /ilink/bot/getconfig
{ "ilink_user_id": "wxid_abc@im.wechat", "context_token": "CKsB..." }
→ { "ret": 0, "typing_ticket": "dGlja2V0..." }
```

**用途：** 拿到 `typingTicket` 后，可调用 `sendTyping` 让用户看到"对方正在输入..."提示：

```text
POST /ilink/bot/sendtyping
{ "ilink_user_id": "wxid_abc@im.wechat", "typing_ticket": "dGlja2V0...", "status": 1 }

status=1 → 开始显示"对方正在输入..."
status=2 → 停止显示
```

**关键特性：**

- **完全可选**——不获取 ticket、不发 typing，消息收发完全不受影响
- 有效期有限，生产环境中通常做缓存 + 定期刷新
- 仅影响 UX：对于回复很快（<2s）的场景，加不加 typing 区别不大；但 AI 生成需要 5-10 秒时，typing 能让用户知道 Bot 没有卡死

**对 clawbot 包的影响：** 如果选择方案 A/B，`typingTicket` 可以完全忽略，或提供一个独立的 `SendTyping` 函数让调用方手动调用。如果选择方案 C/D（流式），`StreamSender` 内部会自动管理 typing 的 start/keepalive/stop 生命周期。

### 概念关系总览

```text
QR 登录成功
  → 微信返回 Credentials {BotToken, BaseURL, BotID, UserID}
                                          ↓
收到用户消息（Poll）                       ↓
  → WeixinMessage                         ↓
     ├── FromUserID ← 谁发的（= userID）  ↓
     ├── ToUserID   ← 发给谁（= botID）
     └── ContextToken ← 对话上下文令牌

回复用户（SendText 等）
  ├── target = ReplyTarget{
  │       ToUserID:     msg.FromUserID,      ← 回复目标
  │       ContextToken: msg.ContextToken,    ← 原样回传
  │   }
  └── creds.BotToken                         ← HTTP Authorization 头

（可选）发送 typing 指示器
  ├── ticket, _ := GetTypingTicket(creds, target)
  └── SendTyping(creds, TypingAction{target.ToUserID, ticket})
```

---

## 1. `entity.go` — 数据结构与常量

定义 `clawbot` 包所有公开和内部使用的数据结构，包括：

- **常量**：`DefaultBaseURL` / `DefaultCDNBaseURL`，以及 `MessageType`、`MessageItemType`、`MessageState` 三组枚举类型
- **凭据**：`Credentials`（BotToken / BaseURL / BotID / UserID）
- **登录类型**：`QRCodeResult`、`LoginCallbacks`
- **轮询类型**：`PollResult`
- **消息协议**：`WeixinMessage` → `MessageItem` → `TextItem` / `ImageItem` / `VoiceItem` / `FileItem` / `VideoItem`，以及 `CDNMedia`、`RefMessage`
- **错误类型**：`APIError`、`ErrSessionExpired`
- **便捷方法**：`WeixinMessage.GetTextBody()`、`WeixinMessage.HasMedia()`

详见 [`entity.go`](entity.go)，每个类型和字段均有注释。

---

## 2. `login.go` — QR 扫码登录

获取 QR 码 → 轮询扫码状态 → 扫码通过后返回 `Credentials`。**不做任何持久化**，凭据交由调用方保存。

导出函数：

| 函数 | 说明 |
|------|------|
| `FetchQRCode(ctx, baseURL)` | 获取登录二维码；`baseURL` 传空自动使用 `DefaultBaseURL` |
| `WaitForLogin(ctx, baseURL, qr, cb)` | 轮询扫码状态直到成功或超时（默认 8 分钟） |

内部行为：QR 过期自动刷新（最多 3 次）、单次轮询 35s 长超时、通过 `LoginCallbacks` 通知扫码/刷新事件。

典型用法：

```go
qr, _ := clawbot.FetchQRCode(ctx, "")                          // 首次登录用默认地址
creds, _ := clawbot.WaitForLogin(ctx, "", qr, clawbot.LoginCallbacks{
    OnScanned:     func() { fmt.Println("已扫码，等待确认...") },
    OnQRRefreshed: func(url string) { fmt.Println("新二维码:", url) },
})
// creds.BotToken / creds.BaseURL / creds.BotID / creds.UserID 全部来自微信返回
```

详见 [`login.go`](login.go)。

---

## 3. `poll.go` — 消息轮询

封装 `getUpdates` 长轮询，单次调用阻塞等待（≈35s）直到有消息或超时。调用方自行管理轮询循环和游标。

导出函数：

| 函数 | 说明 |
|------|------|
| `Poll(ctx, creds, getUpdatesBuf)` | 执行一次长轮询，返回 `PollResult`（消息列表 + 新游标） |
| `IsSessionExpired(err)` | 判断错误是否为会话过期（`errors.Is` 的便捷包装） |

典型用法：

```go
buf := ""
for {
    result, err := clawbot.Poll(ctx, creds, buf)
    if clawbot.IsSessionExpired(err) {
        break // 会话过期，需重新登录
    }
    if err != nil {
        time.Sleep(3 * time.Second)
        continue
    }
    buf = result.GetUpdatesBuf
    for _, msg := range result.Messages {
        go handleMessage(ctx, msg, creds)
    }
}
```

详见 [`poll.go`](poll.go)。

---

## 4. `send.go` — 消息回复

实现了文本发送、多媒体发送（CDN 加密上传）、流式分块发送三类能力，对应原设计中的方案 A + B + C。

导出函数：

| 函数 | 说明 |
|------|------|
| `SendText(ctx, creds, target, text)` | 发送纯文本消息 → `sendmessage` |
| `GetTypingTicket(ctx, creds, target)` | 获取 typing 凭证 → `getconfig` |
| `SendTyping(ctx, creds, action)` | 发送"正在输入…" → `sendtyping` |
| `SendImage(ctx, creds, target, media)` | 图片上传 + 发送 |
| `SendVideo(ctx, creds, target, media)` | 视频上传 + 发送 |
| `SendFile(ctx, creds, target, media)` | 文件上传 + 发送 |
| `SendMediaByPath(ctx, creds, target, media)` | 按扩展名自动路由到 Image/Video/File |
| `NewStreamSender(ctx, opts)` | 创建流式发送器（`StreamSender` 接口） |

媒体发送内部管线：`readFile → AES-128-ECB 加密 → getUploadUrl → POST CDN → sendMessage(CDNMedia)`

典型用法：

```go
target := clawbot.ReplyTarget{ToUserID: msg.FromUserID, ContextToken: msg.ContextToken}

// 纯文本
clawbot.SendText(ctx, creds, target, "你好")

// 图片（带文字说明）
clawbot.SendImage(ctx, creds, target, clawbot.MediaOpts{FilePath: "/tmp/photo.jpg", Caption: "看这张图"})

// 流式（对接 AI SSE）
sender := clawbot.NewStreamSender(ctx, clawbot.StreamSenderOpts{Creds: creds, Target: target})
defer sender.Close()
for token := range aiStream {
    sender.WriteChunk(token)
}
```

详见 [`send.go`](send.go)。
