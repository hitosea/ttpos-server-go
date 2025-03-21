package passport

import (
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"

	"github.com/gin-gonic/gin"

	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/service"
)

type Handler struct {
	dbm        *database.DBManager
	captchaSrv service.ICaptchaSrv
	encryptSrv service.IEncryptSrv
}

// GetCaptcha 获取验证码
// @Summary 获取验证码
// @Tags 通用
// @Access json
// @Produce json
// @Success 200 {object} dto.Response
// @Router /passport/captcha [get]
func (h *Handler) GetCaptcha(c *gin.Context) {
	captcha, err := h.captchaSrv.Generate()
	if err != nil {
		helper.Fail(c, constant.CodeFail, "生成验证码失败")
		return
	}
	helper.Success(c, captcha)
}

// GetServerPublicKey 获取服务端公钥
// @Summary 获取服务端公钥
// @Tags 通用
// @Access json
// @Produce json
// @param client_id query string true "客户端Id"
// @param type query string true "加密类型: jsencrypt"
// @Success 200 {object} dto.Response
// @Router /passport/server_public_key [get]
func (h *Handler) GetServerPublicKey(c *gin.Context) {
	var getKeyReq req.GetServerPublicKeyRequest
	if err := c.ShouldBindQuery(&getKeyReq); err != nil {
		helper.HandleValidationError(c, err, getKeyReq, req.GetServerPublicKeyRequestMessage)
		return
	}
	resp, err := h.encryptSrv.GetServerPublicKey(getKeyReq.ClientId, getKeyReq.Type)
	if err != nil {
		helper.Fail(c, constant.CodeFail, "获取服务端公钥失败")
		return
	}
	helper.Success(c, resp)
}

// LianLianCallback 连连支付回调
func (h *Handler) LianLianCallback(c *gin.Context) {
	ctx := helper.GetContext(c)
	sign := c.GetHeader("sign")
	//
	var callbackReq req.LianLianCallbackRequest
	if err := c.ShouldBind(&callbackReq); err != nil {
		helper.HandleValidationError(c, err, callbackReq, req.GetServerPublicKeyRequestMessage)
		return
	}
	// 处理支付回调
	err := service.NewPaymentRepo(ctx, h.dbm).HandleCallback(sign, callbackReq)
	if err != nil {
		helper.Fail(c, constant.CodeFail, err.Error())
		return
	}

	c.String(200, "success")
}

// LianLianRefundCallback 连连支付退款回调
func (h *Handler) LianLianRefundCallback(c *gin.Context) {
	ctx := helper.GetContext(c)
	sign := c.GetHeader("sign")
	//
	var callbackReq req.LianLianRefundCallbackRequest
	if err := c.ShouldBind(&callbackReq); err != nil {
		helper.HandleValidationError(c, err, callbackReq, req.GetServerPublicKeyRequestMessage)
		return
	}
	callbackReq.MerchantRefundOrderNo = c.Query("merchant_refund_id")

	// 处理支付回调
	err := service.NewPaymentRepo(ctx, h.dbm).HandleRefundCallback(sign, callbackReq)
	if err != nil {
		helper.Fail(c, constant.CodeFail, err.Error())
		return
	}

	c.String(200, "success")
}

func RegisterHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
	wrapper := &Handler{
		dbm:        dbm,
		captchaSrv: service.NewCaptchaSrv(cache),
		encryptSrv: service.NewEncryptSrv(cache),
	}
	router.GET("/captcha", wrapper.GetCaptcha)
	router.GET("/server_public_key", wrapper.GetServerPublicKey)
	router.POST("/lianlian/callback", wrapper.LianLianCallback)
	router.POST("/lianlian/refund/callback", wrapper.LianLianRefundCallback)
}
