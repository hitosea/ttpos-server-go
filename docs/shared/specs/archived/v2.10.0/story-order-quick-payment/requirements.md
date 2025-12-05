> ⚠️ **已归档** - 此 Spec 已随 v2.10.0 发布。
>
> - 归档时间: 2025-12-05
> - 归档人: weifashi

# 快捷支付功能 需求文档

> 本文档定义快捷支付功能的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                |
| ----------------- | --------------------------------------------------------------------------------------------------- |
| **来源 Proposal** | [docs/team/proposals/2025-11/quick-payment.md](../../../../team/proposals/2025-11/quick-payment.md) |
| **创建日期**      | 2025-11-17                                                                                          |
| **负责人**        | 李四（Tech Lead）                                                                                   |
| **目标 Sprint**   | Sprint 24                                                                                           |
| **涉及技术栈**    | [x] Go (main/) [ ] Go (ttpos-bmp/) [x] PHP (admin/) [x] Vue (admin/views/)                          |

---

## 📋 概述

快捷支付功能旨在简化 POS 收银流程，通过一键支付和快捷键支持，将平均结账时间从 30 秒缩短至 3 秒，提升收银效率 30%，改善高峰期顾客等待体验。

本功能主要涉及订单支付 API 优化、缓存设计优化、以及管理后台配置功能。

## 🎯 产品对齐

该功能支持公司 2025 年 Q4 的核心目标：

- **提升用户体验**: 缩短结账等待时间，提升顾客满意度
- **增强竞争力**: 与竞品（收钱吧、二维火）的快捷支付功能持平
- **扩大市场份额**: 高效收银功能是商户选择的重要因素

## 📝 用户故事

**作为** 收银员  
**我想** 通过 API 快速完成支付流程  
**以便于** 加快结账速度，减少顾客等待时间

---

## 功能需求

### Requirement 1: 快捷支付 API

**用户故事**: 作为 POS 客户端，我想调用快捷支付 API 一键完成支付，以便于简化支付流程

#### 验收标准

1. **WHEN** 调用 `/api/v1/order/quick_payment` **THEN** 系统 **SHALL** 自动选择默认支付方式完成支付
2. **IF** 订单状态不是"待支付" **THEN** 系统 **SHALL** 返回错误提示"订单状态不允许支付"
3. **WHEN** 支付成功 **THEN** 系统 **SHALL** 更新订单状态并返回支付结果
4. **WHEN** 支付成功 **THEN** 系统 **SHALL** 发布"订单支付完成"事件触发打印小票

#### 具体要求

- [ ] 1.1 创建快捷支付 API 接口 `/api/v1/order/quick_payment`
- [ ] 1.2 支持传入订单 UUID 和支付方式（可选）
- [ ] 1.3 如未传支付方式，使用商户默认支付方式（现金）
- [ ] 1.4 一次性完成：订单状态更新 + 支付记录创建 + 库存扣减 + 积分累计
- [ ] 1.5 支付流程使用事务管理，保证数据一致性
- [ ] 1.6 本地响应时间 < 200ms（不含第三方支付调用）
- [ ] 1.7 支持并发支付（使用订单 UUID 锁）

---

### Requirement 2: 默认支付方式配置

**用户故事**: 作为商户管理员，我想在后台配置默认支付方式，以便于收银员快速选择

#### 验收标准

1. **WHEN** 访问商户设置页面 **THEN** 系统 **SHALL** 显示支付方式配置选项
2. **WHEN** 修改默认支付方式 **THEN** 系统 **SHALL** 保存配置并生效
3. **IF** 商户未配置 **THEN** 系统 **SHALL** 使用"现金"作为默认值

#### 具体要求

- [ ] 2.1 在 `ttpos_company` 表增加 `default_payment_method` 字段
- [ ] 2.2 管理后台增加配置页面（Vue）
- [ ] 2.3 创建配置更新 API
- [ ] 2.4 配置数据缓存到 Redis（Key: `ttpos:company:config:{company_uuid}`）

---

### Requirement 3: 快捷支付记录和统计

