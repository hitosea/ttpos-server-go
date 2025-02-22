package event

import "ttpos-server-go/pkg/eventbus"

// EventAddSaleBillRecord 事件名称，每个事件都有一个全局唯一的名称
const EventAddSaleBillRecord EventName = "Event_Add_Sale_Bill_Record"

// AddSaleBillRecordPayload 每个事件有一个数据结构
type AddSaleBillRecordPayload struct {
	SaleBillUuid     uint64  `json:"sale_bill_uuid"`
	SaleOrderUuid    uint64  `json:"sale_order_uuid"`
	OrderProductUuid uint64  `json:"order_product_uuid"`
	OrderProductName string  `json:"order_product_name"`
	OrderProductNum  uint    `json:"order_product_num"`
	CompanyUuid      uint64  `json:"company_uuid"`
	StaffUuid        uint64  `json:"staff_uuid"`
	Price            float64 `json:"price"`
	AttributeNames   string  `json:"attribute_names"`
	Source           string  `json:"source"`
}

// AddSaleBillRecordHandler 每个事件的处理器
type AddSaleBillRecordHandler func(msg AddSaleBillRecordPayload)

// PublishAddSaleBillRecordEvent 发布添加销售账单记录事件
func (system *SystemEventBus) PublishAddSaleBillRecordEvent(msg AddSaleBillRecordPayload) {
	system.bus.Publish(eventbus.Event{Name: string(EventAddSaleBillRecord), Payload: msg})
}

// SubscribeAddSaleBillRecordEvent 订阅添加销售账单记录事件
func (system *SystemEventBus) SubscribeAddSaleBillRecordEvent(handler AddSaleBillRecordHandler) {
	system.bus.Subscribe(string(EventAddSaleBillRecord), func(event eventbus.Event) {
		msg := event.Payload.(AddSaleBillRecordPayload)
		handler(msg)
	})
}
