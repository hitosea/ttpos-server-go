# SaleBill 实体模型说明

## 基本信息

- **实体名称**: SaleBill
- **表名**: sale_bill
- **所属模块**: model
- **描述**: 销售账单实体，用于记录POS系统的销售交易账单信息

## 字段说明

| 字段名 | 类型 | 说明 | 备注 |
|--------|------|------|------|
| ID | uint | 主键ID | 继承自BaseModel |
| Uuid | uint64 | UUID | 继承自BaseModel |
| CreateTime | int64 | 创建时间 | 继承自BaseModel |
| UpdateTime | int64 | 更新时间 | 继承自BaseModel |
| DeleteTime | int64 | 删除时间 | 继承自BaseModel |
| CompanyUuid | uint64 | 公司UUID | 多租户标识 |
| OrderUuid | string | 订单UUID | 关联销售订单 |
| DeskUuid | string | 桌台UUID | 关联桌台信息 |
| OrderType | string | 订单类型 | dine_in堂食, take_away外带, delivery外送 |
| OrderNo | string | 订单号 | 业务订单编号 |
| OrderStatus | string | 订单状态 | order待付款, paid已付款, cancelled已取消 |
| OrderTime | int64 | 下单时间 | 时间戳 |
| PaymentTime | int64 | 支付时间 | 时间戳 |
| CancelTime | int64 | 取消时间 | 时间戳 |
| CancelReason | string | 取消原因 | |
| TotalAmount | float64 | 总金额 | 订单总金额 |
| DiscountAmount | float64 | 折扣金额 | |
| ServiceCharge | float64 | 服务费 | |
| TaxAmount | float64 | 税费 | |
| PaidAmount | float64 | 已付金额 | |
| ChangeAmount | float64 | 找零金额 | |
| PaymentMethod | string | 支付方式 | cash现金, card刷卡, mobile移动支付 |
| PaymentStatus | string | 支付状态 | unpaid未支付, paid已支付, refunded已退款 |
| CashierUuid | string | 收银员UUID | 关联收银员 |
| CustomerUuid | string | 客户UUID | 关联客户（可选） |
| CustomerCount | int | 客户数量 | |
| Remark | string | 备注 | |
| Tags | string | 标签 | JSON格式存储 |
| Source | string | 订单来源 | pos收银台, h5移动端, api接口 |

## 关联关系

### 关联实体
- **CompanyUuid** → Company 实体（通过 CompanyUuid 关联）
- **OrderUuid** → SaleOrder 实体（通过 OrderUuid 关联）
- **DeskUuid** → Desk 实体（通过 DeskUuid 关联）
- **CashierUuid** → 收银员（外部关联）
- **CustomerUuid** → Customer 实体（通过 CustomerUuid 关联，可选）

### 被关联关系
- 被支付记录关联（通过订单UUID）
- 被订单项关联（通过订单UUID）

## 数据流分析

### 数据来源
- POS收银系统创建的销售账单
- H5移动端生成的订单
- API接口创建的订单

### 数据流向
1. **订单创建流程**:
   - 客户下单（堂食/外带/外送）
   - 系统生成订单号和UUID
   - 创建SaleBill记录，状态为order（待付款）
   - 推送订单信息到相关终端

2. **订单处理流程**:
   - 收银员确认订单信息
   - 计算总金额、折扣、服务费等
   - 选择支付方式进行处理
   - 更新订单状态为paid（已付款）

3. **订单取消流程**:
   - 取消订单时记录取消原因和时间
   - 更新订单状态为cancelled
   - 处理退款（如已支付）

4. **支付处理流程**:
   - 支付成功后更新支付时间和状态
   - 计算找零金额
   - 记录支付方式信息

### 业务场景
- 餐厅POS收银系统
- 多种订单类型支持（堂食、外带、外送）
- 多种支付方式支持
- 订单生命周期管理
- 客户数量统计
- 订单取消和退款处理

## 索引建议

- 主键索引: ID
- 唯一索引: Uuid
- 唯一索引: OrderNo（订单号唯一）
- 普通索引: CompanyUuid（公司查询）
- 普通索引: OrderUuid（关联订单查询）
- 普通索引: DeskUuid（桌台查询）
- 普通索引: OrderStatus（状态筛选）
- 普通索引: OrderTime（时间范围查询）
- 普通索引: PaymentStatus（支付状态查询）
- 普通索引: CashierUuid（收银员查询）

## WebSocket推送

### AfterUpdate钩子
- **订单更新推送**: 当订单状态变更时，推送到收银端
- **桌台状态推送**: 当订单支付完成时，更新桌台状态并推送

### 推送目标
- 收银端：订单状态更新、支付完成
- 厨房端：新订单通知、订单取消
- 客户端：订单状态查询结果

## 注意事项

1. **多租户支持**: 通过CompanyUuid实现数据隔离
2. **订单状态管理**: 严格的状态流转控制
3. **支付安全**: 支付信息需要加密传输
4. **软删除机制**: 使用DeleteTime实现软删除
5. **实时同步**: 通过WebSocket实现多终端实时同步
6. **金额精度**: 使用float64需要注意精度问题，建议改用decimal
7. **订单号唯一性**: 需要确保OrderNo的全局唯一性
8. **取消权限**: 只有特定角色才能取消订单
9. **数据审计**: 重要操作需要记录操作日志
10. **异常处理**: 支付失败、网络异常等情况的处理机制

## 业务规则

1. **订单类型限制**: 不同订单类型可能有不同的业务规则
2. **支付方式验证**: 根据订单类型限制可用支付方式
3. **折扣计算**: 折扣金额不能超过总金额
4. **服务费计算**: 服务费可能基于订单类型或金额计算
5. **税费计算**: 税费计算需要符合当地法规
6. **找零限制**: 找零金额不能为负数
7. **取消时间限制**: 已支付订单的取消可能需要特殊处理
8. **客户数量验证**: 客户数量不能为负数