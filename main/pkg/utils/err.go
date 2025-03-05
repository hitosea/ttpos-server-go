package utils

import "strings"

func Panic(err error) {
	if strings.Contains(err.Error(), "Error 1054 (42S22)") {
		panic(err.Error())
	}
}
