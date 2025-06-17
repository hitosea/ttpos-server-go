package req

// DecryptQrCodeReq 解密活动二维码请求
type DecryptQrCodeReq struct {
	Sign string `json:"sign" binding:"required"` // 活动二维码sign
}
