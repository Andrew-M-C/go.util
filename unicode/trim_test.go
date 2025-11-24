package unicode

import (
	"fmt"
	"testing"
)

func TestTrimUTF8(t *testing.T) {
	cv("无限制情况", t, func() { testTrimUTF8NoLimit(t) })
	cv("只限制字符数", t, func() { testTrimUTF8RuneLimit(t) })
	cv("只限制字节数", t, func() { testTrimUTF8ByteLimit(t) })
	cv("同时限制字符数和字节数", t, func() { testTrimUTF8BothLimits(t) })
	cv("tail 为空字符串", t, func() { testTrimUTF8EmptyTail(t) })
	cv("tail 超限情况", t, func() { testTrimUTF8TailExceeds(t) })
	cv("边界情况", t, func() { testTrimUTF8EdgeCases(t) })
	cv("多字节字符", t, func() { testTrimUTF8MultiByte(t) })
	cv("特殊字符", t, func() { testTrimUTF8SpecialChars(t) })
	cv("带调试器选项", t, func() { testTrimUTF8WithDebugger(t) })
}

// 测试无限制情况 (覆盖 line 20-22)
func testTrimUTF8NoLimit(t *testing.T) {
	cv("maxRunes=0 且 maxBytes=0", func() {
		result := TrimUTF8("Hello World", "...", 0, 0)
		so(result, eq, "Hello World")
	})

	cv("maxRunes<0 且 maxBytes<0", func() {
		result := TrimUTF8("你好世界", "...", -1, -1)
		so(result, eq, "你好世界")
	})

	cv("maxRunes=0 且 maxBytes<0", func() {
		result := TrimUTF8("Test String", "...", 0, -5)
		so(result, eq, "Test String")
	})
}

// 测试只限制字符数 (覆盖 line 45-56 的字符数检查)
func testTrimUTF8RuneLimit(t *testing.T) {
	cv("不需要截断", func() {
		result := TrimUTF8("Hello", "...", 10, 0)
		so(result, eq, "Hello")
	})

	cv("ASCII字符需要截断", func() {
		result := TrimUTF8("Hello World", "...", 8, 0)
		so(result, eq, "Hello...")
	})

	cv("中文字符需要截断", func() {
		result := TrimUTF8("你好世界，这是测试", "...", 6, 0)
		so(result, eq, "你好世...")
	})

	cv("混合字符需要截断", func() {
		result := TrimUTF8("Hello世界123", "...", 8, 0)
		so(result, eq, "Hello...")
	})

	cv("恰好等于限制", func() {
		result := TrimUTF8("Hello", "...", 5, 0)
		so(result, eq, "Hello")
	})
}

// 测试只限制字节数 (覆盖 line 50-56 的字节数检查)
func testTrimUTF8ByteLimit(t *testing.T) {
	cv("ASCII字符不需要截断", func() {
		result := TrimUTF8("Hello", "...", 0, 20)
		so(result, eq, "Hello")
	})

	cv("ASCII字符需要截断", func() {
		result := TrimUTF8("Hello World", "...", 0, 8)
		so(result, eq, "Hello...")
	})

	cv("中文字符需要截断", func() {
		// 每个中文3字节，"你好世界"=12字节
		// maxBytes=11，会触发截断（修复后的逻辑：先添加再检查）
		result := TrimUTF8("你好世界", "...", 0, 11)
		so(result, eq, "你好...")

		// maxBytes=9，会触发截断
		result = TrimUTF8("你好世界", "...", 0, 9)
		so(result, eq, "你好...")

		// maxBytes=8，会触发截断
		result = TrimUTF8("你好世界", "...", 0, 8)
		so(result, eq, "你...")
	})

	cv("混合字符需要截断", func() {
		// "Hello世界" = 5+3+3=11字节，7个字符
		// maxBytes=10 会触发截断（修复后的逻辑）
		result := TrimUTF8("Hello世界", "...", 0, 10)
		so(result, eq, "Hello...")

		// maxBytes=7 会触发截断
		result = TrimUTF8("Hello世界", "...", 0, 7)
		so(result, eq, "Hell...")
	})

	cv("恰好等于限制", func() {
		result := TrimUTF8("Hello", "...", 0, 5)
		so(result, eq, "Hello")
	})
}

