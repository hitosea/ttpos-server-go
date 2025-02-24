package middleware

import (
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/constant/jwt"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/service"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/auth"
	"ttpos-server-go/pkg/context"

	"github.com/gin-gonic/gin"
)

func Auth(authSrv service.IAuthSrv) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			helper.Fail(c, constant.CodeBadRequest, "token 为空")
			c.Abort()
			return
		}
		ParseJwt(c, authHeader, authSrv)
		c.Next()
	}
}

func DeskAuth(authSrv service.IAuthSrv) gin.HandlerFunc {
	return func(c *gin.Context) {
		deskTokenHeader := c.GetHeader("Token") // 桌台二维码的token

		if deskTokenHeader == "" {
			helper.H5Fail(c, 0, "二维码已失效，请联系商家")
			c.Abort()
			return
		}
		token, err := auth.DecodeDeskToken(deskTokenHeader)
		if err != nil {
			helper.H5Fail(c, 0, "二维码已失效，请联系商家")
			c.Abort()
			return
		}
		ParseDeskToken(c, token, authSrv)
		c.Next()
	}
}

func ParseJwt(c *gin.Context, authHeader string, authSrv service.IAuthSrv) {

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

	if !slices.Contains([]string{constant.SourceShop, constant.SourceCashier, constant.SourceAssistant, constant.SourceKitchen, constant.SourceTablet}, claims.Source) {
		helper.Fail(c, constant.CodeBadRequest, "用户信息错误")
		c.Abort()
		return
	}

	if !regexp.MustCompile(`^/api/v\d+/` + claims.Source).Match([]byte(c.Request.URL.Path)) {
		helper.Fail(c, constant.CodeBadRequest, "用户信息错误")
		c.Abort()
		return
	}

	// 桌台Uuid
	tableUuid, _ := strconv.ParseUint(c.GetHeader("TABLE-UUID"), 10, 64)

	// 用户鉴权
	ctx := context.NewContext(context.WithSource(claims.Source), context.WithCompanyUuid(claims.CompanyUuid))
	company, companySetting, staff, err := authSrv.Auth(ctx, req.Authenticate{
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

		DeviceUuid: claims.DeviceUuid, // 用于判断是否绑定桌台
		TableUuid:  tableUuid,         // 用于判断是否绑定桌台
	})
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeBadRequest, err)
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
}

func ParseDeskToken(c *gin.Context, token *auth.DeskToken, authSrv service.IAuthSrv) {
	deskUuid, err := strconv.Atoi(token.DeskUuid)
	if err != nil {
		helper.H5Fail(c, constant.CodeBadRequest, "二维码已失效，请联系商家")
		c.Abort()
		return
	}
	ctx := context.NewContext(context.WithDeskUuid(uint64(deskUuid)), context.WithCompanyUuid(uint64(token.CompanyUuid)))

	// 用户鉴权, 查询desk表判断qrcode_token值是否相同
	err = authSrv.AuthDesk(ctx, token.DeskTokenValue)
	if err != nil {
		helper.H5Fail(c, constant.CodeBadRequest, "二维码已失效，请联系商家")
		c.Abort()
		return
	}
	// 将用户信息存储到上下文
	c.Set(jwt.Source, "H5")
	c.Set(jwt.CompanyUuid, ctx.GetCompanyUuid()) // 商家Uuid
	c.Set(jwt.DeskUuid, ctx.GetDeskUuid())       // 桌台Uuid
	fmt.Println(fmt.Sprintf("deskUuid: %d, companyUuid: %d", ctx.GetDeskUuid(), ctx.GetCompanyUuid()))
}
