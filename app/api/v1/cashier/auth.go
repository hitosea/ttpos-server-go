package cashier

import (
	"github.com/gin-gonic/gin"
	"ttpos-server-go/app/constant"

	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/service"
)

type AuthHandler struct {
	cashierAuthService *service.CashierAuthService
}

func NewCashierAuthHandler(cashierAuthService *service.CashierAuthService) *AuthHandler {
	return &AuthHandler{
		cashierAuthService: cashierAuthService,
	}
}

// Login 登录
// @Summary 登录
// @Tags 收银端
// @Access json
// @Produce json
// @Param X-SIGN header string true "验证码sign"
// @param data body req.CashierLoginRequest true "登录参数"
// @Success 200 {object} dto.Response
// @Router /cashier/passport/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var loginRequest req.CashierLoginRequest
	if err := c.ShouldBindJSON(&loginRequest); err != nil {
		helper.HandleValidationError(c, err, loginRequest, req.CashierLoginRequestMessage)
		return
	}
	sign := c.GetHeader("X-Sign")
	if sign == "" {
		helper.Fail(c, constant.CodeBadRequest, "验证码签名不能为空")
		return
	}
	token, err := h.cashierAuthService.Login(loginRequest, sign, loginRequest.Code)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeUnauthorized, err)
		return
	}
	helper.Success(c, gin.H{"token": token})
}

// Logout 退出登录
// @Summary 退出登录
// @Tags 收银端
// @Security JwtToken
// @Access json
// @Produce json
// @Success 200 {object} dto.Response
// @Router /cashier/passport/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	h.cashierAuthService.Logout(c.Copy())
	helper.Success(c, gin.H{})
}
