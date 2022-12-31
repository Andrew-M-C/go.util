package main

import "fmt"

func main() {}

func infof(f string, a ...interface{}) {
	s := fmt.Sprintf(f, a...)
	fmt.Sprintln(s)
}
