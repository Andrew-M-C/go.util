package draw

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

var (
	cv = convey.Convey
	so = convey.So

	isNil = convey.ShouldBeNil
)

func TestDraw(t *testing.T) {
	cv("普通的画圆圈", t, func() { testCanvasFilledCircle(t) })
}

func testCanvasFilledCircle(t *testing.T) {
	c := NewCanvas(801, 401, WithZoomOutFactor(1))
	c.DrawPoint(200, 200, 200)

	err := c.WritePNGFile("./simple_circle_test.png")
	so(err, isNil)
}
