package anthropic

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	jsonvalue "github.com/Andrew-M-C/go.jsonvalue"
	"github.com/sashabaranov/go-openai"
)

const (
	defaultMaxTokens = 100000

	// signaturePrefix/Suffix 用于在 reasoning_content 中嵌入 thinking block 的 signature
	signaturePrefix = "\n\n<signature>"
	signatureSuffix = "</signature>"

	// redactedThinkingMark 用于标记 redacted_thinking block
	redactedThinkingMark = "\n\n<redacted_thinking>true</redacted_thinking>"
)

var (
	signatureRe        = regexp.MustCompile(`(?s)\n\n<signature>(.*?)</signature>`)
	redactedThinkingRe = regexp.MustCompile(`\n\n<redacted_thinking>true</redacted_thinking>`)
)

// BuildRequest 将 OpenAI ChatCompletionRequest 转换为 Anthropic Messages Request JSON。
// tools 为打包好的 OpenAI Tool 列表，extraFields 为额外字段（可为 nil）。
// 返回的 []byte 为可直接发送的 JSON body。
func BuildRequest(
	model string,
	messages []openai.ChatCompletionMessage,
	tools []openai.Tool,
	extraFields *jsonvalue.V,
) ([]byte, error) {
	// 构建基础请求 map
	reqMap := map[string]any{
		"model":  model,
		"stream": true,
	}

	// 处理 system 和 messages
	if err := buildMessagesIntoMap(reqMap, messages); err != nil {
		return nil, err
	}

	// 构建 tools 和 tool_choice
	if len(tools) > 0 {
		reqMap["tools"] = buildTools(tools)
		reqMap["tool_choice"] = map[string]any{"type": "auto"}
	}

	// 注入 extraFields（含字段名映射）
	toolChoiceOverride := false
	if extraFields != nil {
		extraFields.RangeObjects(func(key string, value *jsonvalue.V) bool {
			switch key {
			case "stop":
				// OpenAI stop → Anthropic stop_sequences
				b, _ := value.Marshal()
				var v any
				_ = json.Unmarshal(b, &v)
				reqMap["stop_sequences"] = v
			case "tool_choice":
				mapped := mapToolChoice(value)
				if mapped != nil {
					reqMap["tool_choice"] = mapped
					toolChoiceOverride = true
				}
			default:
				b, _ := value.Marshal()
				var v any
				_ = json.Unmarshal(b, &v)
				reqMap[key] = v
			}
			return true
		})
	}
	_ = toolChoiceOverride

	// 确保 max_tokens 有默认值
	if _, ok := reqMap["max_tokens"]; !ok {
		reqMap["max_tokens"] = defaultMaxTokens
	} else if mt, ok := reqMap["max_tokens"].(float64); ok && mt == 0 {
		reqMap["max_tokens"] = defaultMaxTokens
	}

	return json.Marshal(reqMap)
}

// buildMessagesIntoMap 将 OpenAI messages 处理后写入 reqMap 的 system 和 messages 字段
func buildMessagesIntoMap(reqMap map[string]any, msgs []openai.ChatCompletionMessage) error {
	// 提取 system 消息（Anthropic 要求 system 作为顶层字段）
	start := 0
	var systemParts []string
	for start < len(msgs) && msgs[start].Role == openai.ChatMessageRoleSystem {
		systemParts = append(systemParts, msgs[start].Content)
		start++
	}
	if len(systemParts) == 1 {
		reqMap["system"] = systemParts[0]
	} else if len(systemParts) > 1 {
		reqMap["system"] = strings.Join(systemParts, "\n\n")
	}

	msgs = msgs[start:]

	// 转换 messages
	result, err := convertMessages(msgs)
	if err != nil {
		return err
	}
	reqMap["messages"] = result
	return nil
}

// convertMessages 将 OpenAI messages 转换为 Anthropic messages 数组
func convertMessages(msgs []openai.ChatCompletionMessage) ([]map[string]any, error) {
	var result []map[string]any

	type pendingToolResult struct {
		toolUseID string
		content   string
	}
	var pendingTools []pendingToolResult

	flushPendingTools := func() {
		if len(pendingTools) == 0 {
			return
		}
		blocks := make([]map[string]any, 0, len(pendingTools))
		for _, t := range pendingTools {
			blocks = append(blocks, map[string]any{
				"type":        "tool_result",
				"tool_use_id": t.toolUseID,
				"content":     t.content,
			})
		}
		result = append(result, map[string]any{
			"role":    "user",
			"content": blocks,
		})
		pendingTools = nil
	}

	for _, msg := range msgs {
		switch msg.Role {
		case openai.ChatMessageRoleTool:
			pendingTools = append(pendingTools, pendingToolResult{
				toolUseID: msg.ToolCallID,
				content:   msg.Content,
			})

		case openai.ChatMessageRoleAssistant:
			flushPendingTools()
			blocks, err := buildAssistantContent(msg)
			if err != nil {
				return nil, err
			}
			result = append(result, map[string]any{
				"role":    "assistant",
				"content": blocks,
			})

		case openai.ChatMessageRoleUser:
			flushPendingTools()
			content, err := buildUserContent(msg)
			if err != nil {
				return nil, err
			}
			// 合并连续 user 消息
			if len(result) > 0 && result[len(result)-1]["role"] == "user" {
				prev := result[len(result)-1]
				prev["content"] = mergeUserContent(prev["content"], content)
			} else {
				result = append(result, map[string]any{
					"role":    "user",
					"content": content,
				})
			}
		}
	}

	flushPendingTools()
	return result, nil
}

