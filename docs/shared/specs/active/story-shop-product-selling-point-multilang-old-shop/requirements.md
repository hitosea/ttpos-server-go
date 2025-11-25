# 旧商家后台 商品卖点多语言支持 需求文档

> 本文档定义旧商家后台商品卖点多语言能力的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                                         |
| ----------------- | ---------------------------------------------------------------------------------------------------------------------------- |
| **来源 Proposal** | [docs/team/proposals/2025-11-25-product-selling-point-multilang-old-shop.md](../../../team/proposals/2025-11-25-product-selling-point-multilang-old-shop.md) |
| **创建日期**      | 2025-11-25                                                                                                                   |
| **负责人**        | {姓名}                                                                                                                       |
| **目标 Sprint**   | Sprint {N}                                                                                                                   |
| **涉及技术栈**    | [x] Go (main/) [ ] Go (ttpos-bmp/) [x] PHP (admin/) [x] Vue (admin/views/)                                                   |
| **关联任务**      | DooTask #36924                                                                                                               |

---

## 📋 概述

旧商家后台（PHP + Vue 旧管理端）仅支持单语言的 `describe` 字段保存商品卖点，无法针对不同终端语言展示本地化文案。需要在商品管理模块（套餐 + 成品）引入 `describe_multi_language_name_uuid` 字段，复用 `ttpos_multi_language_name` 表存储卖点的多语言内容，并确保会员端、平板、H5、菜单等前台终端按用户语言展示一致的卖点与最新商品图片。

## 🎯 产品对齐

- 为跨语言门店提供统一的卖点管理，保障品牌调性一致
- 降低运营多语言拷贝与人工同步的成本
- 确保消费者终端展示的卖点/图片实时且匹配所选语言，提升下单转化率

## 📝 用户故事

**作为** 商家后台运营  
**我想** 在编辑套餐或成品商品时能为不同语言录入独立卖点，并控制 500 字符限制  
**以便于** 会员端、平板和 H5 菜单能够展示符合语言环境的卖点，提高顾客转化

---

## 功能需求

### Requirement 1: 数据模型扩展 `describe_multi_language_name_uuid`

**用户故事**: 作为平台维护人员，我需要在商品主表里存储卖点多语言引用，以便后端 API 可以按语言返回正确内容。

#### 验收标准

1. **WHEN** 数据库迁移执行完成 **THEN** `ttpos_product_package` 表新增 `describe_multi_language_name_uuid`（BIGINT UNSIGNED，默认 0），并建立索引。
2. **WHEN** 保存商品时 **THEN** 系统 **SHALL** 将卖点文案写入 `ttpos_multi_language_name` 并将返回的 `uuid` 写入 `describe_multi_language_name_uuid`。
3. **IF** 旧数据存在 `describe` 文案 **THEN** 迁移脚本 **SHALL** 创建对应的多语言记录（以中文为默认），并回填 `describe_multi_language_name_uuid`。
4. **WHEN** 终端请求的语言在多语言记录中为空 **THEN** 系统 **SHALL** 回退到主语言（中文）并记录缺失指标。

#### 具体要求

- [ ] 1.1 编写迁移文件为 `ttpos_product_package` 添加 `describe_multi_language_name_uuid` 字段与 `idx_describe_multi_lang_uuid` 索引。
- [ ] 1.2 更新 Go model `main/app/model/product_package.go` 与 PHP Model `admin/app/shop/model/product/Product.php`。
- [ ] 1.3 编写一次性脚本/迁移，将 `describe` 现有值迁移到 `ttpos_multi_language_name`，并填充新列。
- [ ] 1.4 迁移后保留 `describe` 字段作为 fallback，但写入逻辑以多语言表为准。

---

### Requirement 2: 旧后台录入体验（套餐 & 成品）

**用户故事**: 作为商家运营，我希望在旧后台界面中针对所有支持语言填写卖点并实时查看字数限制，以减少错误输入。

#### 验收标准

1. **WHEN** 商家在商品新增/编辑页录入卖点 **THEN** 每种语言均提供独立输入域，限制 500 字符并显示计数。
2. **WHEN** 提交表单 **THEN** 后端 **SHALL** 校验各语言长度、必填逻辑（至少提供主语言），并将数据写入多语言表。
3. **WHEN** 商户切换套餐/成品商品 **THEN** 表单会加载对应的多语言卖点并支持编辑。
4. **WHEN** 卖点或图片保存成功 **THEN** 提示文案明确告知所有终端需要的刷新时间（≤ 5 分钟）。

