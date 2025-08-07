package erpnext

import (
	"time"
	"ttpos-bmp/app/ttpos-erp/internal/consts"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto"
	"ttpos-bmp/app/ttpos-erp/internal/service"
	"ttpos-bmp/internal/pkg/queue"
)

var Item = new(sItem)

type sItem struct{}

func init() {
	service.RegisterItem(Item)
}
func (s sItem) SyncDelay() {
	queue.Push(string(consts.TopicItemSync), &dto.Item{})
	queue.DelayPush(string(consts.TopicItemSyncDelay), &dto.Item{}, 10*time.Second)
}
