package minheap_test

import (
	"testing"

	"github.com/Andrew-M-C/go.util/datastructure/minheap"
	"github.com/smartystreets/goconvey/convey"
)

var (
	cv = convey.Convey
	so = convey.So
	eq = convey.ShouldEqual

	isTrue = convey.ShouldBeTrue
)

func TestBasic(t *testing.T) {
	cv("测试 Basic 类型的基本功能", t, func() { testBasicPushPop(t) })
	cv("测试 Basic 类型的最小堆特性", t, func() { testBasicMinHeap(t) })
	cv("测试 Basic 类型的不同数据类型", t, func() { testBasicDifferentTypes(t) })
	cv("测试 Basic 类型的 Len 方法", t, func() { testBasicLen(t) })
	cv("测试 Basic 类型的空堆操作", t, func() { testBasicEmptyHeap(t) })
}

func TestHeap(t *testing.T) {
	cv("测试 Heap 类型的基本功能", t, func() { testHeapPushPop(t) })
	cv("测试 Heap 类型的最小堆特性", t, func() { testHeapMinHeap(t) })
	cv("测试 Heap 类型的复杂场景", t, func() { testHeapComplexScenarios(t) })
	cv("测试 Heap 类型的 Len 方法", t, func() { testHeapLen(t) })
	cv("测试 Heap 类型的空堆操作", t, func() { testHeapEmptyHeap(t) })
}

// ============================================================================
// Basic 类型测试
// ============================================================================

func testBasicPushPop(t *testing.T) {
	cv("基本的 Push 和 Pop 操作", func() {
		var h minheap.Basic[int]

		so(h.Len(), eq, 0)

		h.Push(5)
		so(h.Len(), eq, 1)

		h.Push(3)
		h.Push(7)
		h.Push(1)
		so(h.Len(), eq, 4)

		v, ok := h.Pop()
		so(ok, isTrue)
		so(v, eq, 1) // 最小的元素
		so(h.Len(), eq, 3)

		v, ok = h.Pop()
		so(ok, isTrue)
		so(v, eq, 3)
		so(h.Len(), eq, 2)

		v, ok = h.Pop()
		so(ok, isTrue)
		so(v, eq, 5)
		so(h.Len(), eq, 1)

		v, ok = h.Pop()
		so(ok, isTrue)
		so(v, eq, 7)
		so(h.Len(), eq, 0)
	})
}

func testBasicMinHeap(t *testing.T) {
	cv("最小堆 - 按升序弹出整数", func() {
		var h minheap.Basic[int]

		// 无序插入
		values := []int{50, 10, 30, 20, 40, 5, 15, 25}
		for _, v := range values {
			h.Push(v)
		}

		so(h.Len(), eq, len(values))

		// 应该按升序弹出
		expected := []int{5, 10, 15, 20, 25, 30, 40, 50}
		for _, exp := range expected {
			v, ok := h.Pop()
			so(ok, isTrue)
			so(v, eq, exp)
		}

		so(h.Len(), eq, 0)
	})

	cv("最小堆 - 处理负数", func() {
		var h minheap.Basic[int]

		values := []int{-5, 10, -20, 0, 15, -10}
		for _, v := range values {
			h.Push(v)
		}

		// 应该按升序弹出
		expected := []int{-20, -10, -5, 0, 10, 15}
		for _, exp := range expected {
			v, ok := h.Pop()
			so(ok, isTrue)
			so(v, eq, exp)
		}
	})

	cv("最小堆 - 处理重复值", func() {
		var h minheap.Basic[int]

		values := []int{5, 3, 5, 1, 3, 1, 5}
		for _, v := range values {
			h.Push(v)
		}

		expected := []int{1, 1, 3, 3, 5, 5, 5}
		for _, exp := range expected {
			v, ok := h.Pop()
			so(ok, isTrue)
			so(v, eq, exp)
		}
	})
}

