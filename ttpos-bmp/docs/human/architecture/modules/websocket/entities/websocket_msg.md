### 实体：WebsocketMsg（WebSocket消息记录）

定位：存储 WebSocket 消息的详细记录，用于消息追踪、状态管理和问题排查。

数据结构
```go
type WebsocketMsg struct {
    Id           uint   `json:"id"`            // 主键ID
    Uuid         uint64 `json:"uuid"`          // 消息唯一标识
    CompanyUuid  uint64 `json:"company_uuid"`  // 公司UUID
    Uid          string `json:"uid"`           // 用户/设备标识
    Msg          string `json:"msg"`           // 消息内容（JSON格式）
    Type         string `json:"type"`          // 消息类型
    SourceClient string `json:"source_client"` // 来源客户端
    Status       int    `json:"status"`        // 消息状态
    IsOffline    int    `json:"is_offline"`    // 是否离线消息
    CreateTime   uint   `json:"create_time"`   // 创建时间
    UpdateTime   uint   `json:"update_time"`   // 更新时间
    DeleteTime   uint   `json:"delete_time"`   // 删除时间（软删除）
}
```

字段说明
- **Id**: 数据库主键，自增长
- **Uuid**: 消息的全局唯一标识，用于消息去重和追踪
- **CompanyUuid**: 所属公司的唯一标识，用于数据隔离
- **Uid**: 用户或设备的标识符，用于消息路由
- **Msg**: 消息的具体内容，通常为 JSON 格式的业务数据
- **Type**: 消息类型，如 "heartbeat"、"order"、"notification" 等
- **SourceClient**: 消息来源客户端类型，如 "pos"、"tablet"、"kitchen" 等
- **Status**: 消息状态，0-待发送，1-发送中，2-发送成功，3-发送失败
- **IsOffline**: 是否为离线消息，0-在线消息，1-离线消息
- **CreateTime**: 消息创建时间戳
- **UpdateTime**: 消息最后更新时间戳
- **DeleteTime**: 软删除时间戳，0表示未删除

消息状态枚举
```go
const (
    MessageStatusPending = 0  // 待发送
    MessageStatusSending = 1  // 发送中
    MessageStatusSuccess = 2  // 发送成功
    MessageStatusFailed  = 3  // 发送失败
)
```

消息类型枚举
```go
const (
    MessageTypeHeartbeat     = "heartbeat"     // 心跳消息
    MessageTypeOrder         = "order"         // 订单消息
    MessageTypeNotification  = "notification"  // 通知消息
    MessageTypeSystem        = "system"        // 系统消息
    MessageTypeBroadcast     = "broadcast"     // 广播消息
)
```

客户端类型枚举
```go
const (
    ClientTypePOS     = "pos"     // 收银机
    ClientTypeTablet  = "tablet"  // 平板
    ClientTypeKitchen = "kitchen" // 厨显
    ClientTypeH5      = "h5"      // H5页面
    ClientTypeMobile  = "mobile"  // 移动端
)
```

索引设计
- 主键索引：`PRIMARY KEY (id)`
- 唯一索引：`UNIQUE KEY idx_uuid (uuid)`
- 复合索引：`KEY idx_company_uid (company_uuid, uid)`
- 状态索引：`KEY idx_status (status)`
- 时间索引：`KEY idx_create_time (create_time)`

使用场景
1. **消息追踪**: 通过 uuid 追踪消息的完整生命周期
2. **状态查询**: 查询特定消息的发送状态
3. **离线消息**: 存储用户离线期间的消息，重连时重发
4. **统计分析**: 分析消息发送成功率、失败原因等
5. **问题排查**: 通过日志记录排查消息发送问题

注意事项
- 消息内容较大时，考虑使用压缩存储
- 定期清理过期的消息记录，避免数据量过大
- 敏感信息不应直接存储在消息内容中
- 离线消息需要设置合理的过期时间
