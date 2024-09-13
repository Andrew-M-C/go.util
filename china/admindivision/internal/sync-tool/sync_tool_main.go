// Package main 用来生成中国最新的行政区划列表
package main

import (
	"log"
)

func main() {
	printf("Starts")
	defer printf("Done")

	sess := &session{}

	procedures := []func() error{
		sess.getAndParseProvinces,
		sess.getAndParseCities,
		sess.getAndParseCounties,
		sess.getAndParseTowns,
		sess.getAndParseVillages,
		sess.writeToGoFile,
	}
	for i, p := range procedures {
		if err := p(); err != nil {
			errorf("执行第 %d 阶段操作失败: %v", i+1, err)
			return
		}
	}
}

var (
	printf = log.Printf
	errorf = log.Fatalf
)
