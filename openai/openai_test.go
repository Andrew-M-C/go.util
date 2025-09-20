package openai_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	utils "github.com/Andrew-M-C/go.util/openai"
	"github.com/fatih/color"
	"github.com/sashabaranov/go-openai"
	"github.com/smartystreets/goconvey/convey"
)

var (
	cv = convey.Convey
	so = convey.So
	eq = convey.ShouldEqual

	isNil  = convey.ShouldBeNil
	notNil = convey.ShouldNotBeNil

	// 测试环境变量
	deepseekModel   = ""
	deepseekBaseURL = ""
	deepseekAPIKey  = ""
	deepseekMCPURL  = ""
)

func TestMain(m *testing.M) {
	if !readEnv() {
		fmt.Println("测试环境变量未设置, 不进行测试")
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func readEnv() bool {
	if deepseekBaseURL = os.Getenv("DEEPSEEK_BASE_URL"); deepseekBaseURL == "" {
		return false
	}
	if deepseekAPIKey = os.Getenv("DEEPSEEK_API_KEY"); deepseekAPIKey == "" {
		return false
	}
	if deepseekModel = os.Getenv("DEEPSEEK_MODEL"); deepseekModel == "" {
		deepseekModel = "deepseek-reasoning"
	}
	if deepseekMCPURL = os.Getenv("DEEPSEEK_MCP_URL"); deepseekMCPURL == "" {
		return false
	}
	return true
}

func printf(s string, a ...any) {
	_, _ = fmt.Printf(s+"\n", a...)
}

func TestProcess(t *testing.T) {
	cv("简单对话", t, func() {
		ctx := context.Background()
		config := utils.ModelConfig{
			Model:   deepseekModel,
			BaseURL: deepseekBaseURL,
			APIKey:  deepseekAPIKey,
		}
		req := []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "你喜欢简短地回答, 稳重、不废话",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: "你好",
			},
		}

		reasoningBuilder := strings.Builder{}
		contentBuilder := strings.Builder{}

		reasoning := func(c string) {
			fmt.Printf("%s", color.BlueString(c))
			reasoningBuilder.WriteString(c)
		}
		content := func(c string) {
			fmt.Printf("%s", c)
			contentBuilder.WriteString(c)
		}

		finishCalled := false
		finish := func(f openai.FinishReason) {
			printf("结束: %v", f)
			finishCalled = true
		}

		rsp, err := utils.Process(ctx, config, req,
			utils.WithDebugger(printf),
			utils.WithContentCallback(content),
			utils.WithReasoningCallback(reasoning),
			utils.WithFinishCallback(finish),
		)
		so(err, isNil)
		so(rsp, notNil)
		so(len(rsp.Messages), eq, 3)
		printf("获得思考: %v", rsp.Messages[2].ReasoningContent)
		printf("获得响应: %v", rsp.Messages[2].Content)
		so(reasoningBuilder.String(), eq, rsp.Messages[2].ReasoningContent)
		so(contentBuilder.String(), eq, rsp.Messages[2].Content)
		so(finishCalled, eq, true)
	})

	cv("带两次 MCP 的请求", t, func() {
		ctx := context.Background()
		config := utils.ModelConfig{
			Model:   deepseekModel,
			BaseURL: deepseekBaseURL,
			APIKey:  deepseekAPIKey,
		}
		req := []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: "请告诉我现在几点以及广州市今天的天气",
				// Content: "现在几点了？",
			},
		}

		reasoning := func(c string) { fmt.Printf("%s", color.BlueString(c)) }
		content := func(c string) { fmt.Printf("%s", c) }
		finish := func(f openai.FinishReason) { printf("阶段性结束: %v\n\n", f) }

		tcStart := func(tc openai.ToolCall) {
			printf("%s\n", color.YellowString("工具调用: %s", tc.Function.Name))
		}
		tcEnds := func(m openai.ChatCompletionMessage) {
			printf("%s\n", color.GreenString("工具调用结束: %s", m.Content))
		}

		rsp, err := utils.Process(ctx, config, req,
			// utils.WithDebugger(printf),
			utils.WithContentCallback(content),
			utils.WithReasoningCallback(reasoning),
			utils.WithFinishCallback(finish),
			utils.WithToolCallRequestCallback(tcStart),
			utils.WithToolCallResponseCallback(tcEnds),
			utils.WithRemoteMCP(deepseekMCPURL),
		)
		so(err, isNil)
		so(rsp, notNil)
		so(len(rsp.Messages), eq, 5) // 1问、1答、2工具调用、1答
	})
}
