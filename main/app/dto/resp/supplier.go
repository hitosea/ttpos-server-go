package resp

import "ttpos-server-go/app/dto"

// SupplierInfo 供应商信息
type SupplierInfo struct {
	Uuid uint64 `json:"uuid"` // 供应商UUID
	Name string `json:"name"` // 供应商名称
}

// SupplierListResp 供应商列表响应
type SupplierListResp struct {
	List []*SupplierInfo  `json:"list"`
	Meta dto.PageResponse `json:"meta"`
}

// SupplierCreateResp 创建供应商响应
type SupplierCreateResp struct {
	Uuid uint64 `json:"uuid"` // 供应商UUID
}

// SupplierDetailResp 供应商详情响应
type SupplierDetailResp struct {
	*SupplierInfo
}

// SupplierSimpleInfo 供应商简单信息（用于选择器）
type SupplierSimpleInfo struct {
	Name string `json:"name"` // 供应商名称
}

// SupplierSelectResp 供应商选择器响应
type SupplierSelectResp struct {
	List []*SupplierSimpleInfo `json:"list"`
}
