package math

import (
	"math"
	"testing"
)

// go test -bench=. -run=none -benchmem -benchtime=2s

func Benchmark_1to1000_newtonIntSqrt(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for n := 1; n <= 1000; n++ {
			newtonIntSqrt(uint64(n))
		}
	}
}

func Benchmark_1to1000_bitwiseSqrt(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for n := 1; n <= 1000; n++ {
			bitwiseSqrt(uint64(n))
		}
	}
}

func Benchmark_1000to1000000_newtonIntSqrt(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for n := 1000; n <= 1000000; n += 1000 {
			newtonIntSqrt(uint64(n))
		}
	}
}

func Benchmark_1000to1000000_bitwiseSqrt(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for n := 1000; n <= 1000000; n += 1000 {
			bitwiseSqrt(uint64(n))
		}
	}
}

func Benchmark_1000000to1000000000_newtonIntSqrt(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for n := 1000000; n <= 1000000000; n += 1000000 {
			newtonIntSqrt(uint64(n))
		}
	}
}

func Benchmark_1000000to1000000000_bitwiseSqrt(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for n := 1000000; n <= 1000000000; n += 1000000 {
			bitwiseSqrt(uint64(n))
		}
	}
}

func Benchmark_0x1000000000000000_float64(b *testing.B) {
	for i := 0; i < b.N; i++ {
		math.Sqrt(float64(0x1000000000000000))
	}
}

func Benchmark_0x1000000000000000_SqrtUint(b *testing.B) {
	for i := 0; i < b.N; i++ {
		math.Sqrt(0x1000000000000000)
	}
}

func Benchmark_0x1FFFFFFFFFFFFFFF_float64(b *testing.B) {
	for i := 0; i < b.N; i++ {
		math.Sqrt(float64(0x1FFFFFFFFFFFFFFF))
	}
}

func Benchmark_0x1FFFFFFFFFFFFFFF_SqrtUint(b *testing.B) {
	for i := 0; i < b.N; i++ {
		math.Sqrt(0x1FFFFFFFFFFFFFFF)
	}
}
