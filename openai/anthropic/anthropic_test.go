package anthropic_test

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"testing"

	jsonvalue "github.com/Andrew-M-C/go.jsonvalue"
	"github.com/Andrew-M-C/go.util/openai/anthropic"
	"github.com/sashabaranov/go-openai"
	"github.com/smartystreets/goconvey/convey"
)

var (
	cv = convey.Convey
	so = convey.So
	eq = convey.ShouldEqual

	isNil  = convey.ShouldBeNil
	isTrue = convey.ShouldBeTrue
)

// ================== 请求转换辅助函数 ==================

// buildReqJSON 调用 BuildRequest 并将结果解析为 map[string]any
func buildReqJSON(
	model string,
	msgs []openai.ChatCompletionMessage,
	tools []openai.Tool,
	extra *jsonvalue.V,
) (map[string]any, error) {
	b, err := anthropic.BuildRequest(model, msgs, tools, extra)
	if err != nil {
		return nil, err
	}
	var m map[string]any
	if err2 := json.Unmarshal(b, &m); err2 != nil {
		return nil, err2
	}
	return m, nil
}

// extraFrom 从 map[string]any 构建 *jsonvalue.V（用于构造 extraFields）
func extraFrom(m map[string]any) *jsonvalue.V {
	v, err := jsonvalue.Import(m)
	if err != nil {
		panic(fmt.Sprintf("extraFrom: %v", err))
	}
	return v
}

// nav 从 any 按 key 路径导航到嵌套字段
func nav(v any, keys ...string) any {
	for _, k := range keys {
		m, ok := v.(map[string]any)
		if !ok {
			return nil
		}
		v = m[k]
	}
	return v
}

func navStr(v any, keys ...string) string  { s, _ := nav(v, keys...).(string); return s }
func navBool(v any, keys ...string) bool   { b, _ := nav(v, keys...).(bool); return b }
func navFloat(v any, keys ...string) float64 {
	f, _ := nav(v, keys...).(float64)
	return f
}
func navSlice(v any, keys ...string) []any {
	s, _ := nav(v, keys...).([]any)
	return s
}
func navMap(v any, keys ...string) map[string]any {
	m, _ := nav(v, keys...).(map[string]any)
	return m
}

// ================== SSE 辅助函数 ==================

// sseEvent 构建单个 Anthropic SSE 事件字符串
func sseEvent(eventType string, data map[string]any) string {
	b, _ := json.Marshal(data)
	return fmt.Sprintf("event: %s\ndata: %s\n\n", eventType, b)
}

// wrapSSE 将多个 SSE 事件字符串拼接后包装为 io.ReadCloser
func wrapSSE(events ...string) io.ReadCloser {
	return io.NopCloser(strings.NewReader(strings.Join(events, "")))
}

// sseResult 保存从 NewSSEReader 读取到的所有 OpenAI SSE 数据
type sseResult struct {
	Chunks   []openai.ChatCompletionStreamResponse
	Done     bool     // 是否收到 data: [DONE]
	RawLines []string // 所有 data: 行的原始 payload（不含 "data: " 前缀）
}

// readSSE 从 io.ReadCloser（NewSSEReader 的输出）读取所有 OpenAI SSE chunks
func readSSE(r io.ReadCloser) sseResult {
	defer r.Close()
	scanner := bufio.NewScanner(r)
	var res sseResult
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		payload := strings.TrimPrefix(line, "data: ")
		if payload == "[DONE]" {
			res.Done = true
			break
		}
		res.RawLines = append(res.RawLines, payload)
		var chunk openai.ChatCompletionStreamResponse
		if json.Unmarshal([]byte(payload), &chunk) == nil {
			res.Chunks = append(res.Chunks, chunk)
		}
	}
	return res
}

// translateSSE 将 Anthropic SSE 事件串翻译为 OpenAI SSE chunks
func translateSSE(events ...string) sseResult {
	return readSSE(anthropic.NewSSEReader(wrapSSE(events...)))
}

// ================== BuildRequest 测试 ==================

