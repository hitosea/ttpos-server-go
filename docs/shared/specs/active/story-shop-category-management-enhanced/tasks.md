# 新管理端-分类管理-增强分类 任务分解

> 本文档定义新管理端分类管理增强功能的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 12  
**已完成**: 8  
**进行中**: -  
**完成率**: 67%

**补充需求**（已实现）：
- ✅ 店内显示不允许取消：`is_display_in_store` 必须始终为 1
- ✅ 被 Grab 商品勾选的分类不允许取消外卖显示：如果 `takeout_product_count > 0`，则 `is_display_in_takeout` 不能设置为 0

---

## Phase 1: 数据库设计和迁移

- [x] 1.1 创建数据库迁移文件

  - File: `admin/database/migrations/{YYYYMMDDHHMMSS}_add_category_display_fields.php`
  - Purpose: 在 `ttpos_product_category` 表增加 `is_display_in_store` 和 `is_display_in_takeout` 字段
  - Requirements: 1.1, 1.2
  - Leverage: 现有迁移文件: `admin/database/migrations/20251208232558_create_product_package_takeout_table.php`
  - Prompt: Role: Database Engineer | Task: 创建迁移文件，为 ttpos_product_category 表添加 is_display_in_store 和 is_display_in_takeout 字段 | Context: 必须检查字段是否存在（使用 IF NOT EXISTS），字段类型 tinyint(1)，默认值分别为 1 和 0，添加在 status 字段之后 | Restrictions: 遵循 .cursor/rules/database.mdc，迁移前检查字段是否存在确保幂等性 | Success: 迁移文件创建成功，字段定义正确

- [ ] 1.2 执行数据库迁移

  - File: -
  - Purpose: 在数据库中执行迁移，添加字段
  - Requirements: 1.1, 1.2
  - Leverage: Task 1.1 的迁移文件
  - Command: `cd admin && php think migrate:run`
  - Success: 迁移执行成功，字段已添加

- [x] 1.3 更新 Go Model

  - File: `main/app/model/category.go`
  - Purpose: 在 ProductCategory 结构体中添加新字段
  - Requirements: 1.1, 1.2
  - Leverage: 现有 Model: `main/app/model/category.go`
  - Prompt: Role: Go Developer | Task: 在 ProductCategory 结构体中添加 IsDisplayInStore 和 IsDisplayInTakeout 字段 | Context: 使用 gorm 标签，字段类型 int，默认值分别为 1 和 0 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: Model 更新成功，字段映射正确

---

## Phase 2: DTO 和 Service 层

- [x] 2.1 更新 Request DTO

  - File: `main/app/dto/req/product.go`
  - Purpose: 在分类请求 DTO 中添加 `is_display_in_store` 和 `is_display_in_takeout` 字段
  - Requirements: 1.3, 1.4
  - Leverage: 现有 DTO: `main/app/dto/req/product.go`
  - Prompt: Role: Go Developer | Task: 在 CategoryCreateReq 和 CategoryUpdateReq 中添加 IsDisplayInStore 和 IsDisplayInTakeout 字段 | Context: 使用 *int 类型，允许 nil（使用默认值） | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: DTO 更新成功

- [x] 2.2 更新 Response DTO

  - File: `main/app/dto/resp/product_resp/category.go`（需创建或增强）
  - Purpose: 在分类响应 DTO 中添加新字段
  - Requirements: 1.5, 1.7
  - Leverage: 现有 DTO: `main/app/dto/resp/product_resp/`
  - Prompt: Role: Go Developer | Task: 创建或更新 CategoryResp，添加 IsDisplayInStore、IsDisplayInTakeout 和 TakeoutProductCount 字段 | Context: 包含所有新字段，TakeoutProductCount 类型为 int64 | Restrictions: data 必须是对象 | Success: DTO 更新成功

- [x] 2.3 增强分类 Service Create 方法

  - File: `main/app/service/product.go`（已增强）
  - Purpose: 在创建分类时支持新字段，并验证业务规则
  - Requirements: 1.3, 1.6, 1.8
  - Leverage: 现有 Service: `main/app/service/product.go`
  - Note: 已实现验证逻辑：店内显示不允许取消（is_display_in_store 必须为 1）

