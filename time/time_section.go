package time

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Point 表示一个时间点, 用于比较。支持 30 小时制 (实际上支持到 48 小时)
type Point struct {
	Hour   int
	Minute int
	Second int
}

// NewPoint 创建一个时间点
func NewPoint[I Integer](hour, minute, second I) Point {
	return Point{
		Hour:   int(hour),
		Minute: int(minute),
		Second: int(second),
	}
}

func (p Point) String() string {
	return fmt.Sprintf("%d:%02d:%02d", p.Hour, p.Minute, p.Second)
}

func (p Point) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

func (p *Point) UnmarshalText(text []byte) error {
	var err error
	parts := strings.Split(string(text), ":")

	// hour
	if p.Hour, err = parseInt(parts[0]); err != nil {
		return fmt.Errorf("invalid hour: '%v'", parts[0])
	}
	if p.Hour > 48 {
		p.Hour = 0
		return fmt.Errorf("invalid hour: '%v'", parts[0])
	}

	// minute
	p.Minute, p.Second = 0, 0
	if len(parts) < 2 {
		return nil
	}
	if p.Minute, err = parseInt(parts[1]); err != nil {
		return fmt.Errorf("invalid minute: '%v'", parts[1])
	}
	if p.Minute > 60 {
		p.Minute = 0
		return fmt.Errorf("invalid minute: '%v'", parts[1])
	}

	// second
	if len(parts) < 3 {
		return nil
	}
	if p.Second, err = parseInt(parts[2]); err != nil {
		return fmt.Errorf("invalid second: '%v'", parts[2])
	}
	if p.Second > 60 {
		p.Second = 0
		return fmt.Errorf("invalid second: '%v'", parts[2])
	}

	return nil
}

func parseInt(s string) (int, error) {
	u, err := strconv.ParseUint(s, 10, 8)
	if err != nil {
		return 0, err
	}
	return int(u), nil
}

// Range 表示一个时间范围, 用于判断时间是否在这个范围内
type Range struct {
	tm         time.Time
	left       Point
	leftEqual  bool
	right      Point
	rightEqual bool
}

// DetermineRange 确定时间是否在指定范围内
func DetermineRange(tm time.Time) *Range {
	return &Range{
		tm: tm,
	}
}

// GE 设定 p <= tm
func (r *Range) GE(p Point) *Range {
	r.left = p
	r.leftEqual = true
	return r
}

// GT 设定 p < tm
func (r *Range) GT(p Point) *Range {
	r.left = p
	r.leftEqual = false
	return r
}

// LE 设定 tm <= p
func (r *Range) LE(p Point) *Range {
	r.right = p
	r.rightEqual = true
	return r
}

// LT 设定 tm < p
func (r *Range) LT(p Point) *Range {
	r.right = p
	r.rightEqual = false
	return r
}

// Check 根据已经配置好的范围, 检查时间是否在范围内
func (r *Range) Check() bool {
	loc := r.tm.Location()

	var left, right time.Time

	if r.left.Hour < 24 && r.right.Hour > 24 {
		// 跨越子夜的情况
		left = time.Date(r.tm.Year(), r.tm.Month(), r.tm.Day(), r.left.Hour, r.left.Minute, r.left.Second, 0, loc)
		right = time.Date(r.tm.Year(), r.tm.Month(), r.tm.Day(), r.right.Hour-24, r.right.Minute, r.right.Second, 0, loc)
		right = right.Add(24 * time.Hour)
	} else if r.left.Hour > 24 && r.right.Hour > 24 {
		// 没跨越子夜, 但是两面都有在 24 小时i之后
		left = time.Date(r.tm.Year(), r.tm.Month(), r.tm.Day(), r.left.Hour-24, r.left.Minute, r.left.Second, 0, loc)
		right = time.Date(r.tm.Year(), r.tm.Month(), r.tm.Day(), r.right.Hour-24, r.right.Minute, r.right.Second, 0, loc)
	} else {
		// 没跨越子夜, 数据都是正常的
		left = time.Date(r.tm.Year(), r.tm.Month(), r.tm.Day(), r.left.Hour, r.left.Minute, r.left.Second, 0, loc)
		right = time.Date(r.tm.Year(), r.tm.Month(), r.tm.Day(), r.right.Hour, r.right.Minute, r.right.Second, 0, loc)
	}

	// left > right 的话, 必然是 false
	if left.After(right) {
		return false
	}

	// 左边界
	if !left.Before(r.tm) {
		return r.leftEqual && left.Equal(r.tm)
	}

	// 右边界
	if !right.After(r.tm) {
		return r.rightEqual && right.Equal(r.tm)
	}

	return true
}
