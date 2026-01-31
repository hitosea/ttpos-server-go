package v1

// LinemanCommonResData LINE MAN 通用响应数据字段
// 所有 LINE MAN API 响应的公共字段
type LinemanCommonResData struct {
	Status  string `json:"status" dc:"结果状态：ok 表示成功，fail 表示失败"`
	Code    string `json:"code" dc:"结果代码"`
	Message string `json:"message,omitempty" dc:"结果描述"`
}
