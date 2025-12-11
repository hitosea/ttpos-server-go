package middleware

import (
	builtinerrors "errors"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/constant/jwt"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/service"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/auth"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"ttpos-server-go/app/errors"
)

var (
	authWhiteList = []string{
		"/api/v1/cashier/old/order/cash/balance",
		"/api/v1/cashier/old/order/member/balance",
	}
)

func Auth(authSrv service.IAuthSrv, dbm *database.DBManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			helper.Fail(c, constant.CodeTokenInvalid, "token is required")
			c.Abort()
			return
		}
		ParseJwt(c, authHeader, authSrv, dbm)
		c.Next()
	}
}

func ParseJwt(c *gin.Context, authHeader string, authSrv service.IAuthSrv, dbm *database.DBManager) {

	parts := strings.SplitN(authHeader, " ", 2)
	if !(len(parts) == 2 && parts[0] == "Bearer") {
		helper.Fail(c, constant.CodeTokenInvalid, "登录失效，请重新登录")
		c.Abort()
		return
	}

	// 验证token
	claims, err := auth.ParseToken(parts[1], config.JWT.Secret)
	if err != nil {
		helper.Fail(c, constant.CodeTokenInvalid, "登录失效，请重新登录")
		c.Abort()
		return
	}

	if !slices.Contains([]string{constant.SourceShop, constant.SourceCashier, constant.SourceAssistant, constant.SourceKitchen, constant.SourceTablet}, claims.Source) {
		helper.Fail(c, constant.CodeAccessDenied, "用户信息错误")
		c.Abort()
		return
	}

	urlPath := c.Request.URL.Path
	if !slices.Contains(authWhiteList, urlPath) && !regexp.MustCompile(`^/api/v\d+/`+claims.Source).Match([]byte(urlPath)) {
		helper.Fail(c, constant.CodeAccessDenied, "用户信息错误")
		c.Abort()
		return
	}

	if strings.HasSuffix(urlPath, "/refresh_token") && !claims.IsRefreshToken {
		helper.Fail(c, constant.CodeTokenInvalid, "无效的refresh_token")
		c.Abort()
		return
	} else if !strings.HasSuffix(urlPath, "/refresh_token") && claims.IsRefreshToken {
		helper.Fail(c, constant.CodeTokenInvalid, "无效的token")
		c.Abort()
		return
	}

	// 如果 company_uuid 为 0，只允许访问 base 接口和切换门店接口
	if claims.CompanyUuid == 0 {
		// 定义允许访问的接口列表
		allowedPaths := []string{
			fmt.Sprintf("/api/v1/%s/base", claims.Source),
			fmt.Sprintf("/api/v1/%s/store_switch", claims.Source),
			fmt.Sprintf("/api/v1/%s/change_password", claims.Source),
		}

		// 检查是否在允许列表中
		isAllowed := slices.Contains(allowedPaths, urlPath)

		if !isAllowed {
			helper.Fail(c, constant.CodeNeedCompanyUuid, "请先选择门店")
			c.Abort()
			return
		}

		// 允许访问，但跳过 Auth 验证（因为 company_uuid 为 0）
		// 设置基本的上下文信息
		c.Set(jwt.Source, claims.Source)
		c.Set(jwt.CompanyUuid, 0)
		c.Set(jwt.StaffUuid, claims.StaffUuid)
		c.Set(jwt.DeviceId, claims.DeviceId)
		c.Set(jwt.RequestUuid, uuid.New().String())
		c.Set(jwt.AssistantStaffUuid, claims.Assistant.StaffUuid)
		c.Set(jwt.AssistantDeviceId, claims.Assistant.DeviceId)
		c.Set(jwt.DeviceUuid, claims.DeviceUuid)
		c.Set(jwt.Brand, claims.Brand)
		return
	}

	// company_uuid 不为 0，走原有逻辑
	// 用户鉴权
	ctx := context.NewContext(
		context.WithGinContext(c.Copy()),
		context.WithSource(claims.Source),
		context.WithCompanyUuid(claims.CompanyUuid),
		context.WithDeviceUuid(claims.DeviceUuid),
		context.WithLogger(logger.Logger),
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
		var appErr errors.AppError
		if builtinerrors.As(err, &appErr) {
			code := appErr.GetCode()
			if code == constant.CodeCashierHandedOver { // 已交班
				helper.ErrorWithData(c, code, appErr.GetData(), err)
				c.Abort()
				return
			}
			if code == constant.CodeTokenInvalid && appErr.GetData() != nil {
				helper.ErrorWithData(c, constant.CodeAccessDenied, appErr.GetData(), err)
				c.Abort()
				return
			}
		}
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
	c.Set(jwt.Brand, claims.Brand)               // 品牌名称
}

func DeskAuth(authSrv service.IAuthSrv, dbm *database.DBManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		deskTokenHeader := c.GetHeader("Authorization") // 桌台二维码的token

		if deskTokenHeader == "" {
			helper.Fail(c, constant.CodeTokenInvalid, "二维码已失效，请联系商家")
			c.Abort()
			return
		}
		parts := strings.SplitN(deskTokenHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			helper.Fail(c, constant.CodeTokenInvalid, "二维码已失效，请联系商家")
			c.Abort()
			return
		}
		token, err := auth.DecodeDeskToken(parts[1])
		if err != nil {
			helper.Fail(c, constant.CodeTokenInvalid, "二维码已失效，请联系商家")
			c.Abort()
			return
		}
		ParseDeskToken(c, token, authSrv, dbm)
		c.Next()
	}
}

