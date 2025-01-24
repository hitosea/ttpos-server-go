package cashier

import (
	"ttpos-server-go/app/constant"

	"github.com/gin-gonic/gin"

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
// @Success 200 {object} dto.Response{data=resp.CashierLoginResponse}
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
	data, err := h.cashierAuthService.Login(loginRequest.Username, loginRequest.Password, sign, loginRequest.Code)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeUnauthorized, err)
		return
	}
	helper.Success(c, data)
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

// Unlock 解锁
// @Summary 解锁
// @Tags 收银端
// @Security JwtToken
// @Access json
// @Produce json
// @Param data body req.ReqCashierUnlock true "解锁参数"
// @Success 200 {object} dto.Response
// @Router /cashier/passport/unlock [post]
func (h *AuthHandler) Unlock(c *gin.Context) {
	//
}
