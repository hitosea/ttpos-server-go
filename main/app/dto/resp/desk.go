package resp

import "ttpos-server-go/app/dto"

type Desk struct {
	Uuid          uint64  `json:"uuid"`           // 桌台UUID
	DeskNo        string  `json:"desk_no"`        // 桌台名称
	CustomerCount uint    `json:"customer_count"` // 桌台人数
	Status        uint    `json:"status"`         // 桌台状态	0:空闲 1:非自助餐 2:自助餐 3:待清台 4:锁单
	IsLock        bool    `json:"is_lock"`        // 是否锁单
	IsBuffet      bool    `json:"is_buffet"`      // 是否自助餐
	Time          uint    `json:"time"`           // 桌台用餐时间（秒）
	Price         float64 `json:"price"`          // 桌台价格
	Remark        string  `json:"remark"`         // 桌台备注
	TypeUuid      uint64  `json:"type_uuid"`      // 桌台类型ID
	RegionUuid    uint64  `json:"region_uuid"`    // 桌台区域ID
}

type DeskExtra struct {
	AvailableNum       uint `json:"available_num"`         // 桌台可用数量
	LockNum            uint `json:"lock_num"`              // 桌台锁定数量
	OccupyBuffetNum    uint `json:"occupy_buffet_num"`     // 桌台自助餐数量
	OccupyNotBuffetNum uint `json:"occupy_not_buffet_num"` // 桌台不是自助餐数量
	OccupyWaitNum      uint `json:"occupy_wait_num"`       // 桌台等待数量
	TotalNum           uint `json:"total_num"`             // 桌台总计数量
}

type DeskRegion struct {
	Uuid uint64 `json:"uuid"` // 餐桌区域ID
	Name string `json:"name"` // 餐桌区域名称
}

type DeskType struct {
	Uuid uint64 `json:"uuid"` // 餐桌类型ID
	Name string `json:"name"` // 餐桌类型名称
}

// DeskRegionAndTypeListWithPaginationResp 桌台区域和类型列表响应
type DeskRegionAndTypeListWithPaginationResp struct {
	Region struct { // 餐桌区域
		List []DeskRegion `json:"list"`
	} `json:"region"`
	Type struct { // 餐桌类型
		List []DeskType `json:"list"`
	} `json:"type"`
}

// DeskListWithPaginationResp 桌台列表响应
type DeskListWithPaginationResp struct {
	Extra DeskExtra        `json:"extra"` // 桌台额外信息
	List  []Desk           `json:"list"`  // 桌台列表
	Meta  dto.PageResponse `json:"meta"`  // 分页信息
}

// DeskInfoResp 桌台详情响应
type DeskInfoResp struct {
	Uuid          uint64 `json:"uuid"`                     // 桌台UUID
	SaleBillUuid  uint64 `json:"sale_bill_uuid"`           // 订单UUID
	DeskNo        string `json:"desk_no"`                  // 桌台名称
	TypeUuid      uint64 `json:"type_uuid"`                // 桌台类型ID
	RegionUuid    uint64 `json:"region_uuid"`              // 桌台区域ID
	Status        uint   `json:"status"`                   // 桌台状态
	CustomerCount uint   `json:"customer_count,omitempty"` // 桌台人数
	IsLock        bool   `json:"is_lock"`                  // 是否锁单
	IsBuffet      bool   `json:"is_buffet"`                // 是否自助餐
	Remark        string `json:"remark"`                   // 桌台备注
	Time          uint   `json:"time"`                     // 桌台用餐时间（秒）
}

// 创建桌台订单响应
type CreateDeskOrderResp struct {
	SaleBillUuid  uint64 `json:"sale_bill_uuid"`  // 销售账单UUID
	SaleOrderUuid uint64 `json:"sale_order_uuid"` // 销售订单UUID
}
