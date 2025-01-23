package auth

import (
	"testing"
	"ttpos-server-go/app/constant"
)

func TestGenerateToken(t *testing.T) {
	token, err := GenerateToken(constant.SOURCE_CASHIER, 1724054154, 1111000, "2323423414", 3600)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(token)
}
