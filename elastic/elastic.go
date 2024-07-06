// Package elastic 提供 elastic 搜索工具
package elastic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	es "github.com/olivere/elastic/v7"
)

// BoolQuerier 表示简单的 bool 搜索器
type BoolQuerier[T any] struct {
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
func NewBoolQuerier[T any](index string) *BoolQuerier[T] {
	return &BoolQuerier[T]{
		Index: index,
	}
}

func (q *BoolQuerier[T]) lazyInit() {
	if q.debug == nil {
		q.debug = func(string, ...any) {}
	}
	if q.sorts == nil {
		q.sorts = map[string]bool{}
	}
	if q.ranges == nil {
		q.ranges = map[string][]rangeOp{}
	}
}

func (q *BoolQuerier[T]) Debug(f func(string, ...any)) *BoolQuerier[T] {
	if f != nil {
		q.debug = f
	}
	return q
}

// EQ 表示相等条件。请注意, EQ 不适用于 text 类型字段
func (q *BoolQuerier[T]) EQ(field string, value any) *BoolQuerier[T] {
	q.lazyInit()
	q.filters = append(q.filters, es.NewTermQuery(field, value))
	q.debug("es.NewTermQuery(%v, %v)", field, value)
	return q
}

// In 表示类似 SQL 的 IN 逻辑。其中 values 必须是一个 slice, 但可以是任意类型的 slice
func (q *BoolQuerier[T]) In(field string, values any) *BoolQuerier[T] {
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
func (q *BoolQuerier[T]) Contains(field string, keyword any) *BoolQuerier[T] {
	q.lazyInit()
	q.musts = append(q.musts, es.NewMatchPhraseQuery(field, keyword))
	q.debug("es.NewMatchPhraseQuery(%v, %v)", field, keyword)
	return q
}

// Contains 表示模糊搜索某个关键字, 适合 text 类型
func (q *BoolQuerier[T]) Fuzzy(field string, keyword any) *BoolQuerier[T] {
	q.lazyInit()
	q.musts = append(q.musts, es.NewMatchQuery(field, keyword))
	q.debug("es.NewMatchQuery(%v, %v)", field, keyword)
	return q
}

// SortAsc 升序
func (q *BoolQuerier[T]) SortAsc(field string) *BoolQuerier[T] {
	q.lazyInit()
	q.sorts[field] = true
	return q
}

// SortDesc 降序
func (q *BoolQuerier[T]) SortDesc(field string) *BoolQuerier[T] {
	q.lazyInit()
	q.sorts[field] = false
	return q
}

// From 开始偏移
func (q *BoolQuerier[T]) From(from int) *BoolQuerier[T] {
	q.from = &from
	return q
}

// Limit 限制个数
func (q *BoolQuerier[T]) Limit(limit int) *BoolQuerier[T] {
	q.size = &limit
	return q
}

// Compare 非等比较
func (q *BoolQuerier[T]) Compare(field string, op RangeOperator, target any) *BoolQuerier[T] {
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

// ESHit 包含 ES 命中之后的外层参数
type ESHit[T any] struct {
	Index  string
	Type   string
	ID     string // doc ID
	Score  float64
	Source T
}

// Do 执行搜索
func (q *BoolQuerier[T]) Do(ctx context.Context, cli *es.Client) ([]T, error) {
	esRes, err := q.packAndDo(ctx, cli)
	if err != nil {
		return nil, err
	}

	if f := q.debug; f != nil {
		f("Elastic response: '%s'", toJSON(esRes))
	}
	if esRes == nil || esRes.Hits == nil {
		return nil, errors.New("es hits nil")
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
	return res, nil
}

// DoWrapped 执行搜索, 同时返回每一个 hit 外层的参数
func (q *BoolQuerier[T]) DoWrapped(ctx context.Context, cli *es.Client) ([]ESHit[T], error) {
	esRes, err := q.packAndDo(ctx, cli)
	if err != nil {
		return nil, err
	}

	if f := q.debug; f != nil {
		f("Elastic response: '%s'", toJSON(esRes))
	}
	if esRes == nil || esRes.Hits == nil {
		return nil, errors.New("es hits nil")
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
	return res, nil
}

func (q *BoolQuerier[T]) packAndDo(ctx context.Context, cli *es.Client) (*es.SearchResult, error) {
	q.lazyInit()
	if len(q.errs) > 0 {
		return nil, errors.Join(q.errs...)
	}

	query := es.NewBoolQuery()

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
		q.filters = append(q.filters, rq)
	}

	// 打包 filter 和 must
	if len(q.filters) > 0 {
		query = query.Filter(q.filters...)
	}
	if len(q.musts) > 0 {
		query = query.Must(q.musts...)
	}

	// 构建搜索
	search := cli.Search().Index(q.Index).Query(query)

	// 排序
	for field, direction := range q.sorts {
		search = search.Sort(field, direction)
	}

	// offset / limit
	if q.from != nil {
		search = search.From(*q.from)
	}
	if q.size != nil {
		search = search.Size(*q.size)
	}

	// 执行
	if f := q.debug; f != nil {
		f("Search: %+v", toJSON(search))
	}
	return search.Do(ctx)
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
