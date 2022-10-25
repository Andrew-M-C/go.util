package bytes

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

var (
	cv = convey.Convey

	printf = convey.Printf
)

func test(t *testing.T, scene string, f func(*testing.T)) {
	if t.Failed() {
		return
	}
	cv(scene, t, func() {
		f(t)
	})
}

func TestSDunp(t *testing.T) {
	test(t, "SDump", testSDump)
}

func testSDump(t *testing.T) {
	s := "你好, 世界! Hello, world！Wonderful.\n"

	s = SDump([]byte(s), s)
	printf(s)
	// TODO:
}
