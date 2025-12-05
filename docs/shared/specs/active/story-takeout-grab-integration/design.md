# GrabFood 外卖平台对接 (v1.1.3) 设计文档

> 本文档定义 GrabFood 对接功能的技术设计和实现方案。

## 📋 概述

本功能旨在通过 `ttpos-takeout` 模块实现与 GrabFood 外卖平台 (API v1.1.3) 的对接。
核心架构采用 **Webhook + Persistence + MQ** 模式：
1. `ttpos-takeout` 接收 GrabFood Webhook 回调。
2. 验证签名后，**将订单数据持久化到 `ttpos-takeout` 数据库**。
3. 将标准化消息投递到 RocketMQ，供下游系统（本设计不包含消费者实现）使用。

---

## 🎯 规范对齐

### Go BMP 规范 (go-bmp.mdc)

- **Takeout 网关端**:
    - Controller 负责 HTTP 接口处理。
    - Logic 层处理签名、**数据入库**、MQ 发送。
    - Model 层新增 `TakeoutOrder` 等实体。

---

## 🏗️ 架构设计

### 架构图

```mermaid
graph TD
    Grab[GrabFood Platform] -- Webhook --> BMP[ttpos-takeout]
    BMP -- Validate --> DB[(Takeout DB)]
    DB -- Persist --> BMP
    BMP -- Convert --> MQ[RocketMQ]
    MQ -.-> ERP[ttpos-erp (Future)]
    
    subgraph "ttpos-takeout"
        Controller[Webhook Controller]
        Logic[Grab Logic]
        DAO[Takeout DAO]
        Producer[MQ Producer]
    end
```

---

## 🗄️ 数据库设计

### 数据表设计

新增表用于存储第三方平台原始订单数据，作为 ERP 订单的 Source of Truth 备份。

#### 表 1: `takeout_order` (外送订单主表)

