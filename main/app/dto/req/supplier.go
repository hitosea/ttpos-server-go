package req

import "ttpos-server-go/app/dto"

// SupplierListReq 供应商列表请求
type SupplierListReq struct {
	dto.PageReq
	Name string `form:"name" json:"name"` // 供应商名称
}

// SupplierCreateReq 创建供应商请求
type SupplierCreateReq struct {
	Name         string `json:"name" binding:"required,max=100"`         // 供应商名称
	Address      string `json:"address" binding:"omitempty,max=200"`     // 供应商地址
	ContactName  string `json:"contact_name" binding:"required,max=50"`  // 联系人姓名
	ContactPhone string `json:"contact_phone" binding:"required,max=20"` // 联系人电话
	Position     string `json:"position" binding:"omitempty,max=50"`     // 职位
	StaffUuid    uint64 `json:"staff_uuid" binding:"required"`           // 采购负责人UUID
}

// SupplierUpdateReq 更新供应商请求
type SupplierUpdateReq struct {
	Uuid         uint64 `json:"uuid" binding:"required"`                 // 供应商UUID
	Name         string `json:"name" binding:"required,max=100"`         // 供应商名称
	Address      string `json:"address" binding:"omitempty,max=200"`     // 供应商地址
	ContactName  string `json:"contact_name" binding:"required,max=50"`  // 联系人姓名
	ContactPhone string `json:"contact_phone" binding:"required,max=20"` // 联系人电话
	Position     string `json:"position" binding:"omitempty,max=50"`     // 职位
	StaffUuid    uint64 `json:"staff_uuid" binding:"required"`           // 采购负责人UUID
}

// SupplierDeleteReq 删除供应商请求
type SupplierDeleteReq struct {
	Uuid uint64 `json:"uuid" binding:"required"` // 供应商UUID
}

// 供应商请求验证消息
var SupplierCreateReqMessage = map[string]string{
	"name.required":          "供应商名称不能为空",
	"name.max":               "供应商名称不能超过100个字符",
	"contact_name.required":  "联系人姓名不能为空",
	"contact_name.max":       "联系人姓名不能超过50个字符",
	"contact_phone.required": "联系人电话不能为空",
	"contact_phone.max":      "联系人电话不能超过20个字符",
	"position.max":           "职位不能超过50个字符",
	"address.max":            "地址不能超过200个字符",
	"staff_uuid.required":    "采购负责人不能为空",
}

var SupplierUpdateReqMessage = map[string]string{
	"uuid.required":          "供应商UUID不能为空",
	"name.required":          "供应商名称不能为空",
	"name.max":               "供应商名称不能超过100个字符",
	"contact_name.required":  "联系人姓名不能为空",
	"contact_name.max":       "联系人姓名不能超过50个字符",
	"contact_phone.required": "联系人电话不能为空",
	"contact_phone.max":      "联系人电话不能超过20个字符",
	"position.max":           "职位不能超过50个字符",
	"address.max":            "地址不能超过200个字符",
	"staff_uuid.required":    "采购负责人不能为空",
}

var SupplierDeleteReqMessage = map[string]string{
	"uuid.required": "供应商UUID不能为空",
}
