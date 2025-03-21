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
	SauceUuidList     []uint64 `json:"sauce_uuid"`      // 小料ID
	AttributeUuidList []uint64 `json:"attribute_uuid"`  // 规格ID

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

type TabletOrderCartProductAddReq struct {
	SaleBillUuid  uint64          `json:"sale_bill_uuid" binding:"required"`      // 销售账单ID。
	SaleOrderUuid uint64          `json:"sale_order_uuid" binding:"required"`     // 销售订单ID。
	Products      []ProductParams `json:"products" binding:"required,min=1,dive"` // 商品信息列表·
}

// ProductParams 商品参数
type ProductParams struct {
	FlavorProductBomUuid            uint64   `json:"flavor_product_bom_uuid" binding:"required"` // 商品规格uuid
	Num                             uint     `json:"num"  binding:"required"`                    // 数量数量
	SauceProductBomUuidList         []uint64 `json:"sauce_product_bom_uuid_list"`                // 加料信息
	ProductPackageAttributeUuidList []uint64 `json:"product_package_attribute_uuid_list"`        // 属性信息
}

// OrderCartProductNumReq 修改购物车商品数量请求参数
type OrderCartProductNumReq struct {
	SaleBillUuid         uint64 `json:"sale_bill_uuid"`          // 销售账单ID
	SaleOrderUuid        uint64 `json:"sale_order_uuid"`         // 销售订单ID
	SaleOrderProductUuid uint64 `json:"sale_order_product_uuid"` // 销售订单商品ID
	Num                  int    `json:"num"`                     // 数量
}

// OrderCartProductCookingReq 送厨购物车商品请求参数
type OrderCartProductCookingReq struct {
	SaleBillUuid  uint64 `json:"sale_bill_uuid"`  // 销售账单ID
	SaleOrderUuid uint64 `json:"sale_order_uuid"` // 销售订单ID。废弃
	IgnoreMust    bool   `json:"ignore_must"`     // 是否忽略必点方案
}

// OrderCartProductReturningReq 退菜购物车商品退菜请求参数
type OrderCartProductReturningReq struct {
	SaleBillUuid         uint64   `json:"sale_bill_uuid"`          // 销售账单ID
	SaleOrderUuid        uint64   `json:"sale_order_uuid"`         // 销售订单ID
	SaleOrderProductUuid uint64   `json:"sale_order_product_uuid"` // 销售订单商品ID
	Num                  uint     `json:"num"`                     // 退菜数量
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
