# 设备在线状态追踪功能说明

## 概述
本功能用于记录和追踪 WebSocket 连接的设备在线状态，包括连接建立、心跳更新、连接断开等完整生命周期。

**重要特性**：同一个设备（基于 `company_uuid` + `device_id` + `source_client` 组合）在表中只保留一条记录，每次重新连接时会更新该记录而不是创建新记录。

## 数据库表结构

### device_online 表
设备在线状态记录表，用于存储设备的连接状态和历史记录。

#### 字段说明
| 字段名 | 类型 | 说明 |
|--------|------|------|
| id | bigint(20) | 主键ID，自增长 |
| company_uuid | bigint(20) | 公司UUID |
| staff_uuid | bigint(20) | 员工UUID |
| device_id | varchar(100) | 设备ID |
| source_client | varchar(50) | 来源客户端（pos/tablet/kitchen/h5/mobile） |
| connection_key | varchar(200) | 连接唯一标识 |
| status | tinyint(4) | 在线状态：0-离线，1-在线 |
| connect_time | int(11) | 连接时间戳 |
| disconnect_time | int(11) | 断开时间戳 |
| last_heartbeat_time | int(11) | 最后心跳时间戳 |
| ip_address | varchar(50) | 客户端IP地址 |
| user_agent | varchar(500) | 用户代理信息 |
| create_time | int(11) | 创建时间戳 |
| update_time | int(11) | 更新时间戳 |
| delete_time | int(11) | 删除时间戳（软删除） |

#### 索引设计
- `PRIMARY KEY (id)`: 主键索引
- `UNIQUE KEY idx_connection_key (connection_key)`: 连接唯一性索引
- `KEY idx_company_device (company_uuid, device_id)`: 公司设备复合索引
- `KEY idx_status (status)`: 状态查询索引
- `KEY idx_staff_uuid (staff_uuid)`: 员工查询索引
- `KEY idx_source_client (source_client)`: 客户端类型索引
- `KEY idx_last_heartbeat (last_heartbeat_time)`: 心跳时间索引
- `KEY idx_connect_time (connect_time)`: 连接时间索引

## 功能实现

### 1. 连接建立时
当设备成功建立 WebSocket 连接时，系统会：
- 检查该设备是否已有记录（基于 `company_uuid` + `device_id` + `source_client`）
- **如果记录已存在**：更新为在线状态，更新连接时间、IP地址等信息
- **如果记录不存在**：创建新记录
- 生成唯一的连接标识（connection_key）
- 将状态设置为在线（status=1）
- 初始化心跳时间

**代码位置**: `internal/logic/websocket/websocket.go` - `handleConnectionSuccess` 方法

```go
// 创建或更新设备在线记录（同一设备只保留一条记录）
nowTime := int(time.Now().Unix())

// 先查询是否已存在该设备的记录
var existingRecord *entity.DeviceOnline
err = dao.DeviceOnline.Ctx(ctx).
    Where(dao.DeviceOnline.Columns().CompanyUuid, companyUuid).
    Where(dao.DeviceOnline.Columns().DeviceId, deviceId).
    Where(dao.DeviceOnline.Columns().SourceClient, client).
    Scan(&existingRecord)

if err == nil && existingRecord != nil {
    // 记录已存在，更新为在线状态
    _, err = dao.DeviceOnline.Ctx(ctx).
        Where(dao.DeviceOnline.Columns().Id, existingRecord.Id).
        Update(do.DeviceOnline{
            StaffUuid:         int64(staffUuid),
            ConnectionKey:     connKey,
            Status:            1, // 在线
            ConnectTime:       nowTime,
            LastHeartbeatTime: nowTime,
            IpAddress:         ipAddress,
            UserAgent:         userAgent,
            UpdateTime:        nowTime,
        })
} else {
    // 记录不存在，创建新记录
    deviceOnlineRecord := &do.DeviceOnline{
        CompanyUuid:       int64(companyUuid),
        StaffUuid:         int64(staffUuid),
        DeviceId:          deviceId,
        SourceClient:      client,
        ConnectionKey:     connKey,
        Status:            1, // 在线
        ConnectTime:       nowTime,
        LastHeartbeatTime: nowTime,
        IpAddress:         ipAddress,
        UserAgent:         userAgent,
        CreateTime:        nowTime,
        UpdateTime:        nowTime,
    }
    dao.DeviceOnline.Ctx(ctx).Data(deviceOnlineRecord).Insert()
}
```

### 2. 心跳更新时
当收到设备心跳消息时，系统会：
- 更新最后心跳时间（last_heartbeat_time）
- 更新记录的更新时间（update_time）

**代码位置**: `internal/logic/websocket/websocket.go` - `handleHeartbeat` 方法

```go
// 更新数据库中的心跳时间
if connKey != "" {
    s.updateDeviceHeartbeat(ctx, connKey)
}
```

### 3. 连接断开时
当设备断开连接时（正常断开或异常断开），系统会：
- 将状态更新为离线（status=0）
- 记录断开时间（disconnect_time）
- 更新记录的更新时间（update_time）

**代码位置**: `internal/logic/websocket/websocket.go` - 多处调用 `updateDeviceOffline` 方法

触发场景：
1. 客户端主动关闭连接
2. 连接异常中断
3. 服务端主动关闭连接（CloseConnection）
4. 服务器关闭时清理所有连接（CloseAllConnections）

## 辅助方法

### updateDeviceHeartbeat
更新设备心跳时间

