package dto

type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

// 分页响应
type PageResponse struct {
	PageNo   int   `json:"page_no"`   // 当前页码
	PageSize int   `json:"page_size"` // 每页大小
	Total    int64 `json:"total"`     // 总数
}

// 语言响应
type LocaleResponse struct {
	ZH   string `json:"zh"`   // 中文
	TH   string `json:"th"`   // 泰语
	EN   string `json:"en"`   // 英语
	ZHTW string `json:"zhtw"` // 粤语
	JA   string `json:"ja"`   // 日语
	KO   string `json:"ko"`   // 韩语
	MY   string `json:"my"`   // 缅甸语
	TR   string `json:"tr"`   // 土耳其语
}
