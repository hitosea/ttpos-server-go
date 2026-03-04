package websocket

import (
	"context"
	"strconv"
	"time"

	"github.com/gogf/gf/v2/frame/g"

	"ttpos-bmp/app/ttpos-websocket/internal/consts"
	"ttpos-bmp/app/ttpos-websocket/internal/dao"
	"ttpos-bmp/app/ttpos-websocket/internal/model/do"
)

// flushHeartbeatToDB 定期将 Redis 中缓存的心跳数据批量写入数据库
// 每 60 秒执行一次，进程退出前最后刷一次
func flushHeartbeatToDB() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-globalCtx.Done():
			// 进程退出前最后刷一次
			doFlushHeartbeat(context.Background())
			return
		case <-ticker.C:
			doFlushHeartbeat(context.Background())
		}
	}
}

// doFlushHeartbeat 执行一次心跳数据从 Redis 刷入 DB
func doFlushHeartbeat(ctx context.Context) {
	// 1. HGETALL 取出所有待刷心跳
	result, err := g.Redis().HGetAll(ctx, consts.RedisKeyDeviceHeartbeat)
	if err != nil {
		g.Log().Warning(ctx, "获取心跳缓存失败", err)
		return
	}

	data := result.Map()
	if len(data) == 0 {
		return
	}

	// 2. 删除 Redis key（后续心跳写入新数据，不会丢失）
	_, err = g.Redis().Del(ctx, consts.RedisKeyDeviceHeartbeat)
	if err != nil {
		g.Log().Warning(ctx, "删除心跳缓存失败", err)
		// 继续执行，数据已取出，最多下次重复刷一次
	}

	// 3. 按 timestamp 分组，相同 timestamp 的 connection_key 合并为一次批量 UPDATE
	grouped := make(map[int64][]string) // timestamp -> []connectionKey
	for connKey, tsVal := range data {
		ts, err := strconv.ParseInt(tsVal.(string), 10, 64)
		if err != nil {
			continue
		}
		grouped[ts] = append(grouped[ts], connKey)
	}

	// 4. 批量 UPDATE
	for ts, connKeys := range grouped {
		_, err := dao.DeviceOnline.Ctx(ctx).
			Where(dao.DeviceOnline.Columns().ConnectionKey, connKeys).
			Where(dao.DeviceOnline.Columns().Status, 1).
			Update(do.DeviceOnline{
				LastHeartbeatTime: int(ts),
				UpdateTime:        int(ts),
			})
		if err != nil {
			g.Log().Warning(ctx, "批量更新心跳时间失败", err,
				"count", len(connKeys),
				"timestamp", ts,
			)
		}
	}

	g.Log().Debug(ctx, "心跳数据已刷入DB", "device_count", len(data))
}