```go
func (s *sWebSocket) updateDeviceHeartbeat(ctx context.Context, connectionKey string) {
    nowTime := int(time.Now().Unix())
    _, err := dao.DeviceOnline.Ctx(ctx).
        Where(dao.DeviceOnline.Columns().ConnectionKey, connectionKey).
        Where(dao.DeviceOnline.Columns().Status, 1).
        Update(do.DeviceOnline{
            LastHeartbeatTime: nowTime,
            UpdateTime:        nowTime,
        })
    // ...
}
```

### updateDeviceOffline
更新设备为离线状态

```go
func (s *sWebSocket) updateDeviceOffline(ctx context.Context, connectionKey string) {
    nowTime := int(time.Now().Unix())
    _, err := dao.DeviceOnline.Ctx(ctx).
        Where(dao.DeviceOnline.Columns().ConnectionKey, connectionKey).
        Update(do.DeviceOnline{
            Status:         0, // 离线
            DisconnectTime: nowTime,
            UpdateTime:     nowTime,
        })
    // ...
}
```

### getClientIP
获取客户端真实IP地址

```go
func getClientIP(r *http.Request) string {
    // 优先从 X-Forwarded-For 头获取
    // 其次从 X-Real-IP 头获取
    // 最后从 RemoteAddr 获取
}
```

## 使用场景

### 1. 设备在线状态查询
可以通过以下条件查询设备在线状态：
- 按公司UUID查询
- 按设备ID查询
- 按员工UUID查询
- 按来源客户端查询
- 按在线状态查询

示例SQL：
```sql
-- 查询某公司下所有在线设备
SELECT * FROM device_online 
WHERE company_uuid = ? AND status = 1;

-- 查询某设备的连接历史
SELECT * FROM device_online 
WHERE company_uuid = ? AND device_id = ? 
ORDER BY connect_time DESC;

-- 查询某时间段内的连接记录
SELECT * FROM device_online 
WHERE company_uuid = ? 
  AND connect_time >= ? 
  AND connect_time <= ?;
```

### 2. 设备在线时长统计
可以通过 connect_time 和 disconnect_time 计算设备在线时长：

```sql
-- 计算设备在线时长（秒）
SELECT 
    device_id,
    source_client,
    SUM(disconnect_time - connect_time) as total_online_seconds
FROM device_online 
WHERE company_uuid = ? 
  AND status = 0  -- 已断开的连接
  AND disconnect_time > 0
GROUP BY device_id, source_client;
```

### 3. 心跳超时检测
可以通过 last_heartbeat_time 检测心跳超时的设备：

```sql
-- 查询超过2分钟没有心跳的在线设备
SELECT * FROM device_online 
WHERE status = 1 
  AND last_heartbeat_time < UNIX_TIMESTAMP(NOW() - INTERVAL 2 MINUTE);
```

### 4. 设备连接统计
可以统计设备的连接次数、平均在线时长等：

```sql
-- 统计每个设备的连接次数
SELECT 
    device_id,
    source_client,
    COUNT(*) as connection_count,
    MAX(connect_time) as last_connect_time
FROM device_online 
WHERE company_uuid = ?
GROUP BY device_id, source_client;
```

## 注意事项

1. **设备唯一性**：
   - 同一设备的唯一标识为：`company_uuid` + `device_id` + `source_client` 组合
   - 每个设备在表中只保留一条记录
   - 重新连接时会更新现有记录，而不是创建新记录

2. **连接标识**：
   - 每个 WebSocket 连接都有唯一的 connection_key，格式为：`{company_uuid}_{client}_{device_id}_{timestamp_nano}`
   - connection_key 在每次连接时都会更新

3. **状态管理**：
   - status=1 表示在线
   - status=0 表示离线
   - 只有在线状态的记录才会更新心跳时间

4. **时间字段**：
   - 所有时间字段使用 Unix 时间戳（秒）
   - connect_time：最后一次连接建立时间
   - disconnect_time：最后一次连接断开时间（在线时为0）
   - last_heartbeat_time：最后心跳时间

5. **IP地址获取**：
   - 优先从 X-Forwarded-For 头获取（代理环境）
   - 其次从 X-Real-IP 头获取
   - 最后从 RemoteAddr 获取（直连）

6. **性能考虑**：
   - 心跳更新是异步的，不会阻塞消息处理
   - 使用了合适的索引优化查询性能
   - 使用"插入或更新"模式，避免重复记录
   - 建议定期清理长时间离线的记录

## 数据库迁移

### 执行迁移
```bash
# 在 ttpos-websocket 目录下执行
cd ttpos-bmp/app/ttpos-websocket

# 使用 Docker 执行迁移
make db_up.docker
```

### 回滚迁移
```bash
# 回滚最后一次迁移
make db_down
```

### 生成代码
```bash
# 生成 DAO、DO、Entity 代码
make dao
```

## 相关文件

- 迁移文件：`manifest/sql/20251115173527_add_device_online_table.up.sql`
- 回滚文件：`manifest/sql/20251115173527_add_device_online_table.down.sql`
- DAO 文件：`internal/dao/device_online.go`
- DO 文件：`internal/model/do/device_online.go`
- Entity 文件：`internal/model/entity/device_online.go`
- 业务逻辑：`internal/logic/websocket/websocket.go`

## 后续优化建议

1. **定时清理**：添加定时任务清理过期的离线记录
2. **统计报表**：基于此表生成设备在线统计报表
3. **告警功能**：当设备长时间离线时发送告警
4. **数据分析**：分析设备在线规律，优化资源分配

