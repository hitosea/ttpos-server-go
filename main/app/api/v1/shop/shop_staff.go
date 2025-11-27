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

// StaffHandler 员工管理
type StaffHandler struct {
	staffSrv service.IStaffSrv
}

// GetStaffList 员工列表
// @Summary 员工列表
// @Description 员工列表
// @Tags 商家端.员工账号
// @Accept json
// @Produce json
// @Security JwtToken
// @Param page_no query int false "页码"
// @Param page_size query int false "每页条数"
// @Param is_filter_super query int false "是否过滤超级管理员"
// @Param keyword query string false "关键词, 姓名、邮箱、手机号"
// @Success 200 {object} dto.Response{data=resp.StaffListPaginationResp}
// @Router /shop/staff/list [get]
func (h *StaffHandler) GetStaffList(c *gin.Context) {
	ctx := helper.GetContext(c)
	var getStaffListReq req.GetStaffListReq
	if err := c.ShouldBindQuery(&getStaffListReq); err != nil {
		helper.HandleValidationError(c, err, getStaffListReq, req.GetStaffListReqReqMessage)
		return
	}
	res, err := h.staffSrv.PaginateGetStaffs(ctx, getStaffListReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeSystemError, err)
		return
	}
	helper.Success(c, res)
}

// UpdateStaff 编辑员工
// @Summary 编辑员工
// @Description 编辑员工
// @Tags 商家端.员工账号
// @Accept json
// @Produce json
// @Security JwtToken
// @Param update_staff_req body req.UpdateStaffReq true "编辑员工请求"
// @Success 200 {object} dto.Response
// @Router /shop/staff/update [post]
func (h *StaffHandler) UpdateStaff(c *gin.Context) {
	ctx := helper.GetContext(c)
	var updateStaffReq req.UpdateStaffReq
	if err := c.ShouldBindJSON(&updateStaffReq); err != nil {
		helper.HandleValidationError(c, err, updateStaffReq, nil)
		return
	}
	// 版本兼容性验证：低于 2.10.0 版本且权限密码为空时，设置默认值
	if err := updateStaffReq.Validate(ctx); err != nil {
		helper.ErrorWithDetail(c, constant.CodeParamError, err)
		return
	}
	err, exists := h.staffSrv.UpdateStaff(ctx, updateStaffReq)
	if err != nil {
		helper.ErrorWithData(c, constant.CodeSystemError, gin.H{
			"exists": exists,
		}, err)
		return
	}
	helper.Success(c, nil)
}

// AddStaff 添加员工
// @Summary 添加员工
// @Description 添加员工
// @Tags 商家端.员工账号
// @Accept json
// @Produce json
// @Security JwtToken
// @Param add_staff_req body req.AddStaffReq true "添加员工请求"
// @Success 200 {object} dto.Response
// @Router /shop/staff/add [post]
func (h *StaffHandler) AddStaff(c *gin.Context) {
	ctx := helper.GetContext(c)
	var addStaffReq req.AddStaffReq
	if err := c.ShouldBindJSON(&addStaffReq); err != nil {
		helper.HandleValidationError(c, err, addStaffReq, req.AddStaffRequestMessage)
		return
	}
	// 版本兼容性验证：低于 2.10.0 版本且权限密码为空时，设置默认值
	if err := addStaffReq.Validate(ctx); err != nil {
		helper.ErrorWithDetail(c, constant.CodeParamError, err)
		return
	}
	err, exists := h.staffSrv.AddStaff(ctx, addStaffReq)
	if err != nil {
		helper.ErrorWithData(c, constant.CodeSystemError, gin.H{
			"exists": exists,
		}, err)
		return
	}
	helper.Success(c, gin.H{})
}

// GetRoleList 获取角色列表
// @Summary 获取角色列表
// @Description 获取角色列表
// @Tags 商家端.员工账号
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
		// 员工账号
		privateApi.GET("/staff/list", wrapper.GetStaffList)   // 员工列表
		privateApi.POST("/staff/add", wrapper.AddStaff)       // 添加员工
		privateApi.POST("/staff/update", wrapper.UpdateStaff) // 编辑员工
		privateApi.GET("/role", wrapper.GetRoleList)          // 角色列表
	}
}
