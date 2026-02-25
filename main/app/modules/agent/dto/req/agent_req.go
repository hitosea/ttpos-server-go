package req

// ChatReq 对话请求
type ChatReq struct {
	Message string `json:"message" binding:"required"` // 用户消息
}
