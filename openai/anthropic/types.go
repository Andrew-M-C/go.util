// Package anthropic 实现 OpenAI 协议与 Anthropic Messages API 协议之间的相互转换。
// 转换方向:
//   - 请求: OpenAI ChatCompletionRequest → Anthropic Messages Request JSON
//   - 响应: Anthropic SSE 流 → 包装为 OpenAI SSE 格式的 io.Reader
package anthropic

// -------- 请求结构 --------

// Request 是发往 Anthropic /v1/messages 的请求体
type Request struct {
	Model     string          `json:"model"`
	System    any             `json:"system,omitempty"` // string 或 []ContentBlock
	Messages  []Message       `json:"messages"`
	MaxTokens int             `json:"max_tokens"`
	Stream    bool            `json:"stream"`
	Tools     []Tool          `json:"tools,omitempty"`
	ToolChoice any            `json:"tool_choice,omitempty"`
	// 其余字段 (temperature, top_p, top_k, stop_sequences 等) 由 extraFields 注入
}

// Message 是 Anthropic 的消息结构
type Message struct {
	Role    string `json:"role"` // "user" | "assistant"
	Content any    `json:"content"` // string 或 []ContentBlock
}

// Tool 是 Anthropic 的工具定义
type Tool struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	InputSchema any    `json:"input_schema"` // JSON Schema object
}

// ToolChoiceAuto 对应 OpenAI "auto"
type ToolChoiceAuto struct {
	Type string `json:"type"` // "auto"
}

// ToolChoiceAny 对应 OpenAI "required"
type ToolChoiceAny struct {
	Type string `json:"type"` // "any"
}

// ToolChoiceTool 对应 OpenAI {"type": "function", "function": {"name": "..."}}
type ToolChoiceTool struct {
	Type string `json:"type"` // "tool"
	Name string `json:"name"`
}

// -------- Content Block 结构 --------

// ContentBlock 是 Anthropic content 数组中的单个 block
type ContentBlock struct {
	Type string `json:"type"`

	// type=text
	Text string `json:"text,omitempty"`

	// type=image
	Source *ImageSource `json:"source,omitempty"`

	// type=tool_use
	ID    string `json:"id,omitempty"`
	Name  string `json:"name,omitempty"`
	Input any    `json:"input,omitempty"`

	// type=tool_result
	ToolUseID string `json:"tool_use_id,omitempty"`
	Content   any    `json:"content,omitempty"` // string 或 []ContentBlock

	// type=thinking
	Thinking  string `json:"thinking,omitempty"`
	Signature string `json:"signature,omitempty"`
}

// ImageSource 是 Anthropic image content block 的 source 字段
type ImageSource struct {
	Type      string `json:"type"`       // "base64" | "url"
	MediaType string `json:"media_type"` // "image/png" 等
	Data      string `json:"data,omitempty"`
	URL       string `json:"url,omitempty"`
}

// -------- SSE 事件结构 --------

// SSEEvent 表示一次 Anthropic SSE 事件
type SSEEvent struct {
	EventType string // 从 "event: xxx" 行解析
	Data      string // 从 "data: xxx" 行解析
}

// MessageStartEvent 对应 event: message_start
type MessageStartEvent struct {
	Type    string          `json:"type"`
	Message MessageStartMsg `json:"message"`
}

// MessageStartMsg 是 message_start 中的 message 字段
type MessageStartMsg struct {
	ID    string `json:"id"`
	Model string `json:"model"`
	Role  string `json:"role"`
	Usage *Usage `json:"usage,omitempty"`
}

// ContentBlockStartEvent 对应 event: content_block_start
type ContentBlockStartEvent struct {
	Type         string            `json:"type"`
	Index        int               `json:"index"`
	ContentBlock ContentBlockStart `json:"content_block"`
}

// ContentBlockStart 是 content_block_start 中的 content_block 字段
type ContentBlockStart struct {
	Type string `json:"type"` // "text" | "thinking" | "tool_use" | "redacted_thinking"
	// type=tool_use
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
	// type=thinking / type=redacted_thinking
	Thinking  string `json:"thinking,omitempty"`
	Signature string `json:"signature,omitempty"`
}

// ContentBlockDeltaEvent 对应 event: content_block_delta
type ContentBlockDeltaEvent struct {
	Type  string            `json:"type"`
	Index int               `json:"index"`
	Delta ContentBlockDelta `json:"delta"`
}

// ContentBlockDelta 是 content_block_delta 中的 delta 字段
type ContentBlockDelta struct {
	Type string `json:"type"` // "text_delta" | "thinking_delta" | "signature_delta" | "input_json_delta"

	// type=text_delta
	Text string `json:"text,omitempty"`

	// type=thinking_delta
	Thinking string `json:"thinking,omitempty"`

	// type=signature_delta
	Signature string `json:"signature,omitempty"`

	// type=input_json_delta
	PartialJSON string `json:"partial_json,omitempty"`
}

// ContentBlockStopEvent 对应 event: content_block_stop
type ContentBlockStopEvent struct {
	Type  string `json:"type"`
	Index int    `json:"index"`
}

// MessageDeltaEvent 对应 event: message_delta
type MessageDeltaEvent struct {
	Type  string       `json:"type"`
	Delta MessageDelta `json:"delta"`
	Usage *Usage       `json:"usage,omitempty"`
}

// MessageDelta 是 message_delta 中的 delta 字段
type MessageDelta struct {
	StopReason   string `json:"stop_reason,omitempty"`
	StopSequence string `json:"stop_sequence,omitempty"`
}

// MessageStopEvent 对应 event: message_stop
type MessageStopEvent struct {
	Type string `json:"type"`
}

// ErrorEvent 对应 event: error
type ErrorEvent struct {
	Type  string    `json:"type"`
	Error ErrorBody `json:"error"`
}

// ErrorBody 是 error 事件中的 error 字段
type ErrorBody struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// Usage 是 Anthropic 的 token 使用情况
type Usage struct {
	InputTokens              int `json:"input_tokens,omitempty"`
	OutputTokens             int `json:"output_tokens,omitempty"`
	CacheCreationInputTokens int `json:"cache_creation_input_tokens,omitempty"`
	CacheReadInputTokens     int `json:"cache_read_input_tokens,omitempty"`
}