**用户故事**: 作为商户管理员，我想查看快捷支付使用情况，以便于了解收银效率

#### 验收标准

1. **WHEN** 访问报表页面 **THEN** 系统 **SHALL** 显示快捷支付使用次数和占比
2. **WHEN** 查看订单详情 **THEN** 系统 **SHALL** 标识是否使用快捷支付

#### 具体要求

- [ ] 3.1 订单表增加 `is_quick_payment` 字段标识
- [ ] 3.2 报表统计 API 支持快捷支付筛选
- [ ] 3.3 订单列表显示快捷支付标识

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 严格遵循 Controller → Service → Repository 分层
- **单一职责原则**: QuickPaymentService 专注快捷支付逻辑，复用 OrderService 和 PaymentService
- **模块化设计**: 支付逻辑可独立测试和复用
- **依赖管理**: QuickPaymentService 依赖 IOrderService 和 IPaymentService 接口
- **遵循规范**:
  - `.cursor/rules/go-main.mdc` - Go Main 开发规范
  - `.cursor/rules/php.mdc` - PHP 开发规范
  - `.cursor/rules/vue.mdc` - Vue 前端规范
  - `.cursor/rules/api.mdc` - API 设计规范
  - `.cursor/rules/database.mdc` - 数据库开发规范

### API 设计要求

- [x] URL 使用 snake_case 命名：`/api/v1/order/quick_payment`
- [x] data 字段必须是对象：`{"data": {"order_uuid": 123, "payment_status": 1}}`
- [x] 响应格式：`{code: 1, message: "success", data: {}}`
- [x] 参考: `.cursor/rules/api.mdc` - API 设计规范

### 数据库设计要求

- [x] `ttpos_company` 表增加 `default_payment_method` 字段（int, default 1 表示现金）
- [x] `ttpos_order` 表增加 `is_quick_payment` 字段（tinyint, default 0）
- [x] 使用数据库迁移文件管理变更
- [x] 字段使用 snake_case
- [x] 参考: `.cursor/rules/database.mdc` - 数据库开发规范

### 性能要求

- [x] 快捷支付 API 本地响应时间 < 200ms（P0）
- [x] 商户配置数据缓存到 Redis，缓存时间 30 分钟
- [x] 订单查询使用 uuid 索引优化
- [x] 并发场景使用 UUID 锁防止重复支付

### 浏览器兼容性（管理后台）

- [x] Chrome 90+
- [x] Safari 14+
- [x] Firefox 88+
- [x] Edge 90+

### 测试要求

- [x] QuickPaymentService 测试覆盖率 ≥ 70%
- [x] **Payment 相关模块测试覆盖率 100%**（高风险，P0）
- [x] Repository 层测试覆盖率 ≥ 80%
- [x] 集成测试覆盖完整支付流程
- [x] 并发测试：10 个并发请求同一订单，只能成功 1 次
- [x] 参考: `.cursor/rules/go-main.mdc` - 测试规范

### 安全要求

- [x] 快捷支付 API 需要身份验证（JWT Token）
- [x] 订单 UUID 验证：订单必须属于当前商户
- [x] 幂等性保证：同一订单只能支付一次
- [x] 参考: `.cursor/rules/security.mdc` - 安全开发规范

### 可靠性要求

- [x] 支付失败时事务回滚，保证数据一致性
- [x] Redis 缓存失败时降级到数据库查询
- [x] 记录详细错误日志（Logger）
- [x] 支付异常时发送告警通知

---

## 验收标准

### 功能验收

1. **快捷支付流程**: 调用 API → 订单状态更新 → 支付记录创建 → 事件发布 → 小票打印
2. **配置功能**: 商户可在后台设置默认支付方式，配置立即生效
3. **统计功能**: 报表显示快捷支付使用情况，订单列表显示快捷支付标识

### 测试验收

1. **单元测试**: Service 和 Repository 覆盖率达标
2. **API 测试**: 所有接口测试通过，包含异常场景
3. **集成测试**: 端到端支付流程测试通过
4. **并发测试**: 并发场景测试通过，无重复支付

