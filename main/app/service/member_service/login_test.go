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
		Source:      constant.SourceMember,
		CompanyUuid: 7684296282112000,
		MemberUuid:  7717741662208000,
	}
	token, err := auth.GenerateToken(claims, "ttoposerewrwgbngdf", 360000000, false)
	if err != nil {
		t.Error("生成token失败")
	}
	t.Log(token)
}

func TestCallbackToken(t *testing.T) {
	member_sale_order_uuid := "7717741662208000"
	token := cryptor.Md5String(member_sale_order_uuid + "ttoposerewrwgbngdf")
	t.Log(token)
}
