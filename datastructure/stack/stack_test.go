package stack_test

import (
	"testing"

	"github.com/Andrew-M-C/go.util/datastructure/stack"
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

func TestStack(t *testing.T) {
	cv("测试 FILO 栈", t, func() {
		testNewStack(t)
		testPushAndPop(t)
		testPopEmpty(t)
		testMultipleOperations(t)
		testStackOrder(t)
		testSingleElement(t)
	})
}

func testNewStack(t *testing.T) {
	cv("创建新栈", func() {
		s := stack.New[int]()
		so(s, notNil)
		so(s.Len(), eq, 0)

		cv("从空栈 Pop 应该返回 false", func() {
			_, ok := s.Pop()
			so(ok, isFalse)
		})
	})
}

func testPushAndPop(t *testing.T) {
	cv("Push 和 Pop 基本操作", func() {
		s := stack.New[string]()

		s.Push("first")
		so(s.Len(), eq, 1)

		val, ok := s.Pop()
		so(ok, isTrue)
		so(val, eq, "first")
		so(s.Len(), eq, 0)
	})
}

func testPopEmpty(t *testing.T) {
	cv("从空栈 Pop", func() {
		s := stack.New[int]()

		cv("第一次 Pop 返回 false", func() {
			val, ok := s.Pop()
			so(ok, isFalse)
			so(val, eq, 0)
			so(s.Len(), eq, 0)
		})

		cv("多次 Pop 空栈", func() {
			for i := 0; i < 5; i++ {
				val, ok := s.Pop()
				so(ok, isFalse)
				so(val, eq, 0)
				so(s.Len(), eq, 0)
			}
		})
	})
}

func testMultipleOperations(t *testing.T) {
	cv("多次 Push 和 Pop", func() {
		s := stack.New[int]()

		for i := 1; i <= 5; i++ {
			s.Push(i * 10)
		}
		so(s.Len(), eq, 5)

		// 栈是 FILO，所以应该按照相反的顺序弹出
		expected := []int{50, 40, 30, 20, 10}
		for i, exp := range expected {
			val, ok := s.Pop()
			so(ok, isTrue)
			so(val, eq, exp)
			so(s.Len(), eq, 5-i-1)
		}

		so(s.Len(), eq, 0)
		_, ok := s.Pop()
		so(ok, isFalse)
	})
}

func testStackOrder(t *testing.T) {
	cv("验证 FILO 顺序", func() {
		s := stack.New[string]()

		items := []string{"apple", "banana", "cherry", "date", "elderberry"}

		for _, item := range items {
			s.Push(item)
		}
		so(s.Len(), eq, len(items))

		// 栈是后进先出，所以应该反向弹出
		for i := len(items) - 1; i >= 0; i-- {
			val, ok := s.Pop()
			so(ok, isTrue)
			so(val, eq, items[i])
			so(s.Len(), eq, i)
		}
	})
}

func testSingleElement(t *testing.T) {
	cv("单元素栈的边界情况", func() {
		s := stack.New[int]()

		cv("Push 一个元素后 Pop", func() {
			s.Push(100)
			so(s.Len(), eq, 1)

			val, ok := s.Pop()
			so(ok, isTrue)
			so(val, eq, 100)
			so(s.Len(), eq, 0)
		})

		cv("Pop 后栈应该完全为空", func() {
			_, ok := s.Pop()
			so(ok, isFalse)
			so(s.Len(), eq, 0)
		})

		cv("可以再次 Push 和 Pop", func() {
			s.Push(200)
			so(s.Len(), eq, 1)

			val, ok := s.Pop()
			so(ok, isTrue)
			so(val, eq, 200)
			so(s.Len(), eq, 0)
		})
	})
}

func TestStackWithStruct(t *testing.T) {
	cv("测试结构体类型", t, func() {
		type Person struct {
			Name string
			Age  int
		}

		s := stack.New[Person]()

		s.Push(Person{Name: "Alice", Age: 30})
		s.Push(Person{Name: "Bob", Age: 25})
		s.Push(Person{Name: "Charlie", Age: 35})
		so(s.Len(), eq, 3)

		// 栈是 FILO，所以 Charlie 先出
		p1, ok := s.Pop()
		so(ok, isTrue)
		so(p1.Name, eq, "Charlie")
		so(p1.Age, eq, 35)

		p2, ok := s.Pop()
		so(ok, isTrue)
		so(p2.Name, eq, "Bob")
		so(p2.Age, eq, 25)

		p3, ok := s.Pop()
		so(ok, isTrue)
		so(p3.Name, eq, "Alice")
		so(p3.Age, eq, 30)

		so(s.Len(), eq, 0)
	})
}

func TestStackInterleaved(t *testing.T) {
	cv("交替 Push 和 Pop", t, func() {
		s := stack.New[int]()

		cv("交替操作", func() {
			s.Push(1)
			s.Push(2)
			so(s.Len(), eq, 2)

			val, ok := s.Pop()
			so(ok, isTrue)
			so(val, eq, 2) // 栈顶是 2
			so(s.Len(), eq, 1)

			s.Push(3)
			s.Push(4)
			so(s.Len(), eq, 3)

			val, ok = s.Pop()
			so(ok, isTrue)
			so(val, eq, 4) // 最后 Push 的是 4

			val, ok = s.Pop()
			so(ok, isTrue)
			so(val, eq, 3)

			val, ok = s.Pop()
			so(ok, isTrue)
			so(val, eq, 1) // 最早 Push 的 1 最后出来

			so(s.Len(), eq, 0)
		})
	})
}

func BenchmarkPush(b *testing.B) {
	s := stack.New[int]()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Push(i)
	}
}

func BenchmarkPop(b *testing.B) {
	s := stack.New[int]()
	for i := 0; i < b.N; i++ {
		s.Push(i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Pop()
	}
}

func BenchmarkPushPop(b *testing.B) {
	s := stack.New[int]()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Push(i)
		s.Pop()
	}
}
