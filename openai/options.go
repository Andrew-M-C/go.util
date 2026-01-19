package openai

import (
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
	extraFields *jsonvalue.V
}

type initializedMCPParams struct {
	id     string
	client InitializedMCPClient
}

type remoteMCPParams struct {
	id      string
	baseURL string
	options []transport.ClientOption
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
		if baseURL != "" {
			o.remoteMCPs = append(o.remoteMCPs, remoteMCPParams{
				id:      stripMcpID(id),
				baseURL: baseURL,
				options: opts,
			})
		}
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

func stripMcpID(id string) string {
	id = strings.TrimSpace(id)
	id = strings.Replace(id, " ", "-", -1)
	id = strings.Replace(id, mcpClientNameSeparator, "-", -1)
	return id
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