func testBasicDifferentTypes(t *testing.T) {
	cv("测试 float64 类型", func() {
		var h minheap.Basic[float64]

		values := []float64{3.14, 2.71, 1.41, 1.73, 2.23}
		for _, v := range values {
			h.Push(v)
		}

		expected := []float64{1.41, 1.73, 2.23, 2.71, 3.14}
		for _, exp := range expected {
			v, ok := h.Pop()
			so(ok, isTrue)
			so(v, eq, exp)
		}
	})

	cv("测试 string 类型", func() {
		var h minheap.Basic[string]

		values := []string{"dog", "cat", "bird", "elephant", "ant"}
		for _, v := range values {
			h.Push(v)
		}

		// 字符串按字典序排序
		expected := []string{"ant", "bird", "cat", "dog", "elephant"}
		for _, exp := range expected {
			v, ok := h.Pop()
			so(ok, isTrue)
			so(v, eq, exp)
		}
	})

	cv("测试 int64 类型", func() {
		var h minheap.Basic[int64]

		values := []int64{1000000000, 999, 1000000001, 1, 500}
		for _, v := range values {
			h.Push(v)
		}

		expected := []int64{1, 500, 999, 1000000000, 1000000001}
		for _, exp := range expected {
			v, ok := h.Pop()
			so(ok, isTrue)
			so(v, eq, exp)
		}
	})
}

func testBasicLen(t *testing.T) {
	cv("测试 Len 方法的准确性", func() {
		var h minheap.Basic[int]

		so(h.Len(), eq, 0)

		// 逐个添加，检查长度变化
		for i := 1; i <= 100; i++ {
			h.Push(i)
			so(h.Len(), eq, i)
		}

		// 逐个删除，检查长度变化
		for i := 99; i >= 0; i-- {
			_, ok := h.Pop()
			so(ok, isTrue)
			so(h.Len(), eq, i)
		}
	})

	cv("测试空堆的 Len", func() {
		var h minheap.Basic[int]
		so(h.Len(), eq, 0)

		h.Push(1)
		_, ok := h.Pop()
		so(ok, isTrue)
		so(h.Len(), eq, 0)
	})
}

func testBasicEmptyHeap(t *testing.T) {
	cv("从空堆 Pop 应该返回 false", func() {
		var h minheap.Basic[int]

		v, ok := h.Pop()
		so(ok, eq, false)
		so(v, eq, 0) // 零值
		so(h.Len(), eq, 0)
	})

	cv("多次从空堆 Pop", func() {
		var h minheap.Basic[int]

		for i := 0; i < 5; i++ {
			v, ok := h.Pop()
			so(ok, eq, false)
			so(v, eq, 0)
		}
		so(h.Len(), eq, 0)
	})

	cv("Pop 完所有元素后再 Pop", func() {
		var h minheap.Basic[int]

		h.Push(1)
		h.Push(2)
		h.Push(3)

		// Pop 所有元素
		for i := 0; i < 3; i++ {
			_, ok := h.Pop()
			so(ok, isTrue)
		}

		// 再次 Pop 应该失败
		v, ok := h.Pop()
		so(ok, eq, false)
		so(v, eq, 0)
		so(h.Len(), eq, 0)
	})

	cv("空堆 Pop 后再 Push", func() {
		var h minheap.Basic[string]

		// 空堆 Pop
		v, ok := h.Pop()
		so(ok, eq, false)
		so(v, eq, "")

		// 再 Push
		h.Push("hello")
		h.Push("world")

		v, ok = h.Pop()
		so(ok, isTrue)
		so(v, eq, "hello")

		v, ok = h.Pop()
		so(ok, isTrue)
		so(v, eq, "world")
	})
}

// ============================================================================
// Heap 类型测试
// ============================================================================

func testHeapPushPop(t *testing.T) {
	cv("基本的 Push 和 Pop 操作", func() {
		var h minheap.Heap[int, string]

		so(h.Len(), eq, 0)

		h.Push(5, "five")
		so(h.Len(), eq, 1)

		h.Push(3, "three")
		h.Push(7, "seven")
		h.Push(1, "one")
		so(h.Len(), eq, 4)

		score, value, ok := h.Pop()
		so(ok, isTrue)
		so(score, eq, 1)
		so(value, eq, "one")
		so(h.Len(), eq, 3)

		score, value, ok = h.Pop()
		so(ok, isTrue)
		so(score, eq, 3)
		so(value, eq, "three")
		so(h.Len(), eq, 2)

		score, value, ok = h.Pop()
		so(ok, isTrue)
		so(score, eq, 5)
		so(value, eq, "five")
		so(h.Len(), eq, 1)

		score, value, ok = h.Pop()
		so(ok, isTrue)
		so(score, eq, 7)
		so(value, eq, "seven")
		so(h.Len(), eq, 0)
	})
}

