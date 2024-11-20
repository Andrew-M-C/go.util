package elastic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"slices"

	es "github.com/olivere/elastic/v7"
)

// BoolQuerier 表示简单的 bool 搜索器
type BoolQuerier struct {
	Index string
	errs  []error
	debug func(string, ...any)

	filters []es.Query
	musts   []es.Query

	ranges map[string][]rangeOp
	sorts  map[string]bool

	from, size *int
}

// NewBoolQuerier 新建 BoolQuerier 对象
func NewBoolQuerier(index string) *BoolQuerier {
	return &BoolQuerier{
		Index: index,
	}
}

func emptyDebug(string, ...any) {}

func (q *BoolQuerier) lazyInit() {
	if q.debug == nil {
		q.debug = emptyDebug
	}
	if q.sorts == nil {
		q.sorts = map[string]bool{}
	}
	if q.ranges == nil {
		q.ranges = map[string][]rangeOp{}
	}
}

// Debug 指定调试输出器
func (q *BoolQuerier) Debug(f func(string, ...any)) *BoolQuerier {
	if f != nil {
		q.debug = f
	} else {
		q.debug = emptyDebug
	}
	return q
}

// EQ 表示相等条件。请注意, EQ 不适用于 text 类型字段, 因为 text 类型通常会被分词器处理成多个
// tokens, 而 EQ 使用了 terms, 这是用来精确匹配单个词条的。
func (q *BoolQuerier) EQ(field string, value any) *BoolQuerier {
	q.lazyInit()
	q.filters = append(q.filters, es.NewTermQuery(field, value))
	q.debug("es.NewTermQuery(%v, %v)", field, value)
	return q
}

// In 表示类似 SQL 的 IN 逻辑。其中 values 必须是一个 slice, 但可以是任意类型的 slice
func (q *BoolQuerier) In(field string, values any) *BoolQuerier {
	q.lazyInit()
	if sli, ok := values.([]any); ok {
		q.filters = append(q.filters, es.NewTermsQuery(field, sli...))
		q.debug("es.NewTermsQuery(%v, %v...)", field, sli)
		return q
	}

	v := reflect.ValueOf(values)
	if t := v.Type(); t.Kind() != reflect.Slice {
		q.errs = append(q.errs, fmt.Errorf("in 操作只接受 slice, 请勿使用 '%v' 类型", t))
		return q
	}

	sli := make([]any, 0, v.Len())
	for i := 0; i < v.Len(); i++ {
		elem := v.Index(i)
		sli = append(sli, elem.Interface())
	}
	q.filters = append(q.filters, es.NewTermsQuery(field, sli...))
	q.debug("es.NewTermsQuery(%v, %v...)", field, sli)
	return q
}

// Contains 表示搜索包含某个关键字, 适合 text 类型
func (q *BoolQuerier) Contains(field string, keyword any) *BoolQuerier {
	q.lazyInit()
	q.musts = append(q.musts, es.NewMatchPhraseQuery(field, keyword))
	q.debug("es.NewMatchPhraseQuery(%v, %v)", field, keyword)
	return q
}

// Contains 表示模糊搜索某个关键字, 适合 text 类型
func (q *BoolQuerier) Fuzzy(field string, keyword any) *BoolQuerier {
	q.lazyInit()
	q.musts = append(q.musts, es.NewMatchQuery(field, keyword))
	q.debug("es.NewMatchQuery(%v, %v)", field, keyword)
	return q
}

// SortAsc 升序
func (q *BoolQuerier) SortAsc(field string) *BoolQuerier {
	q.lazyInit()
	q.sorts[field] = true
	return q
}

// SortDesc 降序
func (q *BoolQuerier) SortDesc(field string) *BoolQuerier {
	q.lazyInit()
	q.sorts[field] = false
	return q
}

// From 开始偏移
func (q *BoolQuerier) From(from int) *BoolQuerier {
	q.from = &from
	return q
}

// Limit 限制个数
func (q *BoolQuerier) Limit(limit int) *BoolQuerier {
	q.size = &limit
	return q
}

// Compare 非等比较
func (q *BoolQuerier) Compare(field string, op RangeOperator, target any) *BoolQuerier {
	q.lazyInit()
	q.ranges[field] = append(q.ranges[field], rangeOp{
		op: op,
		v:  target,
	})
	return q
}

// RangeOperator 表示 range 所支持的操作符, 也可以通用字符串直接传参
type RangeOperator string

const (
	LT RangeOperator = "<"
	LE RangeOperator = "<="
	GT RangeOperator = ">"
	GE RangeOperator = ">="
)

type rangeOp struct {
	op RangeOperator
	v  any
}

func (q *BoolQuerier) packQuery() *es.BoolQuery {
	q.lazyInit()

	query := es.NewBoolQuery()
	filters := slices.Clone(q.filters)

	// 打包 range
	for field, conditions := range q.ranges {
		rq := es.NewRangeQuery(field)
		for _, cond := range conditions {
			switch cond.op {
			case LT:
				rq = rq.Lt(cond.v)
			case LE:
				rq = rq.Lte(cond.v)
			case GT:
				rq = rq.Gt(cond.v)
			case GE:
				rq = rq.Gte(cond.v)
			default:
				// do nothing
			}
		}
		filters = append(filters, rq)
	}

	// 打包 filter 和 must
	if len(filters) > 0 {
		query = query.Filter(filters...)
	}
	if len(q.musts) > 0 {
		query = query.Must(q.musts...)
	}

	return query
}

func (q *BoolQuerier) Do(ctx context.Context, cli *es.Client) (*es.SearchResult, error) {
	q.lazyInit()
	if len(q.errs) > 0 {
		return nil, errors.Join(q.errs...)
	}

	query := q.packQuery()

	// 构建搜索
	search := cli.Search().Index(q.Index).Query(query)

	// 排序
	for field, ascending := range q.sorts {
		search = search.Sort(field, ascending)
	}

	// offset / limit
	if q.from != nil {
		search = search.From(*q.from)
	}
	if q.size != nil {
		search = search.Size(*q.size)
	}

	// 执行
	q.debug("Search: %v", q)

	return search.Do(ctx)
}

func (q *BoolQuerier) String() string {
	s := stringer{
		Index: q.Index,
		From:  q.from,
		Size:  q.size,
	}
	if dsl, err := q.packQuery().Source(); err == nil {
		s.DSL.Query = dsl
	}
	for field, ascending := range q.sorts {
		if ascending {
			s.Sort = append(s.Sort, []string{field, "ASC"})
		} else {
			s.Sort = append(s.Sort, []string{field, "DESC"})
		}
	}
	b, _ := json.Marshal(s)
	return string(b)
}

type stringer struct {
	Index string     `json:"index"`
	Sort  [][]string `json:"sort,omitempty"`
	From  *int       `json:"from,omitempty"`
	Size  *int       `json:"size,omitempty"`
	DSL   struct {
		Query any `json:"query"`
	} `json:"dsl"`
}

type jsonWrapper struct {
	v any
}

func (j jsonWrapper) String() string {
	b, err := json.Marshal(j.v)
	if err != nil {
		return fmt.Sprint(j.v)
	}
	return string(b)
}

func toJSON(v any) fmt.Stringer {
	return jsonWrapper{v: v}
}
