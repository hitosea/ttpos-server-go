package utils

// removeDuplicates 去除字符串切片中的重复元素
func RemoveDuplicates(strSlice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range strSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

// RemoveDuplicateUint64 去除uint64切片中的重复元素
func RemoveDuplicateUint64(uint64Slice []uint64) []uint64 {
	keys := make(map[uint64]bool)
	list := []uint64{}
	for _, entry := range uint64Slice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}
