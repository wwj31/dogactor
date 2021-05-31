//Copyright (c) 2017 Tap4Fun.Co.Ltd. All rights reserved.
package slice

func UniqueSlice(slice *[]interface{}) {
	found := make(map[interface{}]bool)
	total := 0
	for i, val := range *slice {
		if _, ok := found[val]; !ok {
			found[val] = true
			(*slice)[total] = (*slice)[i]
			total++
		}
	}

	*slice = (*slice)[:total]
}

func InsertToSlice(slice, insertion []interface{}, index int) []interface{} {
	result := make([]interface{}, len(slice)+len(insertion))
	at := copy(result, slice[:index])
	at += copy(result[at:], insertion)
	copy(result[at:], slice[index:])
	return result
}

func AppendToSlice(slice, appendItems []interface{}) []interface{} {
	return InsertToSlice(slice, appendItems, len(slice))
}

func ReverseSlice(slice []interface{}) []interface{} {
	if len(slice) == 0 || len(slice) == 1 {
		return slice
	}
	return append(ReverseSlice(slice[1:]), slice[0])
}

func RemoveFromSlice(slice []interface{}, start int, end int) []interface{} {
	return append(slice[:start], slice[end+1:]...)
}

// condition is the function which is used to check if the element shall be removed
func RemoveFromSliceIf(slice []interface{}, condition func(interface{}) bool) []interface{} {
	result := make([]interface{}, len(slice))
	for _, item := range slice {
		// if matched, just ignore
		if condition(item) == false {
			result = append(result, item)
		}
	}
	return result
}