func TestBuildRequest(t *testing.T) {
	cv("BuildRequest: OpenAI → Anthropic 请求格式转换", t, func() {

		// ---- §10 ----------------------------------------------------------------
		cv("§10: stream 始终为 true", func() {
			req, err := buildReqJSON("claude", []openai.ChatCompletionMessage{
				{Role: openai.ChatMessageRoleUser, Content: "hi"},
			}, nil, nil)
			so(err, isNil)
			so(req["stream"], eq, true)
		})

		// ---- §5 ----------------------------------------------------------------
		cv("§5: system 消息提取为顶层 system 字段", func() {
			cv("单条 system 消息", func() {
				msgs := []openai.ChatCompletionMessage{
					{Role: openai.ChatMessageRoleSystem, Content: "You are helpful"},
					{Role: openai.ChatMessageRoleUser, Content: "hi"},
				}
				req, err := buildReqJSON("claude", msgs, nil, nil)
				so(err, isNil)
				so(req["system"], eq, "You are helpful")
				// system 消息不出现在 messages 列表中
				messages := navSlice(req, "messages")
				so(len(messages), eq, 1)
				so(navStr(messages[0], "role"), eq, "user")
			})

			cv("多条 system 消息用 \\n\\n 合并", func() {
				msgs := []openai.ChatCompletionMessage{
					{Role: openai.ChatMessageRoleSystem, Content: "Part 1"},
					{Role: openai.ChatMessageRoleSystem, Content: "Part 2"},
					{Role: openai.ChatMessageRoleUser, Content: "hi"},
				}
				req, err := buildReqJSON("claude", msgs, nil, nil)
				so(err, isNil)
				so(req["system"], eq, "Part 1\n\nPart 2")
			})

			cv("无 system 消息时不设置 system 字段", func() {
				msgs := []openai.ChatCompletionMessage{
					{Role: openai.ChatMessageRoleUser, Content: "hi"},
				}
				req, err := buildReqJSON("claude", msgs, nil, nil)
				so(err, isNil)
				_, ok := req["system"]
				so(ok, eq, false)
			})
		})

		// ---- §2 ----------------------------------------------------------------
		cv("§2: extraFields 处理", func() {
			cv("未知字段原样透传到请求 JSON", func() {
				req, err := buildReqJSON("claude", []openai.ChatCompletionMessage{
					{Role: openai.ChatMessageRoleUser, Content: "hi"},
				}, nil, extraFrom(map[string]any{
					"temperature": 0.7,
					"top_p":       0.9,
				}))
				so(err, isNil)
				so(navFloat(req, "temperature"), eq, 0.7)
				so(navFloat(req, "top_p"), eq, 0.9)
			})

			cv("stop 映射为 stop_sequences", func() {
				req, err := buildReqJSON("claude", []openai.ChatCompletionMessage{
					{Role: openai.ChatMessageRoleUser, Content: "hi"},
				}, nil, extraFrom(map[string]any{
					"stop": []string{"STOP", "END"},
				}))
				so(err, isNil)
				_, hasStop := req["stop"]
				so(hasStop, eq, false) // 原始 stop 键不应存在
				seqs := navSlice(req, "stop_sequences")
				so(len(seqs), eq, 2)
				so(seqs[0], eq, "STOP")
				so(seqs[1], eq, "END")
			})

			cv("max_tokens 缺失时默认设为 100000", func() {
				req, err := buildReqJSON("claude", []openai.ChatCompletionMessage{
					{Role: openai.ChatMessageRoleUser, Content: "hi"},
				}, nil, nil)
				so(err, isNil)
				so(navFloat(req, "max_tokens"), eq, float64(100000))
			})

			cv("显式设置 max_tokens 时保留用户值", func() {
				req, err := buildReqJSON("claude", []openai.ChatCompletionMessage{
					{Role: openai.ChatMessageRoleUser, Content: "hi"},
				}, nil, extraFrom(map[string]any{"max_tokens": 512}))
				so(err, isNil)
				so(navFloat(req, "max_tokens"), eq, float64(512))
			})
		})

		// ---- §18 ---------------------------------------------------------------
		cv("§18: content 字段格式", func() {
			cv("纯文本 user 消息 → string", func() {
				msgs := []openai.ChatCompletionMessage{
					{Role: openai.ChatMessageRoleUser, Content: "hello world"},
				}
				req, err := buildReqJSON("claude", msgs, nil, nil)
				so(err, isNil)
				messages := navSlice(req, "messages")
				so(navStr(messages[0], "content"), eq, "hello world")
			})

			cv("MultiContent 纯文本 → text block 数组", func() {
				msgs := []openai.ChatCompletionMessage{
					{
						Role: openai.ChatMessageRoleUser,
						MultiContent: []openai.ChatMessagePart{
							{Type: openai.ChatMessagePartTypeText, Text: "desc"},
						},
					},
				}
				req, err := buildReqJSON("claude", msgs, nil, nil)
				so(err, isNil)
				messages := navSlice(req, "messages")
				content := navSlice(messages[0], "content")
				so(len(content), eq, 1)
				so(navStr(content[0], "type"), eq, "text")
				so(navStr(content[0], "text"), eq, "desc")
			})
		})

		// ---- §7 ----------------------------------------------------------------
		cv("§7: image_url → Anthropic image content block", func() {
			cv("data URI (base64)", func() {
				msgs := []openai.ChatCompletionMessage{
					{
						Role: openai.ChatMessageRoleUser,
						MultiContent: []openai.ChatMessagePart{
							{
								Type: openai.ChatMessagePartTypeImageURL,
								ImageURL: &openai.ChatMessageImageURL{
									URL: "data:image/png;base64,iVBORw0KGgo=",
								},
							},
						},
					},
				}
				req, err := buildReqJSON("claude", msgs, nil, nil)
				so(err, isNil)
				messages := navSlice(req, "messages")
				content := navSlice(messages[0], "content")
				so(len(content), eq, 1)
				so(navStr(content[0], "type"), eq, "image")
				so(navStr(content[0], "source", "type"), eq, "base64")
				so(navStr(content[0], "source", "media_type"), eq, "image/png")
				so(navStr(content[0], "source", "data"), eq, "iVBORw0KGgo=")
			})

			cv("普通 URL", func() {
				msgs := []openai.ChatCompletionMessage{
					{
						Role: openai.ChatMessageRoleUser,
						MultiContent: []openai.ChatMessagePart{
							{
								Type: openai.ChatMessagePartTypeImageURL,
								ImageURL: &openai.ChatMessageImageURL{
									URL: "https://example.com/img.jpg",
								},
							},
						},
					},
				}
				req, err := buildReqJSON("claude", msgs, nil, nil)
				so(err, isNil)
				messages := navSlice(req, "messages")
				content := navSlice(messages[0], "content")
				so(navStr(content[0], "type"), eq, "image")
				so(navStr(content[0], "source", "type"), eq, "url")
				so(navStr(content[0], "source", "url"), eq, "https://example.com/img.jpg")
			})
		})

		// ---- §16 ---------------------------------------------------------------
		cv("§16: 连续 user 消息用 \\n\\n 合并", func() {
			msgs := []openai.ChatCompletionMessage{
				{Role: openai.ChatMessageRoleUser, Content: "first"},
				{Role: openai.ChatMessageRoleUser, Content: "second"},
			}
			req, err := buildReqJSON("claude", msgs, nil, nil)
			so(err, isNil)
			messages := navSlice(req, "messages")
			so(len(messages), eq, 1) // 合并为一条
			so(navStr(messages[0], "role"), eq, "user")
			so(navStr(messages[0], "content"), eq, "first\n\nsecond")
		})

		// ---- §14 / §6 tool 定义转换 --------------------------------------------
		cv("§14: OpenAI Tool 定义 → Anthropic tool 格式", func() {
			tools := []openai.Tool{
				{
					Type: openai.ToolTypeFunction,
					Function: &openai.FunctionDefinition{
						Name:        "get_weather",
						Description: "获取天气",
						Parameters: map[string]any{
							"type": "object",
							"properties": map[string]any{
								"city": map[string]any{"type": "string"},
							},
							"required": []string{"city"},
						},
					},
				},
			}
			req, err := buildReqJSON("claude", []openai.ChatCompletionMessage{
				{Role: openai.ChatMessageRoleUser, Content: "天气如何"},
			}, tools, nil)
			so(err, isNil)

			toolList := navSlice(req, "tools")
			so(len(toolList), eq, 1)
			so(navStr(toolList[0], "name"), eq, "get_weather")
			so(navStr(toolList[0], "description"), eq, "获取天气")
			// input_schema 应存在
			schema := navMap(toolList[0], "input_schema")
			so(schema, convey.ShouldNotBeNil)
			so(navStr(schema, "type"), eq, "object")

			// 有 tools 时默认 tool_choice 为 auto
			tc := navMap(req, "tool_choice")
			so(navStr(tc, "type"), eq, "auto")
		})

		// ---- §6 tool_calls → tool_use block ------------------------------------
		cv("§6: assistant tool_calls → Anthropic tool_use content block", func() {
			msgs := []openai.ChatCompletionMessage{
				{Role: openai.ChatMessageRoleUser, Content: "天气如何"},
				{
					Role: openai.ChatMessageRoleAssistant,
					ToolCalls: []openai.ToolCall{{
						ID:   "call_abc",
						Type: openai.ToolTypeFunction,
						Function: openai.FunctionCall{
							Name:      "get_weather",
							Arguments: `{"city":"Beijing"}`,
						},
					}},
				},
			}
			req, err := buildReqJSON("claude", msgs, nil, nil)
			so(err, isNil)
			messages := navSlice(req, "messages")
			so(len(messages), eq, 2)

			assistantMsg := navMap(messages[1], "")
			so(navStr(messages[1], "role"), eq, "assistant")
			content := navSlice(messages[1], "content")
			_ = assistantMsg
			so(len(content), eq, 1)
			so(navStr(content[0], "type"), eq, "tool_use")
			so(navStr(content[0], "id"), eq, "call_abc")
			so(navStr(content[0], "name"), eq, "get_weather")
			// input 应是被解析后的对象
			input := navMap(content[0], "input")
			so(navStr(input, "city"), eq, "Beijing")
		})

		// ---- §6 / §15 role=tool → tool_result ----------------------------------
		cv("§6/§15: role=tool 消息 → tool_result content block", func() {
			msgs := []openai.ChatCompletionMessage{
				{Role: openai.ChatMessageRoleUser, Content: "天气如何"},
				{
					Role: openai.ChatMessageRoleAssistant,
					ToolCalls: []openai.ToolCall{{
						ID:   "call_abc",
						Type: openai.ToolTypeFunction,
						Function: openai.FunctionCall{
							Name:      "get_weather",
							Arguments: `{"city":"Beijing"}`,
						},
					}},
				},
				{
					Role:       openai.ChatMessageRoleTool,
					ToolCallID: "call_abc",
					Content:    "Beijing: 25°C",
				},
			}
			req, err := buildReqJSON("claude", msgs, nil, nil)
			so(err, isNil)
			messages := navSlice(req, "messages")
			// 期望：user, assistant, user(tool_result)
			so(len(messages), eq, 3)
			so(navStr(messages[2], "role"), eq, "user")
			content := navSlice(messages[2], "content")
			so(len(content), eq, 1)
			so(navStr(content[0], "type"), eq, "tool_result")
			so(navStr(content[0], "tool_use_id"), eq, "call_abc")
			so(navStr(content[0], "content"), eq, "Beijing: 25°C")
			// §19: is_error 始终不设置
			_, hasIsError := nav(content[0], "is_error").(bool)
			so(hasIsError, eq, false)
		})

		cv("§6/§15: 多个 tool_result 合并到同一个 user 消息", func() {
			msgs := []openai.ChatCompletionMessage{
				{Role: openai.ChatMessageRoleUser, Content: "查询"},
				{
					Role: openai.ChatMessageRoleAssistant,
					ToolCalls: []openai.ToolCall{
						{ID: "c1", Type: openai.ToolTypeFunction, Function: openai.FunctionCall{Name: "t1", Arguments: `{}`}},
						{ID: "c2", Type: openai.ToolTypeFunction, Function: openai.FunctionCall{Name: "t2", Arguments: `{}`}},
					},
				},
				{Role: openai.ChatMessageRoleTool, ToolCallID: "c1", Content: "result1"},
				{Role: openai.ChatMessageRoleTool, ToolCallID: "c2", Content: "result2"},
			}
			req, err := buildReqJSON("claude", msgs, nil, nil)
			so(err, isNil)
			messages := navSlice(req, "messages")
			// user, assistant, user(tool_results)
			so(len(messages), eq, 3)
			content := navSlice(messages[2], "content")
			so(len(content), eq, 2)
			so(navStr(content[0], "tool_use_id"), eq, "c1")
			so(navStr(content[1], "tool_use_id"), eq, "c2")
		})

		// ---- §11 / §21 / §22 reasoning_content --------------------------------
		cv("§11/§21: reasoning_content → thinking block", func() {
			cv("无 signature 的普通 thinking", func() {
				msgs := []openai.ChatCompletionMessage{
					{Role: openai.ChatMessageRoleUser, Content: "think"},
					{
						Role:             openai.ChatMessageRoleAssistant,
						ReasoningContent: "这是思维过程",
						Content:          "这是答案",
					},
					{Role: openai.ChatMessageRoleUser, Content: "继续"},
				}
				req, err := buildReqJSON("claude", msgs, nil, nil)
				so(err, isNil)
				messages := navSlice(req, "messages")
				content := navSlice(messages[1], "content")
				// §26: 顺序应为 thinking → text
				so(navStr(content[0], "type"), eq, "thinking")
				so(navStr(content[0], "thinking"), eq, "这是思维过程")
				so(navStr(content[1], "type"), eq, "text")
			})

			cv("含 signature 的 thinking（从 reasoning_content 中解析还原）", func() {
				rc := "我的分析过程\n\n<signature>ErUBxxxSIGNATURExxx</signature>"
				msgs := []openai.ChatCompletionMessage{
					{Role: openai.ChatMessageRoleUser, Content: "think"},
					{
						Role:             openai.ChatMessageRoleAssistant,
						ReasoningContent: rc,
						Content:          "答案",
					},
					{Role: openai.ChatMessageRoleUser, Content: "继续"},
				}
				req, err := buildReqJSON("claude", msgs, nil, nil)
				so(err, isNil)
				messages := navSlice(req, "messages")
				content := navSlice(messages[1], "content")
				so(navStr(content[0], "type"), eq, "thinking")
				so(navStr(content[0], "thinking"), eq, "我的分析过程")
				so(navStr(content[0], "signature"), eq, "ErUBxxxSIGNATURExxx")
			})
		})

		cv("§22: redacted_thinking 块的还原", func() {
			rc := "\n\n<signature>ErUBREDACTEDSIG</signature>\n\n<redacted_thinking>true</redacted_thinking>"
			msgs := []openai.ChatCompletionMessage{
				{Role: openai.ChatMessageRoleUser, Content: "think"},
				{
					Role:             openai.ChatMessageRoleAssistant,
					ReasoningContent: rc,
					Content:          "答案",
				},
				{Role: openai.ChatMessageRoleUser, Content: "继续"},
			}
			req, err := buildReqJSON("claude", msgs, nil, nil)
			so(err, isNil)
			messages := navSlice(req, "messages")
			content := navSlice(messages[1], "content")
			so(navStr(content[0], "type"), eq, "redacted_thinking")
			so(navStr(content[0], "signature"), eq, "ErUBREDACTEDSIG")
		})

		// ---- §26 thinking block 顺序 --------------------------------------------
		cv("§26: assistant content 顺序为 thinking → text → tool_use", func() {
			rc := "思维\n\n<signature>SIG123</signature>"
			msgs := []openai.ChatCompletionMessage{
				{Role: openai.ChatMessageRoleUser, Content: "q"},
				{
					Role:             openai.ChatMessageRoleAssistant,
					ReasoningContent: rc,
					Content:          "中间文本",
					ToolCalls: []openai.ToolCall{{
						ID:   "c1",
						Type: openai.ToolTypeFunction,
						Function: openai.FunctionCall{Name: "fn", Arguments: `{}`},
					}},
				},
			}
			req, err := buildReqJSON("claude", msgs, nil, nil)
			so(err, isNil)
			messages := navSlice(req, "messages")
			content := navSlice(messages[1], "content")
			so(len(content), eq, 3)
			so(navStr(content[0], "type"), eq, "thinking")
			so(navStr(content[1], "type"), eq, "text")
			so(navStr(content[2], "type"), eq, "tool_use")
		})

		// ---- §23 tool_choice 格式映射 -------------------------------------------
		cv("§23: tool_choice 格式映射", func() {
			msgs := []openai.ChatCompletionMessage{
				{Role: openai.ChatMessageRoleUser, Content: "hi"},
			}
			tools := []openai.Tool{{
				Type: openai.ToolTypeFunction,
				Function: &openai.FunctionDefinition{
					Name:       "fn",
					Parameters: map[string]any{"type": "object"},
				},
			}}

			cv(`"auto" → {"type":"auto"}`, func() {
				req, err := buildReqJSON("claude", msgs, tools, extraFrom(map[string]any{
					"tool_choice": "auto",
				}))
				so(err, isNil)
				so(navStr(req, "tool_choice", "type"), eq, "auto")
			})

			cv(`"required" → {"type":"any"}`, func() {
				req, err := buildReqJSON("claude", msgs, tools, extraFrom(map[string]any{
					"tool_choice": "required",
				}))
				so(err, isNil)
				so(navStr(req, "tool_choice", "type"), eq, "any")
			})

			cv(`{"type":"function","function":{"name":"fn"}} → {"type":"tool","name":"fn"}`, func() {
				req, err := buildReqJSON("claude", msgs, tools, extraFrom(map[string]any{
					"tool_choice": map[string]any{
						"type":     "function",
						"function": map[string]any{"name": "fn"},
					},
				}))
				so(err, isNil)
				so(navStr(req, "tool_choice", "type"), eq, "tool")
				so(navStr(req, "tool_choice", "name"), eq, "fn")
			})
		})
	})
}