#### 具体要求

- [ ] 2.1 PHP Controller `admin/app/shop/controller/product/store/Product.php` 增加多语言请求体解析与 500 字符校验。
- [ ] 2.2 PHP Model 保存逻辑调用 `MultiLanguageName` Service，生成/更新 `describe_multi_language_name_uuid`。
- [ ] 2.3 Vue 旧后台页面（`admin/views/shop/src/views/product/store/product/*.vue`）复用多语言输入组件，支持批量复制主语言。
- [ ] 2.4 表单提交 payload 增加 `selling_point_i18n` 字段，兼容老字段 `selling_point` 作为主语言 fallback。

---

### Requirement 3: 会员端/平板/H5 多语言展示

**用户故事**: 作为顾客，我希望在会员端、平板和 H5 菜单查看商品详情时看到符合我语言的卖点与最新图片。

#### 验收标准

1. **WHEN** 前台终端（会员、平板、H5）获取商品详情 **THEN** 返回对象包含 `selling_point`（当前语言）与 `selling_point_i18n`（完整语言 map）。
2. **WHEN** 终端携带 `lang` 参数 **THEN** 商品服务 **SHALL** 根据 `describe_multi_language_name_uuid` 联表 `ttpos_multi_language_name`，返回匹配语言。
3. **WHEN** 没有对应语言翻译 **THEN** 返回主语言文本并在日志中记录缺失语言。
4. **WHEN** 商品图片替换 **THEN** 详情接口和缓存将在 5 分钟内返回最新图片 URL。

#### 具体要求

- [ ] 3.1 Go API `main/app/api/v1/cashier/cashier_product.go` 及 `main/app/service/product.go` 增加 `describe_multi_language_name_uuid` 查询。
- [ ] 3.2 DTO `main/app/dto/resp/product_resp/product.go` 新增 `SellingPoint` & `SellingPointI18n` 字段。
- [ ] 3.3 Redis/本地缓存命中时也要携带多语言数据，避免二次查询。
- [ ] 3.4 Tablet/H5/Member 前端在下游仓库（ttpos-flutter/ttpos-menu 等）创建跟踪任务，确保消费新的字段。

---

### Requirement 4: 缓存与图片刷新策略

**用户故事**: 作为系统管理员，我需要卖点或图片变更后快速覆盖终端缓存，避免顾客仍看到旧内容。

#### 验收标准

1. **WHEN** 卖点或图片更新成功 **THEN** Shop 后台 **SHALL** 发布刷新事件或直接清理 `product_detail:{uuid}` 缓存。
2. **WHEN** 终端在 5 分钟后重新获取详情 **THEN** 必须命中新文本与图片。
3. **IF** 缓存清理失败 **THEN** 记录告警日志并提供重试任务（支持后台按钮手动触发）。

#### 具体要求

- [ ] 4.1 复用 `pkg/cache/product_cache.go`（或现有缓存封装）提供 `InvalidateDescribe(uuid)` 方法。
- [ ] 4.2 后台保存商品后调用缓存刷新并触发消息队列（如目前已有的 `ProductUpdated` 事件）。
- [ ] 4.3 在操作日志中记录卖点/图片更新的操作者与时间，便于追踪。

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: Controller → Service → Repository 分层；PHP 遵循 MVC，Go 服务遵循接口实现约束。
- **遵循规范**:
  - `.cursor/rules/go-main.mdc`
  - `.cursor/rules/php.mdc`
  - `.cursor/rules/vue.mdc`
  - `.cursor/rules/api.mdc`
  - `.cursor/rules/database.mdc`

### API 设计要求

- URL 使用 snake_case，响应 `data` 必须为对象。
- Shop 后台接口保持 `/shop/product/*` 命名，Go API 返回 `{code,message,data{}}`。
- 所有语言相关字段统一命名为 `*_i18n`。

### 数据库设计要求