func testHeapMinHeap(t *testing.T) {
	cv("最小堆 - 按分数升序弹出", func() {
		var h minheap.Heap[int, string]

		// 无序插入
		data := map[int]string{
			50: "fifty",
			10: "ten",
			30: "thirty",
			20: "twenty",
			40: "forty",
			5:  "five",
			15: "fifteen",
			25: "twenty-five",
		}

		for score, value := range data {
			h.Push(score, value)
		}

		so(h.Len(), eq, len(data))

		// 应该按分数升序弹出
		expectedScores := []int{5, 10, 15, 20, 25, 30, 40, 50}
		for _, expScore := range expectedScores {
			score, value, ok := h.Pop()
			so(ok, isTrue)
			so(score, eq, expScore)
			so(value, eq, data[expScore])
		}

		so(h.Len(), eq, 0)
	})

	cv("最小堆 - 处理负分数", func() {
		var h minheap.Heap[int, string]

		data := []struct {
			score int
			value string
		}{
			{-5, "minus five"},
			{10, "ten"},
			{-20, "minus twenty"},
			{0, "zero"},
			{15, "fifteen"},
			{-10, "minus ten"},
		}

		for _, d := range data {
			h.Push(d.score, d.value)
		}

		// 应该按分数升序弹出
		expectedScores := []int{-20, -10, -5, 0, 10, 15}
		for _, expScore := range expectedScores {
			score, _, ok := h.Pop()
			so(ok, isTrue)
			so(score, eq, expScore)
		}
	})

	cv("最小堆 - 相同分数的不同值", func() {
		var h minheap.Heap[int, string]

		h.Push(5, "first")
		h.Push(3, "second")
		h.Push(5, "third")
		h.Push(3, "fourth")
		h.Push(5, "fifth")

		// 分数应该按升序弹出
		scores := []int{}
		for h.Len() > 0 {
			score, _, ok := h.Pop()
			so(ok, isTrue)
			scores = append(scores, score)
		}

		// 验证分数是升序的
		so(scores[0], eq, 3)
		so(scores[1], eq, 3)
		so(scores[2], eq, 5)
		so(scores[3], eq, 5)
		so(scores[4], eq, 5)
	})
}

func testHeapComplexScenarios(t *testing.T) {
	cv("使用自定义结构体作为值", func() {
		type Task struct {
			Name        string
			Description string
		}

		var h minheap.Heap[int, Task]

		tasks := []struct {
			priority int
			task     Task
		}{
			{3, Task{"任务C", "低优先级"}},
			{1, Task{"任务A", "高优先级"}},
			{2, Task{"任务B", "中优先级"}},
			{1, Task{"任务A2", "另一个高优先级"}},
		}

		for _, t := range tasks {
			h.Push(t.priority, t.task)
		}

		// 应该按优先级升序弹出
		priority, task, ok := h.Pop()
		so(ok, isTrue)
		so(priority, eq, 1)
		// 两个优先级为 1 的任务，名字都包含 "任务A"
		so(len(task.Name) >= 3, isTrue)

		priority, task, ok = h.Pop()
		so(ok, isTrue)
		so(priority, eq, 1)
		so(len(task.Name) >= 3, isTrue)

		priority, task, ok = h.Pop()
		so(ok, isTrue)
		so(priority, eq, 2)
		so(task.Name, eq, "任务B")

		priority, task, ok = h.Pop()
		so(ok, isTrue)
		so(priority, eq, 3)
		so(task.Name, eq, "任务C")
	})

	cv("使用 float64 作为分数", func() {
		var h minheap.Heap[float64, string]

		h.Push(3.14, "pi")
		h.Push(2.71, "e")
		h.Push(1.41, "sqrt2")
		h.Push(1.73, "sqrt3")

		expectedScores := []float64{1.41, 1.73, 2.71, 3.14}
		for _, exp := range expectedScores {
			score, _, ok := h.Pop()
			so(ok, isTrue)
			so(score, eq, exp)
		}
	})

	cv("使用 string 作为分数", func() {
		var h minheap.Heap[string, int]

		h.Push("C", 3)
		h.Push("A", 1)
		h.Push("D", 4)
		h.Push("B", 2)

		expectedScores := []string{"A", "B", "C", "D"}
		for _, exp := range expectedScores {
			score, _, ok := h.Pop()
			so(ok, isTrue)
			so(score, eq, exp)
		}
	})

	cv("大量数据测试", func() {
		var h minheap.Heap[int, int]

		// 插入 1000 个元素
		for i := 1000; i > 0; i-- {
			h.Push(i, i*2)
		}

		so(h.Len(), eq, 1000)

		// 验证前 10 个是最小的
		for i := 1; i <= 10; i++ {
			score, value, ok := h.Pop()
			so(ok, isTrue)
			so(score, eq, i)
			so(value, eq, i*2)
		}

		so(h.Len(), eq, 990)
	})
}

