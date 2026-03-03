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

// DeskManageHandler 桌台管理处理程序
type DeskManageHandler struct {
	deskManageSrv service.IDeskManageSrv
}

// ==================== 区域管理 ====================

// GetDeskRegionList 获取区域列表
// @Summary 获取区域列表
// @Description 获取桌台区域列表
// @Tags 商家端.桌台管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} resp.DeskRegionListResp "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/desk/region/list [get]
func (h *DeskManageHandler) GetDeskRegionList(c *gin.Context) {
	ctx := helper.GetContext(c)

	res, err := h.deskManageSrv.GetDeskRegionList(ctx)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, res)
}

// AddDeskRegion 添加区域
// @Summary 添加区域
// @Description 添加桌台区域
// @Tags 商家端.桌台管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.DeskRegionAddReq true "区域添加请求"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/desk/region/add [post]
func (h *DeskManageHandler) AddDeskRegion(c *gin.Context) {
	ctx := helper.GetContext(c)
	addReq := req.DeskRegionAddReq{}

	if err := c.ShouldBindJSON(&addReq); err != nil {
		helper.HandleValidationError(c, err, addReq, nil)
		return
	}

	err := h.deskManageSrv.AddDeskRegion(ctx, addReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, gin.H{}, "添加成功")
}

// EditDeskRegion 编辑区域
// @Summary 编辑区域
// @Description 编辑桌台区域
// @Tags 商家端.桌台管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.DeskRegionEditReq true "区域编辑请求"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/desk/region/edit [post]
func (h *DeskManageHandler) EditDeskRegion(c *gin.Context) {
	ctx := helper.GetContext(c)
	editReq := req.DeskRegionEditReq{}

	if err := c.ShouldBindJSON(&editReq); err != nil {
		helper.HandleValidationError(c, err, editReq, nil)
		return
	}

	err := h.deskManageSrv.EditDeskRegion(ctx, editReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, gin.H{}, "保存成功")
}

// DeleteDeskRegion 删除区域
// @Summary 删除区域
// @Description 删除桌台区域
// @Tags 商家端.桌台管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.DeskRegionDeleteReq true "区域删除请求"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/desk/region/delete [post]
func (h *DeskManageHandler) DeleteDeskRegion(c *gin.Context) {
	ctx := helper.GetContext(c)
	deleteReq := req.DeskRegionDeleteReq{}

	if err := c.ShouldBindJSON(&deleteReq); err != nil {
		helper.HandleValidationError(c, err, deleteReq, nil)
		return
	}

	err := h.deskManageSrv.DeleteDeskRegion(ctx, deleteReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, gin.H{}, "删除成功")
}

// ==================== 类型管理 ====================

// GetDeskTypeList 获取类型列表
// @Summary 获取类型列表
// @Description 获取桌台类型列表
// @Tags 商家端.桌台管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} resp.DeskTypeListResp "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/desk/type/list [get]
func (h *DeskManageHandler) GetDeskTypeList(c *gin.Context) {
	ctx := helper.GetContext(c)

	res, err := h.deskManageSrv.GetDeskTypeList(ctx)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, res)
}

// AddDeskType 添加类型
// @Summary 添加类型
// @Description 添加桌台类型
// @Tags 商家端.桌台管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.DeskTypeAddReq true "类型添加请求"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/desk/type/add [post]
func (h *DeskManageHandler) AddDeskType(c *gin.Context) {
	ctx := helper.GetContext(c)
	addReq := req.DeskTypeAddReq{}

	if err := c.ShouldBindJSON(&addReq); err != nil {
		helper.HandleValidationError(c, err, addReq, nil)
		return
	}

	err := h.deskManageSrv.AddDeskType(ctx, addReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, gin.H{}, "添加成功")
}

// EditDeskType 编辑类型
// @Summary 编辑类型
// @Description 编辑桌台类型
// @Tags 商家端.桌台管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.DeskTypeEditReq true "类型编辑请求"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/desk/type/edit [post]
func (h *DeskManageHandler) EditDeskType(c *gin.Context) {
	ctx := helper.GetContext(c)
	editReq := req.DeskTypeEditReq{}

	if err := c.ShouldBindJSON(&editReq); err != nil {
		helper.HandleValidationError(c, err, editReq, nil)
		return
	}

	err := h.deskManageSrv.EditDeskType(ctx, editReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, gin.H{}, "保存成功")
}

// DeleteDeskType 删除类型
// @Summary 删除类型
// @Description 删除桌台类型
// @Tags 商家端.桌台管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.DeskTypeDeleteReq true "类型删除请求"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/desk/type/delete [post]
func (h *DeskManageHandler) DeleteDeskType(c *gin.Context) {
	ctx := helper.GetContext(c)
	deleteReq := req.DeskTypeDeleteReq{}

	if err := c.ShouldBindJSON(&deleteReq); err != nil {
		helper.HandleValidationError(c, err, deleteReq, nil)
		return
	}

	err := h.deskManageSrv.DeleteDeskType(ctx, deleteReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, gin.H{}, "删除成功")
}

