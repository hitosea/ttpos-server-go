package shop

import (
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
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

// GetSupplierList 获取供应商列表
// @Summary 获取供应商列表
// @Description 获取供应商列表，支持名称和编码关键字搜索
// @Tags 商家端.供应商管理
// @Accept json
// @Produce json
// @Param keyword query string false "关键字搜索：名称、编码"
// @Param page_no query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Security JwtToken
// @Success 200 {object} dto.Response{data=resp.SupplierListResp} "成功"
// @Router /shop/supplier/list [get]
func (h *SupplierHandler) GetSupplierList(c *gin.Context) {
	var request req.SupplierListReq
	if err := c.ShouldBind(&request); err != nil {
		helper.Fail(c, constant.CodeParamError, "参数错误")
		return
	}
	ctx := helper.GetContext(c)
	result, err := h.supplierSrv.GetSupplierList(ctx, request)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, result)
}

// CreateSupplier 创建供应商
// @Summary 创建供应商
// @Description 创建供应商，包含名称和编码重复性验证
// @Tags 商家端.供应商管理
// @Accept json
// @Produce json
// @Param request body req.SupplierCreateReq true "创建供应商请求"
// @Security JwtToken
// @Success 200 {object} dto.Response{data=resp.SupplierCreateResp} "成功"
// @Router /shop/supplier/create [post]
func (h *SupplierHandler) CreateSupplier(c *gin.Context) {
	var request req.SupplierCreateReq
	if err := c.ShouldBindJSON(&request); err != nil {
		helper.HandleValidationError(c, err, request, req.SupplierCreateReqMessage)
		return
	}
	ctx := helper.GetContext(c)
	err := h.supplierSrv.CreateSupplier(ctx, request)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, gin.H{})
}

// UpdateSupplier 更新供应商
// @Summary 更新供应商
// @Description 更新供应商信息，包含名称和编码重复性验证
// @Tags 商家端.供应商管理
// @Accept json
// @Produce json
// @Param request body req.SupplierUpdateReq true "更新供应商请求"
// @Security JwtToken
// @Success 200 {object} dto.Response "成功"
// @Router /shop/supplier/update [post]
func (h *SupplierHandler) UpdateSupplier(c *gin.Context) {
	var request req.SupplierUpdateReq
	if err := c.ShouldBindJSON(&request); err != nil {
		helper.HandleValidationError(c, err, request, req.SupplierUpdateReqMessage)
		return
	}
	ctx := helper.GetContext(c)
	err := h.supplierSrv.UpdateSupplier(ctx, request)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, nil)
}

// DeleteSupplier 删除供应商
// @Summary 删除供应商
// @Description 删除供应商
// @Tags 商家端.供应商管理
// @Accept json
// @Produce json
// @Param request body req.SupplierDeleteReq true "删除供应商请求"
// @Security JwtToken
// @Success 200 {object} dto.Response "成功"
// @Router /shop/supplier/delete [delete]
func (h *SupplierHandler) DeleteSupplier(c *gin.Context) {
	var request req.SupplierDeleteReq
	if err := c.ShouldBindJSON(&request); err != nil {
		helper.HandleValidationError(c, err, request, req.SupplierDeleteReqMessage)
		return
	}

	ctx := helper.GetContext(c)
	err := h.supplierSrv.DeleteSupplier(ctx, request)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, nil)
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
		privateApi.GET("/supplier/list", wrapper.GetSupplierList)
		privateApi.POST("/supplier/create", wrapper.CreateSupplier)
		privateApi.POST("/supplier/update", wrapper.UpdateSupplier)
		privateApi.DELETE("/supplier/delete", wrapper.DeleteSupplier)
	}
}
