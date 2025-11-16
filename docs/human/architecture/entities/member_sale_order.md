# MemberSaleOrder 实体模型说明

## 基本信息

- **实体名称**: MemberSaleOrder
- **表名**: ttpos_member_sale_order
- **所属模块**: model
- **描述**: 会员端销售订单实体，用于管理会员外卖订单的完整生命周期

## 字段说明

| 字段名 | 类型 | 说明 | 备注 |
|--------|------|------|------|
| ID | uint | 主键ID | 继承自BaseModel |
| Uuid | uint64 | UUID | 继承自BaseModel |
| CreateTime | int64 | 创建时间 | 继承自BaseModel |
| UpdateTime | int64 | 更新时间 | 继承自BaseModel |
| DeleteTime | int64 | 删除时间 | 继承自BaseModel |
| MemberUuid | uint64 | 会员UUID | 关联会员信息 |
| SaleBillUuid | uint64 | 销售账单UUID | 关联销售账单 |
| SaleOrderUuid | uint64 | 销售订单UUID | 关联销售订单 |
| Status | uint | 订单状态 | 0-选购中 1-待付款 2-待商家接单 3-商家备餐中 4-待骑手接单 5-骑手正在赶往商家 6-骑手配送中 7-已完成 8-已取消 |
| SerialNumber | string | 订单流水号 | |
| OrderNo | string | 订单号 | |
| CancelScene | string | 取消场景 | merchant_cancel-商家取消；member_cancel-用户取消；merchant_reject-商家拒单 |
| IsAutoAccept | uint | 是否自动接单 | 0-否；1-是 |
| Remark | string | 订单备注 | |
| CancelReason | string | 取消原因 | |
| IsVerifiedPhone | uint | 是否已验证手机号 | 0-未验证 1-已验证 |
| PaymentMethodUuid | uint64 | 支付方式UUID | 订单已选择的支付方式 |
| ProductNum | float64 | 商品数量 | 订单中商品的总数量 |
| ProductAmount | float64 | 商品金额 | 折前价，已含税 |
| OriginProductAmount | float64 | 商品原价 | 折前价，已含税 |
| MemberDiscountFee | float64 | 会员折扣 | |
| Amount | float64 | 订单总金额 | 商品金额-会员折扣+配送费 |
| RefundAmount | float64 | 退款金额 | |
| IsDistanceCalculated | int | 是否已计算距离费 | -1-未计算，1-已计算 |
| DeliveryDistance | float64 | 配送距离 | 单位km |
| DeliveryFeeAmount | float64 | 配送费 | |
| DeliveryFeeDistance | float64 | 距离费送费 | |
| DeliveryFeeMinFee | float64 | 起步配送费 | |
| DeliveryFeeBaseFee | float64 | 基础服务费 | |
| DeliveryFeePerKm | float64 | 每公里配送费 | |
| RiderAcceptTimeout | int | 骑手接单超时时间 | 单位分钟 |
| RelatedOrderNo | string | 关联订单号 | skootar、grab等第三方平台上的订单号 |
| RelatedOrderType | string | 关联订单类型 | skootar、grab |
| RiderName | string | 骑手名称 | |
| RiderPhone | string | 骑手电话 | |
| Location | string | 骑手位置 | 格式:纬度,经度 |
| RiderAvatar | string | 骑手头像 | |
| RiderRating | float64 | 骑手评分 | |
| RemainingDistance | float64 | 剩余距离 | |
| MemberAddressUuid | uint64 | 会员收货地址UUID | |
| ContactLocation | string | 位置坐标 | "纬度,经度" |
| ContactAddress | string | 详细地址 | |
| ContactAddressDetail | string | 详细地址 | |
| ContactName | string | 联系人 | |
| ContactPhone | string | 联系电话 | |
| ContactPhonePrefix | string | 联系电话前缀 | |
| ContactGender | int | 联系人性别 | 0-女士 1-先生 |
| Sort | int | 排序 | 0-其他状态，1-骑手正在赶往商家，2-骑手配送中，降序排序 |
| PayTime | int64 | 支付完成时间 | 时间戳 |
| SubmitPayTime | int64 | 提交支付时间戳 | |
| AcceptTime | int64 | 商家接单时间 | 时间戳 |
| CookTime | int64 | 商家备餐完成时间 | 时间戳 |
| RiderAcceptTime | int64 | 骑手接单时间 | 时间戳 |
| RiderStartTime | int64 | 骑手开始配送时间 | 时间戳 |
| FinishTime | int64 | 骑手送达时间 | 时间戳 |
| ExpectedFinishTime | string | 预计送达时间 | |
| CancelTime | int64 | 取消时间 | 时间戳 |

