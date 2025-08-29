package queue

import (
	"context"
	"ttpos-server-go/pkg/logger"

	"github.com/apache/rocketmq-client-go/v2/primitive"
	"go.uber.org/zap"
)

// erpDocChangeHandler 处理文档变更
func erpDocChangeHandler(ctx context.Context, msg *primitive.MessageExt) error {

	logger.Logger.Info("处理doc变更", zap.String("msg_id", msg.MsgId))

	return nil
}
