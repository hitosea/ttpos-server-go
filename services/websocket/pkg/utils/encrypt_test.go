package utils

import "testing"

func TestEncryptPassword(t *testing.T) {
	password := EncryptPassword("123456")
	t.Log(password)
}
