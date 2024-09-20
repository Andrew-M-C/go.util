package main

import (
	"context"
	"fmt"
	"strings"

	jsonvalue "github.com/Andrew-M-C/go.jsonvalue"
	"github.com/Andrew-M-C/go.util/net/http"
)

const (
	// Reference: [中国行政区划数据](https://passer-by.com/data_location/)
	sourceHistoryURL = "https://passer-by.com/data_location/history.json"
)

func (sess *session) getHistoryNodes() (outSideErr error) {
	printf("开始解析行政区划历史数据")
	ctx := context.Background()
	nodes, err := http.JSON[jsonvalue.V](ctx, sourceHistoryURL)
	if err != nil {
		return err
	}

	printf("总共解析到 %d 行历史行政区划数据", nodes.Len())

	// 解析每一个节点
	nodes.RangeObjectsBySetSequence(func(k string, v *jsonvalue.V) bool {
		switch {
		case strings.HasSuffix(k, "0000"):
			printf("省级 %v - %v", k, v)
			// 省级行政区
			if _, err := sess.searchNode(k[:2]); err == nil {
				return true // OK
			}
			sess.provinces = append(sess.provinces, &node{
				code:     k[:2],
				fullCode: k,
				name:     v.String(),
				history:  true,
			})
			sortNodes(sess.provinces)
			return true

		case strings.HasSuffix(k, "00"):
			// 市级行政区
			printf("市级 %v - %v", k, v)
			if _, err := sess.searchNode(k[:4]); err == nil {
				return true
			}
			node := &node{
				code:     k[2:4],
				fullCode: k,
				name:     v.String(),
				virtual:  isCityVirtual(v.String()),
				history:  true,
			}
			province, err := searchNode(sess.provinces, k[:2])
			if err != nil {
				outSideErr = fmt.Errorf("操作 %v 失败 (%w)", k, err)
				return false
			}
			province.sub = append(province.sub, node)
			sortNodes(province.sub)
			return true

		default:
			// 区县级行政区
			node := &node{
				code:     k[4:],
				fullCode: k,
				name:     v.String(),
				virtual:  false,
				history:  true,
			}
			city, err := searchNode(sess.provinces, k[:4])
			if err != nil {
				if strings.HasPrefix(k, "1102") || strings.HasPrefix(k, "1202") {
					city, err = addMunicipality(sess.provinces, k, "市辖区")
				} else if strings.HasPrefix(k, "3102") {
					city, err = addMunicipality(sess.provinces, k, "上海县")
				} else if k[2:4] == "90" {
					city, err = addProvinceMunicipality(sess.provinces, k)
				} else if k[:4] == "4229" {
					city, err = addHubeiForest(sess.provinces)
				} else if k[:4] == "4230" {
					city, err = addHubeiChenzhou(sess.provinces)
				} else if k[2:4] == "00" && k[4:6] != "00" {
					city, err = addProvinceMunicipalityCity(sess.provinces, k)
				} else if k[:1] == "8" {
					// 非中国内地
					return true
				}
			}
			if err != nil {
				outSideErr = fmt.Errorf("操作 %v 失败 (%w)", k, err)
				return false
			}
			city.sub = append(city.sub, node)
			return true
		}
	})

	return outSideErr
}

// 直辖市
func addMunicipality(provinces []*node, code string, name string) (*node, error) {
	node := &node{
		code:     code[2:],
		fullCode: code[:4],
		name:     name,
		virtual:  true,
		history:  true,
	}
	province, err := searchNode(provinces, code[:2])
	if err != nil {
		return nil, fmt.Errorf("操作直辖市市辖区 %s 失败 (%w)", code, err)
	}
	province.sub = append(province.sub, node)
	sortNodes(provinces)
	return node, nil
}

// 省直辖县级市, 海南有
func addProvinceMunicipalityCity(provinces []*node, code string) (*node, error) {
	node := &node{
		code:     code[2:],
		fullCode: code[:4],
		name:     "省直辖县级行政区划",
		virtual:  true,
		history:  true,
	}
	province, err := searchNode(provinces, code[:2])
	if err != nil {
		return nil, fmt.Errorf("操作直辖县级行政区划 %s 失败 (%w)", code, err)
	}
	province.sub = append(province.sub, node)
	sortNodes(provinces)
	return node, nil
}

// 省直辖县
func addProvinceMunicipality(provinces []*node, code string) (*node, error) {
	node := &node{
		code:     "90",
		fullCode: code[:4],
		name:     "省直辖县级行政区划",
		virtual:  true,
		history:  true,
	}
	province, err := searchNode(provinces, code[:2])
	if err != nil {
		return nil, fmt.Errorf("操作直辖县行政区划 %s 失败 (%w)", code, err)
	}
	province.sub = append(province.sub, node)
	sortNodes(provinces)
	return node, nil
}

// 湖北省林区
func addHubeiForest(provinces []*node) (*node, error) {
	code := "4229"
	node := &node{
		code:     "29",
		fullCode: code,
		name:     "林区",
		virtual:  true,
		history:  true,
	}
	province, err := searchNode(provinces, code[:2])
	if err != nil {
		return nil, fmt.Errorf("操作湖北省林区 %s 失败 (%w)", code, err)
	}
	province.sub = append(province.sub, node)
	sortNodes(provinces)
	return node, nil
}

// 历史上的郴州市
func addHubeiChenzhou(provinces []*node) (*node, error) {
	code := "4230"
	node := &node{
		code:     "30",
		fullCode: code,
		name:     "郴州市",
		history:  true,
	}
	province, err := searchNode(provinces, code[:2])
	if err != nil {
		return nil, fmt.Errorf("操作湖北省郴州市 %s 失败 (%w)", code, err)
	}
	province.sub = append(province.sub, node)
	sortNodes(provinces)
	return node, nil
}
