# 新管理端-支付管理 需求文档

> 本文档定义 新管理端-支付管理 的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2025-12/new-admin-payment-management.md](../../../../team/proposals/2025-12/new-admin-payment-management.md) |
| **创建日期**      | 2025-12-11                                                                                                 |
| **负责人**        | {姓名}                                                                                                       |
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

在新管理端（Shop 商家管理端）新增支付管理功能模块，提供完整的支付方式 CRUD 操作 API 接口，支持商户通过 API 自主管理支付方式，包括添加、编辑、状态变更（启用/禁用）、删除等操作。同时提供 LianlianPay 专项配置 API，支持配置白名单 IP、商户号、支付方式公钥、商户私钥、Token、站点 ID 等参数。

该功能将显著提升运营效率，减少对技术人员的依赖，增强支付方式管理的灵活性。前端界面由其他团队或后续迭代实现。

## 🎯 产品对齐

该功能支持 TTPOS 产品的核心定位：**收银 + ERP + 会员 + 外卖**中的收银模块，通过完整的支付管理 API 接口，提升商户自主运营能力，降低运维成本，适应业务快速变化的需求。

## 📝 用户故事

**作为** 商户管理员  
**我想** 在管理端自主管理支付方式，包括添加、编辑、启用/禁用、删除，以及配置 LianlianPay 的相关参数  
**以便于** 快速响应业务需求，减少对技术人员的依赖，提升运营效率

---

## 功能需求

### Requirement 1: 支付方式列表查询 API

**用户故事**: 作为商户管理员，我想通过 API 查询所有支付方式列表，以便于了解当前可用的支付方式

#### 验收标准

1. **WHEN** 调用支付方式列表查询 API **THEN** 系统 **SHALL** 返回所有支付方式列表
2. **WHEN** 查询支付方式列表 **THEN** 系统 **SHALL** 返回支付方式的基本信息（名称、代码、状态、排序等）
3. **WHEN** 传入搜索参数 **THEN** 系统 **SHALL** 支持按支付方式名称、代码进行搜索
4. **WHEN** 传入筛选参数 **THEN** 系统 **SHALL** 支持按状态（启用/禁用）、来源（系统/手动/LianLianPay）进行筛选

#### 具体要求

- [ ] 1.1 API 支持分页参数（page, page_size）
- [ ] 1.2 返回字段：名称、代码、支付名称、来源、状态、排序、创建时间
- [ ] 1.3 支持按排序字段排序（sort 参数）
- [ ] 1.4 API 响应格式符合规范：`{code, message, data: {list: [], meta: {}}}`

---

### Requirement 2: 添加支付方式 API

**用户故事**: 作为商户管理员，我想通过 API 添加新的支付方式，以便于扩展支付渠道

#### 验收标准

1. **WHEN** 调用添加支付方式 API **THEN** 系统 **SHALL** 创建新的支付方式记录
2. **IF** 必填字段未填写 **THEN** 系统 **SHALL** 返回错误信息（code != 0）
3. **IF** 支付方式代码已存在 **THEN** 系统 **SHALL** 返回错误并阻止创建
4. **WHEN** 创建成功 **THEN** 系统 **SHALL** 返回创建的支付方式信息

#### 具体要求

- [ ] 2.1 API 请求字段：名称（必填）、代码（必填）、支付名称（必填）、来源（必填）、Logo 图片 UUID、二维码图片 UUID、手续费百分比、显示设置（收银机/助手/会员充值）、状态、排序
- [ ] 2.2 代码字段需校验唯一性（数据库唯一索引）
- [ ] 2.3 手续费百分比范围：0-1，需参数验证
- [ ] 2.4 Logo 和二维码图片通过 UUID 关联（需先上传文件获取 UUID）
- [ ] 2.5 API 响应格式：`{code, message, data: {payment_method: {}}}`

---

### Requirement 3: 编辑支付方式 API

**用户故事**: 作为商户管理员，我想通过 API 编辑支付方式的配置信息，以便于更新支付方式参数

#### 验收标准

1. **WHEN** 调用编辑支付方式 API **THEN** 系统 **SHALL** 更新支付方式记录
2. **IF** 修改后的代码与其他支付方式冲突 **THEN** 系统 **SHALL** 返回错误并阻止更新
3. **IF** 支付方式来源为"系统" **THEN** 系统 **SHALL** 限制部分字段的修改权限（返回错误）
4. **WHEN** 更新成功 **THEN** 系统 **SHALL** 返回更新后的支付方式信息

#### 具体要求

