# 订单数据导入 需求文档

> 本文档定义订单数据导入功能的详细需求和验收标准。
>
> **💡 MVP 方案**：最小可执行版本，快速验证可行性

## 📋 基本信息

| 项目              | 内容                                                                                                          |
| ----------------- | ------------------------------------------------------------------------------------------------------------- |
| **来源 Proposal** | [docs/team/proposals/2025-11/import-orders-data.md](../../../../team/proposals/2025-11/import-orders-data.md) |
| **创建日期**      | 2025-11-19                                                                                                    |
| **负责人**        | xiezhihuan                                                                                                    |
| **目标 Sprint**   | 待定                                                                                                          |
| **涉及技术栈**    | [x] Go (main/) [ ] Go (ttpos-bmp/) [x] PHP (admin/) [x] Vue (admin/views/)                                    |

---

## 📋 概述

华莱士从旧 POS 系统迁移到 TTPOS，需要将历史订单数据批量导入。本功能提供简单的 Excel 文件上传和导入能力，支持将旧系统的订单数据快速迁移到新系统。

**MVP 核心思路**：华莱士提供标准 Excel → Shop 管理端上传 → 解析写入数据库 → 返回结果

## 🎯 产品对齐

- 支持数据迁移，保证业务连续性
- 快速验证可行性，后续可扩展

## 📝 用户故事

**作为** 系统管理员  
**我想** 将华莱士旧系统的历史订单数据批量导入到 TTPOS 系统  
**以便于** 保证数据的连续性和完整性，支持历史数据查询和分析

---

## 功能需求

### Requirement 1: Excel 文件上传

**用户故事**: 作为系统管理员，我想上传 Excel 文件，以便于导入订单数据

#### 验收标准

1. **WHEN** 管理员在 Shop 管理端选择 Excel 文件 **THEN** 系统 **SHALL** 接收文件并开始解析
2. **WHEN** 上传的文件格式不是 `.xlsx` **THEN** 系统 **SHALL** 提示格式错误
3. **WHEN** 上传的文件大小超过 10MB **THEN** 系统 **SHALL** 提示文件过大

#### 具体要求

- [ ] 1.1 在 Shop 管理端添加"导入订单"页面
- [ ] 1.2 提供文件上传组件，支持拖拽和点击上传
- [ ] 1.3 限制文件格式为 `.xlsx`，大小限制 10MB
- [ ] 1.4 上传时显示加载状态

---

### Requirement 2: Excel 数据解析和校验

**用户故事**: 作为系统管理员，我想系统自动解析和校验 Excel 数据，以便于确保数据正确性

#### 验收标准

1. **WHEN** Excel 文件上传成功 **THEN** 系统 **SHALL** 解析文件内容
2. **WHEN** Excel 格式不符合约定 **THEN** 系统 **SHALL** 提示格式错误
3. **WHEN** 必填字段缺失 **THEN** 系统 **SHALL** 提示具体缺失字段
4. **WHEN** 关联数据不存在（门店、商品） **THEN** 系统 **SHALL** 提示具体错误信息

#### 具体要求

- [ ] 2.1 使用 `excelize` 库解析 Excel 文件
- [ ] 2.2 校验必填字段：订单号、下单时间、订单金额、门店、商品
- [ ] 2.3 校验数据格式：日期格式、金额格式
- [ ] 2.4 校验关联数据：门店是否存在、商品是否存在
- [ ] 2.5 单次导入限制 5000 条，超过建议分批

---

### Requirement 3: 订单数据导入

**用户故事**: 作为系统管理员，我想将解析后的订单数据写入数据库，以便于完成数据迁移

#### 验收标准

1. **WHEN** 数据校验通过 **THEN** 系统 **SHALL** 开始导入数据
2. **WHEN** 订单号已存在 **THEN** 系统 **SHALL** 跳过该订单并记录
3. **WHEN** 导入过程中发生错误 **THEN** 系统 **SHALL** 回滚事务，不写入任何数据
4. **WHEN** 导入完成 **THEN** 系统 **SHALL** 返回成功数量和失败详情

#### 具体要求

- [ ] 3.1 使用事务保证数据一致性
- [ ] 3.2 批量插入数据（每批 500 条）
- [ ] 3.3 订单号重复则跳过，不覆盖
- [ ] 3.4 记录导入结果：成功数量、失败数量、失败原因

---

### Requirement 4: 导入结果反馈

**用户故事**: 作为系统管理员，我想查看导入结果，以便于了解导入成功和失败情况

#### 验收标准

1. **WHEN** 导入完成 **THEN** 系统 **SHALL** 显示成功数量和失败数量
2. **WHEN** 有失败记录 **THEN** 系统 **SHALL** 显示失败原因（简要信息）
3. **WHEN** 导入失败 **THEN** 系统 **SHALL** 显示错误提示

#### 具体要求

- [ ] 4.1 前端显示导入结果：成功 X 条，失败 Y 条
- [ ] 4.2 显示失败原因列表（简要信息，不超过 10 条）
- [ ] 4.3 导入完成后可重新上传

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 严格遵循 Controller → Service → Repository 分层
- **单一职责原则**: 每个文件应有单一、明确的目的
- **模块化设计**: Service 和 Repository 应独立且可复用
- **依赖管理**: Service 只能依赖其他 Service 接口，不能直接依赖 Repository
- **遵循规范**:
  - `.cursor/rules/go-main.mdc` - Go Main 开发规范
  - `.cursor/rules/php.mdc` - PHP 开发规范
  - `.cursor/rules/vue.mdc` - Vue 前端规范

