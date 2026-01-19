# ERP 支付方式启用状态（enabled）同步需求提案

> 本文档用于需求发起阶段，经团队评审后创建正式 Spec。

---

## 📋 提案信息

| 项目       | 内容     |
| ---------- | -------- |
| **提案人** | rikugun |
| **日期**   | 2025-12-12 |
| **目标版本** | v2.11.x |
| **状态**   | 待评审   |
| **关联任务** | - |
| **关联 Spec** | `docs/shared/specs/archived/v2.12/story-ttpos-erp-mode-of-payment-enabled/requirements.md` |

---

## 🎯 背景和动机

### 问题描述

- 目前 `GetModeOfPaymentList` 返回的 `ModeOfPayment` 仅包含 `name`，缺少“是否启用”的状态信息，无法支撑前端/同步侧做“隐藏/禁用支付方式”的一致性逻辑。
- 目前 `SaveModeOfPaymentReq` 仅能保存/同步支付方式的命名与类型信息，无法对 ERPNext 中支付方式启用状态进行更新。

### 业务价值

- 支撑商户/运营在 ERP 侧统一维护支付方式启用状态，并在 TTPOS 使用侧（列表、选择器、同步）保持一致。
- 降低因支付方式“已停用但仍可被选择/展示”导致的业务误操作风险。

### 目标用户

- [ ] 收银员
- [x] 商户管理员 / 运营
- [ ] 厨房人员
- [ ] 顾客
- [x] 其他: ERP/TTPOS 集成维护者

---

## 💡 解决方案概述

### 方案描述

在 `ttpos-bmp/app/ttpos-erp/manifest/protobuf/selling/selling.proto` 中：

1) `SaveModeOfPaymentReq` 新增 `enabled`（**可选**）参数：
- 调用方传入时，服务端将该值更新到 ERPNext 的 Mode of Payment 启用状态。
- 调用方不传入时，服务端**不更新** ERP 现有启用状态（保持兼容与最小影响）。

2) `ModeOfPayment` 增加 `enabled` 字段：
- `GetModeOfPaymentList` 返回时携带 ERP 中该支付方式的启用状态。

### 核心功能点

1. Protobuf 变更
   - `ModeOfPayment` 增加 `enabled` 字段。
   - `SaveModeOfPaymentReq` 增加 `enabled` 字段（可选）。

2. 语义与兼容
   - `SaveModeOfPaymentReq.enabled`：
     - **IF** 字段存在 **THEN** 更新 ERP 启用状态。
     - **IF** 字段不存在 **THEN** 不更新启用状态。
   - `GetModeOfPaymentList`：
     - 服务端需确保返回的 `enabled` 为真实值；若 ERP 未返回该字段或无法读取，默认按 `true` 处理并记录审计日志（避免“默认为 false 导致全量变为禁用”的误判）。

3. ERP 对接
   - 保存/同步支付方式时，将 `enabled` 写入 ERP 对应字段（如 ERPNext Doctype `Mode of Payment.enabled`）。

### 影响范围

**涉及终端**：
- [ ] POS 收银端
- [x] Shop 商家管理端（若使用支付方式列表/配置）
- [ ] KDS 厨显端
- [ ] QDS 排号叫号端
- [ ] Assistant 助手端
- [ ] Tablet 平板端
- [ ] Mobile 扫码端
- [ ] Menu 电子菜单端
- [ ] Member 会员端
- [x] 其他: ERP/TTPOS 支付同步服务

**涉及模块**：
- [ ] UI 组件
- [x] API 接口（protobuf/gRPC）
- [ ] 数据模型
- [x] 业务逻辑（保存/查询映射）
- [x] 第三方集成（ERPNext）
- [ ] 其他: ________

---

## 📊 初步评估

### 技术复杂度

- [ ] **低**：纯 UI 调整，无业务逻辑变更
- [x] **中**：需要服务端改造 + 与 ERP 联调
- [ ] **高**：涉及架构调整、第三方集成、复杂算法

### 工作量预估

- **预计天数**: 2 天（proto 变更 + 服务端实现 + ERP 联调 + 回归）
- **预估 SP**: 2（待技术评审确认）

### 风险识别

**潜在风险**：
1. `proto3 optional` 的代码生成与现有工具链版本不一致，导致无法判断字段是否“传入”。
2. ERPNext 字段命名/类型差异（例如 `enabled` 为 int/bool）导致写入失败。
3. 新增 `ModeOfPayment.enabled` 后，若服务端未显式赋值可能被客户端误解为“禁用”。

**缓解措施**：
1. 明确采用 `proto3 optional` 的方案（在 Spec/设计阶段锁定），并在 CI 里固定 protoc / buf / go 生成版本。
2. 增加 ERP 写入的字段映射与参数校验，失败记录审计日志并返回明确错误。
3. `GetModeOfPaymentList` 返回时强制填充 `enabled`（读取失败默认 true + 记录日志）。

---

## 🔗 相关资源

### 参考需求

- 接口文件: `ttpos-bmp/app/ttpos-erp/manifest/protobuf/selling/selling.proto`
- 相关提案: `docs/team/proposals/2025-12/2025-12-10-erp-payment-mode-save.md`

### 相关文档

- API 设计规范: `.cursor/rules/api.mdc`
- BMP Go 规范: `ttpos-bmp/.cursor/rules/go-rules.mdc`

---

## 🤝 需求评审

### 评审参与人

| 角色         | 姓名   | 签名/日期 |
| ------------ | ------ | --------- |
| 产品经理     | {姓名} |           |
| 技术负责人   | {姓名} |           |
| 开发代表     | {姓名} |           |
| 测试代表     | {姓名} |           |
| UI/UX 设计师 | {姓名} |           |

### 评审结论

- [ ] ✅ **批准**：进入技术方案设计阶段
- [ ] 🔄 **修改后批准**：需补充以下内容
- [ ] ❌ **拒绝**：不符合产品规划或优先级

**评审意见**：

```
[记录评审会议的关键讨论和决策]
```

**下一步行动**：

- [ ] 创建 Spec：`story-ttpos-erp-mode-of-payment-enabled`
- [ ] 分配负责人：{姓名}
- [ ] 目标 Sprint：Sprint {N}

---

## 📝 附录

### User Story（初稿）

作为 ERP/TTPOS 集成维护者  
我想 在支付方式同步/列表中引入启用状态（enabled），并支持可选更新到 ERP  
以便于 在 TTPOS 侧准确展示/过滤禁用的支付方式，减少误操作。

### AC 验收标准（初稿）

1. **WHEN** 调用 `SaveModeOfPayment` 且携带 `enabled=true/false` **THEN** 系统 **SHALL** 将启用状态更新到 ERP 对应支付方式。
2. **IF** 调用 `SaveModeOfPayment` 未携带 `enabled` **THEN** 系统 **SHALL** 不改变 ERP 中该支付方式启用状态。
3. **WHEN** 调用 `GetModeOfPaymentList` **THEN** 系统 **SHALL** 返回每个支付方式的 `enabled` 状态，且与 ERP 一致。
4. **IF** ERP 未返回/无法读取启用状态 **THEN** 系统 **SHALL** 默认按启用处理并记录审计日志（避免误判为禁用）。

### 线框图/原型（可选）

[暂无]
