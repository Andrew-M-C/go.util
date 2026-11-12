package openai_test

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"sync"
	"testing"

	utils "github.com/Andrew-M-C/go.util/openai"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/sashabaranov/go-openai"
)

// singleTool 是一个最小化的独立工具，记录 CallTool 收到的参数。
type singleTool struct {
	def    mcp.Tool
	result string

	mu     sync.Mutex
	called []mcp.CallToolParams
}

func (s *singleTool) GetTool() mcp.Tool {
	return s.def
}

func (s *singleTool) CallTool(_ context.Context, request mcp.CallToolParams) (*mcp.CallToolResult, error) {
	s.mu.Lock()
	s.called = append(s.called, request)
	s.mu.Unlock()
	return utils.NewMCPCallToolResultWithString(s.result)
}

func (s *singleTool) calls() []mcp.CallToolParams {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]mcp.CallToolParams(nil), s.called...)
}

func TestWithTool(t *testing.T) {
	cv("WithTool", t, func() {
		cv("用工具名作为 MCP id，并把调用参数交回该工具", func() {
			tool := &singleTool{
				def: mcp.Tool{
					Name:        "echo",
					Description: "回声",
				},
				result: "pong",
			}

			var (
				mu   sync.Mutex
				reqs []capturedChatRequest
			)
			srv := newMockChatServer(t, func(req capturedChatRequest) string {
				mu.Lock()
				reqs = append(reqs, req)
				n := len(reqs)
				mu.Unlock()
				if n == 1 {
					return sseToolCall("call_echo_1", "echo__echo", `{"text":"hi"}`)
				}
				return sseAssistantStop("done")
			})
			defer srv.Close()

			rsp, err := utils.Process(
				context.Background(),
				mockModelConfig(srv.URL),
				[]openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleUser, Content: "hi"}},
				utils.WithTool(tool),
			)
			so(err, isNil)

			mu.Lock()
			gotReqs := append([]capturedChatRequest(nil), reqs...)
			mu.Unlock()

			so(len(gotReqs), eq, 2)
			so(len(gotReqs[0].Tools), eq, 1)
			so(gotReqs[0].Tools[0].Function, notNil)
			so(gotReqs[0].Tools[0].Function.Name, eq, "echo__echo")
			so(gotReqs[0].Tools[0].Function.Description, eq, "回声")

			calls := tool.calls()
			so(len(calls), eq, 1)
			so(calls[0].Name, eq, "echo")
			args, _ := calls[0].Arguments.(map[string]any)
			so(args["text"], eq, "hi")

			so(len(gotReqs[1].Messages), eq, 3)
			so(gotReqs[1].Messages[2].Role, eq, openai.ChatMessageRoleTool)
			so(gotReqs[1].Messages[2].Content, eq, "pong")
			so(gotReqs[1].Messages[2].ToolCallID, eq, "call_echo_1")
			so(rsp.Messages[len(rsp.Messages)-1].Content, eq, "done")
		})

		cv("工具名里的空格和 __ 只改写 id 前缀", func() {
			const rawName = "foo bar__baz"
			tool := &singleTool{
				def:    mcp.Tool{Name: rawName, Description: "sanitize"},
				result: "ok",
			}

			var (
				mu   sync.Mutex
				reqs []capturedChatRequest
			)
			srv := newMockChatServer(t, func(req capturedChatRequest) string {
				mu.Lock()
				reqs = append(reqs, req)
				n := len(reqs)
				mu.Unlock()
				if n == 1 {
					return sseToolCall("call_s", "foo-bar-baz__foo bar__baz", `{}`)
				}
				return sseAssistantStop("done")
			})
			defer srv.Close()

			_, err := utils.Process(
				context.Background(),
				mockModelConfig(srv.URL),
				[]openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleUser, Content: "go"}},
				utils.WithTool(tool),
			)
			so(err, isNil)

			mu.Lock()
			gotReqs := append([]capturedChatRequest(nil), reqs...)
			mu.Unlock()

			so(len(gotReqs), eq, 2)
			so(gotReqs[0].Tools[0].Function.Name, eq, "foo-bar-baz__foo bar__baz")
			calls := tool.calls()
			so(len(calls), eq, 1)
			so(calls[0].Name, eq, rawName)
		})

		cv("清洗后的工具名与其他 MCP id 冲突时报 MCP ID 重复", func() {
			_, err := utils.Process(
				context.Background(),
				utils.ModelConfig{Model: "mock-model", BaseURL: "http://127.0.0.1:1", APIKey: "k"},
				[]openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleUser, Content: "hi"}},
				utils.WithInitializedMCP(&simpleMCP{toolName: "other", toolResult: "x"}, "foo-bar"),
				utils.WithTool(&singleTool{def: mcp.Tool{Name: "foo bar"}}),
			)
			so(err, notNil)
			so(err.Error(), eq, "MCP ID 重复 (foo-bar)")
		})

		cv("空工具名报 MCP ID 为空", func() {
			for _, name := range []string{"", "  "} {
				_, err := utils.Process(
					context.Background(),
					utils.ModelConfig{Model: "mock-model", BaseURL: "http://127.0.0.1:1", APIKey: "k"},
					[]openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleUser, Content: "hi"}},
					utils.WithTool(&singleTool{def: mcp.Tool{Name: name}}),
				)
				so(err, notNil)
				so(err.Error(), eq, "MCP ID 为空")
			}
		})

		cv("工具返回图片时，下一轮请求带上图片内容", func() {
			tool := &userDrawingTool{data: testPNG}
			var (
				mu   sync.Mutex
				reqs []capturedChatRequest
			)
			srv := newMockChatServer(t, func(req capturedChatRequest) string {
				mu.Lock()
				reqs = append(reqs, req)
				n := len(reqs)
				mu.Unlock()
				if n == 1 {
					return sseToolCall("call_img_1", "read_user_drawing__read_user_drawing", `{}`)
				}
				return sseAssistantStop("是一座房子")
			})
			defer srv.Close()

			_, err := utils.Process(
				context.Background(),
				mockModelConfig(srv.URL),
				[]openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleUser, Content: "看看我画的图片是什么"}},
				utils.WithTool(tool),
			)
			so(err, isNil)
			so(tool.calls(), eq, 1)

			mu.Lock()
			gotReqs := append([]capturedChatRequest(nil), reqs...)
			mu.Unlock()
			so(len(gotReqs), eq, 2)
			so(len(gotReqs[1].Messages), eq, 3)
			toolMsg := gotReqs[1].Messages[2]
			so(toolMsg.Role, eq, openai.ChatMessageRoleTool)
			so(toolMsg.ToolCallID, eq, "call_img_1")
			so(len(toolMsg.MultiContent), eq, 2)
			so(toolMsg.MultiContent[0].Type, eq, openai.ChatMessagePartTypeText)
			so(toolMsg.MultiContent[0].Text, eq, "用户创作的图片")
			so(toolMsg.MultiContent[1].Type, eq, openai.ChatMessagePartTypeImageURL)
			so(toolMsg.MultiContent[1].ImageURL, notNil)
			got, decErr := base64.StdEncoding.DecodeString(strings.TrimPrefix(
				toolMsg.MultiContent[1].ImageURL.URL, "data:image/png;base64,",
			))
			so(decErr, isNil)
			so(string(got), eq, string(testPNG))
		})

		cv("同名工具重复注册报 MCP ID 重复", func() {
			_, err := utils.Process(
				context.Background(),
				utils.ModelConfig{Model: "mock-model", BaseURL: "http://127.0.0.1:1", APIKey: "k"},
				[]openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleUser, Content: "hi"}},
				utils.WithTool(&singleTool{def: mcp.Tool{Name: "echo"}}),
				utils.WithTool(&singleTool{def: mcp.Tool{Name: "echo"}}),
			)
			so(err, notNil)
			so(err.Error(), eq, "MCP ID 重复 (echo)")
		})
	})
}

