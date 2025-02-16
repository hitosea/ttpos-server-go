package passport

import (
	"github.com/gin-gonic/gin"
	"ttpos-server-go/pkg/cache"

	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/service"
)

type Handler struct {
	captchaService service.ICaptchaSrv
	encryptService service.IEncryptSrv
}

// GetCaptcha 获取验证码
// @Summary 获取验证码
// @Tags 通用
// @Access json
// @Produce json
// @Success 200 {object} dto.Response
// @Router /passport/captcha [get]
func (h *Handler) GetCaptcha(c *gin.Context) {
	captcha, err := h.captchaService.Generate()
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
	resp, err := h.encryptService.GetServerPublicKey(getKeyReq.ClientId, getKeyReq.Type)
	if err != nil {
		helper.Fail(c, constant.CodeFail, "获取服务端公钥失败")
		return
	}
	helper.Success(c, resp)
}

func RegisterHandlers(router gin.IRouter, cache cache.Cache) {
	wrapper := &Handler{
		captchaService: service.NewCaptchaSrv(cache),
		encryptService: service.NewEncryptSrv(cache),
	}
	router.GET("/captcha", wrapper.GetCaptcha)
	router.GET("/server_public_key", wrapper.GetServerPublicKey)
}
