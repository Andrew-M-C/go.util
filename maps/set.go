package maps

import (
	"github.com/Andrew-M-C/go.util/constraints"
	"github.com/Andrew-M-C/go.util/slice"
)

// Set 表示一个集合
type Set[K comparable] map[K]struct{}

// NewSet 返回一个集合类型
func NewSet[K comparable](vals ...K) Set[K] {
	res := make(Set[K], len(vals))
	for _, v := range vals {
		res.Add(v)
	}
	return res
}

// NewSetFromSlice 从一个切片转为 Set 类型
func NewSetFromSlice[T constraints.Ordered](sli slice.List[T]) Set[T] {
	s := make(Set[T], len(sli))
	for _, v := range sli {
		s.Add(v)
	}
	return s
}

// NewSetWithCapacity 返回一个集合类型并初始化容量
func NewSetWithCapacity[K comparable, I constraints.Integer](cap I) Set[K] {
	return make(Set[K], cap)
}

// Add 添加一个值
func (s Set[K]) Add(key K) {
	s[key] = struct{}{}
}

// Has 是否包含某个 key
func (s Set[K]) Has(key K) bool {
	_, b := s[key]
	return b
}

// Del 删除某个 key, 并且返回删除之前是否已存在
func (s Set[K]) Del(key K) bool {
	if _, b := s[key]; !b {
		return false
	}
	delete(s, key)
	return true
}

// Equal 判断两个 set 是不是相等
func (s Set[K]) Equal(another Set[K]) bool {
	if len(s) != len(another) {
		return false
	}
	for k, v := range s {
		anotherV, exist := another[k]
		if !exist {
			return false
		}
		if v != anotherV {
			return false
		}
	}
	return true
}

// Clone 浅复制
func (s Set[K]) Clone() Set[K] {
	res := make(Set[K], len(s))
	for k, v := range s {
		res[k] = v
	}
	return res
}

// Len 返回集合的大小
func (s Set[K]) Len() int {
	return len(s)
}

// Difference 差集
func (s Set[K]) Difference(toSub Set[K]) Set[K] {
	res := make(Set[K], len(s))
	for k, v := range s {
		if !toSub.Has(k) {
			res[k] = v
		}
	}
	return res
}

// Union 并集
func (s Set[K]) Union(another Set[K]) Set[K] {
	res := s.Clone()
	for k, v := range another {
		res[k] = v
	}
	return res
}

// Intersection 交集
func (s Set[K]) Intersection(another Set[K]) Set[K] {
	res := make(Set[K], len(s))
	for k, v := range s {
		if another.Has(k) {
			res[k] = v
		}
	}
	return res
}

// SymmetricDifference 对称差集, 表示两个集合中不相交的部份, 即 并集 减去 交集
func (s Set[K]) SymmetricDifference(another Set[K]) Set[K] {
	res := s.Clone()
	for k := range another {
		if s.Has(k) {
			res.Del(k)
		} else {
			res.Add(k)
		}
	}
	return res
}
