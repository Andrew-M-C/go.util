package tea_test

import (
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/Andrew-M-C/go.util/ids/tea"
	"github.com/smartystreets/goconvey/convey"
)

var (
	cv = convey.Convey
	so = convey.So
	eq = convey.ShouldEqual
	ne = convey.ShouldNotEqual
)

func TestTea(t *testing.T) {
	cv("基本逻辑", t, func() {
		id := uint64(0x1234)
		key := tea.Key{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}
		enc := tea.Encrypt(id, key)

		t.Logf("%x --> %X", id, enc)
		so(enc, ne, id)

		dec := tea.Decrypt(enc, key)
		t.Logf("%X", dec)
		so(dec, eq, id)
	})

	cv("简单并发一下", t, func() {
		const concurrency = 10000
		wg := sync.WaitGroup{}
		lck := sync.Mutex{}

		wg.Add(concurrency)

		var inList []uint64
		var encList []uint64
		var outList []uint64

		start := time.Now()
		for i := 0; i < concurrency; i++ {
			go func() {
				defer wg.Done()
				key := tea.Key{}
				for i := range key {
					key[i] = byte(rand.Int31n(256))
				}

				id := rand.Uint64()
				enc := tea.Encrypt(id, key)
				dec := tea.Decrypt(enc, key)

				lck.Lock()
				inList = append(inList, id)
				encList = append(encList, enc)
				outList = append(outList, dec)
				lck.Unlock()
			}()
		}

		wg.Wait()
		ela := time.Since(start)
		t.Logf("并发数 %d, 耗时 %v", concurrency, ela)

		for i, id := range inList {
			so(id, ne, encList[i])
			so(id, eq, outList[i])
		}
	})
}
