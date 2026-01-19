> ⚠️ **已归档** - 此 Spec 已随 v2.12 发布。
>
> - 归档时间: 2026-01-04
> - 归档人: weifashi

# story-admin-payment-mode-management / 新管理端支付方式管理 需求文档

> 本文档定义新管理端支付方式管理与 ERPNext 双向同步的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2025-12/admin-payment-erp-integration.md](../../../../team/proposals/2025-12/admin-payment-erp-integration.md) |
| **创建日期**      | 2025-12-12                                                                                                 |
| **负责人**        | 王昱                                                                                                       |
| **目标 Sprint**   | 待定                                                                                                   |
| **涉及技术栈**    | [x] Go (main/) [x] Go (ttpos-bmp/) [ ] PHP (admin/) [x] Vue (admin/views/)                                   |

## 📋 审核状态

| 项目         | 内容                     |
| ------------ | ------------------------ |
| **审核状态** | 待审核 |
| **审核人**   | 待定             |
| **审核日期** | -             |
| **审核意见** | -         |

---

## 📋 概述

新管理端（Go Main 模块）需要实现支付方式管理功能，支持支付方式的创建、更新、删除，并与 ERPNext 系统进行双向同步。核心目标是：

1. **支付方式类型管理**：使用现有 `source` 字段支持 LianLianPay（2）、通用（1）、系统默认（0）三种类型
2. **标准化命名规则**：按照"渠道+支付方式+序号+商家缩写"规则生成 Mode of Payment ID，存储在 `erpnext_payment` 字段
3. **双向数据同步**：TTPOS ↔ ERPNext 自动同步支付方式数据
4. **公司账号关联**：添加支付方式后自动将当前公司加入 ERP 账号配置

**技术说明**：使用现有数据表 `ttpos_payment_method`，无需创建新表。branch、company、abbr 信息从 `ttpos_company_setting` 关联表获取。

该功能是新管理端迁移的重要组成部分，为后续财务核算、支付流水同步奠定基础。

## 🎯 产品对齐

该功能支持以下产品目标：

- **数据统一管理**：避免支付数据孤岛，实现 TTPOS 和 ERP 的数据一致性
- **多渠道支付**：支持连连支付等第三方支付渠道接入
- **财务自动化**：为后续财务核算、对账功能提供数据基础
- **系统迁移**：加速从旧 PHP 管理端向新 Go 管理端的迁移

## 📝 用户故事

**作为** 商户管理员  
**我想** 在新管理端配置支付方式并自动同步到 ERP  
**以便于** 实现支付数据的统一管理和财务核算

**作为** 系统管理员  
**我想** 系统自动处理 TTPOS 和 ERP 之间的支付方式同步  
**以便于** 减少手动操作，避免数据不一致

---

## 功能需求

### Requirement 1: 支付方式类型（Type）管理

**用户故事**: 作为商户管理员，我想管理不同类型的支付方式，以便于区分渠道来源和自定义支付方式

#### 验收标准

1. **WHEN** 查看支付方式类型列表 **THEN** 系统 **SHALL** 显示以下类型选项：
   - **LianLianPay**（连连支付渠道，用于存放 lianlianpay 的支付方式）
   - **通用**（自行添加的支付方式）
   - **系统默认**（Cash、Balance payment 等系统内置支付方式）

2. **WHEN** 创建支付方式 **THEN** 系统 **SHALL** 根据类型自动生成符合规范的 Mode of Payment ID

3. **WHEN** 选择 LianLianPay 类型 **THEN** 系统 **SHALL** 标记该支付方式为渠道同步类型

4. **WHEN** 选择通用类型 **THEN** 系统 **SHALL** 标记该支付方式为自行添加类型

#### 具体要求

- [x] 1.1 支持 LianLianPay、通用、系统默认三种类型（使用 `source` 字段：0=系统，1=手动，2=LianLianPay）
- [x] 1.2 类型字段为必填项
- [x] 1.3 类型不可修改（创建后锁定）
- [x] 1.4 在前端 UI 中清晰展示类型标签
- [x] 1.5 类型枚举值：`source=0`（系统默认）、`source=1`（通用/手动）、`source=2`（LianLianPay）

---

### Requirement 2: Mode of Payment ID 命名规则

