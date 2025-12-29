# 支付方式快捷添加（Kbank渠道）需求文档

> 本文档定义支付方式快捷添加（Kbank渠道）功能的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2025-12/payment-method-quick-add-kbank.md](../../../../team/proposals/2025-12/payment-method-quick-add-kbank.md) |
| **创建日期**      | 2025-12-29                                                                                                 |
| **负责人**        | 王昱                                                                                                       |
| **目标 Sprint**   | Sprint {N}                                                                                                   |
| **涉及技术栈**    | [x] Go (main/) [ ] Go (ttpos-bmp/) [ ] PHP (admin/) [ ] Vue (admin/views/)                                   |

## 📋 审核状态

| 项目         | 内容                     |
| ------------ | ------------------------ |
| **审核状态** | 待审核 |
| **审核人**   | {审核人姓名}             |
| **审核日期** | {YYYY-MM-DD}             |
| **审核意见** | {审核意见，如有}         |

---

## 📋 概述

当前在管理端添加支付方式时，需要手动填写多个字段（名称、支付方式、source、图片等），操作繁琐。特别是对于Kbank渠道的支付方式，商户需要逐个添加Alipay（Kbank）、WeChatPay（Kbank）、Credit QR（Kbank）、Thai QR（Kbank）、Credit Card（Kbank）这5种支付方式，每次都需要重复填写相同的信息，效率低下。

本功能通过扩展 `GetDefaultPayList` 接口返回Kbank支付方式列表，并标记是否可添加，前端调用现有的 `Create` 接口批量创建，实现快捷添加Kbank支付方式的功能。

**业务价值**：
- 提升配置效率：一键快速添加Kbank渠道的5种支付方式，减少90%的重复操作
- 降低配置错误：避免手动填写导致的名称、source不一致问题
- 改善用户体验：简化支付方式配置流程，提升商户满意度
- 减少支持成本：减少因配置错误导致的客服咨询

## 🎯 产品对齐

本功能支持Shop商家管理端和Admin管理端的支付方式管理，提升商户配置支付方式的效率，降低操作复杂度，符合产品"简单易用"的设计理念。

## 📝 用户故事

**作为** 商户管理员  
**我想** 通过快捷添加功能一键添加Kbank渠道的支付方式  
**以便于** 快速完成支付方式配置，减少重复操作，提高配置效率

---

## 功能需求

### Requirement 1: 扩展 GetDefaultPayList 接口返回Kbank支付方式

**用户故事**: 作为商户管理员，我想在默认支付方式列表中看到Kbank支付方式，以便于快速选择添加

#### 验收标准

1. **WHEN** 前端调用 `GET /shop/payment_method/default_pay` 接口 **THEN** 系统 **SHALL** 返回包含5种Kbank支付方式的列表，且Kbank选项在最上面（sort值最小）
2. **WHEN** 接口返回支付方式列表 **THEN** 系统 **SHALL** 标记已添加的Kbank支付方式（can_add=false），未添加的标记为可添加（can_add=true）
3. **WHEN** 接口返回Kbank支付方式 **THEN** 系统 **SHALL** 包含以下字段：code、name（即payment_name）、url、img、sort、can_add、source

#### 具体要求

- [ ] 1.1 在 `GetDefaultPayList` 接口中增加5种Kbank支付方式定义
  - Alipay（Kbank）- code: 93000
  - WeChatPay（Kbank）- code: 93100
  - Credit QR（Kbank）- code: 93200
  - Thai QR（Kbank）- code: 93300
  - Credit Card（Kbank）- code: 93400
- [ ] 1.2 Kbank支付方式在返回列表最前面（sort值最小，如0-4）
- [ ] 1.3 扩展 `DefaultPaymentMethodResp` 响应结构，增加 `can_add`、`source` 字段（`name` 字段即为 `payment_name`，无需新增）
- [ ] 1.4 实现重复检测逻辑：查询当前商户已添加的Kbank支付方式（通过payment_name和source=3匹配），标记can_add=false
- [ ] 1.5 未添加的Kbank支付方式标记can_add=true

---

### Requirement 2: 批量创建Kbank支付方式

**用户故事**: 作为商户管理员，我想批量创建Kbank支付方式，以便于一次性添加多个支付方式

#### 验收标准

1. **WHEN** 前端调用 `POST /shop/payment_method/create` 接口并传入Kbank支付方式信息（包含source参数） **THEN** 系统 **SHALL** 批量创建支付方式，使用传入的source值
2. **WHEN** 前端传入已添加的Kbank支付方式（通过payment_name和source匹配） **THEN** 系统 **SHALL** 返回错误提示，不重复创建
3. **IF** 前端传入空列表 **THEN** 系统 **SHALL** 返回参数错误提示
4. **WHEN** 批量创建成功 **THEN** 系统 **SHALL** 返回"创建成功"提示

#### 具体要求

- [ ] 2.1 在 `PaymentMethodCreateItem` 中增加 `source` 字段（可选字段）
- [ ] 2.2 前端根据 `GetDefaultPayList` 返回的信息构造请求参数，包含source=3
- [ ] 2.3 后端创建支付方式时，如果传入source字段，使用传入的值；否则使用默认值（source=1）
- [ ] 2.4 实现重复检测：创建前检查是否已存在相同的payment_name和source的支付方式，如存在则返回错误
- [ ] 2.5 支持批量创建多个Kbank支付方式

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

- [x] URL 使用 snake_case 命名（如：`/api/v1/payment_method/default_pay`）
- [x] data 字段必须是对象，不能是 null 或数组
- [x] 分页信息统一放在 meta 中
- [x] 响应格式：`{code, message, data{}}`
- [x] 参考: `.cursor/rules/api.mdc` - API 设计规范

