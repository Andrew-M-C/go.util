package main

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/Andrew-M-C/go.util/net/http"
	"github.com/Andrew-M-C/go.util/slice"
)

const (
	sourceProvinceURL = "https://github.com/modood/Administrative-divisions-of-China/raw/master/dist/provinces.json"
	sourceCityURL     = "https://github.com/modood/Administrative-divisions-of-China/raw/master/dist/cities.json"
	sourceCountyURL   = "https://github.com/modood/Administrative-divisions-of-China/raw/master/dist/areas.json"
	sourceTownURL     = "https://github.com/modood/Administrative-divisions-of-China/blob/master/dist/streets.json?raw=true"
	sourceVillageURL  = "https://github.com/modood/Administrative-divisions-of-China/blob/master/dist/villages.json?raw=true"
)

// node 表示一层节点
type node struct {
	code     string
	fullCode string
	name     string
	virtual  bool
	sub      []*node
	history  bool
}

func searchSubNode(sub []*node, code string) *node {
	// 二分查找
	target := &node{
		code: code,
	}
	i := slice.BinarySearchOne(sub, target, func(a, b *node) int {
		return strings.Compare(a.code, b.code)
	})
	if i < 0 {
		return nil
	}
	return sub[i]
}

func (sess *session) searchNode(code string) (*node, error) {
	return searchNode(sess.provinces, code)
}

func searchNode(provinces []*node, code string) (*node, error) {
	// 省级
	province := searchSubNode(provinces, code[0:2])
	if province == nil {
		return nil, fmt.Errorf("查找 %s 失败: 无法找到省级代码 %s", code, code[0:2])
	}
	if len(code) == 2 {
		return province, nil
	}

	// 市级
	city := searchSubNode(province.sub, code[2:4])
	if city == nil {
		return nil, fmt.Errorf("查找 %s 失败: 无法找到市级代码 %s", code, code[2:4])
	}
	if len(code) == 4 {
		return city, nil
	}

	// 区县级
	county := searchSubNode(city.sub, code[4:6])
	if county == nil {
		return nil, fmt.Errorf("查找 %s 失败: 无法找到区县级代码 %s", code, code[4:6])
	}
	if len(code) == 6 {
		return county, nil
	}

	// 镇级
	town := searchSubNode(county.sub, code[6:9])
	if town == nil {
		return nil, fmt.Errorf("查找 %s 失败: 无法找到镇 / 居委会级代码 %s", code, code[6:9])
	}
	if len(code) == 9 {
		return town, nil
	}

	// 村街道级
	village := searchSubNode(town.sub, code[9:])
	if village == nil {
		return nil, fmt.Errorf("查找 %s 失败: 无法找到村 / 街道级代码 %s", code, code[9:])
	}
	return village, nil
}

type session struct {
	provinces []*node
}

type rawNode struct {
	Code         string `json:"code"` // 完整代码
	Name         string `json:"name"`
	ProvinceCode string `json:"provinceCode"`
	CityCode     string `json:"cityCode"`
	CountyCode   string `json:"areaCode"`
	VillageCode  string `json:"streetCode"`
}

func sortNodes(nodes []*node) {
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].code < nodes[j].code
	})
}

func (sess *session) getAndParseProvinces() error {
	printf("开始解析省级行政区")
	ctx := context.Background()
	nodes, err := http.JSON[[]rawNode](ctx, sourceProvinceURL)
	if err != nil {
		return err
	}

	for _, n := range *nodes {
		node := &node{
			code:     n.Code,
			fullCode: n.Code,
			name:     n.Name,
			virtual:  strings.Contains(n.Name, "市"),
		}
		sess.provinces = append(sess.provinces, node)
	}

	// 然后添加港澳台
	tw := &node{
		code:     "71",
		fullCode: "71",
		name:     "台湾省",
		virtual:  true,
	}
	hk := &node{
		code:     "81",
		fullCode: "81",
		name:     "香港特别行政区",
		virtual:  true,
	}
	macau := &node{
		code:     "82",
		fullCode: "82",
		name:     "澳门特别行政区",
		virtual:  true,
	}
	sess.provinces = append(sess.provinces, tw, hk, macau)
	sortNodes(sess.provinces)

	return nil
}

func isCityVirtual(n string) bool {
	switch n {
	case "市辖区", "省直辖县级行政区划":
		return true
	default:
		return false
	}
}

