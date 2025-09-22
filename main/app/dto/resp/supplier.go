package resp

import "ttpos-server-go/app/dto"

// SupplierInfo 供应商信息
type SupplierInfo struct {
	Uuid          uint64 `json:"uuid"`           // 供应商UUID
	Name          string `json:"name"`           // 供应商名称
	Code          string `json:"code"`           // 供应商编码
	IsHeadquarter bool   `json:"is_headquarter"` // 是否为总部
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
	Code string `json:"code"` // 供应商编码
}

// SupplierSelectResp 供应商选择器响应
type SupplierSelectResp struct {
	List []*SupplierSimpleInfo `json:"list"`
}

type SupplierResp struct {
	*SupplierInfo
	Address                 string `json:"address"`                    // 供应商地址
	ContactName             string `json:"contact_name"`               // 联系人姓名
	ContactPhone            string `json:"contact_phone"`              // 联系人电话
	Status                  int    `json:"status"`                     // 状态：0-禁用；1-启用
	HasRelatedPurchaseOrder bool   `json:"has_related_purchase_order"` // 是否有关联的采购单
}
