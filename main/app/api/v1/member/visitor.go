package member

import (
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/service"
	"ttpos-server-go/pkg/database"

	"github.com/gin-gonic/gin"
)

// VisitorHandler 游客控制器
type VisitorHandler struct {
	memberSrv service.IMemberSrv
}

// Login 游客登录
// @Summary 游客登录
// @Description 游客登录，如果游客不存在则自动创建
// @Tags 会员端.游客
// @Accept json
// @Produce json
// @param data body req.VisitorLoginReq true "游客登录参数"
// @Success 200 {object} dto.Response{data=resp.VisitorInfoResp}
// @Router /member/visitor/login [post]
func (h *VisitorHandler) Login(c *gin.Context) {
	ctx := helper.GetContext(c)
	loginReq := req.VisitorLoginReq{}
	if err := c.ShouldBindJSON(&loginReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}

	if err := loginReq.Validate(); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}

	visitorInfo, err := h.memberSrv.VisitorLogin(ctx, loginReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}

	helper.Success(c, visitorInfo)
}

// BindPhone 游客绑定手机号
// @Summary 游客绑定手机号
// @Description 游客绑定手机号，转为正式会员
// @Tags 会员端.游客
// @Accept json
// @Produce json
// @param data body req.VisitorBindPhoneReq true "绑定手机号参数"
// @Success 200 {object} dto.Response
// @Router /member/visitor/bind_phone [post]
func (h *VisitorHandler) BindPhone(c *gin.Context) {
	ctx := helper.GetContext(c)
	bindReq := req.VisitorBindPhoneReq{}
	if err := c.ShouldBindJSON(&bindReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}

	if err := bindReq.Validate(); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}

	if err := h.memberSrv.BindPhone(ctx, bindReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}

	helper.Success(c, gin.H{
		"message": "绑定成功",
	})
}

// RegisterVisitorHandlers 注册游客路由
func RegisterVisitorHandlers(router gin.IRouter, dbm *database.DBManager) {
	memberSrv := service.NewMemberSrv(dbm)

	wrapper := &VisitorHandler{
		memberSrv: memberSrv,
	}

	visitorApi := router.Group("/visitor")
	{
		visitorApi.POST("/login", wrapper.Login)
		visitorApi.POST("/bind_phone", wrapper.BindPhone)
	}
}
