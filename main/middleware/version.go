package middleware

import (
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/pkg/context"

	"github.com/gin-gonic/gin"
)

// 版本检查
func MinVersionCheck(minVersion string) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := helper.GetContext(c)
		if ctx.Version(context.LT, minVersion) {
			helper.Fail(c, constant.CodeVersionError, "软件版本过低，请升级软件")
			c.Abort()
			return
		}
		c.Next()
	}
}
