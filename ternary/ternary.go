package ternary

// Check 实现一个三元表达式
func Check[T any](b bool, ifTrue, ifFalse T) T {
	if b {
		return ifTrue
	}
	return ifFalse
}
