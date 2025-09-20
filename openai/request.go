package openai

import (
	"context"
	"net/http"

	hutil "github.com/Andrew-M-C/go.util/net/http"
	"github.com/sashabaranov/go-openai"
)

func connect(
	ctx context.Context, config ModelConfig,
	messages []openai.ChatCompletionMessage,
	tools []openai.Tool,
	opt *options,
) (*http.Response, error) {
	h := http.Header{
		"Authorization": {"Bearer " + config.APIKey},
	}

	req := openai.ChatCompletionRequest{
		Model:      config.Model,
		Messages:   messages,
		Stream:     true,
		Tools:      tools,
		ToolChoice: "auto",
	}
	rsp, err := hutil.Request(ctx, config.BaseURL,
		hutil.WithRequestHeader(h),
		hutil.WithRequestBody(req),
		hutil.WithMethod("POST"),
		hutil.WithDebugger(opt.debugf),
	)
	return rsp, err
}
