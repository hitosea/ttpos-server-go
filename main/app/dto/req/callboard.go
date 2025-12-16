package req

// GetBindCodeReq 获取绑定码请求
type GetBindCodeReq struct {
	DeviceId string `form:"device_id" binding:"required"`
}

// GetBindInfoReq 获取绑定信息请求
type GetBindInfoReq struct {
	DeviceId string `form:"device_id" binding:"required"`
}

// GetQueueDataReq 获取队列数据请求
type GetQueueDataReq struct {
	DeviceId   string `form:"device_id" binding:"required"`
	Limit      int64  `form:"limit" binding:"required"`
	UpdateTime int64  `form:"update_time"`
	// Timestamp  int64  `form:"timestamp" binding:"required"`
	// Sign       string `form:"sign" binding:"required"`
}

// BindDeviceReq 绑定设备请求
type BindDeviceReq struct {
	BindCode           string `json:"bind_code" binding:"required,max=10"`
	Lang1              string `json:"lang1"`
	Lang2              string `json:"lang2"`
	Name               string `json:"name"`                 // 设备名称（可选，版本 >= 2.11.0 时必填）
	BackgroundImageUrl string `json:"background_image_url"` // 背景图片 URL（可选，可传空字符串）
	TimeoutLimit       *int   `json:"timeout_limit"`        // 超时限制（分钟，可选）
	VoiceCallEnabled   *bool  `json:"voice_call_enabled"`   // 语音叫号开关（可选）
	CallCount          int    `json:"call_count"`           // 叫号次数（可选，版本 >= 2.11.0 时必填，最小1，最大3）
}

// GetDeviceListReq 获取设备列表请求
type GetDeviceListReq struct {
}

// UnbindDeviceReq 解绑设备请求
type UnbindDeviceReq struct {
	Uuid uint64 `json:"uuid" binding:"required"`
}

// UpdateBindInfoReq 更新绑定信息请求
type UpdateBindInfoReq struct {
	Uuid               uint64 `json:"uuid" binding:"required"`
	Lang1              string `json:"lang1"`
	Lang2              string `json:"lang2"`
	Name               string `json:"name"`                 // 设备名称（可选，版本 >= 2.11.0 时必填）
	BackgroundImageUrl string `json:"background_image_url"` // 背景图片 URL（可选，可传空字符串）
	TimeoutLimit       *int   `json:"timeout_limit"`        // 超时限制（分钟，可选）
	VoiceCallEnabled   *bool  `json:"voice_call_enabled"`   // 语音叫号开关（可选）
	CallCount          int    `json:"call_count"`           // 叫号次数（可选，版本 >= 2.11.0 时必填，最小1，最大3）
}
