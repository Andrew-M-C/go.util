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
	SetDrawColor(clr color.Color)
	DrawPoint(center Point, radius float64)
	DrawCircle(center Point, radius, width float64)
	DrawLine(from, to Point, width float64)
	Save(filepath string) error
}
