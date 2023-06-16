package redblacktree_test

import (
	"fmt"
	"testing"

	rbt "github.com/Andrew-M-C/go.util/datastructure/redblacktree"
	"github.com/smartystreets/goconvey/convey"
)

// go test -v -failfast -cover -coverprofile cover.out && go tool cover -html cover.out -o cover.html

var (
	cv = convey.Convey
	so = convey.So
	eq = convey.ShouldEqual
)

func TestRedBlackTree(t *testing.T) {
	cv("测试红黑树", t, func() { testTree(t) })
}

func testTree(t *testing.T) {
	cv("Size()", func() { testTree_Size(t) })
	cv("Clear()", func() { testTree_Clear(t) })
	cv("Keys(), Values()", func() { testTree_Keys_Values(t) })
	cv("Set(), Put(), Get()", func() { testTree_Set_Put_Get(t) })
	cv("Floor(), Ceiling()", func() { testTree_Floor_Ceiling(t) })
	cv("Left(), Right()", func() { testTree_Left_Right(t) })
	cv("Iterate()", func() { testTree_Iterate(t) })
}

func testTree_Size(t *testing.T) {
	tree := &rbt.Tree[int, int]{}
	so(tree.Size(), eq, 0)

	tree = rbt.New[int, int]()
	so(tree.Size(), eq, 0)

	tree.Set(1, 2)
	so(tree.Size(), eq, 1)

	tree.Set(2, 3)
	so(tree.Size(), eq, 2)

	tree.Set(2, 2)
	so(tree.Size(), eq, 2)
}

func testTree_Clear(t *testing.T) {
	tree := &rbt.Tree[int, struct{}]{}
	tree.Clear() // not panic

	tree = rbt.New[int, struct{}]()
	so(tree.Size(), eq, 0)

	tree.Put(1, struct{}{})
	so(tree.Size(), eq, 1)

	tree.Clear()
	so(tree.Size(), eq, 0)
}

func testTree_Keys_Values(t *testing.T) {
	tree := &rbt.Tree[int, string]{}
	keys := tree.Keys()
	vals := tree.Values()
	so(len(keys), eq, 0)
	so(len(vals), eq, 0)

	tree.Set(1, "1")
	tree.Put(2, "22")
	tree.Set(3, "33")

	vals = tree.Values()
	so(len(vals), eq, 3)
	so(vals[0], eq, "1")
	so(vals[1], eq, "22")
	so(vals[2], eq, "33")

	keys = tree.Keys()
	so(len(keys), eq, 3)
	so(keys[0], eq, 1)
	so(keys[1], eq, 2)
	so(keys[2], eq, 3)
}

func testTree_Set_Put_Get(t *testing.T) {
	tree := rbt.New[float64, string]()
	v, exist := tree.Get(0)
	so(exist, eq, false)
	so(v, eq, "")

	tree.Put(0, "0")
	tree.Set(1, "1")

	v, exist = tree.Get(0)
	so(exist, eq, true)
	so(v, eq, "0")

	v, exist = tree.Get(0.5)
	so(exist, eq, false)
	so(v, eq, "")
}

func testTree_Floor_Ceiling(t *testing.T) {
	tree := rbt.New[string, int]()
	n := tree.Floor("0")
	so(n, eq, nil)
	n = tree.Ceiling("9")
	so(n, eq, nil)

	tree.Set("1", 1)
	tree.Set("8", 8)

	n = tree.Floor("0")
	so(n, eq, nil)
	n = tree.Ceiling("9")
	so(n, eq, nil)

	n = tree.Floor("1")
	so(n.K, eq, "1")
	so(n.V, eq, 1)

	n = tree.Ceiling("8")
	so(n.K, eq, "8")
	so(n.V, eq, 8)

	n = tree.Floor("5")
	so(n.K, eq, "1")
	so(n.V, eq, 1)

	n = tree.Ceiling("5")
	so(n.K, eq, "8")
	so(n.V, eq, 8)
}

func testTree_Left_Right(t *testing.T) {
	tree := rbt.New[int, int]()
	n := tree.Left()
	so(n, eq, nil)
	n = tree.Right()
	so(n, eq, nil)

	tree.Set(1, 1)
	tree.Set(10, 10)

	n = tree.Left()
	so(n.K, eq, 1)
	n = tree.Right()
	so(n.K, eq, 10)

	tree.Clear()
	n = tree.Left()
	so(n, eq, nil)
	n = tree.Right()
	so(n, eq, nil)
}

func testTree_Iterate(t *testing.T) {
	const size = 10000
	tree := rbt.New[int, string]()
	tree.Iterate(nil)                                    // not panic
	tree.Iterate(func(int, string) bool { return true }) // not panic

	for i := 0; i < size; i++ {
		tree.Set(i, fmt.Sprintf("%d%d", i, i))
	}

	keys := make([]int, 0, size)
	tree.Iterate(func(k int, v string) bool {
		if k >= size/2 {
			return false
		}
		so(k, eq, len(keys))
		so(v, eq, fmt.Sprintf("%d%d", k, k))
		keys = append(keys, k)
		return true
	})

	so(len(keys), eq, size/2)
	for i, v := range keys {
		so(i, eq, v)
	}
}
