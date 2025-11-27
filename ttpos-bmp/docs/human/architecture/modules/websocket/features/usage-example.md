# WebSocket 防抖功能使用示例

## 📖 功能说明

WebSocket 消息推送现在支持防抖功能，通过设置 `message_key` 参数，可以在短时间内（900ms）合并相同的推送请求，只执行最后一次推送。

## 🎯 使用场景

### 适合使用防抖的场景

1. **订单更新推送**
   - 订单状态频繁变化时，避免重复推送
   - 例如：订单从"待支付"→"支付中"→"已支付"，只推送最终状态

2. **库存变化通知**
   - 商品库存频繁变动时，合并推送
   - 减少客户端刷新次数

3. **配置更新通知**
   - 管理员频繁修改配置时，只推送最终配置

4. **打印任务推送**
   - 避免重复打印相同内容

### 不适合使用防抖的场景

1. **顾客呼叫**
   - 每次呼叫都需要立即响应
   - 不应该合并

2. **紧急通知**
   - 需要实时推送的消息

3. **一次性事件**
   - 不会重复发生的事件

## 💻 代码示例

### 示例 1：订单更新推送（使用防抖）

```go
package main

import (
    "context"
    "google.golang.org/grpc"
    v1 "ttpos-bmp/app/ttpos-websocket/api/websocket"
)

func pushOrderUpdate(orderID uint64, orderData string) error {
    // 连接 WebSocket gRPC 服务
    conn, err := grpc.Dial("localhost:14052", grpc.WithInsecure())
    if err != nil {
        return err
    }
    defer conn.Close()
    
    client := v1.NewWebSocketClient(conn)
    
    // 构建请求
    req := &v1.PushMessageReq{
        CompanyUuid:  1,                                    // 公司UUID
        MessageType:  "update_order",                       // 消息类型
        MessageKey:   fmt.Sprintf("order_%d_update", orderID), // 设置 MessageKey 启用防抖
        Data:         orderData,                            // 订单数据（JSON字符串）
        SourceClient: "*",                                  // 推送给所有客户端
        DeviceId:     "*",                                  // 推送给所有设备
    }
    
    // 调用推送接口
    resp, err := client.PushMessage(context.Background(), req)
    if err != nil {
        return err
    }
    
    fmt.Printf("推送结果: %s\n", resp.Message)
    return nil
}

// 使用示例
func main() {
    orderID := uint64(12345)
    orderData := `{"order_id": 12345, "status": "paid", "amount": 100.00}`
    
    // 即使在 900ms 内多次调用，也只会推送最后一次
    pushOrderUpdate(orderID, orderData)
    time.Sleep(100 * time.Millisecond)
    pushOrderUpdate(orderID, orderData)
    time.Sleep(100 * time.Millisecond)
    pushOrderUpdate(orderID, orderData)
    // 最终只会推送一次
}
```

### 示例 2：顾客呼叫（不使用防抖）

```go
func pushCustomerCall(tableNumber string) error {
    conn, err := grpc.Dial("localhost:14052", grpc.WithInsecure())
    if err != nil {
        return err
    }
    defer conn.Close()
    
    client := v1.NewWebSocketClient(conn)
    
    req := &v1.PushMessageReq{
        CompanyUuid:  1,
        MessageType:  "customer_call",
        // 不设置 MessageKey，每次都会立即推送
        Data:         fmt.Sprintf(`{"table": "%s"}`, tableNumber),
        SourceClient: "cashier",  // 只推送给收银端
    }
    
    resp, err := client.PushMessage(context.Background(), req)
    if err != nil {
        return err
    }
    
    fmt.Printf("推送结果: %s, 推送数量: %d\n", resp.Message, resp.PushCount)
    return nil
}
```

### 示例 3：批量推送不同设备

```go
func pushToSpecificDevice(companyUUID, staffUUID uint64, deviceID, message string) error {
    conn, err := grpc.Dial("localhost:14052", grpc.WithInsecure())
    if err != nil {
        return err
    }
    defer conn.Close()
    
    client := v1.NewWebSocketClient(conn)
    
    req := &v1.PushMessageReq{
        CompanyUuid:  companyUUID,
        StaffUuid:    staffUUID,      // 指定员工
        MessageType:  "notification",
        MessageKey:   fmt.Sprintf("notify_%d_%s", staffUUID, deviceID), // 按设备防抖
        Data:         message,
        SourceClient: "shop",          // 指定来源
        DeviceId:     deviceID,        // 指定设备
    }
    
    resp, err := client.PushMessage(context.Background(), req)
    if err != nil {
        return err
    }
    
    return nil
}
```

### 示例 4：排除特定设备或员工

```go
func pushExcludeDevice(companyUUID uint64, excludeDeviceID string, message string) error {
    conn, err := grpc.Dial("localhost:14052", grpc.WithInsecure())
    if err != nil {
        return err
    }
    defer conn.Close()
    
    client := v1.NewWebSocketClient(conn)
    
    req := &v1.PushMessageReq{
        CompanyUuid:  companyUUID,
        MessageType:  "system_update",
        NotDeviceId:  excludeDeviceID,  // 排除特定设备
        Data:         message,
        SourceClient: "*",
        DeviceId:     "*",
    }
    
    resp, err := client.PushMessage(context.Background(), req)
    return err
}
```

