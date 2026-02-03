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

// MaterialCategoryVisibilityHandler 物品分类可见性配置控制器
type MaterialCategoryVisibilityHandler struct {
	visibilitySrv service.IMaterialCategoryVisibilitySrv
}

// GetList 获取可见性配置列表
// @Summary 获取可见性配置列表
// @Description 获取物品分类可见性配置列表（仅总店可用）
// @Tags 商家端.物品管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{data=resp.VisibilityListResp}
// @Router /shop/material_category_visibility/list [get]
func (h *MaterialCategoryVisibilityHandler) GetList(c *gin.Context) {
	ctx := helper.GetContext(c)

	res, err := h.visibilitySrv.GetList(ctx)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeSystemError, err)
		return
	}

	helper.Success(c, res)
}

// GetDetail 获取可见性配置详情
// @Summary 获取可见性配置详情
// @Description 获取物品分类可见性配置详情（仅总店可用）
// @Tags 商家端.物品管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param uuid query uint64 true "配置UUID"
// @Success 200 {object} dto.Response{data=resp.VisibilityDetailResp}
// @Router /shop/material_category_visibility/detail [get]
func (h *MaterialCategoryVisibilityHandler) GetDetail(c *gin.Context) {
	ctx := helper.GetContext(c)
	var detailReq req.VisibilityDetailReq
	if err := c.ShouldBindQuery(&detailReq); err != nil {
		helper.HandleValidationError(c, err, detailReq, nil)
		return
	}

	res, err := h.visibilitySrv.GetDetail(ctx, detailReq.Uuid)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeSystemError, err)
		return
	}

	helper.Success(c, res)
}

// Create 创建可见性配置
// @Summary 创建可见性配置
// @Description 创建物品分类可见性配置（仅总店可用，创建后自动同步到子店）
// @Tags 商家端.物品管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.VisibilityCreateReq true "创建请求"
// @Success 200 {object} dto.Response
// @Router /shop/material_category_visibility/create [post]
func (h *MaterialCategoryVisibilityHandler) Create(c *gin.Context) {
	ctx := helper.GetContext(c)
	var createReq req.VisibilityCreateReq
	if err := c.ShouldBindJSON(&createReq); err != nil {
		helper.HandleValidationError(c, err, createReq, nil)
		return
	}

	err := h.visibilitySrv.Create(ctx, createReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeSystemError, err)
		return
	}

	helper.Success(c, gin.H{}, "创建成功")
}

// Update 更新可见性配置
// @Summary 更新可见性配置
// @Description 更新物品分类可见性配置（仅总店可用，更新后自动同步到子店）
// @Tags 商家端.物品管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.VisibilityUpdateReq true "更新请求"
// @Success 200 {object} dto.Response
// @Router /shop/material_category_visibility/update [post]
func (h *MaterialCategoryVisibilityHandler) Update(c *gin.Context) {
	ctx := helper.GetContext(c)
	var updateReq req.VisibilityUpdateReq
	if err := c.ShouldBindJSON(&updateReq); err != nil {
		helper.HandleValidationError(c, err, updateReq, nil)
		return
	}

	err := h.visibilitySrv.Update(ctx, updateReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeSystemError, err)
		return
	}

	helper.Success(c, gin.H{}, "更新成功")
}

// Delete 删除可见性配置
// @Summary 删除可见性配置
// @Description 删除物品分类可见性配置（仅总店可用，删除后自动同步到子店）
// @Tags 商家端.物品管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.VisibilityDeleteReq true "删除请求"
// @Success 200 {object} dto.Response
// @Router /shop/material_category_visibility/delete [delete]
func (h *MaterialCategoryVisibilityHandler) Delete(c *gin.Context) {
	ctx := helper.GetContext(c)
	var deleteReq req.VisibilityDeleteReq
	if err := c.ShouldBindJSON(&deleteReq); err != nil {
		helper.HandleValidationError(c, err, deleteReq, nil)
		return
	}

	err := h.visibilitySrv.Delete(ctx, deleteReq.Uuid)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeSystemError, err)
		return
	}

	helper.Success(c, gin.H{}, "删除成功")
}

// GetSubShopRoles 获取子店角色列表
// @Summary 获取子店角色列表
// @Description 获取所有子店的角色列表（仅总店可用，用于配置可见性规则）
// @Tags 商家端.物品管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{data=resp.SubShopRolesResp}
// @Router /shop/material_category_visibility/sub_shop_roles [get]
func (h *MaterialCategoryVisibilityHandler) GetSubShopRoles(c *gin.Context) {
	ctx := helper.GetContext(c)

	res, err := h.visibilitySrv.GetSubShopRoles(ctx)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeSystemError, err)
		return
	}

	helper.Success(c, res)
}

// RegisterMaterialCategoryVisibilityHandlers 注册物品分类可见性配置路由
func RegisterMaterialCategoryVisibilityHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
	// 初始化服务
	roleAccessSrv := service.NewRoleAccessSrv(dbm)
	visibilitySrv := service.NewMaterialCategoryVisibilitySrv(dbm)

	wrapper := &MaterialCategoryVisibilityHandler{
		visibilitySrv: visibilitySrv,
	}

	// 需要认证
	privateApi := router.Group("", middleware.Auth(service.NewAuthSrv(dbm, service.NewCaptchaSrv(cache), roleAccessSrv, nil, nil, nil), dbm))
	{
		privateApi.GET("/material_category_visibility/list", wrapper.GetList)                   // 获取配置列表
		privateApi.GET("/material_category_visibility/detail", wrapper.GetDetail)               // 获取配置详情
		privateApi.POST("/material_category_visibility/create", wrapper.Create)                 // 创建配置
		privateApi.POST("/material_category_visibility/update", wrapper.Update)                 // 更新配置
		privateApi.DELETE("/material_category_visibility/delete", wrapper.Delete)               // 删除配置
		privateApi.GET("/material_category_visibility/sub_shop_roles", wrapper.GetSubShopRoles) // 获取子店角色
	}
}
