package resp

// ChannelSalesBlock 渠道营业统计数据块
type ChannelSalesBlock struct {
	TotalOrderNum      int64   `json:"total_order_num"`       // 订单数
	MinOrderAmount     float64 `json:"min_order_amount"`      // 最小订单金额
	MaxOrderAmount     float64 `json:"max_order_amount"`      // 最大订单金额
	AvgOrderAmount     float64 `json:"avg_order_amount"`      // 平均订单金额
	TotalDeskNum       int64   `json:"total_desk_num"`        // 桌数（仅桌台渠道）
	TotalMealNum       int64   `json:"total_meal_num"`        // 人数（仅桌台渠道）
	OrderAmountMealAvg float64 `json:"order_meal_avg_amount"` // 人均订单金额（仅桌台渠道）
}

// ChannelSalesMeta 渠道营业统计元数据
type ChannelSalesMeta struct {
	StartTime int64 `json:"start_time"` // 查询开始时间戳
	EndTime   int64 `json:"end_time"`   // 查询结束时间戳
}

// ChannelSalesResp 渠道营业统计响应
type ChannelSalesResp struct {
	Summary         *ChannelSalesBlock `json:"summary"`          // 合计
	Table           *ChannelSalesBlock `json:"table"`            // 桌台
	DineIn          *ChannelSalesBlock `json:"dine_in"`          // 点餐-店内
	Scan            *ChannelSalesBlock `json:"scan"`             // 点餐-扫码
	TakeoutShop     *ChannelSalesBlock `json:"takeout_shop"`     // 点餐-外卖
	TakeoutDelivery *ChannelSalesBlock `json:"takeout_delivery"` // 外送
	DineInStore     *ChannelSalesBlock `json:"dine_in_store"`    // 堂食
	Takeaway        *ChannelSalesBlock `json:"takeaway"`         // 外带
	Grab            *ChannelSalesBlock `json:"grab"`             // Grab 外卖
	Lineman         *ChannelSalesBlock `json:"lineman"`          // LINE MAN 外卖
	Takeout         *ChannelSalesBlock `json:"takeout"`          // 外卖
	Meta            *ChannelSalesMeta  `json:"meta"`             // 元数据
}
