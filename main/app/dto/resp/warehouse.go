package resp

import "ttpos-server-go/app/dto"

// WarehouseResp 仓库响应
type WarehouseResp struct {
	Uuid      uint64             `json:"uuid"`
	LocalName dto.LocaleResponse `json:"local_name"`
	Type      string             `json:"type"`
	Code      string             `json:"code"`
	Status    int                `json:"status"`
	Contact   string             `json:"contact"`
	Phone     string             `json:"phone"`
	Address   string             `json:"address"`
	IsDefault int                `json:"is_default"`
}

// WarehouseListResp 仓库列表响应
type WarehouseListResp struct {
	List []WarehouseResp  `json:"list"`
	Meta dto.PageResponse `json:"meta"`
}
