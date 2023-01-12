package svg

import (
	"image/color"
	"os"
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
	const filepath = "./svg_test.svg"
	os.Remove(filepath)

	c := NewCanvas(200, 100)

	// 边缘镶一圈矩形
	c.SetDrawColor(draw.White)
	c.DrawHollowRect(draw.NewPoint(0, 0), draw.NewPoint(200, 100), 1)
	draw.DrawHollowRectXY(c, 0, 0, 200, 100, 1)

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

	// 写一个文字
	draw.DrawTextXY(c, 100, 50, "你好, 世界!", draw.OptColor(draw.Red), draw.OptFontSize(7))
	draw.DrawTextXY(c, 10, 50, "Hello, world!", draw.OptColor(draw.Red), draw.OptRotate(30))

	// 画一个空心平行四边形
	c.DrawHollowPolygon(1, draw.P(100, 0), draw.P(125, 25), draw.P(150, 25), draw.P(125, 0))

	// 画一个是心三角形
	c.SetDrawColor(draw.Red)
	c.DrawSolidPolygon(draw.P(200, 0), draw.P(200, 25), draw.P(175, 0))

	// 保存 svg
	err := c.Save(filepath)
	so(err, isNil)
}
