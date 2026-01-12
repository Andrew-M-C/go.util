// Package sortedset 提供一个有序集合实现
package sortedset

import (
	"github.com/Andrew-M-C/go.util/datastructure/constraints"
	rbt "github.com/Andrew-M-C/go.util/datastructure/redblacktree"
)

// Set 定义一个有序集合。注意, 该集合协程不安全!!!
type Set[K constraints.Ordered, V any, S constraints.Ordered] interface {
	// Set 设置一个元素
	Set(key K, value V, score S)
	// Get 获取一个元素
	Get(key K) (V, bool)
	// Del 删除一个元素
	Del(key K)
	// Len 返回集合中的元素数量
	Len() int
	// GetLowest 获取分数最低的元素。如果集合为空, 那么 score 和 key 都会返回默认零值, 并且 values 返回空
	GetLowest() (score S, values map[K]V)
	// GetHighest 获取分数最高的元素。如果集合为空, 那么 score 和 key 都会返回默认零值, 并且 values 返回空
	GetHighest() (score S, values map[K]V)

	// 调试用
	SetLogger(func(string, ...any))
}

func NewSortedSet[K constraints.Ordered, V any, S constraints.Ordered]() Set[K, V, S] {
	return &sortedSetImpl[K, V, S]{
		byScore: rbt.New[S, *kvItem[K, V, S]](),
		byKey:   map[K]*kvItem[K, V, S]{},
		log:     func(s string, a ...any) {},
	}
}

type kvItem[K constraints.Ordered, V any, S constraints.Ordered] struct {
	next *kvItem[K, V, S]
	prev *kvItem[K, V, S]

	key   K
	value V
	score S
}

type sortedSetImpl[K constraints.Ordered, V any, S constraints.Ordered] struct {
	byScore *rbt.Tree[S, *kvItem[K, V, S]]
	byKey   map[K]*kvItem[K, V, S]

	log func(string, ...any)
}

func (set *sortedSetImpl[K, V, S]) SetLogger(f func(string, ...any)) {
	if f != nil {
		set.log = f
	}
}

func (set *sortedSetImpl[K, V, S]) Set(key K, value V, score S) {
	item := &kvItem[K, V, S]{
		key:   key,
		value: value,
		score: score,
	}

	// 首先判断下 map 之前是否存在相同的 key
	prev, exist := set.byKey[key]
	if exist {
		// 之前的元素要删除掉
		set.delItem(prev)
	}

	// 然后添加新的元素
	// 首先添加到 byKey 中
	set.byKey[key] = item

	// 然后添加到 byScore 中
	chain, _ := set.byScore.Get(score)
	if chain == nil {
		set.log("添加 score 为 %v 的唯一元素 %v", score, key)
		set.byScore.Set(score, item)
		return
	}

	chain.prev = item
	item.next = chain
	set.log("添加 score 为 %v 的元素 %v", score, key)
	set.byScore.Set(score, item)
}

func (set *sortedSetImpl[K, V, S]) Get(key K) (value V, exist bool) {
	v, exist := set.byKey[key]
	if !exist {
		return value, false
	}
	return v.value, true
}

func (set *sortedSetImpl[K, V, S]) delItem(item *kvItem[K, V, S]) {
	// 首先从 byKey 中删掉
	delete(set.byKey, item.key)

	// 然后从 byScore 中删掉
	// 首先要取出元素链
	chain, _ := set.byScore.Get(item.score)
	if chain == nil {
		return
	}

	// 修复：改进链表遍历逻辑，确保能处理所有情况
	for curr := chain; curr != nil; curr = curr.next {
		// 没有找到, 继续遍历
		if curr.key != item.key {
			continue
		}

		// 如果当前元素是链表的第一个元素
		if curr.prev == nil {
			if curr.next == nil {
				set.log("删除 score 为 %v 的唯一元素 %v", item.score, item.key)
				set.byScore.Remove(item.score)
				return
			}
			set.log("删除 score 为 %v 的最新元素 %v", item.score, item.key)
			curr.next.prev = nil
			set.byScore.Set(item.score, curr.next)
			return
		}

		// 如果当前元素是链表的中间元素
		if curr.next != nil {
			set.log("删除 score 为 %v 的中间元素 %v", item.score, item.key)
			curr.prev.next = curr.next
			curr.next.prev = curr.prev
			return
		}

		// 当前元素是链表的最后一个元素
		set.log("删除 score 为 %v 的最老元素 %v", item.score, item.key)
		curr.prev.next = nil
		return
	}
}

func (set *sortedSetImpl[K, V, S]) Del(key K) {
	item, exist := set.byKey[key]
	if !exist {
		set.log("元素 %v 不存在, 不予删除", key)
		return
	}
	set.delItem(item)
}

func (set *sortedSetImpl[K, V, S]) Len() int {
	return len(set.byKey)
}

func (set *sortedSetImpl[K, V, S]) GetLowest() (score S, values map[K]V) {
	node := set.byScore.Left()
	if node == nil {
		return score, map[K]V{}
	}

	return set.returnByNode(node)
}

func (set *sortedSetImpl[K, V, S]) GetHighest() (score S, values map[K]V) {
	node := set.byScore.Right()
	if node == nil {
		return score, map[K]V{}
	}
	return set.returnByNode(node)
}

func (*sortedSetImpl[K, V, S]) returnByNode(node *rbt.Node[S, *kvItem[K, V, S]]) (score S, values map[K]V) {
	values = make(map[K]V)
	score = node.K
	for curr := node.V; curr != nil; curr = curr.next {
		values[curr.key] = curr.value
	}
	return score, values
}
