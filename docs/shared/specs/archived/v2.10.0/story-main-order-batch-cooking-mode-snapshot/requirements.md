> ⚠️ **已归档** - 此 Spec 已随 v2.10.0 发布。
>
> - 归档时间: 2025-12-05
> - 归档人: weifashi

# 订单分批送厨模式快照 需求文档

> 本文档定义订单分批送厨模式快照功能的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                                                    |
| ----------------- | --------------------------------------------------------------------------------------------------------------------------------------- |
| **来源 Proposal** | [docs/team/proposals/2025-12/order-batch-cooking-mode-snapshot.md](../../../../team/proposals/2025-12/order-batch-cooking-mode-snapshot.md) |
| **创建日期**      | 2025-12-02                                                                                                                              |
| **负责人**        | xiezhihuan                                                                                                                               |
| **目标 Sprint**   | Sprint {N}                                                                                                                               |
| **涉及技术栈**    | [x] Go (main/) [ ] Go (ttpos-bmp/) [ ] PHP (admin/) [ ] Vue (admin/views/)                                                               |

## 📋 审核状态

| 项目         | 内容                     |
| ------------ | ------------------------ |
| **审核状态** | 待审核                   |
| **审核人**   | {审核人姓名}             |
| **审核日期** | {YYYY-MM-DD}             |
| **审核意见** | {审核意见，如有}         |

---

## 📋 概述

订单分批送厨模式快照功能旨在解决当前分批送厨模式存储在全局配置中，导致 shop 端修改配置后影响已创建订单的问题。通过在 `sale_bill_setting` 表中记录订单创建时的分批送厨模式，确保订单在整个生命周期内行为一致，不受后续全局配置变更影响。

本功能主要涉及数据库字段扩展、订单创建逻辑修改、送厨逻辑修改和数据迁移。

## 🎯 产品对齐

该功能支持以下业务目标：

- **保证订单一致性**：订单创建时的配置快照，确保订单在整个生命周期内行为一致
- **提升可追溯性**：可以准确查询历史订单创建时的分批送厨模式
- **避免业务风险**：防止因配置变更导致已创建订单的行为异常
- **符合业务逻辑**：订单配置应该在创建时确定，后续不应受全局配置影响

## 📝 用户故事

**作为** 商户管理员  
**我想** 在 shop 端修改分批送厨模式时，已创建的订单保持创建时的模式不变  
**以便于** 确保订单行为一致，避免因配置变更导致已创建订单的行为异常

**作为** 收银员/厨房人员  
**我想** 订单的送厨行为在创建后保持稳定  
**以便于** 按照预期的送厨模式进行工作，不会因配置变更而混乱

---

## 功能需求

### Requirement 1: 数据库字段扩展

**用户故事**: 作为系统，我想在 `sale_bill_setting` 表中记录订单创建时的分批送厨模式，以便于订单后续使用该快照值

#### 验收标准

1. **WHEN** 创建新订单 **THEN** 系统 **SHALL** 在 `sale_bill_setting` 表中保存 `batch_cooking_mode` 字段
2. **IF** `batch_cooking_mode` 字段为空 **THEN** 系统 **SHALL** 使用默认值 "post"
3. **WHEN** 查询订单详情 **THEN** 系统 **SHALL** 返回该订单创建时的 `batch_cooking_mode` 值

#### 具体要求

- [ ] 1.1 在 `ttpos_sale_bill_setting` 表中新增 `batch_cooking_mode` 字段（VARCHAR(10)，默认值 'post'）
- [ ] 1.2 更新 `SaleBillSetting` 模型，添加 `BatchCookingMode` 字段
- [ ] 1.3 数据库迁移脚本编写和测试
- [ ] 1.4 为历史订单的 `batch_cooking_mode` 字段设置默认值 "post"

---

### Requirement 2: 订单创建逻辑修改

**用户故事**: 作为系统，我想在创建订单时从 `business_setting` 读取分批送厨模式并保存到 `sale_bill_setting`，以便于记录订单创建时的配置

#### 验收标准

