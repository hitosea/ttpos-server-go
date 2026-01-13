# Changelog

All notable changes to the ttpos-bmp project will be documented in this file.

## [Unreleased]

### Added
- **LINE MAN 订单状态更新 Webhook**
  - 端点: `POST /v1/partners/{partnerId}/stores/{storeId}/order/status`
  - 接收 LINE MAN 订单完成/取消通知，更新订单状态到数据库
  - 状态映射: `FINISH` → `COMPLETED`, `CANCELED` → `CANCELLED`
  - 通过 RocketMQ 通知 Main 模块（Topic: `takeout_grab_order`）
  - 幂等性处理: 重复请求不重复更新
  - 关联 Spec: tech-takeout-lineman-order-status-update

### Enhanced
- **OpenPosEntry 接口支持通过 PaymentID 指定支付方式**
  - Protobuf: OpenPosEntryDetail 新增 `payment_id` 字段（optional）
  - Logic: 当 `payment_id` 不为空时，自动调用 GetModeOfPayment 查询对应的 `mode_of_payment`
  - 向后兼容: 保持原有 `mode_of_payment` 字段可用
  - 参数校验: `payment_id` 和 `mode_of_payment` 至少需要提供一个
  - 关联 Spec: task-erp-open-pos-entry-payment-id

- **ClosePosEntry 接口支持通过 PaymentID 指定支付方式**
  - Protobuf: ClosePosEntryDetail 新增 `payment_id` 字段（optional）
  - Logic: 当 `payment_id` 不为空时，自动调用 GetModeOfPayment 查询对应的 `mode_of_payment`
  - 向后兼容: 保持原有 `mode_of_payment` 字段可用
  - 参数校验: `payment_id` 和 `mode_of_payment` 至少需要提供一个
  - 关联 Spec: task-erp-close-pos-entry-payment-id
