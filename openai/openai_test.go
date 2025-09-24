package openai_test

import (
	"context"
	_ "embed"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Andrew-M-C/go-bytesize"
	hutil "github.com/Andrew-M-C/go.util/net/http"
	utils "github.com/Andrew-M-C/go.util/openai"
	"github.com/Andrew-M-C/go.util/unsafe"
	"github.com/fatih/color"
	"github.com/mark3labs/mcp-go/mcp"
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
		req := []openai.ChatCompletionMessage{{
			Role:    openai.ChatMessageRoleUser,
			Content: "请告诉我现在几点以及广州市今天的天气",
			// Content: "现在几点了？",
		}}

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

	cv("下载长篇文章, 测试大模型 token 超限", t, func() {
		if true {
			printf("暂时取消测试")
			return
		}
		ctx := context.Background()
		config := utils.ModelConfig{
			Model:   deepseekModel,
			BaseURL: deepseekBaseURL,
			APIKey:  deepseekAPIKey,
		}
		data, err := hutil.Raw(ctx, "https://www.gutenberg.org/cache/epub/1184/pg1184.txt")
		so(err, isNil)
		text := unsafe.BtoS(data)
		printf("全文 %v", bytesize.Base10(len(data)))

		req := []openai.ChatCompletionMessage{{
			Role:    openai.ChatMessageRoleUser,
			Content: text,
		}}
		req = utils.AddOrSetPromptForMessages(req, "请简要介绍一下这部小说")
		reasoning := func(c string) { fmt.Printf("%s", color.BlueString(c)) }
		content := func(c string) { fmt.Printf("%s", c) }
		finish := func(f openai.FinishReason) { printf("结束: %v", f) }

		_, err = utils.Process(ctx, config, req,
			utils.WithContentCallback(content),
			utils.WithReasoningCallback(reasoning),
			utils.WithFinishCallback(finish),
		)
		so(err, notNil)
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

func TestInitializedMCP(t *testing.T) {
	cv("时间 + 天气两个 MCP", t, func() {
		ctx := context.Background()
		config := utils.ModelConfig{
			Model:   deepseekModel,
			BaseURL: deepseekBaseURL,
			APIKey:  deepseekAPIKey,
		}
		req := []openai.ChatCompletionMessage{{
			Role:    openai.ChatMessageRoleUser,
			Content: "请告诉我现在广州24小时制的 HH:MM 格式时间, 以及天气",
		}}

		reasoning := func(c string) { fmt.Printf("%s", color.BlueString(c)) }
		content := func(c string) { fmt.Printf("%s", c) }
		finish := func(f openai.FinishReason) { printf("阶段性结束: %v\n\n", f) }

		weatherMCP := &weatherMCP{}
		timeMCP := &timeMCP{}

		rsp, err := utils.Process(ctx, config, req,
			// utils.WithDebugger(printf),
			utils.WithContentCallback(content),
			utils.WithReasoningCallback(reasoning),
			utils.WithFinishCallback(finish),
			utils.WithInitializedMCP(weatherMCP),
			utils.WithInitializedMCP(timeMCP),
		)
		so(err, isNil)
		so(rsp, notNil)
		// so(len(rsp.Messages), eq, 5) // 1问、1答、2工具调用、1答。但是有时候 LLM 会拆分成两次, 不一定
		so(weatherMCP.Count, eq, 1)
		so(timeMCP.Count, eq, 1)

		s := rsp.Messages[4].Content
		so(strings.Contains(s, "暴雨"), eq, true)

		guangzhouTime := timeMCP.Time.Add(8 * time.Hour).UTC().Format("15:04") // 似乎需要 deepseek-r1 才知道要进行时区转换
		printf("预期获得广州时间: %s", guangzhouTime)
		so(strings.Contains(s, guangzhouTime), eq, true)
	})
}

type weatherMCP struct {
	Count int
}

func (*weatherMCP) ListTools(ctx context.Context, _ mcp.ListToolsRequest) (*mcp.ListToolsResult, error) {
	weatherTool := mcp.Tool{
		Name:        "weather",
		Description: "获取当前天气状况",
	}
	utils.MCPToolParamWithString("location", true, "地区描述").AddToTool(&weatherTool)

	return &mcp.ListToolsResult{
		Tools: []mcp.Tool{weatherTool},
	}, nil
}

func (w *weatherMCP) CallTool(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	w.Count++
	// 无视请求, 固定返回
	return utils.NewMCPCallToolResultWithString("狂风暴雨")
}

type timeMCP struct {
	Time  time.Time
	Count int
}

func (*timeMCP) ListTools(ctx context.Context, _ mcp.ListToolsRequest) (*mcp.ListToolsResult, error) {
	tmTool := mcp.Tool{
		Name:        "time",
		Description: "获取当前的 UTC 时间",
	}
	return &mcp.ListToolsResult{
		Tools: []mcp.Tool{tmTool},
	}, nil
}

func (t *timeMCP) CallTool(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	t.Count++
	t.Time = time.Now().UTC()
	desc := t.Time.Format(time.DateTime)
	return utils.NewMCPCallToolResultWithString(fmt.Sprintf("伦敦时间 %s", desc))
}
