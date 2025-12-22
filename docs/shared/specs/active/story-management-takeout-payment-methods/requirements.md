# Grab/LINE MAN 外卖支付方式自动创建 需求文档

> 本文档定义 Grab/LINE MAN 外卖支付方式自动创建功能的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2025-12/grab-line-man-payment-methods.md](../../../../team/proposals/2025-12/grab-line-man-payment-methods.md) |
| **创建日期**      | 2025-12-22                                                                                                 |
| **负责人**        | 待分配                                                                                                       |
| **目标 Sprint**   | Sprint {N}                                                                                                   |
| **涉及技术栈**    | [x] Go (main/) [ ] Go (ttpos-bmp/) [ ] PHP (admin/) [ ] Vue (admin/views/)                                   |

## 📋 审核状态

| 项目         | 内容                     |
| ------------ | ------------------------ |
| **审核状态** | 已通过 |
| **审核人**   | 系统             |
| **审核日期** | 2025-12-22             |
| **审核意见** | 进入技术设计阶段         |

---

## 📋 概述

在配置 Grab 外卖或 LINE MAN 外卖时，系统自动创建对应的支付方式，确保支付方式数据完整，并支持自动同步到 ERP 系统。

**核心价值**：
- ✅ **自动化配置**：减少手动操作，提升配置效率
- ✅ **数据完整性**：确保支付方式数据与外卖平台配置保持一致
- ✅ **ERP 集成**：自动同步支付方式到 ERP，提升数据一致性
- ✅ **财务对账**：支持按支付方式统计和核对，提升财务准确性

## 🎯 产品对齐

该功能支持以下产品目标：
- **提升商户配置效率**：减少重复性手动操作
- **数据一致性保障**：确保支付方式与外卖平台配置同步
- **ERP 集成完善**：支持支付方式自动同步到 ERP 系统
- **财务对账准确性**：提供完整的支付方式数据支持

## 📝 用户故事

**作为** 商户管理员  
**我想** 在配置 Grab/LINE MAN 外卖时，系统自动创建对应的支付方式  
**以便于** 确保支付方式数据完整，并自动同步到 ERP 系统

---

## 功能需求

### Requirement 1: Grab 外卖支付方式自动创建

**用户故事**: 作为商户管理员，我想在配置 Grab 外卖时系统自动创建"Grab"支付方式，以便于确保支付方式数据完整并支持 ERP 同步

#### 验收标准

1. **WHEN** 商户配置 Grab 外卖并保存 **THEN** 系统 **SHALL** 自动创建"Grab"支付方式
2. **WHEN** 创建 Grab 支付方式 **THEN** 系统 **SHALL** 设置来源标识为"系统"（Source = 0）
3. **WHEN** 创建 Grab 支付方式 **THEN** 系统 **SHALL** 设置固定 code = 91100（参考 LianLianPay 格式，9 开头的 5 位数，为 LianLianPay 预留扩展空间 90111-91099，不与现有 code 冲突）
4. **WHEN** 创建 Grab 支付方式 **THEN** 系统 **SHALL** 设置不在旧后台、新管理端显示（IsShowCashier = 0, IsShowAssistant = 0 等）
5. **IF** 商户已开启 ERP 同步 **THEN** 系统 **SHALL** 自动将 Grab 支付方式同步到 ERP，**AND** ERP 同步失败时返回错误"创建 Grab 支付方式失败"，阻塞流程
6. **IF** Grab 支付方式已存在（通过 payment_name 或 code = 91100 判断） **THEN** 系统 **SHALL** 跳过创建，避免重复
7. **WHEN** 多次保存 Grab 外卖配置 **THEN** 系统 **SHALL** 不会重复创建支付方式（幂等性）

#### 具体要求

- [ ] 1.1 在外卖平台配置保存时触发支付方式创建逻辑
- [ ] 1.2 检查支付方式是否已存在（通过 payment_name = "Grab" 或 code = 91100 判断）
- [ ] 1.3 如果不存在，创建新的支付方式记录
- [ ] 1.4 设置支付方式属性：
  - Source = 0（系统默认）
  - Code = 91100（固定值，参考 LianLianPay 格式，9 开头的 5 位数，Grab 支付方式专用 code，为 LianLianPay 预留扩展空间 90111-91099，不与现有 code 冲突）
  - PaymentName = "Grab"
  - Name = "Grab"
  - IsShowCashier = 0（不在收银端显示）
  - IsShowAssistant = 0（不在助手端显示）
  - IsShowKiosk = 0（不在自助机显示）
  - IsShowMemberRecharge = 0（不在会员充值显示）
  - Status = 1（启用状态）
- [ ] 1.5 如果商户开启 ERP，调用 ERP 同步接口同步支付方式，ERP 同步失败时返回错误，阻塞流程
- [ ] 1.6 记录创建日志，便于追踪和调试

---

### Requirement 2: LINE MAN 外卖支付方式自动创建

**用户故事**: 作为商户管理员，我想在配置 LINE MAN 外卖时系统自动创建"LINE MAN"支付方式，以便于确保支付方式数据完整并支持 ERP 同步

#### 验收标准