func (sess *session) getAndParseCities() error {
	printf("开始解析市级行政区")
	ctx := context.Background()
	nodes, err := http.JSON[[]rawNode](ctx, sourceCityURL)
	if err != nil {
		return err
	}

	for _, n := range *nodes {
		node := &node{
			code:     n.Code[2:],
			fullCode: n.Code,
			name:     n.Name,
			virtual:  isCityVirtual(n.Name),
		}
		province, err := sess.searchNode(n.Code[:2])
		if err != nil {
			return fmt.Errorf("操作 %v 失败 (%w)", n.Code, err)
		}
		province.sub = append(province.sub, node)
	}

	return nil
}

func (sess *session) getAndParseCounties() error {
	printf("开始解析区县级行政区")
	ctx := context.Background()
	nodes, err := http.JSON[[]rawNode](ctx, sourceCountyURL)
	if err != nil {
		return err
	}

	for _, n := range *nodes {
		node := &node{
			code:     n.Code[4:],
			fullCode: n.Code,
			name:     n.Name,
			virtual:  strings.HasSuffix(n.Code, "00"),
		}
		city, err := sess.searchNode(n.Code[:4])
		if err != nil {
			return fmt.Errorf("操作 %v 失败 (%w)", n.Code, err)
		}
		city.sub = append(city.sub, node)
	}

	return nil
}

func (sess *session) getAndParseTowns() error {
	printf("开始解析镇级行政区")
	ctx := context.Background()
	nodes, err := http.JSON[[]rawNode](ctx, sourceTownURL)
	if err != nil {
		return err
	}

	for _, n := range *nodes {
		node := &node{
			code:     n.Code[6:],
			fullCode: n.Code,
			name:     n.Name,
			virtual:  false,
		}
		county, err := sess.searchNode(n.Code[:6])
		if err != nil {
			return fmt.Errorf("操作 %v 失败 (%w)", n.Code, err)
		}
		county.sub = append(county.sub, node)
	}

	return nil
}

func (sess *session) getAndParseVillages() error {
	printf("开始解析村级行政区")
	ctx := context.Background()
	nodes, err := http.JSON[[]rawNode](ctx, sourceVillageURL)
	if err != nil {
		return err
	}

	for _, n := range *nodes {
		node := &node{
			code:     n.Code[9:],
			fullCode: n.Code,
			name:     n.Name,
			virtual:  false,
		}
		town, err := sess.searchNode(n.Code[:9])
		if err != nil {
			return fmt.Errorf("操作 %v 失败 (%w)", n.Code, err)
		}
		town.sub = append(town.sub, node)
	}

	return nil
}

func (sess *session) writeToGoFile() error {
	const fileName = "../../init.go"
	const funcName = "init"

	printf("开始写入文件")
	f, err := os.OpenFile(fileName, os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("打开文件失败 (%w)", err)
	}
	defer f.Close()

	write := func(format string, args ...any) {
		line := fmt.Sprintf(format+"\n", args...)
		_, _ = f.WriteString(line)
	}

	levelToName := []string{
		"",
		"Province",
		"City",
		"County",
		"Town",
		"Village",
	}

	var writeNode func(indent int, isFirst bool, level int, n *node)
	writeNode = func(indent int, isFirst bool, level int, n *node) {
		prefix := strings.Repeat("\t", indent)

		if !isFirst {
			write(prefix + `},{`)
		}
		write(prefix+`	level: %s,`, levelToName[level])
		write(prefix+`	code: "%s",`, n.code)
		write(prefix+`	fullCode: "%s",`, n.fullCode)
		write(prefix+`	name: "%s",`, n.name)

		if n.virtual {
			write(prefix + `	virtual: true,`)
		}
		if n.history {
			write(prefix + `	deprecated: true,`)
		}
		if len(n.sub) == 0 {
			return
		}

		write(prefix + `	sub: []*Division{{`)
		for i, sub := range n.sub {
			writeNode(indent+1, i == 0, level+1, sub)
		}
		write(prefix + `	}},`)
	}

	write(`// Code generated by sync-tool. DO NOT EDIT.`)
	write(`// Time: %s`, time.Now().Format(time.DateTime))
	write("")
	write(`package admindivision`)
	write("")
	write(`func %s() {`, funcName)

	write(`	china.sub = []*Division{{`)
	for i, n := range sess.provinces {
		writeNode(2, i == 0, 1, n)
	}
	write(`	}}`)

	write(`}`)
	return nil
}
