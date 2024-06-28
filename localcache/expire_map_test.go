package localcache

import (
	"testing"
	"time"
)

func TestExpireMap(t *testing.T) {
	cv("基础逻辑", t, func() {
		expire := 100 * time.Millisecond
		_, err := NewExpireMap[string, int](-1)
		so(err, isErr)

		m, err := NewExpireMap[string, int](expire)
		so(err, isNil)

	})
}
