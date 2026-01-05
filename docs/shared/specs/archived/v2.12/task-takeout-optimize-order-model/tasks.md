# 任务分解：优化外卖订单模型

- **Spec**: [task-takeout-optimize-order-model](./requirements.md)
- **Design**: [task-takeout-optimize-order-model](./design.md)
- **状态**: 进行中

## Task List

### 1. Database Migration
- [x] **创建 Migration 文件** @2025-12-09 ✅
    - 文件: `ttpos-bmp/app/ttpos-takeout/manifest/sql/20251209111123_optimize_order_model.up.sql`
    - 文件: `ttpos-bmp/app/ttpos-takeout/manifest/sql/20251209111123_optimize_order_model.down.sql`
    - 内容: Add `shop_uuid`, Rename `merchant_id` -> `provider_merchant_id`, Update `order_type`.

### 2. Model & Entity Update
- [x] **更新 Entity 定义** @2025-12-09 ✅
    - 文件: `ttpos-bmp/app/ttpos-takeout/internal/model/entity/order.go`
    - 修改: Add `ShopUuid`, Rename `MerchantId` -> `ProviderMerchantId`.
- [x] **更新 DAO/DO** @2025-12-09 ✅
    - 手动更新 `internal/model/do/order.go`
    - 手动更新 `internal/dao/internal/order.go`

### 3. Logic Refactoring
- [x] **替换常量 DeliveryByGrab** @2025-12-09 ✅
    - 已替换 `order_service.go` 中的 `DeliveryByGrab` -> `DeliveryByProvider`
    - 已更新测试文件 `order_service_test.go`
- [x] **更新 Order Service 逻辑** @2025-12-09 ✅
    - 文件: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab/order_service.go`
    - 修改: 添加 `ShopUuid` 字段（TODO: 需要从配置获取）
    - 修改: 使用 `ProviderMerchantId` 替代 `MerchantId`
    - 文件: `ttpos-bmp/app/ttpos-takeout/internal/logic/skootar/create_order.go`
    - 修改: 更新 Skootar 订单创建逻辑

### 4. Verification
- [x] **编译检查** @2025-12-09 ✅
    - 编译通过，所有引用 `MerchantId` 的代码已更新
- [ ] **单元测试** @2025-12-09
    - 运行相关测试，确保无 Regression
    - 注意: `ShopUuid` 字段目前为空字符串，需要后续实现获取逻辑
