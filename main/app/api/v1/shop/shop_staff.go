package shop

import (
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/service"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/middleware"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"

	"github.com/gin-gonic/gin"
)

// StaffHandler 管理员管理
type StaffHandler struct {
	staffSrv service.IStaffSrv
}

// GetStaffList 管理员列表
// @Summary 管理员列表
// @Description 管理员列表
// @Tags 商家端.管理员管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param page_no query int false "页码"
// @Param page_size query int false "每页条数"
// @Success 200 {object} dto.Response{data=resp.StaffListPaginationResp}
// @Router /shop/staff/list [get]
func (h *StaffHandler) GetStaffList(c *gin.Context) {
	ctx := helper.GetContext(c)
	var pageReq dto.PageReq
	if err := c.ShouldBindQuery(&pageReq); err != nil {
		helper.HandleValidationError(c, err, pageReq, dto.PageReqMessage)
		return
	}
	res, err := h.staffSrv.PaginateGetStaffs(ctx, pageReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeSystemError, err)
		return
	}
	helper.Success(c, res)
}

// UpdateStaff 修改管理员
// @Summary 修改管理员
// @Description 修改管理员
// @Tags 商家端.管理员管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param update_staff_req body req.UpdateStaffReq true "修改管理员请求"
// @Success 200 {object} dto.Response
// @Router /shop/staff/update [post]
func (h *StaffHandler) UpdateStaff(c *gin.Context) {
	ctx := helper.GetContext(c)
	var updateStaffReq req.UpdateStaffReq
	if err := c.ShouldBindJSON(&updateStaffReq); err != nil {
		helper.HandleValidationError(c, err, updateStaffReq, nil)
		return
	}
	err := h.staffSrv.UpdateStaff(ctx, updateStaffReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeSystemError, err)
		return
	}
	helper.Success(c, nil)
}

// UpdateStaffStatus 设置启用禁用管理员
// @Summary 设置启用禁用管理员
// @Description 设置启用禁用管理员
// @Tags 商家端.管理员管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param update_staff_status_req body req.UpdateStaffStatusReq true "设置启用禁用管理员请求"
// @Success 200 {object} dto.Response
// @Router /shop/staff/status [post]
func (h *StaffHandler) UpdateStaffStatus(c *gin.Context) {
	ctx := helper.GetContext(c)
	var updateStaffStatusReq req.UpdateStaffStatusReq
	if err := c.ShouldBindJSON(&updateStaffStatusReq); err != nil {
		helper.HandleValidationError(c, err, updateStaffStatusReq, nil)
		return
	}
	err := h.staffSrv.UpdateStaffStatus(ctx, updateStaffStatusReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeSystemError, err)
		return
	}
	helper.Success(c, nil)
}

// DeleteStaff 删除管理员
// @Summary 删除管理员
// @Description 删除管理员
// @Tags 商家端.管理员管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param delete_staff_req body req.DeleteStaffReq true "删除管理员请求"
// @Success 200 {object} dto.Response
// @Router /shop/staff [delete]
func (h *StaffHandler) DeleteStaff(c *gin.Context) {
	ctx := helper.GetContext(c)
	var deleteStaffReq req.DeleteStaffReq
	if err := c.ShouldBindJSON(&deleteStaffReq); err != nil {
		helper.HandleValidationError(c, err, deleteStaffReq, nil)
		return
	}
	err := h.staffSrv.DeleteStaff(ctx, deleteStaffReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeSystemError, err)
		return
	}
	helper.Success(c, nil)
}

// AddStaff 添加管理员
// @Summary 添加管理员
// @Description 添加管理员
// @Tags 商家端.管理员管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param add_staff_req body req.AddStaffReq true "添加管理员请求"
// @Success 200 {object} dto.Response
// @Router /shop/staff/add [post]
func (h *StaffHandler) AddStaff(c *gin.Context) {
	ctx := helper.GetContext(c)
	var addStaffReq req.AddStaffReq
	if err := c.ShouldBindJSON(&addStaffReq); err != nil {
		helper.HandleValidationError(c, err, addStaffReq, nil)
		return
	}
	err := h.staffSrv.AddStaff(ctx, addStaffReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeSystemError, err)
		return
	}
	helper.Success(c, nil)
}

// GetRoleList 获取角色列表
// @Summary 获取角色列表
// @Description 获取角色列表
// @Tags 商家端.管理员管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param page_no query int false "页码"
// @Param page_size query int false "每页条数"
// @Success 200 {object} dto.Response{data=resp.RoleListResp}
// @Router /shop/role [get]
func (h *StaffHandler) GetRoleList(c *gin.Context) {
	ctx := helper.GetContext(c)
	var pageReq dto.PageReq
	if err := c.ShouldBindQuery(&pageReq); err != nil {
		helper.HandleValidationError(c, err, pageReq, dto.PageReqMessage)
		return
	}
	res, err := h.staffSrv.GetRoleList(ctx, pageReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeSystemError, err)
		return
	}
	helper.Success(c, res)
}