// ==================== 桌台管理 ====================

// GetDeskManageList 获取桌台列表
// @Summary 获取桌台列表
// @Description 获取桌台管理列表（分页）
// @Tags 商家端.桌台管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param page_no query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Success 200 {object} resp.DeskManageListResp "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/desk/list [get]
func (h *DeskManageHandler) GetDeskManageList(c *gin.Context) {
	ctx := helper.GetContext(c)
	listReq := req.DeskManageListReq{}

	if err := c.ShouldBindQuery(&listReq); err != nil {
		helper.HandleValidationError(c, err, listReq, nil)
		return
	}

	res, err := h.deskManageSrv.GetDeskManageList(ctx, listReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, res)
}

// AddDesk 添加桌台
// @Summary 添加桌台
// @Description 添加桌台
// @Tags 商家端.桌台管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.DeskAddReq true "桌台添加请求"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/desk/add [post]
func (h *DeskManageHandler) AddDesk(c *gin.Context) {
	ctx := helper.GetContext(c)
	addReq := req.DeskAddReq{}

	if err := c.ShouldBindJSON(&addReq); err != nil {
		helper.HandleValidationError(c, err, addReq, nil)
		return
	}

	err := h.deskManageSrv.AddDesk(ctx, addReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, gin.H{}, "添加成功")
}

// EditDesk 编辑桌台
// @Summary 编辑桌台
// @Description 编辑桌台
// @Tags 商家端.桌台管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.DeskEditReq true "桌台编辑请求"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/desk/edit [post]
func (h *DeskManageHandler) EditDesk(c *gin.Context) {
	ctx := helper.GetContext(c)
	editReq := req.DeskEditReq{}

	if err := c.ShouldBindJSON(&editReq); err != nil {
		helper.HandleValidationError(c, err, editReq, nil)
		return
	}

	err := h.deskManageSrv.EditDesk(ctx, editReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, gin.H{}, "保存成功")
}

// DeleteDesk 删除桌台
// @Summary 删除桌台
// @Description 删除桌台
// @Tags 商家端.桌台管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.DeskDeleteReq true "桌台删除请求"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/desk/delete [post]
func (h *DeskManageHandler) DeleteDesk(c *gin.Context) {
	ctx := helper.GetContext(c)
	deleteReq := req.DeskDeleteReq{}

	if err := c.ShouldBindJSON(&deleteReq); err != nil {
		helper.HandleValidationError(c, err, deleteReq, nil)
		return
	}

	err := h.deskManageSrv.DeleteDesk(ctx, deleteReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, gin.H{}, "删除成功")
}

// RegisterDeskManageHandlers 注册桌台管理路由
func RegisterDeskManageHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
	handler := DeskManageHandler{
		deskManageSrv: service.NewDeskManageSrv(dbm),
	}

	captchaSrv := service.NewCaptchaSrv(cache)
	settingSrv := setting.NewSrv(dbm, cache)
	roleAccessSrv := service.NewRoleAccessSrv(dbm)
	deviceSrv := service.NewDeviceSrv(settingSrv, dbm)
	cashBoxSrv := service.NewCashBoxSrv(dbm)
	statisticsSrv := service.NewStatisticsSrv()
	staffShiftSrv := service.NewStaffShiftSrv(cache, dbm, cashBoxSrv, statisticsSrv)
	authSrv := service.NewAuthSrv(dbm, captchaSrv, roleAccessSrv, deviceSrv, staffShiftSrv, settingSrv)

	privateApi := router.Group("", middleware.Auth(authSrv, dbm))
	{
		// 区域管理
		privateApi.GET("/desk/region/list", handler.GetDeskRegionList)
		privateApi.POST("/desk/region/add", handler.AddDeskRegion)
		privateApi.POST("/desk/region/edit", handler.EditDeskRegion)
		privateApi.POST("/desk/region/delete", handler.DeleteDeskRegion)

		// 类型管理
		privateApi.GET("/desk/type/list", handler.GetDeskTypeList)
		privateApi.POST("/desk/type/add", handler.AddDeskType)
		privateApi.POST("/desk/type/edit", handler.EditDeskType)
		privateApi.POST("/desk/type/delete", handler.DeleteDeskType)

		// 桌台管理
		privateApi.GET("/desk/list", handler.GetDeskManageList)
		privateApi.POST("/desk/add", handler.AddDesk)
		privateApi.POST("/desk/edit", handler.EditDesk)
		privateApi.POST("/desk/delete", handler.DeleteDesk)
	}
}
