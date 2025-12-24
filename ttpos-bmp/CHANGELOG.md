# Changelog

All notable changes to the ttpos-bmp project will be documented in this file.

## [Unreleased]

### Enhanced
- **ClosePosEntry 接口支持通过 PaymentID 指定支付方式**
  - Protobuf: ClosePosEntryDetail 新增 `payment_id` 字段（optional）
  - Logic: 当 `payment_id` 不为空时，自动调用 GetModeOfPayment 查询对应的 `mode_of_payment`
  - 向后兼容: 保持原有 `mode_of_payment` 字段可用
  - 参数校验: `payment_id` 和 `mode_of_payment` 至少需要提供一个
  - 关联 Spec: task-erp-close-pos-entry-payment-id
