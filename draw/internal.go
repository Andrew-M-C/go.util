package draw

import (
	"fmt"
	"math"
)

func round(f float64) float64 {
	return math.Ceil(f + 0.5)
}

func debugf(f string, a ...any) {
	// do nothing
	fmt.Printf(f+"\n", a...)
}