### 数据库设计要求

- [x] 已有source字段，无需新增字段
- [x] 使用现有 `ttpos_payment_method` 表
- [x] 参考: `.cursor/rules/database.mdc` - 数据库开发规范

### 性能要求

- [ ] 本地响应时间 < 200ms
- [ ] 数据库查询优化（使用索引，查询payment_name和source）
- [ ] 缓存策略（如需要）
- [ ] 并发处理（使用 UUID 锁）

### 浏览器兼容性（管理后台）

- [ ] Chrome 90+
- [ ] Safari 14+
- [ ] Firefox 88+
- [ ] Edge 90+

### 测试要求

- [ ] Service 层测试覆盖率 ≥ 70%
- [ ] Repository 层测试覆盖率 ≥ 80%
- [x] **Payment/Order 相关模块测试覆盖率 100%**（高风险）
- [ ] 集成测试覆盖核心流程
- [ ] API 测试覆盖所有接口
- [ ] 参考: `.cursor/rules/go-main.mdc` - 测试规范

### 国际化要求

- [ ] 支持 10 种语言（中文、英文、日语、韩语等）
- [ ] 所有文案使用多语言实现
- [ ] 参考: `main/i18n/` - 国际化配置

### 安全要求

- [x] 所有 API 需要身份验证
- [ ] 敏感数据加密存储
- [x] SQL 注入防护（使用参数化查询）
- [ ] XSS 防护（前端输入校验）
- [x] CSRF 防护（Token 验证）
- [ ] 参考: `.cursor/rules/security.mdc` - 安全开发规范

### 可靠性要求

- [ ] 网络异常时优雅降级
- [x] 事务管理（保证数据一致性）
- [x] 错误日志记录（使用 Logger）
- [ ] 故障恢复机制

---

## 验收标准

### 功能验收

1. **GetDefaultPayList接口扩展**: 返回5种Kbank支付方式，且在最前面，正确标记can_add状态
2. **批量创建功能**: 支持批量创建Kbank支付方式，正确使用source=3
3. **重复检测**: 已添加的Kbank支付方式不能重复创建
4. **接口兼容性**: 扩展后的接口不影响现有调用

### 测试验收

1. **单元测试**: 覆盖率达标，Payment相关模块100%覆盖
2. **API 测试**: 所有接口测试通过
3. **集成测试**: 端到端流程测试通过
4. **手动测试**: 浏览器兼容性测试通过

### 文档验收

1. **技术文档**: design.md 完整且准确
2. **API 文档**: API 接口文档完整
3. **数据库文档**: 无需新增字段，已有字段说明完整
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

### 业务约束

- Kbank支付方式code固定为：93000、93100、93200、93300、93400
- source字段值：Kbank使用source=3（需确认）
- 已添加的Kbank支付方式不能重复创建

### 资源约束

- 开发时间: 1-2 天
- Story Point: 2 SP (必须 ≤ 5)

---

## 依赖关系

### 技术依赖

- `main/app/service/payment_method.go` - 支付方式服务
- `main/app/api/v1/shop/shop_payment_method.go` - 支付方式API
- `main/app/dto/resp/payment_method.go` - 响应结构
- `main/app/dto/req/payment_method_req.go` - 请求结构

### 服务依赖

- 无外部服务依赖

### 业务依赖

- 依赖现有的支付方式管理功能
- 依赖现有的 `GetDefaultPayList` 和 `Create` 接口

---

## 风险和缓解

### 风险 1: source字段值定义不明确

**影响**: 中  
**概率**: 中  
**缓解措施**:

- 确认Kbank支付方式的source值（建议使用source=3，或复用现有值）
- 在技术设计阶段明确source值的定义和使用规则

### 风险 2: 接口兼容性问题

**影响**: 高  
**概率**: 低  
**缓解措施**:

- 扩展 `GetDefaultPayList` 时，新增字段使用可选字段，确保现有调用不受影响
- 充分测试现有调用场景

### 风险 3: 重复检测逻辑不准确

**影响**: 中  
**概率**: 低  
**缓解措施**:

- 通过查询 payment_name 和 source 字段判断是否已添加
- 编写充分的单元测试和集成测试

---

## 时间表

- **Phase 1 - 接口扩展**: 0.5-1天（扩展 GetDefaultPayList 接口，增加Kbank支付方式定义、重复检测逻辑、选项排序）
- **Phase 2 - Create接口扩展**: 0.5天（在 PaymentMethodCreateItem 中增加source字段）
- **Phase 3 - 测试**: 0.5-1天（单元测试、API测试、集成测试）
- **总计**: 1-2天（SP = 2）

---

## 参考资料

### 核心规范

- `.cursor/rules/go-main.mdc` - Go Main 核心约束
- `.cursor/rules/api.mdc` - API 设计规范
- `.cursor/rules/database.mdc` - 数据库开发规范
- `.cursor/rules/security.mdc` - 安全开发规范

### 架构文档

- `docs/human/architecture/features/payment_method.md` - 支付方式管理文档

### 开发指南

- `main/app/service/payment_method.go` - 支付方式服务实现
- `main/app/api/v1/shop/shop_payment_method.go` - 支付方式API实现
- `main/app/dto/resp/payment_method.go` - 响应结构定义
- `main/app/dto/req/payment_method_req.go` - 请求结构定义

### 外部参考

- Kbank集成文档: `docs/others/kbank/KBTG-LINKPOS-Specification-v1.5.0.md`

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 提醒：需求评审或范围调整若形成经验，应同步更新 Episode 并在 Proposal/Spec 互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-12-29  
**作者**: 王昱  
**审核者**: {审核者}

