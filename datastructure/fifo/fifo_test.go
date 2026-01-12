package fifo_test

import (
	"testing"

	"github.com/Andrew-M-C/go.util/datastructure/fifo"
	"github.com/smartystreets/goconvey/convey"
)

var (
	cv = convey.Convey
	so = convey.So
	eq = convey.ShouldEqual

	isTrue  = convey.ShouldBeTrue
	isFalse = convey.ShouldBeFalse
	notNil  = convey.ShouldNotBeNil
)

func TestQueue(t *testing.T) {
	cv("测试 FIFO 队列", t, func() {
		testNewQueue(t)
		testPushAndPop(t)
		testPopEmpty(t)
		testMultipleOperations(t)
		testQueueOrder(t)
		testSingleElement(t)
	})
}

func testNewQueue(t *testing.T) {
	cv("创建新队列", func() {
		q := fifo.New[int](0)
		so(q, notNil)
		so(q.Len(), eq, 0)

		cv("从空队列 Pop 应该返回 false", func() {
			_, ok := q.Pop()
			so(ok, isFalse)
		})
	})
}

func testPushAndPop(t *testing.T) {
	cv("Push 和 Pop 基本操作", func() {
		q := fifo.New[string](0)

		q.Push("first")
		so(q.Len(), eq, 1)

		val, ok := q.Pop()
		so(ok, isTrue)
		so(val, eq, "first")
		so(q.Len(), eq, 0)
	})
}

func testPopEmpty(t *testing.T) {
	cv("从空队列 Pop", func() {
		q := fifo.New[int](0)

		cv("第一次 Pop 返回 false", func() {
			val, ok := q.Pop()
			so(ok, isFalse)
			so(val, eq, 0)
			so(q.Len(), eq, 0)
		})

		cv("多次 Pop 空队列", func() {
			for i := 0; i < 5; i++ {
				val, ok := q.Pop()
				so(ok, isFalse)
				so(val, eq, 0)
				so(q.Len(), eq, 0)
			}
		})
	})
}

func testMultipleOperations(t *testing.T) {
	cv("多次 Push 和 Pop", func() {
		q := fifo.New[int](0)

		for i := 1; i <= 5; i++ {
			q.Push(i * 10)
		}
		so(q.Len(), eq, 5)

		expected := []int{10, 20, 30, 40, 50}
		for i, exp := range expected {
			val, ok := q.Pop()
			so(ok, isTrue)
			so(val, eq, exp)
			so(q.Len(), eq, 5-i-1)
		}

		so(q.Len(), eq, 0)
		_, ok := q.Pop()
		so(ok, isFalse)
	})
}

func testQueueOrder(t *testing.T) {
	cv("验证 FIFO 顺序", func() {
		q := fifo.New[string](0)

		items := []string{"apple", "banana", "cherry", "date", "elderberry"}

		for _, item := range items {
			q.Push(item)
		}
		so(q.Len(), eq, len(items))

		for i, expected := range items {
			val, ok := q.Pop()
			so(ok, isTrue)
			so(val, eq, expected)
			so(q.Len(), eq, len(items)-i-1)
		}
	})
}

func testSingleElement(t *testing.T) {
	cv("单元素队列的边界情况", func() {
		q := fifo.New[int](0)

		cv("Push 一个元素后 Pop", func() {
			q.Push(100)
			so(q.Len(), eq, 1)

			val, ok := q.Pop()
			so(ok, isTrue)
			so(val, eq, 100)
			so(q.Len(), eq, 0)
		})

		cv("Pop 后队列应该完全为空", func() {
			_, ok := q.Pop()
			so(ok, isFalse)
			so(q.Len(), eq, 0)
		})

		cv("可以再次 Push 和 Pop", func() {
			q.Push(200)
			so(q.Len(), eq, 1)

			val, ok := q.Pop()
			so(ok, isTrue)
			so(val, eq, 200)
			so(q.Len(), eq, 0)
		})
	})
}

func TestQueueWithStruct(t *testing.T) {
	cv("测试结构体类型", t, func() {
		type Person struct {
			Name string
			Age  int
		}

		q := fifo.New[Person](0)

		q.Push(Person{Name: "Alice", Age: 30})
		q.Push(Person{Name: "Bob", Age: 25})
		q.Push(Person{Name: "Charlie", Age: 35})
		so(q.Len(), eq, 3)

		p1, ok := q.Pop()
		so(ok, isTrue)
		so(p1.Name, eq, "Alice")
		so(p1.Age, eq, 30)

		p2, ok := q.Pop()
		so(ok, isTrue)
		so(p2.Name, eq, "Bob")
		so(p2.Age, eq, 25)

		p3, ok := q.Pop()
		so(ok, isTrue)
		so(p3.Name, eq, "Charlie")
		so(p3.Age, eq, 35)

		so(q.Len(), eq, 0)
	})
}

func TestQueueInterleaved(t *testing.T) {
	cv("交替 Push 和 Pop", t, func() {
		q := fifo.New[int](0)

		cv("交替操作", func() {
			q.Push(1)
			q.Push(2)
			so(q.Len(), eq, 2)

			val, ok := q.Pop()
			so(ok, isTrue)
			so(val, eq, 1)
			so(q.Len(), eq, 1)

			q.Push(3)
			q.Push(4)
			so(q.Len(), eq, 3)

			val, ok = q.Pop()
			so(ok, isTrue)
			so(val, eq, 2)

			val, ok = q.Pop()
			so(ok, isTrue)
			so(val, eq, 3)

			val, ok = q.Pop()
			so(ok, isTrue)
			so(val, eq, 4)

			so(q.Len(), eq, 0)
		})
	})
}

func BenchmarkPush(b *testing.B) {
	q := fifo.New[int](0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q.Push(i)
	}
}

func BenchmarkPop(b *testing.B) {
	q := fifo.New[int](0)
	for i := 0; i < b.N; i++ {
		q.Push(i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q.Pop()
	}
}

func BenchmarkPushPop(b *testing.B) {
	q := fifo.New[int](0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q.Push(i)
		q.Pop()
	}
}