// AddRole 添加角色
// @Summary 添加角色
// @Description 添加角色
// @Tags 商家端.管理员管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param add_role_req body req.AddRoleReq true "添加角色请求"
// @Success 200 {object} dto.Response
// @Router /shop/role/add [post]
func (h *StaffHandler) AddRole(c *gin.Context) {
	ctx := helper.GetContext(c)
	var addRoleReq req.AddRoleReq
	if err := c.ShouldBindJSON(&addRoleReq); err != nil {
		helper.HandleValidationError(c, err, addRoleReq, nil)
		return
	}
	err := h.staffSrv.AddRole(ctx, addRoleReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeSystemError, err)
		return
	}
	helper.Success(c, nil)
}

// UpdateRole 修改角色
// @Summary 修改角色
// @Description 修改角色
// @Tags 商家端.管理员管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param update_role_req body req.UpdateRoleReq true "修改角色请求"
// @Success 200 {object} dto.Response
// @Router /shop/role/update [post]
func (h *StaffHandler) UpdateRole(c *gin.Context) {
	ctx := helper.GetContext(c)
	var updateRoleReq req.UpdateRoleReq
	if err := c.ShouldBindJSON(&updateRoleReq); err != nil {
		helper.HandleValidationError(c, err, updateRoleReq, nil)
		return
	}
	err := h.staffSrv.UpdateRole(ctx, updateRoleReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeSystemError, err)
		return
	}
	helper.Success(c, nil)
}

// DeleteRole 删除角色
// @Summary 删除角色
// @Description 删除角色
// @Tags 商家端.管理员管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param delete_role_req body req.DeleteRoleReq true "删除角色请求"
// @Success 200 {object} dto.Response
// @Router /shop/role [delete]
func (h *StaffHandler) DeleteRole(c *gin.Context) {
	ctx := helper.GetContext(c)
	var deleteRoleReq req.DeleteRoleReq
	if err := c.ShouldBindJSON(&deleteRoleReq); err != nil {
		helper.HandleValidationError(c, err, deleteRoleReq, nil)
		return
	}
	err := h.staffSrv.DeleteRole(ctx, deleteRoleReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeSystemError, err)
		return
	}
	helper.Success(c, nil)
}

// GetRoleAccess 获取角色权限
// @Summary 获取角色权限
// @Description 获取角色权限
// @Tags 商家端.管理员管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param get_role_req query req.GetRoleReq true "获取角色详细请求"
// @Success 200 {object} dto.Response{data=resp.RoleDetailResp}
// @Router /shop/role_access [get]
func (h *StaffHandler) GetRoleAccess(c *gin.Context) {
	ctx := helper.GetContext(c)
	var getRoleReq req.GetRoleReq
	if err := c.ShouldBindQuery(&getRoleReq); err != nil {
		helper.HandleValidationError(c, err, getRoleReq, nil)
		return
	}
	res, err := h.staffSrv.GetRoleAccess(ctx, getRoleReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeSystemError, err)
		return
	}
	helper.Success(c, res)
}

// GetPermissionGroup 获取权限组
// @Summary 获取权限组
// @Description 获取权限组
// @Tags 商家端.管理员管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response
// @Router /shop/access_groups [get]
func (h *StaffHandler) GetPermissionGroup(c *gin.Context) {
	ctx := helper.GetContext(c)
	res, err := h.staffSrv.GetPermissionGroup(ctx)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeSystemError, err)
		return
	}
	helper.Success(c, res)
}

func RegisterStaffHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
	// 初始化服务
	captchaSrv := service.NewCaptchaSrv(cache)
	settingSrv := setting.NewSrv(dbm, cache)
	roleAccessSrv := service.NewRoleAccessSrv(dbm)
	deviceSrv := service.NewDeviceSrv(settingSrv, dbm)
	cashBoxSrv := service.NewCashBoxSrv(dbm)
	statisticsSrv := service.NewStatisticsSrv()
	staffShiftSrv := service.NewStaffShiftSrv(cache, dbm, cashBoxSrv, statisticsSrv)
	authSrv := service.NewAuthSrv(dbm, captchaSrv, roleAccessSrv, deviceSrv, staffShiftSrv, settingSrv)

	wrapper := &StaffHandler{
		staffSrv: service.NewStaffSrv(dbm, cache, roleAccessSrv),
	}

	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv, dbm))
	{
		// 管理员管理
		privateApi.GET("/staff/list", wrapper.GetStaffList)         // 获取管理员列表
		privateApi.POST("/staff/update", wrapper.UpdateStaff)       // 修改管理员
		privateApi.POST("/staff/status", wrapper.UpdateStaffStatus) // 设置启用禁用管理员
		privateApi.DELETE("/staff", wrapper.DeleteStaff)            // 删除管理员
		privateApi.POST("/staff/add", wrapper.AddStaff)             // 添加管理员

		// 获取角色列表
		privateApi.GET("/role", wrapper.GetRoleList)          // 获取角色列表/下拉选择
		privateApi.POST("/role/add", wrapper.AddRole)         // 添加角色
		privateApi.POST("/role/update", wrapper.UpdateRole)   // 修改角色
		privateApi.DELETE("/role", wrapper.DeleteRole)        // 删除角色
		privateApi.GET("/role_access", wrapper.GetRoleAccess) // 获取角色详细

		privateApi.GET("/access_groups", wrapper.GetPermissionGroup) // 获取权限组
	}
}
