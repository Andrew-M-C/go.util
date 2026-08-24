package openai

import (
	"context"
	"net/http"
	"strings"

	jsonvalue "github.com/Andrew-M-C/go.jsonvalue"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/sashabaranov/go-openai"
)

type options struct {
	debugf     func(string, ...any)
	remoteMCPs []remoteMCPParams

	customizeMCPs []initializedMCPParams

	// 简单回调
	reasoningCallback func(string)
	contentCallback   func(string)
	finishCallback    func(openai.FinishReason)

	// 工具调用回调
	toolCallRequestCallback  func(openai.ToolCall)
	toolCallResponseCallback func(openai.ChatCompletionMessage)

	// 额外参数
	extraFields  *jsonvalue.V
	extraHeaders http.Header

	// Anthropic 协议转换模式
	useAnthropic bool

	// preRequestCallback 发送请求之前的回调, nilable
	preRequestCallback PreRequestCallback
}

type initializedMCPParams struct {
	id     string
	client InitializedMCPClient

	includeTools map[string]struct{}
}

type remoteMCPParams struct {
	id      string
	baseURL string
	options []transport.ClientOption

	includeTools map[string]struct{}
}

func mergeOptions(opts []Option) *options {
	o := &options{}
	for _, f := range opts {
		if f == nil {
			continue
		}
		f(o)
	}
	// 兜底值配置
	if o.debugf == nil {
		o.debugf = func(string, ...any) {}
	}
	if o.reasoningCallback == nil {
		o.reasoningCallback = func(string) {}
	}
	if o.contentCallback == nil {
		o.contentCallback = func(string) {}
	}
	if o.finishCallback == nil {
		o.finishCallback = func(openai.FinishReason) {}
	}
	if o.toolCallRequestCallback == nil {
		o.toolCallRequestCallback = func(openai.ToolCall) {}
	}
	if o.toolCallResponseCallback == nil {
		o.toolCallResponseCallback = func(openai.ChatCompletionMessage) {}
	}
	return o
}

// Option 额外选项
type Option func(*options)

// WithDebugger 设置调试函数
func WithDebugger(d func(string, ...any)) Option {
	return func(o *options) {
		if d != nil {
			o.debugf = d
		}
	}
}

// WithRemoteMCP 设置远程 MCP 的 URL, 可以设置多个。
// 参数 id 可以是任意不含空格和毛好的字符串, 多个 MCP 之间不得重复
func WithRemoteMCP(baseURL string, id string, opts ...transport.ClientOption) Option {
	return func(o *options) {
		if baseURL == "" {
			return
		}
		o.remoteMCPs = append(o.remoteMCPs, remoteMCPParams{
			id:      stripMcpID(id),
			baseURL: baseURL,
			options: opts,
		})
	}
}

// WithRemoteMCPAndSpecifyTools 的功能等同于 WithRemoteMCP, 但明确指定只引用其中的某些工具。
func WithRemoteMCPAndSpecifyTools(baseURL string, id string, toolsAndOpts ...any) Option {
	return func(o *options) {
		if baseURL == "" {
			return
		}
		opts := make([]transport.ClientOption, 0, len(toolsAndOpts))
		includeTools := make(map[string]struct{}, len(toolsAndOpts))

		for _, arg := range toolsAndOpts {
			if arg == nil {
				continue
			}
			switch v := arg.(type) {
			case string:
				includeTools[v] = struct{}{}
			case transport.ClientOption:
				opts = append(opts, v)
			default:
				// 不支持的类型
			}
		}

		o.remoteMCPs = append(o.remoteMCPs, remoteMCPParams{
			id:           stripMcpID(id),
			baseURL:      baseURL,
			options:      opts,
			includeTools: includeTools,
		})
	}
}

// WithReasoningCallback 设置思考内容回调函数
func WithReasoningCallback(c func(delta string)) Option {
	return func(o *options) {
		if c != nil {
			o.reasoningCallback = c
		}
	}
}

// WithContentCallback 设置内容回调函数
func WithContentCallback(c func(delta string)) Option {
	return func(o *options) {
		if c != nil {
			o.contentCallback = c
		}
	}
}

// WithFinishCallback 设置 (阶段性的) 结束回调函数
func WithFinishCallback(c func(openai.FinishReason)) Option {
	return func(o *options) {
		if c != nil {
			o.finishCallback = c
		}
	}
}

// WithToolCallRequestCallback 设置工具调用请求回调函数
func WithToolCallRequestCallback(c func(openai.ToolCall)) Option {
	return func(o *options) {
		if c != nil {
			o.toolCallRequestCallback = c
		}
	}
}

