package middleware

import (
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/pkg/context"

	"github.com/gin-gonic/gin"
)

type VersionCheckType string

const (
	TypePurchaseOrder VersionCheckType = "purchase_order"
	TypeTransferOrder VersionCheckType = "transfer_order"
)

// 版本检查
func MinVersionCheck(settingSrv setting.ISrv, typ VersionCheckType, minVersion string) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := helper.GetContext(c)
		if typ == TypePurchaseOrder {
			minVersion = settingSrv.GetShopAppMinVersion()
		}
		if ctx.Version(context.LT, minVersion) {
			helper.Fail(c, constant.CodeAccessDenied, "软件版本过低，请升级软件")
			c.Abort()
			return
		}
		c.Next()
	}
}
