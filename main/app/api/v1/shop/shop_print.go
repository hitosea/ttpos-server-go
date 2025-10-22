package shop

import (
	"strconv"
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

// PrintHandler 打印控制器
type PrintHandler struct {
	authSrv    service.IAuthSrv
	printerSrv service.IPrinterSrv
}

// GetPrintMenuList 获取打印菜单列表
// @Summary 获取打印菜单列表
// @Description 获取打印菜单列表
// @Tags 商家端.打印管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{data=resp.PrintMenuListResp} "成功"
// @Router /shop/printer/menu/list [get]
func (h *PrintHandler) GetPrintMenuList(c *gin.Context) {
	ctx := helper.GetContext(c)
	result, err := h.printerSrv.GetPrintMenuList(ctx)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, result)
}

// GetPrintMenuDetail 获取打印菜单详情
// @Summary 获取打印菜单详情
// @Description 获取打印菜单详情
// @Tags 商家端.打印管理
// @Accept json
// @Produce json
// @Param id query uint64 true "模板ID"
// @Security JwtToken
// @Success 200 {object} dto.Response{data=resp.PrintMenuDetailResp} "成功"
// @Router /shop/printer/menu/detail [get]
func (h *PrintHandler) GetPrintMenuDetail(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 获取路径参数
	id := c.Query("id")
	if id == "" {
		helper.Fail(c, constant.CodeFail, "模板ID参数错误")
		return
	}
	idUint, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		helper.Fail(c, constant.CodeFail, "模板ID参数值错误")
		return
	}
	result, err := h.printerSrv.GetPrintMenuDetail(ctx, idUint)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, result)
}

// EditPrinterCustomize 编辑打印机定制
// @Summary 编辑打印机定制
// @Description 编辑打印机定制
// @Tags 商家端.打印管理
// @Accept json
// @Produce json
// @Param data body req.EditPrinterCustomizeReq true "编辑打印机定制请求"
// @Security JwtToken
// @Success 200 {object} dto.Response{data=string} "成功"
// @Router /shop/printer/customize/edit [post]
func (h *PrintHandler) EditPrinterCustomize(c *gin.Context) {
	ctx := helper.GetContext(c)
	editPrinterCustomizeReq := req.EditPrinterCustomizeReq{}
	if err := c.ShouldBindJSON(&editPrinterCustomizeReq); err != nil {
		helper.HandleValidationError(c, err, editPrinterCustomizeReq, nil)
		return
	}
	err := h.printerSrv.EditPrinterCustomize(ctx, editPrinterCustomizeReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, nil, "编辑打印机定制成功")
}

// DeletePrinterCustomize 删除打印机定制
// @Summary 删除打印机定制
// @Description 删除打印机定制
// @Tags 商家端.打印管理
// @Accept json
// @Produce json
// @Param id query uint64 true "打印机定制ID"
// @Security JwtToken
// @Success 200 {object} dto.Response{data=string} "成功"
// @Router /shop/printer/customize/delete [delete]
func (h *PrintHandler) DeletePrinterCustomize(c *gin.Context) {
	ctx := helper.GetContext(c)
	id := c.Query("id")
	idUint, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		helper.Fail(c, constant.CodeFail, "打印机定制ID参数值错误")
		return
	}
	err = h.printerSrv.DeletePrinterCustomize(ctx, idUint)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, nil, "删除打印机定制成功")
}

// CreatePrinterCustomize 创建打印机定制
// @Summary 创建打印机定制
// @Description 创建打印机定制
// @Tags 商家端.打印管理
// @Accept json
// @Produce json
// @Param data body req.CreatePrinterCustomizeReq true "创建打印机定制请求"
// @Security JwtToken
// @Success 200 {object} dto.Response{data=string} "成功"
// @Router /shop/printer/customize/create [post]
func (h *PrintHandler) CreatePrinterCustomize(c *gin.Context) {
	ctx := helper.GetContext(c)
	createPrinterCustomizeReq := req.CreatePrinterCustomizeReq{}
	if err := c.ShouldBindJSON(&createPrinterCustomizeReq); err != nil {
		helper.HandleValidationError(c, err, createPrinterCustomizeReq, nil)
		return
	}
	err := h.printerSrv.CreatePrinterCustomize(ctx, createPrinterCustomizeReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, nil, "创建打印机定制成功")
}

// UsePrinterCustomize 使用打印机定制

// @Summary 使用打印机定制
// @Description 使用打印机定制
// @Tags 商家端.打印管理
// @Accept json
// @Produce json
// @Param id query uint64 true "打印机定制ID"
// @Security JwtToken
// @Success 200 {object} dto.Response{data=string} "成功"
// @Router /shop/printer/customize/use [post]
func (h *PrintHandler) UsePrinterCustomize(c *gin.Context) {
	ctx := helper.GetContext(c)
	id := c.Query("id")
	idUint, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		helper.Fail(c, constant.CodeFail, "打印机定制ID参数值错误")
		return
	}
	err = h.printerSrv.UsePrinterCustomize(ctx, idUint)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, nil, "使用打印机定制成功")
}

// RegisterPrintHandlers 注册打印相关路由
func RegisterPrintHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
	// 初始化服务
	captchaSrv := service.NewCaptchaSrv(cache)
	settingSrv := setting.NewSrv(dbm, cache)
	roleAccessSrv := service.NewRoleAccessSrv(dbm)
	deviceSrv := service.NewDeviceSrv(settingSrv, dbm)
	cashBoxSrv := service.NewCashBoxSrv(dbm)
	statisticsSrv := service.NewStatisticsSrv()
	staffShiftSrv := service.NewStaffShiftSrv(cache, dbm, cashBoxSrv, statisticsSrv)
	authSrv := service.NewAuthSrv(dbm, captchaSrv, roleAccessSrv, deviceSrv, staffShiftSrv, settingSrv)
	printerSrv := service.NewPrinterSrv(dbm, cache)

	// 初始化控制器
	printerHandler := &PrintHandler{
		authSrv:    authSrv,
		printerSrv: printerSrv,
	}

	// 需要认证的路由
	privateApi := router.Group("", middleware.Auth(authSrv, dbm))
	{
		privateApi.GET("/printer/menu/list", printerHandler.GetPrintMenuList)                 // 打印菜单列表
		privateApi.GET("/printer/menu/detail", printerHandler.GetPrintMenuDetail)             // 打印菜单详情
		privateApi.POST("/printer/customize/edit", printerHandler.EditPrinterCustomize)       // 编辑打印机定制
		privateApi.DELETE("/printer/customize/delete", printerHandler.DeletePrinterCustomize) // 删除打印机定制
		privateApi.POST("/printer/customize/create", printerHandler.CreatePrinterCustomize)   // 创建打印机定制
		privateApi.POST("/printer/customize/use", printerHandler.UsePrinterCustomize)         // 使用打印机定制
	}
}
