package openai

import (
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/h2non/filetype"
	"github.com/h2non/filetype/types"
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
