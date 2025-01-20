package resp

type CaptchaResponse struct {
	Sign   string `json:"sign"`   // 验证码标识
	Base64 string `json:"base64"` // Base64编码的图片
}
