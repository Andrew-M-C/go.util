package openai_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"

	utils "github.com/Andrew-M-C/go.util/openai"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/sashabaranov/go-openai"
)

// simpleMCP 是一个最小化 MCP 实现，固定返回单个工具和指定文本结果，用于测试工具调用回调。
type simpleMCP struct {
	toolName   string
	toolResult string
}

func (c *simpleMCP) ListTools(_ context.Context, _ mcp.ListToolsRequest) (*mcp.ListToolsResult, error) {
	return &mcp.ListToolsResult{
		Tools: []mcp.Tool{{Name: c.toolName, Description: "test tool"}},
	}, nil
}

func (c *simpleMCP) CallTool(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return utils.NewMCPCallToolResultWithString(c.toolResult)
}

// -------- WithResponseChunkCallback --------

func TestWithResponseChunkCallback(t *testing.T) {
	cv("WithResponseChunkCallback (Mock)", t, func() {
		cv("回调按原始 chunk 触发，内容各异", func() {
			var (
				mu     sync.Mutex
				chunks []openai.ChatCompletionStreamResponse
			)

			srv := newMockChatServer(t, func(_ capturedChatRequest) string {
				return sseAssistantStop("hello world")
			})
			defer srv.Close()

			_, err := utils.Process(
				context.Background(),
				mockModelConfig(srv.URL),
				[]openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleUser, Content: "hi"}},
				utils.WithResponseChunkCallback(func(c openai.ChatCompletionStreamResponse) {
					mu.Lock()
					chunks = append(chunks, c)
					mu.Unlock()
				}),
			)

			so(err, isNil)
			so(len(chunks), ge, 1)

			// 累积各次回调的 content，应等于完整响应
			var gotContent string
			for _, c := range chunks {
				if len(c.Choices) > 0 {
					gotContent += c.Choices[0].Delta.Content
				}
			}
			so(gotContent, eq, "hello world")

			// 验证不是每次都重复同一个 chunk（首包 content 非空，次包 finish 非空）
			contentCount := 0
			finishCount := 0
			for _, c := range chunks {
				if len(c.Choices) == 0 {
					continue
				}
				if c.Choices[0].Delta.Content != "" {
					contentCount++
				}
				if c.Choices[0].FinishReason != "" {
					finishCount++
				}
			}
			so(contentCount, ge, 1)
			so(finishCount, ge, 1)
		})

		cv("传入 nil 回调不 panic", func() {
			srv := newMockChatServer(t, func(_ capturedChatRequest) string {
				return sseAssistantStop("ok")
			})
			defer srv.Close()

			_, err := utils.Process(
				context.Background(),
				mockModelConfig(srv.URL),
				[]openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleUser, Content: "hi"}},
				utils.WithResponseChunkCallback(nil),
			)
			so(err, isNil)
		})

		cv("工具调用多轮时每轮的 chunk 均触发回调", func() {
			var (
				mu     sync.Mutex
				chunks []openai.ChatCompletionStreamResponse
			)

			var reqCount int32
			srv := newMockChatServer(t, func(_ capturedChatRequest) string {
				if atomic.AddInt32(&reqCount, 1) == 1 {
					return sseToolCall("call_cb_001", "cbmcp:echo", `{}`)
				}
				return sseAssistantStop("done")
			})
			defer srv.Close()

			_, err := utils.Process(
				context.Background(),
				mockModelConfig(srv.URL),
				[]openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleUser, Content: "test"}},
				utils.WithInitializedMCP(&simpleMCP{toolName: "echo", toolResult: "pong"}, "cbmcp"),
				utils.WithResponseChunkCallback(func(c openai.ChatCompletionStreamResponse) {
					mu.Lock()
					chunks = append(chunks, c)
					mu.Unlock()
				}),
			)

			so(err, isNil)
			// 两轮各至少 1 个 chunk
			so(len(chunks), ge, 2)
		})
	})
}

// -------- WithToolCallResponseCallback --------

