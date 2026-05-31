package anthropic

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/sashabaranov/go-openai"
)

// NewSSEReader 将 Anthropic SSE 流包装为 OpenAI SSE 格式的 io.ReadCloser。
// 上层可以直接用 hutil.ReadSSEJsonData 消费返回的 reader，行为与 OpenAI 原生 SSE 完全一致。
func NewSSEReader(src io.ReadCloser) io.ReadCloser {
	pr, pw := io.Pipe()
	r := &sseTranslator{
		src:    src,
		pw:     pw,
		reader: bufio.NewReader(src),
	}
	go r.run(pw)
	return pr
}

// sseTranslator 从 src 读取 Anthropic SSE，转换后写入 pw
type sseTranslator struct {
	src    io.ReadCloser
	pw     *io.PipeWriter
	reader *bufio.Reader

	// 当前正在构建的各个 content block 状态（以 index 为 key）
	blocks map[int]*blockState
	// message_start 中的 id/model
	msgID    string
	msgModel string
}

type blockState struct {
	blockType   string // "text" | "thinking" | "tool_use" | "redacted_thinking"
	text        strings.Builder
	thinking    strings.Builder
	signature   strings.Builder
	toolID      string
	toolName    string
	toolArgs    strings.Builder
	isRedacted  bool
}

func (t *sseTranslator) run(pw *io.PipeWriter) {
	defer t.src.Close()

	var (
		eventType string
		dataLine  string
	)

	for {
		line, err := t.reader.ReadString('\n')
		line = strings.TrimRight(line, "\r\n")

		if line == "" && (eventType != "" || dataLine != "") {
			// 一个完整的 SSE 事件结束
			if eventType != "" && dataLine != "" {
				if writeErr := t.handleEvent(pw, eventType, dataLine); writeErr != nil {
					pw.CloseWithError(writeErr)
					return
				}
			}
			eventType = ""
			dataLine = ""
		} else if strings.HasPrefix(line, "event:") {
			eventType = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
		} else if strings.HasPrefix(line, "data:") {
			dataLine = strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		}

		if err != nil {
			if err == io.EOF {
				pw.Close()
			} else {
				pw.CloseWithError(err)
			}
			return
		}
	}
}

// handleEvent 将一个 Anthropic SSE 事件转换为零个或多个 OpenAI SSE 行，写入 pw
func (t *sseTranslator) handleEvent(pw *io.PipeWriter, eventType, data string) error {
	switch eventType {
	case "message_start":
		return t.onMessageStart(pw, data)
	case "content_block_start":
		return t.onContentBlockStart(pw, data)
	case "content_block_delta":
		return t.onContentBlockDelta(pw, data)
	case "content_block_stop":
		return t.onContentBlockStop(pw, data)
	case "message_delta":
		return t.onMessageDelta(pw, data)
	case "message_stop":
		return t.writeLine(pw, "data: [DONE]\n\n")
	case "error":
		return t.onError(pw, data)
	case "ping":
		// 忽略
	}
	return nil
}

func (t *sseTranslator) onMessageStart(pw *io.PipeWriter, data string) error {
	var ev MessageStartEvent
	if err := json.Unmarshal([]byte(data), &ev); err != nil {
		return nil // 容错
	}
	t.msgID = ev.Message.ID
	t.msgModel = ev.Message.Model
	t.blocks = make(map[int]*blockState)

	// 输出初始 chunk：携带 id、model、role
	chunk := openai.ChatCompletionStreamResponse{
		ID:    ev.Message.ID,
		Model: ev.Message.Model,
		Choices: []openai.ChatCompletionStreamChoice{{
			Delta: openai.ChatCompletionStreamChoiceDelta{
				Role: openai.ChatMessageRoleAssistant,
			},
		}},
	}
	return t.writeChunk(pw, chunk)
}

func (t *sseTranslator) onContentBlockStart(pw *io.PipeWriter, data string) error {
	var ev ContentBlockStartEvent
	if err := json.Unmarshal([]byte(data), &ev); err != nil {
		return nil
	}

	bs := &blockState{
		blockType:  ev.ContentBlock.Type,
		toolID:     ev.ContentBlock.ID,
		toolName:   ev.ContentBlock.Name,
		isRedacted: ev.ContentBlock.Type == "redacted_thinking",
	}
	if t.blocks == nil {
		t.blocks = make(map[int]*blockState)
	}
	t.blocks[ev.Index] = bs

	// tool_use block 开始时，输出一个带 tool_call id 和 name 的 chunk
	if ev.ContentBlock.Type == "tool_use" {
		idx := ev.Index
		chunk := openai.ChatCompletionStreamResponse{
			ID:    t.msgID,
			Model: t.msgModel,
			Choices: []openai.ChatCompletionStreamChoice{{
				Delta: openai.ChatCompletionStreamChoiceDelta{
					ToolCalls: []openai.ToolCall{{
						Index: &idx,
						ID:    ev.ContentBlock.ID,
						Type:  openai.ToolTypeFunction,
						Function: openai.FunctionCall{
							Name: ev.ContentBlock.Name,
						},
					}},
				},
			}},
		}
		return t.writeChunk(pw, chunk)
	}

	return nil
}

