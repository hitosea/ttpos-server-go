# 商品管理 增加商品详情（新管理端） 任务分解

> 本文档定义商品详情功能的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 13  
**已完成**: 0  
**进行中**: -  
**完成率**: 0%

---

## Phase 1: 数据库设计和迁移

- [ ] 1.1 创建数据库迁移文件

  - File: `admin/database/migrations/{YYYYMMDDHHMMSS}_add_detail_to_ttpos_product_package_table.php`
  - Purpose: 在 `ttpos_product_package` 表中添加 `detail` 字段（LONGTEXT）
  - Requirements: 1.1
  - Leverage: 现有迁移文件示例 `admin/database/migrations/`
  - Prompt: Role: Database Engineer | Task: 为 `ttpos_product_package` 表新增 `detail` LONGTEXT 字段 | Context: 字段注释 `商品详情（富文本）`，遵循 `.cursor/rules/database.mdc` | Restrictions: 迁移中需考虑字段为大文本类型，无默认值，提供回滚方案 | Success: 迁移文件创建成功，字段定义正确

- [ ] 1.2 执行数据库迁移

  - File: -
  - Purpose: 在数据库中新增 `detail` 字段
  - Requirements: 1.1
  - Command: `cd admin && php think migrate:run`
  - Success: 迁移执行成功，表结构更新

- [ ] 1.3 更新 Go Model

  - File: `main/app/model/product.go`
  - Purpose: 在 `ProductPackage` 结构体中增加 `Detail` 字段（longtext 映射）
  - Requirements: 1.4
  - Leverage: 现有字段 `Describe` 定义
  - Prompt: Role: Go Developer | Task: 为 `ProductPackage` 结构体新增 `Detail` 字段，映射 `detail` 列 | Context: 使用 `gorm` 和 `json` 标签，注释 `商品详情（富文本）` | Restrictions: 遵循 `.cursor/rules/go-main.mdc` | Success: 编译通过，字段映射正确

---

## Phase 2: DTO 和 API 层

- [ ] 2.1 更新编辑请求 DTO

  - File: `main/app/dto/req/product_req.go`
  - Purpose: 在 `ProductShopEditReq` 中增加 `Detail` 字段
  - Requirements: 3.1
  - Leverage: 现有字段定义
  - Prompt: Role: Go Developer | Task: 在 `ProductShopEditReq` 中新增 `Detail` 字段，类型 string，可选 | Context: 仅在传入时更新 | Success: DTO 更新成功，binding 校验通过

- [ ] 2.2 更新新增请求 DTO

  - File: `main/app/dto/req/product_req.go`
  - Purpose: 在 `ProductShopAddReq` 中增加 `Detail` 字段
  - Requirements: 4.1
  - Leverage: 现有字段定义
  - Prompt: Role: Go Developer | Task: 在 `ProductShopAddReq` 中新增 `Detail` 字段，类型 string，可选 | Context: 创建商品时可传入富文本 | Success: DTO 更新成功，编译通过

- [ ] 2.3 更新响应 DTO

  - File: `main/app/dto/resp/product_resp/product.go`
  - Purpose: 在 `ProductDetailResp` 中增加 `Detail` 字段
  - Requirements: 2.1
  - Leverage: 现有字段定义
  - Success: DTO 更新成功，API 返回包含 detail

- [ ] 2.4 API 文档同步（可选）

  - File: `docs/shared/api/shop-product-detail.md`（如需新增）
  - Purpose: 记录接口字段变更
  - Requirements: 2.1, 3.1
  - Leverage: `docs/agent/templates/api-doc-template.md`
  - Success: API 文档更新，包含 detail 字段说明

---

## Phase 3: Service 层

- [ ] 3.1 更新商品详情查询逻辑

  - File: `main/app/service/product.go`
  - Purpose: 在 `GetProductDetail` 中返回 `Detail` 字段
  - Requirements: 2.1
  - Leverage: 现有 `ProductDetailResp` 构建逻辑
  - Prompt: Role: Go Developer | Task: 更新 `GetProductDetail`，将 `ProductPackage.Detail` 映射到响应 | Context: 需处理空字符串 | Success: 查询接口返回 detail 字段

- [ ] 3.2 更新商品编辑逻辑

  - File: `main/app/service/product.go`
  - Purpose: 在 `EditProductShop` 处理 `Detail` 字段更新
  - Requirements: 3.1
  - Leverage: 现有商品更新逻辑
  - Prompt: Role: Go Developer | Task: 在保存商品时将 `Detail` 字段写入数据库 | Context: Detail 为空字符串时覆盖旧值 | Success: 保存成功后再次查询返回最新 detail

- [ ] 3.3 更新商品新增逻辑

  - File: `main/app/service/product.go`
  - Purpose: 在 `AddProductShop` 处理 `Detail` 字段写入
  - Requirements: 4.2
  - Leverage: 现有商品新增逻辑
  - Prompt: Role: Go Developer | Task: 在新增商品时将请求体中的 `Detail` 写入 `ProductPackage` 模型 | Context: Detail 可为空字符串，需默认 `""` | Success: 新增后查询接口可返回 detail

---

## Phase 4: 测试

- [ ] 4.1 Repository 层单元测试（可选）

  - File: `main/app/repository/product_repo_test.go`
  - Purpose: 确保 `detail` 字段读写正确
  - Requirements: 1.1, 2.1, 3.1
  - Leverage: 现有测试用例
  - Success: 测试覆盖 CRUD 场景

- [ ] 4.2 Service 层单元测试

  - File: `main/app/service/product_test.go`
  - Purpose: 测试 `GetProductDetail`、`EditProductShop`、`AddProductShop` 的 detail 字段逻辑
  - Requirements: 2.1, 3.1, 4.2
  - Success: 覆盖率达标，detail 字段校验通过

- [ ] 4.3 API 测试

  - File: `main/app/api/v1/shop/shop_product_test.go`（如有）或 Postman 用例
  - Purpose: 验证查询、编辑、添加接口的 detail 字段
  - Requirements: 2.1, 3.1, 4.2
  - Success: 接口测试通过

---

## Phase 5: 文档与发布

- [ ] 5.1 更新 Spec / Proposal 状态

  - File: `docs/team/proposals/2025-11/shop-product-detail-editor.md`
  - Purpose: 在提案中记录 Spec 链接、状态
  - Requirements: 文档要求
  - Success: 提案状态更新，关联 Spec

- [ ] 5.2 活动日志记录

  - File: `docs/team/activities/{user}/{YYYY-MM}/{YYYY-MM-DD}.md`
  - Purpose: 记录 `/spec-create` 活动
  - Requirements: 团队规范
  - Success: 日志追加成功

---

## 提交清单

- requirements/design/tasks 文档已更新
- 数据库迁移脚本和 Go Model 同步
- 商品查询/新增/编辑接口支持 detail 字段
- 单元测试/API 测试全部通过
- 提案与 Spec 互链完成
- 活动日志更新

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 执行任务过程中若总结经验，请记录 Graphiti Episode。

---

**模板版本**: v1.0.0  
**最后更新**: 2025-11-25  
**维护者**: 后端开发组

