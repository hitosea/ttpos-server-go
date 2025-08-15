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

// SupplierHandler 供应商控制器
type SupplierHandler struct {
	authSrv     service.IAuthSrv
	supplierSrv service.ISupplierSrv
}

// GetSupplierSelect 获取供应商列表选择器
// @Summary 获取供应商列表选择器
// @Description 获取供应商列表选择器
// @Tags 商家端.供应商管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{data=resp.SupplierSelectResp} "成功"
// @Router /shop/supplier/select [get]
func (h *SupplierHandler) GetSupplierSelect(c *gin.Context) {
	ctx := helper.GetContext(c)

	result, err := h.supplierSrv.GetSupplierSelect(ctx)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, result)
}

func RegisterSupplierHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
	// 初始化服务
	captchaSrv := service.NewCaptchaSrv(cache)
	settingSrv := setting.NewSrv(dbm, cache)
	roleAccessSrv := service.NewRoleAccessSrv(dbm)
	deviceSrv := service.NewDeviceSrv(settingSrv, dbm)
	cashBoxSrv := service.NewCashBoxSrv(dbm)
	statisticsSrv := service.NewStatisticsSrv()
	staffShiftSrv := service.NewStaffShiftSrv(cache, dbm, cashBoxSrv, statisticsSrv)
	authSrv := service.NewAuthSrv(dbm, captchaSrv, roleAccessSrv, deviceSrv, staffShiftSrv, settingSrv)

	// 供应商服务
	supplierSrv := service.NewSupplierSrv(dbm)

	wrapper := &SupplierHandler{
		authSrv:     authSrv,
		supplierSrv: supplierSrv,
	}

	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv, dbm))
	{
		privateApi.GET("/supplier/select", wrapper.GetSupplierSelect)
	}
}
