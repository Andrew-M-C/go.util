package openai_test

import (
	"context"
	_ "embed"
	"encoding/base64"
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
	hunyuanAPIKey   = ""
)

//go:embed test_image.png
var testPNG []byte

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
	if hunyuanAPIKey = os.Getenv("HUNYUAN_API_KEY"); hunyuanAPIKey == "" {
		return false
	}
	return true
}

func printf(s string, a ...any) {
	_, _ = fmt.Printf(s+"\n", a...)
}

func TestAddImageDataToLastMessage(t *testing.T) {
	cv("为最后一个消息添加图片数据", t, func() {
		messages := []openai.ChatCompletionMessage{{
			Role:    openai.ChatMessageRoleUser,
			Content: "你好",
		}}
		messages, err := utils.AddImageDataToLastMessage(messages, testPNG)
		so(err, isNil)
		so(len(messages), eq, 1)
		so(messages[0].Content, eq, "")
		so(len(messages[0].MultiContent), eq, 2)
		so(messages[0].MultiContent[0].Type, eq, openai.ChatMessagePartTypeText)
		so(messages[0].MultiContent[0].Text, eq, "你好")
		so(messages[0].MultiContent[1].Type, eq, openai.ChatMessagePartTypeImageURL)

		expectedImageURL := "data:image/png;base64," + base64.StdEncoding.EncodeToString(testPNG)
		so(messages[0].MultiContent[1].ImageURL.URL, eq, expectedImageURL)
	})
}

func TestAddOrSetPromptForMessages(t *testing.T) {
	cv("为空的消息数组添加系统提示词", t, func() {
		messages := []openai.ChatCompletionMessage{}
		prompt := "你是一个有用的助手"

		result := utils.AddOrSetPromptForMessages(messages, prompt)

		so(len(result), eq, 1)
		so(result[0].Role, eq, openai.ChatMessageRoleSystem)
		so(result[0].Content, eq, prompt)
	})

	cv("为已存在系统消息的数组替换提示词", t, func() {
		messages := []openai.ChatCompletionMessage{{
			Role:    openai.ChatMessageRoleSystem,
			Content: "旧的系统提示词",
		}, {
			Role:    openai.ChatMessageRoleUser,
			Content: "用户消息",
		}}
		newPrompt := "新的系统提示词"

		result := utils.AddOrSetPromptForMessages(messages, newPrompt)

		so(len(result), eq, 2)
		so(result[0].Role, eq, openai.ChatMessageRoleSystem)
		so(result[0].Content, eq, newPrompt)
		so(result[1].Role, eq, openai.ChatMessageRoleUser)
		so(result[1].Content, eq, "用户消息")
	})

	cv("为没有系统消息的数组添加系统提示词", t, func() {
		messages := []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: "用户消息1",
			},
			{
				Role:    openai.ChatMessageRoleAssistant,
				Content: "助手回复",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: "用户消息2",
			},
		}
		prompt := "你是一个专业的AI助手"

		result := utils.AddOrSetPromptForMessages(messages, prompt)

		so(len(result), eq, 4)
		so(result[0].Role, eq, openai.ChatMessageRoleSystem)
		so(result[0].Content, eq, prompt)
		so(result[1].Role, eq, openai.ChatMessageRoleUser)
		so(result[1].Content, eq, "用户消息1")
		so(result[2].Role, eq, openai.ChatMessageRoleAssistant)
		so(result[2].Content, eq, "助手回复")
		so(result[3].Role, eq, openai.ChatMessageRoleUser)
		so(result[3].Content, eq, "用户消息2")
	})

	cv("为单个用户消息添加系统提示词", t, func() {
		messages := []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: "你好",
			},
		}
		prompt := "请简洁回答问题"

		result := utils.AddOrSetPromptForMessages(messages, prompt)

		so(len(result), eq, 2)
		so(result[0].Role, eq, openai.ChatMessageRoleSystem)
		so(result[0].Content, eq, prompt)
		so(result[1].Role, eq, openai.ChatMessageRoleUser)
		so(result[1].Content, eq, "你好")
	})

	cv("替换系统消息时不影响原有消息顺序", t, func() {
		messages := []openai.ChatCompletionMessage{{
			Role:    openai.ChatMessageRoleSystem,
			Content: "原系统提示",
		}, {
			Role:    openai.ChatMessageRoleUser,
			Content: "第一个用户消息",
		}, {
			Role:    openai.ChatMessageRoleAssistant,
			Content: "第一个助手回复",
		}}
		newPrompt := "替换后的系统提示"

		result := utils.AddOrSetPromptForMessages(messages, newPrompt)

		so(len(result), eq, 3)
		so(result[0].Role, eq, openai.ChatMessageRoleSystem)
		so(result[0].Content, eq, newPrompt)
		so(result[1].Role, eq, openai.ChatMessageRoleUser)
		so(result[1].Content, eq, "第一个用户消息")
		so(result[2].Role, eq, openai.ChatMessageRoleAssistant)
		so(result[2].Content, eq, "第一个助手回复")

		// 确保原始数组也被修改了（引用传递）
		so(messages[0].Content, eq, newPrompt)
	})

	cv("测试空的提示词字符串", t, func() {
		messages := []openai.ChatCompletionMessage{{
			Role:    openai.ChatMessageRoleUser,
			Content: "用户消息",
		}}
		prompt := ""

		result := utils.AddOrSetPromptForMessages(messages, prompt)

		so(len(result), eq, 2)
		so(result[0].Role, eq, openai.ChatMessageRoleSystem)
		so(result[0].Content, eq, "")
		so(result[1].Role, eq, openai.ChatMessageRoleUser)
		so(result[1].Content, eq, "用户消息")
	})
}

