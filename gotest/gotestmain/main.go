package main

import (
	"fmt"

	utilgotest "github.com/Andrew-M-C/go.util/gotest"
)

func main() {
	if utilgotest.IsInGoTest() {
		fmt.Println("ERROR, should NOT in go test mode")
	}
	fmt.Println("OK, now is not in go test mode")
	return
}
