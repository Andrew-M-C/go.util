package heap_test

import (
	"testing"

	"github.com/Andrew-M-C/go.util/container/heap"
	"github.com/smartystreets/goconvey/convey"
)

var (
	cv = convey.Convey
	so = convey.So
	eq = convey.ShouldEqual

	isTrue = convey.ShouldBeTrue
	notNil = convey.ShouldNotBeNil
)

func TestHeap(t *testing.T) {
	cv("测试基本的 Push 和 Pop 功能", t, func() { testBasicPushPop(t) })
	cv("测试最小堆", t, func() { testMinHeap(t) })
	cv("测试最大堆", t, func() { testMaxHeap(t) })
	cv("测试 Len 方法", t, func() { testLen(t) })
	cv("测试 panic 情况", t, func() { testPanic(t) })
	cv("测试自定义类型", t, func() { testCustomType(t) })
}

func testBasicPushPop(t *testing.T) {
	cv("基本的 Push 和 Pop 操作", func() {
		h := heap.New(func(i, j int) bool {
			return i < j
		})

		so(h, notNil)
		so(h.Len(), eq, 0)

		h.Push(5)
		so(h.Len(), eq, 1)

		h.Push(3)
		h.Push(7)
		h.Push(1)
		so(h.Len(), eq, 4)

		v := h.Pop()
		so(v, eq, 1) // 最小的元素
		so(h.Len(), eq, 3)

		v = h.Pop()
		so(v, eq, 3)
		so(h.Len(), eq, 2)

		v = h.Pop()
		so(v, eq, 5)
		so(h.Len(), eq, 1)

		v = h.Pop()
		so(v, eq, 7)
		so(h.Len(), eq, 0)
	})
}

func testMinHeap(t *testing.T) {
	cv("最小堆 - 按升序弹出", func() {
		h := heap.New(func(i, j int) bool {
			return i < j
		})

		// 无序插入
		values := []int{50, 10, 30, 20, 40, 5, 15, 25}
		for _, v := range values {
			h.Push(v)
		}

		so(h.Len(), eq, len(values))

		// 应该按升序弹出
		expected := []int{5, 10, 15, 20, 25, 30, 40, 50}
		for _, exp := range expected {
			v := h.Pop()
			so(v, eq, exp)
		}

		so(h.Len(), eq, 0)
	})

	cv("最小堆 - 处理负数", func() {
		h := heap.New(func(i, j int) bool {
			return i < j
		})

		values := []int{-5, 10, -20, 0, 15, -10}
		for _, v := range values {
			h.Push(v)
		}

		// 应该按升序弹出
		expected := []int{-20, -10, -5, 0, 10, 15}
		for _, exp := range expected {
			v := h.Pop()
			so(v, eq, exp)
		}
	})
}

func testMaxHeap(t *testing.T) {
	cv("最大堆 - 按降序弹出", func() {
		h := heap.New(func(i, j int) bool {
			return i > j // 注意这里是大于号
		})

		// 无序插入
		values := []int{50, 10, 30, 20, 40, 5, 15, 25}
		for _, v := range values {
			h.Push(v)
		}

		so(h.Len(), eq, len(values))

		// 应该按降序弹出
		expected := []int{50, 40, 30, 25, 20, 15, 10, 5}
		for _, exp := range expected {
			v := h.Pop()
			so(v, eq, exp)
		}

		so(h.Len(), eq, 0)
	})
}

func testLen(t *testing.T) {
	cv("测试 Len 方法的准确性", func() {
		h := heap.New(func(i, j int) bool {
			return i < j
		})

		so(h.Len(), eq, 0)

		// 逐个添加，检查长度变化
		for i := 1; i <= 100; i++ {
			h.Push(i)
			so(h.Len(), eq, i)
		}

		// 逐个删除，检查长度变化
		for i := 99; i >= 0; i-- {
			h.Pop()
			so(h.Len(), eq, i)
		}
	})
}

