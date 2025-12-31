package utility

import (
	"github.com/gogf/gf/v2/crypto/gmd5"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
)

// GenerateTtposAuth 生成 TTPOS 认证头
// 使用 MD5(identifier + callbackSecret) 生成认证字符串
// identifier 可以是 shopUUID、shopRefNo 等标识符
func GenerateTtposAuth(identifier string) (string, error) {
	callbackSecret := g.Cfg().MustGet(gctx.GetInitCtx(), "app.callbackSecret").String()
	auth, err := gmd5.EncryptString(identifier + callbackSecret)
	if err != nil {
		return "", gerror.Wrap(err, "failed to generate TTPOS auth")
	}
	return auth, nil
}
