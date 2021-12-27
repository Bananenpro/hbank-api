package services

import (
	"fmt"
	"strings"
)

func StrToBool(value string) bool {
	return strings.EqualFold(value, "true") || strings.EqualFold(value, "t") ||
		strings.EqualFold(value, "yes") || strings.EqualFold(value, "y") ||
		strings.EqualFold(value, "on") || strings.EqualFold(value, "1")
}

func SizeInBytesToStr(size int64) string {
	if size >= 1000000000 {
		return fmt.Sprintf("%d GB", size/1000000000)
	} else if size >= 1000000 {
		return fmt.Sprintf("%d MB", size/1000000)
	} else if size >= 1000 {
		return fmt.Sprintf("%d kB", size/1000)
	} else {
		return fmt.Sprintf("%d B", size)
	}
}
