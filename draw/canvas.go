package draw

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"math"
	"os"

	"github.com/schollz/progressbar/v3"
)

// Canvas 表示一个画布, 需要用 NewCanvas 初始化
type Canvas struct {
	img *image.NRGBA

	width  int
	height int

	// div 表示坐标除以的系数, 必须不小于 1
	div float64

	defaultColor color.Color
}

func NewCanvas(width, height float64, opts ...Option) *Canvas {
	if width <= 1 {
		width = 1
	}
	if height <= 1 {
		height = 1
	}

	opt := mergeOptions(opts)

	c := &Canvas{}
	c.div = opt.div
	c.width = int(round(width / c.div))
	c.height = int(round(height / c.div))

	c.img = image.NewNRGBA(image.Rect(0, 0, c.width-1, c.height-1))
	c.defaultColor = opt.defaultColor

	return c
}

// 获取当次绘图颜色
func (c *Canvas) getDrawColor(opt *option) color.Color {
	if clr := opt.drawColor; clr != nil {
		return clr
	}
	return c.defaultColor
}

// 将点转为一个像素位
func (c *Canvas) pointToPixel(x, y float64) (int, int) {
	if x < 0 || y < 0 {
		return -1, -1
	}

	px := int(round(x / c.div)) // p 表示 pixel
	py := int(round(y / c.div))

	return px, py
}

func (c *Canvas) lengthToPixel(l float64) float64 {
	return l / c.div
}

func (c *Canvas) set(x, y int, clr color.Color) {
	if x < 0 || y < 0 {
		debugf("x(%d) < 0 or y(%d) < 0", x, y)
		return
	}
	if x >= c.width {
		debugf("x >= c.width (%d >= %d)", x, c.width)
		return
	}
	if y >= c.height {
		debugf("y >= c.height (%d >= %d)", y, c.height)
		return
	}
	c.img.Set(x, y, clr)
}

// 迭代第一象限的四分之一个圆
func (c *Canvas) iterateQuadrant(radius float64, f func(plusX, plusY int)) {
	radius = c.lengthToPixel(radius)
	r2 := radius * radius
	r08 := int(round(radius * 0.8))

	// 我们只需要计算八分之一个圆就可以了
	cacheX := make([]int, 0, int(r08)+1)
	cacheY := make([]int, 0, int(r08)+1)

	// Y 轴 -> 45 度
	cacheX = append(cacheX, 0)
	cacheY = append(cacheY, int(round(radius)))
	f(cacheX[0], cacheY[0])
	for x := 1; x <= r08; x++ {
		x2 := x * x
		y := math.Sqrt(r2 - float64(x2))
		cacheX = append(cacheX, x)
		cacheY = append(cacheY, int(round(y)))
		f(cacheX[x], cacheY[x])
	}

	// 45 度 -> X 轴
	for i := len(cacheX) - 1; i >= 0; i-- {
		x := cacheY[i] // 反转 X, Y
		y := cacheX[i]
		f(x, y)
	}
}

// WritePNGFile 写入 PNG 格式的文件
func (c *Canvas) WritePNGFile(filepath string) error {
	f, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("create file error: %w", err)
	}
	defer f.Close()

	bar := progressbar.NewOptions(-1,
		progressbar.OptionShowBytes(true),
		progressbar.OptionShowIts(),
	)
	stream := io.MultiWriter(f, bar)

	err = png.Encode(stream, c.img)
	if err != nil {
		return fmt.Errorf("png.Encode error: %w", err)
	}
	return nil
}

// SetDefaultColor 设置默认画笔
func (c *Canvas) SetDefaultColor(clr color.Color) {
	if clr != nil {
		c.defaultColor = clr
	}
}

// DrawDot 绘制最大为一个像素的点
func (c *Canvas) DrawDot(x, y float64, opts ...Option) {
	px, py := c.pointToPixel(x, y)
	o := mergeOptions(opts)
	clr := c.getDrawColor(o)
	c.set(px, py, clr)
}

// DrawPoint 绘制一个点, 也可以用来绘制一个实心圆
func (c *Canvas) DrawPoint(centerX, centerY, width float64, opts ...Option) {
	cx, cy := c.pointToPixel(centerX, centerY) // c 表示 center
	o := mergeOptions(opts)
	clr := c.getDrawColor(o)

	// 绘制圆圈
	c.iterateQuadrant(width, func(plusX, plusY int) {
		debugf("(%v, %v)", plusX, plusY)
		for x := cx - plusX; x <= cx+plusX; x++ {
			c.set(x, cy-plusY, clr)
			c.set(x, cy+plusY, clr)
		}
	})
}