// ================== SSE 转换测试 ==================

// messageStart 构建标准 message_start 事件
func messageStart(id, model string) string {
	return sseEvent("message_start", map[string]any{
		"type": "message_start",
		"message": map[string]any{
			"id":    id,
			"model": model,
			"role":  "assistant",
		},
	})
}

// messageStop 构建标准 message_stop 事件
func messageStop() string {
	return sseEvent("message_stop", map[string]any{"type": "message_stop"})
}

// cbStart 构建 content_block_start 事件
func cbStart(index int, blockType string, extra map[string]any) string {
	cb := map[string]any{"type": blockType}
	for k, v := range extra {
		cb[k] = v
	}
	return sseEvent("content_block_start", map[string]any{
		"type":          "content_block_start",
		"index":         index,
		"content_block": cb,
	})
}

// cbDelta 构建 content_block_delta 事件
func cbDelta(index int, deltaType string, extra map[string]any) string {
	delta := map[string]any{"type": deltaType}
	for k, v := range extra {
		delta[k] = v
	}
	return sseEvent("content_block_delta", map[string]any{
		"type":  "content_block_delta",
		"index": index,
		"delta": delta,
	})
}

// cbStop 构建 content_block_stop 事件
func cbStop(index int) string {
	return sseEvent("content_block_stop", map[string]any{
		"type":  "content_block_stop",
		"index": index,
	})
}

