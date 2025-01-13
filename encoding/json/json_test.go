package json_test

import (
	"github.com/Andrew-M-C/go.util/encoding/json"
	"github.com/smartystreets/goconvey/convey"
	"golang.org/x/exp/constraints"
)

var (
	cv = convey.Convey
	so = convey.So
	eq = convey.ShouldEqual

	isNil = convey.ShouldBeNil
)

type jsonOrdered interface {
	constraints.Integer | ~string
}

type collection[T jsonOrdered] map[T]struct{}

func (c collection[T]) Set(k T) {
	c[k] = struct{}{}
}

func (c collection[T]) Has(k T) bool {
	_, exist := c[k]
	return exist
}

func (c collection[T]) MarshalJSON() ([]byte, error) {
	return json.MarshalMapKeyToArray(c)
}

func (c *collection[T]) UnmarshalJSON(b []byte) error {
	return json.UnmarshalArrayToCollection(b, c)
}
