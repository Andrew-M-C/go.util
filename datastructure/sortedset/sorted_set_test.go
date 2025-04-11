package sortedset_test

import (
	"testing"

	ss "github.com/Andrew-M-C/go.util/datastructure/sortedset"
	"github.com/smartystreets/goconvey/convey"
)

// go test -v -failfast -cover -coverprofile cover.out && go tool cover -html cover.out -o cover.html

var (
	cv = convey.Convey
	so = convey.So
	eq = convey.ShouldEqual
)

func TestSortedSet(t *testing.T) {
	cv("测试有序集合", t, func() { testSet(t) })
}

func testSet(t *testing.T) {
	cv("基本操作", func() { testSet_Basic(t) })
	cv("获取最高/最低分数", func() { testSet_GetHighestLowest(t) })
	cv("相同分数的元素", func() { testSet_SameScore(t) })
	cv("删除操作", func() { testSet_Del(t) })
	cv("空集合操作", func() { testSet_Empty(t) })
	cv("更新已存在元素", func() { testSet_Update(t) })
	cv("链表操作", func() { testSet_LinkedList(t) })
}

func testSet_Basic(t *testing.T) {
	set := ss.NewSortedSet[string, int, float64]()
	set.SetLogger(t.Logf)

	// 测试初始状态
	so(set.Len(), eq, 0)

	// 添加元素
	set.Set("a", 1, 1.0)
	so(set.Len(), eq, 1)

	val, exist := set.Get("a")
	so(exist, eq, true)
	so(val, eq, 1)

	// 获取不存在的元素
	val, exist = set.Get("b")
	so(exist, eq, false)
	so(val, eq, 0) // 默认零值

	// 添加更多元素
	set.Set("b", 2, 2.0)
	set.Set("c", 3, 3.0)
	so(set.Len(), eq, 3)
}

func testSet_GetHighestLowest(t *testing.T) {
	set := ss.NewSortedSet[string, string, int]()
	set.SetLogger(t.Logf)

	// 添加元素
	set.Set("medium", "中", 5)
	set.Set("low", "低", 1)
	set.Set("high", "高", 10)

	// 测试最低分数
	score, values := set.GetLowest()
	so(score, eq, 1)
	so(len(values), eq, 1)
	so(values["low"], eq, "低")

	// 测试最高分数
	score, values = set.GetHighest()
	so(score, eq, 10)
	so(len(values), eq, 1)
	so(values["high"], eq, "高")
}

func testSet_SameScore(t *testing.T) {
	set := ss.NewSortedSet[string, int, int]()
	set.SetLogger(t.Logf)

	// 添加分数相同的多个元素
	set.Set("a", 1, 5)
	set.Set("b", 2, 5)
	set.Set("c", 3, 5)

	// 测试分数相同的元素
	score, values := set.GetHighest() // 或 GetLowest，因为只有一个分数
	so(score, eq, 5)
	so(len(values), eq, 3)
	so(values["a"], eq, 1)
	so(values["b"], eq, 2)
	so(values["c"], eq, 3)
}

func testSet_Del(t *testing.T) {
	set := ss.NewSortedSet[string, int, float64]()
	set.SetLogger(t.Logf)

	// 添加元素
	set.Set("a", 1, 1.0)
	set.Set("b", 2, 2.0)
	set.Set("c", 3, 2.0) // 同样分数
	set.Set("d", 4, 3.0)

	// 删除不存在的元素
	set.Del("x")
	so(set.Len(), eq, 4)

	// 删除存在的元素
	set.Del("a")
	so(set.Len(), eq, 3)
	_, exist := set.Get("a")
	so(exist, eq, false)

	// 删除分数相同元素列表中的一个元素
	set.Del("c")
	so(set.Len(), eq, 2)
	score, values := set.GetLowest()
	so(score, eq, 2.0)
	so(len(values), eq, 1) // 修复：实际上删除 c 后，分数 2.0 只剩下 b 一个元素
	so(values["b"], eq, 2)

	// 删除分数相同元素列表中的最后一个元素
	set.Del("b")
	so(set.Len(), eq, 1)

	// 删除最后一个元素
	set.Del("d")
	so(set.Len(), eq, 0)
}

func testSet_Empty(t *testing.T) {
	set := ss.NewSortedSet[int, int, int]()
	set.SetLogger(t.Logf)

	// 测试空集合获取最高/最低分数
	score, values := set.GetLowest()
	so(score, eq, 0) // 默认零值
	so(len(values), eq, 0)

	score, values = set.GetHighest()
	so(score, eq, 0) // 默认零值
	so(len(values), eq, 0)

	// 测试删除不存在的元素
	set.Del(123)
	so(set.Len(), eq, 0)

	// 测试获取不存在的元素
	_, exist := set.Get(456)
	so(exist, eq, false)
}

func testSet_Update(t *testing.T) {
	set := ss.NewSortedSet[string, string, int]()
	set.SetLogger(t.Logf)

	// 添加元素
	set.Set("key", "原始值", 10)

	// 更新值
	set.Set("key", "新值", 10)
	val, _ := set.Get("key")
	so(val, eq, "新值")

	// 更新分数
	set.Set("key", "新值", 20)
	score, values := set.GetHighest()
	so(score, eq, 20)
	so(values["key"], eq, "新值")

	// 旧分数应该不在集合中了
	lowest, _ := set.GetLowest()
	so(lowest, eq, 20)
}

// 增加一个测试链表操作的函数
func testSet_LinkedList(t *testing.T) {
	set := ss.NewSortedSet[string, int, int]()
	set.SetLogger(t.Logf)

	// 测试相同分数下的链表删除
	set.Set("a", 1, 5)
	set.Set("b", 2, 5)
	set.Set("c", 3, 5)

	// 删除链表头
	set.Del("c")
	score, values := set.GetHighest()
	so(score, eq, 5)
	so(len(values), eq, 2)
	so(values["a"], eq, 1)
	so(values["b"], eq, 2)

	// 删除链表中间元素
	set.Set("c", 3, 5) // 重新添加 c
	set.Del("b")
	score, values = set.GetHighest()
	so(score, eq, 5)
	so(len(values), eq, 2)
	so(values["a"], eq, 1)
	so(values["c"], eq, 3)

	// 删除链表尾部元素
	set.Set("b", 2, 5) // 重新添加 b
	set.Del("a")
	score, values = set.GetHighest()
	so(score, eq, 5)
	so(len(values), eq, 2)
	so(values["b"], eq, 2)
	so(values["c"], eq, 3)
}
