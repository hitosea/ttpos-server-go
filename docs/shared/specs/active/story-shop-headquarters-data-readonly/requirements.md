# 参考商品单位实现，来源总部的数据不可编辑 需求文档

> 本文档定义参考商品单位（ProductUnit）的实现方式，为多个模块实现总部来源数据不可编辑功能的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2025-12/reference-product-unit-headquarters-readonly.md](../../../../team/proposals/2025-12/reference-product-unit-headquarters-readonly.md) |
| **来源任务** | DooTask #37479 |
| **创建日期**      | 2025-12-08                                                                                                 |
| **负责人**        | 待分配                                                                                                       |
| **目标 Sprint**   | Sprint TBD                                                                                                   |
| **涉及技术栈**    | [x] Go (main/) [ ] Go (ttpos-bmp/) [ ] PHP (admin/) [x] Vue (admin/views/)                                   |

## 📋 审核状态

| 项目         | 内容                     |
| ------------ | ------------------------ |
| **审核状态** | 待审核                   |
| **审核人**   | {审核人姓名}             |
| **审核日期** | {YYYY-MM-DD}             |
| **审核意见** | {审核意见，如有}         |

---

## 📋 概述

参考商品单位（ProductUnit）模块的实现方式，为以下模块实现总部来源数据不可编辑的功能：
- **菜品标签**（ProductLabel）
- **满额减**（FullReductionActivity）
- **商品**（ProductPackage）

**参考实现**：商品单位（ProductUnit）模块已经完整实现了总部来源数据不可编辑的功能：
- ✅ 列表/详情接口返回 `is_editable` 字段
- ✅ 编辑/删除接口增加总部来源数据校验
- ✅ 使用 `isEditable(ctx, headquarterUuid)` 函数判断
- ✅ 前端根据 `is_editable` 字段控制 UI

**业务价值**：
- **数据一致性**：确保总部数据在分店中保持统一，避免分店误操作导致数据不一致
- **统一管理**：总部可以统一管理数据，分店只能使用，不能修改
- **降低风险**：减少因分店误编辑导致的数据错误和业务风险
- **用户体验**：明确区分可编辑和不可编辑的数据，提升用户操作体验

---

## 🎯 产品对齐

此功能是"总部-分店颗粒化同步"功能的一部分，确保同步后的总部数据在分店中保持只读状态，与商品单位（ProductUnit）模块保持一致的用户体验。

---

## 📝 用户故事

**作为** 店长/运营人员  
**我想** 在分店中查看和使用总部的数据（菜品标签、满额减、商品），但不能编辑来源总部的数据  
**以便于** 保持数据一致性，避免误操作导致的数据错误

**作为** 总部运营人员  
**我想** 创建可被分店同步使用的数据  
**以便于** 统一管理数据，减少分店重复创建工作

---

## 功能需求

### Requirement 1: 菜品标签（ProductLabel）总部来源不可编辑

**用户故事**: 作为店长/运营人员，我想查看和使用总部的菜品标签，但不能编辑来源总部的菜品标签

#### 验收标准

1. **WHEN** 调用菜品标签列表接口 **THEN** 系统 **SHALL** 返回每个标签的 `is_editable` 字段
2. **WHEN** 调用菜品标签详情接口 **THEN** 系统 **SHALL** 返回标签的 `is_editable` 字段
3. **WHEN** 用户尝试编辑来源总部的菜品标签 **THEN** 系统 **SHALL** 返回错误提示 "标签不可编辑"
4. **WHEN** 用户尝试删除来源总部的菜品标签 **THEN** 系统 **SHALL** 返回错误提示 "标签不可删除"
5. **IF** 菜品标签的 `headquarter_uuid` 字段为 0 **THEN** 系统 **SHALL** 返回 `is_editable: true`，允许正常编辑

#### 具体要求

- [ ] 1.1 在 `ProductLabelDetail` 响应结构中添加 `is_editable` 字段
- [ ] 1.2 在 `GetProductLabelList()` 方法中返回 `is_editable` 字段
- [ ] 1.3 在 `EditProductLabel()` 方法中增加总部来源数据校验
- [ ] 1.4 在 `DeleteProductLabel()` 方法中增加总部来源数据校验
- [ ] 1.5 前端列表页根据 `is_editable` 控制编辑按钮
- [ ] 1.6 前端详情页根据 `is_editable` 禁用表单字段