1. **WHEN** 创建新订单 **THEN** 系统 **SHALL** 将当前 `business_setting.BatchCookingMode` 的值保存到 `sale_bill_setting.batch_cooking_mode`
2. **IF** `business_setting.BatchCookingMode` 为空 **THEN** 系统 **SHALL** 使用默认值 "post"
3. **WHEN** 订单创建成功 **THEN** 系统 **SHALL** 确保 `sale_bill_setting.batch_cooking_mode` 已正确保存

#### 具体要求

- [ ] 2.1 修改 `NewSaleBillSetting` 函数，从 `businessSetting.BatchCookingMode` 读取值
- [ ] 2.2 如果值为空，使用默认值 "post"
- [ ] 2.3 保存到 `saleBillSetting.BatchCookingMode` 字段
- [ ] 2.4 确保所有订单创建路径都调用该逻辑（POS、Assistant、会员端等）

---

### Requirement 3: 送厨逻辑修改

**用户故事**: 作为系统，我想在订单送厨时使用 `sale_bill_setting.batch_cooking_mode` 的值，以便于已创建订单不受全局配置变更影响

#### 验收标准

1. **WHEN** 订单送厨时 **THEN** 系统 **SHALL** 使用 `sale_bill_setting.batch_cooking_mode` 的值，而不是 `business_setting.BatchCookingMode`
2. **IF** shop 端修改了 `business_setting.BatchCookingMode` **THEN** 已创建的订单 **SHALL** 仍使用创建时保存的 `batch_cooking_mode` 值
3. **IF** 历史订单的 `batch_cooking_mode` 为空 **THEN** 系统 **SHALL** 使用默认值 "post"

#### 具体要求

- [ ] 3.1 全面查找所有使用 `business_setting.BatchCookingMode` 的地方
- [ ] 3.2 修改送厨相关逻辑，从 `sale_bill_setting.batch_cooking_mode` 读取
- [ ] 3.3 确保 POS、Assistant、KDS 等终端的送厨逻辑都使用快照值
- [ ] 3.4 添加兼容逻辑：如果 `batch_cooking_mode` 为空，使用默认值 "post"

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 严格遵循 Controller → Service → Repository 分层
- **单一职责原则**: `NewSaleBillSetting` 函数专注订单设置创建逻辑
- **模块化设计**: 送厨逻辑可独立测试和复用
- **依赖管理**: Service 只能依赖其他 Service 接口
- **遵循规范**:
  - `.cursor/rules/go-main.mdc` - Go Main 开发规范
  - `.cursor/rules/database.mdc` - 数据库开发规范
  - `.cursor/rules/structs.mdc` - 项目结构规范

### API 设计要求

- [ ] URL 使用 snake_case 命名（如：`/api/v1/order_info`）
- [ ] data 字段必须是对象，不能是 null 或数组
- [ ] 响应格式：`{code, message, data{}}`
- [ ] 参考: `.cursor/rules/api.mdc` - API 设计规范

### 数据库设计要求

- [x] 必须包含: `id`, `uuid`, `create_time`, `update_time`, `delete_time`
- [x] 时间字段使用 int 类型，\_time 结尾，默认值 0
- [x] UUID 字段使用 bigint unsigned
- [x] 表名使用 ttpos\_ 前缀
- [x] 字段名使用 snake_case
- [x] `batch_cooking_mode` 字段使用 VARCHAR(10)，默认值 'post'
- [x] 参考: `.cursor/rules/database.mdc` - 数据库开发规范

### 性能要求

- [ ] 订单创建性能不受影响（本地响应时间 < 200ms）
- [ ] 数据库查询优化（使用索引）
- [ ] 送厨逻辑查询性能不受影响

### 测试要求

- [ ] Service 层测试覆盖率 ≥ 70%
- [ ] Repository 层测试覆盖率 ≥ 80%
- [ ] **Order 相关模块测试覆盖率 100%**（高风险）
- [ ] 集成测试覆盖订单创建和送厨流程
- [ ] 测试用例覆盖前置/后置模式、订单创建、送厨等场景
- [ ] 参考: `.cursor/rules/go-main.mdc` - 测试规范

### 可靠性要求

- [ ] 网络异常时优雅降级
- [ ] 事务管理（保证数据一致性）
- [ ] 错误日志记录（使用 Logger）
- [ ] 历史数据兼容性处理

---

## 验收标准

### 功能验收

