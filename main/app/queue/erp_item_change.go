package queue

import (
	"context"
	"ttpos-server-go/pkg/logger"

	"github.com/apache/rocketmq-client-go/v2/primitive"
	"go.uber.org/zap"
)

// erpItemChangeHandler 处理商品变更
func erpItemChangeHandler(ctx context.Context, msg *primitive.MessageExt) error {
	//TODO 处理商品变更
	logger.Logger.Info("处理商品变更", zap.String("msg_id", msg.MsgId))

	return nil
}
