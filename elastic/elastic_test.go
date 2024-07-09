package elastic_test

import (
	"context"
	"os"
	"testing"

	esutil "github.com/Andrew-M-C/go.util/elastic"
	"github.com/olivere/elastic/v7"
	"github.com/smartystreets/goconvey/convey"
)

func TestMain(m *testing.M) {
	initializeTesting()
	os.Exit(m.Run())
}

var (
	cv = convey.Convey
	so = convey.So
	eq = convey.ShouldEqual
	gt = convey.ShouldBeGreaterThan

	isNil = convey.ShouldBeNil

	globalCli *elastic.Client
)

const (
	testIndex = "test_index"
)

func initializeTesting() {
	cli, err := elastic.NewClient(
		elastic.SetURL("http://localhost:9200"),
		elastic.SetSniff(false),
	)
	if err != nil {
		panic(err)
	}
	globalCli = cli
}

type PaperDocument struct {
	Title        string `json:"title"`
	CreateTSMsec int64  `json:"create_ts_msec"`
	Author       string `json:"author"`
	Content      string `json:"content"`
}

/*
创建索引:

curl -X PUT "http://localhost:9200/test_index" -H 'Content-Type: application/json' -d'
{
  "settings": {
    "analysis": {
      "analyzer": {
        "ik_content_analyzer": {
          "type": "custom",
          "tokenizer": "ik_smart"
        }
      }
    }
  },
  "mappings": {
    "properties": {
      "title": {
        "type": "text",
        "analyzer": "ik_content_analyzer"
      },
      "create_ts_msec": {
        "type": "long"
      },
      "author": {
        "type": "text",
        "analyzer": "ik_content_analyzer"
      },
      "content": {
        "type": "text",
        "analyzer": "ik_content_analyzer"
      }
    }
  }
}
'

查询索引:

curl -X GET "http://localhost:9200/test_index/_mappings?pretty=true"

插入两个数据

curl -X POST "http://localhost:9200/test_index/_doc" -H 'Content-Type: application/json' -d'
{
  "title": "示例文章01",
  "create_ts_msec": 1633087521000,
  "author": "张三",
  "content": "这是一篇关于 Golang 的示例文章。"
}
'

curl -X POST "http://localhost:9200/test_index/_doc" -H 'Content-Type: application/json' -d'
{
  "title": "示例文章02",
  "create_ts_msec": 1720190934000,
  "author": "李四",
  "content": "这是一篇关于 Elastic 的示例文章。"
}
'

查询全部:

curl -X GET "http://localhost:9200/test_index/_search?pretty=true" -H 'Content-Type: application/json'

查询一条

curl -X GET "localhost:9200/test_index/_search?pretty=true" -H 'Content-Type: application/json' -d'
{
  "query": {
    "bool": {
      "filter": [
        {
          "term": {
            "title": "示例文章02"
          }
        }
      ]
    }
  }
}
'

*/

func TestBoolQuerier(t *testing.T) {
	cv("ParseSearchResult", t, func() {
		ctx := context.Background()
		cli := globalCli

		// EQ 不适合 text 类型字段
		q := esutil.NewBoolQuerier(testIndex)
		esRes, err := q.Debug(t.Logf).EQ("title", "示例文章02").Do(ctx, cli)
		res := esutil.ParseSearchResult[PaperDocument](esRes)
		so(err, isNil)
		so(len(res), eq, 0)

		// EQ 适合 keyword 或者是其他精确字段
		q = esutil.NewBoolQuerier(testIndex)
		esRes, err = q.EQ("create_ts_msec", 1720190934000).Do(ctx, cli)
		res = esutil.ParseSearchResult[PaperDocument](esRes)
		so(err, isNil)
		so(len(res), eq, 1)
		t.Log(res)

		// 范围搜索
		q = esutil.NewBoolQuerier(testIndex)
		esRes, err = q.Compare("create_ts_msec", ">", 1633087521000).Do(ctx, cli)
		res = esutil.ParseSearchResult[PaperDocument](esRes)
		so(err, isNil)
		so(len(res), eq, 1)
		t.Log(res)

		// 全量搜索
		q = esutil.NewBoolQuerier(testIndex)
		esRes, err = q.Do(ctx, cli)
		res = esutil.ParseSearchResult[PaperDocument](esRes)
		so(err, isNil)
		so(len(res), eq, 2)
		t.Log(res)

		// offset, limit
		q = esutil.NewBoolQuerier(testIndex)
		esRes, err = q.From(0).Limit(1).SortAsc("create_ts_msec").Do(ctx, cli)
		res = esutil.ParseSearchResult[PaperDocument](esRes)
		so(err, isNil)
		so(len(res), eq, 1)
		so(res[0].Title, eq, "示例文章01")

		q = esutil.NewBoolQuerier(testIndex)
		esRes, err = q.From(1).Limit(1).SortAsc("create_ts_msec").Do(ctx, cli)
		res = esutil.ParseSearchResult[PaperDocument](esRes)
		so(err, isNil)
		so(len(res), eq, 1)
		so(res[0].Title, eq, "示例文章02")
	})

	cv("ParseWrappedSearchResult", t, func() {
		ctx := context.Background()
		cli := globalCli

		// 包含式的搜索
		q := esutil.NewBoolQuerier(testIndex)
		esRes, err := q.Contains("content", "这是一篇关于 Golang").Do(ctx, cli)
		res := esutil.ParseWrappedSearchResult[PaperDocument](esRes)
		so(err, isNil)
		so(len(res), eq, 1)
		t.Log(res)

		// 模糊搜索
		q = esutil.NewBoolQuerier(testIndex)
		esRes, err = q.Fuzzy("content", "这是一篇关于 Elastic").SortDesc("_score").Do(ctx, cli)
		res = esutil.ParseWrappedSearchResult[PaperDocument](esRes)
		so(err, isNil)
		so(len(res), eq, 2)
		so(res[0].Source.Title, eq, "示例文章02")
		so(res[1].Source.Title, eq, "示例文章01")
		so(res[0].Score, gt, res[1].Score)
		t.Log(res)
	})
}
