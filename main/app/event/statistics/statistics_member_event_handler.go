package event

import (
	"sync"
	"ttpos-server-go/app/service"
	"ttpos-server-go/pkg/eventbus/event"
)

var once_statistics_member_event_handler sync.Once

// StatisticsMemberEventHandler "统计会员"事件处理器
func StatisticsMemberEventHandler() {
	once_statistics_member_event_handler.Do(func() {
		event.NewSystemBus().SubscribeStatisticsMemberEvent(func(payload event.StatisticsMemberPayload) {
			service.NewStatisticsSrv().SaveMember(payload.Ctx, service.SaveMemberReq{
				MemberRechargeOrderUuid: payload.MemberRechargeOrderUuid,
				OnlyDelete:              payload.OnlyDelete,
			})
		})
	})
}
