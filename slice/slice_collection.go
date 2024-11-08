package slice

// 本文件定义切片集合操作

// CollectionEqual 将两个切片视为集合, 判断是否相等 (相同元素视为同一个元素)
func CollectionEqual[T comparable](a, b []T) bool {
	aSet := sliceToSet(a)
	for _, key := range b {
		if _, exist := aSet[key]; !exist {
			return false
		}
		delete(aSet, key)
	}
	return len(aSet) == 0
}

func sliceToSet[T comparable](sli []T) map[T]struct{} {
	set := make(map[T]struct{}, len(sli))
	for _, key := range sli {
		set[key] = struct{}{}
	}
	return set
}

// CollectionDifference 将两个切片视为集合计算差集, a - b (相同元素视为同一个元素)
func CollectionDifference[T comparable](a, b []T) []T {
	bSet := sliceToSet(b)
	processedKey := make(map[T]struct{}, len(a))

	var res []T
	for _, key := range a {
		if _, exist := processedKey[key]; exist {
			continue
		}
		if _, exist := bSet[key]; !exist {
			res = append(res, key)
		}
		processedKey[key] = struct{}{}
	}
	return res
}

// CollectionUnion 将两个切片视为集合计算并集, a ⋃ b (相同元素视为同一个元素)
func CollectionUnion[T comparable](a, b []T) []T {
	res := make([]T, 0, len(a))
	aKey := make(map[T]struct{}, len(a))

	for _, key := range a {
		if _, exist := aKey[key]; exist {
			continue
		}
		res = append(res, key)
		aKey[key] = struct{}{}
	}

	for _, key := range b {
		if _, exist := aKey[key]; exist {
			continue
		}
		res = append(res, key)
		aKey[key] = struct{}{}
	}
	return res
}

// CollectionIntersection 将两个切片视为集合计算交集, a ∩ b (相同元素视为同一个元素)
func CollectionIntersection[T comparable](a, b []T) []T {
	var res []T
	bKey := sliceToSet(b)
	processed := make(map[T]struct{}, len(a))

	for _, key := range a {
		if _, exist := processed[key]; exist {
			continue
		}
		if _, exist := bKey[key]; exist {
			res = append(res, key)
		}
		processed[key] = struct{}{}
	}

	return res
}

// CollectionSymmetricDifference 将两个切片视为集合计算对成交集, a △ b (相同元素视为同一个元素)
func CollectionSymmetricDifference[T comparable](a, b []T) []T {
	var res []T
	bKey := sliceToSet(b)
	processed := make(map[T]struct{}, len(a))

	for _, key := range a {
		if _, exist := processed[key]; exist {
			continue
		}
		if _, exist := bKey[key]; !exist {
			res = append(res, key)
		}
		processed[key] = struct{}{}
	}
	for _, key := range b {
		if _, exist := processed[key]; exist {
			continue
		}
		res = append(res, key)
		processed[key] = struct{}{}
	}

	return res
}
