package shop

import (
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/service"
	"ttpos-server-go/middleware"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"

	"github.com/gin-gonic/gin"
)

type CompanyAreaHandler struct {
	companyAreaSrv service.ICompanyAreaSrv
}

// RegisterCompanyAreaHandlers 注册门店区域管理路由
func RegisterCompanyAreaHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
	roleAccessSrv := service.NewRoleAccessSrv(dbm)

	wrapper := &CompanyAreaHandler{
		companyAreaSrv: service.NewCompanyAreaSrv(dbm),
	}

	privateApi := router.Group("", middleware.Auth(service.NewAuthSrv(dbm, service.NewCaptchaSrv(cache), roleAccessSrv, nil, nil, nil), dbm))
	{
		privateApi.GET("/company_area/list", wrapper.GetList)             // 区域列表
		privateApi.GET("/company_area/info", wrapper.GetInfo)             // 区域详情
		privateApi.POST("/company_area/create", wrapper.Create)           // 创建区域
		privateApi.POST("/company_area/update", wrapper.Update)           // 编辑区域
		privateApi.DELETE("/company_area/delete", wrapper.Delete)         // 删除区域
		privateApi.GET("/company_area/unassigned", wrapper.GetUnassigned) // 未分配门店列表
		privateApi.GET("/company_area/shop_list", wrapper.GetShopList)   // 区域管理用门店列表
	}
}

// GetList 获取区域列表
// @Summary 获取区域列表
// @Description 获取当前总部下所有区域列表，包含每个区域关联的门店数量、总门店数和未分配门店数
// @Tags 商家端.门店区域
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{data=resp.CompanyAreaListResp}
// @Router /shop/company_area/list [get]
func (h *CompanyAreaHandler) GetList(c *gin.Context) {
	ctx := helper.GetContext(c)
	result, err := h.companyAreaSrv.GetList(ctx)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeSystemError, err)
		return
	}
	helper.Success(c, result)
}

// GetInfo 获取区域详情
// @Summary 获取区域详情
// @Description 获取区域详情，包含区域关联的门店列表
// @Tags 商家端.门店区域
// @Accept json
// @Produce json
// @Security JwtToken
// @Param uuid query uint64 true "区域UUID"
// @Success 200 {object} dto.Response{data=resp.CompanyAreaInfoResp}
// @Router /shop/company_area/info [get]
func (h *CompanyAreaHandler) GetInfo(c *gin.Context) {
	ctx := helper.GetContext(c)
	var infoReq req.CompanyAreaInfoReq
	if err := c.ShouldBindQuery(&infoReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeParamError, err)
		return
	}
	result, err := h.companyAreaSrv.GetInfo(ctx, infoReq.Uuid)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeSystemError, err)
		return
	}
	helper.Success(c, result)
}

// Create 创建区域
// @Summary 创建区域
// @Description 创建门店区域，支持多语言名称，可同时关联门店。区域名称不可重复，每个门店只能属于一个区域
// @Tags 商家端.门店区域
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.CompanyAreaCreateReq true "创建区域请求"
// @Success 200 {object} dto.Response
// @Router /shop/company_area/create [post]
func (h *CompanyAreaHandler) Create(c *gin.Context) {
	ctx := helper.GetContext(c)
	var createReq req.CompanyAreaCreateReq
	if err := c.ShouldBindJSON(&createReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeParamError, err)
		return
	}
	if err := h.companyAreaSrv.Create(ctx, createReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeSystemError, err)
		return
	}
	helper.Success(c, gin.H{})
}

// Update 编辑区域
// @Summary 编辑区域
// @Description 编辑区域名称、排序和关联门店。支持差量更新门店归属，移除的门店变为未分配
// @Tags 商家端.门店区域
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.CompanyAreaUpdateReq true "编辑区域请求"
// @Success 200 {object} dto.Response
// @Router /shop/company_area/update [post]
func (h *CompanyAreaHandler) Update(c *gin.Context) {
	ctx := helper.GetContext(c)
	var updateReq req.CompanyAreaUpdateReq
	if err := c.ShouldBindJSON(&updateReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeParamError, err)
		return
	}
	if err := h.companyAreaSrv.Update(ctx, updateReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeSystemError, err)
		return
	}
	helper.Success(c, gin.H{})
}

// Delete 删除区域
// @Summary 删除区域
// @Description 删除区域，该区域下所有门店自动变为未分配状态
// @Tags 商家端.门店区域
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.CompanyAreaDeleteReq true "删除区域请求"
// @Success 200 {object} dto.Response
// @Router /shop/company_area/delete [delete]
func (h *CompanyAreaHandler) Delete(c *gin.Context) {
	ctx := helper.GetContext(c)
	var deleteReq req.CompanyAreaDeleteReq
	if err := c.ShouldBindJSON(&deleteReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeParamError, err)
		return
	}
	if err := h.companyAreaSrv.Delete(ctx, deleteReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeSystemError, err)
		return
	}
	helper.Success(c, gin.H{})
}

// GetUnassigned 获取未分配门店列表
// @Summary 获取未分配门店列表
// @Description 获取总部下所有未分配区域的门店列表
// @Tags 商家端.门店区域
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{data=resp.CompanyAreaUnassignedResp}
// @Router /shop/company_area/unassigned [get]
func (h *CompanyAreaHandler) GetUnassigned(c *gin.Context) {
	ctx := helper.GetContext(c)
	result, err := h.companyAreaSrv.GetUnassigned(ctx)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeSystemError, err)
		return
	}
	helper.Success(c, result)
}

// GetShopList 获取区域管理用门店列表
// @Summary 获取区域管理用门店列表
// @Description 获取总部下所有门店（含总部自身），返回门店UUID、名称和区域归属信息，用于区域管理选择门店
// @Tags 商家端.门店区域
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{data=resp.CompanyAreaShopListResp}
// @Router /shop/company_area/shop_list [get]
func (h *CompanyAreaHandler) GetShopList(c *gin.Context) {
	ctx := helper.GetContext(c)
	result, err := h.companyAreaSrv.GetShopList(ctx)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeSystemError, err)
		return
	}
	helper.Success(c, result)
}
