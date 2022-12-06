// Package svg 提供基于 svg 的绘图工具
package svg

import "fmt"

func debugf(f string, a ...interface{}) {
	fmt.Printf(f+"\n", a...)
}

var _ = debugf
