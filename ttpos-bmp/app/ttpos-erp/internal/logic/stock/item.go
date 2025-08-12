package stock

import (
	"time"
	"ttpos-bmp/app/ttpos-erp/internal/consts"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto/erp"
	"ttpos-bmp/app/ttpos-erp/internal/service"
	"ttpos-bmp/internal/pkg/queue"
)

var Item = new(sItem)

type sItem struct{}

func init() {
	service.RegisterItem(Item)
}
func (s *sItem) SyncDelay() {
	queue.Push(string(consts.TopicItemSync), &erp.Item{})
	queue.DelayPush(string(consts.TopicItemSyncDelay), &erp.Item{}, 10*time.Second)
}
