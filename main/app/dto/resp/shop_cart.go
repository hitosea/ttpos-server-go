package resp

import (
	"encoding/json"
	"sort"
	"strings"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/resp/product_resp"

	"github.com/shopspring/decimal"
)

// 购物车商品的“规格/属性”
type ProductFlavorAndAttributeRes struct {
	ProductType uint         `json:"product_type"` // 商品类型 0-商品 1-套餐
	ProductInfo *ProductInfo `json:"product_info"` // 商品/套餐信息
}

// 商品信息
type ProductInfo struct {
	Product                product_resp.Product     `json:"product"`                  // 商品/套餐信息
	SelectedProduct        *ProductPackageDetail    `json:"selected_product"`         // 已选购的商品信息
	SelectedProductPackage *ProductSelectedInfoList `json:"selected_product_package"` // 已选购的商品套餐信息
}

// 商品选择信息
type ProductSelectedInfoList struct {
	List []PackageSelectedInfo `json:"list"`
}

// 套餐选择信息
type PackageSelectedInfo struct {
	ProductPackageGroupUuid uint64   `json:"product_package_group_uuid"` // 套餐分组UUID
	FlavorUuid              uint64   `json:"flavor_uuid"`                // 某个规格商品ID
	AttributeUuidList       []uint64 `json:"attribute_uuid"`             // 属性ID列表
	Num                     float64  `json:"num"`                        // 套餐子商品数量
	UnitNum                 float64  `json:"unit_num"`                   // 每份数量
}

type OrderRemarkRes struct {
	List []OrderRemarkResItem `json:"list"`
}

type OrderRemarkResItem struct {
	IsLatest     bool               `json:"is_latest"`     // 是否是最新备注
	Uuids        []uint64           `json:"uuids"`         // 备注UUID列表
	Remark       dto.LocaleResponse `json:"remark"`        // 备注
	CustomRemark string             `json:"custom_remark"` // 自定义备注
	CreateTime   int64              `json:"create_time"`   // 创建时间(时间戳)
}

// 整单备注信息
type OrderRemarkInfo struct {
	List []OrderRemarkItem `json:"list"`
}

func (res *OrderRemarkInfo) ToJson() string {
	json, err := json.Marshal(res)
	if err != nil {
		return ""
	}
	return string(json)
}

func (res *OrderRemarkInfo) GetOrderRemarkResponse() OrderRemarkRes {
	result := make([]OrderRemarkResItem, 0)
	for _, item := range res.List {
		result = append(result, OrderRemarkResItem{
			IsLatest:     item.IsLatest,
			Uuids:        item.Uuids,
			Remark:       item.GetOrderRemarkResponse(),
			CustomRemark: item.Remark, // 自定义备注
			CreateTime:   item.CreateTime,
		})
	}

	// 按创建时间排序，最新的在最后
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreateTime > result[j].CreateTime
	})

	// 判断是否有多个最新的整单备注
	latestMap := make(map[int64]bool)
	for _, item := range result {
		if item.IsLatest {
			latestMap[item.CreateTime] = true
		}
	}
	if len(latestMap) > 1 {
		for index, item := range result {
			if item.IsLatest {
				result[index].IsLatest = false
			}
		}
		result[len(result)-1].IsLatest = true // 最后一个整单备注是最新的
	}

	return OrderRemarkRes{
		List: result,
	}
}

type OrderRemarkItem struct {
	IsLatest   bool                 `json:"is_latest"`   // 是否是最新备注
	Uuids      []uint64             `json:"uuids"`       // 备注UUID列表
	Remarks    []dto.LocaleResponse `json:"remarks"`     // 备注列表
	Remark     string               `json:"remark"`      // 备注，自定义备注
	CreateTime int64                `json:"create_time"` // 创建时间(时间戳)
}

