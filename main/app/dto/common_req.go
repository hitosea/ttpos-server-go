package dto

// PageReq 分页请求参数
type PageReq struct {
	PageNo   int `form:"page_no,default=1" binding:"min=1"`             // 页码
	PageSize int `form:"page_size,default=20" binding:"min=1,max=1000"` // 每页大小
}
