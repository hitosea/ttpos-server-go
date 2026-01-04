> ⚠️ **已归档** - 此 Spec 已随 v2.12 发布。
>
> - 归档时间: 2026-01-04
> - 归档人: weifashi

# 厨显（Grab/LINE MAN外卖相关） 需求文档

> 本文档定义厨显系统中 Grab/LINE MAN 外卖订单标识和商品名称统一显示的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2025-12/v2.12.0-kitchen-display-grab-line-man.md](../../../../team/proposals/2025-12/v2.12.0-kitchen-display-grab-line-man.md) |
| **创建日期**      | 2025-12-25                                                                                                 |
| **负责人**        | weifashi                                                                                                       |
| **目标 Sprint**   | Sprint {N}                                                                                                   |
| **涉及技术栈**    | [x] Go (main/) [ ] Go (ttpos-bmp/) [ ] PHP (admin/) [ ] Vue (admin/views/)                                   |

## 📋 审核状态

| 项目         | 内容                     |
| ------------ | ------------------------ |
| **审核状态** | 已通过 |
| **审核人**   | weifashi             |
| **审核日期** | 2025-12-25             |
| **审核意见** | 需求清晰，验收标准明确，可以进入设计阶段         |

---

## 📋 概述

在厨显系统中，当处理 Grab/LINE MAN 等第三方外卖平台的订单时，需要在"按订单显示"和"按分类显示"两种模式下，清晰标记外卖订单，并确保商品名称统一使用店内商品名称显示，而非第三方平台的商品名称。这将帮助厨房人员快速识别订单类型，减少沟通成本，提升制作效率。

## 🎯 产品对齐

该功能支持以下产品目标：
- **提升厨房效率**：清晰的外卖订单标记帮助厨房人员快速识别订单类型，减少误操作
- **减少沟通成本**：统一的商品名称显示避免因名称差异导致的沟通问题
- **改善用户体验**：确保外卖订单能够准确、及时地制作完成
- **支持多平台扩展**：为未来接入更多第三方外卖平台奠定基础

## 📝 用户故事

**作为** 厨房人员  
**我想** 在厨显中看到外卖订单的明确标识和统一的商品名称  
**以便于** 快速识别订单类型，准确制作订单，减少沟通成本

---

## 功能需求

### Requirement 1: 按订单显示模式下的外卖订单标记

**用户故事**: 作为厨房人员，我想在按订单显示模式下看到外卖订单的明确标识，以便于快速识别订单来源

#### 验收标准

1. **WHEN** 厨显按订单显示模式 **THEN** 显示的订单需要标记外卖标识 **AND** 商品名称均使用店内商品名称
2. **IF** 订单来源为 Grab/LINE MAN 等第三方外卖平台 **THEN** 系统 **SHALL** 自动识别并标记为外卖订单
3. **WHEN** 订单为外卖订单 **THEN** 前端 **SHALL** 在订单列标题或订单卡片上显示"外卖"或"外送"标识

#### 具体要求

- [ ] 1.1 后端接口 `/api/v1/kitchen/product/list_by_order` 返回的数据中，外卖订单的 `is_takeout_bill` 字段为 `true`
- [ ] 1.2 后端接口返回的商品名称（`locale_name`）统一使用店内商品名称，而非第三方平台的商品名称
- [ ] 1.3 前端在订单列标题或订单卡片上显示外卖标识（如"外卖"、"外送"标签或图标）
- [ ] 1.4 外卖标识的视觉样式应清晰醒目，易于识别（如使用不同颜色、图标等）
- [ ] 1.5 支持 Grab 和 LINE MAN 两个平台的订单识别和标记

#### 相关接口

- **接口路径**: `GET /api/v1/kitchen/product/list_by_order`
- **接口文件**: `main/app/api/v1/kitchen/kitchen_product.go`
- **服务实现**: `main/app/service/production.go::GetProductListByOrder`
- **响应结构**: `main/app/dto/resp/production.go::ProductionListWithPagination`
  - `list[].is_takeout_bill`: 是否是外送订单（bool）
  - `list[].product_list.list[].locale_name`: 商品名称（多语言）

