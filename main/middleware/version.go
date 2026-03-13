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
	TypePurchaseOrder VersionCheckType = VersionCheckType(setting.ModulePurchaseOrder)
	TypeTransferOrder VersionCheckType = VersionCheckType(setting.ModuleTransferOrder)
	TypeStatistics    VersionCheckType = VersionCheckType(setting.ModuleStatistics)
)

// MinVersionCheck 版本检查，从配置中动态获取指定模块的最低版本
func MinVersionCheck(settingSrv setting.ISrv, typ VersionCheckType) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := helper.GetContext(c)
		minVersion := settingSrv.GetShopAppMinVersion(string(typ))
		if ctx.Version(context.LT, minVersion) {
			helper.Fail(c, constant.CodeAccessDenied, "软件版本过低，请升级软件")
			c.Abort()
			return
		}
		c.Next()
	}
}
