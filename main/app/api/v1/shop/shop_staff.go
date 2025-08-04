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

// GetStaff 管理员列表
// @Summary 管理员列表
// @Description 管理员列表
// @Tags 移动管理端.管理员管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param page_no query int false "页码"
// @Param page_size query int false "每页条数"
// @Success 200 {object} dto.Response{data=resp.StaffListPaginationResp}
// @Router /shop/staff [get]
func (h *StaffHandler) GetStaff(c *gin.Context) {
	ctx := helper.GetContext(c)
	var pageReq dto.PageReq
	if err := c.ShouldBindQuery(&pageReq); err != nil {
		helper.HandleValidationError(c, err, pageReq, dto.PageReqMessage)
		return
	}
	res, err := h.staffSrv.GetStaffs(ctx, pageReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeSystemError, err)
		return
	}
	helper.Success(c, res)
}

// UpdateStaff 修改管理员
// @Summary 修改管理员
// @Description 修改管理员
// @Tags 移动管理端.管理员管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param update_staff_req body req.UpdateStaffReq true "修改管理员请求"
// @Success 200 {object} dto.Response
// @Router /shop/staff [post]
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

// UpdateStaffStatus 设置启用禁用员工
// @Summary 设置启用禁用员工
// @Description 设置启用禁用员工
// @Tags 移动管理端.管理员管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param update_staff_status_req body req.UpdateStaffStatusReq true "设置启用禁用员工请求"
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

// DeleteStaff 删除员工
// @Summary 删除员工
// @Description 删除员工
// @Tags 移动管理端.管理员管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param delete_staff_req body req.DeleteStaffReq true "删除员工请求"
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
		staffSrv: service.NewStaffSrv(dbm),
	}

	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv, dbm))
	{
		// 员工管理
		privateApi.GET("/staff", wrapper.GetStaff)                  // 获取员工列表
		privateApi.POST("/staff", wrapper.UpdateStaff)              // 修改员工
		privateApi.POST("/staff/status", wrapper.UpdateStaffStatus) // 设置启用禁用员工
		privateApi.DELETE("/staff", wrapper.DeleteStaff)            // 删除员工
	}
}
