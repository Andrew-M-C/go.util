package openai

import (
	"context"
	"net/http"

	jsonvalue "github.com/Andrew-M-C/go.jsonvalue"
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
		"Content-Type":  {"application/json"},
		"Authorization": {"Bearer " + config.APIKey},
	}

	req := openai.ChatCompletionRequest{
		Model:    config.Model,
		Messages: messages,
		Stream:   true,
	}
	if len(tools) > 0 {
		req.Tools = tools
		req.ToolChoice = "auto"
	}

	options := []hutil.RequestOption{
		hutil.WithRequestHeader(h),
		hutil.WithMethod("POST"),
		hutil.WithDebugger(opt.debugf),
	}

	if opt.extraFields == nil {
		options = append(options, hutil.WithRequestBody(req))
	} else {
		j := jsonvalue.New(req)
		opt.extraFields.RangeObjects(func(key string, value *jsonvalue.V) bool {
			j.At(key).Set(value)
			return true
		})
		b, _ := j.Marshal(jsonvalue.OptUTF8())
		options = append(options, hutil.WithRequestBody(b))
	}

	// if opt.extraFields == nil {
	// 	j := jsonvalue.New(req)
	// 	j.At("thinking", "type").Set("enabled")
	// 	b, _ := j.Marshal(jsonvalue.OptUTF8())
	// 	body = b
	// }

	rsp, err := hutil.Request(ctx, config.BaseURL, options...)
	return rsp, err
}
