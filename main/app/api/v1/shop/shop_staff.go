package shop

import (
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/service"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/middleware"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"

	"github.com/gin-gonic/gin"
)

// StaffHandler 员工管理
type StaffHandler struct {
	staffSrv     service.IStaffSrv
	saasStaffSrv service.ISaasStaffSrv
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
// @Param company_uuid query uint64 false "门店UUID"
// @Success 200 {object} dto.Response{data=resp.StaffListPaginationResp}
// @Router /shop/staff/list [get]
func (h *StaffHandler) GetStaffList(c *gin.Context) {
	ctx := helper.GetContext(c)
	var getStaffListReq req.GetStaffListReq
	if err := c.ShouldBindQuery(&getStaffListReq); err != nil {
		helper.HandleValidationError(c, err, getStaffListReq, req.GetStaffListReqReqMessage)
		return
	}
	var res resp.StaffListPaginationResp
	var err error
	// 2.11.0版本后，使用统一账号员工列表
	if ctx.Version(context.GTE, constant.ClientVersionV2110) {
		if getStaffListReq.CompanyUuid == 0 {
			helper.ErrorWithDetail(c, constant.CodeParamError, errors.New("参数错误"))
		}
		res, err = h.staffSrv.SaasPaginateGetStaffs(ctx, getStaffListReq)
	} else {
		res, err = h.staffSrv.PaginateGetStaffs(ctx, getStaffListReq)
	}
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeSystemError, err)
		return
	}
	helper.Success(c, res)
}

// SearchStaff 搜索员工
// @Summary 搜索员工
// @Description 搜索员工
// @Tags 商家端.员工账号
// @Accept json
// @Produce json
// @Security JwtToken
// @Param keyword query string false "关键词, 邮箱、手机号"
// @Success 200 {object} dto.Response{data=resp.SearchStaffResp}
// @Router /shop/staff/search [get]
func (h *StaffHandler) SearchStaff(c *gin.Context) {
	ctx := helper.GetContext(c)
	var searchStaffReq req.SearchStaffByKeywordReq
	if err := c.ShouldBindQuery(&searchStaffReq); err != nil {
		helper.HandleValidationError(c, err, searchStaffReq, nil)
		return
	}
	res := h.saasStaffSrv.SearchStaff(ctx, searchStaffReq.Keyword)

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
	var err error
	var exists []string
	// 2.11.0版本后，使用统一账号更新员工
	if ctx.Version(context.GTE, constant.ClientVersionV2110) {
		err, exists = h.staffSrv.SaasUpdateStaff(ctx, updateStaffReq)
	} else {
		err, exists = h.staffSrv.UpdateStaff(ctx, updateStaffReq)
	}
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
	var err error
	var exists []string

	// 2.11.0版本后，使用统一账号添加员工
	if ctx.Version(context.GTE, constant.ClientVersionV2110) {
		err, exists = h.staffSrv.SaasAddStaff(ctx, addStaffReq)
	} else {
		err, exists = h.staffSrv.AddStaff(ctx, addStaffReq)
	}
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

// GetStaffDetail 获取员工详情
// @Summary 获取员工详情
// @Description 根据员工uuid查询员工详情，包括员工在各门店的角色信息和当前门店的启用禁用状态
// @Tags 商家端.员工账号
// @Accept json
// @Produce json
// @Security JwtToken
// @Param uuid query uint64 true "员工UUID"
// @Success 200 {object} dto.Response{data=resp.Staff}
// @Router /shop/staff/detail [get]
func (h *StaffHandler) GetStaffDetail(c *gin.Context) {
	ctx := helper.GetContext(c)
	var getStaffDetailReq req.GetStaffDetailReq
	if err := c.ShouldBindQuery(&getStaffDetailReq); err != nil {
		helper.HandleValidationError(c, err, getStaffDetailReq, req.GetStaffDetailRequestMessage)
		return
	}
	res, err := h.staffSrv.GetStaffDetail(ctx, getStaffDetailReq.Uuid)
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
		staffSrv:     service.NewStaffSrv(dbm, cache, roleAccessSrv),
		saasStaffSrv: service.NewSaasStaffSrv(dbm),
	}

	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv, dbm))
	{
		// 员工账号
		privateApi.GET("/staff/list", wrapper.GetStaffList)     // 员工列表
		privateApi.GET("/staff/detail", wrapper.GetStaffDetail) // 获取员工详情
		privateApi.POST("/staff/add", wrapper.AddStaff)         // 添加员工
		privateApi.POST("/staff/update", wrapper.UpdateStaff)   // 编辑员工
		privateApi.GET("/role", wrapper.GetRoleList)            // 角色列表

		privateApi.GET("/staff/search", wrapper.SearchStaff) // 搜索员工
	}
}
