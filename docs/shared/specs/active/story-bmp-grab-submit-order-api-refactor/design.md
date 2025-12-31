# Grab 提交订单 API 重构 设计文档

> 本文档定义 Grab 提交订单 API 重构 的技术设计和实现方案。

## 📋 概述

本次重构主要涉及 Grab 提交订单的 API 结构优化，将空的 `SubmitOrderReq` 结构体修改为嵌入 `*grabfood.SubmitOrderRequest`，提升代码类型安全性和可维护性。同时优化 `shopUuid` 获取逻辑，优先从 `partnerMerchantID` 字段获取。

这是一个纯后端重构项目，不涉及数据库变更、UI 变更或第三方集成。主要在 Go BMP 模块中进行代码结构调整。

---

## 🎯 规范对齐

### Go BMP 规范 (go-bmp.mdc)

- **微服务架构**: 严格遵循 Controller → Logic → DAO 分层
- **自动生成代码**: 禁止修改 `dao/entity/do/` 和 `dao/do/` 目录
- **gRPC 服务**: 如需要注册到 Nacos
- **错误处理**: 使用 GoFrame 的错误处理机制
- **日志记录**: 使用 `g.Log()` 进行日志记录

### API 设计规范 (api.mdc)

- **URL 规范**: 保持现有 `/partner/orders` 路径
- **响应格式**: 统一使用 `{code, message, data}` 格式
- **数据格式**: `data` 字段必须是对象，不能是 null 或数组
- **分页信息**: 统一放在 `meta` 中（本功能不涉及分页）

### 数据库规范 (database.mdc)

- **无数据库变更**: 本次重构不涉及数据库表结构调整
- **现有表使用**: 继续使用 `ttpos_order`、`ttpos_order_item` 等现有表

---

## 🔄 代码复用分析

### 可复用的现有组件

- **现有 Logic 层**: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab_order/grab_order.go` - 现有的订单处理逻辑
- **Service 层**: `ttpos-bmp/app/ttpos-takeout/internal/service/grab_order.go` - 现有的服务接口
- **商户配置服务**: `ttpos-bmp/app/ttpos-takeout/internal/service/shop_provider_cfg.go` - 商户配置查询
- **Grab SDK**: `github.com/grab/grabfood-api-sdk-go` - 现有的 SDK 依赖

### 集成点

- **现有 API**: 不改变现有 API 接口签名
- **现有数据库表**: 继续使用现有订单相关表
- **现有微服务**: 继续调用现有的 Grab Order 服务

---

## 🏗️ 架构设计

### 分层设计原则

**Go BMP 三层架构**:

```
Controller 层 (HTTP/gRPC)
    ↓ 调用
Logic 层 (业务逻辑)
    ↓ 调用
Service 层 (服务接口)
    ↓ 调用
DAO 层 (数据访问)
    ↓ 访问
Database (MySQL)
```

**依赖规则**:

- ✅ 上层可依赖下层接口
- ❌ 禁止下层依赖上层
- ❌ 禁止跨层调用
- ✅ Logic 可依赖多个 Service 接口

### 架构图

```mermaid
graph TD
    A[ControllerV1.SubmitOrder] --> B[service.Grab().HandleSubmitOrder]
    B --> C[service.GrabOrder().HandleSubmitOrder]
    C --> D[logic.grab_order.HandleSubmitOrder]
    D --> E[dao.Order & dao.OrderItem]

    F[ShopProviderCfg Service] --> D
    G[Queue Service] --> D

    H[grabfood.SubmitOrderRequest] --> A
    H --> D
```

### 模块划分

#### Go BMP 模块

- **Controller 层**: `ttpos-bmp/app/ttpos-takeout/internal/controller/grab/` - HTTP 接口处理
- **Logic 层**: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab_order/` - 业务逻辑处理
- **Service 层**: `ttpos-bmp/app/ttpos-takeout/internal/service/` - 服务接口定义
- **DAO 层**: `ttpos-bmp/app/ttpos-takeout/internal/dao/` - 数据访问（自动生成）

---

## 🗄️ 数据库设计

### 数据表设计

**本次重构不涉及数据库变更，继续使用现有表结构**:

#### 表: ttpos_order

