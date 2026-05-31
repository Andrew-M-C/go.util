package anthropic

import (
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
)

var (
	signatureRe        = regexp.MustCompile(`(?s)\n\n<signature>(.*?)</signature>`)
	redactedThinkingRe *regexp.Regexp
	// redactedThinkingMark 嵌入 reasoning_content，用于标记 redacted_thinking block
	redactedThinkingMark string
)

func init() {
	tag := redactedThinkingTag()
	redactedThinkingMark = "\n\n<" + tag + ">true</" + tag + ">"
	redactedThinkingRe = regexp.MustCompile(`(?s)\n\n<` + tag + `>true</` + tag + `>`)
}

func redactedThinkingTag() string {
	return "redacted" + "_" + "thinking"
}

// RedactedThinkingMarker 返回嵌入 reasoning_content 的 redacted_thinking 标记。
func RedactedThinkingMarker() string {
	return redactedThinkingMark
}

// jsonClone 深拷贝 JSON 节点。Append 不会复制子节点，复用对象池时后续 NewObject 可能覆盖已挂到数组上的节点。
func jsonClone(v *jsonvalue.V) *jsonvalue.V {
	if v == nil {
		return nil
	}
	return jsonvalue.MustUnmarshal(v.MustMarshal())
}

func appendV(arr *jsonvalue.V, v *jsonvalue.V) {
	arr.MustAppend(jsonClone(v)).InTheEnd()
}

// BuildRequest 将 OpenAI ChatCompletionRequest 转换为 Anthropic Messages Request JSON。
// tools 为打包好的 OpenAI Tool 列表，extraFields 为额外字段（可为 nil）。
// 返回的 []byte 为可直接发送的 JSON body。
func BuildRequest(
	model string,
	messages []openai.ChatCompletionMessage,
	tools []openai.Tool,
	extraFields *jsonvalue.V,
) ([]byte, error) {
	req := jsonvalue.NewObject()
	req.MustSetString(model).At("model")
	req.MustSetBool(true).At("stream")

	if err := buildMessagesInto(req, messages); err != nil {
		return nil, err
	}

	if len(tools) > 0 {
		req.MustSet(buildTools(tools)).At("tools")
		req.MustSet(toolChoiceObject("auto")).At("tool_choice")
	}

	if extraFields != nil {
		extraFields.RangeObjects(func(key string, value *jsonvalue.V) bool {
			switch key {
			case "stop":
				req.At("stop_sequences").Set(value)
			case "tool_choice":
				if mapped := mapToolChoice(value); mapped != nil {
					req.At("tool_choice").Set(mapped)
				}
			default:
				req.At(key).Set(value)
			}
			return true
		})
	}

	ensureMaxTokens(req)

	return req.Marshal(jsonvalue.OptUTF8())
}

func ensureMaxTokens(req *jsonvalue.V) {
	mt := req.MustGet("max_tokens")
	if mt.ValueType() == jsonvalue.NotExist {
		req.MustSetInt(defaultMaxTokens).At("max_tokens")
		return
	}
	if mt.IsNumber() && mt.Int() == 0 {
		req.MustSetInt(defaultMaxTokens).At("max_tokens")
	}
}

func toolChoiceObject(typ string) *jsonvalue.V {
	o := jsonvalue.NewObject()
	o.MustSetString(typ).At("type")
	return o
}

// buildMessagesInto 将 OpenAI messages 处理后写入 req 的 system 和 messages 字段
func buildMessagesInto(req *jsonvalue.V, msgs []openai.ChatCompletionMessage) error {
	start := 0
	var systemParts []string
	for start < len(msgs) && msgs[start].Role == openai.ChatMessageRoleSystem {
		systemParts = append(systemParts, msgs[start].Content)
		start++
	}
	if len(systemParts) == 1 {
		req.MustSetString(systemParts[0]).At("system")
	} else if len(systemParts) > 1 {
		req.MustSetString(strings.Join(systemParts, "\n\n")).At("system")
	}

	msgs = msgs[start:]
	if len(msgs) == 0 {
		req.MustSet(jsonvalue.NewArray()).At("messages")
		return nil
	}

	messages, err := convertMessages(msgs)
	if err != nil {
		return err
	}
	req.MustSet(messages).At("messages")
	return nil
}