// msgDelta 构建 message_delta 事件
func msgDelta(stopReason string, usage map[string]any) string {
	m := map[string]any{
		"type":  "message_delta",
		"delta": map[string]any{"stop_reason": stopReason},
	}
	if usage != nil {
		m["usage"] = usage
	}
	return sseEvent("message_delta", m)
}

func TestSSEReader(t *testing.T) {
	cv("NewSSEReader: Anthropic SSE → OpenAI SSE 格式转换", t, func() {

		// ---- message_start ----------------------------------------------------
		cv("message_start → 首个 chunk 携带 role=assistant / id / model", func() {
			result := translateSSE(
				messageStart("msg_001", "claude-3-5-sonnet"),
				messageStop(),
			)
			so(result.Done, isTrue)
			so(len(result.Chunks) >= 1, isTrue)
			first := result.Chunks[0]
			so(first.ID, eq, "msg_001")
			so(first.Model, eq, "claude-3-5-sonnet")
			so(len(first.Choices) >= 1, isTrue)
			so(first.Choices[0].Delta.Role, eq, openai.ChatMessageRoleAssistant)
		})

		// ---- text_delta ------------------------------------------------------
		cv("text_delta → OpenAI content delta", func() {
			result := translateSSE(
				messageStart("msg_t", "claude"),
				cbStart(0, "text", nil),
				cbDelta(0, "text_delta", map[string]any{"text": "Hello"}),
				cbDelta(0, "text_delta", map[string]any{"text": " world"}),
				cbStop(0),
				msgDelta("end_turn", nil),
				messageStop(),
			)
			so(result.Done, isTrue)

			var textParts []string
			for _, c := range result.Chunks {
				if len(c.Choices) > 0 && c.Choices[0].Delta.Content != "" {
					textParts = append(textParts, c.Choices[0].Delta.Content)
				}
			}
			so(strings.Join(textParts, ""), eq, "Hello world")
		})

		// ---- §11 thinking_delta ----------------------------------------------
		cv("§11: thinking_delta → OpenAI reasoning_content delta", func() {
			result := translateSSE(
				messageStart("msg_th", "claude"),
				cbStart(0, "thinking", nil),
				cbDelta(0, "thinking_delta", map[string]any{"thinking": "分析中..."}),
				cbDelta(0, "thinking_delta", map[string]any{"thinking": "继续分析"}),
				cbStop(0),
				msgDelta("end_turn", nil),
				messageStop(),
			)
			so(result.Done, isTrue)

			var reasoning []string
			for _, c := range result.Chunks {
				if len(c.Choices) > 0 && c.Choices[0].Delta.ReasoningContent != "" {
					reasoning = append(reasoning, c.Choices[0].Delta.ReasoningContent)
				}
			}
			combined := strings.Join(reasoning, "")
			so(strings.Contains(combined, "分析中..."), isTrue)
			so(strings.Contains(combined, "继续分析"), isTrue)
		})

		// ---- §21 signature_delta (普通 thinking) ----------------------------
		cv("§21: signature_delta → reasoning_content 末尾嵌入 <signature>SIG</signature>", func() {
			result := translateSSE(
				messageStart("msg_sig", "claude"),
				cbStart(0, "thinking", nil),
				cbDelta(0, "thinking_delta", map[string]any{"thinking": "思考"}),
				cbDelta(0, "signature_delta", map[string]any{"signature": "SIGVAL123"}),
				cbStop(0),
				msgDelta("end_turn", nil),
				messageStop(),
			)
			so(result.Done, isTrue)

			var reasoning []string
			for _, c := range result.Chunks {
				if len(c.Choices) > 0 && c.Choices[0].Delta.ReasoningContent != "" {
					reasoning = append(reasoning, c.Choices[0].Delta.ReasoningContent)
				}
			}
			combined := strings.Join(reasoning, "")
			so(strings.Contains(combined, "\n\n<signature>SIGVAL123</signature>"), isTrue)
			// 非 redacted_thinking 不应含 redacted 标记
			so(strings.Contains(combined, "<redacted_thinking>"), eq, false)
		})

		// ---- §22 signature_delta (redacted_thinking) -----------------------
		cv("§22: redacted_thinking signature_delta → reasoning_content 附加 <redacted_thinking> 标记", func() {
			result := translateSSE(
				messageStart("msg_red", "claude"),
				cbStart(0, "redacted_thinking", nil),
				cbDelta(0, "signature_delta", map[string]any{"signature": "REDACTED_SIG"}),
				cbStop(0),
				msgDelta("end_turn", nil),
				messageStop(),
			)
			so(result.Done, isTrue)

			var reasoning []string
			for _, c := range result.Chunks {
				if len(c.Choices) > 0 && c.Choices[0].Delta.ReasoningContent != "" {
					reasoning = append(reasoning, c.Choices[0].Delta.ReasoningContent)
				}
			}
			combined := strings.Join(reasoning, "")
			so(strings.Contains(combined, "<signature>REDACTED_SIG</signature>"), isTrue)
			so(strings.Contains(combined, "<redacted_thinking>true</redacted_thinking>"), isTrue)
		})

		// ---- §6 tool_use content_block_start / input_json_delta -----------
		cv("§6: tool_use block → OpenAI tool_calls 格式", func() {
			result := translateSSE(
				messageStart("msg_tool", "claude"),
				cbStart(0, "tool_use", map[string]any{
					"id":   "call_xyz",
					"name": "get_weather",
				}),
				cbDelta(0, "input_json_delta", map[string]any{"partial_json": `{"city`}),
				cbDelta(0, "input_json_delta", map[string]any{"partial_json": `":"Beijing"}`}),
				cbStop(0),
				msgDelta("tool_use", nil),
				messageStop(),
			)
			so(result.Done, isTrue)

			// 找到 tool_calls chunk（content_block_start 时发出的）
			var foundToolStart bool
			var argsAccum strings.Builder
			for _, c := range result.Chunks {
				if len(c.Choices) == 0 {
					continue
				}
				for _, tc := range c.Choices[0].Delta.ToolCalls {
					if tc.ID == "call_xyz" {
						foundToolStart = true
						so(tc.Function.Name, eq, "get_weather")
					}
					argsAccum.WriteString(tc.Function.Arguments)
				}
			}
			so(foundToolStart, isTrue)
			so(argsAccum.String(), eq, `{"city":"Beijing"}`)
		})

		// ---- §17 / §20 message_delta stop_reason 映射 ---------------------
		cv("§17/§20: message_delta stop_reason → OpenAI finish_reason", func() {
			mapping := []struct {
				anthropicReason string
				openaiReason    openai.FinishReason
			}{
				{"end_turn", openai.FinishReasonStop},
				{"max_tokens", openai.FinishReasonLength},
				{"tool_use", openai.FinishReasonToolCalls},
				{"stop_sequence", openai.FinishReasonStop},
			}

			for _, m := range mapping {
				m := m
				cv(fmt.Sprintf("%s → %s", m.anthropicReason, m.openaiReason), func() {
					result := translateSSE(
						messageStart("msg_fr", "claude"),
						msgDelta(m.anthropicReason, nil),
						messageStop(),
					)
					so(result.Done, isTrue)
					var gotReason openai.FinishReason
					for _, c := range result.Chunks {
						if len(c.Choices) > 0 && c.Choices[0].FinishReason != "" {
							gotReason = c.Choices[0].FinishReason
						}
					}
					so(gotReason, eq, m.openaiReason)
				})
			}
		})

		// ---- §20 usage 映射 ------------------------------------------------
		cv("§20: message_delta usage → OpenAI prompt/completion/total tokens", func() {
			result := translateSSE(
				messageStart("msg_usage", "claude"),
				msgDelta("end_turn", map[string]any{
					"input_tokens":  100,
					"output_tokens": 50,
				}),
				messageStop(),
			)
			so(result.Done, isTrue)

			var gotUsage *openai.Usage
			for _, c := range result.Chunks {
				if c.Usage != nil {
					gotUsage = c.Usage
				}
			}
			so(gotUsage, convey.ShouldNotBeNil)
			so(gotUsage.PromptTokens, eq, 100)
			so(gotUsage.CompletionTokens, eq, 50)
			so(gotUsage.TotalTokens, eq, 150)
		})

		// ---- §17 message_stop → [DONE] ------------------------------------
		cv("§17: message_stop → data: [DONE]", func() {
			result := translateSSE(
				messageStart("msg_done", "claude"),
				messageStop(),
			)
			so(result.Done, isTrue)
		})

		// ---- §25 error event ----------------------------------------------
		cv("§25: Anthropic error event → 错误行 + [DONE]", func() {
			result := translateSSE(
				sseEvent("error", map[string]any{
					"type": "error",
					"error": map[string]any{
						"type":    "overloaded_error",
						"message": "Overloaded",
					},
				}),
			)
			so(result.Done, isTrue)
			// RawLines 中应包含 overloaded_error 信息
			found := false
			for _, line := range result.RawLines {
				if strings.Contains(line, "overloaded_error") {
					found = true
					break
				}
			}
			so(found, isTrue)
		})

		// ---- ping 被忽略 --------------------------------------------------
		cv("ping 事件被忽略，不产生任何 chunk", func() {
			result := translateSSE(
				messageStart("msg_ping", "claude"),
				fmt.Sprintf("event: ping\ndata: {}\n\n"),
				msgDelta("end_turn", nil),
				messageStop(),
			)
			so(result.Done, isTrue)
			// ping 不应产生额外 content delta
			for _, c := range result.Chunks {
				if len(c.Choices) > 0 {
					so(c.Choices[0].Delta.Content, eq, "")
				}
			}
		})

		// ---- 完整多轮 SSE 流（集成验证）--------------------------------------
		cv("完整 SSE 流集成：thinking + text + tool_use", func() {
			result := translateSSE(
				messageStart("msg_full", "claude-3-7-sonnet"),
				// thinking block
				cbStart(0, "thinking", nil),
				cbDelta(0, "thinking_delta", map[string]any{"thinking": "让我思考"}),
				cbDelta(0, "signature_delta", map[string]any{"signature": "SIG_FULL"}),
				cbStop(0),
				// text block
				cbStart(1, "text", nil),
				cbDelta(1, "text_delta", map[string]any{"text": "好的"}),
				cbStop(1),
				// tool_use block
				cbStart(2, "tool_use", map[string]any{"id": "c99", "name": "search"}),
				cbDelta(2, "input_json_delta", map[string]any{"partial_json": `{"q":"go"}`}),
				cbStop(2),
				// message_delta with usage
				msgDelta("tool_use", map[string]any{
					"input_tokens":  200,
					"output_tokens": 80,
				}),
				messageStop(),
			)
			so(result.Done, isTrue)

			var hasReasoning, hasContent, hasToolCall bool
			var gotUsage *openai.Usage
			for _, c := range result.Chunks {
				if len(c.Choices) == 0 {
					continue
				}
				d := c.Choices[0].Delta
				if d.ReasoningContent != "" {
					hasReasoning = true
				}
				if d.Content != "" {
					hasContent = true
				}
				if len(d.ToolCalls) > 0 {
					hasToolCall = true
				}
				if c.Usage != nil {
					gotUsage = c.Usage
				}
			}

			so(hasReasoning, isTrue)
			so(hasContent, isTrue)
			so(hasToolCall, isTrue)
			so(gotUsage, convey.ShouldNotBeNil)
			so(gotUsage.TotalTokens, eq, 280)
		})
	})
}
