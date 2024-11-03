package sync

import "testing"

func testPool(*testing.T) {
	pool := NewPool[*int](func() *int {
		return new(int)
	})

	ptr1 := pool.Get()
	so(*ptr1, eq, 0)
	*ptr1 = 1

	ptr2 := pool.Get()
	so(*ptr2, eq, 0)
	*ptr2 = 2

	pool.Put(ptr2)
	ptr3 := pool.Get()
	so(*ptr3, eq, 2)
}
