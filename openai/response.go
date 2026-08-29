package openai

import "github.com/sashabaranov/go-openai"

// MergeResponseDeltaChunks 合并多个 SSE 响应段，从 chunks[0] 开始做前缀合并，
// 遇到格式不同时立即停止，返回已合并的结果和合并个数。调用方可用 mergedCount
// 切掉已处理的前缀，再对剩余 chunks 继续调用以拆分后续段。
//
// 注意: 本函数仅处理 Choices[0] 的 delta，多 choice 场景不适用。
//
// delta 内容按语义分为以下几类，同类才合并:
//   - empty/role-only：空 delta 或仅含 role 字段，透明地吸收进后续包。
//   - reasoning：Delta.ReasoningContent != ""
//   - content：Delta.Content != ""
//   - tool_calls：len(Delta.ToolCalls) > 0，连续的 tool_calls 包（含跨 index）合并为同一段。
//   - finish：Choices[0].FinishReason != ""，单独成段，不与其他类型合并。
//   - usage：len(Choices)==0 且 Usage != nil
//   - mixed：同时含多种实质内容类型，单独成段，不与其他类型合并。
//
// 字段合并规则：
//   - ID / Object / Created / Model / SystemFingerprint：取第一个包的值
//   - Usage：取最后一个非 nil 的值
//   - Choices[0].FinishReason：取最后一个非空的值
//   - Choices[0].Delta.Role：取第一个非空的值
//   - Choices[0].Delta.Content / ReasoningContent：拼接
//   - Choices[0].Delta.ToolCalls：按 Index 合并，Arguments 拼接；新 Index 追加
func MergeResponseDeltaChunks(
	chunks ...openai.ChatCompletionStreamResponse,
) (response openai.ChatCompletionStreamResponse, mergedCount int) {
	if len(chunks) == 0 {
		return openai.ChatCompletionStreamResponse{}, 0
	}

	response = chunks[0]
	mergedCount = 1
	baseType := classifyChunk(chunks[0])

	for i := 1; i < len(chunks); i++ {
		ct := classifyChunk(chunks[i])

		canMerge := false
		switch {
		case ct == chunkTypeEmpty:
			// empty/role-only 包始终透明吸收
			canMerge = true
		case ct == chunkTypeFinish || ct == chunkTypeMixed:
			// finish 和 mixed 不并入任何段
		case baseType == chunkTypeEmpty:
			// 当前段尚未确定类型，吸收第一个有实质内容的包并锁定类型
			canMerge = true
			baseType = ct
		case ct == baseType:
			canMerge = true
		}

		if !canMerge {
			break
		}
		mergeChunkInto(&response, chunks[i])
		mergedCount++
	}
	return response, mergedCount
}

// chunkType 是 delta 内容的语义分类，用于 MergeResponseDeltaChunks 内部合并判断。
type chunkType int

const (
	chunkTypeEmpty     chunkType = iota // 空 delta 或仅含 role 字段
	chunkTypeReasoning                  // Delta.ReasoningContent != ""
	chunkTypeContent                    // Delta.Content != ""
	chunkTypeToolCalls                  // len(Delta.ToolCalls) > 0
	chunkTypeFinish                     // Choices[0].FinishReason != ""
	chunkTypeUsageOnly                  // len(Choices)==0 且 Usage != nil
	chunkTypeMixed                      // 同时含多种实质内容类型
)

func classifyChunk(c openai.ChatCompletionStreamResponse) chunkType {
	if len(c.Choices) == 0 {
		if c.Usage != nil {
			return chunkTypeUsageOnly
		}
		return chunkTypeEmpty
	}

	choice := c.Choices[0]
	delta := choice.Delta

	hasToolCalls := len(delta.ToolCalls) > 0
	hasReasoning := delta.ReasoningContent != ""
	hasContent := delta.Content != ""

	realCount := 0
	if hasToolCalls {
		realCount++
	}
	if hasReasoning {
		realCount++
	}
	if hasContent {
		realCount++
	}

	switch {
	case realCount > 1:
		return chunkTypeMixed
	case hasToolCalls:
		return chunkTypeToolCalls
	case hasReasoning:
		return chunkTypeReasoning
	case hasContent:
		return chunkTypeContent
	case choice.FinishReason != "":
		// finish 优先级低于实质内容：只有 delta 中没有任何内容时才归为 finish 段。
		// 若包同时带 content/reasoning + finish_reason，仍归入内容类型，
		// finish_reason 由 mergeChunkInto 写入合并结果。
		return chunkTypeFinish
	default:
		return chunkTypeEmpty
	}
}

// mergeChunkInto 将 src 合并进 dst，仅处理 Choices[0]。
func mergeChunkInto(dst *openai.ChatCompletionStreamResponse, src openai.ChatCompletionStreamResponse) {
	if src.Usage != nil {
		dst.Usage = src.Usage
	}
	if len(src.Choices) == 0 {
		return
	}
	if len(dst.Choices) == 0 {
		dst.Choices = []openai.ChatCompletionStreamChoice{src.Choices[0]}
		return
	}

	srcChoice := src.Choices[0]
	dstChoice := &dst.Choices[0]
	srcDelta := srcChoice.Delta
	dstDelta := &dstChoice.Delta

	if dstDelta.Role == "" && srcDelta.Role != "" {
		dstDelta.Role = srcDelta.Role
	}
	dstDelta.Content += srcDelta.Content
	dstDelta.ReasoningContent += srcDelta.ReasoningContent
	if srcChoice.FinishReason != "" {
		dstChoice.FinishReason = srcChoice.FinishReason
	}

	for _, tc := range srcDelta.ToolCalls {
		srcIdx := 0
		if tc.Index != nil {
			srcIdx = *tc.Index
		}
		merged := false
		for j := range dstDelta.ToolCalls {
			dstIdx := 0
			if dstDelta.ToolCalls[j].Index != nil {
				dstIdx = *dstDelta.ToolCalls[j].Index
			}
			if dstIdx != srcIdx {
				continue
			}
			existing := &dstDelta.ToolCalls[j]
			existing.Function.Arguments += tc.Function.Arguments
			if existing.ID == "" {
				existing.ID = tc.ID
			}
			if existing.Type == "" {
				existing.Type = tc.Type
			}
			if existing.Function.Name == "" {
				existing.Function.Name = tc.Function.Name
			}
			merged = true
			break
		}
		if !merged {
			idx := srcIdx
			dstDelta.ToolCalls = append(dstDelta.ToolCalls, openai.ToolCall{
				ID:   tc.ID,
				Type: tc.Type,
				Function: openai.FunctionCall{
					Name:      tc.Function.Name,
					Arguments: tc.Function.Arguments,
				},
				Index: &idx,
			})
		}
	}
}
