package middleware

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/constant/jwt"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/service"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/auth"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func Auth(authSrv service.IAuthSrv, dbm *database.DBManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			helper.Fail(c, constant.CodeTokenInvalid, "token 为空")
			c.Abort()
			return
		}
		ParseJwt(c, authHeader, authSrv, dbm)
		c.Next()
	}
}

func DeskAuth(authSrv service.IAuthSrv, dbm *database.DBManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		deskTokenHeader := c.GetHeader("Token") // 桌台二维码的token

		if deskTokenHeader == "" {
			helper.Fail(c, constant.CodeTokenInvalid, "二维码已失效，请联系商家")
			c.Abort()
			return
		}
		token, err := auth.DecodeDeskToken(deskTokenHeader)
		if err != nil {
			helper.Fail(c, constant.CodeTokenInvalid, "二维码已失效，请联系商家")
			c.Abort()
			return
		}
		ParseDeskToken(c, token, authSrv, dbm)
		c.Next()
	}
}

func ParseJwt(c *gin.Context, authHeader string, authSrv service.IAuthSrv, dbm *database.DBManager) {

	parts := strings.SplitN(authHeader, " ", 2)
	if !(len(parts) == 2 && parts[0] == "Bearer") {
		helper.Fail(c, constant.CodeTokenInvalid, "token 格式错误")
		c.Abort()
		return
	}

	// 验证token
	claims, err := auth.ParseToken(parts[1], config.JWT.Secret)
	if err != nil {
		helper.Fail(c, constant.CodeTokenInvalid, "无效的token")
		c.Abort()
		return
	}

	if !slices.Contains([]string{constant.SourceShop, constant.SourceCashier, constant.SourceAssistant, constant.SourceKitchen, constant.SourceTablet}, claims.Source) {
		helper.Fail(c, constant.CodeAccessDenied, "用户信息错误")
		c.Abort()
		return
	}

	if !regexp.MustCompile(`^/api/v\d+/` + claims.Source).Match([]byte(c.Request.URL.Path)) {
		helper.Fail(c, constant.CodeAccessDenied, "用户信息错误")
		c.Abort()
		return
	}

	// 用户鉴权
	ctx := context.NewContext(
		context.WithSource(claims.Source),
		context.WithCompanyUuid(claims.CompanyUuid),
		context.WithDeviceUuid(claims.DeviceUuid),
	)
	company, companySetting, staff, desk, err := authSrv.Auth(ctx, req.Authenticate{
		Source:      claims.Source,
		DeviceId:    claims.DeviceId,
		CompanyUuid: claims.CompanyUuid,
		StaffUuid:   claims.StaffUuid,
		UrlPath:     c.Request.URL.Path,
		Assistant: req.Assistant{
			DeviceId:  claims.Assistant.DeviceId,
			StaffUuid: claims.Assistant.StaffUuid,
		},
		TokenIssuedAt: claims.IssuedAt.Unix(),
	})
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeAccessDenied, err)
		c.Abort()
		return
	}

	// 将用户信息存储到上下文
	c.Set(jwt.Source, claims.Source)
	c.Set(jwt.Company, company)               // 商家信息
	c.Set(jwt.CompanySetting, companySetting) // 商家设置信息
	c.Set(jwt.Staff, staff)                   // 员工信息，如果是点餐助手，应该是收银员

	c.Set(jwt.CompanyUuid, claims.CompanyUuid) // 商家Uuid
	c.Set(jwt.DeviceId, claims.DeviceId)       // 设备ID，如果是点餐助手，应该是收银机设备ID
	c.Set(jwt.StaffUuid, claims.StaffUuid)     // 员工Uuid，如果是点餐助手，应该是收银员Uuid

	c.Set(jwt.AssistantStaffUuid, claims.Assistant.StaffUuid) // 点餐助手员工Uuid
	c.Set(jwt.AssistantDeviceId, claims.Assistant.DeviceId)   // 点餐助手设备ID

	c.Set(jwt.DeviceUuid, claims.DeviceUuid) // 桌台绑定的设备uuid
	// 注入一个uuid

	c.Set(jwt.RequestUuid, uuid.New().String()) // 桌台绑定的设备uuid

	c.Set(jwt.DeskUuid, desk.Uuid) // 桌台Uuid，平板端绑定的桌台Uuid

	c.Set(jwt.DB, dbm.GetDB(claims.CompanyUuid)) // 数据库连接
}

func ParseDeskToken(c *gin.Context, token *auth.DeskToken, authSrv service.IAuthSrv, dbm *database.DBManager) {
	ctx := context.NewContext(context.WithDeskUuid(token.DeskUuid), context.WithCompanyUuid(token.CompanyUuid))

	// 用户鉴权, 查询desk表判断qrcode_token值是否相同
	company, err := authSrv.AuthDesk(ctx, token.DeskTokenValue)
	if err != nil {
		helper.Fail(c, constant.CodeTokenInvalid, "二维码已失效，请联系商家")
		c.Abort()
		return
	}
	// 将用户信息存储到上下文
	c.Set(jwt.Source, jwt.SourceH5)
	c.Set(jwt.CompanyUuid, token.CompanyUuid)          // 商家Uuid
	c.Set(jwt.DeskUuid, token.DeskUuid)                // 桌台Uuid
	c.Set(jwt.Company, *company)                       // 商家信息
	c.Set(jwt.CompanySetting, *company.CompanySetting) // 商家设置信息
	c.Set(jwt.DB, dbm.GetDB(token.CompanyUuid))        // 数据库连接
	fmt.Println(fmt.Sprintf("ParseDeskToken deskUuid: %d, companyUuid: %d", token.DeskUuid, token.CompanyUuid))
}
