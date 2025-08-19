package member_service

import (
	"testing"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/pkg/auth"

	"github.com/duke-git/lancet/cryptor"
)

func TestLogin(t *testing.T) {
	// 生成token
	claims := auth.Claims{
		Source:      constant.SourceShop,
		CompanyUuid: 6593110384640000,
		StaffUuid:   6593571758080000,
	}
	token, err := auth.GenerateToken(claims, "dkjhd00a08", 360000000, false)
	if err != nil {
		t.Error("生成token失败")
	}
	t.Log(token)
}

func TestCallbackToken(t *testing.T) {
	member_sale_order_uuid := "3676433861967873"
	token := cryptor.Md5String(member_sale_order_uuid + "dkjhd00a08")
	t.Log(token)
}
