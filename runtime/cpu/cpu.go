package cpu

// CoreNum 返回当前环境的 CPU 数
func CoreNum() int {
	return internal.cpuNum
}
