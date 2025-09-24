package openai

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Andrew-M-C/go.util/unsafe"
	"github.com/h2non/filetype"
	"github.com/h2non/filetype/types"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/sashabaranov/go-openai"
)

// -------- 图片相关 --------

// AddImageURLToLastMessage 为最后一个消息添加图片 URL, 可能修改 message 中的数据
//
// 如果 messages 中没有成员, 会报错退出
func AddImageURLToLastMessage(
	messages []openai.ChatCompletionMessage, url string,
) ([]openai.ChatCompletionMessage, error) {
	return addImageURLToLastMessage(messages, url)
}

// AddImageDataToLastMessage 为最后一个消息添加图片数据, 可能修改 message 中的数据。
//
// 如果 messages 中没有成员, 会报错退出
func AddImageDataToLastMessage(
	messages []openai.ChatCompletionMessage, data []byte,
) ([]openai.ChatCompletionMessage, error) {
	mime, err := analyzeFileData(data)
	if err != nil {
		return messages, err
	}
	if mime.Type != "image" {
		return messages, fmt.Errorf("data is not an image (got '%s')", mime.Type)
	}

	// 添加数据
	imageURL := "data:image/png;base64," + base64.StdEncoding.EncodeToString(data)
	return addImageURLToLastMessage(messages, imageURL)
}

func analyzeFileData(data []byte) (types.MIME, error) {
	typ, err := filetype.Match(data)
	if err != nil {
		return types.MIME{}, fmt.Errorf("分析文件数据失败 (%w)", err)
	}
	return typ.MIME, nil
}

func addImageURLToLastMessage(
	messages []openai.ChatCompletionMessage, imageURL string,
) ([]openai.ChatCompletionMessage, error) {
	contentItem := openai.ChatMessagePart{
		Type:     openai.ChatMessagePartTypeImageURL,
		ImageURL: &openai.ChatMessageImageURL{URL: imageURL},
	}
	if len(messages) == 0 {
		return append(messages, openai.ChatCompletionMessage{
			Role:         openai.ChatMessageRoleUser,
			MultiContent: []openai.ChatMessagePart{contentItem},
		}), nil
	}

	// 添加数据
	lastMsg := messages[len(messages)-1]
	if lastMsg.Content != "" && len(lastMsg.MultiContent) > 0 {
		return messages, errors.New("最后一个消息的 Content 和 MultiContent 字段同时存在, 请检查")
	}

	// 最后一个消息没有内容, 那么直接 append 一个就行了
	if lastMsg.Content == "" && len(lastMsg.MultiContent) == 0 {
		lastMsg.MultiContent = append(lastMsg.MultiContent, openai.ChatMessagePart{
			Type:     openai.ChatMessagePartTypeImageURL,
			ImageURL: &openai.ChatMessageImageURL{URL: imageURL},
		})
		messages[len(messages)-1] = lastMsg
		return messages, nil
	}

	// 最后一个消息有内容, 那么转换为 multi content
	if lastMsg.Content != "" {
		lastMsg.MultiContent = append(lastMsg.MultiContent,
			openai.ChatMessagePart{
				Type: openai.ChatMessagePartTypeText,
				Text: lastMsg.Content,
			},
			openai.ChatMessagePart{
				Type:     openai.ChatMessagePartTypeImageURL,
				ImageURL: &openai.ChatMessageImageURL{URL: imageURL},
			},
		)
		lastMsg.Content = ""
		messages[len(messages)-1] = lastMsg
		return messages, nil
	}

	// 没有内容的话直接 append 就行了
	lastMsg.MultiContent = append(lastMsg.MultiContent, openai.ChatMessagePart{
		Type:     openai.ChatMessagePartTypeImageURL,
		ImageURL: &openai.ChatMessageImageURL{URL: imageURL},
	})
	messages[len(messages)-1] = lastMsg
	return messages, nil
}

// -------- prompt 相关 --------

// AddOrSetPromptForMessages 添加或设置整个上下文的。如果 prompt 不存在则设置, 存在则替换
func AddOrSetPromptForMessages(
	messages []openai.ChatCompletionMessage, prompt string,
) []openai.ChatCompletionMessage {
	// 替换模式
	if len(messages) > 0 && (messages[0].Role == openai.ChatMessageRoleSystem) {
		messages[0].Content = prompt
		return messages
	}

	// 添加模式
	res := make([]openai.ChatCompletionMessage, 0, len(messages)+1)
	res = append(res, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: prompt,
	})
	res = append(res, messages...)

	return res
}

// -------- MCP 相关 --------

// ReadMCPCallToolRequest 读取 MCP 工具调用参数, 方便实现 InitializedMCPClient 接口。
func ReadMCPCallToolRequest[T any](req mcp.CallToolRequest) (T, error) {
	var t T
	b, err := json.Marshal(req.Params.Arguments)
	if err != nil {
		return t, err
	}
	err = json.Unmarshal(b, &t)
	return t, err
}

// NewMCPCallToolResultWithString 生成一个 MCP 工具调用结果, 用于适配 InitializedMCPClient 接口
func NewMCPCallToolResultWithString(s string) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultText(s), nil
}

// NewMCPCallToolResultWithJSON 生成一个 MCP 工具调用结果, 使用 json.Marshal 之后的数据
// 打包为 JSON 数据, 用于适配 InitializedMCPClient 接口
func NewMCPCallToolResultWithJSON(v any) (*mcp.CallToolResult, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	s := unsafe.BtoS(b)
	return mcp.NewToolResultText(s), nil
}

// MCPToolParamWithString 创建一个 MCP string 参数构建器, 用于构建 MCP tool 列表参数
func MCPToolParamWithString(name string, required bool, desc string) MCPToolBuilder {
	var res mcpToolBuilder
	if required {
		res.opt = mcp.WithString(name, mcp.Description(desc), mcp.Required())
	} else {
		res.opt = mcp.WithString(name, mcp.Description(desc))
	}
	return res
}

// MCPToolParamWithNumber 创建一个 MCP 数字参数构建器, 用于构建 MCP tool 列表参数
func MCPToolParamWithNumber(name string, required bool, desc string) MCPToolBuilder {
	var res mcpToolBuilder
	if required {
		res.opt = mcp.WithNumber(name, mcp.Description(desc), mcp.Required())
	} else {
		res.opt = mcp.WithNumber(name, mcp.Description(desc))
	}
	return res
}

// MCPToolParamWithBoolean 创建一个 MCP boolean 参数构建器, 用于构建 MCP tool 列表参数
func MCPToolParamWithBoolean(name string, required bool, desc string) MCPToolBuilder {
	var res mcpToolBuilder
	if required {
		res.opt = mcp.WithBoolean(name, mcp.Description(desc), mcp.Required())
	} else {
		res.opt = mcp.WithBoolean(name, mcp.Description(desc))
	}
	return res
}

type MCPToolBuilder interface {
	AddToTool(tool *mcp.Tool)
}

type mcpToolBuilder struct {
	opt mcp.ToolOption
}

func (b mcpToolBuilder) AddToTool(t *mcp.Tool) {
	if t == nil {
		return
	}
	if t.InputSchema.Properties == nil {
		t.InputSchema.Properties = map[string]any{}
	}
	b.opt(t)
}
