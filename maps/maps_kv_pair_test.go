package maps_test

import (
	"testing"

	"github.com/Andrew-M-C/go.util/maps"
)

func testKVPairs(*testing.T) {
	m := map[string]int{
		"1":    1,
		"-1":   -1,
		"1000": 1000,
	}

	kvs := maps.KeyValuesAndSortByKeys(m, maps.Descend)
	so(kvs[0].K, eq, "1000")
	so(kvs[0].V, eq, 1000)
	so(kvs[1].K, eq, "1")
	so(kvs[1].V, eq, 1)
	so(kvs[2].K, eq, "-1")
	so(kvs[2].V, eq, -1)

	kvs = maps.KeyValuesAndSortByValues(m, maps.Ascend)
	so(kvs[0].K, eq, "-1")
	so(kvs[0].V, eq, -1)
	so(kvs[1].K, eq, "1")
	so(kvs[1].V, eq, 1)
	so(kvs[2].K, eq, "1000")
	so(kvs[2].V, eq, 1000)
}
