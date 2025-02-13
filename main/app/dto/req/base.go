package req

type VerifyPasswordReq struct {
	Password string `json:"password" binding:"required"` // 密码
}
