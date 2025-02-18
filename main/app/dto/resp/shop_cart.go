package resp

// 桌台购物车
type ShopCart struct {
	ShopCartType  int                `json:"shop_cart_type"`  // 购物车类型 0:桌台购物车 1:点餐购物车
	Desk          DeskInfo           `json:"desk"`            // 桌台信息
	Buffet        BuffetInfo         `json:"buffet"`          // 自助餐信息
	DiningMethod  int                `json:"dining_method"`   // 用餐方式 0:堂食 1:打包
	ProductList   []Product          `json:"product_list"`    // 商品列表
	SaleOrderList []SaleOrder        `json:"sale_order_list"` // 销售订单列表
	AmountInfo    ShopCartAmountInfo `json:"amount_info"`     // 金额信息
}

type ShopCartAmountInfo struct {
	ProductAmount        float64 `json:"product_amount"`         // 商品金额
	ServiceAmount        float64 `json:"service_amount"`         // 服务费
	TaxAmount            float64 `json:"tax_amount"`             // 税费
	DiscountAmount       float64 `json:"discount_amount"`        // 优惠折扣金额
	MemberDiscountAmount float64 `json:"member_discount_amount"` // 会员优惠折扣金额
	TotalAmount          float64 `json:"total_amount"`           // 总金额
}

// 自助餐信息
type BuffetInfo struct {
	EndTime string `json:"end_time"` // 自助餐结束时间
}

// 桌台信息
type DeskInfo struct {
	Uuid      uint64 `json:"uuid"`       // 桌台ID
	DeskNo    string `json:"desk_no"`    // 桌台编号
	MealNum   uint   `json:"meal_num"`   // 就餐人数
	StartTime string `json:"start_time"` // 开台时间
}

// 购物车销售订单信息
type SaleOrder struct {
	Uuid    uint64 `json:"uuid"`
	OrderNo string `json:"order_no"`
}

// 购物车商品
type Product struct {
	Uuid          uint64  `json:"uuid"`           // 商品uuid
	Name          string  `json:"name"`           // 商品名称
	Attribute     string  `json:"attribute"`      // 商品属性
	Num           uint    `json:"num"`            // 数量
	Price         float64 `json:"price"`          // 原价
	DiscountPrice float64 `json:"discount_price"` // 折扣价
	Status        int     `json:"status"`         // 0: 未送厨 1:已送厨 2:退菜 3:赠菜
	Remark        string  `json:"remark"`         // 备注
	IsMust        bool    `json:"is_must"`        // 是否必点
}
