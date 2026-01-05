# Grab 店铺集成状态落库与旅程关联 任务分解

> 本文档定义本功能的执行任务清单（颗粒度 1-4 小时）。

## 📊 进度总览

**总任务数**: 11  
**已完成**: 8  
**进行中**: -  
**完成率**: 73%

---

## Phase 1: 数据库与模型

- [x] 1.1 创建迁移脚本 (up/down)
  - File: `ttpos-bmp/app/ttpos-takeout/manifest/sql/20251211163000_create_shop_provider_cfg_table.{up|down}.sql`
  - Purpose: 创建 `shop_provider_cfg`，字段含 `id, uuid, shop_uuid, provider_name, provider_merchant_id, provider_shop_status enum, created_at, updated_at, deleted_at`，唯一键 `shop_uuid+provider_name`
  - Requirements: 1.1-1.3
  - Leverage: 现有迁移示例 `ttpos-bmp/app/ttpos-takeout/manifest/sql/`

- [x] 1.2 执行迁移并生成 dao/entity/do
  - Command: `cd ttpos-bmp/app/ttpos-takeout && make dao`
  - Purpose: 应用表结构并生成访问层代码
  - Requirements: 1.1-1.3

---

## Phase 2: Webhook 集成状态落库 + 消息

- [x] 2.1 状态枚举/映射整理
  - File: `internal/consts/consts.go`（ProviderShopStatus）、`internal/logic/grab/shop_provider_cfg_service.go`（MapGrabIntegrationStatus）
  - Purpose: 将 integrationStatus 映射到枚举 `INACTIVE/ACTIVE/SYNCING/FAILED`（provider_name=grab）
  - Requirements: 2.1

- [x] 2.2 Upsert Service/Logic
  - File: `internal/logic/grab/shop_provider_cfg_service.go`
  - Purpose: Upsert `shop_provider_cfg` 按 `shop_uuid + provider_name=grab`，更新状态与 merchant_id
  - Requirements: 2.1-2.3

- [x] 2.3 Webhook 调用 Upsert + 发消息
  - File: `internal/logic/grab/store_service.go`（HandleIntegrationStatus）
  - Purpose: 处理回调后调用 Upsert，并发布 `takeout_store_integration_state`（含 shop_uuid, provider_name, provider_shop_status, provider_merchant_id, updated_at）
  - Requirements: 2.1-2.4

- [ ] 2.4 单元/集成测试（Webhook）
  - File: `internal/logic/grab/` 测试文件
  - Purpose: 覆盖映射、Upsert 幂等、消息发送调用
  - Requirements: 2.1-2.4

---

## Phase 3: CreateSelfServeJourney 落库

- [x] 3.1 旅程成功后 Upsert
  - File: `internal/logic/grab/self_serve_journey.go`
  - Purpose: 成功时写/更新 `shop_provider_cfg`（状态 SYNCING），失败不写成功态
  - Requirements: 3.1-3.3

- [ ] 3.2 测试（旅程路径）
  - File: 同上测试文件
  - Purpose: 验证旅程成功/失败时落库逻辑
  - Requirements: 3.1-3.3

---

## Phase 4: gRPC 查询接口

- [x] 4.1 更新 protobuf
  - File: `ttpos-bmp/app/ttpos-takeout/manifest/protobuf/grab/grab.proto`
  - Purpose: 新增 `GetShopProviderCfg`（request: shop_uuid, provider_name=grab；response: status/merchant_id/updated_at 等）
  - Requirements: 4.1

- [x] 4.2 生成 pb & 注册路由
  - Command: `cd ttpos-bmp/app/ttpos-takeout && make pb`
  - Purpose: 生成 gRPC 代码（自动注册到 GrabServer）
  - Requirements: 4.1, 4.3

- [x] 4.3 实现 gRPC Controller/Logic
  - File: `internal/controller/rpc/grab/grab_v1_get_shop_provider_cfg.go`
  - Purpose: 按 shop_uuid(+provider_name) 查询并返回；无记录返回 not found
  - Requirements: 4.1-4.4

- [ ] 4.4 gRPC 测试
  - File: rpc/logic 测试文件
  - Purpose: 覆盖有数据/无数据/错误场景
  - Requirements: 4.1-4.4

---

## Phase 5: 验收与文档

- [ ] 5.1 验收自查
  - Purpose: 核对 requirements/design 的所有 AC
  - Requirements: 全部

- [ ] 5.2 文档更新
  - File: `design.md`（必要时补充）、相关 API/Topic 文档
  - Purpose: 保持文档与实现一致
  - Requirements: 全部

---

## Graphiti & 活动日志
- Related Episode: `[待补充]`
- 活动日志：完成关键节点后更新 `docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
