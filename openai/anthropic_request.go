package openai

import (
	"context"
	"net/http"

	hutil "github.com/Andrew-M-C/go.util/net/http"
	"github.com/Andrew-M-C/go.util/openai/anthropic"
	"github.com/sashabaranov/go-openai"
)

func connectAnthropic(
	ctx context.Context, config ModelConfig,
	messages []openai.ChatCompletionMessage,
	tools []openai.Tool,
	opt *options,
) (openai.ChatCompletionRequest, *http.Response, error) {
	req := openai.ChatCompletionRequest{
		Model:    config.Model,
		Messages: messages,
		Stream:   true,
	}
	if len(tools) > 0 {
		req.Tools = tools
		req.ToolChoice = "auto"
	}

	// 构建 Anthropic 格式的请求 body
	body, err := anthropic.BuildRequest(config.Model, messages, tools, opt.extraFields)
	if err != nil {
		return req, nil, err
	}

	// 构建 header：Anthropic 使用 x-api-key，而非 Authorization: Bearer
	h := http.Header{
		"Content-Type":      {"application/json"},
		"x-api-key":         {config.APIKey},
		"anthropic-version": {"2023-06-01"},
	}
	// 用户自定义 header 覆盖默认值
	for key, vals := range opt.extraHeaders {
		h[key] = vals
	}

	reqOpts := []hutil.RequestOption{
		hutil.WithRequestHeader(h),
		hutil.WithMethod("POST"),
		hutil.WithRequestBody(body),
		hutil.WithDebugger(opt.debugf),
	}

	rsp, err := hutil.Request(ctx, config.BaseURL, reqOpts...)
	if err != nil {
		return req, nil, err
	}

	// 用 SSE 中间层包装 Body，使其输出 OpenAI 格式的 SSE
	rsp.Body = anthropic.NewSSEReader(rsp.Body)
	return req, rsp, nil
}
