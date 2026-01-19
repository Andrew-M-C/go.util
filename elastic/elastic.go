// Package elastic 定义 ES 搜索相关的工具
package elastic

import es "github.com/olivere/elastic/v7"

// ESClient 定义 ES 客户端的接口, 方便进行 mock
type ESClient interface {
	Search(indices ...string) *es.SearchService
}
