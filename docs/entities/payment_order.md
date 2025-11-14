# PaymentOrder 实体模型说明

## 基本信息

- **实体名称**: PaymentOrder
- **表名**: ttpos_payment_order
- **所属模块**: main
- **描述**: 支付订单表，管理各种支付方式的支付记录

## 字段说明

| 字段名 | 类型 | 说明 | 备注 |
|--------|------|------|------|
| Uuid | uint64 | UUID | 主键标识 |
| PaymentMethodName | string | 支付类型名称 | NOT NULL |
| PaymentMethodUuid | uint64 | 支付类型ID | NOT NULL |
| PaymentFeePercent | float64 | 支付手续费百分比 | NOT NULL |
| RelatedType | int | 关联订单类型 | 0-销售订单；1-充值订单 |
| RelatedUuid | uint64 | 关联的充值订单、销售订单ID | NOT NULL |
| CurrencyUnit | string | 货币单位 | NOT NULL |
| PaymentAmount | float64 | 支付金额 | NOT NULL |
| PaymentCommissionFee | float64 | 支付手续费 | 支付金额*支付手续费百分比 |
| Amount | float64 | 实收金额 | 实收金额=支付金额+支付手续费 |
| TransactionNumber | string | 交易号 | NOT NULL |
| Status | int | 支付状态 | 0-未支付 1-已支付 2-已退款 3-支付异常 |
| StatusReason | string | 支付状态原因 | NOT NULL |
| BalanceAmount | float64 | 主账户金额 | 用于反结账时退款 |
| GiftBalanceAmount | float64 | 赠送帐户金额 | 用于反结账时退款 |

## 关联关系

### 关联实体
- **PaymentMethod** → 支付方式（通过 PaymentMethodUuid 关联）
- **MemberRechargeOrder** → 会员充值订单（通过 RelatedUuid 关联，当 RelatedType=1 时）
- **ReturnOrderAmounts** → 退款单金额（通过 PaymentOrderUuid 关联）
- **RefundOrder** → 退款单（通过 PaymentOrderUuid 关联）

## 业务方法

### 退款相关
- **NewRefundOrder()**: 创建退款单
- **GetCanReturnAmount()**: 获取可退款金额（支付金额-已退款金额）

### 支付方式相关
- **GetSource()**: 获取支付方式来源
- **GetSourceText(language)**: 获取来源文本
- **Cancel()**: 取消支付订单

### 其他方法
- **SetBaseModel(baseModel)**: 设置基础模型
- **SetNil()**: 清空关联对象

## 数据流分析

### 数据来源
- 销售订单支付
- 会员充值支付
- 各种支付方式的支付记录

### 数据流向
1. **支付创建流程**:
   - 创建支付订单关联到销售订单或充值订单
   - 根据支付方式计算手续费
   - 设置支付金额和实收金额

2. **支付处理流程**:
   - 调用第三方支付接口
   - 更新支付状态和交易号
   - 处理支付结果

3. **退款流程**:
   - 创建退款单关联到支付订单
   - 处理退款金额分配（主账户/赠送账户）
   - 更新支付订单状态为已退款

4. **反结账流程**:
   - 创建反结账退款单
   - 恢复会员余额
   - 记录退款原因

### 业务场景
- 多种支付方式支持（现金、余额、微信、支付宝等）
- 手续费计算和收取
- 支付状态管理
- 退款和反结账处理
- 会员余额支付分离（主账户/赠送账户）

## 索引建议

- 主键索引: Uuid
- 唯一索引: TransactionNumber（交易号唯一）
- 普通索引: PaymentMethodUuid（支付方式查询）
- 普通索引: RelatedType + RelatedUuid（关联订单查询）
- 普通索引: Status（状态查询）
- 普通索引: CreatedAt（时间范围查询）

## 注意事项

1. 支持多种支付方式，包括第三方支付和内部支付
2. 手续费计算支持百分比和固定金额
3. 余额支付支持主账户和赠送账户分离
4. 支持退款和反结账操作
5. 所有金额字段使用decimal类型保证精度
6. 支付状态需要及时同步更新