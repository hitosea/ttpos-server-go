package event

import (
	"sync"
	"time"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/eventbus/event"
)

var once_change_member_points_event_handler sync.Once

// init 自动注册"会员积分变动"事件处理器
func init() {
	// 只初始化一次
	changeMemberPointsEventHandler()
}

// changeMemberBalanceEventHandler "会员余额变动"事件处理器
func changeMemberPointsEventHandler() {
	once_change_member_points_event_handler.Do(func() {
		event.NewSystemBus().SubscribeChangeMemberPointsEvent(func(payload event.ChangeMemberPointsPayload) {
			db := database.GetDBManager(config.DatabaseConf{}).GetDB(payload.CompanyUuid)
			go func() {
				time.Sleep(400 * time.Microsecond)
				HandleMemberPoints(db)
			}()
		})
	})
}
