package openai

import (
	"strings"

	"github.com/sashabaranov/go-openai"
)

// streamBuilder 流式响应构建器
type streamBuilder struct {
	// 入参
	opts *options

	// 中间参数
	rsp *openai.ChatCompletionStreamResponse

	reasoningContent strings.Builder
	content          strings.Builder

	toolCalls map[int]*toolCallBuilder // key 是 index

	finishReason openai.FinishReason
}

type toolCallBuilder struct {
	id           string
	typ          openai.ToolType
	functionName string
	functionArgs strings.Builder
}

func (b *streamBuilder) Reset() {
	b.rsp = nil
	b.reasoningContent.Reset()
	b.content.Reset()
	b.toolCalls = nil
	b.finishReason = ""
}

// AddResponse 添加一次响应数据
func (b *streamBuilder) AddResponse(from openai.ChatCompletionStreamResponse) {
	if b.rsp == nil {
		b.rsp = &from
	}
	if from.Usage != nil {
		b.rsp.Usage = from.Usage
	}
	if len(b.rsp.Choices) == 0 && len(from.Choices) > 0 {
		b.rsp.Choices = from.Choices
	}
	if len(from.Choices) == 0 {
		return
	}

	choice := from.Choices[0]
	if choice.FinishReason != "" {
		b.finishReason = choice.FinishReason
		b.opts.finishCallback(b.finishReason)
	}
	delta := choice.Delta
	if delta.ReasoningContent != "" {
		b.reasoningContent.WriteString(delta.ReasoningContent)
		b.opts.reasoningCallback(delta.ReasoningContent)
	}
	if delta.Content != "" {
		b.content.WriteString(delta.Content)
		b.opts.contentCallback(delta.Content)
	}

	if delta.ToolCalls != nil {
		if b.toolCalls == nil {
			b.toolCalls = make(map[int]*toolCallBuilder)
		}
		for _, tc := range delta.ToolCalls {
			index := 0
			if tc.Index != nil {
				index = *tc.Index
			}
			bdr := b.toolCalls[index]
			if bdr == nil {
				bdr = &toolCallBuilder{
					id:           tc.ID,
					typ:          tc.Type,
					functionName: tc.Function.Name,
				}
				b.toolCalls[index] = bdr
			}
			bdr.functionArgs.WriteString(tc.Function.Arguments)
		}
	}
}

// Done 完成构建
func (b *streamBuilder) Done() openai.ChatCompletionStreamResponse {
	if len(b.rsp.Choices) == 0 {
		return *b.rsp
	}
	b.rsp.Choices[0].Delta.ReasoningContent = b.reasoningContent.String()
	b.rsp.Choices[0].Delta.Content = b.content.String()
	b.rsp.Choices[0].FinishReason = b.finishReason

	if len(b.toolCalls) == 0 {
		return *b.rsp
	}

	b.rsp.Choices[0].Delta.ToolCalls = make([]openai.ToolCall, len(b.toolCalls))
	for index, tc := range b.toolCalls {
		b.rsp.Choices[0].Delta.ToolCalls[index] = openai.ToolCall{
			ID:   tc.id,
			Type: tc.typ,
			Function: openai.FunctionCall{
				Name:      tc.functionName,
				Arguments: tc.functionArgs.String(),
			},
			Index: &index,
		}
	}
	return *b.rsp
}
