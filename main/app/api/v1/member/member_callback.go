package member

import (
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req/member_req"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/service"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/utils"

	"github.com/duke-git/lancet/cryptor"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// CallbackHandler 回调控制器
type CallbackHandler struct {
	memberCallbackSrv service.IMemberCallbackSrv
}

// Callback 回调
// @Summary 回调
// @Description 回调
// @Tags 会员端.回调
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body member_req.CallbackReq true "详情参数"
// @Param companyUuid query string true "商家uuid"
// @param X-Ttpos-Callback-Auth header string true "回调token"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /member/order/callback [post]
func (h *CallbackHandler) Callback(c *gin.Context) {
	// 绑定请求参数
	ctx := helper.GetContext(c)
	companyUuidStr := c.Query("company_uuid")
	if companyUuidStr == "" {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.New("company_uuid不能为空"))
		return
	}
	companyUuid, err := utils.StringToUint64(companyUuidStr)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	ctx.SetCompanyUuid(companyUuid)
	callbackReq := member_req.CallbackReq{}
	if err := c.ShouldBindJSON(&callbackReq); err != nil {
		helper.HandleValidationError(c, err, callbackReq, nil)
		return
	}
	ctx.Log().Debug("回调", zap.Any("params", callbackReq))

	// 验证token. token格式： md5（订单号+secret）
	secret := config.JWT.Secret
	token := c.GetHeader("X-Ttpos-Callback-Auth")
	if token != cryptor.Md5String(callbackReq.ShopRefNo+secret) {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.New("token不合法"))
		return
	}
	// 回调
	if err := h.memberCallbackSrv.ParseMemberCallbackData(ctx, callbackReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, nil)
}

func RegisterMemberCallbackHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
	// 初始化服务
	settingSrv := setting.NewSrv(dbm, cache)
	// 初始化处理器
	orderSrv := service.NewOrderSrv(
		dbm,
		service.NewLocaleSrv(),
		settingSrv,
		service.NewMustPlanSrv(dbm),
		service.NewPaymentMethodSrv(dbm, settingSrv),
		service.NewMemberSrv(dbm, cache),
		service.NewCashBoxSrv(dbm),
		service.WithSmsSrv(dbm),
	)
	// 初始化处理器
	memberCallbackSrv := service.NewMemberCallbackSrv(dbm, service.NewLocaleSrv(), settingSrv, orderSrv)
	wrapper := &CallbackHandler{
		memberCallbackSrv: memberCallbackSrv,
	}
	// 需要认证
	privateApi := router.Group("")
	{
		privateApi.POST("/order/callback", wrapper.Callback) // 回调
	}
}
