package slices

// CombineEvenly 将两个切片均匀混合成一个新切片。
// 两个切片的元素保持各自的相对顺序，同时尽可能均匀地分布在结果中。
//
// 算法思路：将较长的切片均匀分成 (短切片长度+1) 段，然后在各段之间插入短切片的元素。
// 使用纯整数运算实现均匀分段，避免浮点数精度问题。
func CombineEvenly[T any](s1, s2 []T) []T {
	// 确保 s1 是较长的切片
	if len(s1) < len(s2) {
		s1, s2 = s2, s1
	}

	longLen := len(s1)
	shortLen := len(s2)

	// 边界情况处理
	if longLen == 0 {
		return nil
	}
	if shortLen == 0 {
		// 只有一个切片有元素，直接复制返回
		result := make([]T, longLen)
		copy(result, s1)
		return result
	}

	// 将长切片 s1 均匀分成 (shortLen+1) 段，在各段之间插入短切片 s2 的元素
	//
	// 例1（整除情况）：s1=[A,B,C,D,E,F,G,H](8个), s2=[+,-,*](3个)
	//   分成 4 段，每段 8/4=2 个：[A,B] | [C,D] | [E,F] | [G,H]
	//   结果：A,B,+,C,D,-,E,F,*,G,H
	//
	// 例2（不整除情况）：s1=[A,B,C,D,E,F,G,H,I,J,K,L](12个), s2=[+,-,*,/](4个)
	//   分成 5 段，12/5=2.4，利用整数除法自动产生 2 和 3 交替的段长：
	//   第0段结束位置: 1*12/5=2  → 段长 2: [A,B]
	//   第1段结束位置: 2*12/5=4  → 段长 2: [C,D]
	//   第2段结束位置: 3*12/5=7  → 段长 3: [E,F,G]
	//   第3段结束位置: 4*12/5=9  → 段长 2: [H,I]
	//   第4段结束位置: 5*12/5=12 → 段长 3: [J,K,L]
	//   结果：A,B,+,C,D,-,E,F,G,*,H,I,/,J,K,L

	result := make([]T, 0, longLen+shortLen)
	longIdx := 0
	segments := shortLen + 1

	for i := 0; i < segments; i++ {
		// 计算第 i 段结束时，长切片应该累计取了多少个元素
		// 使用整数除法：(i+1) * longLen / segments
		endPos := (i + 1) * longLen / segments

		// 添加本段的长切片元素
		for longIdx < endPos {
			result = append(result, s1[longIdx])
			longIdx++
		}

		// 在段后插入短切片元素（最后一段之后不插入）
		if i < shortLen {
			result = append(result, s2[i])
		}
	}

	return result
}
