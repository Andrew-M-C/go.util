// Package bits 提供位操作相关的逻辑
package bits

import "golang.org/x/exp/constraints"

// New64 初始化一个 uint64 的位掩码
func New64(offsetsToSet ...int) uint64 {
	return Set(uint64(0), offsetsToSet...)
}

// Set 按偏移量设置位
func Set[T constraints.Unsigned](orig T, offsets ...int) T {
	for _, offset := range offsets {
		orig |= (1 << offset)
	}
	return orig
}

// Clear 按偏移量清除位
func Clear[T constraints.Unsigned](orig T, offsets ...int) T {
	for _, offset := range offsets {
		orig &= ^(1 << offset)
	}
	return orig
}

// HasAny 返回指定的位是否有任何一个置位了
func HasAny[T constraints.Unsigned](b T, offsets ...int) bool {
	var mask T
	for _, offset := range offsets {
		mask |= (1 << offset)
	}
	return (b & mask) > 0
}

// HasAll 返回指定的位是否全部一个置位了
func HasAll[T constraints.Unsigned](b T, offsets ...int) bool {
	var mask T
	for _, offset := range offsets {
		mask |= (1 << offset)
	}
	return (b & mask) == mask
}
