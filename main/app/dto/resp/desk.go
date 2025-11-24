package resp

import (
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/resp/product_resp"
)

type Desk struct {
	Uuid                   uint64  `json:"uuid"`                       // 桌台UUID
	DeskNo                 string  `json:"desk_no"`                    // 桌台名称
	CustomerCount          uint    `json:"customer_count"`             // 桌台人数
	Status                 uint    `json:"status"`                     // 桌台状态  0-未开台 1-已开台
	IsLock                 bool    `json:"is_lock"`                    // 是否锁单
	IsBuffet               bool    `json:"is_buffet"`                  // 是否自助餐
	IsWait                 bool    `json:"is_wait"`                    // 是否待清台
	Time                   int64   `json:"time"`                       // 桌台用餐时间（秒）
	LockTime               int64   `json:"lock_time"`                  // 锁单时间
	Price                  float64 `json:"price"`                      // 桌台价格
	Remark                 string  `json:"remark"`                     // 桌台备注
	TypeUuid               uint64  `json:"type_uuid"`                  // 桌台类型ID
	RegionUuid             uint64  `json:"region_uuid"`                // 桌台区域ID
	SaleBillUuid           uint64  `json:"sale_bill_uuid"`             // 销售账单UUID
	SaleOrderUuid          uint64  `json:"sale_order_uuid"`            // 第一个销售订单UUID
	IsSplitOrder           bool    `json:"is_split_order"`             // 是否拆单
	UpdateTime             int64   `json:"update_time"`                // 更新时间
	IsDisabled             bool    `json:"-"`                          // 是否被禁用
	DefaultPeopleNum       uint    `json:"default_people_num"`         // 默认人数
	IsOpenDefaultPeopleNum bool    `json:"is_open_default_people_num"` // 是否开启默认人数
	BatchTagColor          string  `json:"batch_tag_color"`            // 桌台分批类型的颜色
}

type DeskNo struct {
	DeskNo string `json:"desk_no"`
}

type DeskExtra struct {
	AvailableNum       uint  `json:"available_num"`         // 桌台可用数量
	LockNum            uint  `json:"lock_num"`              // 桌台锁定数量
	OccupyBuffetNum    uint  `json:"occupy_buffet_num"`     // 桌台自助餐数量
	OccupyNotBuffetNum uint  `json:"occupy_not_buffet_num"` // 桌台非自助餐数量
	OccupyWaitNum      uint  `json:"occupy_wait_num"`       // 桌台待清台数量
	TotalNum           uint  `json:"total_num"`             // 桌台总计数量
	UpdateTime         int64 `json:"update_time"`           // 更新时间

	// 分批信息
	BatchTags []BatchTagRes `json:"batch_tags"`
}

type BatchTagRes struct {
	Uuid       uint64             `json:"uuid"`        // 分批类型的uuid
	LocaleName dto.LocaleResponse `json:"locale_name"` // 分批类型的名称
	Color      string             `json:"color"`       // 分批类型的颜色
	Sort       int                `json:"sort"`        // 分批类型的排序
	Count      uint               `json:"count"`       // 桌台分批类型的数量
}

type DeskRegion struct {
	Uuid      uint64 `json:"uuid"`        // 餐桌区域ID
	Name      string `json:"name"`        // 餐桌区域名称
	IsOpenMap bool   `json:"is_open_map"` // 是否开启地图
}

type DeskType struct {
	Uuid uint64 `json:"uuid"` // 餐桌类型ID
	Name string `json:"name"` // 餐桌类型名称
}

// DeskRegionAndTypeListWithPaginationResp 桌台区域和类型列表响应
type DeskRegionAndTypeListWithPaginationResp struct {
	UpdateTime int64    `json:"update_time"` // 更新时间
	Region     struct { // 餐桌区域
		List []DeskRegion `json:"list"`
	} `json:"region"`
	Type struct { // 餐桌类型
		List []DeskType `json:"list"`
	} `json:"type"`
}

// DeskListWithPaginationResp 桌台列表响应
type DeskListWithPaginationResp struct {
	Extra DeskExtra        `json:"extra"` // 桌台额外信息
	List  []Desk           `json:"list"`  // 桌台列表
	Meta  dto.PageResponse `json:"meta"`  // 分页信息
}

