package utils

import (
	"strings"
)

func EllipsisString(s string, max int) string {
	if max > len(s) {
		return s
	}
	return s[:strings.LastIndexAny(s[:max], " .,:;-")] + "..."
}