**实现位置**：
- `main/app/service/product_label.go`
- `main/app/dto/resp/product_label.go`

---

### Requirement 2: 满额减（FullReductionActivity）总部来源不可编辑

**用户故事**: 作为店长/运营人员，我想查看和使用总部的满额减活动，但不能编辑来源总部的满额减活动

#### 验收标准

1. **WHEN** 调用满额减列表接口 **THEN** 系统 **SHALL** 返回每个活动的 `is_editable` 字段
2. **WHEN** 调用满额减详情接口 **THEN** 系统 **SHALL** 返回活动的 `is_editable` 字段
3. **WHEN** 用户尝试编辑来源总部的满额减活动 **THEN** 系统 **SHALL** 返回错误提示 "活动不可编辑"
4. **WHEN** 用户尝试删除来源总部的满额减活动 **THEN** 系统 **SHALL** 返回错误提示 "活动不可删除"
5. **IF** 满额减活动的 `headquarter_uuid` 字段为 0 **THEN** 系统 **SHALL** 返回 `is_editable: true`，允许正常编辑

#### 具体要求

- [ ] 2.1 在 `FullReductionActivityResp` 响应结构中添加 `is_editable` 字段
- [ ] 2.2 在满额减列表/详情接口中返回 `is_editable` 字段
- [ ] 2.3 在满额减编辑接口中增加总部来源数据校验
- [ ] 2.4 在满额减删除接口中增加总部来源数据校验
- [ ] 2.5 前端列表页根据 `is_editable` 控制编辑按钮
- [ ] 2.6 前端详情页根据 `is_editable` 禁用表单字段

**实现位置**：
- `main/app/service/full_reduction_activity_srv.go`
- `main/app/dto/resp/full_reduction_activity_resp.go`

---

### Requirement 3: 商品（ProductPackage）总部来源不可编辑（特殊规则）

**用户故事**: 作为店长/运营人员，我想查看和使用总部的商品，但只能修改外卖的价格、上下架，其他字段不可编辑

#### 验收标准

1. **WHEN** 调用商品列表接口 **THEN** 系统 **SHALL** 返回每个商品的 `is_editable` 字段
2. **WHEN** 调用商品详情接口 **THEN** 系统 **SHALL** 返回商品的 `is_editable` 字段
3. **WHEN** 用户尝试编辑来源总部的商品（非外卖价格、上下架字段）**THEN** 系统 **SHALL** 返回错误提示 "商品不可编辑"
4. **WHEN** 用户尝试修改来源总部的商品的外卖价格或上下架状态 **THEN** 系统 **SHALL** 允许修改
5. **WHEN** 用户尝试删除来源总部的商品 **THEN** 系统 **SHALL** 返回错误提示 "商品不可删除"
6. **IF** 商品的 `headquarter_uuid` 字段为 0 **THEN** 系统 **SHALL** 返回 `is_editable: true`，允许正常编辑

#### 具体要求

- [ ] 4.1 确认商品响应结构中已有 `is_editable` 字段（✅ 已存在）
- [ ] 4.2 在商品编辑接口中增加总部来源数据校验（特殊规则：允许修改外卖价格、上下架）
- [ ] 4.3 在商品删除接口中增加总部来源数据校验
- [ ] 4.4 前端列表页根据 `is_editable` 控制编辑按钮
- [ ] 4.5 前端详情页根据 `is_editable` 禁用表单字段（外卖价格、上下架除外）

**实现位置**：
- `main/app/service/product.go`
- `main/app/dto/resp/product_resp/product.go`

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 严格遵循 Controller → Service → Repository 分层 ✅
- **单一职责原则**: 每个文件应有单一、明确的目的 ✅
- **模块化设计**: Service 和 Repository 应独立且可复用 ✅
- **依赖管理**: Service 只能依赖其他 Service 接口，不能直接依赖 Repository ✅
- **代码复用**: 参考商品单位的 `isEditable()` 函数实现，保持一致性 ✅
- **遵循规范**:
  - `.cursor/rules/go-main.mdc` - Go Main 开发规范 ✅
  - `.cursor/rules/vue.mdc` - Vue 前端规范（待前端实现）