// WithToolCallResponseCallback 设置工具调用响应回调函数
func WithToolCallResponseCallback(c func(openai.ChatCompletionMessage)) Option {
	return func(o *options) {
		if c != nil {
			o.toolCallResponseCallback = c
		}
	}
}

// WithInitializedMCP 设置自定义的已初始化完成的 MCP 客户端, 可以设置多个
// 参数 id 可以是任意不含空格和毛好的字符串, 多个 MCP 之间不得重复
func WithInitializedMCP(c InitializedMCPClient, id string) Option {
	return func(o *options) {
		if c != nil {
			m := initializedMCPParams{
				id:     stripMcpID(id),
				client: c,
			}
			o.customizeMCPs = append(o.customizeMCPs, m)
		}
	}
}

// WithTools 设置可调用的工具
func WithTools(tools ToolManager, id string) Option {
	return WithInitializedMCP(tools, id)
}

// WithInitializedMCPAndSpecifyTools 同 WithInitializedMCP, 但明确指定只引用其中的某些工具。
func WithInitializedMCPAndSpecifyTools(c InitializedMCPClient, id string, tools ...string) Option {
	return func(o *options) {
		if c == nil {
			return
		}
		m := initializedMCPParams{
			id:           stripMcpID(id),
			client:       c,
			includeTools: sliceToCollection(tools),
		}
		o.customizeMCPs = append(o.customizeMCPs, m)
	}
}

func stripMcpID(id string) string {
	id = strings.TrimSpace(id)
	id = strings.Replace(id, " ", "-", -1)
	id = strings.Replace(id, mcpClientNameSeparator, "-", -1)
	return id
}

func sliceToCollection[T comparable](sli []T) map[T]struct{} {
	res := make(map[T]struct{}, len(sli))
	for _, v := range sli {
		res[v] = struct{}{}
	}
	return res
}

// WithExtraFields 设置请求 completion 的额外参数。后设置的会覆盖前面设置的 key。
// 如果传入的参数不是一个有效的 JSON object, 则不进行设置
func WithExtraFields(fields any) Option {
	if fields == nil {
		return nil
	}
	j, err := jsonvalue.Import(fields)
	if err != nil {
		return nil
	}
	return func(o *options) {
		if o.extraFields == nil {
			o.extraFields = j
			return
		}
		j.RangeObjects(func(key string, value *jsonvalue.V) bool {
			o.extraFields.At(key).Set(value)
			return true
		})
	}
}

// WithAnthropic 在请求时转换为 Anthropic 协议并发往 /v1/messages 端点，拿到响应后转换回
// OpenAI 协议。ModelConfig.BaseURL 应设置为完整的 Anthropic 端点 URL，例如
// "https://api.anthropic.com/v1/messages"。
//
// 注意事项：
//   - extra fields 中的字段会直接注入到 Anthropic 请求 JSON 中，其中 stop → stop_sequences 会自动映射
//   - 若需要设置 anthropic-beta 等 header，请使用 WithHeader()
//   - 建议不要通过 WithExtraFields 设置 tool_choice，库会自动处理工具调用格式
func WithAnthropic() Option {
	return func(o *options) {
		o.useAnthropic = true
	}
}

// WithHeader 设置自定义 HTTP header，会覆盖默认 header 中的同名字段。
// 常用于设置 anthropic-version、anthropic-beta 等 Anthropic 专用 header。
func WithHeader(h http.Header) Option {
	return func(o *options) {
		if o.extraHeaders == nil {
			o.extraHeaders = h.Clone()
			return
		}
		for key, vals := range h {
			o.extraHeaders[key] = vals
		}
	}
}

// WithIncludeUsage 设置是否包含 usage 字段
func WithIncludeUsage() Option {
	return func(o *options) {
		if o.extraFields == nil {
			o.extraFields = jsonvalue.NewObject()
		}
		o.extraFields.At("stream_options", "include_usage").Set(true)
	}
}

// PreRequestCallback 表示准备发出请求之前的回调函数, 支持调用方直接改写底层 messages。
// 每次向模型发起请求前都会调用（含工具调用后的后续轮次）。最终发出的消息是否正确由回调方自行负责。
type PreRequestCallback func(
	ctx context.Context,
	params *PreRequestContext,
) error

// PreRequestContext 表示准备发出请求之前的参数。
// RequestMessages 即处理器当前持有的消息切片，回调可原地修改元素，也可整体替换该切片。
type PreRequestContext struct {
	RequestMessages []openai.ChatCompletionMessage
}

// WithPreRequestCallback 设置准备发出请求之前的回调函数, 支持调用方直接改写底层 messages。
// 传入 nil 时不覆盖已有回调。
func WithPreRequestCallback(c PreRequestCallback) Option {
	return func(o *options) {
		if c != nil {
			o.preRequestCallback = c
		}
	}
}