func ParseDeskToken(c *gin.Context, token *auth.DeskToken, authSrv service.IAuthSrv, dbm *database.DBManager) {
	ctx := context.NewContext(context.WithDeskUuid(token.DeskUuid), context.WithCompanyUuid(token.CompanyUuid))

	// 用户鉴权, 查询desk表判断qrcode_token值是否相同
	company, err := authSrv.AuthDesk(ctx, token.DeskTokenValue)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeTokenInvalid, err)
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
}

func BusinessMenu(authSrv service.IAuthSrv, dbm *database.DBManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		menuHeader := c.GetHeader("Authorization") // 电子菜单二维码的token

		if menuHeader == "" {
			helper.Fail(c, constant.CodeTokenInvalid, "二维码已失效，请联系商家")
			c.Abort()
			return
		}
		parts := strings.SplitN(menuHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			helper.Fail(c, constant.CodeTokenInvalid, "二维码已失效，请联系商家")
			c.Abort()
			return
		}
		token, err := auth.DecodeMenuToken(parts[1])
		if err != nil {
			helper.Fail(c, constant.CodeTokenInvalid, "二维码已失效，请联系商家")
			c.Abort()
			return
		}

		ParseMenuToken(c, token, authSrv, dbm)
		c.Next()
	}
}

func ParseMenuToken(c *gin.Context, token *auth.MenuToken, authSrv service.IAuthSrv, dbm *database.DBManager) {
	ctx := context.NewContext(context.WithCompanyUuid(token.CompanyUuid))
	// 用户鉴权, 查询desk表判断qrcode_token值是否相同
	company, err := authSrv.AuthMenu(ctx, token.QrCode)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeTokenInvalid, err)
		c.Abort()
		return
	}
	// 将用户信息存储到上下文
	c.Set(jwt.Source, jwt.SourceH5)
	c.Set(jwt.CompanyUuid, token.CompanyUuid)          // 商家Uuid
	c.Set(jwt.Company, *company)                       // 商家信息
	c.Set(jwt.CompanySetting, *company.CompanySetting) // 商家设置信息
	c.Set(jwt.DB, dbm.GetDB(token.CompanyUuid))        // 数据库连接
	fmt.Println(fmt.Sprintf("ParseMenuToken companyUuid: %d", token.CompanyUuid))
}

// MemberAuth
func MemberAuth(authSrv service.IAuthSrv, dbm *database.DBManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		memberTokenHeader := c.GetHeader("Authorization")
		if memberTokenHeader == "" {
			helper.Fail(c, constant.CodeTokenInvalid, "Token无效, 请重新登录")
			c.Abort()
			return
		}
		parts := strings.SplitN(memberTokenHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			helper.Fail(c, constant.CodeTokenInvalid, "Token无效, 请重新登录")
			c.Abort()
			return
		}
		// 验证token
		claims, err := auth.ParseToken(parts[1], config.JWT.Secret)
		if err != nil {
			helper.Fail(c, constant.CodeTokenInvalid, "登录失效，请重新登录")
			c.Abort()
			return
		}
		//
		urlPath := c.Request.URL.Path
		if !regexp.MustCompile(`^/api/v\d+/` + claims.Source).Match([]byte(urlPath)) {
			helper.Fail(c, constant.CodeAccessDenied, "用户信息错误")
			c.Abort()
			return
		}
		//
		if strings.HasSuffix(urlPath, "/refresh_token") && !claims.IsRefreshToken {
			helper.Fail(c, constant.CodeTokenInvalid, "无效的refresh_token")
			c.Abort()
			return
		} else if !strings.HasSuffix(urlPath, "/refresh_token") && claims.IsRefreshToken {
			helper.Fail(c, constant.CodeTokenInvalid, "无效的token")
			c.Abort()
			return
		}
		// 验证会员是否存在
		db := dbm.GetDB(claims.CompanyUuid)
		if db == nil {
			helper.Fail(c, constant.CodeTokenInvalid, "无法使用该功能，请联系商家")
			c.Abort()
			return
		}
		memberRepo := repository.NewMemberRepo(db)
		member, err := memberRepo.GetMemberRecord(memberRepo.WhereUuid(claims.MemberUuid))
		if err != nil || member.IsDelete() {
			helper.Fail(c, constant.CodeTokenInvalid, "会员不存在")
			c.Abort()
			return
		}
		company, err := repository.NewCompanyRepo(dbm.GetDB(claims.CompanyUuid)).GetCompanyInfoByUuid(claims.CompanyUuid)
		if err != nil || company.IsExpired() || company.IsDelete() {
			helper.Fail(c, constant.CodeTokenInvalid, "无法使用该功能，请联系商家")
			c.Abort()
			return
		}
		// 将用户信息存储到上下文
		c.Set(jwt.Source, jwt.SourceMember)
		c.Set(jwt.CompanyUuid, claims.CompanyUuid)         // 商家Uuid
		c.Set(jwt.MemberUuid, claims.MemberUuid)           // 会员Uuid
		c.Set(jwt.Member, *member)                         // 会员信息
		c.Set(jwt.Company, *company)                       // 商家信息
		c.Set(jwt.CompanySetting, *company.CompanySetting) // 商家设置信息
		c.Set(jwt.DB, dbm.GetDB(claims.CompanyUuid))       // 数据库连接
		//
		c.Next()
	}
}
