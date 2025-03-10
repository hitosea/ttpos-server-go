package utils

import "strings"

func IsNotFoundRecord(err error) bool {
	return strings.Contains(err.Error(), "record not found")
}
