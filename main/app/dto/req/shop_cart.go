package req

// OrderCartProduct 购物车商品请求参数
type OrderCartProduct struct {
	SaleBillUuid         uint64 `json:"sale_bill_uuid"`          // 销售账单ID
	SaleOrderUuid        uint64 `json:"sale_order_uuid"`         // 销售订单ID
	SaleOrderProductUuid uint64 `json:"sale_order_product_uuid"` // 销售订单商品ID
}

// OrderCartProductAddReq 向购物车添加商品请求参数
type OrderCartProductAddReq struct {
	SaleBillUuid      uint64   `json:"sale_bill_uuid"`  // 销售账单ID。可选，参数不填时表示要新建销售账单，添加商品后创建点餐销售账单。
	SaleOrderUuid     uint64   `json:"sale_order_uuid"` // 销售订单ID。可选，参数不填时默认加购到第一个销售订单中
	FlavorUuid        uint64   `json:"flavor_uuid"`     // 某个规格商品ID
	SauceUuidList     []uint64 `json:"sauce_uuid"`      // 小料ID列表
	AttributeUuidList []uint64 `json:"attribute_uuid"`  // 属性ID列表
	Operation         string   `json:"operation"`       // 操作类型。add: 加购，sub: 减购. 不填，默认是加购
	Num               *float64 `json:"num"`             // 商品数量。可选，不填时，默认是1。
	MustPlanUuid      uint64   `json:"must_plan_uuid"`  // 必点方案uuid. 可选，在必点方案弹窗中加购时填写
	Price             *float64 `json:"price"`           // 商品价格。可选，当商品价格与后台设置的最新价格不一致时，加购失败并返回最新价格
	IsBuffet          *bool    `json:"is_buffet"`       // 是否是自助餐商品。可选，不填时，表示不判断是不是最新价格。该参数仅在判断价格时使用
	// 后端内部使用的参数
	isH5Product bool `json:"-"` // 是否是H5商品
}

func (req *OrderCartProductAddReq) SetIsH5Product() {
	req.isH5Product = true
}

func (req *OrderCartProductAddReq) IsH5Product() bool {
	return req.isH5Product
}

// ProductAddReq 添加商品请求参数
type ProductAddReq struct {
	SaleBillUuid  uint64          `json:"sale_bill_uuid"`  // 销售账单ID。
	SaleOrderUuid uint64          `json:"sale_order_uuid"` // 销售订单ID。
	Products      []ProductParams `json:"products"`        // 商品信息列表·
	IsH5Product   bool            `json:"is_h5_product"`   // 是否是H5商品
}

type GetProductPackageDetailReq struct {
	SaleBillUuid       uint64 `form:"sale_bill_uuid"`       // 销售账单ID
	SaleOrderUuid      uint64 `form:"sale_order_uuid"`      // 销售订单ID
	ProductPackageUuid uint64 `form:"product_package_uuid"` // 商品包uuid
}

type TabletOrderCartProductAddReq struct {
	SaleBillUuid  uint64          `json:"sale_bill_uuid" binding:"required"`      // 销售账单ID。
	SaleOrderUuid uint64          `json:"sale_order_uuid" binding:"required"`     // 销售订单ID。
	Products      []ProductParams `json:"products" binding:"required,min=1,dive"` // 商品信息列表·
	IgnoreMust    bool            `json:"ignore_must"`                            // 是否忽略必点方案
}

// ProductParams 商品参数
type ProductParams struct {
	FlavorProductBomUuid            uint64   `json:"flavor_product_bom_uuid" binding:"required"` // 商品规格uuid
	Num                             float64  `json:"num"  binding:"required"`                    // 数量数量
	Price                           *float64 `json:"price"`                                      // 商品价格，商品单价。当商品价格与后台设置的最新价格不一致时，加购失败并返回最新价格。可选，不传时，不进行价格校验
	IsBuffet                        *bool    `json:"is_buffet"`                                  // 是否是自助餐商品。可选，不填时，表示不判断是不是最新价格。该参数仅在判断价格时使用
	SauceProductBomUuidList         []uint64 `json:"sauce_product_bom_uuid_list"`                // 加料信息
	ProductPackageAttributeUuidList []uint64 `json:"product_package_attribute_uuid_list"`        // 属性信息
	Operation                       string   `json:"operation"`                                  // 操作类型。add: 加购，sub: 减购
	MustPlanUuid                    uint64   `json:"must_plan_uuid"`                             // 必点方案uuid. 可选，在必点方案弹窗中加购时填写
	Remark                          string   `json:"remark"`                                     // 备注，平板端离线购物车提交
}