## 关联关系

### 关联实体
- **MemberUuid** → Member 实体（通过 MemberUuid 关联）
- **SaleBillUuid** → SaleBill 实体（通过 SaleBillUuid 关联）
- **SaleOrderUuid** → SaleOrder 实体（通过 SaleOrderUuid 关联）
- **PaymentMethodUuid** → PaymentMethod 实体（通过 PaymentMethodUuid 关联）
- **MemberAddressUuid** → MemberAddress 实体（通过 MemberAddressUuid 关联）

## 数据流分析

### 数据来源
- 会员端APP下单
- 微信小程序下单
- 支付宝小程序下单
- 第三方平台订单同步

### 数据流向
1. **下单流程**:
   - 会员选择商品和收货地址
   - 系统计算配送费和订单金额
   - 创建MemberSaleOrder记录，状态为选购中
   - 提交支付，状态变为待付款

2. **支付流程**:
   - 会员完成支付
   - 更新支付时间和状态为待商家接单
   - 推送订单到商家端
   - 开始配送费计算

3. **商家接单流程**:
   - 商家查看订单详情
   - 确认接单或拒单
   - 接单后状态变为商家备餐中
   - 开始制作商品

4. **配送流程**:
   - 备餐完成后等待骑手接单
   - 骑手接单后开始配送
   - 配送过程中更新位置信息
   - 送达后订单完成

### 业务场景
- 会员外卖订单管理
- 配送费自动计算
- 骑手配送跟踪
- 订单取消和退款
- 第三方平台集成

## 索引建议

- 主键索引: ID
- 唯一索引: Uuid
- 唯一索引: OrderNo（订单号唯一）
- 普通索引: MemberUuid（会员查询）
- 普通索引: SaleBillUuid（销售账单查询）
- 普通索引: SaleOrderUuid（销售订单查询）
- 普通索引: Status（状态筛选）
- 普通索引: PayTime（支付时间查询）
- 普通索引: FinishTime（完成时间查询）

## WebSocket推送

### AfterUpdate钩子
- **订单状态推送**: 当订单状态变更时，推送到会员端和商家端
- **骑手位置推送**: 骑手位置更新时推送到会员端

### 推送目标
- **会员端**: 订单状态更新、骑手位置、预计送达时间
- **商家端**: 新订单通知、状态更新
- **骑手端**: 订单分配、导航信息

## 业务方法

### GetActualConsumptionAmount() float64
- **功能**: 获取订单实际消费金额（订单金额-配送费）
- **返回**: 实际消费金额
- **用途**: 会员积分计算

### GetContactPhoneMask() string
- **功能**: 获取脱敏手机号（只显示后四位）
- **返回**: 脱敏手机号
- **用途**: 隐私保护

### IsSelfCancel() bool
- **功能**: 判断是否是自主取消
- **返回**: 是否自主取消
- **用途**: 取消类型判断

### IsMerchantCancel() bool
- **功能**: 判断是否是商家取消
- **返回**: 是否商家取消
- **用途**: 取消责任判断

