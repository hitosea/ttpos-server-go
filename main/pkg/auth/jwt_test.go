package auth

import (
	"testing"
	"ttpos-server-go/app/constant"
)

func TestGenerateToken(t *testing.T) {
	token, err := GenerateToken(constant.SourceCashier, "device_id", 1724054154, 1111000, "ttpos_kitchen", 86400)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(token)
}
