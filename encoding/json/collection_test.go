package json_test

import (
	js "encoding/json"
	"testing"

	"github.com/Andrew-M-C/go.util/encoding/json"
)

func TestMarshalMapKeyToArray(t *testing.T) {
	cv("MarshalMapKeyToArray", t, func() {
		testMarshalMapKeyToArray(t)
	})
}

func testMarshalMapKeyToArray(*testing.T) {
	cv("string-int", func() {
		s := map[string]int{
			"a": 1,
			"b": 2,
			"c": 3,
		}
		b, err := json.MarshalMapKeyToArray(s)
		so(err, isNil)
		so(string(b), eq, `["a","b","c"]`)
	})
	cv("string-bool", func() {
		s := map[string]bool{
			"a": true,
			"b": false,
			"c": true,
		}
		b, err := json.MarshalMapKeyToArray(s)
		so(err, isNil)
		so(string(b), eq, `["a","b","c"]`)
	})
	cv("string-struct", func() {
		s := map[string]bool{
			"a": true,
			"b": false,
			"c": true,
		}
		b, err := json.MarshalMapKeyToArray(s)
		so(err, isNil)
		so(string(b), eq, `["a","b","c"]`)
	})
	cv("typed map", func() {
		c := collection[uint]{}
		c.Set(11)
		c.Set(22)
		c.Set(33)
		b, err := json.MarshalMapKeyToArray(c)
		so(err, isNil)
		so(string(b), eq, `[11,22,33]`)
	})
}

func TestMarshalBoolMapKeyToArray(t *testing.T) {
	cv("MarshalBoolMapKeyToArray", t, func() {
		testMarshalBoolMapKeyToArray(t)
	})
}

func testMarshalBoolMapKeyToArray(*testing.T) {
	cv("string-bool", func() {
		s := map[string]bool{
			"a": true,
			"b": false,
			"c": true,
		}
		b, err := json.MarshalBoolMapKeyToArray(s)
		so(err, isNil)
		so(string(b), eq, `["a","c"]`)
	})
}

func TestUnmarshalArrayToCollection(t *testing.T) {
	cv("UnmarshalArrayToCollection", t, func() {
		testUnmarshalArrayToCollection(t)
	})
}

func testUnmarshalArrayToCollection(*testing.T) {
	cv("string-struct", func() {
		in := `["a","b","c"]`
		var out map[string]struct{}
		err := json.UnmarshalArrayToCollection([]byte(in), &out)
		so(err, isNil)
		so(len(out), eq, 3)
		for _, k := range []string{"a", "b", "c"} {
			_, exist := out[k]
			so(exist, eq, true)
		}
	})

	cv("uint64-collection", func() {
		in := `[4,5,6]`
		var out collection[uint64]
		err := json.UnmarshalArrayToCollection([]byte(in), &out)
		so(err, isNil)
		so(len(out), eq, 3)
		for _, k := range []uint64{4, 5, 6} {
			_, exist := out[k]
			so(exist, eq, true)
		}
	})
}

func TestUnmarshalArrayToBoolMap(t *testing.T) {
	cv("UnmarshalArrayToBoolMap", t, func() {
		testUnmarshalArrayToBoolMap(t)
	})
}

func testUnmarshalArrayToBoolMap(*testing.T) {
	cv("string-bool", func() {
		in := `["a","b","c"]`
		var out map[string]bool
		err := json.UnmarshalArrayToBoolMap([]byte(in), &out)
		so(err, isNil)
		so(len(out), eq, 3)
		for _, k := range []string{"a", "b", "c"} {
			b, exist := out[k]
			so(exist, eq, true)
			so(b, eq, true)
		}
	})
}

func TestCollectionGo(t *testing.T) {
	cv("collection.go", t, func() {
		testCollectionGo(t)
	})
}

func testCollectionGo(*testing.T) {
	cv("encoding/json", func() {
		m := collection[int]{}
		m.Set(1)
		m.Set(2)
		m.Set(3)

		b, err := js.Marshal(m)
		so(err, isNil)
		so(string(b), eq, "[1,2,3]")

		var newM collection[int]
		err = js.Unmarshal(b, &newM)
		so(err, isNil)
		so(len(newM), eq, 3)
		so(newM.Has(1), eq, true)
		so(newM.Has(2), eq, true)
		so(newM.Has(3), eq, true)
	})
}
