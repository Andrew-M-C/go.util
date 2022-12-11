// Package draw 提供简单的绘图工具
package draw

import (
	"image/color"

	"golang.org/x/exp/constraints"
)

// Number 表示任意数字类型
type Number interface {
	constraints.Integer | constraints.Float
}

// Point 表示一个点
type Point struct {
	X float64
	Y float64
}

// NewPoint 构建一个新的点
func NewPoint[T1, T2 Number](x T1, y T2) Point {
	return Point{
		X: float64(x),
		Y: float64(y),
	}
}

// Canvas 表示一个画板
type Canvas interface {
	Size() (width, height float64)

	SetDrawColor(clr color.Color)
	CurrentDrawColor() color.Color

	DrawPoint(center Point, radius float64)
	DrawHollowCircle(center Point, radius, width float64)

	DrawLine(from, to Point, width float64)

	DrawHollowRect(endpoint1, endpoint2 Point, width float64)
	DrawSolidRect(endpoint1, endpoint2 Point)

	Save(filepath string) error
}

// DrawPointXY 封装 DrawPoint, 但是提供具体的 XY 值而不是 Point
func DrawPointXY[TP, TR Number](canvas Canvas, centerX, centerY TP, radius TR) {
	canvas.DrawPoint(NewPoint(centerX, centerY), float64(radius))
}

// DrawHollowCircle 封装 DrawHollowCircle, 但是提供具体的 XY 值而不是 Point
func DrawHollowCircleXY[TC, TR, TW Number](canvas Canvas, centerX, centerY TC, radius TR, width TW) {
	center := NewPoint(centerX, centerY)
	canvas.DrawHollowCircle(center, float64(radius), float64(width))
}

// DrawLineXY 封装 DrawLine, 但是提供具体的 XY 值而不是 Point
func DrawLineXY[TP, TW Number](canvas Canvas, x1, y1, x2, y2 TP, width TW) {
	from := NewPoint(x1, y1)
	to := NewPoint(x2, y2)
	canvas.DrawLine(from, to, float64(width))
}

// DrawHollowRectXY 封装 DrawHollowRect, 但是提供具体的 XY 值而不是 Point
func DrawHollowRectXY[TP, TW Number](canvas Canvas, x1, y1, x2, y2 TP, width TW) {
	p1 := NewPoint(x1, y1)
	p2 := NewPoint(x2, y2)
	canvas.DrawHollowRect(p1, p2, float64(width))
}

// DrawSolidRectXY 封装 DrawSolidRect, 但是提供具体的 XY 值而不是 Point
func DrawSolidRectXY[T Number](canvas Canvas, x1, y1, x2, y2 T) {
	p1 := NewPoint(x1, y1)
	p2 := NewPoint(x2, y2)
	canvas.DrawSolidRect(p1, p2)
}

// TextCanvas 表示一个能绘图的画布
type TextCanvas interface {
	DrawText(origin Point, text string, opts ...Option)
}

// DrawTextXY 封装 DrawText, 但是提供具体的 XY 值而不是 Point
func DrawTextXY[T Number](canvas TextCanvas, x, y T, text string, opts ...Option) {
	p := NewPoint(x, y)
	canvas.DrawText(p, text, opts...)
}
