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

func hexRGBString(o draw.MergedOptions) string {
	r, g, b, _ := o.Color().RGBA()
	return fmt.Sprintf("#%02X%02X%02X", r>>8, g>>8, b>>8)
}

func rgbString(o draw.MergedOptions) string {
	r, g, b, _ := o.Color().RGBA()
	return fmt.Sprintf("rgb(%d,%d,%d)", r>>8, g>>8, b>>8)
}

func opacityString(o draw.MergedOptions) string {
	_, _, _, alpha := o.Color().RGBA()
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

// Size 返回画布大小
func (c *Canvas) Size() (width, height float64) {
	return c.width, c.height
}

// SetDrawColor 设置后续的绘制颜色
func (c *Canvas) SetDrawColor(clr color.Color) {
	if clr != nil {
		c.storage.color = clr
	}
}

// CurrentDrawColor 返回当前的颜色
func (c *Canvas) CurrentDrawColor() color.Color {
	if clr := c.storage.color; clr != nil {
		return clr
	}
	return color.Black
}

// DrawPoint 绘制一个实心圆
func (c *Canvas) DrawPoint(center draw.Point, radius float64) {
	o := draw.MergeOptions(c, nil)
	opt := fmt.Sprintf(`fill="%s" style="opacity:%s"`, hexRGBString(o), opacityString(o))
	c.svg().Circle(int(center.X), int(center.Y), absint(radius), opt)
}

// DrawHollowCircle 绘制一个空心圆
func (c *Canvas) DrawHollowCircle(center draw.Point, radius, width float64) {
	o := draw.MergeOptions(c, nil)
	opt := fmt.Sprintf(
		`stroke-width="%d" stroke="%s" style="opacity:%s" fill="none"`,
		absint(width), hexRGBString(o), opacityString(o),
	)
	c.svg().Circle(int(center.X), int(center.Y), absint(radius), opt)
}

// DrawLine 绘制一条线
func (c *Canvas) DrawLine(from, to draw.Point, width float64) {
	o := draw.MergeOptions(c, nil)
	opt := fmt.Sprintf(`style="stroke:%s;stroke-width:%d"`, rgbString(o), absint(width))
	c.svg().Line(int(from.X), int(from.Y), int(to.X), int(to.Y), opt)
}

// DrawHollowRect 绘制一个空心矩形
func (c *Canvas) DrawHollowRect(from, to draw.Point, width float64) {
	o := draw.MergeOptions(c, nil)
	opt := fmt.Sprintf(
		`stroke-width="%d" stroke="%s" style="opacity:%s" fill="none"`,
		absint(width), hexRGBString(o), opacityString(o),
	)
	c.svg().Rect(int(from.X), int(from.Y), int(to.X), int(to.Y), opt)
}

// DrawSolidRect 绘制一个实心矩形
func (c *Canvas) DrawSolidRect(from, to draw.Point) {
	o := draw.MergeOptions(c, nil)
	opt := fmt.Sprintf(`stroke-width="0" style="fill:%s;opacity:%s"`, rgbString(o), opacityString(o))
	c.svg().Rect(int(from.X), int(from.Y), int(to.X), int(to.Y), opt)
}

// Save 关闭并保存到文件
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

// -------- svg 特有逻辑 --------

func (c *Canvas) DrawText(origin draw.Point, text string, opts ...draw.Option) {
	o := draw.MergeOptions(c, opts)

	// 无需旋转
	if o.Rotate() == 0 {
		opt := fmt.Sprintf(
			`fill="%s" style="opacity:%s" font-size="%d"`,
			hexRGBString(o), opacityString(o), o.FontSize(),
		)
		c.svg().Text(int(origin.X), int(origin.Y), text, opt)
		return
	}

	// 需要旋转
	opt := fmt.Sprintf(
		`fill="%s" style="opacity:%s" font-size="%d" transform="translate(%d,%d) rotate(%.0f)"`,
		hexRGBString(o), opacityString(o), o.FontSize(),
		int(origin.X), int(origin.Y), o.Rotate(),
	)
	c.svg().Text(0, 0, text, opt)
}
