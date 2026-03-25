package req

// OrderCartProduct 购物车商品请求参数
type OrderCartProduct struct {
	SaleBillUuid         uint64 `json:"sale_bill_uuid"`          // 销售账单ID
	SaleOrderUuid        uint64 `json:"sale_order_uuid"`         // 销售订单ID
	SaleOrderProductUuid uint64 `json:"sale_order_product_uuid"` // 销售订单商品ID
}

// OrderCartProductPackageAddReq 向购物车添加套餐请求参数
type OrderCartProductPackageAddReq struct {
	SaleBillUuid       uint64           `json:"sale_bill_uuid"`       // 销售账单UUID
	SaleOrderUuid      uint64           `json:"sale_order_uuid"`      // 销售订单UUID
	ProductPackageUuid uint64           `json:"product_package_uuid"` // 套餐UUID
	Products           []ProductRequest `json:"products"`             // 套餐商品请求列表
	Num                float64          `json:"num"`                  // 套餐商品数量

	isH5Product bool `json:"-"` // 是否是H5端下单的商品
}

func (req *OrderCartProductPackageAddReq) SetIsH5Product() {
	req.isH5Product = true
}

func (req *OrderCartProductPackageAddReq) IsH5Product() bool {
	return req.isH5Product
}

// ProductRequest 套餐商品请求参数
type ProductRequest struct {
	ProductPackageGroupUuid uint64 `json:"product_package_group_uuid"` // 套餐分组UUID
	// 普通商品的参数
	EditProductReq
	Num      float64 `json:"num"`       // 商品份数
	UnitNum  float64 `json:"unit_num"`  // 一个套餐的单个子商品的数量
	AddPrice float64 `json:"add_price"` // 加价金额
}

type OrderCartProductFlavorAndAttributeChangeReq struct {
	OrderCartProductPackageAddReq        // 套餐商品请求参数
	ProductType                   uint   `json:"product_type"`            // 商品类型 0-商品 1-套餐 2-套餐子商品
	SaleOrderProductUuid          uint64 `json:"sale_order_product_uuid"` // 销售订单商品ID

	// 普通商品的参数
	EditProductReq
}

type EditProductReq struct {
	FlavorUuid        uint64   `json:"flavor_uuid"`    // 某个规格商品ID
	SauceUuidList     []uint64 `json:"sauce_uuid"`     // 小料ID列表
	AttributeUuidList []uint64 `json:"attribute_uuid"` // 属性ID列表
}

// OrderCartProductFlavorAndAttributeReq 查询购物车商品“规格/属性”请求参数
type OrderCartProductFlavorAndAttributeReq struct {
	SaleBillUuid         uint64 `form:"sale_bill_uuid"`          // 销售账单ID
	SaleOrderUuid        uint64 `form:"sale_order_uuid"`         // 销售订单ID
	SaleOrderProductUuid uint64 `form:"sale_order_product_uuid"` // 销售订单商品ID
	ProductType          uint   `form:"product_type"`            // 商品类型 0-商品 1-套餐
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
	BatchTagUuid      uint64   `json:"batch_tag_uuid"`  // 分批类型UUID, 可选（前置模式时使用）
	// 后端内部使用的参数
	isH5Product bool `json:"-"` // 是否是H5商品
}

func (req *OrderCartProductAddReq) SetIsH5Product() {
	req.isH5Product = true
}

func (req *OrderCartProductAddReq) IsH5Product() bool {
	return req.isH5Product
}

// OrderProductQuotationReq 订单商品报价请求参数
type OrderProductQuotationReq struct {
	SaleBillUuid  uint64          `json:"sale_bill_uuid" binding:"required"`      // 销售账单UUID
	SaleOrderUuid uint64          `json:"sale_order_uuid" binding:"required"`     // 销售订单UUID
	Products      []ProductParams `json:"products" binding:"required,min=1,dive"` // 商品信息列表
}