### API 设计要求

- [x] URL 使用 snake_case 命名 ✅
- [x] data 字段必须是对象，不能是 null 或数组 ✅
- [x] 分页信息统一放在 meta 中 ✅
- [x] 响应格式：`{code, message, data{}}` ✅
- [x] 参考: `.cursor/rules/api.mdc` - API 设计规范 ✅

### 性能要求

- [x] 本地响应时间 < 200ms ✅
- [x] 数据库查询优化（使用索引）✅
- [ ] 缓存策略（Redis）（如需要）
- [ ] 并发处理（使用 UUID 锁）（如需要）

### 浏览器兼容性（管理后台）

- [ ] Chrome 90+（待前端实现）
- [ ] Safari 14+（待前端实现）
- [ ] Firefox 88+（待前端实现）
- [ ] Edge 90+（待前端实现）

### 测试要求

- [ ] Service 层测试覆盖率 ≥ 70%（待补充测试）
- [ ] Repository 层测试覆盖率 ≥ 80%（待补充测试）
- [ ] API 测试覆盖所有接口（待补充测试）
- [ ] 集成测试覆盖核心流程（待补充测试）
- [ ] 参考: `.cursor/rules/go-main.mdc` - 测试规范

### 国际化要求

- [x] 支持 10 种语言（中文、英文、日语、韩语等）✅
- [x] 所有文案使用多语言实现 ✅
- [ ] 错误提示信息国际化（待确认）
- [ ] 参考: `main/i18n/` - 国际化配置

### 安全要求

- [x] 所有 API 需要身份验证 ✅
- [x] SQL 注入防护（使用参数化查询）✅
- [ ] XSS 防护（前端输入校验）（待前端实现）
- [ ] CSRF 防护（Token 验证）（待前端实现）
- [ ] 参考: `.cursor/rules/security.mdc` - 安全开发规范

### 可靠性要求

- [x] 网络异常时优雅降级 ✅
- [x] 事务管理（保证数据一致性）✅
- [x] 错误日志记录（使用 Logger）✅
- [x] 故障恢复机制 ✅

---

## 验收标准

### 功能验收

1. **菜品标签总部来源不可编辑**: ⏳ 待实现
   - 列表/详情接口返回 `is_editable` 字段
   - 编辑/删除接口拒绝总部来源数据
   - 前端 UI 控制

2. **满额减总部来源不可编辑**: ⏳ 待实现
   - 列表/详情接口返回 `is_editable` 字段
   - 编辑/删除接口拒绝总部来源数据
   - 前端 UI 控制

3. **商品总部来源不可编辑（特殊规则）**: ⏳ 待实现
   - 列表/详情接口返回 `is_editable` 字段（✅ 已存在）
   - 编辑接口拒绝总部来源数据（允许修改外卖价格、上下架）
   - 删除接口拒绝总部来源数据
   - 前端 UI 控制

### 测试验收

1. **单元测试**: ⏳ 待补充
   - Service 层测试覆盖率 ≥ 70%
   - Repository 层测试覆盖率 ≥ 80%

2. **API 测试**: ⏳ 待补充
   - 所有接口测试通过
   - 边界场景测试通过

3. **集成测试**: ⏳ 待补充
   - 端到端流程测试通过
   - 同步功能测试通过

4. **手动测试**: ⏳ 待前端实现
   - 浏览器兼容性测试通过
   - UI 交互测试通过

### 文档验收

1. **技术文档**: ✅ 已完成
   - requirements.md 完整且准确

2. **API 文档**: ⏳ 待更新
   - API 接口文档完整（Swagger）

3. **代码注释**: ⏳ 待补充
   - 关键函数有注释说明

---

## 约束条件

### 技术约束

#### Go Main 模块

- ✅ 必须使用 Gin 框架
- ✅ 接口以 `I` 开头，实现以 `Impl` 结尾
- ✅ Service 只能依赖其他 Service 接口
- ✅ Repository 只能持有 db 实例，不能持有 DBManager
- ✅ 不使用 panic，返回 error
- ✅ 参考商品单位的实现方式，保持代码一致性

#### Vue 模块