- [x] 2.4 增强分类 Service Update 方法

  - File: `main/app/service/product.go`（已增强）
  - Purpose: 在更新分类时支持新字段，并验证业务规则
  - Requirements: 1.4, 1.6, 1.8, 1.9
  - Leverage: 现有 Service: `main/app/service/product.go`
  - Note: 已实现验证逻辑：
    - 店内显示不允许取消（is_display_in_store 不能设置为 0）
    - 被 Grab 商品勾选的分类不允许取消外卖显示（如果 takeout_product_count > 0，is_display_in_takeout 不能设置为 0）

- [x] 2.5 增强分类 Service GetByUuid 方法

  - File: `main/app/service/category_srv.go`（需创建或增强）
  - Purpose: 在获取分类详情时统计被外卖商品选中的数量
  - Requirements: 1.7
  - Leverage: 现有 Service: `main/app/service/`，外卖商品 Model: `main/app/model/product_package_takeout.go`
  - Prompt: Role: Go Developer | Task: 增强 CategorySrv 的 GetByUuid 方法，查询该分类被外卖商品选中的数量（统计 ttpos_product_package_takeout 表中 category_uuid 等于分类 uuid 且 delete_time=0 的记录数） | Context: 查询失败时设置为 0，不影响主流程 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: Service 方法增强成功，统计逻辑正确

---

## Phase 3: API 层

- [x] 3.1 增强分类创建 API

  - File: `main/app/api/v1/shop/shop_product.go`
  - Purpose: 分类创建 API 支持新字段
  - Requirements: 1.3
  - Leverage: 现有 API: `main/app/api/v1/shop/shop_product.go`
  - Note: API 层无需修改，已自动支持新字段（通过 DTO 绑定）

- [x] 3.2 增强分类编辑 API

  - File: `main/app/api/v1/shop/shop_product.go`
  - Purpose: 分类编辑 API 支持新字段
  - Requirements: 1.4
  - Leverage: 现有 API: `main/app/api/v1/shop/shop_product.go`
  - Note: API 层无需修改，已自动支持新字段（通过 DTO 绑定）

- [x] 3.3 增强分类查询 API（列表和详情）

  - File: `main/app/api/v1/shop/shop_product.go`
  - Purpose: 分类查询 API 返回新字段
  - Requirements: 1.5, 1.7
  - Leverage: 现有 API: `main/app/api/v1/shop/shop_product.go`
  - Note: API 层无需修改，Service 已返回包含新字段的响应

---

## Phase 4: 测试

- [ ] 4.1 编写 Service 单元测试

  - File: `main/app/service/category_srv_test.go`（需创建或增强）
  - Purpose: 测试分类创建和更新时的验证逻辑
  - Requirements: 测试要求
  - Leverage: 现有测试: `main/app/service/*_srv_test.go`
  - Prompt: Role: QA Engineer | Task: 为 CategorySrv 编写单元测试，测试至少开启一个显示渠道的验证逻辑 | Context: 测试正常场景和边界场景 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试通过

- [ ] 4.2 编写 API 集成测试

  - File: `main/app/api/v1/shop/shop_product_test.go`（需创建或增强）
  - Purpose: 测试分类 API 接口
  - Requirements: 测试要求
  - Leverage: 现有测试: `main/app/api/*_api_test.go`
  - Prompt: Role: QA Engineer | Task: 为分类 API 编写集成测试 | Context: 测试创建、编辑、查询接口，测试新字段 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 所有 API 测试通过

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 `go fmt` 和 `go vet`
- [ ] 所有测试通过

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成

### 规范遵循

- [ ] 遵循 `.cursor/rules/go-main.mdc`
- [ ] 遵循 `.cursor/rules/api.mdc`
- [ ] 遵循 `.cursor/rules/database.mdc`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/story-shop-category-management-enhanced/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-shop-category-management-enhanced/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-shop-category-management-enhanced/tasks.md
```

---

**模板版本**: v1.0.0  
**最后更新**: 2025-12-09  
**维护者**: 后端开发组
