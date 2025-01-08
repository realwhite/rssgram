package utils

import (
	"slices"
)

func MergeStrSlices(a, b []string) []string {
	return slices.Concat(a, b)
}
