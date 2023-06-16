// Package redblacktree 提供红黑树实现
package redblacktree

import (
	rbt "github.com/emirpasic/gods/trees/redblacktree"
	"golang.org/x/exp/constraints"
)

type treeCapability[K constraints.Ordered, V any] interface {
	Size() int
	Clear()

	Keys() []K
	Values() []V

	Set(k K, v V)
	Put(k K, v V)
	Get(k K) (V, bool)
	Floor(k K) *Node[K, V]
	Ceiling(k K) *Node[K, V]

	Left() *Node[K, V]
	Right() *Node[K, V]

	Iterate(func(k K, v V) bool)
}

var _ treeCapability[int, struct{}] = (*Tree[int, struct{}])(nil)

// Tree 表示一个红黑树实现
type Tree[K constraints.Ordered, V any] struct {
	tree *rbt.Tree
}

// New 新建一个红黑树实现
func New[K constraints.Ordered, V any]() *Tree[K, V] {
	return &Tree[K, V]{}
}

func (t *Tree[K, V]) lazyInit() {
	if t.tree != nil {
		return
	}
	t.tree = rbt.NewWith(t.compare)
}

func (t *Tree[K, V]) compare(i, j any) int {
	ii, _ := i.(K)
	jj, _ := j.(K)
	switch {
	case ii < jj:
		return -1
	default:
		return 0
	case ii > jj:
		return 1
	}
}

// Size 返回 KV 对的数量
func (t *Tree[K, V]) Size() int {
	if t.tree == nil {
		return 0
	}
	return t.tree.Size()
}

// Clear 清空红黑树
func (t *Tree[K, V]) Clear() {
	if t.tree != nil {
		t.tree.Clear()
	}
}

// Values 从低到高返回所有值。如果没有值, 则返回 nil
func (t *Tree[K, V]) Values() []V {
	if t.Size() == 0 {
		return nil
	}
	vals := t.tree.Values()
	res := make([]V, 0, len(vals))
	for _, intf := range vals {
		v, _ := intf.(V)
		res = append(res, v)
	}
	return res
}

// Set 设置一个值
func (t *Tree[K, V]) Set(k K, v V) {
	t.lazyInit()
	t.tree.Put(k, v)
}

// Put 设置一个值
func (t *Tree[K, V]) Put(k K, v V) {
	t.Set(k, v)
}

// Get 获取某值
func (t *Tree[K, V]) Get(k K) (v V, exist bool) {
	if t.tree == nil {
		return
	}
	val, exist := t.tree.Get(k)
	if !exist {
		return
	}
	v, _ = val.(V)
	return v, exist
}

// Floor 获取某个值左边的第一个值, 如果集合为空或者当前值是最小值, 则返回 nil
func (t *Tree[K, V]) Floor(k K) *Node[K, V] {
	if t.tree == nil {
		return nil
	}
	node, exist := t.tree.Floor(k)
	if !exist || node == nil {
		return nil
	}

	key, _ := node.Key.(K)
	val, _ := node.Value.(V)
	return &Node[K, V]{
		K: key,
		V: val,
	}
}

// Ceiling 获取某个值右边的第一个值, 如果集合为空或者当前值是最大值, 则返回 nil
func (t *Tree[K, V]) Ceiling(k K) *Node[K, V] {
	if t.tree == nil {
		return nil
	}
	node, exist := t.tree.Ceiling(k)
	if !exist || node == nil {
		return nil
	}

	key, _ := node.Key.(K)
	val, _ := node.Value.(V)
	return &Node[K, V]{
		K: key,
		V: val,
	}
}

// Keys 返回所有的 key, 列表为空则返回 nil
func (t *Tree[K, V]) Keys() (keys []K) {
	if t.Size() == 0 {
		return nil
	}
	intfKeys := t.tree.Keys()
	keys = make([]K, 0, len(intfKeys))
	for _, v := range intfKeys {
		key, _ := v.(K)
		keys = append(keys, key)
	}
	return keys
}

// Left 返回整个集合的左值, 如果没有则返回 nil
func (t *Tree[K, V]) Left() *Node[K, V] {
	if t.Size() == 0 {
		return nil
	}
	intfNode := t.tree.Left()
	if intfNode == nil {
		return nil
	}
	k, _ := intfNode.Key.(K)
	v, _ := intfNode.Key.(V)
	return &Node[K, V]{
		K: k,
		V: v,
	}
}

// Right 返回整个集合的左值, 如果没有则返回 nil
func (t *Tree[K, V]) Right() *Node[K, V] {
	if t.Size() == 0 {
		return nil
	}
	intfNode := t.tree.Right()
	if intfNode == nil {
		return nil
	}
	k, _ := intfNode.Key.(K)
	v, _ := intfNode.Key.(V)
	return &Node[K, V]{
		K: k,
		V: v,
	}
}

// ITerate 遍历每一个值
func (t *Tree[K, V]) Iterate(f func(k K, v V) bool) {
	if f == nil || t.Size() == 0 {
		return
	}
	for it := t.tree.Iterator(); it.Next(); {
		intfK, intfV := it.Key(), it.Value()
		k, _ := intfK.(K)
		v, _ := intfV.(V)
		goOn := f(k, v)
		if !goOn {
			break
		}
	}
}