func (t *sseTranslator) onContentBlockDelta(pw *io.PipeWriter, data string) error {
	var ev ContentBlockDeltaEvent
	if err := json.Unmarshal([]byte(data), &ev); err != nil {
		return nil
	}

	bs := t.blocks[ev.Index]
	if bs == nil {
		return nil
	}

	switch ev.Delta.Type {
	case "text_delta":
		bs.text.WriteString(ev.Delta.Text)
		chunk := openai.ChatCompletionStreamResponse{
			ID:    t.msgID,
			Model: t.msgModel,
			Choices: []openai.ChatCompletionStreamChoice{{
				Delta: openai.ChatCompletionStreamChoiceDelta{
					Content: ev.Delta.Text,
				},
			}},
		}
		return t.writeChunk(pw, chunk)

	case "thinking_delta":
		bs.thinking.WriteString(ev.Delta.Thinking)
		chunk := openai.ChatCompletionStreamResponse{
			ID:    t.msgID,
			Model: t.msgModel,
			Choices: []openai.ChatCompletionStreamChoice{{
				Delta: openai.ChatCompletionStreamChoiceDelta{
					ReasoningContent: ev.Delta.Thinking,
				},
			}},
		}
		return t.writeChunk(pw, chunk)

	case "signature_delta":
		bs.signature.WriteString(ev.Delta.Signature)
		// signature 嵌入 reasoning_content 末尾，格式：\n\n<signature>SIG</signature>
		// 如果是 redacted_thinking，额外附加 redacted 标记
		sigText := signaturePrefix + ev.Delta.Signature + signatureSuffix
		if bs.isRedacted {
			sigText += redactedThinkingMark
		}
		chunk := openai.ChatCompletionStreamResponse{
			ID:    t.msgID,
			Model: t.msgModel,
			Choices: []openai.ChatCompletionStreamChoice{{
				Delta: openai.ChatCompletionStreamChoiceDelta{
					ReasoningContent: sigText,
				},
			}},
		}
		return t.writeChunk(pw, chunk)

	case "input_json_delta":
		bs.toolArgs.WriteString(ev.Delta.PartialJSON)
		idx := ev.Index
		chunk := openai.ChatCompletionStreamResponse{
			ID:    t.msgID,
			Model: t.msgModel,
			Choices: []openai.ChatCompletionStreamChoice{{
				Delta: openai.ChatCompletionStreamChoiceDelta{
					ToolCalls: []openai.ToolCall{{
						Index: &idx,
						Function: openai.FunctionCall{
							Arguments: ev.Delta.PartialJSON,
						},
					}},
				},
			}},
		}
		return t.writeChunk(pw, chunk)
	}

	return nil
}

func (t *sseTranslator) onContentBlockStop(_ *io.PipeWriter, data string) error {
	// 不需要输出任何 chunk，状态已在 delta 阶段维护
	return nil
}

func (t *sseTranslator) onMessageDelta(pw *io.PipeWriter, data string) error {
	var ev MessageDeltaEvent
	if err := json.Unmarshal([]byte(data), &ev); err != nil {
		return nil
	}

	finishReason := mapStopReason(ev.Delta.StopReason)

	chunk := openai.ChatCompletionStreamResponse{
		ID:    t.msgID,
		Model: t.msgModel,
		Choices: []openai.ChatCompletionStreamChoice{{
			FinishReason: finishReason,
			Delta:        openai.ChatCompletionStreamChoiceDelta{},
		}},
	}

	// 映射 usage
	if ev.Usage != nil {
		promptTokens := ev.Usage.InputTokens
		completionTokens := ev.Usage.OutputTokens
		chunk.Usage = &openai.Usage{
			PromptTokens:     promptTokens,
			CompletionTokens: completionTokens,
			TotalTokens:      promptTokens + completionTokens,
		}
	}

	return t.writeChunk(pw, chunk)
}

func (t *sseTranslator) onError(_ *io.PipeWriter, data string) error {
	var ev ErrorEvent
	if err := json.Unmarshal([]byte(data), &ev); err != nil {
		return fmt.Errorf("anthropic API error: unknown")
	}
	return fmt.Errorf("anthropic API error: %s: %s", ev.Error.Type, ev.Error.Message)
}

// mapStopReason 将 Anthropic stop_reason 映射为 OpenAI finish_reason
func mapStopReason(reason string) openai.FinishReason {
	switch reason {
	case "end_turn":
		return openai.FinishReasonStop
	case "max_tokens":
		return openai.FinishReasonLength
	case "tool_use":
		return openai.FinishReasonToolCalls
	case "stop_sequence":
		return openai.FinishReasonStop
	default:
		return openai.FinishReason(reason)
	}
}

// writeChunk 将 OpenAI chunk 序列化后写入 pw
func (t *sseTranslator) writeChunk(pw *io.PipeWriter, chunk openai.ChatCompletionStreamResponse) error {
	b, err := json.Marshal(chunk)
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	buf.WriteString("data: ")
	buf.Write(b)
	buf.WriteString("\n\n")
	_, err = pw.Write(buf.Bytes())
	return err
}

// writeLine 直接写一行原始数据到 pw
func (t *sseTranslator) writeLine(pw *io.PipeWriter, line string) error {
	_, err := io.WriteString(pw, line)
	return err
}