- 所有新增字段遵循 `snake_case`，类型、索引符合 `.cursor/rules/database.mdc`。
- `describe_multi_language_name_uuid` 默认值 0，不允许 NULL，并建立普通索引。
- 迁移脚本需幂等，支持回滚。

### 性能要求

- 后台保存商品整流程 ≤ 2s。
- 商品详情接口本地响应时间 < 200ms，JOIN 多语言表需使用索引。
- 缓存命中率 > 80%。

### 测试要求

- PHP Service 单元测试 ≥ 70%。
- Go Service/Repository 单元测试 ≥ 70% / ≥ 80%。
- 前端提供 e2e 用例覆盖卖点多语言输入。

### 国际化要求

- 支持现有 9 种语言（ZH/EN/TH/ZHTW/JA/KO/MY/TR/SV），后续新增语言需免代码改动。
- 输入验证应支持全角字符统计。

### 安全与可靠性

- 所有接口需要认证与权限检查。
- 防范 XSS：前端对输入进行转义，后端仅存文本。
- 缓存刷新失败需记录可观察性指标并支持重试。

---

## 验收标准

### 功能验收

1. **多语言字段**: `ttpos_product_package` 新增列并成功回填历史数据。
2. **后台录入**: 套餐与成品商品页面支持多语言卖点，限制 500 字符。
3. **终端展示**: 会员/平板/H5 端能够按语言展示卖点，且 5 分钟内看到图片更新。
4. **缓存策略**: 卖点或图片更新后缓存被清理，接口返回最新数据。

### 测试验收

1. PHP & Go 单元测试覆盖率达标。
2. 前端联调通过，包含多语言切换与长度校验。
3. 集成测试验证缓存刷新和多终端展示路径。

### 文档验收

1. design.md、tasks.md 与实现保持一致。
2. 数据库迁移与 API 变更同步在项目文档及 CHANGELOG 中记录。
3. 若形成经验，需要创建 Graphiti Episode 并互链。

---

## 约束条件

### 技术约束

- Go Main：不允许 panic，Service 不得直接依赖 Repository。
- PHP：沿用 ThinkPHP6，Controller 不写业务逻辑，使用验证器。
- Vue：必须为 Composition API + Element Plus，多语言文本使用 i18n。

### 业务约束

- 旧商家后台仍需兼容尚未升级的新商户，不能强制一次性填写所有语言。
- 缓存刷新时间 ≤ 5 分钟，否则需提供回滚方案。

### 资源约束

- 开发时间：5 天
- 预估 Story Point：8

---

## 依赖关系

- `ttpos_multi_language_name` 表及相关 Service/Model。
- 会员/平板/H5 终端仓库需配合消费新字段（需在对应仓库创建任务）。
- 现有商品图片缓存通道（若使用 CDN，需要更新 purge 策略）。

---

## 风险和缓解

### 风险 1: 老数据迁移失败或重复写入

- **影响**: 高；历史卖点丢失
- **概率**: 中
- **缓解措施**:
  - 迁移脚本分批执行并记录进度
  - 提供回滚脚本，将 `describe` 重新写回

### 风险 2: 终端未同步新字段导致空白展示

- **影响**: 中
- **概率**: 中
- **缓解措施**:
  - 在响应中保留 `selling_point`（字符串）作为兼容字段
  - 为终端新增 feature flag，灰度切换

---

## 时间表

- **Phase 1 - 数据模型与迁移**: 1.5 天
- **Phase 2 - 后台录入体验**: 2 天
- **Phase 3 - 前台展示 & 缓存**: 1.5 天
- **Phase 4 - 测试与发布**: 0.5 天
- **总计**: 5 天（SP = 8）

---

## 参考资料

- `.cursor/rules/api.mdc`
- `.cursor/rules/database.mdc`
- `.cursor/rules/go-main.mdc`
- `.cursor/rules/php.mdc`
- `.cursor/rules/vue.mdc`
- `docs/shared/specs/story-shop-product-detail-editor/`（同一模块字段扩展参考）

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 活动日志：`docs/team/activities/王昱/2025-11/2025-11-25.md`
- 若迁移方案或多语言适配形成经验，请同步创建 Episode。

---

**版本**: v1.0.0  
**创建日期**: 2025-11-25  
**作者**: 产品 + 后端联合  
**审核者**: 待评审

