package transfer_order

import (
	"context"
	"fmt"
	"time"
	"ttpos-server-go/pkg/cache"

	"gorm.io/gorm"
)

// transferOrderHelper 调拨单辅助方法
type transferOrderHelper struct{}

// GenerateOrderNo 生成调拨单订单编号
// 格式：TR+12位数字（全平台唯一）
// 组成：8位日期(YYYYMMDD) + 4位序号
func (h *transferOrderHelper) GenerateOrderNo(db *gorm.DB) string {
	now := time.Now()
	dateStr := now.Format("20060102") // 8位日期

	// Redis key: transfer_order:seq:YYYYMMDD
	redisKey := fmt.Sprintf("transfer_order:seq:%s", dateStr)
	ctx := context.Background()

	// 获取Redis客户端（兼容集群和单机模式）
	var seq int64
	var err error

	if clusterClient := cache.Global.GetClusterClient(); clusterClient != nil {
		// 集群模式
		seq, err = clusterClient.Incr(ctx, redisKey).Result()
		if err == nil {
			clusterClient.Expire(ctx, redisKey, 3*24*time.Hour)
			if seq > 9999 {
				clusterClient.Set(ctx, redisKey, 1, 3*24*time.Hour)
				seq = 1
			}
		}
	} else if client := cache.Global.GetClient(); client != nil {
		// 单机模式
		seq, err = client.Incr(ctx, redisKey).Result()
		if err == nil {
			client.Expire(ctx, redisKey, 3*24*time.Hour)
			if seq > 9999 {
				client.Set(ctx, redisKey, 1, 3*24*time.Hour)
				seq = 1
			}
		}
	}

	// Redis失败时降级使用时间戳
	if err != nil {
		return fmt.Sprintf("TR%s%04d", dateStr, now.Unix()%10000)
	}

	// 生成12位数字：TR + 8位日期 + 4位序号
	orderNo := fmt.Sprintf("TR%s%04d", dateStr, seq)
	return orderNo
}
