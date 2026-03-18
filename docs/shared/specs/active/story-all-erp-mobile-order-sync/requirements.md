# story-all-erp-mobile-order-sync 需求文档

## 基本信息

| 项目 | 内容 |
|------|------|
| Spec ID | story-all-erp-mobile-order-sync |
| Level | story |
| 创建人 | weifashi |
| 创建日期 | 2026-03-11 |
| DooTask 关联 | #40303 |
| 父 Spec | story-all-erp-sales-invoice-reform |
| 目标版本 | V2.20.0 |

## 审核状态

| 项目 | 内容 |
|------|------|
| 审核状态 | 已通过 |
| 审核人 | weifashi |
| 审核日期 | 2026-03-11 |

---

## 用户故事

**作为** ERP 商家/财务人员
**我想** 会员端和扫码点餐(Mobile)产生的订单也能同步进入 ERP，生成 Sales Invoice、Payment Entry 和 Stock Entry
**以便于** 在 ERP 中看到全渠道（POS + 会员 + 扫码点餐）的完整销售数据，确保财务数据完整性和全渠道营收可见

## 背景

当前会员端和扫码点餐(Mobile)产生的订单下到 POS 后，未能同步进入 ERP，导致**财务数据不完整**，财务人员无法在 ERP 中看到全渠道的销售数据。

父 Spec `story-all-erp-sales-invoice-reform` 已定义了 POS 端结账后生成 Sales Invoice + Payment Entry + Stock Entry 的完整流水线，但该流水线仅覆盖 POS 直接下单的场景，未涵盖会员/扫码点餐渠道的订单。

---

## 功能需求

### Requirement 1: 接单后推送 ERP（开启接单场景）

**用户故事**: 作为财务人员，我想会员/扫码点餐订单在 POS 接单成功后立即推送到中台队列进入 ERP，以便于实时记录收入

**流程**:
```
[Mobile/会员端] 顾客下单 → [POS 收银端] 收到订单 → 接单确认 → 接单成功 → 推送到中台队列 → 生成 ERP 单据
```

#### 验收标准

1. **WHEN** 会员/扫码点餐订单到达 POS 且 POS **开启接单** **THEN** 订单 **SHALL** 在接单成功后推送到中台队列进入 ERP
2. **WHEN** POS 设置为自动接单 **THEN** 自动接单完成后 **SHALL** 同样触发 ERP 推送

### Requirement 2: 结账后推送 ERP（未开启接单场景）

**用户故事**: 作为财务人员，我想未开启接单的会员/扫码点餐订单在结账完成后推送到 ERP，以便于确保所有渠道订单都进入财务系统

**流程**:
```
[Mobile/会员端] 顾客下单 → [POS 收银端] 直接进入订单列表 → 结账完成 → 推送到中台队列 → 生成 ERP 单据
```

#### 验收标准

1. **WHEN** 会员/扫码点餐订单到达 POS 且 POS **未开启接单** **THEN** 订单 **SHALL** 在结账完成后推送到中台队列进入 ERP

### Requirement 3: 新旧方案过渡（基于班次状态）

**用户故事**: 作为系统运维人员，我想系统基于班次状态平滑切换新旧 ERP 方案，以便于确保旧班次数据一致性且新班次立即享受新方案

#### 方案对比

| 条件 | 方案 | ERP 单据 |
|------|------|----------|
| **旧班次未关班** | 旧方案 | POS Invoice + Opening/Closing Entry |
| **新班次产生** | 新方案 | Sales Invoice + Payment Entry + Stock Entry（合并扣减）|

#### 验收标准

1. **WHEN** 当前运行的是**旧班次**（未关班） **THEN** 所有订单（含会员/扫码点餐） **SHALL** 走旧方案（POS Invoice + Opening/Closing Entry）
2. **WHEN** **新班次产生**（旧班次关班后重新开班） **THEN** 所有订单（含会员/扫码点餐） **SHALL** 走新方案（Sales Invoice + Stock Entry）
3. **WHEN** 系统切换到新方案后 **THEN** 系统 **SHALL** 不可回退到旧方案，后续所有班次均走新方案

### Requirement 4: Sales Invoice 订单来源标识

**用户故事**: 作为财务人员，我想在 ERP 的 Sales Invoice 中清楚区分订单来源渠道，以便于按渠道分析销售数据

#### 额外字段映射

| ERP 字段 | 来源/规则 | 说明 |
|---------|----------|------|
| `order_source_uuid` | 订单来源 UUID | 标识为会员端/扫码点餐来源 |
| `order_source_name` | 订单来源名称 | 如"会员扫码点餐"、"Mobile 点餐" |
| `customer` | 会员信息 | 若关联会员，使用会员对应的 ERP Customer；否则使用 POS Profile 默认客户 |

#### 验收标准

1. **WHEN** 会员/扫码点餐订单走新方案 **THEN** Sales Invoice **SHALL** 包含正确的 `order_source_uuid` 和 `order_source_name`，标识订单来源
2. **WHEN** 订单关联了会员 **THEN** Sales Invoice 的 `customer` **SHALL** 使用会员对应的 ERP Customer
3. **WHEN** 订单未关联会员 **THEN** Sales Invoice 的 `customer` **SHALL** 使用 POS Profile 默认客户

---

## 非功能需求

### 测试要求

- [ ] 单元测试覆盖率 >= 80%
- [ ] 集成测试覆盖接单/未接单两种场景
- [ ] 新旧方案切换边界测试

### 性能要求

- [ ] 推送到中台队列延迟 < 1 秒
- [ ] ERP 单据生成遵循父 Spec 的异步+重试机制（5 分钟间隔，最多 3 次）

---

## 技术约束

- Go 版本: 1.23+
- 框架: Gin + GORM (Main), GoFrame v2.x (BMP)
- 分层架构: API -> Service -> Repository -> Model
- 必须遵循 CLAUDE.md 和 .cursor/rules/go-main.mdc 规范
- ERP 同步逻辑复用父 Spec `story-all-erp-sales-invoice-reform` 已建立的 Sales Invoice + Payment Entry + Stock Entry 流水线
- 新旧方案切换以班次为边界，不可逆

---

## 涉及终端

- Mobile/会员端（下单入口）
- POS 收银端（接单/结账触发 ERP 推送）
- BMP 中台（ttpos-erp 服务处理队列消息，生成 ERP 单据）

## 涉及模块

- **Main 模块**: 订单服务（接单/结账时推送队列）、班次服务（新旧方案判断）
- **BMP 模块**: ttpos-erp 服务（处理会员/扫码订单的 Sales Invoice 生成，增加来源字段）
- **中台队列**: RocketMQ 消息处理

## 关联文档

- 父 Spec: [story-all-erp-sales-invoice-reform](../story-all-erp-sales-invoice-reform/requirements.md)
- DooTask: #40303

---

## 风险和缓解

### 风险 1: 新旧方案切换边界订单丢失

**影响**: 高
**缓解措施**: 以班次为原子切换单位，旧班次内所有订单（含中途进入的扫码订单）统一走旧方案，确保数据一致性

### 风险 2: 接单/未接单场景判断不准确

**影响**: 中
**缓解措施**: 接单设置读取 POS 配置，通过订单状态流转（已接单/已结账）明确触发时机

---

**版本**: v1.0.0
**创建日期**: 2026-03-11
