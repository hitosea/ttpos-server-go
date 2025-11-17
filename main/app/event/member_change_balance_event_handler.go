package event

import (
	"sync"
	"time"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/eventbus/event"
	"ttpos-server-go/pkg/utils"
)

var once_change_member_balance_event_handler sync.Once

// changeMemberBalanceEventHandler "会员余额变动"事件处理器
func changeMemberBalanceEventHandler() {
	once_change_member_balance_event_handler.Do(func() {
		event.NewSystemBus().SubscribeChangeMemberBalanceEvent(func(payload event.ChangeMemberBalancePayload) {
			db := database.GetDBManager(config.DatabaseConf{}).GetDB(payload.CompanyUuid)
			utils.Go(func() {
				time.Sleep(1 * time.Second)
				HandleMemberBalance(db)
			})
		})
	})
}