- [ ] 必须使用 Vue 3 + TypeScript + Vite（待前端实现）
- [ ] 使用 Element Plus 组件库（待前端实现）
- [ ] 遵循 `.cursor/rules/vue.mdc`（待前端实现）

### 业务约束

- **数据一致性**：总部数据在分店中必须保持只读，不能修改（商品有特殊规则）
- **用户体验**：必须明确告知用户为什么不能编辑
- **兼容性**：必须与商品单位（ProductUnit）模块保持一致的用户体验

### 资源约束

- 开发时间: 3.5-4.5 天（3个模块，每个模块约1-1.5天）
- Story Point: 3 SP（参考商品单位的实现方式，工作量相对可控）

---

## 依赖关系

### 技术依赖

- **参考实现**：`main/app/service/product.go` - ProductUnit 相关方法
- **判断函数**：`main/app/service/product.go` - `isEditable(ctx, headquarterUuid)`

### 服务依赖

- **Frontend → Main**: HTTP API 调用（各模块的列表/详情/编辑/删除接口）

### 业务依赖

- **总部-分店颗粒化同步功能**：确保同步后的数据正确标记 `headquarter_uuid`
- **参考实现**：商品单位（ProductUnit）模块的完整实现

---

## 风险和缓解

### 风险 1: 实现不一致

**影响**: 中  
**概率**: 中  
**缓解措施**:

- 严格参考商品单位的实现方式
- 使用统一的 `isEditable()` 函数
- 代码审查时检查一致性

### 风险 2: 前端未实现 UI 控制

**影响**: 中  
**概率**: 低  
**缓解措施**:

- 明确标注前端需要实现的功能点
- 参考商品单位的前端实现
- 提供详细的 UI 交互说明

### 风险 3: 特殊规则理解偏差

**影响**: 高  
**概率**: 中  
**缓解措施**:

- 商品：明确只能修改外卖价格、上下架
- 与产品确认特殊规则的具体实现方式

---

## 时间表

- **Phase 1 - 菜品标签实现**: 1 天
  - 后端 API 实现
  - 前端 UI 实现

- **Phase 2 - 满额减实现**: 1 天
  - 后端 API 实现
  - 前端 UI 实现

- **Phase 3 - 商品实现（特殊规则）**: 1.5 天
  - 后端 API 实现（特殊规则处理）
  - 前端 UI 实现

- **Phase 4 - 测试和优化**: 1 天
  - 单元测试补充
  - API 测试补充
  - 集成测试
  - UI 交互测试

- **总计**: 4.5 天（SP = 3）

---

## 参考资料

### 核心规范

- `.cursor/rules/go-main.mdc` - Go Main 核心约束
- `.cursor/rules/vue.mdc` - Vue 开发规范
- `.cursor/rules/api.mdc` - API 设计规范
- `.cursor/rules/security.mdc` - 安全开发规范

### 参考实现

- **商品单位（ProductUnit）**：
  - Service 实现：`main/app/service/product.go`
    - `GetProductUnitList()` - 第2006行
    - `GetProductUnit()` - 第2044行
    - `EditProductUnit()` - 第2260行
    - `DeleteProductUnit()` - 第2375行
    - `isEditable()` - 第2634行
  - 响应结构：`main/app/dto/resp/product_resp/product.go`
    - `ProductUnitItem` - 第225行
    - `ProductUnitDetail` - 第248行

### 相关文档

- **总部-分店颗粒化同步**：`docs/shared/specs/active/shop-headquarters-branch-granular-sync-backend/`
### 代码位置

- **菜品标签**：
  - Service：`main/app/service/product_label.go`
  - 响应结构：`main/app/dto/resp/product_label.go`

- **满额减**：
  - Service：`main/app/service/full_reduction_activity_srv.go`
  - 响应结构：`main/app/dto/resp/full_reduction_activity_resp.go`

- **商品**：
  - Service：`main/app/service/product.go`
  - 响应结构：`main/app/dto/resp/product_resp/product.go`

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{user}/{YYYY-MM}/{YYYY-MM-DD}.md`
- 提醒：需求评审或范围调整若形成经验，应同步更新 Episode 并在 Proposal/Spec 互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-12-08  
**作者**: 曾振华  
**审核者**: {审核者}