---

### Requirement 2: 按分类显示模式下的外卖订单标记

**用户故事**: 作为厨房人员，我想在按分类显示模式下看到外卖商品的明确标识，以便于快速识别商品来源

#### 验收标准

1. **WHEN** 厨显按分类显示模式 **THEN** 显示的商品需要标记外卖标识 **AND** 商品名称均使用店内商品名称
2. **IF** 商品属于 Grab/LINE MAN 等第三方外卖平台的订单 **THEN** 系统 **SHALL** 自动识别并标记为外卖商品
3. **WHEN** 商品为外卖订单商品 **THEN** 前端 **SHALL** 在商品卡片上显示"外卖"或"外送"标识

#### 具体要求

- [ ] 2.1 后端接口 `/api/v1/kitchen/product/list_by_category` 返回的数据中，外卖订单商品的 `is_takeout_bill` 字段为 `true`
- [ ] 2.2 后端接口返回的商品名称（`locale_name`）统一使用店内商品名称，而非第三方平台的商品名称
- [ ] 2.3 前端在商品卡片上显示外卖标识（如"外卖"、"外送"标签或图标）
- [ ] 2.4 外卖标识的视觉样式应清晰醒目，易于识别（如使用不同颜色、图标等）
- [ ] 2.5 支持 Grab 和 LINE MAN 两个平台的订单识别和标记

#### 相关接口

- **接口路径**: `GET /api/v1/kitchen/product/list_by_category`
- **接口文件**: `main/app/api/v1/kitchen/kitchen_product.go`
- **服务实现**: `main/app/service/production.go::GetProductListByCategory`
- **响应结构**: `main/app/dto/resp/production.go::ProductionListWithPagination`
  - `list[].is_takeout_bill`: 是否是外送订单（bool）
  - `list[].product_list.list[].locale_name`: 商品名称（多语言）

---

### Requirement 3: 商品名称统一显示

**用户故事**: 作为厨房人员，我想看到统一的商品名称，以便于准确理解商品信息

#### 验收标准

1. **IF** 第三方平台商品名称与店内商品名称不同 **THEN** 系统 **SHALL** 显示店内商品名称
2. **WHEN** 商品名称显示在厨显界面 **THEN** 系统 **SHALL** 使用多语言支持的商品名称（`locale_name`）
3. **WHEN** 商品名称映射失败或不存在 **THEN** 系统 **SHALL** 使用第三方平台商品名称作为兜底

#### 具体要求

- [ ] 3.1 后端接口返回的商品名称（`locale_name`）必须使用店内商品名称
- [ ] 3.2 商品名称支持多语言显示（根据厨显设备语言设置）
- [ ] 3.3 商品名称映射逻辑应确保准确性，避免映射错误
- [ ] 3.4 当商品映射关系不存在时，应记录日志并返回第三方平台商品名称

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 严格遵循 Controller → Service → Repository 分层
- **单一职责原则**: 每个文件应有单一、明确的目的
- **模块化设计**: Service 和 Repository 应独立且可复用
- **依赖管理**: Service 只能依赖其他 Service 接口，不能直接依赖 Repository
- **遵循规范**:
  - `.cursor/rules/go-main.mdc` - Go Main 开发规范
  - `.cursor/rules/go-bmp.mdc` - Go BMP 微服务规范
  - `.cursor/rules/php.mdc` - PHP 开发规范
  - `.cursor/rules/vue.mdc` - Vue 前端规范
  - `.cursor/rules/structs.mdc` - 项目结构规范

### API 设计要求

- [x] URL 使用 snake_case 命名（如：`/api/v1/kitchen/product/list_by_order`）
- [x] data 字段必须是对象，不能是 null 或数组
- [x] 分页信息统一放在 meta 中
- [x] 响应格式：`{code, message, data{}}`
- [x] 参考: `.cursor/rules/api.mdc` - API 设计规范

### 数据库设计要求

