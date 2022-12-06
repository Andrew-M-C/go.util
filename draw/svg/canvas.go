package svg

import (
	"bytes"
	"fmt"
	"image/color"
	"os"

	"github.com/Andrew-M-C/go.util/draw"
	svg "github.com/ajstarks/svgo"
)

var _ draw.Canvas = (*Canvas)(nil)

// Canvas 表示一个 svg 画布
type Canvas struct {
	width  float64
	height float64

	storage struct {
		buff  bytes.Buffer
		color color.Color
		svg   *svg.SVG
	}
}

func abs[T draw.Number](i T) float64 {
	if i < 0 {
		return float64(-i)
	}
	return float64(i)
}

func absint[T draw.Number](i T) int {
	if i < 0 {
		return int(-i)
	}
	return int(i)
}

func (c *Canvas) color() color.Color {
	if clr := c.storage.color; clr != nil {
		return clr
	}
	return color.Black
}

func (c *Canvas) hexRGBString() string {
	r, g, b, _ := c.color().RGBA()
	return fmt.Sprintf("#%02X%02X%02X", r>>8, g>>8, b>>8)
}

func (c *Canvas) rgbString() string {
	r, g, b, _ := c.color().RGBA()
	return fmt.Sprintf("rgb(%d,%d,%d)", r>>8, g>>8, b>>8)
}

func (c *Canvas) opacityString() string {
	_, _, _, alpha := c.color().RGBA()
	return fmt.Sprintf("%f", float64(alpha)/0xFFFF)
}

func (c *Canvas) svg() *svg.SVG {
	if cvs := c.storage.svg; cvs != nil {
		return cvs
	}
	s := svg.New(&c.storage.buff)
	s.Start(int(c.width), int(c.height))
	c.storage.svg = s
	return s
}

// 新建一个 svg 画布
func NewCanvas[T1, T2 draw.Number](width T1, height T2) *Canvas {
	c := &Canvas{
		width:  abs(width),
		height: abs(height),
	}
	return c
}

// SetDrawColor 设置后续的绘制颜色
func (c *Canvas) SetDrawColor(clr color.Color) {
	if clr != nil {
		c.storage.color = clr
	}
}

// DrawPoint 绘制一个实心圆
func (c *Canvas) DrawPoint(center draw.Point, radius float64) {
	opt := fmt.Sprintf(`fill="%s" style="opacity:%s"`, c.hexRGBString(), c.opacityString())
	c.svg().Circle(int(center.X), int(center.Y), absint(radius), opt)
}

// DrawHollowCircle 绘制一个空心圆
func (c *Canvas) DrawHollowCircle(center draw.Point, radius, width float64) {
	opt := fmt.Sprintf(
		`stroke-width="%d" stroke="%s" style="opacity:%s" fill="none"`,
		absint(width), c.hexRGBString(), c.opacityString(),
	)
	c.svg().Circle(int(center.X), int(center.Y), absint(radius), opt)
}

// DrawLine 绘制一条线
func (c *Canvas) DrawLine(from, to draw.Point, width float64) {
	opt := fmt.Sprintf(`style="stroke:%s;stroke-width:%d"`, c.rgbString(), absint(width))
	c.svg().Line(int(from.X), int(from.Y), int(to.X), int(to.Y), opt)
}

// DrawHollowRect 绘制一个空心矩形
func (c *Canvas) DrawHollowRect(from, to draw.Point, width float64) {
	opt := fmt.Sprintf(
		`stroke-width="%d" stroke="%s" style="opacity:%s" fill="none"`,
		absint(width), c.hexRGBString(), c.opacityString(),
	)
	c.svg().Rect(int(from.X), int(from.Y), int(to.X), int(to.Y), opt)
}

// DrawSolidRect 绘制一个实心矩形
func (c *Canvas) DrawSolidRect(from, to draw.Point) {
	opt := fmt.Sprintf(`stroke-width="0" style="fill:%s;opacity:%s"`, c.rgbString(), c.opacityString())
	c.svg().Rect(int(from.X), int(from.Y), int(to.X), int(to.Y), opt)
}

// Save 保存到文件
func (c *Canvas) Save(filepath string) error {
	f, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("open filepath '%s' error: %w", filepath, err)
	}
	defer f.Close()

	c.svg().End()
	_, err = f.Write(c.storage.buff.Bytes())
	if err != nil {
		return fmt.Errorf("write to file error: %w", err)
	}

	c.storage.buff.Reset()
	c.storage.svg = nil

	return nil
}