// convertMessages 将 OpenAI messages 转换为 Anthropic messages 数组
func convertMessages(msgs []openai.ChatCompletionMessage) (*jsonvalue.V, error) {
	arr := jsonvalue.NewArray()

	type pendingToolResult struct {
		toolUseID string
		content   string
	}
	var pendingTools []pendingToolResult

	var lastMsg *jsonvalue.V

	flushPendingTools := func() {
		if len(pendingTools) == 0 {
			return
		}
		blocks := jsonvalue.NewArray()
		for _, t := range pendingTools {
			block := jsonvalue.NewObject()
			block.MustSetString("tool_result").At("type")
			block.MustSetString(t.toolUseID).At("tool_use_id")
			block.MustSetString(t.content).At("content")
			appendV(blocks, block)
		}
		m := jsonvalue.NewObject()
		m.MustSetString("user").At("role")
		m.MustSet(blocks).At("content")
		appendV(arr, m)
		lastMsg = arr.MustGet(arr.Len() - 1)
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
			blocks := buildAssistantContent(msg)
			m := jsonvalue.NewObject()
			m.MustSetString("assistant").At("role")
			m.MustSet(blocks).At("content")
			appendV(arr, m)
			lastMsg = arr.MustGet(arr.Len() - 1)

		case openai.ChatMessageRoleUser:
			flushPendingTools()
			content, err := buildUserContent(msg)
			if err != nil {
				return nil, err
			}
			if lastMsg != nil {
				if lastMsg.MustGet("role").String() == "user" {
					prev := lastMsg.MustGet("content")
					lastMsg.MustSet(mergeUserContent(prev, content)).At("content")
					continue
				}
			}
			m := jsonvalue.NewObject()
			m.MustSetString("user").At("role")
			m.MustSet(content).At("content")
			appendV(arr, m)
			lastMsg = arr.MustGet(arr.Len() - 1)
		}
	}

	flushPendingTools()
	return arr, nil
}

// buildAssistantContent 将 assistant 消息转换为 Anthropic content block 数组
// 顺序：thinking block → text block → tool_use block
func buildAssistantContent(msg openai.ChatCompletionMessage) *jsonvalue.V {
	blocks := jsonvalue.NewArray()

	if msg.ReasoningContent != "" {
		for _, b := range parseReasoningContent(msg.ReasoningContent) {
			appendV(blocks, b)
		}
	}

	if msg.Content != "" {
		textBlock := jsonvalue.NewObject()
		textBlock.MustSetString("text").At("type")
		textBlock.MustSetString(msg.Content).At("text")
		appendV(blocks, textBlock)
	}

	for _, tc := range msg.ToolCalls {
		block := jsonvalue.NewObject()
		block.MustSetString("tool_use").At("type")
		block.MustSetString(tc.ID).At("id")
		block.MustSetString(tc.Function.Name).At("name")
		if tc.Function.Arguments != "" {
			input, err := jsonvalue.UnmarshalString(tc.Function.Arguments)
			if err != nil {
				block.MustSetString(tc.Function.Arguments).At("input")
			} else {
				block.MustSet(input).At("input")
			}
		}
		appendV(blocks, block)
	}

	return blocks
}

// parseReasoningContent 从 reasoning_content 字符串中解析出 thinking/redacted_thinking block
func parseReasoningContent(s string) []*jsonvalue.V {
	sigMatch := signatureRe.FindStringSubmatchIndex(s)
	if sigMatch == nil {
		b := jsonvalue.NewObject()
		b.MustSetString("thinking").At("type")
		b.MustSetString(s).At("thinking")
		return []*jsonvalue.V{b}
	}

	signature := s[sigMatch[2]:sigMatch[3]]
	thinkingText := s[:sigMatch[0]]

	if redactedThinkingRe.MatchString(s) {
		b := jsonvalue.NewObject()
		b.MustSetString("redacted" + "_" + "thinking").At("type")
		b.MustSetString(signature).At("signature")
		return []*jsonvalue.V{b}
	}

	b := jsonvalue.NewObject()
	b.MustSetString("thinking").At("type")
	b.MustSetString(thinkingText).At("thinking")
	b.MustSetString(signature).At("signature")
	return []*jsonvalue.V{b}
}

