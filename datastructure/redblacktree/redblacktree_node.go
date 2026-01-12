package redblacktree

import (
	"github.com/Andrew-M-C/go.util/datastructure/constraints"
)

// Node 表示一个节点
type Node[K constraints.Ordered, V any] struct {
	K K
	V V
}
