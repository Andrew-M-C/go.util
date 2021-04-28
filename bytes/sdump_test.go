package bytes

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func test(t *testing.T, scene string, f func(*testing.T)) {
	if t.Failed() {
		return
	}
	Convey(scene, t, func() {
		f(t)
	})
}

func TestSDunp(t *testing.T) {
	test(t, "SDump", testSDump)
}

func testSDump(t *testing.T) {
	s := "你好, 世界! Hello, world！Wonderful.\n"

	s = SDump([]byte(s), s)
	Printf(s)
	// TODO:
}
