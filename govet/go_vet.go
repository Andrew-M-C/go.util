// Package govet 封装在 go vet 检查的工具
package govet

// NoCopy 用来在代码中嵌入, 用于 go vet 中提示不要 copy
type NoCopy struct{}

func (*NoCopy) Lock()   {}
func (*NoCopy) Unlock() {}
