package shop

import (
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/service"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/middleware"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"

	"github.com/gin-gonic/gin"
)

// CountryHandler 国家处理器
//
// 任务: story-shop-material-origin Phase 3.8
// 需求: R3.4
//
// @version v2.11.0
type CountryHandler struct {
	countrySrv service.ICountrySrv
}

// GetList 获取国家列表
// @Summary 获取国家列表
// @Description 获取所有国家列表（197个国家/地区）
// @Tags 商家端.物品管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{data=material_resp.CountryListResp}
// @Failure 400 {object} nil "错误请求"
// @Router /shop/country/list [get]
func (h *CountryHandler) GetList(c *gin.Context) {
	ctx := helper.GetContext(c)

	list, err := h.countrySrv.GetList(ctx)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, list)
}

// RegisterCountryRoutes 注册国家路由
func RegisterCountryRoutes(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
	// 初始化服务
	captchaSrv := service.NewCaptchaSrv(cache)
	settingSrv := setting.NewSrv(dbm, cache)
	roleAccessSrv := service.NewRoleAccessSrv(dbm)
	deviceSrv := service.NewDeviceSrv(settingSrv, dbm)
	cashBoxSrv := service.NewCashBoxSrv(dbm)
	statisticsSrv := service.NewStatisticsSrv()
	staffShiftSrv := service.NewStaffShiftSrv(cache, dbm, cashBoxSrv, statisticsSrv)
	authSrv := service.NewAuthSrv(dbm, captchaSrv, roleAccessSrv, deviceSrv, staffShiftSrv, settingSrv)

	countrySrv := service.NewCountrySrv()
	handler := &CountryHandler{
		countrySrv: countrySrv,
	}

	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv, dbm))
	{
		privateApi.GET("/country/list", handler.GetList)
	}
}

