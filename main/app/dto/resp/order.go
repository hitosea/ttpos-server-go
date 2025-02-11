package resp

import "ttpos-server-go/app/dto"

// 创建点餐订单响应
type CreateInstantOrderResp struct {
	SaleBillUuid uint64 `json:"sale_bill_uuid"` // 销售账单UUID
}

// 创建桌台订单响应
type CreateDeskOrderResp struct {
	SaleBillUuid uint64 `json:"sale_bill_uuid"` // 销售账单UUID
}

type OrderList struct {
	SaleBillUuid  uint64  `json:"sale_bill_uuid"` // 销售单UUID
	SerialNo      string  `json:"serial_no"`      // 桌台名称
	CustomerCount uint    `json:"customer_count"` // 桌台人数
	Status        uint    `json:"status"`         // 桌台状态	0:空闲 1:非自助餐 2:自助餐 3:待清台 4:锁单
	IsLock        uint    `json:"is_lock"`        // 是否锁单	0:否 1:是
	IsBuffet      uint    `json:"is_buffet"`      // 是否自助餐	0:否 1:是
	Time          uint    `json:"time"`           // 桌台用餐时间（秒）
	Price         float64 `json:"price"`          // 桌台价格
	Remark        string  `json:"remark"`         // 桌台备注
	TypeUuid      uint64  `json:"type_uuid"`      // 桌台类型ID
	RegionUuid    uint64  `json:"region_uuid"`    // 桌台区域ID
}

// 创建订单响应
type CreateOrderResp struct {
	Uuid uint64 `json:"uuid"` // 订单UUID
}

// 获取订单列表响应
type OrderListPaginationResp struct {
	List []OrderList      `json:"list"` // 订单列表
	Meta dto.PageResponse `json:"meta"` // 分页信息
}
