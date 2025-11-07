package req

import "ttpos-server-go/app/dto"

// ExportRecordListReq 导出记录列表请求
type ExportRecordListReq struct {
	dto.PageReq
	ExportType *uint8 `json:"export_type" form:"export_type"` // 导出类型，不传则查询全部
}

// ExportRecordDeleteReq 批量删除导出记录请求
type ExportRecordDeleteReq struct {
	Uuids []uint64 `json:"uuids" binding:"required,min=1"` // 导出记录UUID数组
}