// DeskInfoResp 桌台详情响应
type DeskInfoResp struct {
	Uuid          uint64 `json:"uuid"`           // 桌台UUID
	SaleBillUuid  uint64 `json:"sale_bill_uuid"` // 订单UUID
	DeskNo        string `json:"desk_no"`        // 桌台名称
	TypeUuid      uint64 `json:"type_uuid"`      // 桌台类型ID
	RegionUuid    uint64 `json:"region_uuid"`    // 桌台区域ID
	Status        uint   `json:"status"`         // 桌台状态
	CustomerCount uint   `json:"customer_count"` // 桌台人数
	IsLock        bool   `json:"is_lock"`        // 是否锁单
	IsBuffet      bool   `json:"is_buffet"`      // 是否自助餐
	Remark        string `json:"remark"`         // 桌台备注
	Time          uint   `json:"time"`           // 桌台用餐时间（秒）
}

type DeskPing struct {
	UnsentKitchenInfo   UnsentKitchenInfo      `json:"unsent_kitchen_info"`    // 未送厨商品信息，即将废弃，随时被删掉，信息从unsent_kitchen中读取
	DeskInfo            Desk                   `json:"desk_info"`              // 桌台信息
	SentKitchen         SentKitchen            `json:"sent_kitchen"`           // 已送厨商品信息
	UnsentKitchen       UnsentKitchen          `json:"unsent_kitchen"`         // 未送厨商品信息
	SentKitchenProducts SentKitchenProductList `json:"sent_kitchen_products"`  // 已送厨商品列表，用于显示送厨数量和完成数量
	Buffet              BuffetInfo             `json:"buffet"`                 // 自助餐信息
	MustPlans           ProductMustPlanList    `json:"must_plans"`             // 必点方案列表
	SaleOrderList       []SaleOrder            `json:"sale_order_list"`        // 销售订单列表
	UpdateTime          int64                  `json:"update_time"`            // 更新时间
	Product             *product_resp.Product  `json:"product,omitempty"`      // 商品信息。 当加购商品时商品价格变化时，返回最新的商品信息
	OrderRemark         *OrderRemarkRes        `json:"order_remark,omitempty"` // 整单备注信息
}

type H5DeskPing struct {
	DeskInfo      Desk                  `json:"desk_info"`              // 桌台信息
	SentKitchen   H5CartSendProduct     `json:"sent_kitchen"`           // 已送厨商品信息
	UnsentKitchen UnsentKitchen         `json:"unsent_kitchen"`         // 未送厨商品信息
	Buffet        BuffetInfo            `json:"buffet"`                 // 自助餐信息
	MustPlans     ProductMustPlanList   `json:"must_plans"`             // 必点方案列表
	MustProducts  BuffetProductList     `json:"must_products"`          // 必点商品列表
	UpdateTime    int64                 `json:"update_time"`            // 更新时间
	Product       *product_resp.Product `json:"product,omitempty"`      // 商品信息。 当加购商品时商品价格变化时，返回最新的商品信息
	OrderRemark   *OrderRemarkRes       `json:"order_remark,omitempty"` // 整单备注信息
}

type UnsentKitchenInfo struct {
	ProductNum    uint    `json:"product_num"`    // 未下单商品数量
	ProductAmount float64 `json:"product_amount"` // 未下单商品金额
}

type SentKitchenProduct struct {
	ProductPackageUuid uint64  `json:"product_package_uuid"`     // 商品Uuid
	SentKitchenNum     float64 `json:"sent_kitchen_product_num"` // 已送厨商品数量
	FinishedNum        float64 `json:"finished_num"`             // 制作完成数量
}
type SentKitchenProductList struct {
	List []SentKitchenProduct `json:"list"`
}

// CreateDeskOrderResp 创建桌台订单响应
type CreateDeskOrderResp struct {
	SaleBillUuid  uint64 `json:"sale_bill_uuid"`  // 销售账单UUID
	SaleOrderUuid uint64 `json:"sale_order_uuid"` // 销售订单UUID
}

type CreateMemberOrderResp struct {
	MemberSaleOrderInfo *MemberSaleOrderInfo `json:"member_sale_order_info"` // 会员端销售订单信息。当提交订单成功后，返回会员端销售订单信息
}

