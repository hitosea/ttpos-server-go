# 旧商家后台 商品卖点多语言支持 任务分解

> 本文档定义旧商家后台商品卖点多语言支持的执行任务清单。

## 📋 任务分解原则

- 颗粒度 1-4 小时，便于并行
- 明确关联需求编号（requirements.md）
- 便于 AI 协作（每项提供上下文）

## 📊 进度总览

**总任务数**: 16  
**已完成**: 0  
**进行中**: -  
**完成率**: 0%

---

## Phase 1: 数据库与模型

- [ ] **1.1 创建字段迁移文件**
  - File: `admin/database/migrations/20251125103000_add_describe_multi_language_uuid_to_product_package.php`
  - Purpose: 为 `ttpos_product_package` 新增 `describe_multi_language_name_uuid` 列与索引
  - Requirements: 1.1
  - Leverage: `docs/agent/templates/database-migration-template.md`

- [ ] **1.2 数据回填脚本**
  - File: `admin/database/migrations/20251125103500_fill_product_describe_multi_language_uuid.php`
  - Purpose: 将 `describe` 旧值写入 `ttpos_multi_language_name` 并回填新列
  - Requirements: 1.3
  - Leverage: 参考 `20251120083445_seed_order_source_and_nationality_data.php`

- [ ] **1.3 Go Model 更新**
  - File: `main/app/model/product_package.go`
  - Purpose: 添加 `DescribeMultiLanguageNameUuid` 字段
  - Requirements: 1.2

- [ ] **1.4 PHP Model 更新**
  - File: `admin/app/shop/model/product/Product.php`
  - Purpose: 读取/写入 `describe_multi_language_name_uuid`
  - Requirements: 1.2

---

## Phase 2: 旧后台（PHP + Vue）

- [ ] **2.1 PHP 验证器扩展**
  - File: `admin/app/shop/validate/ProductValidate.php`
  - Purpose: 校验 `selling_point_i18n`，限制每语种 500 字符
  - Requirements: 2.1

- [ ] **2.2 Controller 请求解析**
  - File: `admin/app/shop/controller/product/store/Product.php`
  - Purpose: 将 `selling_point_i18n` 解析为结构化数据，写入 Model
  - Requirements: 2.1, 2.2

- [ ] **2.3 Model 多语言保存**
  - File: `admin/app/shop/model/product/Product.php`
  - Purpose: 调用 `MultiLanguageName` Service 保存卖点文本，回填 `describe_multi_language_name_uuid`
  - Requirements: 2.2

- [ ] **2.4 图片/卖点更新事件**
  - File: `admin/app/shop/service/product/ProductService.php`（或现有事件触发点）
  - Purpose: 保存成功后发布 `ProductUpdated` 事件，附带 `describe_multi_language_name_uuid`
  - Requirements: 4.2

- [ ] **2.5 Vue 表单改造（套餐）**
  - File: `admin/views/shop/src/views/product/store/product/part/Buyset.vue`
  - Purpose: 增加多语言输入组件、字数限制、复制主语言按钮
  - Requirements: 2.3

- [ ] **2.6 Vue 表单改造（成品）**
  - File: `admin/views/shop/src/views/product/store/product/{add,edit}.vue`
  - Purpose: 读取/提交 `selling_point_i18n`
  - Requirements: 2.3

---

## Phase 3: Go Main & 缓存

- [ ] **3.1 Repository JOIN 多语言**
  - File: `main/app/repository/product_repo.go`
  - Purpose: 提供 `WithDescribeMultiLang()` 选项 JOIN `ttpos_multi_language_name`
  - Requirements: 3.1

- [ ] **3.2 Service 语言解析**
  - File: `main/app/service/product.go`
  - Purpose: 根据 `lang` 参数返回 `SellingPoint` 与 `SellingPointI18n`
  - Requirements: 3.2

- [ ] **3.3 DTO 更新**
  - File: `main/app/dto/resp/product_resp/product.go`
  - Purpose: 增加 `SellingPoint`/`SellingPointI18n` 字段
  - Requirements: 3.2

- [ ] **3.4 API 返回值**
  - File: `main/app/api/v1/cashier/cashier_product.go`
  - Purpose: 返回新的 DTO 字段，透传 `lang`
  - Requirements: 3.1, 3.3

- [ ] **3.5 缓存封装**
  - File: `main/pkg/cache/product_cache.go`
  - Purpose: 缓存/读取 `selling_point` 多语言，新增 `InvalidateDescribe`
  - Requirements: 4.1

---

## Phase 4: 测试与联调

- [ ] **4.1 单元测试（PHP & Go）**
  - Files: `admin/tests/...`, `main/app/service/product_test.go`
  - Purpose: 覆盖多语言保存/读取、lang fallback
  - Requirements: 全部

- [ ] **4.2 前端联调 & 端到端验证**
  - Files: `admin/views/shop/tests/`, Postman 集成用例
  - Purpose: 验证 500 字符限制、多语言展示、缓存刷新 ≤ 5 分钟
  - Requirements: 2.x, 3.x, 4.x

---

## 提交清单

- [ ] 迁移脚本执行并记录结果
- [ ] Go/PHP/Vue 代码通过 lint & 测试
- [ ] 缓存刷新在日志中可见
- [ ] requirements/design/tasks 三文档同步更新
- [ ] 在 Proposal 中更新关联 Spec 状态

---

## Graphiti & 活动日志

- Graphiti Episode: `[待补充]`
- 活动日志：`docs/team/activities/王昱/2025-11/2025-11-25.md`

---

**模板版本**: v1.0.0  
**创建日期**: 2025-11-25  
**维护者**: 后端开发组

