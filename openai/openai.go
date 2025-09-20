// Package openai 封装一些调用 OpenAI 兼容协议的能力
package openai

import (
	"context"

	"github.com/sashabaranov/go-openai"
)

// ModelConfig 模型配置
type ModelConfig struct {
	Model   string `json:"mode,omitempty"`
	BaseURL string `json:"base_url,omitempty"`
	APIKey  string `json:"api_key,omitempty"`
}

// Process 完全自助式地处理一次完整的流式响应, 从发起请求开始, 自动调用工具, 直到模型返回完成为止
func Process(
	ctx context.Context, config ModelConfig, messages []openai.ChatCompletionMessage,
	options ...Option,
) (ProcessResponse, error) {
	opts := mergeOptions(options)
	p := &processor{
		Conf:     config,
		Opts:     opts,
		Messages: messages,
	}
	return p.do(ctx)
}

// ProcessResponse 表示一次请求的结果
type ProcessResponse struct {
	Messages     []openai.ChatCompletionMessage
	FinishReason openai.FinishReason
}
