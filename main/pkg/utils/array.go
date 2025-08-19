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
