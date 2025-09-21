package openai

import (
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/h2non/filetype"
	"github.com/h2non/filetype/types"
	"github.com/sashabaranov/go-openai"
)

// AddImageURLToLastMessage 为最后一个消息添加图片 URL, 可能修改 message 中的数据
//
// 如果 messages 中没有成员, 会报错退出
func AddImageURLToLastMessage(messages []openai.ChatCompletionMessage, url string) error {
	return addImageURLToLastMessage(messages, url)
}

// AddImageDataToLastMessage 为最后一个消息添加图片数据, 可能修改 message 中的数据。
//
// 如果 messages 中没有成员, 会报错退出
func AddImageDataToLastMessage(messages []openai.ChatCompletionMessage, data []byte) error {
	mime, err := analyzeFileData(data)
	if err != nil {
		return err
	}
	if mime.Type != "image" {
		return fmt.Errorf("data is not an image (got '%s')", mime.Type)
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

func addImageURLToLastMessage(messages []openai.ChatCompletionMessage, imageURL string) error {
	if len(messages) == 0 {
		return errors.New("messages 中没有成员")
	}

	// 添加数据
	lastMsg := messages[len(messages)-1]
	if lastMsg.Content != "" && len(lastMsg.MultiContent) > 0 {
		return errors.New("最后一个消息的 Content 和 MultiContent 字段同时存在, 请检查")
	}

	// 最后一个消息没有内容, 那么直接 append 一个就行了
	if lastMsg.Content == "" && len(lastMsg.MultiContent) == 0 {
		lastMsg.MultiContent = append(lastMsg.MultiContent, openai.ChatMessagePart{
			Type:     openai.ChatMessagePartTypeImageURL,
			ImageURL: &openai.ChatMessageImageURL{URL: imageURL},
		})
		messages[len(messages)-1] = lastMsg
		return nil
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
		return nil
	}

	// 没有内容的话直接 append 就行了
	lastMsg.MultiContent = append(lastMsg.MultiContent, openai.ChatMessagePart{
		Type:     openai.ChatMessagePartTypeImageURL,
		ImageURL: &openai.ChatMessageImageURL{URL: imageURL},
	})
	messages[len(messages)-1] = lastMsg
	return nil
}
