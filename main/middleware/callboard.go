package middleware

import (
	"errors"
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	tterrors "ttpos-server-go/app/errors"
	"ttpos-server-go/pkg/context"

	"github.com/gin-gonic/gin"
)

type BindedDeviceValidator interface {
	BindedDeviceValidate(deviceId string) (companyUuid uint64, err error)
}

// BindedDeviceAuth 已绑定设备认证中间件
func BindedDeviceAuth(validator BindedDeviceValidator) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取设备认证参数
		deviceId := getDeviceId(c)
		companyUuid, err := validator.BindedDeviceValidate(deviceId)
		if err != nil {
			var appError tterrors.AppError
			if errors.As(err, &appError) {
				helper.Fail(c, appError.Code, appError.Message)
				c.Abort()
				return
			}
			helper.Fail(c, constant.CodeFail, "auth failed")
			c.Abort()
			return
		}
		ctx := c.Request.Context()
		ctx = context.WithCompanyUuidV2(ctx, companyUuid) // 在上下文中添加公司uuid
		ctx = context.WithDeviceId(ctx, deviceId)         // 在上下文中添加设备id

		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func getDeviceId(c *gin.Context) string {
	if c.GetHeader("Device-Id") != "" {
		return c.GetHeader("Device-Id")
	}
	// 从请求参数中获取
	if c.Query("device_id") != "" {
		return c.Query("device_id")
	}
	if c.Param("device_id") != "" {
		return c.Param("device_id")
	}
	return ""
}

func getTimestamp(c *gin.Context) string {
	if c.GetHeader("Timestamp") != "" {
		return c.GetHeader("Timestamp")
	}
	// 从请求参数中获取
	if c.Query("timestamp") != "" {
		return c.Query("timestamp")
	}
	return ""
}

func getSignature(c *gin.Context) string {
	return c.GetHeader("Signature")
}
