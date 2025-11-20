package openai

import (
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/sashabaranov/go-openai"
)

type options struct {
	debugf     func(string, ...any)
	remoteMCPs []remoteMCPParams

	customizeMCPs []InitializedMCPClient

	// 简单回调
	reasoningCallback func(string)
	contentCallback   func(string)
	finishCallback    func(openai.FinishReason)

	// 工具调用回调
	toolCallRequestCallback  func(openai.ToolCall)
	toolCallResponseCallback func(openai.ChatCompletionMessage)
}

type remoteMCPParams struct {
	baseURL string
	options []transport.ClientOption
}

func mergeOptions(opts []Option) *options {
	o := &options{}
	for _, f := range opts {
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

// WithRemoteMCP 设置远程 MCP 的 URL, 可以设置多个
func WithRemoteMCP(baseURL string, opts ...transport.ClientOption) Option {
	return func(o *options) {
		if baseURL != "" {
			o.remoteMCPs = append(o.remoteMCPs, remoteMCPParams{
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

// WithInitializedMCP 设置自定义的已初始化完成的 MCP 客户端
func WithInitializedMCP(c InitializedMCPClient) Option {
	return func(o *options) {
		if c != nil {
			o.customizeMCPs = append(o.customizeMCPs, c)
		}
	}
}
