package utils

import (
	"strings"
)

func EllipsisString(s string, max int) string {
	if max <= 0 {
		return ""
	}
	if max >= len(s) {
		return s
	}
	cut := strings.LastIndexAny(s[:max], " .,:;-")
	if cut <= 0 {
		cut = max
	}
	return s[:cut] + "..."
}
