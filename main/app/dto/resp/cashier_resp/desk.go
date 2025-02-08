package cashier_resp

import "ttpos-server-go/app/dto"

type Desk struct {
	Uuid          uint    `json:"uuid"`           // 桌台UUID
	TableNo       string  `json:"table_no"`       // 桌台名称
	CustomerCount uint    `json:"customer_count"` // 桌台人数
	Status        uint    `json:"status"`         // 桌台状态	0:空闲 1:非自助餐 2:自助餐 3:待清台 4:锁单
	IsBuffet      uint    `json:"is_buffet"`      // 是否自助餐
	Time          uint    `json:"time"`           // 桌台用餐时间（秒）
	Price         float64 `json:"price"`          // 桌台价格
	TypeUuid      uint    `json:"type_uuid"`      // 桌台类型ID
	RegionUuid    uint    `json:"region_uuid"`    // 桌台区域ID
}

type DeskExtra struct {
	AvailableNum       uint `json:"available_num"`         // 桌台可用数量
	LockNum            uint `json:"lock_num"`              // 桌台锁定数量
	OccupyBuffetNum    uint `json:"occupy_buffet_num"`     // 桌台自助餐数量
	OccupyNotBuffetNum uint `json:"occupy_not_buffet_num"` // 桌台不是自助餐数量
	OccupyWaitNum      uint `json:"occupy_wait_num"`       // 桌台等待数量
	TotalNum           uint `json:"total_num"`             // 桌台总计数量
}

// DeskListWithPaginationResp 桌台列表响应
type DeskListWithPaginationResp struct {
	Extra DeskExtra        `json:"extra"`
	List  []Desk           `json:"list"`
	Meta  dto.PageResponse `json:"meta"`
}