**用户故事**: 作为开发人员，我想系统按照统一的命名规则生成 Mode of Payment ID，以便于追溯来源和避免冲突

#### 验收标准

1. **GIVEN** 类型为 LianLianPay **THEN** ID 格式 **SHALL** 为：`{渠道}-{支付方式}-{序号}-{商家缩写}`
   - 渠道：目前仅有 LianLianPay 渠道，其他的均为自行添加和系统默认
   - 支付方式：TTPOS 具体的支付方式字段（如 Alipay、WeChat Pay）
   - 序号：0000/0001，用作如果有同名的拓展，比如自行添加多个支付方式，默认的使用 0000，自行添加使用 0001 起进行同名序号
   - 商家缩写：Company Abbr（如 No.21）
   - 示例：`LianLianPay-Alipay-0000-No.21`

2. **GIVEN** 类型为通用（自行添加）**THEN** ID 格式 **SHALL** 为：`{支付方式}-{序号}-{商家缩写}`
   - 示例：`Alipay-0001-No.21`（第一个自行添加的 Alipay）
   - 示例：`Alipay-0002-No.21`（第二个自行添加的 Alipay）
   - 示例：`wechat pay-0001-No.21`（第一个自行添加的 WeChat Pay）

3. **GIVEN** 类型为系统默认 **THEN** ID 格式 **SHALL** 为：`{支付方式}-0000-{商家缩写}`
   - 示例：`Cash-0000-No.21`
   - 示例：`Balance payment-0000-No.21`

4. **WHEN** 存在同名支付方式 **THEN** 系统 **SHALL** 自动递增序号（0001, 0002, 0003...）

5. **WHEN** 商家缩写为空 **THEN** 系统 **SHALL** 使用商家 ID 或默认值

#### 具体要求

- [x] 2.1 实现命名规则生成器（Go Service 层）
- [x] 2.2 序号自动递增逻辑（查询已有同名支付方式）
- [x] 2.3 商家缩写（Abbr）从 `ttpos_company_setting` 表获取（`erpnext_company_abbr`）
- [x] 2.4 ID 唯一性校验（数据库层 + 业务层）
- [x] 2.5 支持命名规则配置化（便于后续调整）
- [x] 2.6 序号从 0000（系统/默认）或 0001（自建）起递增
- [x] 2.7 生成的 Mode of Payment ID 存储在 `erpnext_payment` 字段

---

### Requirement 3: 支付方式字段映射

**用户故事**: 作为开发人员，我想明确 TTPOS 和 ERP 之间的字段映射关系，以便于数据同步的准确性

#### 验收标准

1. **GIVEN** TTPOS 支付方式字段 **THEN** 系统 **SHALL** 映射到 ERP Mode of Payment 字段：

| TTPOS 字段（表字段） | ERP 字段 | 说明 | 必填 |
|-----------|---------|------|------|
| `source` | `type` | 类型（0=System，1=General，2=LianLianPay） | ✅ |
| `erpnext_payment` | `name` | Mode of Payment ID（按命名规则生成） | ✅ |
| `payment_name` | `mode_of_payment` | 支付方式名称（如 Alipay, WeChat Pay） | ✅ |
| `status` | `enabled` | 状态（0=禁用，1=启用） | ✅ |
| `branch`（从关联表获取） | `custom_branch` | 商家分店 | ✅ |
| `company`（从关联表获取） | `custom_company` | 商家公司 | ✅ |
| `abbr`（从关联表获取） | - | 商家缩写（用于 ID 生成） | ✅ |
| `name` | `description` | 支付方式名称（可选） | ❌ |
| `logo_file_uuid` | - | 支付图标（TTPOS 专用） | ❌ |
| `sort` | - | 排序顺序（TTPOS 专用） | ❌ |

2. **WHEN** 同步到 ERP **THEN** 系统 **SHALL** 忽略 TTPOS 专用字段（icon, sort_order 等）

3. **WHEN** 从 ERP 同步 **THEN** 系统 **SHALL** 保留 TTPOS 原有的 icon 和 sort_order

#### 具体要求

- [x] 3.1 定义字段映射配置（DTO 层）
- [x] 3.2 实现双向转换器（TTPOS ↔ ERP）
- [x] 3.3 必填字段校验
- [x] 3.4 字段值归一化处理（如 status 与 enabled 的转换：0↔0，1↔1）
- [x] 3.5 从 `ttpos_company_setting` 表获取 branch、company、abbr 信息

