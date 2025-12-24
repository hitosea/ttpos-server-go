package req

// UpdatePrintSettingReq 更新打印设置请求
type UpdatePrintSettingReq struct {
	EnableCustomCopies string `json:"enable_custom_copies" binding:"required,oneof=0 1"` // 是否启用自定义打印联数 "0"-关闭 "1"-开启
	CheckoutSlipCopies int    `json:"checkout_slip_copies"`                              // 结账单打印联数 0-10
}