// buildAssistantContent 将 assistant 消息转换为 Anthropic content block 数组
// 顺序：thinking block → text block → tool_use block
func buildAssistantContent(msg openai.ChatCompletionMessage) ([]map[string]any, error) {
	var blocks []map[string]any

	// reasoning_content → thinking block（含 signature 解析）
	if msg.ReasoningContent != "" {
		thinkingBlocks := parseReasoningContent(msg.ReasoningContent)
		blocks = append(blocks, thinkingBlocks...)
	}

	// text content
	if msg.Content != "" {
		blocks = append(blocks, map[string]any{
			"type": "text",
			"text": msg.Content,
		})
	}

	// tool_calls → tool_use blocks
	for _, tc := range msg.ToolCalls {
		var input any
		if tc.Function.Arguments != "" {
			if err := json.Unmarshal([]byte(tc.Function.Arguments), &input); err != nil {
				input = tc.Function.Arguments
			}
		}
		blocks = append(blocks, map[string]any{
			"type":  "tool_use",
			"id":    tc.ID,
			"name":  tc.Function.Name,
			"input": input,
		})
	}

	return blocks, nil
}

// parseReasoningContent 从 reasoning_content 字符串中解析出 thinking/redacted_thinking block
func parseReasoningContent(s string) []map[string]any {
	sigMatch := signatureRe.FindStringSubmatchIndex(s)
	if sigMatch == nil {
		return []map[string]any{{
			"type":     "thinking",
			"thinking": s,
		}}
	}

	signature := s[sigMatch[2]:sigMatch[3]]
	thinkingText := s[:sigMatch[0]]

	isRedacted := redactedThinkingRe.MatchString(s)
	if isRedacted {
		return []map[string]any{{
			"type":      "redacted_thinking",
			"signature": signature,
		}}
	}

	return []map[string]any{{
		"type":      "thinking",
		"thinking":  thinkingText,
		"signature": signature,
	}}
}

// buildUserContent 将 user 消息转换为 Anthropic content 格式（string 或 []block）
func buildUserContent(msg openai.ChatCompletionMessage) (any, error) {
	if msg.Content != "" && len(msg.MultiContent) == 0 {
		return msg.Content, nil
	}
	if len(msg.MultiContent) > 0 {
		var blocks []map[string]any
		for _, part := range msg.MultiContent {
			switch part.Type {
			case openai.ChatMessagePartTypeText:
				blocks = append(blocks, map[string]any{
					"type": "text",
					"text": part.Text,
				})
			case openai.ChatMessagePartTypeImageURL:
				block, err := buildImageBlock(part)
				if err != nil {
					return nil, err
				}
				blocks = append(blocks, block)
			}
		}
		return blocks, nil
	}
	return "", nil
}

// buildImageBlock 将 OpenAI image_url content part 转换为 Anthropic image content block
func buildImageBlock(part openai.ChatMessagePart) (map[string]any, error) {
	if part.ImageURL == nil {
		return nil, fmt.Errorf("image_url part 缺少 ImageURL 字段")
	}
	rawURL := part.ImageURL.URL

	// data URI: "data:image/png;base64,xxxx"
	if strings.HasPrefix(rawURL, "data:") {
		commaIdx := strings.Index(rawURL, ",")
		if commaIdx < 0 {
			return nil, fmt.Errorf("无效的 data URI: %s", rawURL)
		}
		header := rawURL[5:commaIdx] // "image/png;base64"
		data := rawURL[commaIdx+1:]
		mediaType := strings.SplitN(header, ";", 2)[0]

		return map[string]any{
			"type": "image",
			"source": map[string]any{
				"type":       "base64",
				"media_type": mediaType,
				"data":       data,
			},
		}, nil
	}

	// 普通 URL
	return map[string]any{
		"type": "image",
		"source": map[string]any{
			"type": "url",
			"url":  rawURL,
		},
	}, nil
}

// mergeUserContent 将两个 user content 合并（用 \n\n 分隔）
func mergeUserContent(prev, curr any) any {
	prevStr, prevIsStr := prev.(string)
	currStr, currIsStr := curr.(string)
	if prevIsStr && currIsStr {
		return prevStr + "\n\n" + currStr
	}
	prevBlocks := toContentBlocks(prev)
	currBlocks := toContentBlocks(curr)
	return append(prevBlocks, currBlocks...)
}

func toContentBlocks(v any) []map[string]any {
	switch val := v.(type) {
	case string:
		if val == "" {
			return nil
		}
		return []map[string]any{{"type": "text", "text": val}}
	case []map[string]any:
		return val
	default:
		return nil
	}
}

// buildTools 将 OpenAI Tool 列表转换为 Anthropic tool 格式
func buildTools(tools []openai.Tool) []map[string]any {
	result := make([]map[string]any, 0, len(tools))
	for _, t := range tools {
		if t.Function == nil {
			continue
		}
		entry := map[string]any{
			"name":         t.Function.Name,
			"input_schema": t.Function.Parameters,
		}
		if t.Function.Description != "" {
			entry["description"] = t.Function.Description
		}
		result = append(result, entry)
	}
	return result
}

// mapToolChoice 将 jsonvalue 格式的 OpenAI tool_choice 映射为 Anthropic 格式
func mapToolChoice(v *jsonvalue.V) any {
	if v == nil {
		return nil
	}
	if v.IsString() {
		switch v.String() {
		case "auto":
			return map[string]any{"type": "auto"}
		case "required":
			return map[string]any{"type": "any"}
		case "none":
			return nil
		}
		return nil
	}
	if v.IsObject() {
		typ, _ := v.GetString("type")
		if typ == "function" {
			name, _ := v.GetString("function", "name")
			if name != "" {
				return map[string]any{"type": "tool", "name": name}
			}
		}
	}
	return nil
}
