package req

type TakeoutAddress struct {
	AddressName string // 地址名称。商家名称
	Address     string // 地址
	Lat         string // 纬度
	Lng         string // 经度
}

type TakeoutDistanceReq struct {
	ProviderName string            // 外送渠道
	Address      []*TakeoutAddress // 起止地址和经纬度
}

type TakeoutLocation struct {
	TakeoutAddress
	ContactName  string // 联系人。外送订单的收获人或商家名
	ContactPhone string // 联系电话。外送订单的收获人或商家电话
}

type CreateTakeoutOrderReq struct {
	ProviderName     string           // 外送渠道
	CustomerLocation *TakeoutLocation // 顾客地址
	MerchantLocation *TakeoutLocation // 商户地址
	Remark           string           // 备注
	CallbackUrl      string           // 回调地址
	ShopOrderUuid    string           // 商户订单号。member_sale_order_uuid
}

type ConfirmTakeoutOrderReq struct {
	ShopOrderUuid string // 商户订单号。member_sale_order_uuid
}

type CancelTakeoutOrderReq struct {
	ShopOrderUuid string // 商户订单号。member_sale_order_uuid
	Reason        string // 取消原因
}

type GetDriverInfoReq struct {
	ShopOrderUuid string // 商户订单号。member_sale_order_uuid
}

// GetImportProgressReq 获取导入进度请求
type GetImportProgressReq struct {
	Platform string `form:"platform" binding:"required" json:"platform"` // 外卖平台
}

// GetImportLogsReq 获取导入日志列表请求
type GetImportLogsReq struct {
	Platform   string `form:"platform" json:"platform"`                           // 外卖平台筛选
	ImportType *int8  `form:"import_type" json:"import_type"`                     // 导入类型筛选（1-TTPOS推送到平台 2-平台推送到TTPOS）
	Status     *int8  `form:"status" json:"status"`                               // 状态筛选（0-进行中 1-成功 2-失败）
	PageNo     int    `form:"page_no" binding:"required,min=1" json:"page_no"`    // 页码
	PageSize   int    `form:"page_size" binding:"min=1,max=100" json:"page_size"` // 每页数量
}

// PrintProductItem 打印送厨单商品项
type PrintProductItem struct {
	ProductUuid    uint64 `json:"product_uuid"`     // 商品UUID
	ProductBomUuid uint64 `json:"product_bom_uuid"` // 规格UUID (BOM UUID)
}
