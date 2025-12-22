package response

// TakeoutOrderResp 订单响应
type TakeoutOrderResp struct {
	Uuid             uint64                  `json:"uuid"`
	Platform         string                  `json:"platform"`
	PlatformOrderId  string                  `json:"platform_order_id"`
	ShortOrderNumber string                  `json:"short_order_number"`
	OrderState       int                     `json:"order_state"`
	IsAbnormal       int                     `json:"is_abnormal"`
	AbnormalDetail   string                  `json:"abnormal_detail"`
	StockStatus      int                     `json:"stock_status"`
	Subtotal         int64                   `json:"subtotal"`
	DeliveryFee      int64                   `json:"delivery_fee"`
	TotalAmount      int64                   `json:"total_amount"`
	CurrencyCode     string                  `json:"currency_code"`
	CurrencySymbol   string                  `json:"currency_symbol"`
	PaymentType      string                  `json:"payment_type"`
	OrderTime        int64                   `json:"order_time"`
	AcceptedTime     int64                   `json:"accepted_time"`
	Cutlery          int                     `json:"cutlery"`
	OrderType        string                  `json:"order_type"`
	Items            []*TakeoutOrderItemResp `json:"items"`
}

// TakeoutOrderItemResp 订单商品响应
type TakeoutOrderItemResp struct {
	Uuid             uint64 `json:"uuid"`
	PlatformItemId   string `json:"platform_item_id"`
	PlatformItemName string `json:"platform_item_name"`
	Quantity         int    `json:"quantity"`
	Price            int64  `json:"price"`
	Tax              int64  `json:"tax"`
	Specifications   string `json:"specifications"`
	IsMapped         int    `json:"is_mapped"`
}

// TakeoutOrderListResp 订单列表响应
type TakeoutOrderListResp struct {
	List []*TakeoutOrderResp `json:"list"`
	Meta Meta                `json:"meta"`
}

// TakeoutSettingsResp 配置响应
type TakeoutSettingsResp struct {
	Uuid       uint64 `json:"uuid"`
	Platform   string `json:"platform"`
	AutoAccept bool   `json:"auto_accept"`
	MaxAmount  int64  `json:"max_amount"`
}

// Meta 分页元数据
type Meta struct {
	PageNo   int `json:"page_no"`
	PageSize int `json:"page_size"`
	Total    int `json:"total"`
}
