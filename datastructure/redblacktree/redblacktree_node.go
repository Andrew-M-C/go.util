package redblacktree

import (
	"golang.org/x/exp/constraints"
)

// Node 表示一个节点
type Node[K constraints.Ordered, V any] struct {
	K K
	V V
}
