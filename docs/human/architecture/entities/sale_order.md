# SaleOrder 实体模型说明

## 基本信息

- **实体名称**: SaleOrder
- **表名**: sale_order
- **所属模块**: model
- **描述**: 销售订单实体，用于记录详细的销售订单信息，包含订单项和支付信息

## 字段说明

| 字段名 | 类型 | 说明 | 备注 |
|--------|------|------|------|
| ID | uint | 主键ID | 继承自BaseModel |
| Uuid | uint64 | UUID | 继承自BaseModel |
| CreateTime | int64 | 创建时间 | 继承自BaseModel |
| UpdateTime | int64 | 更新时间 | 继承自BaseModel |
| DeleteTime | int64 | 删除时间 | 继承自BaseModel |
| CompanyUuid | uint64 | 公司UUID | 多租户标识 |
| OrderNo | string | 订单号 | 业务订单编号 |
| OrderType | string | 订单类型 | dine_in堂食, take_away外带, delivery外送 |
| OrderStatus | string | 订单状态 | pending待处理, processing处理中, completed已完成, cancelled已取消 |
| DeskUuid | string | 桌台UUID | 关联桌台信息 |
| CustomerUuid | string | 客户UUID | 关联客户（可选） |
| CustomerCount | int | 客户数量 | |
| OrderTime | int64 | 下单时间 | 时间戳 |
| TotalAmount | float64 | 总金额 | 订单总金额 |
| DiscountAmount | float64 | 折扣金额 | |
| ServiceCharge | float64 | 服务费 | |
| TaxAmount | float64 | 税费 | |
| Remark | string | 备注 | |
| Tags | string | 标签 | JSON格式存储 |
| Source | string | 订单来源 | pos收银台, h5移动端, api接口 |

## 关联关系

### 关联实体
- **CompanyUuid** → Company 实体（通过 CompanyUuid 关联）
- **DeskUuid** → Desk 实体（通过 DeskUuid 关联）
- **CustomerUuid** → Customer 实体（通过 CustomerUuid 关联，可选）

### 关联的子实体
- **SaleOrderItem** → 订单项（一对多关系）
- **SaleOrderPayment** → 支付记录（一对多关系）

### 被关联关系
- **SaleBill** → 通过OrderUuid关联

## 数据流分析

### 数据来源
- POS收银系统创建的详细订单
- H5移动端提交的订单
- API接口创建的订单

### 数据流向
1. **订单创建流程**:
   - 客户选择商品并下单
   - 系统生成订单号和UUID
   - 创建SaleOrder主记录
   - 创建订单项记录（SaleOrderItem）
   - 订单状态为pending（待处理）

2. **订单处理流程**:
   - 厨房接收订单并开始制作
   - 更新订单状态为processing（处理中）
   - 制作完成后更新为completed（已完成）

3. **订单支付流程**:
   - 创建支付记录（SaleOrderPayment）
   - 处理各种支付方式
   - 支付完成后创建SaleBill记录

4. **订单取消流程**:
   - 取消订单时更新状态
   - 处理退款（如已支付）
   - 记录取消原因

### 业务场景
- 餐厅订单管理系统
- 订单项详细记录
- 多种支付方式处理
- 订单状态跟踪
- 客户订单历史
- 订单统计分析

## 索引建议

- 主键索引: ID
- 唯一索引: Uuid
- 唯一索引: OrderNo（订单号唯一）
- 普通索引: CompanyUuid（公司查询）
- 普通索引: DeskUuid（桌台查询）
- 普通索引: CustomerUuid（客户查询）
- 普通索引: OrderStatus（状态筛选）
- 普通索引: OrderTime（时间范围查询）
- 普通索引: OrderType（订单类型查询）

## 业务方法

### GetPayments() []SaleOrderPayment
- **功能**: 获取订单的所有支付记录
- **返回**: 支付记录列表
- **用途**: 订单支付状态查询和统计

### GetItems() []SaleOrderItem
- **功能**: 获取订单的所有订单项
- **返回**: 订单项列表
- **用途**: 订单详情展示和统计分析

### GetTotalPaid() float64
- **功能**: 计算已支付总金额
- **返回**: 已支付金额
- **用途**: 支付状态判断和找零计算

### GetRemainingAmount() float64
- **功能**: 计算剩余未支付金额
- **返回**: 未支付金额
- **用途**: 支付提醒和部分支付处理

### IsFullyPaid() bool
- **功能**: 判断订单是否已完全支付
- **返回**: 是否完全支付
- **用途**: 订单状态验证

### CanCancel() bool
- **功能**: 判断订单是否可以取消
- **返回**: 是否可取消
- **用途**: 取消权限验证

## 注意事项

1. **多租户支持**: 通过CompanyUuid实现数据隔离
2. **订单状态管理**: 严格的状态流转控制
3. **金额精度**: 使用float64需要注意精度问题，建议改用decimal
4. **软删除机制**: 使用DeleteTime实现软删除
5. **订单项一致性**: 订单项总金额应与订单总金额一致
6. **支付一致性**: 支付记录总金额应与订单总金额一致
7. **库存管理**: 订单创建时需要检查商品库存
8. **取消规则**: 不同状态的订单可能有不同的取消规则

## 业务规则

1. **订单号唯一性**: 订单号必须在系统内唯一
2. **金额验证**: 订单总金额 = 订单项金额总和 - 折扣 + 服务费 + 税费
3. **状态流转**: pending → processing → completed/cancelled
4. **支付限制**: 未完成的订单不能创建支付记录
5. **取消限制**: 已支付订单的取消需要特殊处理
6. **库存扣减**: 订单确认时需要扣减相应商品库存
7. **时间限制**: 订单创建时间和支付时间应在合理范围内
8. **客户数量**: 客户数量不能为负数

## 扩展功能

### 订单拆分
- 支持大额订单拆分为多个子订单
- 拆分后需要重新分配订单号和UUID

### 订单合并
- 支持多个小订单合并为一个订单
- 合并时需要处理订单项和支付记录

### 订单修改
- 支持在特定状态下修改订单项
- 修改时需要重新计算总金额

### 订单复制
- 支持基于历史订单创建新订单
- 复制时需要重置订单状态和支付信息