// 测试同时限制字符数和字节数
func testTrimUTF8BothLimits(t *testing.T) {
	cv("字符数限制更严格", func() {
		result := TrimUTF8("ABCDEFGHIJ", "...", 6, 20)
		so(result, eq, "ABC...")
	})

	cv("字节数限制更严格", func() {
		result := TrimUTF8("你好世界天地", "...", 10, 12)
		so(result, eq, "你好世...")
	})

	cv("两个限制都宽松", func() {
		result := TrimUTF8("Hello", "...", 10, 20)
		so(result, eq, "Hello")
	})

	cv("复杂混合字符串", func() {
		result := TrimUTF8("English中文日本語한국어", "...", 12, 30)
		so(result, eq, "English中文...")
	})
}

// 测试 tail 为空字符串
func testTrimUTF8EmptyTail(t *testing.T) {
	cv("空tail按字符数截断", func() {
		result := TrimUTF8("Hello World", "", 5, 0)
		so(result, eq, "Hello")
	})

	cv("空tail按字节数截断", func() {
		result := TrimUTF8("你好世界", "", 0, 8)
		so(result, eq, "你好")
	})

	cv("空tail不需要截断", func() {
		result := TrimUTF8("Hello", "", 10, 20)
		so(result, eq, "Hello")
	})
}

// 测试 tail 超限的情况 (覆盖 line 24-34)
func testTrimUTF8TailExceeds(t *testing.T) {
	cv("tail字符数超过maxRunes", func() {
		result := TrimUTF8("Hello", "......", 5, 0)
		so(result, eq, "Hello")
	})

	cv("tail字节数超过maxBytes", func() {
		result := TrimUTF8("Hello", "......", 0, 5)
		so(result, eq, "Hello")
	})

	cv("tail同时超过两个限制", func() {
		result := TrimUTF8("Hello World", "很长很长的tail字符串", 10, 10)
		so(result, eq, "Hello Worl")
	})

	cv("中文tail超过限制", func() {
		result := TrimUTF8("测试", "很长的尾部", 5, 0)
		so(result, eq, "测试")
	})
}

// 测试边界情况
func testTrimUTF8EdgeCases(t *testing.T) {
	cv("空字符串输入", func() {
		result := TrimUTF8("", "...", 5, 10)
		so(result, eq, "")
	})

	cv("只能容纳tail", func() {
		result := TrimUTF8("Hello World", "...", 3, 3)
		so(result, eq, "...")
	})

	cv("极小的限制", func() {
		// tail超限，被设为空，maxRunes=1会触发截断，只保留1个字符
		result := TrimUTF8("你好", "...", 1, 0)
		so(result, eq, "你")
	})

	cv("所有内容被截断只剩tail", func() {
		result := TrimUTF8("Hello World", "...", 3, 0)
		so(result, eq, "...")
	})

	cv("长字符串截断", func() {
		result := TrimUTF8("这是一个非常非常非常非常非常非常非常非常非常非常非常非常长的字符串用来测试性能和正确性", "...", 20, 0)
		so(result, eq, "这是一个非常非常非常非常非常非常非...")
	})

	cv("线上 bug", func() {
		result := TrimUTF8("【活动】绿洲打卡🧧  ", "…", 10, 30, WithDebugger(t.Logf))
		so(result, eq, "【活动】绿洲打卡…")
	})
}

// 测试多字节字符
func testTrimUTF8MultiByte(t *testing.T) {
	cv("包含emoji表情", func() {
		result := TrimUTF8("Hello😀World🎉", "...", 10, 0)
		so(result, eq, "Hello😀W...")
	})

	cv("日文字符", func() {
		// "こんにちは世界" 有 7 个字符，maxRunes=5 会触发截断
		result := TrimUTF8("こんにちは世界", "...", 5, 0)
		so(result, eq, "こん...")
	})

	cv("韩文字符", func() {
		// "안녕하세요" 有 5 个字符，maxRunes=3 会触发截断
		result := TrimUTF8("안녕하세요", "...", 3, 0)
		so(result, eq, "...")
	})

	cv("混合多国语言", func() {
		// "Hello你好こんにちは" 有 12 个字符，需要更小的限制才会触发截断
		result := TrimUTF8("Hello你好こんにちは", "...", 10, 0)
		so(result, eq, "Hello你好...")
	})
}

