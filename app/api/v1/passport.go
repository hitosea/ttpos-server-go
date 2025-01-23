package v1

import (
	"github.com/gin-gonic/gin"

	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/service"
)

type PassportHandler struct {
	captchaService *service.CaptchaService
	pgpService     *service.PGPService
}

func NewPassportHandler(captchaService *service.CaptchaService, pgpService *service.PGPService) *PassportHandler {
	return &PassportHandler{
		captchaService: captchaService,
		pgpService:     pgpService,
	}
}

// GetCaptcha 获取验证码
// @Summary 获取验证码
// @Tags 通用
// @Access json
// @Produce json
// @Success 200 {object} dto.Response
// @Router /passport/captcha [get]
func (h *PassportHandler) GetCaptcha(c *gin.Context) {
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
// @param client_id query string false "客户端Id"
// @Success 200 {object} dto.Response
// @Router /passport/server-public-key [get]
func (h *PassportHandler) GetServerPublicKey(c *gin.Context) {
	var getKeyReq req.GetServerPublicKeyRequest
	if err := c.ShouldBindQuery(&getKeyReq); err != nil {
		helper.HandleValidationError(c, err, getKeyReq, req.GetServerPublicKeyRequestMessage)
		return
	}
	resp, err := h.pgpService.GetServerPublicKey(getKeyReq.ClientId)
	if err != nil {
		helper.Fail(c, constant.CodeFail, "获取服务端公钥失败")
		return
	}
	helper.Success(c, resp)
}
