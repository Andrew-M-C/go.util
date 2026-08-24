package openai_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	utils "github.com/Andrew-M-C/go.util/openai"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/sashabaranov/go-openai"
)

type capturedChatRequest struct {
	Model    string                         `json:"model"`
	Messages []openai.ChatCompletionMessage `json:"messages"`
	Stream   bool                           `json:"stream"`
	Tools    []openai.Tool                  `json:"tools"`
}

func newMockChatServer(
	t *testing.T,
	handler func(req capturedChatRequest) string,
) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		var req capturedChatRequest
		if err := json.Unmarshal(body, &req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, handler(req))
	}))
}

func sseAssistantStop(content string) string {
	contentJSON, err := json.Marshal(content)
	if err != nil {
		panic(err)
	}
	return "" +
		`data: {"id":"cmpl-mock","object":"chat.completion.chunk","created":1,"model":"mock","choices":[{"index":0,"delta":{"role":"assistant","content":` +
		string(contentJSON) +
		`},"finish_reason":null}]}` + "\n\n" +
		`data: {"id":"cmpl-mock","object":"chat.completion.chunk","created":1,"model":"mock","choices":[{"index":0,"delta":{},"finish_reason":"stop"}]}` + "\n\n" +
		"data: [DONE]\n\n"
}

func sseToolCall(id, name, args string) string {
	return fmt.Sprintf(
		`data: {"id":"cmpl-mock","object":"chat.completion.chunk","created":1,"model":"mock","choices":[{"index":0,"delta":{"role":"assistant","tool_calls":[{"index":0,"id":%q,"type":"function","function":{"name":%q,"arguments":%q}}]},"finish_reason":"tool_calls"}]}`+"\n\n"+
			"data: [DONE]\n\n",
		id, name, args,
	)
}

func mockModelConfig(baseURL string) utils.ModelConfig {
	return utils.ModelConfig{
		Model:   "mock-model",
		BaseURL: baseURL,
		APIKey:  "mock-key",
	}
}

type ctxKey struct{}