type MemberSaleOrderInfo struct {
	MemberSaleOrderUuid uint64                     `json:"member_sale_order_uuid"` // 会员端销售订单UUID
	Status              uint                       `json:"status"`                 // 会员端销售订单状态。0-选购中 1-待付款 2-待商家接单 3-商家备餐中 4-待骑手接单 5-骑手正在赶往商家 6-骑手配送中 7-已完成 8-已取消
	ProductList         MemberSaleOrderProductList `json:"product_list"`           // 会员端销售订单商品列表
	ProductAmount       float64                    `json:"product_amount"`         // 商品金额(合计)
	MemberDiscount      float64                    `json:"member_discount"`        // 会员折扣金额。大于0时才显示
	Amount              float64                    `json:"amount"`                 // 订单金额。订单金额=商品金额(合计)-会员折扣金额+配送费
	Remark              string                     `json:"remark"`                 // 备注
	Address             MemberSaleOrderAddress     `json:"address"`                // 收货地址
	DeliveryFee         MemberSaleOrderDeliveryFee `json:"delivery_fee"`           // 配送费信息
	PaymentMethods      PaymentMethodList          `json:"payment_methods"`        // 支付方式列表。只显示lianlianpay的微信支付、支付宝支付、QRPromptPay支付
	IsVerifiedPhone     bool                       `json:"is_verified_phone"`      // 订单是否已经验证手机号
	IsInDeliveryRange   bool                       `json:"is_in_delivery_range"`   // 是否在配送范围内。如果不在配送范围内，则置灰提交订单按钮ß

	// 内部使用，用于记录操作日志
	SaleBillUuid  uint64 `json:"-"` // 销售账单UUID
	SaleOrderUuid uint64 `json:"-"` // 销售订单UUID
}

type MemberSaleOrderProductList struct {
	List []MemberSaleOrderProduct `json:"list"`
}

// 会员端销售订单商品
type MemberSaleOrderProduct struct {
	SaleOrderProductUuid uint64             `json:"sale_order_product_uuid"` // 销售订单商品UUID
	LocaleName           dto.LocaleResponse `json:"locale_name"`             // 商品名称。
	LocaleAttributeName  dto.LocaleResponse `json:"locale_attribute_name"`   // 商品属性
	Num                  float64            `json:"num"`                     // 数量
	UnitPrice            float64            `json:"unit_price"`              // 单价（折前）,含税费
	Amount               float64            `json:"amount"`                  // 商品金额（折前）。商品金额=（商品规格价 + 商品小料A价格 + 商品小料B价格）*数量。商品规格价=商品规格原价*会员端折扣率。 商品小料价格=商品小料原价*会员端折扣率。 会员折扣率是指商品在会员端价格上浮比例，一般比堂食贵
	Image                string             `json:"image"`                   // 商品图片URL
}

// 会员端销售订单地址
type MemberSaleOrderAddress struct {
	MemberAddressUuid  uint64 `json:"member_address_uuid"`  // 会员地址UUID
	Longitude          string `json:"longitude"`            // 经度
	Latitude           string `json:"latitude"`             // 纬度
	Address            string `json:"address"`              // 地址
	DetailAddress      string `json:"detail_address"`       // 详细地址。如门牌号
	ContactName        string `json:"contact_name"`         // 联系人
	ContactPhone       string `json:"contact_phone"`        // 联系电话
	ContactPhonePrefix string `json:"contact_phone_prefix"` // 联系电话前缀
	ContactGender      int    `json:"contact_gender"`       // 联系人性别, 0-女士 1-先生
}

// 会员端销售订单配送费信息
type MemberSaleOrderDeliveryFee struct {
	Amount   float64 `json:"amount"`     // 配送费。配送费=基础配送费+（配送距离*每公里配送费）。未达到起步配送按起步配送费收取
	Distance float64 `json:"distance"`   // 配送距离，单位km。选择地址后才有值
	MinFee   float64 `json:"min_fee"`    // 起步配送费，最低配送的费用。未达到起步配送按起步配送费收取
	BaseFee  float64 `json:"base_fee"`   // 基础配送费
	FeePerKm float64 `json:"fee_per_km"` // 每公里配送费
}

// 支付方式信息
type PaymentMethod struct {
	Uuid      uint64 `json:"uuid"`       // 支付方式UUID
	Name      string `json:"name"`       // 支付方式名称
	IsDefault bool   `json:"is_default"` // 是否默认
}

// 合并桌台接口响应 以下桌台已被拆单，不支持合并桌台"
type DeskMergeCheckResp struct {
	List []DeskNo `json:"list"`
}

type DeskMergeShopCartResp struct {
	IsResetDiscount bool      `json:"is_reset_discount"`
	ShopCart        *ShopCart `json:"shop_cart"`
}
