package event

import (
	"ttpos-server-go/app/dto"
	"ttpos-server-go/pkg/eventbus"
	"ttpos-server-go/pkg/utils"
)

// SentCookingPayload 送厨事件数据结构
type SentCookingPayload struct {
	BasePayload
	Products Products `json:"products"` // 用餐人数
}

// OrderProduct 订单商品
type OrderProduct struct {
	OrderProductId        uint64               `json:"order_product_id"`    // 订单商品ID
	ProductId             uint64               `json:"product_id"`          // 商品ID
	ProductName           dto.LocaleResponse   `json:"product_name"`        // 商品名称
	ProductType           uint8                `json:"product_type"`        // 商品类型
	ProductAttr           dto.LocaleResponse   `json:"product_attr"`        // 商品属性, 包含规格、属性、小料
	ProductAttrList       []dto.LocaleResponse `json:"product_attrs"`       // 商品属性, 包含规格、属性、小料
	ProductSauceNamesList []dto.LocaleResponse `json:"product_sauce_names"` // 商品小料
	Attr                  dto.LocaleResponse   `json:"attr"`                // 商品属性
	AttrList              []dto.LocaleResponse `json:"attr_list"`           // 商品属性 - 列表
	FlavorName            dto.LocaleResponse   `json:"flavor_name"`         // 商品规格
	TotalNum              float64              `json:"total_num"`           // 总数量
	NumType               uint                 `json:"num_type"`            // 数量计算方法, 0-整数 1-小数
	IsBuffet              bool                 `json:"is_buffet"`           // 是否自助餐
	IsWrap                bool                 `json:"is_wrap"`             // 是否打包
	IsGift                bool                 `json:"is_gift"`             // 是否赠品
	IsPackage             bool                 `json:"is_package"`          // 是否套餐
	IsSubProduct          bool                 `json:"is_sub_product"`      // 是否套餐子商品
	Remark                string               `json:"remark"`              // 备注
	RemarkLocale          dto.LocaleResponse   `json:"remark_locale"`       // 备注（包含预设备注和自定义备注，多语言）
	IsBatch               bool                 `json:"is_batch"`            // 是否分批商品
	SubProducts           []OrderProduct       `json:"sub_products"`        // 套餐子商品
}

// Products 送厨商品列表
type Products []OrderProduct

func (payload *SentCookingPayload) ToJsonString() string {
	return utils.ToJson(payload)
}

// SentCookingHandler 送厨事件处理器
type SentCookingHandler func(msg SentCookingPayload)

// PublishSentCookingEvent 发布送厨事件
func (system *SystemEventBus) PublishSentCookingEvent(msg SentCookingPayload) {
	system.bus.Publish(eventbus.Event{Name: string(EventSentCooking), Payload: msg})
}

// SubscribeSentCookingEvent 订阅送厨事件
func (system *SystemEventBus) SubscribeSentCookingEvent(handler SentCookingHandler) {
	system.bus.Subscribe(string(EventSentCooking), func(event eventbus.Event) {
		msg := event.Payload.(SentCookingPayload)
		handler(msg)
	})
}