func TestWithPreRequestCallback(t *testing.T) {
	cv("未设置回调时, 原始消息应原样发给模型", t, func() {
		var got capturedChatRequest
		srv := newMockChatServer(t, func(req capturedChatRequest) string {
			got = req
			return sseAssistantStop("pong")
		})
		defer srv.Close()

		req := []openai.ChatCompletionMessage{{
			Role:    openai.ChatMessageRoleUser,
			Content: "hello",
		}}
		rsp, err := utils.Process(context.Background(), mockModelConfig(srv.URL), req)
		so(err, isNil)
		so(len(got.Messages), eq, 1)
		so(got.Messages[0].Role, eq, openai.ChatMessageRoleUser)
		so(got.Messages[0].Content, eq, "hello")
		so(len(rsp.Messages), eq, 2)
		so(rsp.Messages[1].Content, eq, "pong")
	})

	cv("回调可原地修改底层 messages, 无需整体替换切片", t, func() {
		var got capturedChatRequest
		srv := newMockChatServer(t, func(req capturedChatRequest) string {
			got = req
			return sseAssistantStop("inplace-ok")
		})
		defer srv.Close()

		req := []openai.ChatCompletionMessage{{
			Role:    openai.ChatMessageRoleUser,
			Content: "原始提问",
		}}
		rsp, err := utils.Process(
			context.Background(),
			mockModelConfig(srv.URL),
			req,
			utils.WithPreRequestCallback(func(_ context.Context, params *utils.PreRequestContext) error {
				so(len(params.RequestMessages), eq, 1)
				params.RequestMessages[0].Content = "原地改写"
				return nil
			}),
		)
		so(err, isNil)
		so(len(got.Messages), eq, 1)
		so(got.Messages[0].Content, eq, "原地改写")
		so(rsp.Messages[0].Content, eq, "原地改写")
		so(rsp.Messages[1].Content, eq, "inplace-ok")
	})

	cv("回调可以整体替换消息切片, 实际发出的请求应使用新切片", t, func() {
		var got capturedChatRequest
		srv := newMockChatServer(t, func(req capturedChatRequest) string {
			got = req
			return sseAssistantStop("rewritten-ok")
		})
		defer srv.Close()

		req := []openai.ChatCompletionMessage{{
			Role:    openai.ChatMessageRoleUser,
			Content: "原始提问",
		}}
		rsp, err := utils.Process(
			context.Background(),
			mockModelConfig(srv.URL),
			req,
			utils.WithPreRequestCallback(func(_ context.Context, params *utils.PreRequestContext) error {
				so(len(params.RequestMessages), eq, 1)
				so(params.RequestMessages[0].Content, eq, "原始提问")
				params.RequestMessages = []openai.ChatCompletionMessage{{
					Role:    openai.ChatMessageRoleSystem,
					Content: "injected-system",
				}, {
					Role:    openai.ChatMessageRoleUser,
					Content: "改写后的提问",
				}}
				return nil
			}),
		)
		so(err, isNil)
		so(len(got.Messages), eq, 2)
		so(got.Messages[0].Role, eq, openai.ChatMessageRoleSystem)
		so(got.Messages[0].Content, eq, "injected-system")
		so(got.Messages[1].Role, eq, openai.ChatMessageRoleUser)
		so(got.Messages[1].Content, eq, "改写后的提问")
		so(rsp.Messages[0].Content, eq, "injected-system")
		so(rsp.Messages[1].Content, eq, "改写后的提问")
		so(rsp.Messages[2].Content, eq, "rewritten-ok")
	})

	cv("回调返回错误时应中断请求并包装错误, 且不应真正发往模型", t, func() {
		var hit atomic.Int32
		srv := newMockChatServer(t, func(capturedChatRequest) string {
			hit.Add(1)
			return sseAssistantStop("should-not-happen")
		})
		defer srv.Close()

		wantErr := errors.New("故意失败")
		req := []openai.ChatCompletionMessage{{
			Role:    openai.ChatMessageRoleUser,
			Content: "hello",
		}}
		rsp, err := utils.Process(
			context.Background(),
			mockModelConfig(srv.URL),
			req,
			utils.WithPreRequestCallback(func(context.Context, *utils.PreRequestContext) error {
				return wantErr
			}),
		)
		so(err, notNil)
		so(errors.Is(err, wantErr), eq, true)
		so(strings.Contains(err.Error(), "发送请求前回调失败"), eq, true)
		so(strings.Contains(err.Error(), "问答错误"), eq, true)
		so(hit.Load(), eq, int32(0))
		so(len(rsp.Messages), eq, 0)
		so(len(rsp.FullConversation), eq, 0)
	})

	cv("Process 入口会拷贝入参切片, 回调改的是内部 messages, 不回写调用方原切片", t, func() {
		srv := newMockChatServer(t, func(capturedChatRequest) string {
			return sseAssistantStop("ok")
		})
		defer srv.Close()

		req := []openai.ChatCompletionMessage{{
			Role:    openai.ChatMessageRoleUser,
			Content: "原始提问",
		}}
		_, err := utils.Process(
			context.Background(),
			mockModelConfig(srv.URL),
			req,
			utils.WithPreRequestCallback(func(_ context.Context, params *utils.PreRequestContext) error {
				params.RequestMessages[0].Content = "已被回调改写"
				params.RequestMessages = append(params.RequestMessages, openai.ChatCompletionMessage{
					Role:    openai.ChatMessageRoleAssistant,
					Content: "追加消息",
				})
				return nil
			}),
		)
		so(err, isNil)
		so(len(req), eq, 1)
		so(req[0].Content, eq, "原始提问")
	})

	cv("回调应收到调用方传入的 context", t, func() {
		srv := newMockChatServer(t, func(capturedChatRequest) string {
			return sseAssistantStop("ok")
		})
		defer srv.Close()

		const marker = "from-caller"
		ctx := context.WithValue(context.Background(), ctxKey{}, marker)
		var got any
		req := []openai.ChatCompletionMessage{{
			Role:    openai.ChatMessageRoleUser,
			Content: "hello",
		}}
		_, err := utils.Process(
			ctx,
			mockModelConfig(srv.URL),
			req,
			utils.WithPreRequestCallback(func(cbCtx context.Context, _ *utils.PreRequestContext) error {
				got = cbCtx.Value(ctxKey{})
				return nil
			}),
		)
		so(err, isNil)
		so(got, eq, marker)
	})

	cv("传入 nil 回调不应覆盖已设置的回调", t, func() {
		var called atomic.Int32
		srv := newMockChatServer(t, func(capturedChatRequest) string {
			return sseAssistantStop("ok")
		})
		defer srv.Close()

		req := []openai.ChatCompletionMessage{{
			Role:    openai.ChatMessageRoleUser,
			Content: "hello",
		}}
		_, err := utils.Process(
			context.Background(),
			mockModelConfig(srv.URL),
			req,
			utils.WithPreRequestCallback(func(context.Context, *utils.PreRequestContext) error {
				called.Add(1)
				return nil
			}),
			utils.WithPreRequestCallback(nil),
		)
		so(err, isNil)
		so(called.Load(), eq, int32(1))
	})

	cv("空消息时应在回调之前失败, 回调不应被调用", t, func() {
		var called atomic.Int32
		srv := newMockChatServer(t, func(capturedChatRequest) string {
			return sseAssistantStop("ok")
		})
		defer srv.Close()

		_, err := utils.Process(
			context.Background(),
			mockModelConfig(srv.URL),
			nil,
			utils.WithPreRequestCallback(func(context.Context, *utils.PreRequestContext) error {
				called.Add(1)
				return nil
			}),
		)
		so(err, notNil)
		so(strings.Contains(err.Error(), "没有待请求的消息"), eq, true)
		so(called.Load(), eq, int32(0))
	})

	cv("工具调用多轮迭代时, 每一轮向模型发请求前都应触发回调", t, func() {
		var (
			mu       sync.Mutex
			roundN   int
			captured [][]openai.ChatCompletionMessage
		)
		srv := newMockChatServer(t, func(req capturedChatRequest) string {
			mu.Lock()
			n := roundN
			roundN++
			mu.Unlock()
			if n == 0 {
				return sseToolCall("call_1", "mock-echo:echo", `{"x":1}`)
			}
			return sseAssistantStop("tool-done")
		})
		defer srv.Close()

		req := []openai.ChatCompletionMessage{{
			Role:    openai.ChatMessageRoleUser,
			Content: "请调用工具",
		}}
		rsp, err := utils.Process(
			context.Background(),
			mockModelConfig(srv.URL),
			req,
			utils.WithInitializedMCP(&echoMCP{}, "mock-echo"),
			utils.WithPreRequestCallback(func(_ context.Context, params *utils.PreRequestContext) error {
				copied := append([]openai.ChatCompletionMessage{}, params.RequestMessages...)
				mu.Lock()
				captured = append(captured, copied)
				mu.Unlock()
				return nil
			}),
		)
		so(err, isNil)
		so(len(captured), eq, 2)
		so(len(captured[0]), eq, 1)
		so(captured[0][0].Content, eq, "请调用工具")
		so(len(captured[1]) >= 3, eq, true) // user + assistant(tool_calls) + tool result
		so(captured[1][0].Content, eq, "请调用工具")
		so(captured[1][1].Role, eq, openai.ChatMessageRoleAssistant)
		so(len(captured[1][1].ToolCalls), eq, 1)
		so(captured[1][1].ToolCalls[0].Function.Name, eq, "mock-echo:echo")
		so(captured[1][2].Role, eq, openai.ChatMessageRoleTool)
		so(captured[1][2].Content, eq, "echo-ok")
		so(rsp.Messages[len(rsp.Messages)-1].Content, eq, "tool-done")
		so(len(rsp.FullConversation), eq, 2)
	})
}

type echoMCP struct{}

func (*echoMCP) ListTools(context.Context, mcp.ListToolsRequest) (*mcp.ListToolsResult, error) {
	return &mcp.ListToolsResult{
		Tools: []mcp.Tool{{
			Name:        "echo",
			Description: "回显工具",
		}},
	}, nil
}

func (*echoMCP) CallTool(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return utils.NewMCPCallToolResultWithString("echo-ok")
}