// 获取整单备注响应. item1;item2;custom_remark
func (item *OrderRemarkItem) GetOrderRemarkResponse() dto.LocaleResponse {
	name := dto.LocaleResponse{}
	if len(item.Remarks) > 0 {
		for _, remark := range item.Remarks {
			name.EN += remark.EN + ";"
			name.TH += remark.TH + ";"
			name.ZH += remark.ZH + ";"
			name.ZHTW += remark.ZHTW + ";"
			name.JA += remark.JA + ";"
			name.KO += remark.KO + ";"
			name.MY += remark.MY + ";"
			name.TR += remark.TR + ";"
			name.SV += remark.SV + ";"
		}
	}
	if item.Remark != "" {
		name.EN += item.Remark
		name.TH += item.Remark
		name.ZH += item.Remark
		name.ZHTW += item.Remark
		name.JA += item.Remark
		name.KO += item.Remark
		name.MY += item.Remark
		name.TR += item.Remark
		name.SV += item.Remark
	}
	// 去掉最后一个分号
	name.EN = strings.TrimSuffix(name.EN, ";")
	name.TH = strings.TrimSuffix(name.TH, ";")
	name.ZH = strings.TrimSuffix(name.ZH, ";")
	name.ZHTW = strings.TrimSuffix(name.ZHTW, ";")
	name.JA = strings.TrimSuffix(name.JA, ";")
	name.KO = strings.TrimSuffix(name.KO, ";")
	name.MY = strings.TrimSuffix(name.MY, ";")
	name.TR = strings.TrimSuffix(name.TR, ";")
	name.SV = strings.TrimSuffix(name.SV, ";")
	return name
}

// 桌台购物车
type ShopCart struct {
	SaleBillUuid    uint64               `json:"sale_bill_uuid"`         // 销售账单ID
	Takeout         *bool                `json:"takeout,omitempty"`      // 是否是打包订单，false:堂食订单 true:打包订单。只有点餐订单才有这个字段
	IsDeskOrder     bool                 `json:"is_desk_order"`          // 购物车类型 true:桌台购物车 false:点餐购物车
	IsLock          bool                 `json:"is_lock"`                // 购物车是否锁定 true:锁定 false:未锁定
	OrderSourceUuid uint64               `json:"order_source_uuid"`      // 订单来源UUID（0=店内）
	NationalityUuid uint64               `json:"nationality_uuid"`       // 国籍UUID（0=未记录）
	Desk            *DeskInfo            `json:"desk,omitempty"`         // 桌台信息
	Buffet          *BuffetInfo          `json:"buffet,omitempty"`       // 自助餐信息
	MustPlans       *ProductMustPlanList `json:"must_plans,omitempty"`   // 必点方案列表信息
	DiningMethod    uint                 `json:"dining_method"`          // 用餐方式 0:堂食 1:打包。与Takeout重复，废弃
	SaleOrderList   []SaleOrder          `json:"sale_order_list"`        // 销售订单列表
	UpdateTime      int64                `json:"update_time"`            // 更新时间
	OrderRemark     *OrderRemarkRes      `json:"order_remark,omitempty"` // 整单备注信息
	// 分批送厨的模式
	BatchCookingMode string `json:"batch_cooking_mode"` // 分批送厨的模式 "post"：后置模式 "pre"：前置模式

	Product *product_resp.Product `json:"product,omitempty"` // 商品信息。 当加购商品时商品价格变化时，返回最新的商品信息

	// 是否返回code状态码为-209
	code int // 状态码。 当code为-209时，前端会弹出分批送厨弹窗
}

func (res *ShopCart) SetCode(code int) {
	res.code = code
}
func (res *ShopCart) GetCode() int {
	return res.code
}

// H5CartSendProduct 扫码h5购物车已下单品
type H5CartSendProduct struct {
	Groups H5GroupList `json:"groups"` // 商品列表
}

type H5GroupList struct {
	List []H5Group `json:"list"`
}

type H5Group struct {
	SentKitchenProductGroup
	AcceptTime         int64 `json:"accept_time"`            // 接单时间. “17:20:01 接单”。值为0时不显示。
	H5OrderTime        int64 `json:"h5_order_time"`          // h5下单时间 0表示不是h5下单
	IsH5OrderNeedAudit bool  `json:"is_h5_order_need_audit"` // h5订单是否需要审核，是则需要显示接单/拒单时间
	IsAccept           bool  `json:"is_accept"`              // false:已拒单 true:已接单
}

type H5Product struct {
	Product
}

// UnsentKitchen 未送厨商品
type UnsentKitchen struct {
	Products   CartProductList  `json:"products"`    // 商品列表
	AmountInfo SimpleAmountInfo `json:"amount_info"` // 金额信息
}

// SentKitchen 已送厨商品
type SentKitchen struct {
	Groups     GroupList  `json:"groups"`
	AmountInfo AmountInfo `json:"amount_info"`
}

type GroupList struct {
	List []SentKitchenProductGroup `json:"list"`
}

type SentKitchenProductGroup struct {
	SendKitchenTime int64            `json:"send_kitchen_time"` // 送厨时间. “17:20:01 下单”
	Products        GroupProductList `json:"products"`          // 组商品列表
}

