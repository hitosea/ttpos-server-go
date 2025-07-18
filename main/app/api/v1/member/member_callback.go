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
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /member/callback [post]
func (h *CallbackHandler) Callback(c *gin.Context) {
	// 绑定请求参数
	ctx := helper.GetContext(c)
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
	memberCallbackSrv := service.NewMemberCallbackSrv(dbm, service.NewLocaleSrv(), settingSrv)
	wrapper := &CallbackHandler{
		memberCallbackSrv: memberCallbackSrv,
	}
	// 需要认证
	privateApi := router.Group("")
	{
		privateApi.POST("/callback", wrapper.Callback) // 回调
	}
}
