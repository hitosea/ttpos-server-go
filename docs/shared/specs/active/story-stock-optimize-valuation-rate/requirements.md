# 优化库存盘点估值率逻辑 需求文档

> 本文档定义 优化库存盘点估值率逻辑 的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2025-12/v2.12.0-optimize-stock-valuation-rate.md](../../../../team/proposals/2025-12/v2.12.0-optimize-stock-valuation-rate.md) |
| **创建日期**      | 2025-12-23                                                                                                 |
| **负责人**        | {姓名}                                                                                                       |
| **目标 Sprint**   | v2.12.0                                                                                                      |
| **涉及技术栈**    | [ ] Go (main/) [x] Go (ttpos-bmp/) [ ] PHP (admin/) [ ] Vue (admin/views/)                                   |
| **关联任务**      | DooTask #37893                                                                                               |

## 📋 审核状态

| 项目         | 内容                     |
| ------------ | ------------------------ |
| **审核状态** | 待审核 |
| **审核人**   | {审核人姓名}             |
| **审核日期** | {YYYY-MM-DD}             |
| **审核意见** | {审核意见，如有}         |

---

## 📋 概述

当前 TTPOS 系统在创建库存盘点时，若物品估值率为空，会强制赋值为 1，导致覆盖 ERPNext 中该物品的真实估值率（来自采购数据），直接影响财务计算和成本核算。

**本次需求范围**：仅优化 ttpos-erp 后端模块（ttpos-bmp 项目），不涉及前端改动。

核心改进包括：
1. 新增从 ERPNext Bin 表查询物品估值率的 gRPC 服务
2. 优化盘点保存逻辑，移除强制赋值为 1 的行为，改为从 Bin 表读取真实估值率
3. 在盘点提交时验证估值率有效性，若为空则阻止提交并提示明确错误

通过后端逻辑优化，确保估值率数据来源的准确性，规范库存盘点业务流程。

## 🎯 产品对齐

该功能优化支持 TTPOS 产品的核心定位：**收银 + ERP + 会员 + 外卖**中的 ERP 模块（库存管理），通过规范库存盘点流程，确保成本核算的准确性，为商户提供可靠的财务数据基础。

## 📝 用户故事

**作为** 仓库管理员  
**我想** 在进行库存盘点时，系统能够自动读取物品的真实估值率（来自采购数据）  
**以便于** 确保盘点结果准确反映物品的实际成本，不会因为系统的错误赋值导致财务数据失真

---

## 功能需求

### Requirement 1: 新增 Bin 查询服务（ERPNext Bin 表）

**用户故事**: 作为系统开发者，我需要实现查询 ERPNext Bin 表的 gRPC 服务，以便盘点时能够获取物品的真实估值率

#### 验收标准

1. **WHEN** 调用 Bin 查询服务并传入 `item_code` 和 `warehouse` **THEN** 系统 **SHALL** 返回该物品在指定仓库的估值率（Valuation Rate）
2. **IF** 物品在指定仓库中不存在 Bin 记录 **THEN** 系统 **SHALL** 返回空结果或估值率为 0
3. **WHEN** 查询 Bin 记录 **THEN** 系统 **SHALL** 返回字段包括：`item_code`, `warehouse`, `actual_qty`, `valuation_rate`, `stock_value` 等
4. **IF** ERPNext 服务不可用或查询失败 **THEN** 系统 **SHALL** 返回明确的错误信息

#### 具体要求

- [ ] 1.1 定义 Protobuf 消息：`GetBinReq`（包含 `item_code`, `warehouse`, `company_abbr`）和 `GetBinResp`（包含 Bin 记录详情）
- [ ] 1.2 在 `ttpos-bmp/app/ttpos-erp` 模块中实现 Bin 查询 gRPC 服务
- [ ] 1.3 调用 ERPNext Document API 查询 Bin DocType（`frappe.client.get_list` 或 `frappe.client.get`）
- [ ] 1.4 使用 `service.Document().List()` 或 `service.Document().Get()` 方法查询 Bin 数据
- [ ] 1.5 错误处理：ERPNext API 调用失败时返回明确错误信息
- [ ] 1.6 日志记录：记录查询参数和结果，便于调试

