# ERP 支付方式更新（SaveModeOfPayment 扩展）需求提案

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
| **关联 Spec** | -      |

---

## 🎯 背景和动机

### 问题描述

- 目前 `SaveModeOfPayment` 接口仅支持**创建**新的支付方式，无法对已存在的支付方式进行更新。
- 当需要修改已有支付方式的 `enabled` 状态时，缺少标准接口支持，只能手动在 ERP 中操作。
- 业务场景中经常需要临时启用/禁用某个支付方式，缺乏 API 化管理能力。

### 业务价值

- 支持通过 API 更新已有支付方式的属性（如 enabled 状态）。
- 提升支付方式管理的自动化程度，减少人工操作 ERP。
- 为后续扩展更多可更新字段（如渠道配置）奠定基础。

### 目标用户

- [ ] 收银员
- [x] 商户管理员 / 运营
- [ ] 厨房人员
- [ ] 顾客
- [x] 其他: ERP/TTPOS 集成维护者

---

## 💡 解决方案概述

### 方案描述

在 `SaveModeOfPaymentReq` 中增加可选的 `name` 字段：

- **IF** 请求包含 `name` **THEN** 执行**更新**操作，根据 `name` 查找已有支付方式并更新指定字段。
- **IF** 请求不包含 `name` **THEN** 执行**创建**操作（保持现有行为）。

更新操作支持的字段：
- `enabled`：更新支付方式的启用状态

### 核心功能点

1. Protobuf 变更
   - `SaveModeOfPaymentReq` 增加 `name` 字段（可选）

2. 服务端语义
   - `name` 存在：查找并更新已有支付方式
   - `name` 不存在：创建新支付方式（现有逻辑不变）

3. 更新操作支持字段
   - `enabled`：当 `name` 存在且 `enabled` 有值时，更新 ERP 中该支付方式的启用状态

### 影响范围

**涉及终端**：
- [ ] POS 收银端
- [x] Shop 商家管理端（若使用支付方式管理功能）
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
- [x] 业务逻辑（保存/更新分支）
- [x] 第三方集成（ERPNext）
- [ ] 其他: ________

---

## 📊 初步评估

### 技术复杂度

- [x] **低**：在现有接口基础上增加字段和分支逻辑
- [ ] **中**：需要前后端联调，基础业务逻辑
- [ ] **高**：涉及架构调整、第三方集成、复杂算法

### 工作量预估

- **预计天数**: 1 天（proto 变更 + 服务端分支逻辑 + 联调）
- **预估 SP**: 1（待技术评审确认）

### 风险识别

**潜在风险**：
1. `name` 不存在时更新操作应返回明确错误，避免静默失败。
2. 更新操作需确保权限校验，防止越权修改其他公司/分店的支付方式。

**缓解措施**：
1. 更新时先查询 `name` 是否存在，不存在则返回"支付方式不存在"错误。
2. 更新时校验 `name` 对应的支付方式是否属于当前 `company_abbr`。

---

## 🔗 相关资源

### 参考需求

- 接口文件: `ttpos-bmp/app/ttpos-erp/manifest/protobuf/selling/selling.proto`
- 相关 Spec: `docs/shared/specs/active/story-ttpos-erp-mode-of-payment-enabled/`

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

- [ ] 创建 Spec：`story-ttpos-erp-mode-of-payment-update`
- [ ] 分配负责人：{姓名}
- [ ] 目标 Sprint：Sprint {N}

---

## 📝 附录

### User Story（初稿）

作为 ERP/TTPOS 集成维护者  
我想 通过 SaveModeOfPayment 接口更新已有支付方式的属性（如 enabled）  
以便于 统一管理支付方式的创建与更新，减少人工操作 ERP。

### AC 验收标准（初稿）

1. **WHEN** 调用 `SaveModeOfPayment` 且携带 `name` **THEN** 系统 **SHALL** 查找该支付方式并执行更新操作。
2. **IF** `name` 对应的支付方式不存在 **THEN** 系统 **SHALL** 返回"支付方式不存在"错误。
3. **WHEN** 调用 `SaveModeOfPayment` 未携带 `name` **THEN** 系统 **SHALL** 执行创建操作（保持现有行为）。
4. **WHEN** 更新操作携带 `enabled` **THEN** 系统 **SHALL** 更新 ERP 中该支付方式的启用状态。
5. **IF** 更新操作未携带 `enabled` **THEN** 系统 **SHALL** 不更新启用状态。

### Proto 变更草案

```protobuf
// SaveModeOfPaymentReq 保存/同步支付方式请求
message SaveModeOfPaymentReq {
  string company_abbr = 1; // 公司简称，必填
  string branch = 2;       // 分支，必填
  string channel = 3;      // 渠道，如 LianLianPay，创建时必填
  string pay_type = 4;     // 支付类型（TTPOS 定义），创建时必填
  optional bool enabled = 5; // 是否启用，可选
  optional string name = 6;  // 支付方式名称，可选：传入时执行更新操作
}
```

### 线框图/原型（可选）

[暂无]
