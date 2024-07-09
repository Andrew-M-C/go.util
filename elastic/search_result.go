package elastic

import (
	"encoding/json"

	es "github.com/olivere/elastic/v7"
)

// ESHit 包含 ES 命中之后的外层参数
type ESHit[T any] struct {
	Index  string
	Type   string
	ID     string // doc ID
	Score  float64
	Source T
}

// ParseSearchResult 解析 elastic 搜索结果
func ParseSearchResult[T any](esRes *es.SearchResult) []T {
	if esRes == nil || esRes.Hits == nil {
		return nil
	}

	res := make([]T, 0, len(esRes.Hits.Hits))
	for _, hit := range esRes.Hits.Hits {
		if hit == nil {
			continue
		}
		var item T
		if err := json.Unmarshal(hit.Source, &item); err != nil {
			continue
		}
		res = append(res, item)
	}
	return res
}

// ParseSearchResult 解析 elastic 搜索结果, 同时包含外部的其他额外参数
func ParseWrappedSearchResult[T any](esRes *es.SearchResult) []ESHit[T] {
	if esRes == nil || esRes.Hits == nil {
		return nil
	}

	res := make([]ESHit[T], 0, len(esRes.Hits.Hits))
	for _, hit := range esRes.Hits.Hits {
		if hit == nil {
			continue
		}
		var item ESHit[T]
		if err := json.Unmarshal(hit.Source, &item.Source); err != nil {
			continue
		}
		item.Index = hit.Index
		item.Type = hit.Type
		item.ID = hit.Id
		if hit.Score != nil {
			item.Score = *hit.Score
		}
		res = append(res, item)
	}
	return res
}