### IsCanRefund() bool
- **功能**: 判断是否还可退款
- **返回**: 是否可退款
- **用途**: 退款权限验证

### CalculateMemberDiscount() float64
- **功能**: 计算外卖单会员折扣
- **返回**: 会员折扣金额
- **用途**: 折扣计算

### UpdateDeliveryConfig(deliveryConfig DeliveryConfigResponse)
- **功能**: 更新配送费配置
- **参数**: 配送费配置
- **用途**: 配送费配置更新

### CalculateDeliveryFee() float64
- **功能**: 计算配送费
- **返回**: 配送费金额
- **用途**: 配送费计算

### Accept()
- **功能**: 商家接单
- **用途**: 订单接单处理

### Reject()
- **功能**: 商家拒单
- **用途**: 订单拒单处理

### CookFinish()
- **功能**: 备餐完成
- **用途**: 备餐状态更新

### RiderAccept(riderName, riderPhone, location, riderAvatar, riderRating)
- **功能**: 骑手接单
- **用途**: 骑手接单处理

### RiderDelivery(riderName, riderPhone, location)
- **功能**: 骑手配送中
- **用途**: 配送状态更新

### RiderCompleted(riderName, riderPhone, location)
- **功能**: 骑手配送完成
- **用途**: 订单完成处理

## 注意事项

1. **状态管理**: 严格的状态流转控制，确保订单状态的一致性
2. **配送费计算**: 配送费计算需要考虑距离、时间、天气等因素
3. **骑手信息**: 骑手信息需要实时更新，确保位置准确性
4. **取消规则**: 不同状态的订单有不同的取消规则和责任划分
5. **退款处理**: 退款金额需要根据订单状态和配送情况计算
6. **隐私保护**: 联系人信息需要脱敏处理
7. **第三方集成**: 需要处理第三方平台的订单同步和状态更新

## 业务规则

1. **订单状态流转**: 选购中 → 待付款 → 待商家接单 → 商家备餐中 → 待骑手接单 → 骑手配送中 → 已完成
2. **取消规则**: 会员只能在商家接单前取消，商家只能在备餐中取消
3. **配送费计算**: 配送费 = 起步配送费 + 距离费，不低于起步配送费
4. **支付超时**: 订单支付有24小时超时限制
5. **接单超时**: 骑手接单有超时限制，超时后自动重新分配
6. **退款规则**: 已配送的订单只能部分退款，未配送的订单可以全额退款
7. **会员折扣**: 会员折扣只适用于商品金额，不适用于配送费

## 扩展功能

### 智能配送调度
- 基于骑手位置和订单距离的智能调度
- 多骑手协同配送
- 配送路径优化

### 订单预测
- 基于历史数据的订单量预测
- 高峰期预警和资源调度
- 配送能力预估

### 会员营销
- 基于订单历史的个性化推荐
- 会员等级和权益管理
- 积分和奖励系统

### 数据分析
- 订单热力图分析
- 配送效率分析
- 客户行为分析

## 性能优化

### 查询优化
- 使用索引优化订单查询
- 批量加载关联数据
- 分页查询优化

### 状态同步
- 批量状态更新
- 状态变更队列
- 增量状态同步

### 地理位置处理
- 地理位置索引优化
- 距离计算缓存
- 位置更新批量处理

## 统计分析

### 订单统计
- **订单量统计**: 按时间、地区、会员统计订单量
- **订单金额统计**: 订单总金额、平均订单金额分析
- **完成率统计**: 订单完成率和取消率分析

### 配送分析
- **配送时间分析**: 平均配送时间、准时率统计
- **配送距离分析**: 平均配送距离、配送费分析
- **骑手效率分析**: 骑手接单量、配送效率统计

### 会员分析
- **会员消费分析**: 会员消费频率、消费金额分析
- **复购率分析**: 会员复购率和留存率分析
- **偏好分析**: 会员点餐偏好和时间偏好分析