package utils

import "sort"

type SortedEntry[T any] struct {
	Key string
	Val T
}

func SortMapByKey[T any](m map[string]T) []SortedEntry[T] {
	vals := make([]SortedEntry[T], 0, len(m))
	for key, val := range m {
		vals = append(vals, SortedEntry[T]{
			Key: key,
			Val: val,
		})
	}

	sort.Slice(vals, func(i, j int) bool {
		return vals[i].Key < vals[j].Key
	})

	return vals
}
