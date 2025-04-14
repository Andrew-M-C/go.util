// Package admindivision 实现中国统计用行政区划查询工具
package admindivision

import (
	"slices"
	"strconv"
	"strings"

	"github.com/Andrew-M-C/go.util/slice"
	"github.com/agnivade/levenshtein"
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
	// 是否历史已弃用的行政区划
	deprecated bool
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

// 是否已撤销
func (d *Division) Deprecated() bool {
	return d.deprecated
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
	target := &Division{
		code: code,
	}
	idx := slice.BinarySearchOne(d.sub, target, divComp)
	if idx < 0 {
		return nil
	}
	return d.sub[idx]
}

func divComp(a, b *Division) int {
	return strings.Compare(a.code, b.code)
}

var (
	china = &Division{}
)

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

// MatchDivisionByName 按照一个行政区划名称搜索行政节点层级链, 必须以省级行政区开始查询,
// 而且必须与数据库中的名称完全一致
func MatchDivisionByName(name ...string) []*Division {
	if len(name) == 0 {
		return nil
	}

	// 从省级开始查起
	var res []*Division
	curr := china
	for _, n := range name {
		// 遍历当前节点寻找匹配的名称
		found := false
		for _, sub := range curr.SubDivisions() {
			if sub.name != n {
				continue
			}
			curr = sub
			found = true
			res = append(res, sub)
			break
		}
		if !found {
			return res
		}
	}
	return res
}

// SearchDivisionByName 按照一个行政区划名称搜索行政节点层级链, 必须以省级行政区开始查询,
// 按照编辑距离 (levenshtein) 进行匹配查询
func SearchDivisionByName(name ...string) []*Division {
	if len(name) == 0 {
		return nil
	}

	// 从省级开始查起
	var res []*Division
	distance := 0
	curr := china
	for _, n := range name {
		// 遍历当前节点寻找接近的名称
		var closest *Division
		distance = len(n)
		for _, sub := range curr.SubDivisions() {
			dist := levenshtein.ComputeDistance(sub.name, n)
			if dist == 0 {
				closest = sub
				break
			}
			if dist < distance {
				distance = dist
				closest = sub
			}
		}
		if closest == nil {
			return res
		}
		res = append(res, closest)
		curr = closest
	}

	return res
}

// JoinDivisionCodes 将一个区划链的代码连接成一个字符串。注意, 仅按照层级 join, 不包含最后的补零
func JoinDivisionCodes(divisions []*Division) string {
	buff := strings.Builder{}
	for _, d := range divisions {
		buff.WriteString(d.code)
	}
	return buff.String()
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
