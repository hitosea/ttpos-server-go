package middleware

import (
	"github.com/gin-gonic/gin"
	"strings"
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/constant/jwt"
	"ttpos-server-go/app/service"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/auth"
)

func Auth(authSrv service.IAuthSrv) gin.HandlerFunc {
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

		if claims.Source != constant.SourceCashier {
			helper.Fail(c, constant.CodeBadRequest, "用户信息错误")
			c.Abort()
			return
		}

		// 认证用户
		company, companySetting, staff, err := authSrv.AuthenticateStaff(claims.Source, claims.DeviceId, claims.CompanyId, claims.StaffId, c.Request.URL.Path)
		if err != nil {
			helper.ErrorWithDetail(c, constant.CodeBadRequest, err)
			c.Abort()
			return
		}

		// 将用户信息存储到上下文
		c.Set(jwt.Staff, staff)
		c.Set(jwt.Company, company)
		c.Set(jwt.CompanySetting, companySetting)
		c.Set(jwt.CompanyId, claims.CompanyId)
		c.Set(jwt.StaffId, claims.StaffId)
		c.Set(jwt.Source, claims.Source)
		c.Set(jwt.DeviceId, claims.DeviceId)
		c.Next()
	}
}
