// Package admindivision 实现中国统计用行政区划查询工具
package admindivision

import (
	"slices"
	"strconv"
	"strings"
)

// AdministrativeLevel 行政层级
type AdministrativeLevel int

const (
	Province AdministrativeLevel = iota + 1
	City
	County
	// Town
	// Village
)

// Division 表示一级行政区划
type Division struct {
	level    AdministrativeLevel
	code     string
	fullCode string
	name     string
	virtual  bool
	sub      []*Division
}

// Level 返回行政层级
func (d *Division) Level() AdministrativeLevel {
	return d.level
}

// 单独区域代码, 不包含上级节点
func (d *Division) Code() string {
	return d.code
}

// FullCode 完整区域代码, 包含上级节点
func (d *Division) FullCode() string {
	return d.fullCode
}

// 官方名称, 不包含上级节点
func (d *Division) Name() string {
	return d.name
}

// 是否虚拟行政节点。直辖市、香港、澳门的 Virtual() 返回 true
func (d *Division) Virtual() bool {
	return d.virtual
}

// SubDivisions 获取下一层级的区划列表
func (d *Division) SubDivisions() []*Division {
	if len(d.sub) == 0 {
		return nil
	}
	return slices.Clone(d.sub)
}

// SubDivisionByCode 按下一层及的子代码查询行政区划, 如果查不到则返回 nil
func (d *Division) SubDivisionByCode(code string) *Division {
	sub := d.sub

	// 二分查找
	left, right := 0, len(d.sub)
	for left < right {
		mid := (left + right) / 2
		switch {
		case sub[mid].code < code:
			left = mid
		case sub[mid].code > code:
			right = mid
		default:
			return sub[mid]
		}
	}

	return nil
}

var china = &Division{}

// Provinces 返回省份列表
func Provinces() []*Division {
	return china.SubDivisions()
}

// ProvinceByCode 按代码查找省级行政区
func ProvinceByCode(code string) *Division {
	return china.SubDivisionByCode(code)
}

// SearchDivisionByCode 按照一个行政区划搜索行政节点层级链
func SearchDivisionByCode(code string) []*Division {
	codeChain := splitCodesToChain(code)
	if len(codeChain) == 0 {
		return nil
	}

	var res []*Division
	for i, c := range codeChain {
		var div *Division
		if i == 0 {
			div = ProvinceByCode(c)
		} else {
			div = res[len(res)-1].SubDivisionByCode(c)
		}
		if div == nil {
			return res
		}
		res = append(res, div)
	}

	return res
}

// DescribeDivisionChain 描述一个区划链
func DescribeDivisionChain(divisions []*Division, sep string) string {
	var parts []string
	for _, d := range divisions {
		if !d.virtual || d.level == Province {
			parts = append(parts, d.name)
		}
	}
	return strings.Join(parts, sep)
}

func splitCodesToChain(code string) []string {
	if _, err := strconv.ParseUint(code, 10, 64); err != nil {
		return nil
	}

	switch len(code) {
	case 2:
		return []string{code}
	case 4:
		return []string{code[:2], code[2:]}
	case 6:
		return []string{code[:2], code[2:4], code[4:]}
	case 9:
		return []string{code[:2], code[2:4], code[4:6], code[6:]}
	case 12:
		// 例: "110101001001"
		return []string{code[:2], code[2:4], code[4:6], code[6:9], code[9:]}
	default:
		return nil
	}
}
