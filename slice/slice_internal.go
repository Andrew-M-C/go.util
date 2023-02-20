package slice

import (
	"math/rand"
	"time"
)

var internal = struct {
	rander rand.Source
	debugf func(string, ...interface{})
}{
	rander: rand.NewSource(time.Now().UnixMicro()),
	debugf: func(string, ...interface{}) {
		// do nothing
	},
}

func internalInt63n(n int64) int64 {
	if n&(n-1) == 0 { // n is power of two, can mask
		return internal.rander.Int63() & (n - 1)
	}
	max := int64((1 << 63) - 1 - (1<<63)%uint64(n))
	v := internal.rander.Int63()
	for v > max {
		v = internal.rander.Int63()
	}
	return v % n
}

func internalInt31n(n int32) int32 {
	if n&(n-1) == 0 { // n is power of two, can mask
		return internalInt31() & (n - 1)
	}
	max := int32((1 << 31) - 1 - (1<<31)%uint32(n))
	v := internalInt31()
	for v > max {
		v = internalInt31()
	}
	return v % n
}

func internalInt31() int32 {
	return int32(internal.rander.Int63() >> 32)
}