### API 设计要求

- [x] URL 使用 snake_case 命名（如：`/api/v1/order/import`）
- [x] data 字段必须是对象，不能是 null 或数组
- [x] 响应格式：`{code, message, data{}}`
- [x] 参考: `.cursor/rules/api.mdc` - API 设计规范

### 数据库设计要求

- [x] 使用现有订单表结构（不新增表）
- [x] 遵循现有数据库规范
- [x] 参考: `.cursor/rules/database.mdc` - 数据库开发规范

### 性能要求

- [ ] 单次导入 5000 条数据响应时间 < 30 秒
- [ ] 批量插入优化（每批 500 条）
- [ ] 导入过程不影响其他用户正常使用

### 浏览器兼容性（管理后台）

- [ ] Chrome 90+
- [ ] Safari 14+
- [ ] Firefox 88+
- [ ] Edge 90+

### 测试要求

- [ ] Service 层测试覆盖率 ≥ 70%
- [ ] Repository 层测试覆盖率 ≥ 80%
- [ ] **Order 相关模块测试覆盖率 100%**（高风险）
- [ ] 集成测试覆盖核心流程
- [ ] API 测试覆盖所有接口
- [ ] 参考: `.cursor/rules/go-main.mdc` - 测试规范

### 安全要求

- [ ] 所有 API 需要身份验证
- [ ] 文件上传大小限制
- [ ] SQL 注入防护（使用参数化查询）
- [ ] 参考: `.cursor/rules/security.mdc` - 安全开发规范

### 可靠性要求

- [ ] 网络异常时优雅降级
- [ ] 事务管理（保证数据一致性）
- [ ] 错误日志记录（使用 Logger）
- [ ] 导入失败时数据回滚

---

## 验收标准

### 功能验收

1. **文件上传**: 支持上传 `.xlsx` 文件，大小限制 10MB
2. **数据解析**: 成功解析 Excel 文件，提取订单数据
3. **数据校验**: 校验必填字段和关联数据，给出明确错误提示
4. **数据导入**: 成功将数据写入数据库，订单号重复则跳过
5. **结果反馈**: 显示导入成功数量和失败原因

### 测试验收

1. **单元测试**: 覆盖率达标（Service ≥ 70%, Repository ≥ 80%, Order 100%）
2. **API 测试**: 所有接口测试通过
3. **集成测试**: 端到端流程测试通过
4. **手动测试**: 浏览器兼容性测试通过

### 文档验收

1. **技术文档**: design.md 完整且准确
2. **API 文档**: API 接口文档完整
3. **测试文档**: tasks.md 中的测试任务完成

---

## 约束条件

### 技术约束

#### Go Main 模块

- 必须使用 Gin 框架
- 接口以 `I` 开头，实现以 `Impl` 结尾
- Service 只能依赖其他 Service 接口
- Repository 只能持有 db 实例，不能持有 DBManager
- 不使用 panic，返回 error

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

- Excel 格式由华莱士按约定准备（不提供模板下载）
- 单次导入限制 5000 条
- 订单号重复则跳过，不覆盖

### 资源约束

- 开发时间: 2-3 天
- Story Point: 3 SP (必须 ≤ 5)

---

## 依赖关系

### 技术依赖

- `github.com/xuri/excelize/v2` - Excel 文件处理
- 现有订单相关 Service 和 Repository

### 服务依赖

- **Admin → Main**: HTTP API 调用
- **Frontend → Admin**: HTTP API 调用

### 业务依赖

- 订单表结构已存在
- 门店、商品数据已存在

---

## 风险和缓解

### 风险 1: Excel 格式不规范

**影响**: 中  
**概率**: 中  
**缓解措施**:

- 前置沟通，明确 Excel 格式规范
- 严格的数据校验，给出明确错误提示
- 小批量测试，确认无误后再批量导入

### 风险 2: 数据量过大导致超时

**影响**: 中  
**概率**: 低  
**缓解措施**:

- 单次限制 5000 条
- 批量插入优化（每批 500 条）
- 超过限制建议分批导入

### 风险 3: 关联数据缺失

**影响**: 高  
**概率**: 中  
**缓解措施**:

- 严格校验关联数据存在性
- 失败时给出明确提示
- 建议先导入基础数据（门店、商品）

---

## 时间表

- **Phase 1 - 需求确认和格式定义**: 0.5 天
- **Phase 2 - 后端实现**: 1 天
- **Phase 3 - 前端实现**: 0.5 天
- **Phase 4 - 测试和修复**: 0.5 天
- **总计**: 2.5 天（SP = 3）

---

## 参考资料

### 核心规范

- `.cursor/rules/go-main.mdc` - Go Main 核心约束
- `.cursor/rules/php.mdc` - PHP 核心约束
- `.cursor/rules/vue.mdc` - Vue 开发规范
- `.cursor/rules/api.mdc` - API 设计规范
- `.cursor/rules/database.mdc` - 数据库开发规范
- `.cursor/rules/security.mdc` - 安全开发规范

### 参考实现

- 商品导入功能: `admin/views/shop/src/views/product/store/product/importProduct.vue`
- 桌台导入功能: `admin/views/shop/src/views/supplier/table/table/importQrcode.vue`

### 外部参考

- Excelize 文档: https://xuri.me/excelize/

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 提醒：需求评审或范围调整若形成经验，应同步更新 Episode 并在 Proposal/Spec 互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-11-19  
**作者**: xiezhihuan  
**审核者**: 待定