1. **数据库字段**: `ttpos_sale_bill_setting` 表已新增 `batch_cooking_mode` 字段，历史数据已设置默认值
2. **订单创建**: 新创建的订单正确保存 `batch_cooking_mode` 值
3. **送厨逻辑**: 订单送厨时使用 `sale_bill_setting.batch_cooking_mode`，不受全局配置影响
4. **配置变更**: shop 端修改 `business_setting.BatchCookingMode` 后，已创建订单保持原有模式
5. **历史兼容**: 历史订单的 `batch_cooking_mode` 为空时，使用默认值 "post"

### 测试验收

1. **单元测试**: 覆盖率达标
2. **集成测试**: 订单创建和送厨流程测试通过
3. **手动测试**: 前置/后置模式切换测试通过
4. **回归测试**: 现有功能不受影响

### 文档验收

1. **技术文档**: design.md 完整且准确（待创建）
2. **数据库文档**: 迁移脚本和表结构文档完整
3. **测试文档**: tasks.md 中的测试任务完成（待创建）

---

## 约束条件

### 技术约束

#### Go Main 模块

- 必须使用 Gin 框架
- 接口以 `I` 开头，实现以 `Impl` 结尾
- Service 只能依赖其他 Service 接口
- Repository 只能持有 db 实例，不能持有 DBManager
- 不使用 panic，返回 error

### 业务约束

- 订单创建时的配置快照必须准确记录
- 已创建订单的行为不受后续配置变更影响
- 历史订单需要兼容处理

### 资源约束

- 开发时间: 2-3 天
- Story Point: 3-5 SP (必须 ≤ 5)

---

## 依赖关系

### 技术依赖

- `main/app/model/order.go` - SaleBillSetting 模型
- `main/app/service/order.go` - NewSaleBillSetting 函数
- `main/app/dto/resp/setting/business_setting.go` - BatchCookingMode 字段

### 业务依赖

- 订单创建流程
- 送厨流程
- 门店业务设置

---

## 风险和缓解

### 风险 1: 历史数据兼容性

**影响**: 中  
**概率**: 高  
**缓解措施**:

- 为历史订单设置默认值 "post"，与当前业务逻辑保持一致
- 在代码中添加兼容逻辑：如果 `batch_cooking_mode` 为空，使用默认值

### 风险 2: 送厨逻辑分散

**影响**: 高  
**概率**: 中  
**缓解措施**:

- 使用代码搜索工具全面查找所有使用 `BatchCookingMode` 的地方
- 代码审查确保所有相关逻辑都已修改
- 充分测试覆盖所有送厨场景

### 风险 3: 测试覆盖不足

**影响**: 中  
**概率**: 中  
**缓解措施**:

- 制定详细的测试用例，覆盖前置/后置模式、订单创建、送厨等场景
- 进行回归测试，确保现有功能不受影响
- 多终端测试（POS、Assistant、KDS 等）

---

## 时间表

- **Phase 1 - 数据库迁移**: 0.5 天
- **Phase 2 - 模型和订单创建逻辑**: 1 天
- **Phase 3 - 送厨逻辑修改和测试**: 1 天
- **Phase 4 - 联调和回归测试**: 0.5-1 天
- **总计**: 2-3 天（SP = 3-5）

---

## 参考资料

### 核心规范

- `.cursor/rules/go-main.mdc` - Go Main 核心约束
- `.cursor/rules/database.mdc` - 数据库开发规范
- `.cursor/rules/structs.mdc` - 项目结构规范

### 相关代码位置

- 模型定义: `main/app/model/order.go` (SaleBillSetting)
- 订单创建: `main/app/service/order.go` (NewSaleBillSetting)
- 业务设置: `main/app/dto/resp/setting/business_setting.go` (BatchCookingMode)
- 数据库表: `ttpos_sale_bill_setting`

### 相关提案

- `docs/team/proposals/2025-12/order-batch-cooking-mode-snapshot.md` - 需求提案
- `docs/team/proposals/2025-11/assistant-batch-cooking-pre-mode.md` - 类似功能
- `docs/team/proposals/2025-11/cashier-batch-cooking-pre-mode.md` - 类似功能

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 提醒：需求评审或范围调整若形成经验，应同步更新 Episode 并在 Proposal/Spec 互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-12-02  
**作者**: xiezhihuan  
**审核者**: {审核者}

