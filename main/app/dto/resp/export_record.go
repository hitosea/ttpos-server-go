package resp

import "ttpos-server-go/app/dto"

// ExportRecordListResp 导出记录列表响应
type ExportRecordListResp struct {
	Uuid       uint64 `json:"uuid"`        // UUID
	ExportType uint8  `json:"export_type"` // 导出类型: 1-时段营业统计, 2-综合运营统计, 3-营业应收统计, 4-菜品出品明细, 5-菜品出品详情
	ExportName string `json:"export_name"` // 导出文件名称
	FileUuid   uint64 `json:"file_uuid"`   // 文件UUID
	FileName   string `json:"file_name"`   // 文件名称
	FileUrl    string `json:"file_url"`    // 文件下载链接
	FileSize   int    `json:"file_size"`   // 文件大小
	Status     uint8  `json:"status"`      // 状态 0-导出中, 1-导出成功, 2-导出失败
	ErrorMsg   string `json:"error_msg"`   // 错误信息
	CreateTime int64  `json:"create_time"` // 创建时间
	UpdateTime int64  `json:"update_time"` // 更新时间
}

// ExportRecordListPaginationResp 导出记录列表分页响应
type ExportRecordListPaginationResp struct {
	List []ExportRecordListResp `json:"list"`
	Meta dto.PageResponse       `json:"meta"`
}
