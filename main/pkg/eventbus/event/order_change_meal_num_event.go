package event

import (
	"ttpos-server-go/pkg/eventbus"
	"ttpos-server-go/pkg/utils"
)

// EventChangeMealNumSaleBill “修改桌台就餐人数”事件名称
const EventChangeMealNumSaleBill EventName = "Event_Change_Meal_Num_Sale_Bill"

// ChangeMealNumSaleBillPayload “修改桌台就餐人数”事件数据结构
type ChangeMealNumSaleBillPayload struct {
	BasePayload
	OldMealNum uint `json:"old_meal_num"` // 旧桌台就餐人数
	NewMealNum uint `json:"new_meal_num"` // 新桌台就餐人数
}

func (payload *ChangeMealNumSaleBillPayload) ToJsonString() string {
	return utils.ToJson(payload)
}

// ChangeMealNumSaleBillHandler 修改桌台就餐人数事件处理器
type ChangeMealNumSaleBillHandler func(msg ChangeMealNumSaleBillPayload)

// PublishChangeMealNumSaleBillEvent 发布修改桌台就餐人数事件
func (system *SystemEventBus) PublishChangeMealNumSaleBillEvent(msg ChangeMealNumSaleBillPayload) {
	system.bus.Publish(eventbus.Event{Name: string(EventChangeMealNumSaleBill), Payload: msg})
}

// SubscribeChangeMealNumSaleBillEvent 订阅修改桌台就餐人数事件
func (system *SystemEventBus) SubscribeChangeMealNumSaleBillEvent(handler ChangeMealNumSaleBillHandler) {
	system.bus.Subscribe(string(EventChangeMealNumSaleBill), func(event eventbus.Event) {
		msg := event.Payload.(ChangeMealNumSaleBillPayload)
		handler(msg)
	})
}
