# story-ttpos-erp-mode-of-payment-enabled / ERP 支付方式更新（SaveModeOfPayment 扩展）需求说明

> 基于 Proposal `docs/team/proposals/2025-12/erp-mode-of-payment-update.md`，扩展 SaveModeOfPayment 接口支持更新已有支付方式（如 enabled 状态）。

## 1. 概要

| 字段 | 说明 |
| --- | --- |
| 提案链接 | `docs/team/proposals/2025-12/erp-mode-of-payment-update.md` |
| 目标版本 | v2.11.x |
| 所属端 | ttpos-bmp / erp 集成 |
| 主要角色 | ERP/TTPOS 集成维护者、商户管理员 |
| 状态 | 已通过 |

## 2. 背景与目标

- 现状：`SaveModeOfPayment` 接口仅支持**创建**新的支付方式，无法对已存在的支付方式进行更新。
- 现状：当需要修改已有支付方式的 `enabled` 状态时，缺少标准接口支持，只能手动在 ERP 中操作。
- 目标：
  - 支持通过 API 更新已有支付方式的属性（如 enabled 状态）。
  - 提升支付方式管理的自动化程度，减少人工操作 ERP。
  - 为后续扩展更多可更新字段（如渠道配置）奠定基础。

## 3. 业务/功能需求

1. SaveModeOfPayment 接口语义扩展
   - 接口：`SellingService.SaveModeOfPayment`（`selling.proto`）
   - 入参：`SaveModeOfPaymentReq` 增加 `name`（可选）。
   - 语义：
     - **IF** 请求包含 `name` **THEN** 执行**更新**操作，根据 `name` 查找已有支付方式并更新指定字段。
     - **IF** 请求不包含 `name` **THEN** 执行**创建**操作（保持现有行为）。

2. 更新操作支持的字段
   - `enabled`：更新支付方式的启用状态
   - **IF** `name` 存在且 `enabled` 有值 **THEN** 更新 ERP 中该支付方式的启用状态。
   - **IF** `name` 存在但 `enabled` 为空 **THEN** 不更新启用状态。

3. Protobuf 可选字段语义约束（强制）
   - `name` 使用 `optional string`，以区分"未传"和"传空字符串"。
   - `enabled` 使用 `optional bool`，以区分"未传"、"传 true"和"传 false"。

4. 错误处理
   - **IF** `name` 对应的支付方式不存在 **THEN** 返回"支付方式不存在"错误。
   - **IF** `name` 对应的支付方式不属于当前 `company_abbr` **THEN** 返回"无权限修改"错误。

5. 数据兼容与降级
   - 对历史客户端：不携带 `name` 的 `SaveModeOfPayment` 调用必须保持原有行为（创建新支付方式）。
   - 对新客户端：携带 `name` 则执行更新操作。

## 4. 约束与假设

- ERP 存在可更新的支付方式启用字段（ERPNext Doctype `Mode of Payment.enabled` 或等价字段）。
- 本需求仅覆盖"接口扩展"，不包含管理后台 UI 的更新交互改造。

## 5. 验收标准（AC）

1. **WHEN** 调用 `SaveModeOfPayment` 且携带 `name` **THEN** 系统 **SHALL** 查找该支付方式并执行更新操作。
2. **IF** `name` 对应的支付方式不存在 **THEN** 系统 **SHALL** 返回"支付方式不存在"错误。
3. **WHEN** 调用 `SaveModeOfPayment` 未携带 `name` **THEN** 系统 **SHALL** 执行创建操作（保持现有行为）。
4. **WHEN** 更新操作携带 `enabled` **THEN** 系统 **SHALL** 更新 ERP 中该支付方式的启用状态。
5. **IF** 更新操作未携带 `enabled` **THEN** 系统 **SHALL** 不更新启用状态。

## 6. 不在范围

- 批量更新支付方式。
- 支付方式 UI 配置页面改造（仅接口与同步逻辑）。
- 更新除 `enabled` 外的其他字段（预留扩展性）。

## 7. 里程碑

- 需求评审通过：T+0
- Protobuf 变更与服务端实现：T+1
- ERP 联调与验收：T+2