// buildUserContent 将 user 消息转换为 Anthropic content（string 或 block 数组）
func buildUserContent(msg openai.ChatCompletionMessage) (*jsonvalue.V, error) {
	if msg.Content != "" && len(msg.MultiContent) == 0 {
		return jsonvalue.NewString(msg.Content), nil
	}
	if len(msg.MultiContent) > 0 {
		blocks := jsonvalue.NewArray()
		for _, part := range msg.MultiContent {
			switch part.Type {
			case openai.ChatMessagePartTypeText:
				textBlock := jsonvalue.NewObject()
				textBlock.MustSetString("text").At("type")
				textBlock.MustSetString(part.Text).At("text")
				appendV(blocks, textBlock)
			case openai.ChatMessagePartTypeImageURL:
				block, err := buildImageBlock(part)
				if err != nil {
					return nil, err
				}
				appendV(blocks, block)
			}
		}
		return blocks, nil
	}
	return jsonvalue.NewString(""), nil
}

// buildImageBlock 将 OpenAI image_url content part 转换为 Anthropic image content block
func buildImageBlock(part openai.ChatMessagePart) (*jsonvalue.V, error) {
	if part.ImageURL == nil {
		return nil, fmt.Errorf("image_url part 缺少 ImageURL 字段")
	}
	rawURL := part.ImageURL.URL

	block := jsonvalue.NewObject()
	block.MustSetString("image").At("type")
	source := jsonvalue.NewObject()

	if strings.HasPrefix(rawURL, "data:") {
		commaIdx := strings.Index(rawURL, ",")
		if commaIdx < 0 {
			return nil, fmt.Errorf("无效的 data URI: %s", rawURL)
		}
		header := rawURL[5:commaIdx]
		data := rawURL[commaIdx+1:]
		mediaType := strings.SplitN(header, ";", 2)[0]
		source.MustSetString("base64").At("type")
		source.MustSetString(mediaType).At("media_type")
		source.MustSetString(data).At("data")
	} else {
		source.MustSetString("url").At("type")
		source.MustSetString(rawURL).At("url")
	}

	block.MustSet(source).At("source")
	return block, nil
}

// mergeUserContent 将两个 user content 合并（用 \n\n 分隔）
func mergeUserContent(prev, curr *jsonvalue.V) *jsonvalue.V {
	if prev.IsString() && curr.IsString() {
		return jsonvalue.NewString(prev.String() + "\n\n" + curr.String())
	}
	merged := jsonvalue.NewArray()
	appendAsContentBlocks(merged, prev)
	appendAsContentBlocks(merged, curr)
	return merged
}

func appendAsContentBlocks(arr *jsonvalue.V, content *jsonvalue.V) {
	if content.IsString() {
		if s := content.String(); s != "" {
			block := jsonvalue.NewObject()
			block.MustSetString("text").At("type")
			block.MustSetString(s).At("text")
			appendV(arr, block)
		}
		return
	}
	if content.IsArray() {
		content.RangeArray(func(_ int, v *jsonvalue.V) bool {
			appendV(arr, v)
			return true
		})
	}
}

// buildTools 将 OpenAI Tool 列表转换为 Anthropic tool 数组
func buildTools(tools []openai.Tool) *jsonvalue.V {
	arr := jsonvalue.NewArray()
	for _, t := range tools {
		if t.Function == nil {
			continue
		}
		entry := jsonvalue.NewObject()
		entry.MustSetString(t.Function.Name).At("name")
		if t.Function.Description != "" {
			entry.MustSetString(t.Function.Description).At("description")
		}
		schema, err := jsonvalue.Import(t.Function.Parameters)
		if err != nil {
			schema = jsonvalue.NewObject()
		}
		entry.MustSet(schema).At("input_schema")
		appendV(arr, entry)
	}
	return arr
}

// mapToolChoice 将 jsonvalue 格式的 OpenAI tool_choice 映射为 Anthropic 格式
func mapToolChoice(v *jsonvalue.V) *jsonvalue.V {
	if v == nil {
		return nil
	}
	if v.IsString() {
		switch v.String() {
		case "auto":
			return toolChoiceObject("auto")
		case "required":
			return toolChoiceObject("any")
		case "none":
			return nil
		}
		return nil
	}
	if v.IsObject() {
		typ, err := v.GetString("type")
		if err == nil && typ == "function" {
			name, err := v.GetString("function", "name")
			if err == nil && name != "" {
				o := jsonvalue.NewObject()
				o.MustSetString("tool").At("type")
				o.MustSetString(name).At("name")
				return o
			}
		}
	}
	return nil
}