```sql
CREATE TABLE IF NOT EXISTS `takeout_order` (
    `id` bigint(20) NOT NULL AUTO_INCREMENT,
    `uuid` varchar(100) NOT NULL COMMENT '系统唯一ID',
    `merchant_id` varchar(100) NOT NULL COMMENT '商户ID',
    `partner_order_id` varchar(100) NOT NULL COMMENT '平台订单号',
    `short_order_number` varchar(50) DEFAULT NULL COMMENT '短单号',
    `provider_name` varchar(50) NOT NULL COMMENT '渠道: grab, foodpanda',
    `order_time` datetime DEFAULT NULL COMMENT '下单时间',
    `order_status` varchar(50) DEFAULT NULL COMMENT '订单状态',
    `currency` varchar(10) DEFAULT 'THB',
    `total_amount` decimal(14,2) DEFAULT 0.00 COMMENT '总金额',
    `merchant_tax` decimal(14,2) DEFAULT 0.00 COMMENT '商户税费',
    `takeout_tax` decimal(14,2) DEFAULT 0.00 COMMENT '平台税费',
    `discount_amount` decimal(14,2) DEFAULT 0.00 COMMENT '折扣金额',
    `payment_type` varchar(50) DEFAULT NULL COMMENT '支付方式: CASH, CARD',
    `customer_name` varchar(100) DEFAULT NULL,
    `customer_phone` varchar(50) DEFAULT NULL,
    `note` text DEFAULT NULL COMMENT '备注',
    `raw_data` json DEFAULT NULL COMMENT '原始JSON数据',
    `created_at` datetime DEFAULT NULL,
    `updated_at` datetime DEFAULT NULL,
    `deleted_at` datetime DEFAULT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_partner_order` (`provider_name`, `partner_order_id`),
    UNIQUE KEY `uk_uuid` (`uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='外送订单主表';
```

#### 表 2: `takeout_order_item` (外送订单明细表)

```sql
CREATE TABLE IF NOT EXISTS `takeout_order_item` (
    `id` bigint(20) NOT NULL AUTO_INCREMENT,
    `order_uuid` varchar(100) NOT NULL COMMENT '关联订单UUID',
    `partner_item_id` varchar(100) DEFAULT NULL COMMENT '平台商品ID',
    `item_name` varchar(200) NOT NULL,
    `quantity` int(11) NOT NULL DEFAULT 1,
    `price` decimal(14,2) NOT NULL DEFAULT 0.00 COMMENT '单价',
    `total_price` decimal(14,2) NOT NULL DEFAULT 0.00 COMMENT '总价',
    `specifications` text DEFAULT NULL COMMENT '规格描述(JSON)',
    `note` varchar(500) DEFAULT NULL,
    `created_at` datetime DEFAULT NULL,
    PRIMARY KEY (`id`),
    KEY `idx_order_uuid` (`order_uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='外送订单明细表';
```

#### 表 3: `takeout_menu_log` (菜单同步记录表)

```sql
CREATE TABLE IF NOT EXISTS `takeout_menu_log` (
    `id` bigint(20) NOT NULL AUTO_INCREMENT,
    `merchant_id` varchar(100) NOT NULL COMMENT '商户ID',
    `provider_name` varchar(50) NOT NULL COMMENT '渠道: grab',
    `request_id` varchar(100) DEFAULT NULL COMMENT '请求ID',
    `status` varchar(20) DEFAULT NULL COMMENT '同步状态: SUCCESS, QUEUED, FAIL',
    `menu_snapshot` json DEFAULT NULL COMMENT '菜单快照(JSON)',
    `error_msg` text DEFAULT NULL,
    `created_at` datetime DEFAULT NULL,
    PRIMARY KEY (`id`),
    KEY `idx_merchant` (`merchant_id`, `provider_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='菜单同步记录表';
```

#### 表 4: `takeout_order_status_log` (订单状态变更日志表) (New)

```sql
CREATE TABLE IF NOT EXISTS `takeout_order_status_log` (
    `id` bigint(20) NOT NULL AUTO_INCREMENT,
    `order_uuid` varchar(100) NOT NULL COMMENT '关联订单UUID',
    `provider_name` varchar(50) NOT NULL COMMENT '渠道: grab',
    `status_before` varchar(50) DEFAULT NULL COMMENT '变更前状态',
    `status_after` varchar(50) NOT NULL COMMENT '变更后状态',
    `change_source` varchar(50) DEFAULT 'WEBHOOK' COMMENT '变更来源: WEBHOOK, API, SYSTEM',
    `remark` text DEFAULT NULL COMMENT '备注或原因',
    `created_at` datetime DEFAULT NULL,
    PRIMARY KEY (`id`),
    KEY `idx_order_uuid` (`order_uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='订单状态变更日志表';
```

---

## 🔌 API 设计

### Webhook API (ttpos-takeout)

#### API 1: 接收订单 (Submit Order)

- **URL**: `/api/v1/callback/grab/orders`
- **Logic**:
    1. 校验 HMAC 签名。
    2. 解析 JSON。
    3. 开启事务：
        - 插入 `takeout_order`。
        - 插入 `takeout_order_item`。
    4. 提交事务。
    5. 发送 RocketMQ 消息。
    6. 返回 200 OK。

---

## 🧩 组件和接口

### Service 层 (ttpos-takeout)

#### Grab Client (API Endpoints 实现)

封装 Grab API 调用逻辑，提供给 Logic 层使用。

```go
// ttpos-bmp/app/ttpos-takeout/internal/service/grab/client.go

type GrabClient interface {
    // 菜单管理
    UpdateMenu(ctx context.Context, merchantID string, menu *dto.GrabMenu) error
    
    // 门店状态
    UpdateStoreStatus(ctx context.Context, merchantID string, status dto.StoreStatus) error
    
    // 订单操作
    AcceptOrder(ctx context.Context, orderID string) error
    RejectOrder(ctx context.Context, orderID string, reason string) error
}
```

#### Grab Logic

```go
// ttpos-bmp/app/ttpos-takeout/internal/logic/grab/grab.go

func (s *sGrab) HandleOrder(ctx context.Context, payload []byte) error {
    // 1. Verify Signature
    // 2. Parse DTO
    // 3. Save to DB (Transaction)
    // 4. Send MQ (Retry on failure? Or just log error as order is already saved)
}

func (s *sGrab) SyncMenu(ctx context.Context, merchantID string) error {
    // 1. Convert POS Menu -> Grab Menu DTO
    // 2. Save snapshot to takeout_menu_log (QUEUED)
    // 3. Call GrabClient.UpdateMenu
    // 4. Update log status (SUCCESS/FAIL)
}

// HandleOrderStatusUpdate 处理状态变更回调
func (s *sGrab) HandleOrderStatusUpdate(ctx context.Context, payload []byte) error {
     // 1. Verify Signature
     // 2. Parse DTO
     // 3. Save to takeout_order_status_log
     // 4. Update takeout_order.order_status
     // 5. Send MQ
}
```

---

## 📚 实现清单

### Phase 1: 数据库与模型
- [ ] 创建 `takeout_order`, `takeout_order_item`, `takeout_menu_log`, `takeout_order_status_log` 迁移脚本。
- [ ] 生成 Model/DAO。

### Phase 2: 核心逻辑 (ttpos-takeout)
- [ ] 实现签名验证。
- [ ] 实现 Webhook Controller。
- [ ] 实现数据入库逻辑。
- [ ] 实现 MQ 发送逻辑。

### Phase 3: API Client 与菜单
- [ ] 实现 `GrabClient` (API endpoints 对接)。
- [ ] 实现菜单转换与保存逻辑。

**详细任务**: 参见 `tasks.md`

---

**版本**: v1.3.0
**创建日期**: 2025-12-04