- [x] 必须包含: `id`, `uuid`, `create_time`, `update_time`, `delete_time`
- [x] 时间字段使用 int 类型，\_time 结尾，默认值 0
- [x] 金额字段使用 decimal(20,8)
- [x] UUID 字段使用 bigint unsigned
- [x] 表名使用 ttpos\_ 前缀
- [x] 字段名使用 snake_case
- [x] 参考: `.cursor/rules/database.mdc` - 数据库开发规范

### 性能要求

- [ ] 本地响应时间 < 200ms
- [ ] 数据库查询优化（使用索引）
- [ ] 缓存策略（Redis）
- [ ] 并发处理（使用 UUID 锁）

### 浏览器兼容性（管理后台）

- [ ] Chrome 90+
- [ ] Safari 14+
- [ ] Firefox 88+
- [ ] Edge 90+

### 测试要求

- [ ] Service 层测试覆盖率 ≥ 70%
- [ ] Repository 层测试覆盖率 ≥ 80%
- [ ] **Payment/Order 相关模块测试覆盖率 100%**（高风险）
- [ ] 集成测试覆盖核心流程
- [ ] API 测试覆盖所有接口
- [ ] 参考: `.cursor/rules/go-main.mdc` - 测试规范

### 国际化要求

- [ ] 支持 10 种语言（中文、英文、日语、韩语等）
- [ ] 所有文案使用多语言实现
- [ ] 参考: `main/i18n/` - 国际化配置

### 安全要求

- [ ] 所有 API 需要身份验证
- [ ] 敏感数据加密存储
- [ ] SQL 注入防护（使用参数化查询）
- [ ] XSS 防护（前端输入校验）
- [ ] CSRF 防护（Token 验证）
- [ ] 参考: `.cursor/rules/security.mdc` - 安全开发规范

### 可靠性要求

- [ ] 网络异常时优雅降级
- [ ] 事务管理（保证数据一致性）
- [ ] 错误日志记录（使用 Logger）
- [ ] 故障恢复机制

---

## 验收标准

### 功能验收

1. **按订单显示模式验收**: 在按订单显示模式下，外卖订单能够清晰标记，商品名称使用店内商品名称
2. **按分类显示模式验收**: 在按分类显示模式下，外卖商品能够清晰标记，商品名称使用店内商品名称
3. **商品名称统一验收**: 所有显示的商品名称均使用店内商品名称，而非第三方平台商品名称
4. **多平台支持验收**: 支持 Grab 和 LINE MAN 两个平台的订单识别和标记

### 测试验收

1. **单元测试**: 覆盖率达标
2. **API 测试**: 所有接口测试通过
3. **集成测试**: 端到端流程测试通过
4. **手动测试**: 浏览器兼容性测试通过

### 文档验收

1. **技术文档**: design.md 完整且准确
2. **API 文档**: API 接口文档完整（如有）
3. **数据库文档**: 迁移脚本和表结构文档完整
4. **测试文档**: tasks.md 中的测试任务完成

---

## 约束条件

### 技术约束

#### Go Main 模块

- 必须使用 Gin 框架
- 接口以 `I` 开头，实现以 `Impl` 结尾
- Service 只能依赖其他 Service 接口
- Repository 只能持有 db 实例，不能持有 DBManager
- 不使用 panic，返回 error

#### Go BMP 模块

- 必须使用 GoFrame 2.x
- 禁止修改 dao/entity/do/ 目录（自动生成）
- gRPC 服务必须注册到 Nacos
- 遵循 `ttpos-bmp/.cursor/rules/go-rules.mdc`

#### PHP 模块

- 必须使用 ThinkPHP 6.0
- 遵循 MVC 分层
- Controller 不写业务逻辑
- 使用验证器验证参数
- 使用软删除

#### Vue 模块

- 必须使用 Vue 3 + TypeScript + Vite
- 使用 Element Plus 组件库
- 遵循 `.cursor/rules/vue.mdc`

### 业务约束

- 前端仓库：`all-kds-grab-order-display`（厨显）
- 仅影响 KDS 厨显端，不影响其他终端
- 需要确保与现有厨显功能兼容

### 资源约束

- 开发时间: 3-5 天
- Story Point: 3-5 SP (必须 ≤ 5)

---

## 依赖关系