func testPanic(t *testing.T) {
	cv("lessFunc 为 nil 应该 panic", func() {
		defer func() {
			r := recover()
			so(r, notNil)
			so(r, eq, "lessFunc is nil")
		}()

		heap.New[int](nil)
		so(false, isTrue) // 不应该执行到这里
	})
}

func testCustomType(t *testing.T) {
	type Person struct {
		Name string
		Age  int
	}

	cv("自定义类型 - 按年龄排序", func() {
		h := heap.New(func(i, j Person) bool {
			return i.Age < j.Age
		})

		people := []Person{
			{Name: "张三", Age: 30},
			{Name: "李四", Age: 25},
			{Name: "王五", Age: 35},
			{Name: "赵六", Age: 20},
			{Name: "钱七", Age: 28},
		}

		for _, p := range people {
			h.Push(p)
		}

		so(h.Len(), eq, 5)

		// 应该按年龄升序弹出
		expectedAges := []int{20, 25, 28, 30, 35}
		for i, expAge := range expectedAges {
			p := h.Pop()
			so(p.Age, eq, expAge)
			t.Logf("第 %d 个弹出: %s, 年龄 %d", i+1, p.Name, p.Age)
		}

		so(h.Len(), eq, 0)
	})

	cv("自定义类型 - 按字符串排序", func() {
		type Task struct {
			Priority string
			Content  string
		}

		h := heap.New(func(i, j Task) bool {
			// 字母序越小，优先级越高
			return i.Priority < j.Priority
		})

		tasks := []Task{
			{Priority: "C", Content: "低优先级任务"},
			{Priority: "A", Content: "高优先级任务"},
			{Priority: "B", Content: "中优先级任务"},
			{Priority: "D", Content: "最低优先级任务"},
		}

		for _, task := range tasks {
			h.Push(task)
		}

		expectedPriorities := []string{"A", "B", "C", "D"}
		for _, exp := range expectedPriorities {
			task := h.Pop()
			so(task.Priority, eq, exp)
		}
	})

	cv("自定义类型 - 复杂排序逻辑", func() {
		type Job struct {
			Urgent   bool
			Priority int
			ID       int
		}

		h := heap.New(func(i, j Job) bool {
			// 1. 紧急的任务优先
			if i.Urgent != j.Urgent {
				return i.Urgent
			}
			// 2. 优先级数字越小越优先
			if i.Priority != j.Priority {
				return i.Priority < j.Priority
			}
			// 3. ID 越小越优先
			return i.ID < j.ID
		})

		jobs := []Job{
			{Urgent: false, Priority: 2, ID: 3},
			{Urgent: true, Priority: 1, ID: 2},
			{Urgent: true, Priority: 1, ID: 1},
			{Urgent: false, Priority: 1, ID: 4},
			{Urgent: true, Priority: 2, ID: 5},
		}

		for _, job := range jobs {
			h.Push(job)
		}

		// 预期顺序:
		// 1. Urgent=true, Priority=1, ID=1
		// 2. Urgent=true, Priority=1, ID=2
		// 3. Urgent=true, Priority=2, ID=5
		// 4. Urgent=false, Priority=1, ID=4
		// 5. Urgent=false, Priority=2, ID=3
		expected := []Job{
			{Urgent: true, Priority: 1, ID: 1},
			{Urgent: true, Priority: 1, ID: 2},
			{Urgent: true, Priority: 2, ID: 5},
			{Urgent: false, Priority: 1, ID: 4},
			{Urgent: false, Priority: 2, ID: 3},
		}

		for i, exp := range expected {
			job := h.Pop()
			so(job.Urgent, eq, exp.Urgent)
			so(job.Priority, eq, exp.Priority)
			so(job.ID, eq, exp.ID)
			t.Logf("第 %d 个: Urgent=%v, Priority=%d, ID=%d",
				i+1, job.Urgent, job.Priority, job.ID)
		}
	})
}