---

### Requirement 4: TTPOS → ERP 同步

**用户故事**: 作为商户管理员，我想在 TTPOS 添加/修改支付方式时自动同步到 ERP，以便于保持数据一致性

#### 验收标准

1. **WHEN** 在 TTPOS 创建支付方式 **THEN** 系统 **SHALL**：
   - 按命名规则生成 Mode of Payment ID
   - 调用 ERP API 创建 Mode of Payment
   - **将当前公司（Company）加入 Mode of Payment 的账号配置中**
   - 使用支付方式名称（payment_method）作为 ERP 的 mode_of_payment 字段值
   - 记录同步日志

2. **WHEN** 在 TTPOS 修改支付方式 **THEN** 系统 **SHALL**：
   - 调用 ERP API 更新 Mode of Payment
   - 同步状态、描述等可修改字段
   - 记录同步日志

3. **WHEN** 在 TTPOS 删除支付方式 **THEN** 系统 **SHALL**：
   - 调用 ERP API 禁用 Mode of Payment（不删除）
   - 设置 ERP `enabled = 0`
   - TTPOS 本地使用软删除
   - 记录同步日志

4. **IF** ERP API 调用失败 **THEN** 系统 **SHALL**：
   - 记录失败日志（包括错误原因）
   - 标记同步状态为失败
   - 支持手动重试
   - 不回滚 TTPOS 本地操作（异步同步）

5. **WHEN** 添加支付方式后 **THEN** 系统 **SHALL** 将当前公司（Company）加入账号内（Mode of Payment Accounts）

#### 具体要求

- [x] 4.1 实现 ERP RPC 调用（CreateModeOfPayment, UpdateModeOfPayment, DisableModeOfPayment）
- [x] 4.2 异步同步机制（RabbitMQ 队列）
- [x] 4.3 重试机制（最多 3 次）
- [x] 4.4 同步日志表设计（记录状态、错误信息、重试次数）
- [x] 4.5 Company 账号自动添加逻辑（调用 CreateModePaymentAccount）
- [x] 4.6 软删除实现（TTPOS 本地）
- [x] 4.7 使用支付方式名称（payment_method）作为 ERP mode_of_payment 字段值

---

### Requirement 5: ERP → TTPOS 同步

**用户故事**: 作为系统管理员，我想在 ERP 添加/修改支付方式时自动同步到 TTPOS，以便于支持从 ERP 侧管理支付方式

#### 验收标准

1. **WHEN** 店铺触发同步操作 **THEN** 系统 **SHALL**：
   - 调用 ERP API 获取该店铺（Branch）的所有 Mode of Payment
   - 比对 TTPOS 本地数据
   - 新增、更新或禁用本地支付方式
   - 记录同步日志

2. **GIVEN** ERP 新增 Mode of Payment **THEN** TTPOS **SHALL**：
   - 创建对应的支付方式记录
   - 根据命名规则识别类型（LianLianPay/通用/系统默认）
   - 自动提取商家缩写（Abbr）

3. **GIVEN** ERP 更新 Mode of Payment **THEN** TTPOS **SHALL**：
   - 更新对应的支付方式记录
   - 保留 TTPOS 专用字段（icon, sort_order）

4. **GIVEN** ERP 禁用 Mode of Payment **THEN** TTPOS **SHALL**：
   - 禁用对应的支付方式（软删除）

5. **IF** 同步冲突（TTPOS 和 ERP 数据不一致）**THEN** 系统 **SHALL**：
   - 以 ERP 数据为准（ERP 是主数据源）
   - 记录冲突日志
   - 通知管理员审核

#### 具体要求

- [x] 5.1 实现 ERP RPC 调用（ListModeOfPayments）
- [x] 5.2 数据对比算法（增量同步）
- [x] 5.3 类型识别逻辑（根据命名规则反向推断）
- [x] 5.4 冲突处理策略
- [x] 5.5 同步触发机制（手动触发 + 定时任务）
- [x] 5.6 同步日志记录

---

### Requirement 6: 旧数据处理策略

**用户故事**: 作为系统管理员，我想明确旧数据的处理策略，以便于避免数据冲突和迁移问题

