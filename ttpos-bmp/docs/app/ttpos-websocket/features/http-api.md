# WebSocket HTTP API 文档

## 📖 概述

为了方便从旧系统迁移到新的 gRPC 架构，ttpos-websocket 保留了 HTTP 推送接口，与原有的 `websocket/api/api.go` 完全兼容。

## 🔗 接口地址

```
POST http://localhost:14051/ws/push
```

## 📝 请求参数

### 请求头

```
Content-Type: application/json
```

### 请求体

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| company_uuid | uint64 | ✅ | 公司UUID |
| source_client | string | ❌ | 来源客户端（shop/cashier/tablet/kitchen/assistant/H5/*），默认 * |
| device_id | string | ❌ | 设备ID，*表示所有设备，默认 * |
| not_device_id | string | ❌ | 排除的设备ID |
| staff_uuid | uint64 | ❌ | 员工UUID，0表示推送给所有员工 |
| not_staff_uuid | uint64 | ❌ | 排除的员工UUID |
| message_type | string | ✅ | 消息类型 |
| message_key | string | ❌ | 消息键，用于防抖去重 |
| data | object/string | ✅ | 消息数据（可以是对象或字符串） |

### 响应格式

```json
{
  "code": 1,
  "message": "success"
}
```

## 💻 使用示例

### 示例 1：订单更新推送（使用防抖）

```bash
curl -X POST http://localhost:14051/ws/push \
  -H "Content-Type: application/json" \
  -d '{
    "company_uuid": 1,
    "message_type": "update_order",
    "message_key": "order_12345_update",
    "data": {
      "order_id": 12345,
      "status": "paid",
      "amount": 100.00
    }
  }'
```

**响应：**
```json
{
  "code": 1,
  "message": "success"
}
```

### 示例 2：顾客呼叫（不使用防抖）

```bash
curl -X POST http://localhost:14051/ws/push \
  -H "Content-Type: application/json" \
  -d '{
    "company_uuid": 1,
    "source_client": "cashier",
    "message_type": "customer_call",
    "data": {
      "table": "A01",
      "time": "2025-01-15 10:30:00"
    }
  }'
```

### 示例 3：推送给特定设备

```bash
curl -X POST http://localhost:14051/ws/push \
  -H "Content-Type: application/json" \
  -d '{
    "company_uuid": 1,
    "source_client": "shop",
    "device_id": "device_001",
    "staff_uuid": 123,
    "message_type": "notification",
    "message_key": "notify_123_device_001",
    "data": {
      "title": "新订单",
      "content": "您有一个新的订单"
    }
  }'
```

### 示例 4：排除特定设备

```bash
curl -X POST http://localhost:14051/ws/push \
  -H "Content-Type: application/json" \
  -d '{
    "company_uuid": 1,
    "source_client": "*",
    "device_id": "*",
    "not_device_id": "device_002",
    "message_type": "system_update",
    "data": {
      "version": "2.0.0",
      "update_time": "2025-01-15 12:00:00"
    }
  }'
```

## 🔧 JavaScript/TypeScript 示例

### 使用 Fetch API

```javascript
async function pushMessage(params) {
  try {
    const response = await fetch('http://localhost:14051/ws/push', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(params),
    });
    
    const result = await response.json();
    console.log('推送结果:', result);
    return result;
  } catch (error) {
    console.error('推送失败:', error);
    throw error;
  }
}

// 使用示例
pushMessage({
  company_uuid: 1,
  message_type: 'update_order',
  message_key: 'order_12345_update',
  data: {
    order_id: 12345,
    status: 'paid',
  },
});
```

### 使用 Axios

```javascript
import axios from 'axios';

async function pushMessage(params) {
  try {
    const response = await axios.post(
      'http://localhost:14051/ws/push',
      params
    );
    console.log('推送结果:', response.data);
    return response.data;
  } catch (error) {
    console.error('推送失败:', error);
    throw error;
  }
}

// 使用示例
pushMessage({
  company_uuid: 1,
  message_type: 'customer_call',
  data: {
    table: 'A01',
  },
});
```

## 🐘 PHP 示例

```php
<?php
function pushMessage($params) {
    $url = 'http://localhost:14051/ws/push';
    
    $ch = curl_init($url);
    curl_setopt($ch, CURLOPT_POST, 1);
    curl_setopt($ch, CURLOPT_POSTFIELDS, json_encode($params));
    curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
    curl_setopt($ch, CURLOPT_HTTPHEADER, [
        'Content-Type: application/json',
    ]);
    
    $response = curl_exec($ch);
    $httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);
    curl_close($ch);
    
    if ($httpCode === 200) {
        return json_decode($response, true);
    } else {
        throw new Exception("推送失败: HTTP $httpCode");
    }
}

// 使用示例
$result = pushMessage([
    'company_uuid' => 1,
    'message_type' => 'update_order',
    'message_key' => 'order_12345_update',
    'data' => [
        'order_id' => 12345,
        'status' => 'paid',
    ],
]);

print_r($result);
?>
```

## 🐍 Python 示例

```python
import requests
import json

def push_message(params):
    url = 'http://localhost:14051/ws/push'
    headers = {'Content-Type': 'application/json'}
    
    try:
        response = requests.post(url, json=params, headers=headers)
        response.raise_for_status()
        return response.json()
    except requests.exceptions.RequestException as e:
        print(f'推送失败: {e}')
        raise

# 使用示例
result = push_message({
    'company_uuid': 1,
    'message_type': 'update_order',
    'message_key': 'order_12345_update',
    'data': {
        'order_id': 12345,
        'status': 'paid',
    }
})

print('推送结果:', result)
```

## 🔄 与旧 API 的兼容性

### 完全兼容

新的 HTTP API 与旧的 `websocket/api/api.go` 完全兼容：

1. ✅ 相同的请求路径：`/ws/push`
2. ✅ 相同的请求参数格式
3. ✅ 相同的响应格式
4. ✅ 相同的防抖机制（900ms）
5. ✅ 相同的限流机制（500并发）

### 迁移步骤

1. **无需修改客户端代码**
   - 只需将请求地址从旧服务改为新服务
   - 例如：`http://old-host:8099/ws/push` → `http://new-host:14051/ws/push`

2. **逐步迁移**
   - 可以先将部分流量切换到新服务
   - 验证功能正常后再完全切换

3. **监控对比**
   - 对比新旧服务的推送成功率
   - 对比响应时间和性能指标

## 🎯 防抖机制说明

### 工作原理

与旧 API 完全相同：

1. 设置 `message_key` 后启用防抖
2. 900ms 内相同 `message_key` 的请求会被合并
3. 只执行最后一次推送
4. 2秒内超过10次推送会自动执行

### 示例

```bash
# 快速发送3次请求（间隔100ms）
curl -X POST http://localhost:14051/ws/push \
  -H "Content-Type: application/json" \
  -d '{"company_uuid": 1, "message_type": "test", "message_key": "test_key", "data": {"version": 1}}'

sleep 0.1

curl -X POST http://localhost:14051/ws/push \
  -H "Content-Type: application/json" \
  -d '{"company_uuid": 1, "message_type": "test", "message_key": "test_key", "data": {"version": 2}}'

sleep 0.1

curl -X POST http://localhost:14051/ws/push \
  -H "Content-Type: application/json" \
  -d '{"company_uuid": 1, "message_type": "test", "message_key": "test_key", "data": {"version": 3}}'

# 最终只会推送 version: 3
```

## ⚙️ 配置说明

### 服务端口

在 `manifest/config/config.yaml` 中配置：

```yaml
server:
  address: ":14051"  # HTTP服务端口
```

### 限流配置

在代码中配置（`internal/controller/http/push.go`）：

```go
// 最大并发处理的消息数
maxConcurrentMessages = 500
```

### 防抖延迟

在代码中配置（`internal/controller/http/push.go`）：

```go
// 等待900毫秒
time.Sleep(900 * time.Millisecond)
```

## 📊 性能对比

| 指标 | 旧 API | 新 HTTP API | 新 gRPC API |
|------|--------|-------------|-------------|
| 协议 | HTTP | HTTP | gRPC |
| 响应时间 | ~5ms | ~5ms | ~2ms |
| 并发支持 | 500 | 500 | 无限制 |
| 防抖延迟 | 900ms | 900ms | 900ms |
| 跨实例 | ✅ Redis | ✅ Redis | ✅ Redis |

## 🔍 调试技巧

### 查看日志

```bash
# 查看HTTP推送日志
tail -f log/websocket/*.log | grep "HTTP推送"

# 查看防抖日志
tail -f log/websocket/*.log | grep "HTTP防抖"
```

### 测试工具

使用 Postman 或 Insomnia 测试：

1. 创建新的 POST 请求
2. URL: `http://localhost:14051/ws/push`
3. Headers: `Content-Type: application/json`
4. Body: 选择 raw JSON
5. 发送请求

### Redis 监控

```bash
# 查看防抖键
redis-cli keys "*_count"

# 监控 Redis 命令
redis-cli monitor | grep push_client
```

## ❓ 常见问题

### Q1: HTTP API 和 gRPC API 有什么区别？

A: 
- **HTTP API**: 兼容旧系统，方便迁移，性能略低
- **gRPC API**: 新架构推荐，性能更好，类型安全

### Q2: 如何选择使用哪个接口？

A:
- **过渡期**: 使用 HTTP API，无需修改现有代码
- **新功能**: 使用 gRPC API，获得更好的性能
- **长期**: 逐步迁移到 gRPC API

### Q3: HTTP API 会被废弃吗？

A: 短期内不会废弃，会长期保持兼容。但建议新项目使用 gRPC API。

### Q4: 性能差异大吗？

A: HTTP API 和 gRPC API 在防抖场景下性能差异不大（都需要等待900ms），但 gRPC 在高并发场景下性能更好。

## 📚 相关文档

- [USAGE_EXAMPLE.md](./USAGE_EXAMPLE.md) - gRPC API 使用示例
- [DEBOUNCE_MIGRATION.md](./DEBOUNCE_MIGRATION.md) - 防抖功能迁移文档
- [README.MD](./README.MD) - 项目文档

## 🎉 总结

HTTP API 提供了与旧系统完全兼容的接口，方便平滑迁移：

1. ✅ **零改动迁移**: 只需修改服务地址
2. ✅ **完全兼容**: 请求和响应格式完全相同
3. ✅ **功能一致**: 防抖、限流机制完全相同
4. ✅ **易于测试**: 使用 curl、Postman 等工具即可测试
5. ✅ **多语言支持**: 支持任何能发送 HTTP 请求的语言

建议在过渡期使用 HTTP API，待系统稳定后逐步迁移到 gRPC API 以获得更好的性能。