// OrderCartProductNumReq 修改购物车商品数量请求参数
type OrderCartProductNumReq struct {
	SaleBillUuid         uint64  `json:"sale_bill_uuid"`          // 销售账单ID
	SaleOrderUuid        uint64  `json:"sale_order_uuid"`         // 销售订单ID
	SaleOrderProductUuid uint64  `json:"sale_order_product_uuid"` // 销售订单商品ID
	Num                  float64 `json:"num"`                     // 数量
}

// OrderCartProductCookingReq 送厨购物车商品请求参数
type OrderCartProductCookingReq struct {
	SaleBillUuid   uint64 `json:"sale_bill_uuid"`   // 销售账单ID
	SaleOrderUuid  uint64 `json:"sale_order_uuid"`  // 销售订单ID。废弃
	IgnoreMust     bool   `json:"ignore_must"`      // 是否忽略必点方案
	H5OrderUuid    uint64 `json:"h5_order_uuid"`    // h5订单ID。默认为0，表示不送厨h5订单商品。当从H5订单进入桌台时，需要传入h5订单ID，将该h5订单的商品送厨
	Password       string `json:"password"`         // 高级密码后台开启的时候才传
	IsCheckCooking bool   `json:"is_check_cooking"` // 是否只进行送厨检查，而不进行实际的送厨。场景：助手端开启下单校验高级密码时，先检查送厨，再实际送厨。检查送厨时不进行实际送厨
}

type H5ConfirmOrderReq struct {
	IgnoreMust bool `json:"ignore_must"` // 是否忽略必点方案
}

// OrderCartProductReturningReq 退菜购物车商品退菜请求参数
type OrderCartProductReturningReq struct {
	SaleBillUuid         uint64   `json:"sale_bill_uuid"`          // 销售账单ID
	SaleOrderUuid        uint64   `json:"sale_order_uuid"`         // 销售订单ID
	SaleOrderProductUuid uint64   `json:"sale_order_product_uuid"` // 销售订单商品ID
	Num                  float64  `json:"num"`                     // 退菜数量
	Reason               string   `json:"reason"`                  // 退菜原因
	Password             string   `json:"password"`                // 高级密码 后台开启的时候才传
	ReturnIds            []uint64 `json:"return_ids"`              // 退菜标签ids
}

// OrderCartProductGivingReq  购物车商品赠菜请求参数
type OrderCartProductGivingReq struct {
	SaleBillUuid         uint64   `json:"sale_bill_uuid"`          // 销售账单ID
	SaleOrderUuid        uint64   `json:"sale_order_uuid"`         // 销售订单ID
	SaleOrderProductUuid uint64   `json:"sale_order_product_uuid"` // 销售订单商品ID
	Reason               string   `json:"reason"`                  // 赠菜原因
	GiftIds              []uint64 `json:"gift_ids"`                // 赠菜标签ids
}

// OrderCartProductChangeDeskReq 转菜购物车商品请求参数
type OrderCartProductChangeDeskReq struct {
	SaleBillUuid         uint64 `json:"sale_bill_uuid"`          // 销售账单ID
	SaleOrderUuid        uint64 `json:"sale_order_uuid"`         // 销售订单ID
	SaleOrderProductUuid uint64 `json:"sale_order_product_uuid"` // 销售订单商品ID
	DeskUuid             uint64 `json:"desk_uuid"`               // 目标桌台ID
}

// RejectH5OrderReq 拒单请求参数
type RejectH5OrderReq struct {
	H5OrderUuid uint64 `json:"h5_order_uuid"` // h5订单ID
}

// AcceptH5OrderReq 接单请求参数
type AcceptH5OrderReq struct {
	H5OrderUuid uint64 `json:"h5_order_uuid"` // h5订单ID
}
