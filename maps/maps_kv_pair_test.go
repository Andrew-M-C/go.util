package maps

import "testing"

func testKVPairs(t *testing.T) {
	m := map[string]int{
		"1":    1,
		"-1":   -1,
		"1000": 1000,
	}

	kvs := KeyValuesAndSortByKeys(m, Descend)
	so(kvs[0].K, eq, "1000")
	so(kvs[0].V, eq, 1000)
	so(kvs[1].K, eq, "1")
	so(kvs[1].V, eq, 1)
	so(kvs[2].K, eq, "-1")
	so(kvs[2].V, eq, -1)

	kvs = KeyValuesAndSortByValues(m, Ascend)
	so(kvs[0].K, eq, "-1")
	so(kvs[0].V, eq, -1)
	so(kvs[1].K, eq, "1")
	so(kvs[1].V, eq, 1)
	so(kvs[2].K, eq, "1000")
	so(kvs[2].V, eq, 1000)
}
