package resp

import "ttpos-server-go/app/dto"

type H5OrderListExtra struct {
	UnhandledCount int64 `json:"unhandled_count"` // 未处理的接单数量
	HandledCount   int64 `json:"handled_count"`   // 已处理的接单数量
}

type H5OrderList struct {
	Extra H5OrderListExtra `json:"extra"`
	List  []H5OrderItem    `json:"list"`
	Meta  dto.PageResponse `json:"meta"`
}

type H5OrderInfo struct {
	SaleBillUuid uint64  `json:"sale_bill_uuid"` // 销售账单Uuid
	H5OrderUuid  uint64  `json:"h5_order_uuid"`  // h5订单Uuid
	OrderTime    int64   `json:"order_time"`     // h5订单下单时间
	HandleTime   int64   `json:"handle_time"`    // 接单、拒单时间
	WaitTime     int64   `json:"wait_time"`      // 等待时间，单位：秒
	DeskNo       string  `json:"desk_no"`        // 桌台编号
	Price        float64 `json:"price"`          // 订单金额
	Status       uint    `json:"status"`         // 状态：1-待处理; 2-已接单; 3-已拒单
}

type H5OrderItem struct {
	H5OrderInfo
	DeskRegionUuid uint64  `json:"desk_region_uuid"` // 桌台区域Uuid
	Num            float64 `json:"num"`              // 商品数量
}

type H5OrderDetail struct {
	H5OrderInfo
	DeskUuid uint64 `json:"desk_uuid"` // 桌台Uuid
	Cashier  string `json:"cashier"`   // 接单、拒单操作人
}

type ProductItem struct {
	LocaleName  dto.LocaleResponse `json:"locale_name"`  // 商品名称多语言
	Num         float64            `json:"num"`          // 数量
	TotalPrice  float64            `json:"total_price"`  // 总价
	SubProducts SubProductList     `json:"sub_products"` // 套餐子商品列表
}

type SubProductList struct {
	List []SubProductItem `json:"list"`
}

type SubProductItem struct {
	LocaleName dto.LocaleResponse `json:"locale_name"` // 商品名称多语言
	Num        float64            `json:"num"`         // 数量
}

type ProductList struct {
	List []ProductItem `json:"list"`
}

type OperationLog struct {
	List []OrderOperationLog `json:"list"`
}

type H5OrderDetailResp struct {
	H5OrderDetail   H5OrderDetail `json:"h5_order_detail"`  // 订单详情
	NewProduct      ProductList   `json:"new_product"`      // 新增商品列表
	AcceptedProduct ProductList   `json:"accepted_product"` // 已下单商品列表
	OperationLog    OperationLog  `json:"operation"`        // 操作日志列表
}
