package event

import (
	"sync"
	"time"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/eventbus/event"
	"ttpos-server-go/pkg/utils"
)

var once_change_member_points_event_handler sync.Once

// changeMemberPointsEventHandler "会员积分变动"事件处理器
func changeMemberPointsEventHandler() {
	once_change_member_points_event_handler.Do(func() {
		event.NewSystemBus().SubscribeChangeMemberPointsEvent(func(payload event.ChangeMemberPointsPayload) {
			db := database.GetDBManager(config.DatabaseConf{}).GetDB(payload.CompanyUuid)
			utils.Go(func() {
				time.Sleep(1 * time.Second)
				HandleMemberPoints(db)
			})
		})
	})
}