---

### Requirement 2: 调整盘点保存逻辑（SaveStockReconciliation）

**用户故事**: 作为仓库管理员，当我保存盘点单时，系统应该使用物品的真实估值率，而非强制赋值为 1

#### 验收标准

1. **WHEN** 保存盘点单时，物品的 `ValuationRate` 为 0 **THEN** 系统 **SHALL** 调用 Bin 查询服务获取真实估值率
2. **IF** Bin 查询服务返回的估值率 > 0 **THEN** 系统 **SHALL** 使用该估值率作为盘点物品的估值率
3. **IF** Bin 查询服务返回的估值率 = 0 或无记录 **THEN** 系统 **SHALL** 仍使用保底估值率 1（保持当前逻辑，在提交时再验证）
4. **WHEN** 保存盘点单成功 **THEN** 系统 **SHALL** 返回盘点单号
5. **IF** 盘点物品数量与当前仓库库存一致 **THEN** 系统 **SHALL** 跳过该物品（当前逻辑保持不变）

#### 具体要求

- [ ] 2.1 在 `SaveStockReconciliation` 方法中，检查 `item.ValuationRate` 是否为 0
- [ ] 2.2 若为 0，调用新增的 Bin 查询服务（传入 `item_code`, `warehouse`, `company_abbr`）
- [ ] 2.3 若 Bin 查询返回估值率 > 0，使用该估值率；否则使用保底估值率 1
- [ ] 2.4 移除当前代码中第 109 行的强制赋值逻辑：`itemData.ValuationRate = consts.DefaultValuationRate`
- [ ] 2.5 增加日志：记录估值率来源（Bin 查询或保底值）
- [ ] 2.6 保持现有逻辑：若物品数量与库存一致，跳过该物品

---

### Requirement 3: 盘点提交时验证估值率（SubmitStockReconciliation）

**用户故事**: 作为仓库管理员，当我提交盘点单时，系统应该验证所有物品的估值率，确保数据准确性

#### 验收标准

1. **WHEN** 提交盘点单时 **THEN** 系统 **SHALL** 验证盘点单中所有物品的估值率
2. **IF** 存在估值率为 0 或默认值 1 的物品 **THEN** 系统 **SHALL** 阻止提交并返回错误信息
3. **WHEN** 估值率验证失败 **THEN** 系统 **SHALL** 返回错误提示："xx物品库存物品估值率为空，无法提交盘点单。请先通过采购入库建立库存。"
4. **IF** 所有物品估值率有效 **THEN** 系统 **SHALL** 正常提交盘点单到 ERPNext

#### 具体要求

- [ ] 3.1 在 `SubmitStockReconciliation` 方法中，提交前查询盘点单详情
- [ ] 3.2 遍历盘点单中的所有物品，检查估值率是否为 0 或 1
- [ ] 3.3 若存在无效估值率，返回错误信息（包含物品代码和名称）
- [ ] 3.4 错误提示格式：`"物品 [物品代码] 在仓库 [仓库名] 的估值率为空，无法提交盘点单。请先通过采购入库建立库存。"`
- [ ] 3.5 若验证通过，正常调用 `service.Document().ChangeDocStatus()` 提交盘点单


---

## 📌 未来迭代

以下功能不在本次需求范围内，将在后续版本中实现：

### 前端优化（待规划）

1. **移除物品管理页面的估值率字段**
   - 物品添加页面移除估值率输入框
   - 物品查看/编辑页面移除估值率显示

2. **盘点页面增强验证提示**
   - 修改仓库时提示物品不在仓库中
   - 提供"清空当前物品"和"我知道了"选项