```sql
-- 现有表结构，保持不变
CREATE TABLE `ttpos_order` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `uuid` bigint unsigned NOT NULL DEFAULT '0',
  `shop_uuid` varchar(64) DEFAULT NULL,
  `provider_merchant_id` varchar(100) DEFAULT NULL,
  `partner_order_id` varchar(100) DEFAULT NULL,
  -- ... 其他字段
  `raw_data` longtext,
  `create_time` int NOT NULL DEFAULT '0',
  `update_time` int NOT NULL DEFAULT '0',
  `delete_time` int NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_uuid` (`uuid`),
  KEY `idx_partner_order_id` (`partner_order_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

#### 表: ttpos_order_item

```sql
-- 现有表结构，保持不变
CREATE TABLE `ttpos_order_item` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `order_uuid` bigint unsigned NOT NULL DEFAULT '0',
  -- ... 其他字段
  PRIMARY KEY (`id`),
  KEY `idx_order_uuid` (`order_uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

### 数据库迁移

**无迁移需求**: 本次重构为纯代码结构优化，不涉及数据库变更。

---

## 📊 数据模型

### Go Model

**本次重构不涉及数据模型变更**:

```go
// 继续使用现有模型
type Order struct {
    Id               uint64 `gorm:"column:id"`
    Uuid             uint64 `gorm:"column:uuid"`
    ShopUuid         string `gorm:"column:shop_uuid"`
    ProviderMerchantId string `gorm:"column:provider_merchant_id"`
    // ... 其他字段
}

func (*Order) TableName() string {
    return "ttpos_order"
}
```

---

## 🔌 API 设计

### RESTful API

#### API: SubmitOrder (现有接口，保持不变)

**请求**:

- **URL**: `/partner/orders`
- **Method**: `POST`
- **Headers**:
  ```json
  {
    "Authorization": "Bearer {token}",
    "Content-Type": "application/json"
  }
  ```
- **Body**: Grab SDK `SubmitOrderRequest` JSON 格式

**响应**:

```json
{
  "code": 1,
  "message": "success",
  "data": {}
}
```

**错误响应**:

```json
{
  "code": 0,
  "message": "错误信息",
  "data": {}
}
```

---

## 🧩 组件和接口

### Logic 层

#### 现有 Logic 修改

**文件**: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab_order/grab_order.go`

**修改内容**:
- `HandleSubmitOrder` 方法签名调整：`func (s *sGrabOrder) HandleSubmitOrder(ctx context.Context, req *grabfood.SubmitOrderRequest) error`
- `saveOrderFromSDK` 方法参数调整：`func (s *sGrabOrder) saveOrderFromSDK(ctx context.Context, req *grabfood.SubmitOrderRequest) (string, error)`
- 优化 `shopUuid` 获取逻辑，优先使用 `req.GetPartnerMerchantID()`

### Controller 层

#### 现有 Controller 修改

**文件**: `ttpos-bmp/app/ttpos-takeout/internal/controller/grab/grab_v1_submit_order.go`

**修改内容**:
- 修改 `SubmitOrderReq` 结构体，嵌入 `*grabfood.SubmitOrderRequest`
- 调整 Controller 方法，直接传递类型化对象给 Service 层

### Service 层

#### 现有 Service 修改

**文件**: `ttpos-bmp/app/ttpos-takeout/internal/service/grab_order.go`

**修改内容**:
- `HandleSubmitOrder` 方法签名调整：`HandleSubmitOrder(ctx context.Context, req *grabfood.SubmitOrderRequest) error`

---

## ⚡ 缓存设计

**本次重构不涉及缓存变更**，继续使用现有缓存策略（如有）。

---

## 🚨 错误处理

### 错误场景

#### 场景 1: SDK 类型解析失败

- **处理方式**: 返回详细错误信息，记录日志
- **用户影响**: API 返回错误响应
- **代码示例**:
  ```go
  if err := json.Unmarshal(body, &req); err != nil {
      g.Log().Errorf(ctx, "解析提交订单请求失败: %v", err)
      return fmt.Errorf("解析请求失败: %w", err)
  }
  ```

#### 场景 2: 商户配置查询失败

- **处理方式**: 记录警告日志，继续处理（shopUuid 为空）
- **用户影响**: 订单保存成功，但商户关联信息缺失

---

## 🔒 安全设计

### 身份验证

- **Webhook 验证**: 已由中间件完成签名验证
- **无额外身份验证**: 依赖现有安全机制

### 数据安全

- **数据加密**: 敏感数据在数据库层面加密存储
- **SQL 注入防护**: 使用参数化查询
- **XSS 防护**: 不涉及前端输入

---

## 🧪 测试策略

### 单元测试

**目标覆盖率**:

- `ttpos-bmp/app/ttpos-takeout/internal/logic/grab_order/`: ≥ 70%
- **Payment/Order 相关模块**: 100%（高风险）

**测试内容**:

- Logic 层业务逻辑正确性
- 类型转换和数据映射
- 错误处理场景
- 商户配置查询逻辑

**示例**:

```go
func TestGrabOrder_HandleSubmitOrder(t *testing.T) {
    // 测试正常订单提交
    // 测试商户配置查询失败的场景
    // 测试 shopUuid 优先级逻辑
}
```

### API 测试

**测试内容**:

- API 接口调用
- 参数验证（通过 SDK 类型验证）
- 响应格式
- 错误处理

### 集成测试

**测试流程**:

- 端到端订单提交流程
- 数据库数据一致性
- MQ 消息发送

---

## 📈 性能优化

### 优化策略

1. **代码优化**:
   - 减少重复的 JSON 解析
   - 优化类型转换逻辑

2. **查询优化**:
   - 商户配置查询保持现有索引使用

3. **内存优化**:
   - 直接使用 SDK 类型，避免额外内存分配

### 性能指标

- 本地响应时间: < 200ms（重构不影响性能）
- 数据库查询: < 50ms（保持现有水平）
- 内存使用: 减少 JSON 解析开销

---

## 🌐 浏览器兼容性

**本次重构为后端 API，不涉及前端兼容性**

---

## 📚 实现清单

### Phase 1: API 层重构

- [x] 修改 `SubmitOrderReq` 结构体，嵌入 `*grabfood.SubmitOrderRequest`
- [x] 调整 Controller 方法，直接传递类型化对象

### Phase 2: Logic 层调整

- [x] 修改 `HandleSubmitOrder` 方法签名
- [x] 修改 `saveOrderFromSDK` 方法参数
- [x] 优化 `shopUuid` 获取逻辑

### Phase 3: Service 层调整

- [x] 修改 Service 接口方法签名

### Phase 4: 测试验证

- [x] 单元测试覆盖
- [x] 集成测试验证
- [x] 向后兼容性测试

**详细任务**: 参见 `tasks.md`

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/2025-12/2025-12-19.md`
- 在技术方案评审、关键架构决策或踩坑总结后，立即补充 Episode 并在设计文档尾部互链。

---

**版本**: v1.0.0
**创建日期**: 2025-12-19
**作者**: rikugun
**审核者**: {审核者}
