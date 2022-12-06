package svg

import (
	"image/color"
	"testing"

	"github.com/Andrew-M-C/go.util/draw"
	"github.com/smartystreets/goconvey/convey"
)

var (
	cv = convey.Convey
	so = convey.So

	isNil = convey.ShouldBeNil
)

func TestSVG(t *testing.T) {
	cv("测试常规操作", t, func() { testGeneral(t) })
}

func testGeneral(t *testing.T) {
	c := NewCanvas(200, 100)

	// 边缘镶一圈矩形
	c.SetDrawColor(draw.White)
	c.DrawHollowRect(draw.NewPoint(0, 0), draw.NewPoint(200, 100), 1)

	// 角落里画一个实心矩形
	c.SetDrawColor(draw.Blue)
	c.DrawSolidRect(draw.NewPoint(150, 50), draw.NewPoint(198, 98))

	// 画一个小点
	c.SetDrawColor(draw.Red)
	c.DrawPoint(draw.NewPoint(150, 50), 1)

	// 画一个实心圆
	c.SetDrawColor(color.Gray{0x7F})
	c.DrawPoint(draw.NewPoint(50, 50), 50)

	// 画一个空心圆
	c.SetDrawColor(draw.Black)
	c.DrawHollowCircle(draw.NewPoint(100, 50), 25, 5)

	// 画一条直线
	c.SetDrawColor(draw.Green)
	c.DrawLine(draw.NewPoint(150, 150), draw.NewPoint(100, 50), 1)

	// 保存
	err := c.Save("./svg_test.svg")
	so(err, isNil)
}