### 文档验收

1. **技术文档**: design.md 包含完整的数据库设计、API 设计、缓存设计
2. **API 文档**: docs/shared/api/order_api.md 已更新
3. **数据库文档**: 迁移脚本已创建，表结构文档已更新

---

## 约束条件

### 技术约束

#### Go Main 模块

- 必须使用 Gin 框架
- QuickPaymentService 接口以 `I` 开头：`IQuickPaymentService`
- Service 依赖 IOrderService 和 IPaymentService 接口（不直接依赖 Repository）
- Repository 只持有 db 实例
- 不使用 panic，所有错误返回 error

#### PHP 模块

- 使用 ThinkPHP 6.0
- Controller 不写业务逻辑，调用 Service
- 使用验证器验证参数
- 配置更新使用事务

#### Vue 模块

- 使用 Vue 3 + TypeScript
- 使用 Element Plus 组件库
- API 调用封装到独立文件

### 业务约束

- 快捷支付只支持"待支付"状态的订单
- 默认支付方式限定为：1-现金，2-微信，3-支付宝，4-银行卡
- 快捷支付不支持部分支付（必须全额支付）

### 资源约束

- 开发时间: 2-3 天
- Story Point: 5 (符合 ≤ 5 标准)

---

## 依赖关系

### 技术依赖

- `github.com/gin-gonic/gin` - Web 框架
- `gorm.io/gorm` - ORM
- `github.com/go-redis/redis/v8` - Redis 客户端

### 服务依赖

- **无微服务依赖**：纯 Go Main 模块实现

### 业务依赖

- 依赖现有的订单模块（OrderService）
- 依赖现有的支付模块（PaymentService）
- 依赖现有的事件总线（EventBus）

---

## 风险和缓解

### 风险 1: 并发重复支付

**影响**: 高  
**概率**: 中  
**缓解措施**:

- 使用订单 UUID 锁（SystemLock）防止并发
- 订单状态检查：只有"待支付"状态才允许支付
- 幂等性保证：支付记录使用唯一索引

### 风险 2: 缓存不一致

**影响**: 中  
**概率**: 低  
**缓解措施**:

- 配置更新时主动删除缓存
- 缓存过期时间设置为 30 分钟
- 缓存失败时降级到数据库查询

### 风险 3: 性能瓶颈

**影响**: 中  
**概率**: 低  
**缓解措施**:

- 使用 Redis 缓存商户配置
- 订单查询使用 uuid 索引
- 事件发布使用异步 goroutine

---

## 时间表

- **Phase 1 - 数据库和 API**: 1 天
- **Phase 2 - 后台配置功能**: 0.5 天
- **Phase 3 - 测试和优化**: 1 天
- **总计**: 2.5 天（SP = 5）

---

## 参考资料

### 核心规范

- `.cursor/rules/go-main.mdc` - Go Main 核心约束
- `.cursor/rules/php.mdc` - PHP 核心约束
- `.cursor/rules/vue.mdc` - Vue 开发规范
- `.cursor/rules/api.mdc` - API 设计规范
- `.cursor/rules/database.mdc` - 数据库开发规范
- `.cursor/rules/security.mdc` - 安全开发规范

### 架构文档

- `docs/human/architecture/go-main-architecture.md` - Go Main 架构
- `docs/human/architecture/database-design.md` - 数据库设计
- `docs/human/architecture/event_bus_design.md` - 事件总线设计

### 开发指南

- `docs/human/guides/go-main-development.md` - Go Main 开发指南
- `docs/human/guides/php-development.md` - PHP 开发指南
- `docs/human/guides/api-design-guide.md` - API 设计指南

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 活动日志：`docs/team/activities/2025-11/2025-11-17.md`
- 提醒：当快捷支付功能产出可复用经验或重大决策时，创建 Episode 并在此占位更新。

---

**版本**: v1.0.0  
**创建日期**: 2025-11-17  
**作者**: 后端开发组  
**审核者**: CTO
