// Package admindivision 实现中国统计用行政区划查询工具
package admindivision

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	sliceutil "github.com/Andrew-M-C/go.util/slices"
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

func (l AdministrativeLevel) String() string {
	switch l {
	case Province:
		return "省级行政区"
	case City:
		return "市级行政区"
	case County:
		return "区县级行政区"
	default:
		return fmt.Sprintf("非法值 %d", l)
	}
}

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

func (d Division) String() string {
	return fmt.Sprintf("%s (%v, %v)", d.name, d.level, d.code)
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

// SubDivisionByCode 按下一层级的子代码查询行政区划, 如果查不到则返回 nil
func (d *Division) SubDivisionByCode(code string) *Division {
	target := &Division{
		code: code,
	}
	idx := sliceutil.BinarySearchOne(d.sub, target, divComp)
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
// 如果查找不到则按照前缀匹配
func SearchDivisionByName(name ...string) []*Division {
	if len(name) == 0 {
		return nil
	}

	// 从省级开始查起
	var res []*Division
	curr := china
	for _, n := range name {
		closest := findClosestDivision(curr, n)
		if closest == nil {
			break
		}
		res = append(res, closest)
		curr = closest
	}

	// 判断是不是省直辖行政区划
	if len(res) == 1 && len(name) == 2 {
		res = searchDivisionByNameAndTryDirectCounty(res, name)
	}
	// 判断是不是直辖市行政区划
	if len(res) == 1 && len(name) == 2 {
		res = searchDivisionByNameAndTryDirectCity(res, name)
	}
	// 重庆市下辖县的情况
	if len(res) == 1 && len(name) == 2 {
		res = searchDivisionByNameAndTryChongqingCounty(res, name)
	}

	return res
}

// searchDivisionByNameAndTryDirectCounty 是 SearchDivisionByName 的子函数, 用于处理直辖市行政区划
func searchDivisionByNameAndTryDirectCounty(res []*Division, name []string) []*Division {
	// 看看有没有省直辖行政区划节点
	directAdmin := res[0].SubDivisionByCode("90")
	if directAdmin == nil {
		return res // 找不到, 算了
	}
	// 从省直辖行政区划节点开始重新查询
	sub := findClosestDivision(directAdmin, name[1])
	if sub == nil {
		return res
	}
	return append(res, directAdmin, sub)
}

// 处理直辖市行政区划
func searchDivisionByNameAndTryDirectCity(res []*Division, name []string) []*Division {
	// 首先看看是不是直辖市
	switch res[0].code {
	case "11", "12", "31", "50":
		// continue
	default:
		return res // 不是直辖市, 不需要检查
	}

	// 看看有没有直辖市行政区划节点
	directAdmin := res[0].SubDivisionByCode("01")
	if directAdmin == nil {
		return res // 找不到, 算了
	}
	sub := findClosestDivision(directAdmin, name[1])
	if sub == nil {
		return res
	}
	return append(res, directAdmin, sub)
}

// 处理重庆市下面的县。其他的几个直辖市没有县
func searchDivisionByNameAndTryChongqingCounty(res []*Division, name []string) []*Division {
	if res[0].code != "50" {
		return res // 不是重庆市, 不需要检查
	}

	// 县
	directAdmin := res[0].SubDivisionByCode("02")
	if directAdmin == nil {
		return res // 找不到, 算了
	}
	sub := findClosestDivision(directAdmin, name[1])
	if sub == nil {
		return res
	}
	return append(res, directAdmin, sub)
}

func findClosestDivision(curr *Division, name string) *Division {
	// 遍历当前节点寻找接近的名称
	var closest *Division
	for _, sub := range curr.SubDivisions() {
		if strings.HasPrefix(sub.name, name) {
			if !sub.deprecated {
				closest = sub
				break
			}
			closest = sub
		}
	}
	return closest
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
