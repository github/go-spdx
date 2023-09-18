package spdxexp

// flatten will take an array of nested array and return
// all nested elements in an array. e.g. [[1,2,[3]],4] -> [1,2,3,4]
func flatten[T any](lists [][]T) []T {
	var res []T
	for _, list := range lists {
		res = append(res, list...)
	}
	return res
}

// removeDuplicates will remove all duplicates from a slice
func removeDuplicates[T string](sliceList []T) []T {
	allKeys := make(map[T]bool)
	list := []T{}
	for _, item := range sliceList {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}
