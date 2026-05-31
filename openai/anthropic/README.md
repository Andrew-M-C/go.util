# anthropic

在 **OpenAI Chat Completions 协议** 与 **Anthropic Messages API** 之间做双向转换。调用方始终使用 `github.com/sashabaranov/go-openai` 的消息与流式结构；实际 HTTP 可走 Anthropic `/v1/messages`。

本包既可被父级 `openai` 包通过 `WithAnthropic()` 间接使用，也可单独引用做协议层转换。

## 能力概览

| 方向 | 说明 |
|------|------|
| 请求 | `BuildRequest`：OpenAI `[]ChatCompletionMessage` + tools + extraFields → Anthropic Messages JSON |
| 响应 | `NewSSEReader`：Anthropic SSE（`event:` + `data:`）→ OpenAI SSE（仅 `data:`，含 `[DONE]`） |

**仅支持流式**：请求体始终带 `"stream": true`，不提供非流式响应转换。

## 用法一：通过父包 `openai`（推荐）

与现有 `Process` / MCP / tool 流程一致，只需加 option 并改配置。

```go
import (
    "github.com/Andrew-M-C/go.util/openai"
    openaigo "github.com/sashabaranov/go-openai"
)

conf := openai.ModelConfig{
    Model:   "claude-sonnet-4-20250514",
    APIKey:  os.Getenv("ANTHROPIC_API_KEY"),
    // 完整 URL，不会做 path 替换
    BaseURL: "https://api.anthropic.com/v1/messages",
}

rsp, err := openai.Process(ctx, conf, messages,
    openai.WithAnthropic(),
    openai.WithExtraFields(map[string]any{
        "max_tokens": 8192,
        "thinking": map[string]any{
            "type":          "enabled",
            "budget_tokens": 10000,
        },
    }),
    // 覆盖默认 header（见下文「HTTP Header」）
    openai.WithHeader(http.Header{
        "anthropic-version": {"2023-06-01"},
        "anthropic-beta":    {"extended-thinking-2025-01-24"},
    }),
)
```

父包会：

1. 用 `anthropic.BuildRequest` 构造请求体；
2. 将 `Authorization: Bearer` 改为 `x-api-key`，并设置默认 `anthropic-version: 2023-06-01`；
3. 用 `anthropic.NewSSEReader` 包装响应 `Body`，上层 `ReadSSEJsonData` / `streamBuilder` 仍按 OpenAI SSE 解析。

## 用法二：直接引用本包

适合自建 HTTP 客户端、网关或测试。

### 构造请求

```go
import (
    jsonvalue "github.com/Andrew-M-C/go.jsonvalue"
    "github.com/Andrew-M-C/go.util/openai/anthropic"
    "github.com/sashabaranov/go-openai"
)

extra, _ := jsonvalue.Import(map[string]any{
    "temperature": 0.7,
    "stop":        []string{"END"}, // 会自动映射为 stop_sequences
})

body, err := anthropic.BuildRequest(
    "claude-sonnet-4-20250514",
    messages,
    tools,
    extra,
)
// POST body 到 BaseURL，Header: x-api-key, anthropic-version, Content-Type: application/json
```

### 转换 SSE 响应

```go
resp, _ := http.Post(url, headers, bytes.NewReader(body))
// resp.Body 为 Anthropic 原始 SSE 流

openaiBody := anthropic.NewSSEReader(resp.Body)
defer openaiBody.Close()

// 之后按 OpenAI 流式约定读 data: 行即可
scanner := bufio.NewScanner(openaiBody)
for scanner.Scan() {
    line := scanner.Text()
    // data: {...} 或 data: [DONE]
}
```

对外导出 API：

- `BuildRequest(model, messages, tools, extraFields *jsonvalue.V) ([]byte, error)`
- `NewSSEReader(src io.ReadCloser) io.ReadCloser`

## 协议转换说明（简表）

### 请求

- `role=system` → 顶层 `system`（多条用 `\n\n` 拼接）
- `role=user` 纯文本 → `content` 字符串；`MultiContent` → content block 数组
- `image_url`（含 `data:image/...;base64,...`）→ Anthropic `image` + `source`
- `role=assistant` 的 `tool_calls` → `tool_use` blocks；顺序：**thinking → text → tool_use**
- `role=tool` → `tool_result` blocks，合并进后续 **user** 消息（不设 `is_error`）
- 连续多条 `role=user` → 合并为一条（`\n\n` 拼接）
- OpenAI `tools` → Anthropic `tools`（`input_schema` 来自 `Function.Parameters`）
- `extraFields`：`stop` → `stop_sequences`；`tool_choice` 按 OpenAI 语义映射（见下）

### 响应（SSE）