### 技术依赖

- `main/app/service/production.go` - 送厨商品服务
- `main/app/dto/resp/production.go` - 响应数据结构
- `main/app/modules/takeout/domain/value_object/takeout_platform.go` - 外卖平台常量定义

### 服务依赖

- **Main → BMP**: 无（本功能不涉及 gRPC 调用）
- **Admin → Main**: 无（本功能不涉及管理后台）
- **Frontend → Main**: HTTP API 调用（`/api/v1/kitchen/product/list_by_order` 和 `/api/v1/kitchen/product/list_by_category`）

### 业务依赖

- 外卖订单已正确创建并关联到销售账单
- 商品名称映射关系已建立（第三方平台商品与店内商品的映射）
- 厨显设备已正确配置

---

## 风险和缓解

### 风险 1: 第三方平台订单数据结构可能与现有订单结构存在差异

**影响**: 中  
**概率**: 中  
**缓解措施**:

- 提前梳理第三方平台订单数据结构，确保数据兼容性
- 在 Service 层增加数据转换和校验逻辑
- 增加单元测试覆盖边界情况

### 风险 2: 商品名称映射关系需要确保准确性，避免映射错误

**影响**: 高  
**概率**: 中  
**缓解措施**:

- 建立完善的商品名称映射机制，支持手动维护和自动匹配
- 增加映射关系校验逻辑
- 当映射失败时记录日志并返回第三方平台商品名称作为兜底

### 风险 3: 前端显示性能可能受订单数量影响

**影响**: 低  
**概率**: 低  
**缓解措施**:

- 优化前端渲染逻辑，采用虚拟滚动等技术提升性能
- 后端接口支持分页查询
- 增加前端缓存机制

---

## 时间表

- **Phase 1 - 后端接口改造**: 2 天
  - 确保接口返回 `is_takeout_bill` 字段正确标识外卖订单
  - 确保商品名称使用店内商品名称
- **Phase 2 - 前端显示实现**: 2 天
  - 实现按订单显示模式下的外卖标识
  - 实现按分类显示模式下的外卖标识
- **Phase 3 - 测试和优化**: 1 天
  - 单元测试和集成测试
  - 性能优化和 bug 修复
- **总计**: 5 天（SP = 5）

---

## 参考资料

### 核心规范

- `.cursor/rules/go-main.mdc` - Go Main 核心约束
- `.cursor/rules/go-bmp.mdc` - Go BMP 开发规范
- `.cursor/rules/php.mdc` - PHP 核心约束
- `.cursor/rules/vue.mdc` - Vue 开发规范
- `.cursor/rules/api.mdc` - API 设计规范
- `.cursor/rules/database.mdc` - 数据库开发规范
- `.cursor/rules/security.mdc` - 安全开发规范

### 架构文档

- `docs/human/architecture/go-main-architecture.md` - Go Main 架构
- `docs/human/architecture/go-bmp-architecture.md` - Go BMP 架构
- `docs/human/architecture/php-architecture.md` - PHP 架构
- `docs/human/architecture/database-design.md` - 数据库设计

### 开发指南

- `docs/human/guides/go-main-development.md` - Go Main 开发指南
- `docs/human/guides/go-bmp-development.md` - Go BMP 开发指南
- `docs/human/guides/php-development.md` - PHP 开发指南
- `docs/human/guides/api-design-guide.md` - API 设计指南
- `docs/human/guides/database-guide.md` - 数据库开发指南

### 外部参考

- DooTask #37493 - 原始需求任务
- 前端仓库: `all-kds-grab-order-display`（厨显）

### 相关代码文件

- `main/app/api/v1/kitchen/kitchen_product.go` - 厨显产品 API 控制器
- `main/app/service/production.go` - 送厨商品服务实现
- `main/app/dto/resp/production.go` - 响应数据结构
- `main/app/modules/takeout/domain/value_object/takeout_platform.go` - 外卖平台常量定义

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 提醒：需求评审或范围调整若形成经验，应同步更新 Episode 并在 Proposal/Spec 互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-12-25  
**作者**: weifashi  
**审核者**: {审核者}