// userDrawingTool 在被调用时返回用户创作的图片。
type userDrawingTool struct {
	data []byte

	mu     sync.Mutex
	called int
}

func (t *userDrawingTool) GetTool() mcp.Tool {
	return mcp.NewTool("read_user_drawing",
		mcp.WithDescription("读取用户创作的图片。当用户想查看自己画的内容时调用，调用后返回这张图片。"),
	)
}

func (t *userDrawingTool) CallTool(context.Context, mcp.CallToolParams) (*mcp.CallToolResult, error) {
	t.mu.Lock()
	t.called++
	t.mu.Unlock()
	return mcp.NewToolResultImage("用户创作的图片", base64.StdEncoding.EncodeToString(t.data), "image/png"), nil
}

func (t *userDrawingTool) calls() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.called
}

func TestReadUserDrawing(t *testing.T) {
	cfg, ok := readDeepSeekConfig()
	if !ok {
		t.Log("不予测试")
		return
	}

	cv("看看我画的图片是什么", t, func() {
		tool := &userDrawingTool{data: testPNG}
		content := func(c string) { fmt.Printf("%s", c) }
		rsp, err := utils.Process(
			context.Background(),
			cfg,
			[]openai.ChatCompletionMessage{{
				Role:    openai.ChatMessageRoleUser,
				Content: "看看我画的图片是什么",
			}},
			utils.WithContentCallback(content),
			utils.WithTool(tool),
		)
		so(err, isNil)
		so(rsp, notNil)
		so(tool.calls(), eq, 1)

		answer := rsp.Messages[len(rsp.Messages)-1].Content
		printf("获得响应: %v", answer)
		so(strings.Contains(answer, "房"), eq, true)
	})
}
