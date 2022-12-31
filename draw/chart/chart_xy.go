package chart

import (
	"errors"
	"math"
	"strconv"

	"github.com/Andrew-M-C/go.util/draw"
	"github.com/Andrew-M-C/go.util/maps"
)

// NewXY 新建一个简单的绘图工具
func NewXY() *XY {
	xy := &XY{}
	xy.lazyInit()
	return xy
}

type XY struct {
	clusters map[string]map[float64]float64 // string: 行名称
}

func (xy *XY) lazyInit() {
	if xy.clusters == nil {
		xy.clusters = make(map[string]map[float64]float64, 1)
	}
}

// Add 添加一个数据
func (xy *XY) Add(name string, x, y float64) {
	data := xy.clusters[name]
	if data == nil {
		data = make(map[float64]float64, 1)
		xy.clusters[name] = data
	}

	data[x] = y
}

// Draw 绘图
func (xy *XY) Draw(canvas draw.TextCanvas, opts ...Option) error {
	opt := mergeOpts(opts)

	d := &drawXY{
		xy:  xy,
		c:   canvas,
		opt: opt,
	}

	return d.do()

	// 然后绘图
	// TODO:
}

type drawXY struct {
	xy  *XY
	c   draw.TextCanvas
	opt *opt

	allNames     maps.KeyList[string]
	xListByNames map[string][]float64

	xMin, xMax float64
	yMin, yMax float64

	// origin 表示原点
	origin struct {
		x, y float64
	}
}

func (d *drawXY) do() error {
	steps := []func() error{
		d.readInitData,  // 首先读取初始化数据
		d.calcDrawParam, // 计算绘制参数
		// TODO:
	}

	for _, step := range steps {
		if err := step(); err != nil {
			return err
		}
	}
	return nil
}

func (d *drawXY) readInitData() error {
	d.allNames = maps.Keys(d.xy.clusters)
	d.xListByNames = make(map[string][]float64, len(d.allNames))

	d.xMin, d.xMax = math.NaN(), math.NaN()
	d.yMin, d.yMax = math.NaN(), math.NaN()

	// 首先取所有数据
	for _, name := range d.allNames {
		xList := maps.Keys(d.xy.clusters[name]).SortAsc()
		d.xListByNames[name] = xList

		// X 范围
		if math.IsNaN(d.xMin) {
			d.xMin = xList[0]
			d.xMax = xList[len(xList)-1]
			continue
		}
		if x := xList[0]; x < d.xMin {
			d.xMin = x
		}
		if x := xList[len(xList)-1]; x > d.xMax {
			d.xMax = x
		}

		// Y 范围
		for _, y := range xList {
			if math.IsNaN(d.yMin) {
				d.yMin, d.yMax = y, y
				continue
			}
			if y < d.yMin {
				d.yMin = y
			} else if y > d.yMax {
				d.yMax = y
			}
		}
	}

	return nil
}

func (d *drawXY) calcDrawParam() error {
	// 首先计算需要绘制的坐标网格
	// 现阶段方案暂时只支持将所有的数据绘制在同一个 Y 坐标上, 并且只能显式指定坐标间隔
	if d.opt.xScale <= 0 {
		return errors.New("please specify X scale")
	}
	if d.opt.yScale <= 0 {
		return errors.New("please specify Y scale")
	}

	w, h := d.c.Size()
	d.origin.y = h - float64(d.opt.fontSize)*2

	// 首先绘制 Y 轴文字
	maxStringLen := 0
	corY := d.origin.y
	step := (h - d.origin.y) / (d.yMax - d.yMin)
	debug("Y step: %f", step)
	for y := d.yMin; y <= d.yMax; {
		s := strconv.FormatFloat(y, 'g', -1, 64)
		if len(s) > maxStringLen {
			maxStringLen = len(s)
		}

		draw.DrawTextXY(d.c, 0, corY/2, s)

		corY -= step
		y += d.opt.yScale
	}

	d.origin.x = float64(maxStringLen)

	// 绘制 X 轴文字
	// TODO:

	_ = w
	return nil
}
