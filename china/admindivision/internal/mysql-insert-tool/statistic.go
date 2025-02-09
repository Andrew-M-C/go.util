package main

import "log"

func statisticNameLength(data downloadData) {
	longest := ""

	iterateNode := func(n adminNodeFormat) {
		if len(n.Name) <= len(longest) {
			return
		}
		longest = n.Name
	}
	iterate := func(nodes []adminNodeFormat) {
		for _, n := range nodes {
			iterateNode(n)
		}
	}
	iterate(data.Provinces)
	iterate(data.Cities)
	iterate(data.Counties)
	iterate(data.Towns)
	iterate(data.Villages)

	log.Printf("最长的节点名称有 %d 个字符: %s", runeLen(longest), longest)
}

func runeLen(s string) int {
	count := 0
	for range s {
		count++
	}
	return count
}