1. **WHEN** 商户配置 LINE MAN 外卖并保存 **THEN** 系统 **SHALL** 自动创建"LINE MAN"支付方式
2. **WHEN** 创建 LINE MAN 支付方式 **THEN** 系统 **SHALL** 设置来源标识为"系统"（Source = 0）
3. **WHEN** 创建 LINE MAN 支付方式 **THEN** 系统 **SHALL** 设置固定 code = 91200（参考 LianLianPay 格式，9 开头的 5 位数，为 LianLianPay 预留扩展空间 90111-91099，不与现有 code 冲突）
4. **WHEN** 创建 LINE MAN 支付方式 **THEN** 系统 **SHALL** 设置不在旧后台、新管理端显示（IsShowCashier = 0, IsShowAssistant = 0 等）
5. **IF** 商户已开启 ERP 同步 **THEN** 系统 **SHALL** 自动将 LINE MAN 支付方式同步到 ERP，**AND** ERP 同步失败时返回错误"创建 LINE MAN 支付方式失败"，阻塞流程
6. **IF** LINE MAN 支付方式已存在（通过 payment_name 或 code = 91200 判断） **THEN** 系统 **SHALL** 跳过创建，避免重复
7. **WHEN** 多次保存 LINE MAN 外卖配置 **THEN** 系统 **SHALL** 不会重复创建支付方式（幂等性）

#### 具体要求

- [ ] 2.1 在外卖平台配置保存时触发支付方式创建逻辑
- [ ] 2.2 检查支付方式是否已存在（通过 payment_name = "LINE MAN" 或 code = 91200 判断）
- [ ] 2.3 如果不存在，创建新的支付方式记录
- [ ] 2.4 设置支付方式属性：
  - Source = 0（系统默认）
  - Code = 91200（固定值，参考 LianLianPay 格式，9 开头的 5 位数，LINE MAN 支付方式专用 code，为 LianLianPay 预留扩展空间 90111-91099，不与现有 code 冲突）
  - PaymentName = "LINE MAN"
  - Name = "LINE MAN"
  - IsShowCashier = 0（不在收银端显示）
  - IsShowAssistant = 0（不在助手端显示）
  - IsShowKiosk = 0（不在自助机显示）
  - IsShowMemberRecharge = 0（不在会员充值显示）
  - Status = 1（启用状态）
- [ ] 2.5 如果商户开启 ERP，调用 ERP 同步接口同步支付方式，ERP 同步失败时返回错误，阻塞流程
- [ ] 2.6 记录创建日志，便于追踪和调试

---

### Requirement 3: 支付方式去重和幂等性保障

**用户故事**: 作为系统，我想确保支付方式创建操作的幂等性，以便于避免重复创建和数据冲突

#### 验收标准

1. **IF** 支付方式已存在（通过 payment_name 判断） **THEN** 系统 **SHALL** 跳过创建
2. **WHEN** 多次调用创建逻辑 **THEN** 系统 **SHALL** 保证最终只存在一条支付方式记录
3. **IF** 支付方式创建失败 **THEN** 系统 **SHALL** 记录错误日志，不影响外卖配置保存

#### 具体要求

- [ ] 3.1 在创建前检查支付方式是否已存在（通过 payment_name 或 code 字段查询）
- [ ] 3.2 如果已存在，记录日志并跳过创建
- [ ] 3.3 如果不存在，执行创建逻辑
- [ ] 3.4 创建失败时记录错误日志，但不影响外卖配置保存的主流程

---

### Requirement 4: ERP 同步集成

**用户故事**: 作为商户管理员，我想在创建支付方式时自动同步到 ERP，以便于保持数据一致性

#### 验收标准

1. **IF** 商户已开启 ERP 同步 **THEN** 系统 **SHALL** 自动调用 ERP 同步接口
2. **IF** ERP 同步成功 **THEN** 系统 **SHALL** 更新支付方式的 erpnext_payment 字段
3. **IF** ERP 同步失败 **THEN** 系统 **SHALL** 返回错误"创建 xxx 支付方式失败"，阻塞流程，支付方式创建失败（事务回滚）

#### 具体要求

- [ ] 4.1 检测商户是否开启 ERP（通过 Company.IsOpenErp() 判断）
- [ ] 4.2 如果开启 ERP，调用 ERP 同步接口（参考 payment_method.go 中的 Create 方法）
- [ ] 4.3 根据支付方式 Source 确定 ERP Channel（使用 erpService.GetChannelBySource）
- [ ] 4.4 同步成功后更新支付方式的 erpnext_payment 字段
- [ ] 4.5 同步失败时返回错误，阻塞流程，支付方式创建失败（事务回滚）

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

- [ ] URL 使用 snake_case 命名（如：`/api/v1/order_info`）
- [ ] data 字段必须是对象，不能是 null 或数组
- [ ] 分页信息统一放在 meta 中
- [ ] 响应格式：`{code, message, data{}}`
- [ ] 参考: `.cursor/rules/api.mdc` - API 设计规范

### 数据库设计要求

