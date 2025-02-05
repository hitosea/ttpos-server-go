package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/auth"
)

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			helper.Fail(c, constant.CodeBadRequest, "token 为空")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			helper.Fail(c, constant.CodeBadRequest, "token 格式错误")
			c.Abort()
			return
		}

		// 验证token
		claims, err := auth.ParseToken(parts[1], config.JWT.Secret)
		if err != nil {
			helper.Fail(c, constant.CodeBadRequest, "无效的token")
			c.Abort()
			return
		}

		// 将用户信息存储到上下文
		c.Set("company_id", claims.CompanyId)
		c.Set("staff_id", claims.StaffId)
		c.Set("source", claims.Source)
		c.Next()
	}
}
