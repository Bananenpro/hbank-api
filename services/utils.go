package services

import "strings"

func StrToBool(value string) bool {
	return strings.EqualFold(value, "true") || strings.EqualFold(value, "t") ||
		strings.EqualFold(value, "yes") || strings.EqualFold(value, "y") ||
		strings.EqualFold(value, "on")
}