func testHeapLen(t *testing.T) {
	cv("测试 Len 方法的准确性", func() {
		var h minheap.Heap[int, string]

		so(h.Len(), eq, 0)

		// 逐个添加，检查长度变化
		for i := 1; i <= 50; i++ {
			h.Push(i, "value")
			so(h.Len(), eq, i)
		}

		// 逐个删除，检查长度变化
		for i := 49; i >= 0; i-- {
			_, _, ok := h.Pop()
			so(ok, isTrue)
			so(h.Len(), eq, i)
		}
	})

	cv("测试空堆的 Len", func() {
		var h minheap.Heap[int, string]
		so(h.Len(), eq, 0)

		h.Push(1, "test")
		_, _, ok := h.Pop()
		so(ok, isTrue)
		so(h.Len(), eq, 0)
	})
}

func testHeapEmptyHeap(t *testing.T) {
	cv("从空堆 Pop 应该返回 false", func() {
		var h minheap.Heap[int, string]

		score, value, ok := h.Pop()
		so(ok, eq, false)
		so(score, eq, 0)    // 零值
		so(value, eq, "")   // 零值
		so(h.Len(), eq, 0)
	})

	cv("多次从空堆 Pop", func() {
		var h minheap.Heap[int, string]

		for i := 0; i < 5; i++ {
			score, value, ok := h.Pop()
			so(ok, eq, false)
			so(score, eq, 0)
			so(value, eq, "")
		}
		so(h.Len(), eq, 0)
	})

	cv("Pop 完所有元素后再 Pop", func() {
		var h minheap.Heap[int, string]

		h.Push(1, "one")
		h.Push(2, "two")
		h.Push(3, "three")

		// Pop 所有元素
		for i := 0; i < 3; i++ {
			_, _, ok := h.Pop()
			so(ok, isTrue)
		}

		// 再次 Pop 应该失败
		score, value, ok := h.Pop()
		so(ok, eq, false)
		so(score, eq, 0)
		so(value, eq, "")
		so(h.Len(), eq, 0)
	})

	cv("空堆 Pop 后再 Push", func() {
		var h minheap.Heap[float64, string]

		// 空堆 Pop
		score, value, ok := h.Pop()
		so(ok, eq, false)
		so(score, eq, 0.0)
		so(value, eq, "")

		// 再 Push
		h.Push(1.5, "hello")
		h.Push(0.5, "world")

		score, value, ok = h.Pop()
		so(ok, isTrue)
		so(score, eq, 0.5)
		so(value, eq, "world")

		score, value, ok = h.Pop()
		so(ok, isTrue)
		so(score, eq, 1.5)
		so(value, eq, "hello")
	})

	cv("测试复杂类型的零值", func() {
		type ComplexValue struct {
			ID   int
			Name string
		}

		var h minheap.Heap[int, ComplexValue]

		score, value, ok := h.Pop()
		so(ok, eq, false)
		so(score, eq, 0)
		so(value.ID, eq, 0)
		so(value.Name, eq, "")
	})
}
