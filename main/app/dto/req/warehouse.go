package req

import "ttpos-server-go/app/dto"

// CreateWarehouseReq 创建仓库请求
type CreateWarehouseReq struct {
	Type       string             `json:"type" binding:"required,oneof=normal transit"` // 类型：normal-普通；transit-在途
	Code       string             `json:"code" binding:"required,max=20"`               // 编码
	LocaleName dto.LocaleResponse `json:"name" binding:"required"`                      // 名称
	Status     int                `json:"status" binding:"oneof=0 1"`                   // 状态：0-禁用；1-启用
	Contact    string             `json:"contact" binding:"max=100"`                    // 联系人
	Phone      string             `json:"phone" binding:"max=20"`                       // 联系电话
	Address    string             `json:"address" binding:"max=500"`                    // 地址
}

// UpdateWarehouseReq 更新仓库请求
type UpdateWarehouseReq struct {
	Uuid       uint64             `json:"uuid" binding:"required"`
	Type       string             `json:"type" binding:"required,oneof=normal transit"` // 类型：normal-普通；transit-在途
	Code       string             `json:"code" binding:"required,max=20"`               // 编码
	LocaleName dto.LocaleResponse `json:"name" binding:"required"`                      // 名称
	Status     int                `json:"status" binding:"oneof=0 1"`                   // 状态：0-禁用；1-启用
	Contact    string             `json:"contact" binding:"max=100"`                    // 联系人
	Phone      string             `json:"phone" binding:"max=20"`                       // 联系电话
	Address    string             `json:"address" binding:"max=500"`                    // 地址
}

// WarehouseListReq 仓库列表请求
type WarehouseListReq struct {
	dto.PageReq
	Keyword string `json:"keyword" form:"keyword"` // 关键字
	Type    string `json:"type" form:"type"`       // 类型：normal-普通；transit-在途
	Status  *int   `json:"status" form:"status"`   // 状态：0-禁用；1-启用
}

// DeleteWarehouseReq 删除仓库请求
type DeleteWarehouseReq struct {
	Uuid uint64 `json:"uuid" binding:"required"` // 仓库ID
}

// WarehouseDetailReq 仓库详情请求
type WarehouseDetailReq struct {
	Uuid uint64 `json:"uuid" form:"uuid" binding:"required"` // 仓库ID
}

// WarehouseSelectReq 仓库选择器请求
type WarehouseSelectReq struct {
	Status *int `json:"status" form:"status"`
}

// SetDefaultWarehouseReq 设置默认仓库请求
type SetDefaultWarehouseReq struct {
	Uuid uint64 `json:"uuid" binding:"required"` // 仓库ID
}