## 🔑 MessageKey 命名规范

推荐使用以下格式命名 MessageKey：

```
{业务类型}_{业务ID}_{操作类型}
```

### 示例

- `order_12345_update` - 订单 12345 的更新
- `product_678_stock` - 商品 678 的库存变化
- `config_system_update` - 系统配置更新
- `table_A01_status` - 桌台 A01 的状态变化
- `printer_kitchen_task` - 厨房打印机任务

## ⚙️ 参数说明

### PushMessageReq 参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| company_uuid | uint64 | ✅ | 公司UUID |
| staff_uuid | uint64 | ❌ | 员工UUID，0表示推送给所有员工 |
| not_staff_uuid | uint64 | ❌ | 排除的员工UUID |
| source_client | string | ❌ | 来源客户端（shop/cashier/tablet/kitchen/assistant/H5/*） |
| device_id | string | ❌ | 设备ID，*表示所有设备 |
| not_device_id | string | ❌ | 排除的设备ID |
| message_type | string | ✅ | 消息类型 |
| message_key | string | ❌ | 消息键，设置后启用防抖 |
| data | string | ✅ | 消息数据（JSON格式字符串） |

### 客户端类型说明

- `shop` - 商家端
- `cashier` - 收银端
- `tablet` - 平板点餐
- `kitchen` - 厨房显示屏
- `assistant` - 服务员端
- `H5` - H5页面
- `*` - 所有客户端

## 📊 防抖机制说明

### 工作原理

```
请求1 → 设置UUID1 → 等待900ms → 检查UUID → 推送
         ↓
请求2 → 更新UUID2 → 等待900ms → 检查UUID → 推送
         ↓
请求3 → 更新UUID3 → 等待900ms → 检查UUID → 推送 ✅
                                     ↑
                            请求1和2被取消 ❌
```

### 时间线示例

```
0ms:    请求1到达，设置UUID1，开始等待
100ms:  请求2到达，更新UUID2，开始等待
200ms:  请求3到达，更新UUID3，开始等待
900ms:  请求1检查UUID，发现已变成UUID3，取消推送
1000ms: 请求2检查UUID，发现已变成UUID3，取消推送
1100ms: 请求3检查UUID，UUID3未变，执行推送 ✅
```

### 特殊情况

1. **高频推送（>10次/2秒）**
   - 计数器超过10次后，自动推送，不再等待
   - 防止消息积压

2. **Redis 连接失败**
   - 降级为直接推送
   - 记录错误日志

## 🧪 测试建议

### 单元测试

```go
func TestDebounce(t *testing.T) {
    // 测试防抖功能
    messageKey := "test_order_123"
    
    // 快速发送3次请求
    for i := 0; i < 3; i++ {
        pushOrderUpdate(123, fmt.Sprintf(`{"version": %d}`, i))
        time.Sleep(100 * time.Millisecond)
    }
    
    // 等待防抖完成
    time.Sleep(1 * time.Second)
    
    // 验证只推送了一次
    // 检查 WebSocket 客户端接收到的消息数量
}
```

### 压力测试

```go
func BenchmarkDebounce(b *testing.B) {
    for i := 0; i < b.N; i++ {
        pushOrderUpdate(uint64(i), `{"test": true}`)
    }
}
```

## 📝 最佳实践

1. **合理设置 MessageKey**
   - 确保相同业务使用相同的 key
   - 不同业务使用不同的 key

2. **监控推送效果**
   - 记录推送成功率
   - 监控防抖取消次数

3. **处理推送失败**
   - 实现重试机制
   - 记录失败日志

4. **客户端处理**
   - 客户端应该能处理延迟推送
   - 使用消息ID去重

## 🔍 调试技巧

### 查看日志

```bash
# 查看推送日志
tail -f log/websocket/*.log | grep "推送消息"

# 查看防抖日志
tail -f log/websocket/*.log | grep "防抖"
```

### Redis 调试

```bash
# 连接 Redis
redis-cli

# 查看所有防抖键
keys *_count

# 查看特定键的值
get order_123_update

# 查看键的过期时间
ttl order_123_update
```

## ❓ 常见问题

### Q1: 为什么我的消息没有立即推送？

A: 如果设置了 `message_key`，消息会延迟 900ms 推送。如果需要立即推送，不要设置 `message_key`。

### Q2: 如何确认防抖是否生效？

A: 查看日志中的 "防抖取消推送" 和 "防抖推送消息" 记录。

### Q3: 防抖会影响性能吗？

A: 不会。防抖逻辑在 goroutine 中异步执行，不阻塞主流程。

### Q4: 多个服务实例如何协调防抖？

A: 通过 Redis 共享防抖状态，确保跨实例的防抖效果。

## 📚 相关文档

- [DEBOUNCE_MIGRATION.md](./DEBOUNCE_MIGRATION.md) - 迁移文档
- [README.MD](./README.MD) - 项目文档
- [GoFrame Redis 文档](https://goframe.org/docs/components/cache-redis)