// ProductAddReq 添加商品请求参数
type ProductAddReq struct {
	SaleBillUuid   uint64          `json:"sale_bill_uuid"`    // 销售账单ID。
	SaleOrderUuid  uint64          `json:"sale_order_uuid"`   // 销售订单ID。
	Products       []ProductParams `json:"products"`          // 商品信息列表·
	IsH5Product    bool            `json:"is_h5_product"`     // 是否是H5商品
	IsMemberAdd    bool            `json:"is_member_add"`     // 是否是会员端加购
	IsMemberDineIn bool            `json:"is_member_dine_in"` // 是否是会员端堂食下单
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

func (req *TabletOrderCartProductAddReq) FormatPackageSubProductParams() {
	//for _, product := range req.Products {
	//	for j := range product.Products {
	//		subProduct := &product.Products[j]
	//		subProduct.Num = product.Num // 前端传过来的num是错误的. 需要单位数量✖️套餐数量
	//	}
	//}
}

// ProductParams 商品参数
type ProductParams struct {
	FlavorProductBomUuid            uint64           `json:"flavor_product_bom_uuid"`             // 商品规格uuid
	FlavorflavUuid                  uint64           `json:"flavor_uuid"`                         // 商品规格uuid,FlavorProductBomUuid的别名,只有会员端堂食订单创建使用
	Num                             float64          `json:"num"  binding:"required"`             // 商品份数
	UnitNum                         float64          `json:"unit_num"`                            // 单位数量. 目前只有平板的加购并送厨使用该字段
	Price                           *float64         `json:"price"`                               // 商品价格，商品单价。当商品价格与后台设置的最新价格不一致时，加购失败并返回最新价格。可选，不传时，不进行价格校验
	IsBuffet                        *bool            `json:"is_buffet"`                           // 是否是自助餐商品。可选，不填时，表示不判断是不是最新价格。该参数仅在判断价格时使用
	SauceProductBomUuidList         []uint64         `json:"sauce_product_bom_uuid_list"`         // 加料信息
	SauceUuidList                   []uint64         `json:"sauce_uuid"`                          // 加料信息,SauceProductBomUuidList的别名,只有会员端堂食订单创建使用
	ProductPackageAttributeUuidList []uint64         `json:"product_package_attribute_uuid_list"` // 属性信息
	AttributeUuidList               []uint64         `json:"attribute_uuid"`                      // 属性信息,ProductPackageAttributeUuidList的别名,只有会员端堂食订单创建使用
	AddPrice                        float64          `json:"add_price"`                           // 加价金额
	Operation                       string           `json:"operation"`                           // 操作类型。add: 加购，sub: 减购
	MustPlanUuid                    uint64           `json:"must_plan_uuid"`                      // 必点方案uuid. 可选，在必点方案弹窗中加购时填写
	Remark                          string           `json:"remark"`                              // 备注，平板端离线购物车提交
	RemarkUuids                     []uint64         `json:"remark_uuids"`                        // 备注预设UUID列表
	ProductPackageGroupUuid         uint64           `json:"product_package_group_uuid"`          // 套餐分组uuid。可选，当商品是套餐商品时，该字段有值
	ProductType                     uint             `json:"product_type"`                        // 商品类型 0-商品 1-套餐
	Products                        []ProductRequest `json:"products"`                            // 套餐商品请求列表。当商品是套餐商品时，该字段有值
	ProductPackageUuid              uint64           `json:"product_package_uuid"`                // 套餐商品uuid。当商品是套餐商品时，该字段有值

	isPackageProduct        bool   // 是否是套餐商品
	packageSubProductParams string // 套餐子商品参数（JSON格式）

	subProducts         []ProductParams // 套餐子商品列表。当商品是套餐商品时，该字段有值
	isPackageSubProduct bool            // 是否是套餐子商品
	packageUuid         uint64          // 套餐uuid,用于标注套餐子商品的套餐商品（sale_order_product）的uuid
	BatchTagUuid        uint64          `json:"batch_tag_uuid"` // 分批类型UUID, 可选（前置模式时使用）
}

// ApplyCompatibleFields 将兼容字段赋值到标准字段
// 会员端堂食订单创建时,客户端传 flavor_uuid(FlavorflavUuid),需要赋值到 FlavorProductBomUuid
// 会员端堂食订单创建时,客户端传 sauce_uuid(SauceUuidList),需要赋值到 SauceProductBomUuidList
func (req *ProductParams) ApplyCompatibleFields() {
	if req.FlavorflavUuid != 0 {
		req.FlavorProductBomUuid = req.FlavorflavUuid
	}
	if len(req.SauceUuidList) > 0 {
		req.SauceProductBomUuidList = req.SauceUuidList
	}
	if len(req.AttributeUuidList) > 0 {
		req.ProductPackageAttributeUuidList = req.AttributeUuidList
	}
}

func (req *ProductParams) SetIsPackageProduct(subProducts []ProductParams) {
	req.subProducts = subProducts
	req.isPackageProduct = true
}

func (req *ProductParams) SetIsPackageSubProduct(packageUuid uint64) {
	req.isPackageSubProduct = true
	req.packageUuid = packageUuid
}

func (req *ProductParams) SetPackageSubProductParams(params string) {
	req.packageSubProductParams = params
}

func (req *ProductParams) GetPackageSubProductParams() string {
	return req.packageSubProductParams
}

func (req *ProductParams) GetIsPackageProduct() bool {
	return req.isPackageProduct
}

type SubProduct struct {
	FlavorUuid              uint64   `json:"flavor_uuid"`                // 商品规格uuid
	AttributeUuid           []uint64 `json:"attribute_uuid"`             // 属性ID列表
	ProductPackageGroupUuid uint64   `json:"product_package_group_uuid"` // 套餐分组uuid
	Num                     float64  `json:"num"`                        // 套餐子商品数量
	UnitNum                 float64  `json:"unit_num"`                   // 每份数量
}

func (req *ProductParams) GetSubProductList() []SubProduct {
	subProductList := make([]SubProduct, 0)
	for _, subProduct := range req.subProducts {
		subProductList = append(subProductList, SubProduct{
			FlavorUuid:              subProduct.FlavorProductBomUuid,
			AttributeUuid:           subProduct.ProductPackageAttributeUuidList,
			ProductPackageGroupUuid: subProduct.ProductPackageGroupUuid,
			Num:                     subProduct.Num,
			UnitNum:                 subProduct.UnitNum,
		})
	}
	return subProductList
}

// 获取套餐子商品参数
func (req *ProductParams) GetSubProducts() []ProductParams {
	return req.subProducts
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
	SaleBillUuid        uint64 `json:"sale_bill_uuid"`         // 销售账单ID
	IgnoreMust          bool   `json:"ignore_must"`            // 是否忽略必点方案
	H5OrderUuid         uint64 `json:"h5_order_uuid"`          // h5订单ID。默认为0，表示不送厨h5订单商品。当从H5订单进入桌台时，需要传入h5订单ID，将该h5订单的商品送厨
	Password            string `json:"password"`               // 高级密码后台开启的时候才传
	IsCheckCooking      bool   `json:"is_check_cooking"`       // 是否只进行送厨检查，而不进行实际的送厨。场景：助手端开启下单校验高级密码时，先检查送厨，再实际送厨。检查送厨时不进行实际送厨
	IsMemberOrderAccept bool   `json:"is_member_order_accept"` // 是否是会员端接单. 不用返回购物车信息
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

// OrderCartProductWrapReq 打包购物车商品请求参数
type OrderCartProductWrapReq struct {
	SaleBillUuid         uint64 `json:"sale_bill_uuid"`          // 销售账单ID
	SaleOrderUuid        uint64 `json:"sale_order_uuid"`         // 销售订单ID
	SaleOrderProductUuid uint64 `json:"sale_order_product_uuid"` // 销售订单商品ID
}

// OrderCartProductUnwrapReq 取消打包购物车商品请求参数
type OrderCartProductUnwrapReq struct {
	SaleBillUuid         uint64 `json:"sale_bill_uuid"`          // 销售账单ID
	SaleOrderUuid        uint64 `json:"sale_order_uuid"`         // 销售订单ID
	SaleOrderProductUuid uint64 `json:"sale_order_product_uuid"` // 销售订单商品ID
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