#### 验收标准

1. **GIVEN** 以往旧数据 **THEN** 系统 **SHALL** 不处理（不迁移、不重命名、不删除）

2. **WHEN** 查询支付方式列表 **THEN** 系统 **SHALL** 同时显示新数据和旧数据

3. **WHEN** 编辑旧数据 **THEN** 系统 **SHALL** 提示用户该数据为旧数据，建议创建新的支付方式

#### 具体要求

- [x] 6.1 旧数据识别：`erpnext_payment` 字段为空或不符合新命名规则的数据视为旧数据
- [x] 6.2 旧数据只读策略（不允许修改）
- [x] 6.3 旧数据查询过滤（可选）

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

- [x] URL 使用 snake_case 命名（如：`/api/v1/admin/payment_methods`）
- [x] data 字段必须是对象，不能是 null 或数组
- [x] 分页信息统一放在 meta 中
- [x] 响应格式：`{code, message, data{}}`
- [x] 参考: `.cursor/rules/api.mdc` - API 设计规范

### 数据库设计要求

- [x] 使用现有数据表 `ttpos_payment_method`，无需创建新表
- [x] 参考: `.cursor/rules/database.mdc` - 数据库开发规范

**现有表结构**：

```sql
-- 现有支付方式表（已存在）
CREATE TABLE IF NOT EXISTS `ttpos_payment_method` (
    `id` INT(11) NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT NOT NULL DEFAULT 0 COMMENT '支付方式ID',
    `headquarter_uuid` BIGINT NOT NULL DEFAULT 0 COMMENT '总部uuid，0表示本店创建，>0表示从总部同步',
    `name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '支付方式名称',
    `code` INT(11) NOT NULL DEFAULT 0 COMMENT '支付方式代号',
    `payment_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '支付名称',
    `source` INT(10) NOT NULL DEFAULT 1 COMMENT '来源 0-系统 1-手动 2-LianLianPay',
    `logo_file_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'logo图片ID',
    `qrcode_file_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '二维码图片ID',
    `fee_percent`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '手续费百分比,取值范围0-1',
    `is_show_cashier` INT(10) NOT NULL DEFAULT 0 COMMENT '0-不显示 1-收银机结账显示',
    `is_show_assistant` INT(10) NOT NULL DEFAULT 0 COMMENT '0-不显示 1-点餐助手结账显示',
    `is_show_member_recharge` INT(10) NOT NULL DEFAULT 0 COMMENT '0-不显示 1-收银机会员充值显示',
    `status` INT(10) NOT NULL DEFAULT 0 COMMENT '状态 0-禁用 1-启用',
    `sort` INT(11) NOT NULL DEFAULT 0 COMMENT '排序',
    `default_img` TEXT COMMENT '默认图片',
    `erpnext_payment` VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'ERPNext支付方式',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`),
    INDEX `idx_headquarter_uuid` (`headquarter_uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '支付方式表';
```

**字段映射关系**：

| 需求字段 | 现有表字段 | 说明 |
|---------|-----------|------|
| `type` | `source` | 类型：0=系统(System)，1=手动(General)，2=LianLianPay |
| `mode_of_payment_id` | `erpnext_payment` | ERP Mode of Payment ID（按命名规则生成） |
| `payment_method` | `payment_name` | 支付方式名称（如 Alipay, WeChat Pay） |
| `status` | `status` | 状态：0=禁用，1=启用 |
| `sort_order` | `sort` | 排序顺序 |
| `icon` | `logo_file_uuid` | 支付图标（通过 file_uuid 关联） |
| `branch` | - | 从 shop/company 关联表获取 |
| `company` | - | 从 shop/company 关联表获取 |
| `abbr` | - | 从 shop/company 关联表获取（CompanyAbbr） |

**说明**：
- `branch`、`company`、`abbr` 信息从 `ttpos_company_setting` 表关联获取，无需在 `ttpos_payment_method` 表中新增字段
- `erpnext_payment` 字段用于存储按命名规则生成的 Mode of Payment ID
- `source` 字段已支持类型区分（0-系统，1-手动，2-LianLianPay）
- 同步日志可考虑使用现有日志表或创建独立的同步日志表（可选）

### 性能要求

- [x] 本地响应时间 < 200ms
- [x] ERP 同步操作使用异步队列（不阻塞主流程）
- [x] 数据库查询优化（使用索引）
- [x] 缓存支付方式列表（Redis，TTL 10 分钟）
- [x] 并发处理（使用 UUID 锁防止重复同步）

### 测试要求

- [x] Service 层测试覆盖率 ≥ 70%
- [x] Repository 层测试覆盖率 ≥ 80%
- [x] **Payment 相关模块测试覆盖率 100%**（高风险）
- [x] 集成测试覆盖核心流程（创建、同步、删除）
- [x] API 测试覆盖所有接口
- [x] ERP 同步测试（Mock ERP API）
- [x] 参考: `.cursor/rules/go-main.mdc` - 测试规范

### 国际化要求

- [x] 支持 10 种语言（中文、英文、日语、韩语等）
- [x] 所有文案使用多语言实现
- [x] 支付方式名称支持多语言配置
- [x] 参考: `main/i18n/` - 国际化配置

### 安全要求

- [x] 所有 API 需要身份验证（JWT Token）
- [x] 权限控制（只有管理员可以管理支付方式）
- [x] 敏感数据加密存储（如有）
- [x] SQL 注入防护（使用参数化查询）
- [x] XSS 防护（前端输入校验）
- [x] CSRF 防护（Token 验证）
- [x] ERP API 调用使用 HTTPS
- [x] 参考: `.cursor/rules/security.mdc` - 安全开发规范

### 可靠性要求

- [x] ERP 同步失败时优雅降级（本地数据不受影响）
- [x] 事务管理（保证数据一致性）
- [x] 错误日志记录（使用 Logger）
- [x] 同步重试机制（最多 3 次）
- [x] 监控告警（同步失败率超过 10% 时告警）

---

## 验收标准

### 功能验收

1. **支付方式 CRUD API**：提供完整的后端 API 接口（创建、查询、更新、删除）
2. **命名规则**：自动生成的 Mode of Payment ID 符合命名规范
3. **TTPOS → ERP 同步**：在 TTPOS 操作后成功同步到 ERP
4. **ERP → TTPOS 同步**：从 ERP 同步支付方式到 TTPOS
5. **类型识别**：正确识别和处理 LianLianPay、通用、系统默认三种类型
6. **软删除**：删除操作使用软删除，ERP 侧禁用而不删除
7. **Company 账号添加**：添加支付方式后自动将 Company 加入账号配置
8. **同步日志**：所有同步操作有完整的日志记录
9. **旧数据处理**：旧数据不处理，不影响新功能

### 测试验收

1. **单元测试**: 覆盖率达标（Service ≥ 70%, Repository ≥ 80%）
2. **API 测试**: 所有接口测试通过
3. **集成测试**: 端到端流程测试通过（创建 → 同步 → 查询 → 删除）
4. **ERP 同步测试**: Mock ERP API 测试通过
5. **性能测试**: 响应时间 < 200ms

### 文档验收

1. **技术文档**: design.md 完整且准确
2. **API 文档**: API 接口文档完整（包含请求/响应示例）
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
- 遵循 `.cursor/rules/go-main.mdc`

### 业务约束

- **不处理旧数据**：旧 PHP 管理端的支付配置数据不在本需求范围内，不提供迁移工具，不处理
- **ERP 为主数据源**：同步冲突时以 ERP 数据为准
- **软删除策略**：TTPOS 使用软删除，ERP 使用禁用状态
- **命名规则不可变**：Mode of Payment ID 生成后不可修改
- **Company 账号自动添加**：添加支付方式后必须将当前公司加入账号配置

### 资源约束

- 开发时间: 5-8 天
- Story Point: 待技术评审确认（建议拆分为多个子任务，每个 ≤ 5 SP）

---

## 依赖关系

### 技术依赖

- `gorm.io/gorm` - ORM 框架
- `github.com/gin-gonic/gin` - Web 框架
- `github.com/streadway/amqp` - RabbitMQ 客户端
- `github.com/go-redis/redis/v8` - Redis 客户端
- `main/app/service/rpc/erp` - ERP RPC 服务

### 服务依赖

- **Main → ttpos-bmp (ttpos-erp)**: gRPC 调用（Mode of Payment 同步）
- **Admin (Vue) → Main (Go)**: HTTP API 调用
- **Main → RabbitMQ**: 异步同步队列
- **Main → Redis**: 缓存支付方式列表

### 业务依赖

- **商家配置**：需要已配置 Company 和 Branch
- **ERP 连接**：需要 ERPNext 系统正常运行且网络可达
- **权限配置**：需要配置支付方式管理权限
- **相关 Spec**：`story-ttpos-erp-payment-mode-save`（BMP 层 SaveModeOfPayment 接口）

---

## 风险和缓解

### 风险 1: ERP API 调用失败

**影响**: 高  
**概率**: 中  
**缓解措施**:

- 使用异步队列，避免阻塞主流程
- 设计重试机制（最多 3 次）
- 记录详细的失败日志
- 提供手动重试功能
- 监控告警（失败率超过 10% 时告警）

### 风险 2: 命名冲突

**影响**: 中  
**概率**: 低  
**缓解措施**:

- 序号自动递增（0001, 0002...）
- 唯一性校验（数据库 + 业务层）
- 冲突时自动选择下一个可用序号

### 风险 3: 数据同步不一致

**影响**: 高  
**概率**: 中  
**缓解措施**:

- 以 ERP 为主数据源（冲突时以 ERP 为准）
- 记录同步日志（包括请求和响应数据）
- 提供数据对比工具（手动检查一致性）
- 定期全量同步（每天凌晨）

### 风险 4: Company 账号添加失败

**影响**: 中  
**概率**: 低  
**缓解措施**:

- 在创建 Mode of Payment 后立即调用 CreateModePaymentAccount
- 失败时记录日志并告警
- 提供手动修复工具

---

## 时间表

- **Phase 1 - 数据模型设计**: 0.5 天
  - 分析现有表结构 `ttpos_payment_method`
  - 确认字段映射关系（source、erpnext_payment、payment_name 等）
  - 确认从 `ttpos_company_setting` 获取 branch、company、abbr 的逻辑
  - 无需创建新表或迁移脚本

- **Phase 2 - 后端 API 开发**: 3-4 天
  - 实现 CRUD 接口
  - 实现命名规则生成器
  - 实现字段映射转换器
  - 实现 TTPOS → ERP 同步
  - 实现 ERP → TTPOS 同步
  - 实现 Company 账号自动添加
  - 单元测试和集成测试

- **Phase 3 - 前端页面开发**: 1-2 天
  - 支付方式管理页面
  - 同步日志查看页面

- **Phase 4 - 测试和联调**: 0.5-1 天
  - API 测试
  - ERP 同步测试
  - 性能测试

- **总计**: 5-8 天（SP = 待确认）

---

## 参考资料

### 核心规范

- `.cursor/rules/go-main.mdc` - Go Main 核心约束
- `.cursor/rules/api.mdc` - API 设计规范
- `.cursor/rules/database.mdc` - 数据库开发规范
- `.cursor/rules/security.mdc` - 安全开发规范

### 相关 Spec

- `story-admin-payment-erp-integration` - 新管理端支付管理 ERP 对接（总体方案）
- `story-ttpos-erp-payment-mode-save` - BMP 层 SaveModeOfPayment 接口

### 相关代码

- **新管理端后端**: `main/app/api/v1/admin/handler.go`
- **ERP RPC 服务**: `main/app/service/rpc/erp/erp.go`, `main/app/service/rpc/erp/selling.go`
- **ERP 请求 DTO**: `main/app/dto/req/erpnext.go`
- **旧管理端支付模型**: `admin/app/common/model/pay/PaymentApp.php`
- **BMP ERP 模块**: `ttpos-bmp/app/ttpos-erp/api/selling/`
- **BMP SaveModeOfPayment**: `ttpos-bmp/app/ttpos-erp/internal/logic/selling/selling.go`

### 外部参考

- [ERPNext Mode of Payment API](https://frappeframework.com/docs/user/en/api/rest)
- [ERPNext Payment Entry API](https://frappeframework.com/docs/user/en/api/rest)

---

## Graphiti & 活动日志

- Related Episode: `待补充`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/王昱/2025-12/2025-12-12.md`
- 提醒：需求评审或范围调整若形成经验，应同步更新 Episode 并在 Proposal/Spec 互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-12-12  
**作者**: 王昱  
**审核者**: 待定
