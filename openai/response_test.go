package openai_test

import (
	"testing"

	utils "github.com/Andrew-M-C/go.util/openai"
	"github.com/sashabaranov/go-openai"
)

func TestMergeResponseDeltaChunks(t *testing.T) {
	// ---- 构建辅助函数 ----

	mkBase := func(id string) openai.ChatCompletionStreamResponse {
		return openai.ChatCompletionStreamResponse{ID: id, Model: "m", Created: 1}
	}
	withRole := func(c openai.ChatCompletionStreamResponse) openai.ChatCompletionStreamResponse {
		c.Choices = []openai.ChatCompletionStreamChoice{{
			Delta: openai.ChatCompletionStreamChoiceDelta{Role: openai.ChatMessageRoleAssistant},
		}}
		return c
	}
	withContent := func(c openai.ChatCompletionStreamResponse, s string) openai.ChatCompletionStreamResponse {
		c.Choices = []openai.ChatCompletionStreamChoice{{
			Delta: openai.ChatCompletionStreamChoiceDelta{Content: s},
		}}
		return c
	}
	withReasoning := func(c openai.ChatCompletionStreamResponse, s string) openai.ChatCompletionStreamResponse {
		c.Choices = []openai.ChatCompletionStreamChoice{{
			Delta: openai.ChatCompletionStreamChoiceDelta{ReasoningContent: s},
		}}
		return c
	}
	mkFinish := func(reason openai.FinishReason) openai.ChatCompletionStreamResponse {
		c := mkBase("id")
		c.Choices = []openai.ChatCompletionStreamChoice{{FinishReason: reason}}
		return c
	}
	mkToolCall := func(tcIdx int, id, name, args string) openai.ChatCompletionStreamResponse {
		idx := tcIdx
		c := mkBase("id")
		c.Choices = []openai.ChatCompletionStreamChoice{{
			Delta: openai.ChatCompletionStreamChoiceDelta{
				ToolCalls: []openai.ToolCall{{
					Index:    &idx,
					ID:       id,
					Type:     openai.ToolTypeFunction,
					Function: openai.FunctionCall{Name: name, Arguments: args},
				}},
			},
		}}
		return c
	}

	cv("MergeResponseDeltaChunks", t, func() {
		cv("空输入返回零值, count=0", func() {
			rsp, n := utils.MergeResponseDeltaChunks()
			so(n, eq, 0)
			so(rsp.ID, eq, "")
		})

		cv("单个 chunk 原样返回, count=1", func() {
			rsp, n := utils.MergeResponseDeltaChunks(withContent(mkBase("a"), "hello"))
			so(n, eq, 1)
			so(rsp.ID, eq, "a")
			so(rsp.Choices[0].Delta.Content, eq, "hello")
		})

		cv("两个 content chunk 拼接", func() {
			rsp, n := utils.MergeResponseDeltaChunks(
				withContent(mkBase("x"), "foo"),
				withContent(mkBase("x"), "bar"),
			)
			so(n, eq, 2)
			so(rsp.Choices[0].Delta.Content, eq, "foobar")
		})

		cv("role-only 首包透明吸收进 content", func() {
			rsp, n := utils.MergeResponseDeltaChunks(
				withRole(mkBase("x")),
				withContent(mkBase("x"), "hi"),
			)
			so(n, eq, 2)
			so(rsp.Choices[0].Delta.Role, eq, openai.ChatMessageRoleAssistant)
			so(rsp.Choices[0].Delta.Content, eq, "hi")
		})

		cv("role-only 首包透明吸收进 reasoning", func() {
			rsp, n := utils.MergeResponseDeltaChunks(
				withRole(mkBase("x")),
				withReasoning(mkBase("x"), "think"),
			)
			so(n, eq, 2)
			so(rsp.Choices[0].Delta.Role, eq, openai.ChatMessageRoleAssistant)
			so(rsp.Choices[0].Delta.ReasoningContent, eq, "think")
		})

		cv("content 后紧跟 reasoning 停止合并, count=1", func() {
			rsp, n := utils.MergeResponseDeltaChunks(
				withContent(mkBase("x"), "c"),
				withReasoning(mkBase("x"), "r"),
			)
			so(n, eq, 1)
			so(rsp.Choices[0].Delta.Content, eq, "c")
			so(rsp.Choices[0].Delta.ReasoningContent, eq, "")
		})

		cv("两个 reasoning chunk 拼接", func() {
			rsp, n := utils.MergeResponseDeltaChunks(
				withReasoning(mkBase("x"), "aa"),
				withReasoning(mkBase("x"), "bb"),
			)
			so(n, eq, 2)
			so(rsp.Choices[0].Delta.ReasoningContent, eq, "aabb")
		})

		cv("tool_calls 同 index: arguments 拼接, id/name 取首次非空", func() {
			idx0 := 0
			rsp, n := utils.MergeResponseDeltaChunks(
				mkToolCall(0, "call_1", "fn", `{"a":`),
				mkToolCall(0, "", "", `"v"}`),
			)
			so(n, eq, 2)
			tcs := rsp.Choices[0].Delta.ToolCalls
			so(len(tcs), eq, 1)
			so(*tcs[0].Index, eq, idx0)
			so(tcs[0].ID, eq, "call_1")
			so(tcs[0].Function.Name, eq, "fn")
			so(tcs[0].Function.Arguments, eq, `{"a":"v"}`)
		})

		cv("tool_calls 跨 index 追加, 同一段", func() {
			rsp, n := utils.MergeResponseDeltaChunks(
				mkToolCall(0, "call_0", "fn0", `{}`),
				mkToolCall(1, "call_1", "fn1", `{}`),
			)
			so(n, eq, 2)
			so(len(rsp.Choices[0].Delta.ToolCalls), eq, 2)
		})

		cv("tool_calls 后紧跟 content 停止合并, count=1", func() {
			rsp, n := utils.MergeResponseDeltaChunks(
				mkToolCall(0, "call_0", "fn", `{}`),
				withContent(mkBase("x"), "done"),
			)
			so(n, eq, 1)
			so(len(rsp.Choices[0].Delta.ToolCalls), eq, 1)
			so(rsp.Choices[0].Delta.Content, eq, "")
		})

		cv("finish chunk 单独成段, count=1", func() {
			rsp, n := utils.MergeResponseDeltaChunks(mkFinish(openai.FinishReasonStop))
			so(n, eq, 1)
			so(rsp.Choices[0].FinishReason, eq, openai.FinishReasonStop)
		})

		cv("content 后紧跟 finish 停止合并, count=1", func() {
			rsp, n := utils.MergeResponseDeltaChunks(
				withContent(mkBase("x"), "hello"),
				mkFinish(openai.FinishReasonStop),
			)
			so(n, eq, 1)
			so(rsp.Choices[0].Delta.Content, eq, "hello")
			so(string(rsp.Choices[0].FinishReason), eq, "")
		})

		cv("两个 finish chunk 各自成段", func() {
			_, n := utils.MergeResponseDeltaChunks(
				mkFinish(openai.FinishReasonStop),
				mkFinish(openai.FinishReasonStop),
			)
			so(n, eq, 1)
		})

		cv("usage-only chunk 合并, Usage 取最后一次", func() {
			c1 := mkBase("x")
			c1.Usage = &openai.Usage{PromptTokens: 5}
			c2 := openai.ChatCompletionStreamResponse{
				ID:    "x",
				Usage: &openai.Usage{PromptTokens: 10, CompletionTokens: 20, TotalTokens: 30},
			}
			rsp, n := utils.MergeResponseDeltaChunks(c1, c2)
			so(n, eq, 2)
			so(rsp.Usage.TotalTokens, eq, 30)
		})

		cv("壳字段 (ID/Model/Created) 始终取第一个包", func() {
			c1 := withContent(mkBase("first"), "a")
			c1.Model = "model-1"
			c2 := withContent(mkBase("second"), "b")
			c2.Model = "model-2"

			rsp, n := utils.MergeResponseDeltaChunks(c1, c2)
			so(n, eq, 2)
			so(rsp.ID, eq, "first")
			so(rsp.Model, eq, "model-1")
			so(rsp.Choices[0].Delta.Content, eq, "ab")
		})

		cv("混合类型 (content+reasoning 同包) 单独成段", func() {
			c := mkBase("x")
			c.Choices = []openai.ChatCompletionStreamChoice{{
				Delta: openai.ChatCompletionStreamChoiceDelta{
					Content:          "txt",
					ReasoningContent: "think",
				},
			}}
			_, n := utils.MergeResponseDeltaChunks(c, withContent(mkBase("x"), "more"))
			so(n, eq, 1)
		})

		cv("多段流水线: role+contents+finish 逐段拆分", func() {
			chunks := []openai.ChatCompletionStreamResponse{
				withRole(mkBase("x")),
				withContent(mkBase("x"), "hello"),
				withContent(mkBase("x"), " world"),
				mkFinish(openai.FinishReasonStop),
			}

			// 第一段: role + 2 × content = 3 个
			seg1, n1 := utils.MergeResponseDeltaChunks(chunks...)
			so(n1, eq, 3)
			so(seg1.Choices[0].Delta.Role, eq, openai.ChatMessageRoleAssistant)
			so(seg1.Choices[0].Delta.Content, eq, "hello world")

			// 第二段: finish = 1 个
			seg2, n2 := utils.MergeResponseDeltaChunks(chunks[n1:]...)
			so(n2, eq, 1)
			so(seg2.Choices[0].FinishReason, eq, openai.FinishReasonStop)
		})
	})
}