func TestProcessBasic(t *testing.T) {
	cv("简单对话", t, func() {
		ctx := context.Background()
		config := utils.ModelConfig{
			Model:   deepseekModel,
			BaseURL: deepseekBaseURL,
			APIKey:  deepseekAPIKey,
		}
		req := []openai.ChatCompletionMessage{{
			Role:    openai.ChatMessageRoleSystem,
			Content: "你喜欢简短地回答, 稳重、不废话",
		}, {
			Role:    openai.ChatMessageRoleUser,
			Content: "你好",
		}}

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
}

func TestProcessMCP(t *testing.T) {
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
			printf("%s\n", color.BlueString("工具调用: %s", tc.Function.Name))
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

func TestProcessMultiModal(t *testing.T) {
	cv("带图片请求, 混元图生文, 网络链接", t, func() {
		// Reference: https://cloud.tencent.com/document/product/1729/111007
		ctx := context.Background()
		config := utils.ModelConfig{
			Model:   "hunyuan-vision",
			BaseURL: "https://api.hunyuan.cloud.tencent.com/v1/chat/completions",
			APIKey:  hunyuanAPIKey,
		}
		req := []openai.ChatCompletionMessage{
			{
				Role: openai.ChatMessageRoleUser,
				MultiContent: []openai.ChatMessagePart{
					{
						Type: openai.ChatMessagePartTypeText,
						Text: "从这张图中你看到了什么?",
					},
					{
						Type: openai.ChatMessagePartTypeImageURL,
						ImageURL: &openai.ChatMessageImageURL{
							URL: "https://www.baidu.com/img/PCtm_d9c8750bed0b3c7d089fa7d55720d6cf.png",
						},
					},
				},
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
		so(len(rsp.Messages), eq, 2)
		printf("获得思考: %v", rsp.Messages[1].ReasoningContent)
		printf("获得响应: %v", rsp.Messages[1].Content)
		so(reasoningBuilder.String(), eq, rsp.Messages[1].ReasoningContent)
		so(contentBuilder.String(), eq, rsp.Messages[1].Content)
		so(finishCalled, eq, true)
	})

	cv("带图片请求, 混元图生文, 图片二进制数据", t, func() {
		// Reference: https://cloud.tencent.com/document/product/1729/111007
		ctx := context.Background()
		config := utils.ModelConfig{
			Model:   "hunyuan-vision",
			BaseURL: "https://api.hunyuan.cloud.tencent.com/v1/chat/completions",
			APIKey:  hunyuanAPIKey,
		}
		req := []openai.ChatCompletionMessage{
			{
				Role: openai.ChatMessageRoleUser,
				MultiContent: []openai.ChatMessagePart{
					{
						Type: openai.ChatMessagePartTypeText,
						Text: "从这张图中你看到了什么?",
					},
					{
						Type: openai.ChatMessagePartTypeImageURL,
						ImageURL: &openai.ChatMessageImageURL{
							URL: "data:image/png;base64," + base64.StdEncoding.EncodeToString(testPNG),
						},
					},
				},
			},
		}

		content := func(c string) { fmt.Printf("%s", c) }
		rsp, err := utils.Process(ctx, config, req,
			utils.WithContentCallback(content),
		)
		so(err, isNil)
		so(rsp, notNil)
		so(len(rsp.Messages), eq, 2)
		printf("获得响应: %v", rsp.Messages[1].Content)
	})

	cv("带图片请求, 混元图生文,多张图片", t, func() {
		// Reference: https://cloud.tencent.com/document/product/1729/111007
		ctx := context.Background()
		config := utils.ModelConfig{
			Model:   "hunyuan-vision",
			BaseURL: "https://api.hunyuan.cloud.tencent.com/v1/chat/completions",
			APIKey:  hunyuanAPIKey,
		}
		req := []openai.ChatCompletionMessage{
			{
				Role: openai.ChatMessageRoleUser,
				MultiContent: []openai.ChatMessagePart{
					{
						Type: openai.ChatMessagePartTypeText,
						Text: "请分别按顺序说明这三张图片的异同?",
					},
					{
						Type: openai.ChatMessagePartTypeImageURL,
						ImageURL: &openai.ChatMessageImageURL{
							URL: "data:image/png;base64," + base64.StdEncoding.EncodeToString(testPNG),
						},
					},
					{
						Type: openai.ChatMessagePartTypeImageURL,
						ImageURL: &openai.ChatMessageImageURL{
							URL: "https://www.baidu.com/img/PCtm_d9c8750bed0b3c7d089fa7d55720d6cf.png",
						},
					},
					{
						Type: openai.ChatMessagePartTypeImageURL,
						ImageURL: &openai.ChatMessageImageURL{
							URL: "https://qcloudimg.tencent-cloud.cn/raw/42c198dbc0b57ae490e57f89aa01ec23.png",
						},
					},
				},
			},
		}

		content := func(c string) { fmt.Printf("%s", c) }
		rsp, err := utils.Process(ctx, config, req,
			utils.WithContentCallback(content),
		)
		so(err, isNil)
		so(rsp, notNil)
		so(len(rsp.Messages), eq, 2)
		printf("获得响应: %v", rsp.Messages[1].Content)
	})
}
