package dto

// PageReq 分页请求参数
type PageReq struct {
	PageNo   int `form:"page_no,default=1"  json:"page_no,default=1" binding:"omitempty,min=1"`               // 页码
	PageSize int `form:"page_size,default=20" json:"page_size,default=20" binding:"omitempty,min=1,max=1100"` // 每页大小
}

// PageReqMessage 分页请求参数错误信息
var PageReqMessage = map[string]string{
	"page_no.min":   "页码不能小于1",
	"page_size.min": "每页大小不能小于1",
	"page_size.max": "每页大小不能大于1100",
}

type CheckNameResult struct {
	Lang      string `json:"lang"`
	TextExist bool   `json:"text_exist"`
}