3. **物品源过滤优化**
   - 盘点页面只显示选定仓库的物品
   - 实时更新物品源列表

> **说明**: 本次需求仅优化后端逻辑，前端改动将在后续迭代中结合整体 UI/UX 优化一并实现。

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 严格遵循 Controller → Logic → Service 分层（ttpos-bmp 模块）
- **单一职责原则**: 每个文件应有单一、明确的目的
- **模块化设计**: Logic 和 Service 应独立且可复用
- **依赖管理**: Logic 只能依赖 Service 接口，不能直接操作数据库
- **遵循规范**:
  - `.cursor/rules/go-bmp.mdc` - Go BMP 开发规范
  - `.cursor/rules/structs.mdc` - 项目结构规范

### API 设计要求

- [x] gRPC 接口遵循 Protobuf 规范
- [x] 错误码统一使用 gRPC Status Code
- [x] 响应消息包含详细的错误信息
- [x] 参考: `.cursor/rules/go-bmp.mdc` - BMP API 设计规范

### 数据库设计要求

- [x] 本需求不涉及数据库表结构变更
- [x] 查询 ERPNext Bin 表通过 Document API，无需本地表
- [x] 参考: `.cursor/rules/database.mdc` - 数据库开发规范

### 性能要求

- [ ] Bin 查询响应时间 < 500ms（依赖 ERPNext 性能）
- [ ] 盘点保存性能：批量查询 Bin 数据时考虑并发控制
- [ ] 缓存策略：Bin 数据可短期缓存（5 分钟），减少 ERPNext 查询压力
- [ ] 错误重试：ERPNext 查询失败时重试 1 次

### 浏览器兼容性（管理后台）

- [ ] 不适用（本次需求仅涉及后端 ttpos-erp 模块）

### 测试要求

- [ ] Logic 层测试覆盖率 ≥ 70%
- [ ] Service 层测试覆盖率 ≥ 80%
- [x] **Stock 相关模块测试覆盖率 100%**（高风险）
- [ ] 集成测试覆盖核心流程（盘点保存、提交、Bin 查询）
- [ ] 参考: `.cursor/rules/go-bmp.mdc` - 测试规范

### 国际化要求

- [ ] 支持 10 种语言（中文、英文、日语、韩语等）
- [ ] 错误提示信息使用多语言实现
- [ ] 参考: `ttpos-bmp/i18n/` - 国际化配置

### 安全要求

- [x] 所有 gRPC 接口需要身份验证（通过 token 或证书）
- [x] 参数验证：防止 SQL 注入、XSS 攻击
- [x] 敏感数据（估值率）仅对有权限的用户可见
- [x] 参考: `.cursor/rules/security.mdc` - 安全开发规范

### 可靠性要求

- [ ] ERPNext 服务不可用时优雅降级（使用保底估值率 1，并记录警告日志）
- [ ] 事务管理：盘点单提交失败时回滚状态
- [ ] 错误日志记录：记录所有 Bin 查询失败和估值率异常情况
- [ ] 故障恢复机制：支持重新提交盘点单

---

## 验收标准

### 功能验收

1. **Bin 查询服务**: 能够正确查询 ERPNext Bin 表，返回估值率数据
2. **盘点保存优化**: 不再强制赋值估值率为 1，使用 Bin 查询结果
3. **盘点提交验证**: 估值率为空时阻止提交，返回明确错误信息

### 测试验收

1. **单元测试**: 覆盖率达标（Logic ≥ 70%, Service ≥ 80%, Stock 相关 100%）
2. **集成测试**: 端到端流程测试通过（保存盘点 → 提交盘点）
3. **回归测试**: 确保现有盘点功能不受影响
4. **错误场景测试**: 覆盖 ERPNext 不可用、估值率为空等异常情况

### 文档验收

