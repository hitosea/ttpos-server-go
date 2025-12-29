package event

import (
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/eventbus"
	"ttpos-server-go/pkg/utils"
)

type BasePayload struct {
	Ctx              context.Context `json:"-"`
	CompanyUuid      uint64          `json:"-"`
	Source           string          `json:"-"`
	SaleBillUuid     uint64          `json:"-"`
	SaleOrderUuid    uint64          `json:"-"`
	TakeoutOrderUuid uint64          `json:"-"`
	H5OrderUuid      uint64          `json:"-"`
	OperatorUuid     int64           `json:"-"` // -1 是系统平台； -2 是用户； 正数 是员工
	Staff            model.Staff     `json:"-"`
	// 会员端订单时，MemberSaleOrderUuid 是外送订单Uuid
	MemberSaleOrderUuid uint64 `json:"-"`
	MemberUuid          uint64 `json:"-"`
}

func (base BasePayload) GetOperatorUuid() uint64 {
	if base.OperatorUuid < 0 {
		return 0
	}
	if base.Ctx.GetSource() == constant.SourceAssistant {
		return base.Ctx.GetAssistantUuid()
	}
	return uint64(base.OperatorUuid)
}

// OpenDeskPayload 开台事件数据结构
type OpenDeskPayload struct {
	BasePayload
	MealNum  uint   `json:"meal_num"`  // 用餐人数
	IsBuffet bool   `json:"is_buffet"` // 是否是自助餐
	TableId  uint64 `json:"table_id"`  // 桌台ID
	TableNo  string `json:"table_no"`  // 桌台号
}

func (payload *OpenDeskPayload) ToJsonString() string {
	return utils.ToJson(payload)
}

// OpenDeskHandler 开台事件处理器
type OpenDeskHandler func(msg OpenDeskPayload)

// PublishOpenDeskEvent 发布开台事件
func (system *SystemEventBus) PublishOpenDeskEvent(msg OpenDeskPayload) {
	system.bus.Publish(eventbus.Event{Name: string(EventOpenDesk), Payload: msg})
}

// SubscribeOpenDeskEvent 订阅开台事件
func (system *SystemEventBus) SubscribeOpenDeskEvent(handler OpenDeskHandler) {
	system.bus.Subscribe(string(EventOpenDesk), func(event eventbus.Event) {
		msg := event.Payload.(OpenDeskPayload)
		handler(msg)
	})
}