type GroupProductList struct {
	List []Product `json:"list"` // 商品列表
}

// 送厨接口响应：商品 XXX 已下架，请选择其他商品
// 送厨接口响应：规格 商品名称 已下架，请选择其他规格
// 送厨接口响应：以下商品库存不足，请删除后再下单 -商品名稱1（规格大份）-商品名稱2（规格小份）
// 送厨接口响应：已送厨和本次要送厨的商品未选择必点商品，确定要继续送厨吗？· 方案名称1 少点1份 · 方案名称2 少点2份
// 送厨接口响应：以下商品价格有变动，请核对后再下单 -商品名稱1（规格大份）-商品名稱2（规格小份）
// 送厨接口响应：以下商品超出限购数量，请在限购数量内下单 -商品名稱1（规格大份）-商品名稱2（规格小份）
type OrderCheckRes struct {
	Products            *CartProductList     `json:"products,omitempty"`
	ProductMustPlanList *ProductMustPlanList `json:"product_must_plans,omitempty"`
}

type OrderCheckServiceRes struct {
	Code int `json:"code"`
	OrderCheckRes
}

type CartProductList struct {
	List []Product `json:"list"`
}

type InstantShopCart struct {
	SaleBillUuid  uint64      `json:"sale_bill_uuid"`  // 销售账单ID
	DiningMethod  uint        `json:"dining_method"`   // 用餐方式 0:堂食 1:打包
	SaleOrderList []SaleOrder `json:"sale_order_list"` // 销售订单列表
}

type AmountInfo struct {
	ProductOriginalAmount float64 `json:"product_origin_amount"`  // 商品金额(原价)
	ProductAmount         float64 `json:"product_amount"`         // 商品金额(折后价)
	ServiceAmount         float64 `json:"service_amount"`         // 服务费
	TaxAmount             float64 `json:"tax_amount"`             // 税费（商品税费+服务费税费）
	DiscountAmount        float64 `json:"discount_amount"`        // 优惠折扣金额(整单打折优惠金额+订单抹零金额)
	MemberDiscountAmount  float64 `json:"member_discount_amount"` // 会员优惠折扣金额
	Amount                float64 `json:"amount"`                 // 总金额。商品未含税时，总金额=商品金额(折后)+服务费+税费。商品已含税时，总金额=商品金额（折后，含商品消费税）+服务费+税费（只有服务费税）
	ProductNum            float64 `json:"product_num"`            // 总数量，用于点餐助手、平板端、h5
}

// SimpleAmountInfo 简单金额信息。 用于h5购物车、点餐助手
type SimpleAmountInfo struct {
	ProductAmount float64 `json:"product_amount"` // 商品金额(折后价)
	ProductNum    float64 `json:"product_num"`    // 总数量，用于点餐助手、h5
}

// BuffetInfo 自助餐信息
type BuffetInfo struct {
	RemainingSeconds      int64              `json:"remaining_seconds"`       // 自助餐还剩余多少秒。可以为负数，表示自助餐已经结束了多少秒
	IsTimeLimited         bool               `json:"is_time_limited"`         // 是否限时
	LocaleName            dto.LocaleResponse `json:"locale_name"`             // 自助餐名称
	RemainingOrderingTime int64              `json:"remaining_ordering_time"` // 自助餐剩余点餐时间，单位秒
	ReminderOrderTime     int64              `json:"reminder_order_time"`     // 自助餐结束前x秒时提醒即将不可下单，用于助手端、平板端和h5
	IsTabletH5TimeSet     bool               `json:"is_tablet_h5_time_set"`   // 助手、平板、h5时间是否已设置
}

// DeskInfo 桌台信息
type DeskInfo struct {
	Uuid      uint64 `json:"uuid"`       // 桌台ID
	DeskNo    string `json:"desk_no"`    // 桌台编号
	MealNum   uint   `json:"meal_num"`   // 就餐人数
	StartTime int64  `json:"start_time"` // 开台时间
	Duration  int64  `json:"duration"`   // 用餐时长，单位：秒，表示从开台时间到当前时间的时间差。避免前端计算时间不准
}

