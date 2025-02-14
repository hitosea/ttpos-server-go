package cashier

import (
	"strconv"
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/constant/jwt"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/service"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/middleware"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"

	"github.com/gin-gonic/gin"
)

// MemberHandler 结构体
type MemberHandler struct {
	memberSrv service.IMemberSrv
}

// GetMemberLevels 会员等级列表
// @Summary 会员等级列表
// @Description 会员等级列表
// @Tags 收银端.会员
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{data=resp.MemberLevelList}
// @Router /cashier/member/levels [get]
func (h *MemberHandler) GetMemberLevels(c *gin.Context) {
	helper.Success(c, h.memberSrv.GetLevels(c.GetUint64(jwt.CompanyUuid)))
}

// SearchMember 模糊搜索会员
// @Summary 模糊搜索会员
// @Description 模糊搜索会员
// @Tags 收银端.会员
// @Accept json
// @Produce json
// @Security JwtToken
// @param keyword query string false "关键字搜索：uuid\phone，前端处理前后空格"
// @Success 200 {object} dto.Response{data=resp.SearchMemberList}
// @Router /cashier/member/search [get]
func (h *MemberHandler) SearchMember(c *gin.Context) {
	helper.Success(c, h.memberSrv.SearchMember(c.GetUint64(jwt.CompanyUuid), c.Query("keyword")))
}

// RechargeMember 充值会员信息
// @Summary 充值会员信息
// @Description 充值会员信息
// @Tags 收银端.会员
// @Accept json
// @Produce json
// @Security JwtToken
// @param uuid query number true "uuid"
// @Success 200 {object} dto.Response{data=resp.RechargeMember}
// @Router /cashier/member/recharge_member [get]
func (h *MemberHandler) RechargeMember(c *gin.Context) {
	uuid, err := strconv.ParseUint(c.Query("uuid"), 10, 64)
	if err != nil {
		helper.Fail(c, constant.CodeBadRequest, "参数错误")
	}
	info, err := h.memberSrv.GetRechargeMember(c.GetUint64(jwt.CompanyUuid), uuid)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
	}
	helper.Success(c, info)
}

// AddMember 添加会员
// @Summary 添加会员
// @Description 添加会员
// @Tags 收银端.会员
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.AddMemberReq true "添加会员参数"
// @Success 200 {object} dto.Response
// @Router /cashier/member/add [post]
func (h *MemberHandler) AddMember(c *gin.Context) {
	var addMemberReq req.AddMemberReq
	if err := c.ShouldBindJSON(&addMemberReq); err != nil {
		helper.HandleValidationError(c, err, addMemberReq, req.AddMemberReqMessage)
		return
	}
	if err := h.memberSrv.AddMember(c.GetUint64(jwt.CompanyUuid), addMemberReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	helper.Success(c, gin.H{}, "添加会员成功")
}

func RegisterMemberHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
	// 初始化服务
	captchaSrv := service.NewCaptchaSrv(cache)
	settingSrv := setting.NewSrv(dbm, cache)
	roleAccessSrv := service.NewRoleAccessSrv(dbm)
	bindRecordSrv := service.NewBindRecordSrv(settingSrv, dbm)
	staffShiftSrv := service.NewStaffShiftSrv(cache, dbm)
	authSrv := service.NewAuthSrv(dbm, captchaSrv, roleAccessSrv, bindRecordSrv, staffShiftSrv, settingSrv)

	memberSrv := service.NewMemberSrv(dbm)

	wrapper := &MemberHandler{
		memberSrv: memberSrv,
	}

	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv))
	{
		privateApi.GET("/member/levels", wrapper.GetMemberLevels)       // 获取会员等级列表
		privateApi.GET("/member/search", wrapper.SearchMember)          // 模糊搜索会员
		privateApi.GET("/member/charge_member", wrapper.RechargeMember) // 充值会员信息

		privateApi.POST("/member/add", wrapper.AddMember) // 添加会员
	}
}
