package poker

import (
	"testing"
)

func TestFaroShuffle(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		expected []int
	}{
		{
			name:     "空数组",
			input:    []int{},
			expected: []int{},
		},
		{
			name:     "单个元素",
			input:    []int{1},
			expected: []int{1},
		},
		{
			name:     "两个元素",
			input:    []int{1, 2},
			expected: []int{1, 2},
		},
		{
			name:     "四个元素",
			input:    []int{1, 2, 3, 4},
			expected: []int{1, 3, 2, 4},
		},
		{
			name:     "六个元素",
			input:    []int{1, 2, 3, 4, 5, 6},
			expected: []int{1, 4, 2, 5, 3, 6},
		},
		{
			name:     "八个元素",
			input:    []int{1, 2, 3, 4, 5, 6, 7, 8},
			expected: []int{1, 5, 2, 6, 3, 7, 4, 8},
		},
		{
			name:     "奇数个元素(5)",
			input:    []int{1, 2, 3, 4, 5},
			expected: []int{1, 3, 2, 4, 5},
		},
		{
			name:     "奇数个元素(7)",
			input:    []int{1, 2, 3, 4, 5, 6, 7},
			expected: []int{1, 4, 2, 5, 3, 6, 7},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建测试用的牌堆
			deck := &Deck{
				cards: make([]Card, len(tt.input)),
			}

			// 用点数来模拟，方便验证
			for i, v := range tt.input {
				deck.cards[i] = Card{point: Point(v)}
			}

			// 执行洗牌
			deck.FaroShuffle()

			// 验证结果
			if len(deck.cards) != len(tt.expected) {
				t.Fatalf("长度不匹配: got %d, want %d", len(deck.cards), len(tt.expected))
			}

			for i, expected := range tt.expected {
				if int(deck.cards[i].point) != expected {
					t.Errorf("位置 %d: got %d, want %d", i, deck.cards[i].point, expected)
				}
			}

			// 打印结果以便调试
			result := make([]int, len(deck.cards))
			for i := range deck.cards {
				result[i] = int(deck.cards[i].point)
			}
			t.Logf("结果: %v", result)
		})
	}
}

// 测试实际的扑克牌
func TestFaroShuffleWithRealDeck(t *testing.T) {
	deck := NewDeck()

	t.Logf("洗牌前: %v", deck.cards[:10])

	deck.FaroShuffle()

	t.Logf("洗牌后: %v", deck.cards[:10])

	// 验证牌的数量没有变化
	if deck.Len() != 54 {
		t.Errorf("牌的数量错误: got %d, want 54", deck.Len())
	}
}

// 测试多次 Faro Shuffle 的循环性质
// 完美洗牌有一个有趣的性质: 对于 52 张牌, 执行 8 次 Faro Shuffle 会回到原始顺序
func TestFaroShuffleCycle(t *testing.T) {
	deck1 := NewDeck()

	// 保存原始顺序
	original := make([]Card, deck1.Len())
	copy(original, deck1.cards)

	// 执行多次洗牌
	maxIterations := 100
	for i := 1; i <= maxIterations; i++ {
		deck1.FaroShuffle()

		// 检查是否回到原始顺序
		isOriginal := true
		for j := range deck1.cards {
			if deck1.cards[j] != original[j] {
				isOriginal = false
				break
			}
		}

		if isOriginal {
			t.Logf("经过 %d 次 Faro Shuffle 后回到原始顺序", i)
			return
		}
	}

	t.Logf("在 %d 次洗牌内没有回到原始顺序", maxIterations)
}

// 测试 In-shuffle 版本
func TestFaroShuffleIn(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		expected []int
	}{
		{
			name:     "空数组",
			input:    []int{},
			expected: []int{},
		},
		{
			name:     "单个元素",
			input:    []int{1},
			expected: []int{1},
		},
		{
			name:     "两个元素",
			input:    []int{1, 2},
			expected: []int{2, 1},
		},
		{
			name:     "四个元素",
			input:    []int{1, 2, 3, 4},
			expected: []int{3, 1, 4, 2},
		},
		{
			name:     "六个元素",
			input:    []int{1, 2, 3, 4, 5, 6},
			expected: []int{4, 1, 5, 2, 6, 3},
		},
		{
			name:     "八个元素",
			input:    []int{1, 2, 3, 4, 5, 6, 7, 8},
			expected: []int{5, 1, 6, 2, 7, 3, 8, 4},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建测试用的牌堆
			deck := &Deck{
				cards: make([]Card, len(tt.input)),
			}

			// 用点数来模拟，方便验证
			for i, v := range tt.input {
				deck.cards[i] = Card{point: Point(v)}
			}

			// 执行洗牌
			deck.FaroShuffleIn()

			// 验证结果
			if len(deck.cards) != len(tt.expected) {
				t.Fatalf("长度不匹配: got %d, want %d", len(deck.cards), len(tt.expected))
			}

			for i, expected := range tt.expected {
				if int(deck.cards[i].point) != expected {
					t.Errorf("位置 %d: got %d, want %d", i, deck.cards[i].point, expected)
				}
			}

			// 打印结果以便调试
			result := make([]int, len(deck.cards))
			for i := range deck.cards {
				result[i] = int(deck.cards[i].point)
			}
			t.Logf("结果: %v", result)
		})
	}
}
