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

// RoleHandler 角色管理控制器
type RoleHandler struct {
	roleSrv       service.IRoleSrv
	roleAccessSrv service.IRoleAccessSrv
}

// GetRoleDetail 获取角色详情
// @Summary 获取角色详情
// @Description 获取角色详情（包含权限列表和关联员工）
// @Tags 商家端.员工账号
// @Accept json
// @Produce json
// @Security JwtToken
// @Param uuid query uint64 true "角色UUID"
// @Success 200 {object} dto.Response{data=resp.RoleDetailResp}
// @Router /shop/role/detail [get]
func (h *RoleHandler) GetRoleDetail(c *gin.Context) {
	ctx := helper.GetContext(c)
	var detailReq req.GetRoleDetailReq
	if err := c.ShouldBindQuery(&detailReq); err != nil {
		helper.HandleValidationError(c, err, detailReq, req.GetRoleDetailRequestMessage)
		return
	}

	res, err := h.roleSrv.GetRoleDetail(ctx, detailReq.Uuid)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeSystemError, err)
		return
	}

	helper.Success(c, res)
}

// CreateRole 创建角色
// @Summary 创建角色
// @Description 创建角色并配置权限（权限至少选择一个）
// @Tags 商家端.员工账号
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.AddRoleReq true "创建角色请求"
// @Success 200 {object} dto.Response{data=resp.Role}
// @Router /shop/role/create [post]
func (h *RoleHandler) CreateRole(c *gin.Context) {
	ctx := helper.GetContext(c)
	var createReq req.AddRoleReq
	if err := c.ShouldBindJSON(&createReq); err != nil {
		helper.HandleValidationError(c, err, createReq, req.AddRoleRequestMessage)
		return
	}

	res, err := h.roleSrv.CreateRole(ctx, createReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeSystemError, err)
		return
	}

	helper.Success(c, res, "创建角色成功")
}

// UpdateRole 更新角色
// @Summary 更新角色
// @Description 更新角色信息、权限和关联员工（权限至少选择一个）
// @Tags 商家端.员工账号
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.UpdateRoleReq true "更新角色请求"
// @Success 200 {object} dto.Response
// @Router /shop/role/update [post]
func (h *RoleHandler) UpdateRole(c *gin.Context) {
	ctx := helper.GetContext(c)
	var updateReq req.UpdateRoleReq
	if err := c.ShouldBindJSON(&updateReq); err != nil {
		helper.HandleValidationError(c, err, updateReq, req.UpdateRoleRequestMessage)
		return
	}

	err := h.roleSrv.UpdateRole(ctx, updateReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeSystemError, err)
		return
	}

	helper.Success(c, gin.H{}, "更新角色成功")
}

// DeleteRole 删除角色
// @Summary 删除角色
// @Description 删除角色（已关联员工的角色无法删除）
// @Tags 商家端.员工账号
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.DeleteRoleReq true "删除角色请求"
// @Success 200 {object} dto.Response
// @Router /shop/role/delete [delete]
func (h *RoleHandler) DeleteRole(c *gin.Context) {
	ctx := helper.GetContext(c)
	var deleteReq req.DeleteRoleReq
	if err := c.ShouldBindJSON(&deleteReq); err != nil {
		helper.HandleValidationError(c, err, deleteReq, req.DeleteRoleRequestMessage)
		return
	}

	err := h.roleSrv.DeleteRole(ctx, deleteReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeSystemError, err)
		return
	}

	helper.Success(c, gin.H{}, "删除角色成功")
}

// GetPermissions 获取权限组
// @Summary 获取权限组
// @Description 获取店铺的所有权限组（包含管理APP、收银机、点餐助手等所有权限组），用于角色权限配置
// @Tags 商家端.员工账号
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{data=resp.PermissionGroup}
// @Router /shop/permission_group [get]
func (h *RoleHandler) GetPermissionGroup(c *gin.Context) {
	ctx := helper.GetContext(c)

	permissionGroup, err := h.roleAccessSrv.GetCompanyPermissionGroup(ctx,
		[]constant.RouteName{constant.ShopAppRouteName, constant.CashierRouteName, constant.AssistantRouteName})
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeSystemError, err)
		return
	}

	helper.Success(c, permissionGroup)
}

// RegisterRoleHandlers 注册角色管理路由
func RegisterRoleHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
	// 初始化服务
	roleAccessSrv := service.NewRoleAccessSrv(dbm)
	roleSrv := service.NewRoleSrv(dbm, roleAccessSrv)

	wrapper := &RoleHandler{
		roleSrv:       roleSrv,
		roleAccessSrv: roleAccessSrv,
	}

	// 需要认证
	privateApi := router.Group("", middleware.Auth(service.NewAuthSrv(dbm, service.NewCaptchaSrv(cache), roleAccessSrv, nil, nil, nil), dbm))
	{
		privateApi.GET("/role/detail", wrapper.GetRoleDetail) // 获取角色详情
		privateApi.POST("/role/create", wrapper.CreateRole)   // 创建角色
		privateApi.POST("/role/update", wrapper.UpdateRole)   // 更新角色
		privateApi.DELETE("/role/delete", wrapper.DeleteRole) // 删除角色

		privateApi.GET("/permission_group", wrapper.GetPermissionGroup) // 获取权限组
	}
}