func TestWithToolCallResponseCallback(t *testing.T) {
	cv("WithToolCallResponseCallback (Mock)", t, func() {
		cv("工具调用完成后收到 role=tool 消息", func() {
			var (
				mu       sync.Mutex
				toolRsps []openai.ChatCompletionMessage
			)

			const (
				toolID     = "call_tc_001"
				toolResult = "42"
			)

			var reqCount int32
			srv := newMockChatServer(t, func(_ capturedChatRequest) string {
				if atomic.AddInt32(&reqCount, 1) == 1 {
					return sseToolCall(toolID, "tcmcp:calc", `{"x":1}`)
				}
				return sseAssistantStop("答案是" + toolResult)
			})
			defer srv.Close()

			_, err := utils.Process(
				context.Background(),
				mockModelConfig(srv.URL),
				[]openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleUser, Content: "1+1=?"}},
				utils.WithInitializedMCP(&simpleMCP{toolName: "calc", toolResult: toolResult}, "tcmcp"),
				utils.WithToolCallResponseCallback(func(m openai.ChatCompletionMessage) {
					mu.Lock()
					toolRsps = append(toolRsps, m)
					mu.Unlock()
				}),
			)

			so(err, isNil)
			so(len(toolRsps), eq, 1)
			so(toolRsps[0].Role, eq, openai.ChatMessageRoleTool)
			so(toolRsps[0].Content, eq, toolResult)
			so(toolRsps[0].ToolCallID, eq, toolID)
		})

		cv("无工具调用时回调不触发", func() {
			var called int32
			srv := newMockChatServer(t, func(_ capturedChatRequest) string {
				return sseAssistantStop("no tools here")
			})
			defer srv.Close()

			_, err := utils.Process(
				context.Background(),
				mockModelConfig(srv.URL),
				[]openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleUser, Content: "hi"}},
				utils.WithToolCallResponseCallback(func(_ openai.ChatCompletionMessage) {
					atomic.AddInt32(&called, 1)
				}),
			)

			so(err, isNil)
			so(atomic.LoadInt32(&called), eq, int32(0))
		})

		cv("多个工具并发调用时每个都触发回调", func() {
			var (
				mu       sync.Mutex
				toolRsps []openai.ChatCompletionMessage
			)

			// 第一轮返回两个工具调用（两个不同 index），第二轮返回停止
			var reqCount int32
			srv := newMockChatServer(t, func(_ capturedChatRequest) string {
				if atomic.AddInt32(&reqCount, 1) == 1 {
					// 两个工具调用拼在同一个 SSE 包里，finish_reason=tool_calls
					return "" +
						`data: {"id":"c","model":"m","choices":[{"index":0,"delta":{"tool_calls":[` +
						`{"index":0,"id":"c0","type":"function","function":{"name":"multimcp:t1","arguments":"{}"}},` +
						`{"index":1,"id":"c1","type":"function","function":{"name":"multimcp:t2","arguments":"{}"}}` +
						`]},"finish_reason":"tool_calls"}]}` + "\n\n" +
						"data: [DONE]\n\n"
				}
				return sseAssistantStop("two tools done")
			})
			defer srv.Close()

			mcp2 := &simpleMCP{toolName: "t1", toolResult: "r1"}
			_, err := utils.Process(
				context.Background(),
				mockModelConfig(srv.URL),
				[]openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleUser, Content: "run tools"}},
				// 注册同一个 MCP，它同时拥有 t1 和 t2（CallTool 统一返回结果）
				utils.WithInitializedMCP(&dualMCP{}, "multimcp"),
				utils.WithToolCallResponseCallback(func(m openai.ChatCompletionMessage) {
					mu.Lock()
					toolRsps = append(toolRsps, m)
					mu.Unlock()
				}),
			)

			so(err, isNil)
			_ = mcp2
			so(len(toolRsps), eq, 2)
		})
	})
}

// dualMCP 提供两个工具供上面的多工具并发测试使用。
type dualMCP struct{}

func (*dualMCP) ListTools(_ context.Context, _ mcp.ListToolsRequest) (*mcp.ListToolsResult, error) {
	return &mcp.ListToolsResult{
		Tools: []mcp.Tool{
			{Name: "t1", Description: "tool 1"},
			{Name: "t2", Description: "tool 2"},
		},
	}, nil
}

func (*dualMCP) CallTool(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return utils.NewMCPCallToolResultWithString("result of " + req.Params.Name)
}

// -------- 集成测试（需要 DEEPSEEK_* 环境变量）--------

func TestOpenAI_ResponseChunkCallback_Integration(t *testing.T) {
	conf, ok := readDeepSeekConfig()
	if !ok {
		t.Skip("no deepseek config")
		return
	}

	cv("WithResponseChunkCallback + MergeResponseDeltaChunks 集成测试", t, func() {
		cv("逐 chunk 回调并使用 MergeResponseDeltaChunks 逐段拆分", func() {
			var (
				mu     sync.Mutex
				chunks []openai.ChatCompletionStreamResponse
			)

			messages := []openai.ChatCompletionMessage{
				{Role: openai.ChatMessageRoleUser, Content: "请用一句话介绍一下你自己"},
			}

			rsp, err := utils.Process(
				context.Background(), conf, messages,
				utils.WithResponseChunkCallback(func(c openai.ChatCompletionStreamResponse) {
					mu.Lock()
					chunks = append(chunks, c)
					mu.Unlock()
				}),
			)
			so(err, isNil)
			so(len(rsp.Messages), ge, 2)
			so(len(chunks), ge, 1)

			// 逐段拆分，验证：
			//   1. MergeResponseDeltaChunks 不会陷入无限循环
			//   2. 所有 chunks 都被消费完毕
			//   3. 至少有一段包含非空文本内容
			remaining := chunks
			var (
				segments    int
				hasContent  bool
			)
			for len(remaining) > 0 {
				merged, n := utils.MergeResponseDeltaChunks(remaining...)
				so(n, ge, 1) // 每次调用必须至少消费一个 chunk，否则死循环
				remaining = remaining[n:]
				segments++

				if len(merged.Choices) > 0 {
					c := merged.Choices[0].Delta.Content
					rc := merged.Choices[0].Delta.ReasoningContent
					if c != "" {
						hasContent = true
					}
					t.Logf("段 %d: count=%d content=%q reasoning_len=%d finish=%s",
						segments, n, c, len(rc), merged.Choices[0].FinishReason)
				} else {
					t.Logf("段 %d: count=%d (no choices, usage=%v)", segments, n, merged.Usage)
				}
			}
			so(segments, ge, 1)
			so(hasContent, eq, true)
			t.Logf("共 %d 个 chunk, 拆为 %d 段; assistant content: %q",
				len(chunks), segments, rsp.Messages[len(rsp.Messages)-1].Content)
		})
	})
}
