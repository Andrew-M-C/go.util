package slice

import (
	"fmt"
	"reflect"
)

func CombineEvenly(s1, s2 interface{}) (interface{}, error) {
	sliceT, err := checkSliceSameType(s1, s2)
	if err != nil {
		return nil, err
	}

	v1 := reflect.ValueOf(s1)
	v2 := reflect.ValueOf(s2)

	if v1.Len() < v2.Len() {
		v1, v2 = v2, v1
	}
	if v1.Len() == 0 {
		return reflect.MakeSlice(sliceT, 0, 0).Interface(), nil
	}
	if v1.Len() == 1 {
		out := reflect.MakeSlice(sliceT, 1, 2)
		out.Index(0).Set(v1.Index(0))
		if v2.Len() == 1 {
			out = reflect.Append(out, v2.Index(0))
		}
		return out.Interface(), nil
	}

	total := v1.Len() + v2.Len()
	inserted := make([]bool, total)
	out := reflect.MakeSlice(sliceT, total, total)

	// 由于 lenA >= lenB，因此第一个位置必然是 A。
	// 首先计算出 A 插入位置的步长
	step := float64(v1.Len()+v2.Len()-1) / float64(v1.Len()-1)

	// 第一个位置必然是 A
	out.Index(0).Set(v1.Index(0))
	inserted[0] = true

	// 后续位置按照步长插入
	next := step
	for i := 1; i < v1.Len(); i++ {
		pos := round64(next)
		if pos >= total {
			break
		}
		out.Index(pos).Set(v1.Index(i))
		inserted[pos] = true
		next += step
	}

	// 剩余位置用 B 插入
	v2Index := 0
	for i, notNil := range inserted {
		if notNil {
			continue
		}
		if v2Index > v2.Len()-1 {
			break
		}
		out.Index(i).Set(v2.Index(v2Index))
		v2Index++
	}

	return out.Interface(), nil
}

func checkSliceSameType(s1, s2 interface{}) (sliceType reflect.Type, err error) {
	t1 := reflect.TypeOf(s1)
	t2 := reflect.TypeOf(s2)

	if t1.Kind() != reflect.Slice {
		err = fmt.Errorf("illegal type for first parameter: %v, should be a slice", t1)
		return
	}
	if t2.Kind() != reflect.Slice {
		err = fmt.Errorf("illegal type for second parameter: %v, should be a slice", t2)
		return
	}

	elem1 := t1.Elem()
	elem2 := t2.Elem()
	if elem1.String() != elem2.String() {
		err = fmt.Errorf("elemens of two parameters are not exactly the same, first: %v, second: %v", elem1, elem2)
		return
	}

	return t1, nil
}

func round64(f float64) int {
	f += 0.5
	return int(f)
}
