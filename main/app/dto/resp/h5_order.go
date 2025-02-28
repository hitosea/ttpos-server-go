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
	OrderUuid  uint64  `json:"order_uuid"`  // 订单Uuid
	OrderTime  int64   `json:"order_time"`  // 下单时间
	HandleTime int64   `json:"handle_time"` // 接单、拒单时间
	DeskNo     string  `json:"desk_no"`     // 桌台编号
	Price      float64 `json:"price"`       // 订单金额
	Status     uint    `json:"status"`      // 状态：1-待处理; 2-已接单; 3-已拒单
}

type H5OrderItem struct {
	H5OrderInfo
	DeskRegionUuid uint64 `json:"desk_region_uuid"` // 桌台区域Uuid
	Num            uint   `json:"num"`              // 商品数量
}

type H5OrderDetail struct {
	H5OrderInfo
	DeskUuid uint64 `json:"desk_uuid"` // 桌台Uuid
	Cashier  string `json:"cashier"`   // 接单、拒单操作人
}

type ProductItem struct {
	LocaleName dto.LocaleResponse `json:"locale_name"` // 商品名称多语言
	Num        uint               `json:"num"`         // 数量
	TotalPrice float64            `json:"total_price"` // 总价
}

type ProductList struct {
	List []ProductItem `json:"list"`
}

type OperationLogItem struct {
	RealName    string `json:"real_name"`   // real_name 真实姓名
	Email       string `json:"email"`       // 账号
	Source      string `json:"source"`      // 来源：收银端、商家后台等
	CreateTime  int64  `json:"create_time"` // 日志创建时间
	Description string `json:"description"` // 描述
}

type OperationLogList struct {
	List []OperationLogItem `json:"list"`
}

type H5OrderDetailResp struct {
	H5OrderDetail       H5OrderDetail    `json:"h5_order_detail"`       // 订单详情
	NewProductList      ProductList      `json:"new_product_list"`      // 新增商品列表
	AcceptedProductList ProductList      `json:"accepted_product_list"` // 已下单商品列表
	OperationLogList    OperationLogList `json:"operation_log_list"`    // 操作日志列表
}