- [ ] 3.1 API 请求字段与添加接口一致（通过 UUID 标识要更新的支付方式）
- [ ] 3.2 系统来源的支付方式（source=0），代码字段不可修改（返回错误）
- [ ] 3.3 API 响应格式：`{code, message, data: {payment_method: {}}}`
- [ ] 3.4 支持修改所有可编辑字段（除系统限制字段外）

---

### Requirement 4: 支付方式状态变更 API

**用户故事**: 作为商户管理员，我想通过 API 启用或禁用支付方式，以便于控制支付方式在收银端的显示

#### 验收标准

1. **WHEN** 调用支付方式状态变更 API **THEN** 系统 **SHALL** 更新支付方式状态
2. **WHEN** 支付方式状态变更为"禁用" **THEN** 系统 **SHALL** 同步影响收银端显示，该支付方式不再显示
3. **WHEN** 支付方式状态变更为"启用" **THEN** 系统 **SHALL** 同步影响收银端显示，该支付方式正常显示
4. **IF** 支付方式有关联的未完成订单 **THEN** 系统 **SHALL** 返回警告信息（但仍允许操作）

#### 具体要求

- [ ] 4.1 API 支持状态参数（status: 0-禁用, 1-启用）
- [ ] 4.2 API 通过 UUID 标识要更新的支付方式
- [ ] 4.3 状态变更后实时同步到收银端（通过现有同步机制）
- [ ] 4.4 记录状态变更日志（可选，通过操作日志记录）

---

### Requirement 5: 删除支付方式 API

**用户故事**: 作为商户管理员，我想通过 API 删除不再使用的支付方式，以便于清理无效配置

#### 验收标准

1. **WHEN** 调用删除支付方式 API **THEN** 系统 **SHALL** 检查是否有关联订单
2. **IF** 支付方式有关联的历史订单 **THEN** 系统 **SHALL** 返回错误并禁止删除（或仅允许软删除）
3. **IF** 支付方式无关联订单 **THEN** 系统 **SHALL** 执行删除操作（软删除）
4. **IF** 支付方式来源为"系统" **THEN** 系统 **SHALL** 返回错误并禁止删除

#### 具体要求

- [ ] 5.1 API 通过 UUID 标识要删除的支付方式
- [ ] 5.2 检查关联订单逻辑（查询 `ttpos_payment_order` 表）
- [ ] 5.3 系统来源的支付方式（source=0）禁止删除（返回错误）
- [ ] 5.4 删除采用软删除（设置 `delete_time`）
- [ ] 5.5 API 响应格式：`{code, message, data: {}}`

---

### Requirement 6: LianlianPay 专项配置 API

**用户故事**: 作为商户管理员，我想通过 API 配置 LianlianPay 的相关参数，以便于正确接入 LianlianPay 支付渠道

#### 验收标准

1. **WHEN** 调用 LianlianPay 配置 API **THEN** 系统 **SHALL** 保存配置信息
2. **IF** 配置了敏感信息（私钥、Token） **THEN** 系统 **SHALL** 加密存储
3. **IF** 必填配置项未填写 **THEN** 系统 **SHALL** 返回错误并阻止保存
4. **WHEN** 配置保存成功 **THEN** 系统 **SHALL** 返回配置信息（敏感字段返回加密后的值或占位符）

#### 具体要求

- [ ] 6.1 配置项包括：白名单 IP、商户号、支付方式公钥、商户私钥、Token、站点 ID
- [ ] 6.2 敏感字段（商户私钥、Token）需加密存储（使用加密算法）
- [ ] 6.3 敏感字段在 API 响应中返回占位符或加密后的值（不返回明文）
- [ ] 6.4 支持配置项的查询和更新（通过支付方式 UUID 关联）
- [ ] 6.5 配置项验证：IP 格式、密钥格式等（参数验证）
- [ ] 6.6 配置保存后实时生效（立即更新到支付服务配置）

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 严格遵循 Controller → Service → Repository 分层
- **单一职责原则**: 每个文件应有单一、明确的目的
- **模块化设计**: Service 和 Repository 应独立且可复用
- **依赖管理**: Service 只能依赖其他 Service 接口，不能直接依赖 Repository
- **遵循规范**:
  - `.cursor/rules/go-main.mdc` - Go Main 开发规范
  - `.cursor/rules/structs.mdc` - 项目结构规范

### API 设计要求

- [x] URL 使用 snake_case 命名（如：`/api/v1/shop/payment_method`）
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

**注意**: 支付方式表 `ttpos_payment_method` 已存在，本次需求主要新增管理界面，不涉及表结构变更。LianlianPay 配置信息需考虑存储方案（可扩展表或配置表）。

### 性能要求

- [ ] 本地响应时间 < 200ms
- [ ] 数据库查询优化（使用索引）
- [ ] 缓存策略（Redis）- 支付方式列表可缓存
- [ ] 并发处理（使用 UUID 锁）