- [ ] 必须包含: `id`, `uuid`, `create_time`, `update_time`, `delete_time`
- [ ] 时间字段使用 int 类型，\_time 结尾，默认值 0
- [ ] 金额字段使用 decimal(20,8)
- [ ] UUID 字段使用 bigint unsigned
- [ ] 表名使用 ttpos\_ 前缀
- [ ] 字段名使用 snake_case
- [ ] 参考: `.cursor/rules/database.mdc` - 数据库开发规范

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

1. **Grab 支付方式自动创建**: 配置 Grab 外卖时，系统自动创建"Grab"支付方式，code = 91100（固定值，参考 LianLianPay 格式，为 LianLianPay 预留扩展空间 90111-91099），来源标识为"系统"，且不在旧后台、新管理端显示
2. **LINE MAN 支付方式自动创建**: 配置 LINE MAN 外卖时，系统自动创建"LINE MAN"支付方式，code = 91200（固定值，参考 LianLianPay 格式，为 LianLianPay 预留扩展空间 90111-91099），来源标识为"系统"，且不在旧后台、新管理端显示
3. **ERP 同步**: 如果商户开启 ERP，支付方式自动同步到 ERP 系统
4. **去重保障**: 支付方式已存在时，不会重复创建
5. **幂等性**: 多次保存外卖配置时，不会重复创建支付方式

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

- 支付方式创建失败不应影响外卖配置保存
- **ERP 同步失败会返回错误，阻塞流程，支付方式创建失败（事务回滚）**
- 支付方式仅用于系统内部记录和 ERP 同步，不在前端显示

### 资源约束

- 开发时间: 3-5 天
- Story Point: 3-5 (必须 ≤ 5)

---

## 依赖关系

### 技术依赖

- `main/app/service/payment_method.go` - 支付方式服务
- `main/app/service/rpc/erp/selling.go` - ERP 同步服务
- `main/app/constant/payment.go` - 支付方式常量定义

### 服务依赖

- **Main → BMP**: gRPC 调用（外卖平台配置相关）
- **Main → ERP**: HTTP API 调用（支付方式同步）

### 业务依赖

- 外卖平台配置功能（Grab/LINE MAN 配置保存）
- 支付方式管理功能（支付方式创建和 ERP 同步）
- ERP 集成功能（ERP 同步接口）

---

## 风险和缓解

### 风险 1: 支付方式 code 冲突

**影响**: 低  
**概率**: 低  
**缓解措施**:

- 使用固定的 code 值（Grab = 91100, LINE MAN = 91200），参考 LianLianPay 格式（9 开头的 5 位数）
- 现有支付方式 code 范围：
  - 系统默认：-1, 10-190
  - LianLianPay：90111, 90222, 90333（预留扩展空间至 91099）
  - 手动添加：20000+（每次递增 100）
- 固定 code 值（91100, 91200）在 9 开头的专用范围内，为 LianLianPay 预留足够的扩展空间（90111-91099），逻辑上不会与已有 code 重复
- 未来新增其他外卖平台支付方式时，可继续使用 91300, 91400 等，避免累积冲突

### 风险 2: ERP 同步失败

**影响**: 高  
**概率**: 中  
**缓解措施**:

- ERP 同步失败时返回错误，阻塞流程，确保数据一致性
- 外部服务调用时需要在事务中处理，失败时自动回滚
- 监控 ERP 同步成功率，及时发现问题
- 提供重试机制或降级策略（如需要）

### 风险 3: 重复创建支付方式

**影响**: 低  
**概率**: 低  
**缓解措施**:

- 创建前检查支付方式是否已存在（通过 payment_name 字段）
- 如果已存在，跳过创建并记录日志
- 保证操作的幂等性

### 风险 4: 外卖配置删除时的支付方式处理

**影响**: 低  
**概率**: 低  
**缓解措施**:

- 明确删除策略：删除外卖配置时，支付方式保留但标记为禁用
- 或提供配置项，允许用户选择是否删除支付方式

---

## 时间表

- **Phase 1 - 需求分析和技术方案设计**: 1 天
- **Phase 2 - 支付方式创建逻辑实现**: 2 天
- **Phase 3 - ERP 同步集成和测试**: 1-2 天
- **总计**: 3-5 天（SP = 3-5）

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

### 相关代码

- `main/app/service/payment_method.go` - 支付方式服务（参考 Create 方法）
- `main/app/constant/payment.go` - 支付方式常量定义
  - 需要在常量文件中添加：`PaymentMethodCodeGrab = 91100` 和 `PaymentMethodCodeLineMan = 91200`
  - 参考 LianLianPay 的定义方式：`PaymentMethodCodeLianLianWechatPay = 90111`
  - 为 LianLianPay 预留扩展空间：90111-91099
- `main/app/service/rpc/erp/selling.go` - ERP 同步服务
- `main/app/model/payment_method.go` - 支付方式数据模型

### 外部参考

- [Grab 外卖平台配置文档]
- [LINE MAN 外卖平台配置文档]
- [ERP 支付方式同步接口文档]

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 提醒：需求评审或范围调整若形成经验，应同步更新 Episode 并在 Proposal/Spec 互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-12-22  
**作者**: 王昱  
**审核者**: 待审核