// SaleOrder 购物车销售订单信息
type SaleOrder struct {
	Uuid                uint64     `json:"uuid"`
	OrderNo             string     `json:"order_no"`
	Status              uint       `json:"status"`                // 订单状态, 0-未结账 1-已结账
	IsDiscount          bool       `json:"is_discount"`           // 是否存在折扣 true:存在 false:不存在
	IsMemberDiscount    bool       `json:"is_member_discount"`    // 是否存在会员优惠折扣 true:存在 false:不存在
	CustomDiscountRate  float64    `json:"custom_discount_rate"`  // 订单改价折扣率
	ZeroRule            uint8      `json:"zero_rule"`             // 订单抹零规则
	AutoDiscountMessage string     `json:"auto_discount_message"` // 优惠折扣-自动抹零信息. 如"优惠折扣自动抹零-抹分" "优惠折扣自动抹零-抹角" "优惠折扣自动抹零-四舍五入保留一位小数" "优惠折扣自动抹零-四舍五入保留整数"
	ProductList         []Product  `json:"product_list"`          // 商品列表
	ProductNum          float64    `json:"product_num"`           // 商品数量
	AmountInfo          AmountInfo `json:"amount_info"`
}

// Product 购物车商品
type Product struct {
	Uuid                uint64             `json:"uuid"`                  // 商品uuid
	ProductPackageUuid  uint64             `json:"product_package_uuid"`  // 商品包uuid
	MustPlanUuid        uint64             `json:"must_plan_uuid"`        // 必点方案uuid
	LocaleName          dto.LocaleResponse `json:"locale_name"`           // 商品名称。商品名称、自助餐名称、自助餐加钟名称
	LocaleAttributeName dto.LocaleResponse `json:"locale_attribute_name"` // 商品属性
	Num                 float64            `json:"num"`                   // 数量
	NumType             uint               `json:"num_type"`              // 数量计算方法 0-整数 1-小数
	FinishedNum         float64            `json:"finished_num"`          // 制作完成数量
	UnitPrice           float64            `json:"unit_price"`            // 单价（折前）
	SalePrice           float64            `json:"price"`                 // 原价. 原价=单价*数量
	DiscountPrice       float64            `json:"discount_price"`        // 折扣价,折后。折扣价不等于原价时，前端要显示出折扣价。单价(折后)*数量
	Status              int                `json:"status"`                // 0: 未送厨 1:已送厨 2:制作完成（出餐）
	Remark              string             `json:"remark"`                // 备注
	IsMust              bool               `json:"is_must"`               // 是否必点
	IsGift              bool               `json:"is_gift"`               // 是否是赠菜
	IsWrap              bool               `json:"is_wrap"`               // 是否是打包
	IsBuffet            bool               `json:"is_buffet"`             // 是否是自助餐
	IsCancel            bool               `json:"is_cancel"`             // 是否退菜
	CanChangeNum        bool               `json:"can_change_num"`        // 顾客可修改必点数量
	AboutBuffet         AboutBuffet        `json:"about_buffet"`          // 自助餐信息
	IsShowKitchen       uint               `json:"is_show_kitchen"`       // 是否在厨显端显示
	ProductType         uint               `json:"product_type"`          // 商品类型 0-商品 1-套餐
	PackageProductList  PackageProductList `json:"package_product_list"`  // 套餐商品列表
	AddPrice            float64            `json:"add_price"`             // 加价金额（套餐主商品的加价总和）
	CanEdit             bool               `json:"can_edit"`              // 是否可以编辑
	IsBatch             bool               `json:"is_batch"`              // 是否是分批商品
	ShowDelayTag        bool               `json:"show_delay_tag"`        // 是否显示延迟送厨标签. 表示该商品是分批送厨商品,目前处理预送厨状态
	ShowBatchTag        bool               `json:"show_batch_tag"`        // 是否显示分批类型
	BatchTagName        dto.LocaleResponse `json:"batch_tag_name"`        // 分批类型名称
	BatchTagColor       string             `json:"batch_tag_color"`       // 分批类型颜色
	BatchTagUuid        uint64             `json:"batch_tag_uuid"`        // 分批类型UUID
	// 后端使用，前端不返回
	CreateTime         int64   `json:"-"` // 创建时间（点餐助手未送厨）
	SendKitchenTime    int64   `json:"-"` // 送厨时间
	H5OrderTime        int64   `json:"-"` // h5下单时间
	IsH5OrderNeedAudit bool    `json:"-"` // h5订单是否需审核
	AcceptTime         int64   `json:"-"` // 接单时间
	IsAccept           bool    `json:"-"` // 是否已接单
	Sign               string  `json:"-"` // 签名，用于合并商品
	CanReturnNum       uint    `json:"-"` // 可退货数量
	CanReturnAmount    float64 `json:"-"` // 可退款金额
	TotalPrice         float64 `json:"-"` // 总价（单个商品、折后）
}

type PackageProductList struct {
	List []PackageProduct `json:"list"`
}

