package utils

// removeDuplicates 去除字符串切片中的重复元素
func RemoveDuplicates[T comparable](strSlice []T) []T {
	keys := make(map[T]bool)
	list := []T{}
	for _, entry := range strSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

// 获取两个切片中的重复元素
func GetDuplicateElements[T comparable](slice1 []T, slice2 []T) []T {
	keys := make(map[T]bool)
	list := []T{}
	for _, entry := range slice1 {
		keys[entry] = true
	}
	for _, entry := range slice2 {
		if _, value := keys[entry]; value {
			list = append(list, entry)
		}
	}
	return list
}

// Filter 通用切片过滤方法
func Filter[T any](slice []T, predicate func(T) bool) []T {
	result := make([]T, 0)
	for _, item := range slice {
		if predicate(item) {
			result = append(result, item)
		}
	}
	return result
}

// FilterByContains 基于包含关系过滤切片
func FilterByContains[T any, K comparable](slice []T, targetList []K, getKey func(T) K) []T {
	return Filter(slice, func(item T) bool {
		return Contains(targetList, getKey(item))
	})
}

// Contains 检查切片是否包含指定元素
func Contains[T comparable](slice []T, target T) bool {
	for _, item := range slice {
		if item == target {
			return true
		}
	}
	return false
}

// RemoveDuplicatesByKey 基于指定的 key 提取函数去除切片中的重复元素
// 保留第一次出现的元素
func RemoveDuplicatesByKey[T any, K comparable](slice []T, getKey func(T) K) []T {
	if len(slice) == 0 {
		return slice
	}

	seen := make(map[K]bool)
	result := make([]T, 0, len(slice))

	for _, item := range slice {
		key := getKey(item)
		if !seen[key] {
			seen[key] = true
			result = append(result, item)
		}
	}

	return result
}
