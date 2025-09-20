package openai

import "github.com/sashabaranov/go-openai"

// Delta 获取流式响应的 delta 部分
func Delta(m openai.ChatCompletionStreamResponse) openai.ChatCompletionStreamChoiceDelta {
	if len(m.Choices) == 0 {
		return openai.ChatCompletionStreamChoiceDelta{}
	}
	return m.Choices[0].Delta
}

// Content 获取正文文本内容
func Content(m openai.ChatCompletionStreamResponse) string {
	if len(m.Choices) == 0 {
		return ""
	}
	choice := m.Choices[0]
	return choice.Delta.Content
}

// ReasoningContent 获取思考内容
func ReasoningContent(m openai.ChatCompletionStreamResponse) string {
	if len(m.Choices) == 0 {
		return ""
	}
	choice := m.Choices[0]
	return choice.Delta.ReasoningContent
}

// FinishReason 获取结束原因
func FinishReason(m openai.ChatCompletionStreamResponse) openai.FinishReason {
	if len(m.Choices) == 0 {
		return ""
	}
	choice := m.Choices[0]
	return choice.FinishReason
}

// ToolCalls 获取工具调用
func ToolCalls(m openai.ChatCompletionStreamResponse) []openai.ToolCall {
	if len(m.Choices) == 0 {
		return nil
	}
	choice := m.Choices[0]
	return choice.Delta.ToolCalls
}
