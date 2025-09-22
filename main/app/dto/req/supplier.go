package req

import "ttpos-server-go/app/dto"

// SupplierListReq 供应商列表请求
type SupplierListReq struct {
	dto.PageReq
	Keyword string `form:"keyword" json:"keyword"` // 供应商名称
}

// SupplierCreateReq 创建供应商请求
type SupplierCreateReq struct {
	Name         string `json:"name" binding:"required,max=150"`          // 供应商名称
	Code         string `json:"code" binding:"required,max=20"`           // 供应商编码
	Status       int    `json:"status" binding:"oneof=0 1"`               // 供应商状态:0-禁用；1-启用
	Address      string `json:"address" binding:"omitempty,max=500"`      // 供应商地址
	ContactName  string `json:"contact_name" binding:"omitempty,max=100"` // 联系人姓名
	ContactPhone string `json:"contact_phone" binding:"omitempty,max=20"` // 联系人电话
}

// SupplierUpdateReq 更新供应商请求
type SupplierUpdateReq struct {
	Uuid         uint64 `json:"uuid" binding:"required"`                  // 供应商UUID
	Name         string `json:"name" binding:"required,max=150"`          // 供应商名称
	Code         string `json:"code" binding:"required,max=20"`           // 供应商编码
	Status       int    `json:"status" binding:"oneof=0 1"`               // 供应商状态:0-禁用；1-启用
	Address      string `json:"address" binding:"omitempty,max=500"`      // 供应商地址
	ContactName  string `json:"contact_name" binding:"omitempty,max=100"` // 联系人姓名
	ContactPhone string `json:"contact_phone" binding:"omitempty,max=20"` // 联系人电话
}

// SupplierDeleteReq 删除供应商请求
type SupplierDeleteReq struct {
	Uuid uint64 `json:"uuid" binding:"required"` // 供应商UUID
}

// SupplierSelectReq 供应商选择器请求
type SupplierSelectReq struct {
	PurchaseType int `form:"purchase_type" json:"purchase_type"` // 采购类型, 1-外部采购 2-内部采购
}

// SupplierReq 获取供应商请求
type SupplierReq struct {
	Uuid uint64 `json:"uuid" form:"uuid" binding:"required"` // 供应商UUID
}

var SupplierSelectReqMessage = map[string]string{
	"is_external.oneof": "是否外部采购参数值只能是0或1",
}
