package utils

import (
	"strings"
)

func EllipsisString(s string, max int) string {
	if max > len(s) {
		return s
	}
	idx := strings.LastIndexAny(s[:max], " .,:;-")
	if idx == -1 {
		return s[:max] + "..."
	}
	return s[:idx] + "..."
}