### 浏览器兼容性（管理后台）

- [ ] 不适用（仅后端 API，无前端界面）

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
- [x] 敏感数据加密存储（LianlianPay 私钥、Token）
- [x] SQL 注入防护（使用参数化查询）
- [x] CSRF 防护（Token 验证）
- [x] 参考: `.cursor/rules/security.mdc` - 安全开发规范

### 可靠性要求

- [ ] 网络异常时优雅降级
- [ ] 事务管理（保证数据一致性）
- [ ] 错误日志记录（使用 Logger）
- [ ] 故障恢复机制

---

## 验收标准

### 功能验收

1. **支付方式列表 API**: 能够正确返回所有支付方式，支持搜索和筛选参数
2. **添加支付方式 API**: 能够成功创建新的支付方式，字段验证正确
3. **编辑支付方式 API**: 能够成功更新支付方式信息，权限控制正确
4. **状态变更 API**: 能够正确切换支付方式状态，收银端同步生效
5. **删除支付方式 API**: 能够正确删除支付方式，关联检查正确
6. **LianlianPay 配置 API**: 能够正确配置 LianlianPay 参数，敏感信息加密存储

### 测试验收

1. **单元测试**: 覆盖率达标（Service ≥ 70%, Repository ≥ 80%, Payment 相关 100%）
2. **API 测试**: 所有接口测试通过（使用 Postman/curl 等工具）
3. **集成测试**: 端到端流程测试通过

### 文档验收

1. **技术文档**: design.md 完整且准确
2. **API 文档**: API 接口文档完整
3. **数据库文档**: 如有表结构变更，迁移脚本和表结构文档完整
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

- 系统来源的支付方式（source=0）部分字段不可修改，不可删除
- 删除支付方式前需检查关联订单，有历史订单时禁止删除或仅允许软删除
- LianlianPay 配置参数敏感，必须加密存储
- 支付方式状态变更需实时同步到收银端

### 资源约束

- 开发时间: 3-4 天（仅后端 API，无前端界面）
- Story Point: 3 (必须 ≤ 5)

---

## 依赖关系

### 技术依赖

- `main/app/model/payment_method.go` - 支付方式数据模型
- `main/app/repository/payment_method.go` - 支付方式数据访问层
- `main/app/service/payment.go` - 支付服务（LianlianPay 相关）

### 服务依赖

- **外部系统 → Main**: HTTP API 调用（其他系统或前端调用 Go Main API）

### 业务依赖

- 支付方式表 `ttpos_payment_method` 已存在
- 支付订单表 `ttpos_payment_order` 用于关联检查
- LianlianPay 支付服务已集成，需扩展配置管理功能

---

## 风险和缓解

### 风险 1: 支付方式删除导致数据不一致

**影响**: 高  
**概率**: 中  
**缓解措施**:

- 删除前进行关联检查，查询 `ttpos_payment_order` 表
- 有历史订单时禁止删除或仅允许软删除
- 删除操作记录日志，便于追溯

### 风险 2: LianlianPay 配置参数泄露

**影响**: 高  
**概率**: 低  
**缓解措施**:

- 敏感字段（私钥、Token）使用加密存储
- 传输使用 HTTPS
- 界面显示为星号或密文
- 访问权限控制

### 风险 3: 状态变更影响正在进行的订单

**影响**: 中  
**概率**: 低  
**缓解措施**:

- 状态变更采用软删除或状态标记，保留历史记录
- 状态变更时提示风险并允许确认操作
- 实时同步到收银端，避免数据不一致

---

## 时间表

- **Phase 1 - 后端 API 开发**: 2-3 天
- **Phase 2 - 测试与优化**: 1 天
- **总计**: 3-4 天（SP = 3）

---

## 参考资料

### 核心规范

- `.cursor/rules/go-main.mdc` - Go Main 核心约束
- `.cursor/rules/api.mdc` - API 设计规范
- `.cursor/rules/database.mdc` - 数据库开发规范
- `.cursor/rules/security.mdc` - 安全开发规范

### 架构文档

- `docs/human/architecture/features/payment_method.md` - 支付方式架构文档
- `docs/human/architecture/features/payment.md` - 支付架构文档

### 开发指南

- `docs/human/guides/go-main-development.md` - Go Main 开发指南
- `docs/human/guides/api-design-guide.md` - API 设计指南

### 外部参考

- LianlianPay 官方文档
- 现有支付方式管理代码：`main/app/repository/payment_method.go`

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 提醒：需求评审或范围调整若形成经验，应同步更新 Episode 并在 Proposal/Spec 互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-12-11  
**作者**: 王昱  
**审核者**: {审核者}