// 测试特殊字符
func testTrimUTF8SpecialChars(t *testing.T) {
	cv("包含换行符", func() {
		result := TrimUTF8("Hello\nWorld\n", "...", 10, 0)
		so(result, eq, "Hello\nW...")
	})

	cv("包含制表符", func() {
		result := TrimUTF8("Hello\tWorld", "...", 9, 0)
		so(result, eq, "Hello\t...")
	})

	cv("包含空格", func() {
		// "Hello World Test" 有 16 个字符，maxRunes=12 会截断成 12 个字符
		result := TrimUTF8("Hello World Test", "...", 12, 0)
		so(result, eq, "Hello Wor...")
	})

	cv("特殊Unicode字符", func() {
		// "Test→←↑↓" 有 8 个字符，maxRunes=6 会触发截断
		result := TrimUTF8("Test→←↑↓", "...", 6, 0)
		so(result, eq, "Tes...")
	})
}

// 测试带调试器选项 (覆盖 line 26, 32 的 debug 调用)
func testTrimUTF8WithDebugger(t *testing.T) {
	cv("tail超过字符数限制时触发debug", func() {
		debugMessages := []string{}
		debugFunc := func(format string, args ...any) {
			msg := fmt.Sprintf(format, args...)
			debugMessages = append(debugMessages, msg)
		}

		result := TrimUTF8("Hello", "很长的尾部字符串", 5, 0, WithDebugger(debugFunc))
		so(result, eq, "Hello")
		// 修复后的代码会输出更多debug信息：tail信息、无需截断信息、以及tail超限信息
		so(len(debugMessages) >= 1, eq, true)
	})

	cv("tail超过字节数限制时触发debug", func() {
		debugMessages := []string{}
		debugFunc := func(format string, args ...any) {
			msg := fmt.Sprintf(format, args...)
			debugMessages = append(debugMessages, msg)
		}

		result := TrimUTF8("Hello", "very long tail string", 0, 5, WithDebugger(debugFunc))
		so(result, eq, "Hello")
		// 修复后的代码会输出更多debug信息
		so(len(debugMessages) >= 1, eq, true)
	})

	cv("正常截断会输出debug信息", func() {
		debugMessages := []string{}
		debugFunc := func(format string, args ...any) {
			msg := fmt.Sprintf(format, args...)
			debugMessages = append(debugMessages, msg)
		}

		result := TrimUTF8("Hello World", "...", 8, 0, WithDebugger(debugFunc))
		so(result, eq, "Hello...")
		// 修复后的代码在正常截断时也会输出debug信息（tail信息、需截断信息）
		so(len(debugMessages) >= 2, eq, true)
	})

	cv("nil调试函数也能正常工作", func() {
		result := TrimUTF8("Hello World", "...", 8, 0, WithDebugger(nil))
		so(result, eq, "Hello...")
	})
}

// BenchmarkTrimUTF8 性能测试
func BenchmarkTrimUTF8(b *testing.B) {
	testCases := []struct {
		name     string
		orig     string
		tail     string
		maxRunes int
		maxBytes int
	}{
		{
			name:     "short_ascii",
			orig:     "Hello World",
			tail:     "...",
			maxRunes: 8,
			maxBytes: 0,
		},
		{
			name:     "short_chinese",
			orig:     "你好世界，这是测试",
			tail:     "...",
			maxRunes: 6,
			maxBytes: 0,
		},
		{
			name:     "long_mixed",
			orig:     "这是一个非常非常非常非常非常非常非常非常非常非常非常非常长的字符串用来测试性能和正确性Hello World 123",
			tail:     "...",
			maxRunes: 30,
			maxBytes: 100,
		},
		{
			name:     "no_trim",
			orig:     "Short",
			tail:     "...",
			maxRunes: 100,
			maxBytes: 200,
		},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = TrimUTF8(tc.orig, tc.tail, tc.maxRunes, tc.maxBytes)
			}
		})
	}
}

// ExampleTrimUTF8 示例函数
func ExampleTrimUTF8() {
	// 基本用法：限制字符数
	result := TrimUTF8("Hello World", "...", 8, 0)
	fmt.Println(result)

	// 限制字节数
	result = TrimUTF8("你好世界测试", "...", 0, 14)
	fmt.Println(result)

	// 同时限制字符数和字节数
	result = TrimUTF8("Hello 世界", "...", 10, 15)
	fmt.Println(result)

	// Output:
	// Hello...
	// 你好世...
	// Hello 世界
}
