package shop

import (
	"context"
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/service"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/middleware"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/utils"

	"github.com/gin-gonic/gin"
)

// MiscHandler 杂项控制器
type MiscHandler struct {
	authSrv      service.IAuthSrv
	checkNameSrv service.ICheckNameSrv
}

// AiTranslate 翻译
// @Summary 翻译
// @Tags 通用
// @Access json
// @Produce json
// @Security JwtToken
// @param data body req.AiTranslateRequest true "翻译数据"
// @Success 200 {object} dto.Response{data=[]utils.TranslateResponseItem}
// @Router /shop/ai/translate [post]
func (h *MiscHandler) AiTranslate(c *gin.Context) {
	var aiTranslateReq req.AiTranslateRequest
	if err := c.ShouldBind(&aiTranslateReq); err != nil {
		helper.HandleValidationError(c, err, aiTranslateReq, nil)
		return
	}
	res, err := utils.NewTranslateClient().Translate(context.Background(), []utils.TranslateItem{
		{
			Lang:    aiTranslateReq.Language,
			Content: aiTranslateReq.Text,
		},
	})
	if err != nil {
		helper.Fail(c, constant.CodeFail, err.Error())
		return
	}
	helper.Success(c, res.Data[0])
}

// CheckNameExists 检查名称是否存在
// @Summary 检查名称是否存在
// @Tags 通用
// @Access json
// @Produce json
// @Security JwtToken
// @param data body req.CheckNameRequest true "检查名称是否存在"
// @Success 200 {object} dto.Response{data=resp.CheckNameResp}
// @Router /shop/check/name/exists [post]
func (h *MiscHandler) CheckNameExists(c *gin.Context) {
	ctx := helper.GetContext(c)
	var checkNameReq req.CheckNameRequest
	if err := c.ShouldBind(&checkNameReq); err != nil {
		helper.HandleValidationError(c, err, checkNameReq, nil)
		return
	}
	res, err := h.checkNameSrv.CheckNameExists(ctx, checkNameReq)
	if err != nil {
		helper.Fail(c, constant.CodeFail, err.Error())
		return
	}
	helper.Success(c, res)
}

func RegisterMiscHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
	// 初始化服务
	captchaSrv := service.NewCaptchaSrv(cache)
	settingSrv := setting.NewSrv(dbm, cache)
	roleAccessSrv := service.NewRoleAccessSrv(dbm)
	deviceSrv := service.NewDeviceSrv(settingSrv, dbm)
	cashBoxSrv := service.NewCashBoxSrv(dbm)
	statisticsSrv := service.NewStatisticsSrv()
	staffShiftSrv := service.NewStaffShiftSrv(cache, dbm, cashBoxSrv, statisticsSrv)
	authSrv := service.NewAuthSrv(dbm, captchaSrv, roleAccessSrv, deviceSrv, staffShiftSrv, settingSrv)

	wrapper := &MiscHandler{
		authSrv:      authSrv,
		checkNameSrv: service.NewCheckNameSrv(dbm),
	}

	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv, dbm))
	{
		// 翻译
		privateApi.POST("/ai/translate", wrapper.AiTranslate)
		// 检查名称是否存在
		privateApi.POST("/check/name/exists", wrapper.CheckNameExists)
	}
}
