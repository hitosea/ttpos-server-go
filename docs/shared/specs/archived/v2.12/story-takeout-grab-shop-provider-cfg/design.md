# Grab 店铺集成状态落库与旅程关联设计文档

> 本文档定义 Grab 店铺集成状态落库、状态变更消息及 gRPC 查询的技术设计与实现方案。

## 📋 概述

围绕 Grab integrationStatus 及自助旅程创建，新增持久化表 `shop_provider_cfg`，实现 webhook/upsert、消息推送 `takeout_store_integration_state`，并提供 gRPC 查询接口（按 shop_uuid / provider_name=grab）。

---

## 🎯 规范对齐

### Go BMP 规范 (ttpos-bmp/.cursor/rules/go-rules.mdc)
- 禁止直接修改 dao/entity/do，生成代码通过 `make dao`。
- gRPC 服务需注册到 Nacos，遵循统一错误码/鉴权策略。

### API 设计规范 (api.mdc)
- gRPC 响应字段保持对象结构，code/message/data。

### 数据库规范 (database.mdc)
- 必含 `id, uuid, created_at, updated_at, deleted_at`，int 秒级时间。
- 字段 snake_case；索引覆盖 `shop_uuid + provider_name`。

---

## 🔄 代码复用分析
- 现有 Grab webhook 处理：`ttpos-bmp/app/ttpos-takeout/internal/controller/grab/` 及 `internal/logic/grab/`（复用签名校验、状态枚举）。
- 自助旅程创建流程：`internal/logic/grab/self_serve_journey.go`。
- 消息推送：沿用 BMP 内部 MQ 生产封装（参考 `ttpos-bmp/manifest/topics.txt` 和现有 producer 封装）。

---

## 🏗️ 架构设计

### 流程（Webhook）
1) Grab 推送 integrationStatus → Controller 校验 → Logic 做状态映射。
2) Logic 调用 Service Upsert `shop_provider_cfg`（按 shop_uuid + provider_name=grab）。
3) 更新完成后发布消息到 `takeout_store_integration_state`，携带状态与 merchant_id。

### 流程（CreateSelfServeJourney）
1) 旅程创建成功后，调用同一 Service Upsert（状态 SYNCING/ACTIVE）。
2) 失败时不写成功状态，可落 FAILED 供追踪。

### 流程（gRPC 查询）
1) gRPC Controller 接收 `shop_uuid`（可选 provider_name，默认 grab）。
2) Service 只读查询表，返回状态/merchant_id/更新时间；无记录返回 not found。

---

## 🗄️ 数据库设计

### 表：shop_provider_cfg
```sql
CREATE TABLE IF NOT EXISTS `shop_provider_cfg` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `uuid` bigint unsigned NOT NULL DEFAULT 0 COMMENT '唯一标识',
  `shop_uuid` bigint unsigned NOT NULL DEFAULT 0 COMMENT '门店UUID',
  `provider_name` varchar(32) NOT NULL DEFAULT '' COMMENT '第三方名称，如 grab',
  `provider_merchant_id` varchar(128) NOT NULL DEFAULT '' COMMENT '第三方商户ID',
  `provider_shop_status` enum('INACTIVE','ACTIVE','SYNCING','FAILED') NOT NULL DEFAULT 'INACTIVE' COMMENT '门店集成状态',
  `created_at` int NOT NULL DEFAULT 0 COMMENT '创建时间',
  `updated_at` int NOT NULL DEFAULT 0 COMMENT '更新时间',
  `deleted_at` int NOT NULL DEFAULT 0 COMMENT '删除时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_shop_provider` (`shop_uuid`,`provider_name`),
  KEY `idx_provider_name` (`provider_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='门店第三方集成配置';
```

说明：
- 时间字段 int 秒级，默认 0。
- provider_name 统一使用小写 `grab`（预留其他渠道扩展）。

---

## 🧩 组件和接口

### Protobuf（新增）
- 文件：`ttpos-bmp/app/ttpos-takeout/manifest/protobuf/grab/grab.proto`
- Service：`GrabTakeoutService`（沿用命名），新增 `GetShopProviderCfg`：
  - Request: `shop_uuid` (uint64), `provider_name` (string, optional, default grab)
  - Response: `code`, `message`, `data` 包含 `provider_shop_status`, `provider_merchant_id`, `provider_name`, `shop_uuid`, `updated_at`

生成：
```bash
cd ttpos-bmp/app/ttpos-takeout && make dao
```

### Service / Logic
- 新增接口：`UpsertShopProviderCfg(ctx, shopUUID, providerName, merchantID, status)` 幂等。
- 提供查询接口：`GetShopProviderCfg(ctx, shopUUID, providerName)`。
- 状态映射常量放置在 grab 模块公共处，确保 webhook/旅程共享。

### Controller
- Webhook：在 `grab_v1_integration_status.go` 调用逻辑后追加 Upsert + 消息推送。
- gRPC：新增 rpc 文件（如 `grab_v1_get_shop_provider_cfg.go`），调用查询接口，返回数据/错误码。
- 自助旅程：`self_serve_journey.go` 成功后调用 Upsert。

### 消息推送
- Topic：`takeout_store_integration_state`（参考 `ttpos-bmp/manifest/topics.txt`）。
- Payload 建议字段：`shop_uuid`, `provider_name`, `provider_shop_status`, `provider_merchant_id`, `updated_at`。
- 发送位置：状态更新成功后（Webhook/旅程），失败不推送或推送失败态按业务确认（默认不推）。

### 错误与幂等
- Upsert 使用唯一键（shop_uuid + provider_name），采用事务或 `INSERT ... ON DUPLICATE KEY UPDATE`。
- gRPC not found 返回业务错误码（保持与现有 grab gRPC 风格）。

---

## 🧪 测试策略
- 单测：状态映射、Upsert 幂等、查询无/有数据、枚举校验。
- 集成：模拟 webhook -> 表更新 -> 消息推送；旅程创建 -> Upsert；gRPC 查询返回。
- 迁移：执行 up/down 验证字段/索引。

---

## 📈 性能与可靠性
- Upsert 单条写入，索引命中，预计 < 50ms。
- 消息推送异步，不阻塞主流程（失败记录日志）。
- 并发通知通过唯一键避免脏写，必要时乐观重试一次。

---

## Graphiti & 活动日志
- Related Episode: `[待补充]`
- 活动日志：`docs/team/activities/` 按日记录。

---

**版本**: v1.0.0  
**创建日期**: 2025-12-11  
**作者**: TTPOS Team  
**审核者**: -  