1. **技术文档**: design.md 完整且准确
2. **API 文档**: Protobuf 定义和 gRPC 服务文档完整
3. **测试文档**: tasks.md 中的测试任务完成

---

## 约束条件

### 技术约束

#### Go BMP 模块

- 必须使用 GoFrame 2.x 框架
- gRPC 服务使用 Protobuf 定义
- Logic 层只能依赖 Service 接口
- Service 层封装 ERPNext Document API 调用
- 不使用 panic，返回 error

### 业务约束

- 估值率必须来自 ERPNext Bin 表（采购入库生成）
- 不允许在 TTPOS 侧手动维护估值率
- 盘点物品必须在选定仓库中有库存记录
- 估值率为空的物品不允许提交盘点单

### 资源约束

- 开发时间: 2-3 天（仅后端 ttpos-erp 模块）
- Story Point: 3 (必须 ≤ 5)

---

## 依赖关系

### 技术依赖

- `ttpos-bmp/app/ttpos-erp/internal/logic/stock/stock_reconciliation.go` - 盘点逻辑
- `ttpos-bmp/app/ttpos-erp/internal/service/document.go` - ERPNext Document API
- `ttpos-bmp/app/ttpos-erp/api/stock/` - Stock gRPC 接口定义

### 服务依赖

- **ttpos-bmp → ERPNext**: Document API 调用（查询 Bin 表）

### 业务依赖

- ERPNext Bin 表存在且可访问
- 物品必须通过采购入库建立 Bin 记录
- 盘点单与 Stock Reconciliation DocType 关联

---

## 风险和缓解

### 风险 1: ERPNext Bin 查询性能问题

**影响**: 中  
**概率**: 中  
**缓解措施**:

- 实现缓存机制，减少重复查询
- 批量查询时使用并发控制，避免过多请求
- 设置合理的超时时间（5 秒）
- ERPNext 不可用时使用保底估值率 1，记录警告日志

### 风险 2: 历史数据估值率错误

**影响**: 高  
**概率**: 高  
**缓解措施**:

- 提供数据修复脚本或工具
- 在提交盘点时进行严格验证，阻止错误数据
- 提供用户培训和文档，说明正确流程
- 记录详细日志，便于追溯和修复

### 风险 3: 用户流程变更导致使用困难

**影响**: 中  
**概率**: 中  
**缓解措施**:

- 提供清晰的错误提示和操作指引
- 编写用户操作文档和培训材料
- 在 UI 中增加帮助提示和引导
- 收集用户反馈，持续优化体验

---

## 时间表

- **Phase 1 - 后端 Bin 查询服务**: 0.5 天
- **Phase 2 - 盘点保存逻辑优化**: 0.5 天
- **Phase 3 - 盘点提交验证逻辑**: 0.5 天
- **Phase 4 - 测试与优化**: 1 天
- **总计**: 2-3 天（SP = 3）

---

## 参考资料

### 核心规范

- `.cursor/rules/go-bmp.mdc` - Go BMP 开发规范
- `.cursor/rules/database.mdc` - 数据库开发规范
- `.cursor/rules/security.mdc` - 安全开发规范

### 架构文档

- `docs/human/architecture/erp/stock_reconciliation.md` - 库存盘点架构文档（如存在）
- ERPNext Bin DocType: https://github.com/frappe/erpnext/blob/develop/erpnext/stock/doctype/bin/bin.json

### 开发指南

- `docs/human/guides/go-bmp-development.md` - Go BMP 开发指南

### 外部参考

- ERPNext Stock Reconciliation: https://github.com/frappe/erpnext/blob/develop/erpnext/stock/doctype/stock_reconciliation/stock_reconciliation.json
- Frappe Client API: https://frappeframework.com/docs/user/en/api/rest

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 提醒：需求评审或范围调整若形成经验，应同步更新 Episode 并在 Proposal/Spec 互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-12-23  
**作者**: rikugun  
**审核者**: {审核者}

