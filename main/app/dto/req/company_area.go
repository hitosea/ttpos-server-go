package req

import "ttpos-server-go/app/dto"

// CompanyAreaCreateReq 创建区域请求
type CompanyAreaCreateReq struct {
	LocaleName    dto.LocaleResponse `json:"locale_name" binding:"required"`    // 区域多语言名称
	CompanyUuids  []uint64           `json:"company_uuids"`                     // 关联的门店UUID列表
	Sort          int                `json:"sort"`                              // 排序
}

// CompanyAreaUpdateReq 编辑区域请求
type CompanyAreaUpdateReq struct {
	Uuid          uint64             `json:"uuid" binding:"required"`           // 区域UUID
	LocaleName    dto.LocaleResponse `json:"locale_name" binding:"required"`    // 区域多语言名称
	CompanyUuids  []uint64           `json:"company_uuids"`                     // 关联的门店UUID列表
	Sort          int                `json:"sort"`                              // 排序
}

// CompanyAreaDeleteReq 删除区域请求
type CompanyAreaDeleteReq struct {
	Uuid uint64 `json:"uuid" binding:"required"` // 区域UUID
}

// CompanyAreaInfoReq 区域详情请求
type CompanyAreaInfoReq struct {
	Uuid uint64 `form:"uuid" binding:"required"` // 区域UUID
}