type PackageProduct struct {
	Uuid                uint64             `json:"uuid"`                  // 商品uuid. sale_order_product_uuid
	LocaleName          dto.LocaleResponse `json:"locale_name"`           // 商品名称
	LocaleAttributeName dto.LocaleResponse `json:"locale_attribute_name"` // 商品属性
	Num                 float64            `json:"num"`                   // 数量
	UnitNum             float64            `json:"unit_num"`              // 单位数量
	AddPrice            float64            `json:"add_price"`             // 加价金额
}

// GetPrice 获取商品价格(折后价)
func (p *Product) GetPrice() float64 {
	return decimal.NewFromFloat(p.DiscountPrice).Round(2).InexactFloat64()
}

type AboutBuffet struct {
	IsCustomer       bool   `json:"is_customer"`        // 是否是自助餐顾客
	IsDelay          bool   `json:"is_delay"`           // 是否是加钟
	BuffetUuid       uint64 `json:"buffet_uuid"`        // 自助餐Id
	CustomerTypeUuid uint64 `json:"customer_type_uuid"` // 自助餐顾客类型uuid
}

type OrderFinishResp struct {
	SaleBillUuid  uint64        `json:"sale_bill_uuid"`  // 销售账单uuid
	SaleOrderUuid uint64        `json:"sale_order_uuid"` // 销售订单uuid
	AmountInfo    PayAmountInfo `json:"amount_info"`     // 金额信息
	PayMethodList PayMethodList `json:"pay_methods"`     // 支付方式列表
}

type PayAmountInfo struct {
	OrderAmount  float64 `json:"order_amount"`  // 订单应收金额
	PayAmount    float64 `json:"pay_amount"`    // 订单实收金额
	ChangeAmount float64 `json:"change_amount"` // 找零金额
}
type PayMethodList struct {
	List []PayMethod `json:"list"`
}

type PayMethod struct {
	Uuid uint64 `json:"uuid"` // 支付方式uuid
	Name string `json:"name"` // 支付方式名称
}

// BuffetProductList 自助餐商品列表 或 必点商品列表
type BuffetProductList struct {
	List []BuffetProduct `json:"list"` // 商品列表
}

type BuffetProduct struct {
	Uuid  uint64 `json:"uuid"`  // 商品product_package_uuid
	Name  string `json:"name"`  // 商品名称. 不用于前端展示，仅用于开发核对接口数据是否正确
	Limit uint   `json:"limit"` // 限购数量， 限购数量=单人限购数量*用餐人数
}

type ProductPackageDetail struct {
	FlavorUuid     uint64   `json:"flavor_uuid"`     // 商品规格uuid
	AttributesUuid []uint64 `json:"attributes_uuid"` // 商品属性uuid列表
	SaucesUuid     []uint64 `json:"sauces_uuid"`     // 商品加料uuid列表
	Num            float64  `json:"num"`             // 数量
}

type ProductPackageDetailRes struct {
	List []ProductPackageDetail `json:"list"` // 数据
}

type OrderCartProductBatchCookingRes struct {
	List []OrderCartProductBatchCooking    `json:"list"` // 数据
	Tags []OrderCartProductBatchCookingTag `json:"tags"` // 分批类型列表
}

type OrderCartProductBatchCooking struct {
	Uuid                uint64             `json:"uuid"`                  // 商品uuid sale_order_product_uuid
	LocaleName          dto.LocaleResponse `json:"locale_name"`           // 商品名称
	LocaleAttributeName dto.LocaleResponse `json:"locale_attribute_name"` // 商品属性
	Image               string             `json:"image"`                 // 商品图片URL
	BatchTagUuid        uint64             `json:"batch_tag_uuid"`        // 分批类型uuid, 如果已经进行分批送厨，则不为0
	BatchTime           int64              `json:"batch_time"`            // 分批送厨时间，如果已经进行分批送厨，则不为0
	SendKitchenTime     int64              `json:"send_kitchen_time"`     // 送厨时间. 预送厨的时间。用于排序
	CreateTime          int64              `json:"create_time"`           // 下单时间。用于排序
}

type OrderCartProductBatchCookingTag struct {
	Uuid       uint64             `json:"uuid"`        // 分批类型uuid
	LocaleName dto.LocaleResponse `json:"locale_name"` // 分批类型名称，多语言
	Color      string             `json:"color"`       // 分批类型颜色，如 #FF0000
	Sort       uint               `json:"sort"`        // 排序
	Count      uint               `json:"count"`       // 商品数量
}