- `text_delta` → `delta.content`
- `thinking_delta` / `signature_delta` → `delta.reasoning_content`（signature 见下）
- `tool_use` + `input_json_delta` → `delta.tool_calls`
- `message_delta` → `finish_reason` + `usage`（`input_tokens`/`output_tokens` → prompt/completion/total）
- `message_stop` → `data: [DONE]`
- `error` → OpenAI 风格错误行 + `[DONE]`

### Extended thinking 与 `reasoning_content`

Anthropic 的 `signature` 在流式里通过 `signature_delta` 在 thinking block **末尾**发出。本包在写入 OpenAI 时把它嵌进 `reasoning_content`：

```text
{thinking 正文}

<signature>xxxxx</signature>
```

若为 `redacted_thinking`，还会追加：

```text

<think>true</think>
```

下一轮请求时，`BuildRequest` 会从 assistant 的 `ReasoningContent` 中解析上述标记，还原为 Anthropic 的 `thinking` / `redacted_thinking` block（含 `signature`）。**多轮 extended thinking 必须原样带回 assistant 消息中的 `ReasoningContent`。**

## 使用中需要注意的点

### 配置与端点

1. **`BaseURL` 必须是完整 URL**（例如 `https://api.anthropic.com/v1/messages`），不会做 `/v1/chat/completions` → `/v1/messages` 的路径替换。
2. **`APIKey` 原样作为 `x-api-key`**，不校验前缀；父包不会再用 `Authorization: Bearer`。
3. **仅流式**：不要期望非流式 JSON 响应；`BuildRequest` 固定 `stream: true`。

### `max_tokens` 与 extra 字段

4. 未在 extra 中设置 `max_tokens` 时，**默认 `100000`**。
5. `WithExtraFields` / `jsonvalue` 中的字段会 **合并进 Anthropic 请求根对象**；Anthropic 不认识的 key 也会原样下发（由服务端决定是否接受）。
6. **`stop` 会映射为 `stop_sequences`**；不要在 extra 里同时写两套。
7. **建议不要自行设置 `tool_choice`**：有 tools 时父流程会设 `auto`；若 extra 里写了 `tool_choice`，会按规则映射：
   - `"auto"` → `{"type":"auto"}`
   - `"required"` → `{"type":"any"}`
   - `{"type":"function","function":{"name":"..."}}` → `{"type":"tool","name":"..."}`

### HTTP Header

8. 默认：`Content-Type: application/json`、`x-api-key`、`anthropic-version: 2023-06-01`。
9. **`anthropic-beta`、新版 `anthropic-version` 等请用父包 `WithHeader(http.Header)`**；同名 header 会覆盖默认值。

### 消息序列（Anthropic 约束）

10. **对话应以 `user` 消息开头**；本包不会自动插入空 user 消息，顺序错误由调用方负责。
11. **连续多条 `user` 会自动合并**；若你依赖「多条独立 user」的语义，需在业务层自行处理。
12. **`role=tool` 会排在 assistant 的 `tool_use` 之后**，转成带 `tool_result` 的 user 消息；与 OpenAI 多 tool 轮次兼容，但消息顺序需符合「assistant 调工具 → tool 结果」的习惯写法。

### 工具与图片

13. Tool 定义需为 OpenAI `Tool` + `FunctionDefinition`（含 `Parameters` JSON Schema）。
14. 图片支持 `MultiContent` 里的 `image_url`：`data:...;base64,...` 与 `https://...` URL 两种形式。

### Thinking / 温度等

15. 开启 extended thinking 时，在 extra 中传 Anthropic 的 `thinking` 对象（如 `type: enabled`、`budget_tokens`）；具体字段以 Anthropic 文档为准。
16. **`temperature` 不做范围裁剪**（OpenAI 与 Anthropic 上限不同），由调用方自行控制。

### 直接调用 `NewSSEReader` 时

17. 传入的 `src` 在 goroutine 结束后会被 **Close**；若还需读原始 body，请自行 `TeeReader` 或不要交给 `NewSSEReader`。
18. 只处理 Anthropic 官方 SSE 事件类型；`ping` 会忽略。

### 测试

19. 单元测试见同目录 `anthropic_test.go`，覆盖 DUSCUSS 中约定的各转换项；运行：

```bash
cd openai && go test ./anthropic/... -v
```

## 相关文件

| 文件 | 职责 |
|------|------|
| `request.go` | `BuildRequest` 及消息 / tool / thinking 转换 |
| `sse.go` | `NewSSEReader` 与 Anthropic → OpenAI SSE 映射 |
| `types.go` | Anthropic 事件与请求的结构体（解析用） |

父包：`openai/options.go`（`WithAnthropic`、`WithHeader`）、`openai/anthropic_request.go`（`connectAnthropic`）。
