# TTPOS系统数据流分析总结

## 系统架构概述

TTPOS是一个完整的餐饮POS系统，采用多租户架构，支持总部-分店管理模式。系统包含销售、会员、支付、仓储、员工、设备等多个核心模块。

## 核心数据流分析

### 1. 销售业务数据流

#### 1.1 开台流程
```
顾客到店 → 收银员选择桌面 → 创建SaleBill → 分配Desk → 设置Staff
```

**数据流路径:**
- 收银员通过Device登录系统
- 选择或创建Desk（桌面）
- 创建SaleBill记录，关联Company、Staff、Desk
- 通过WebSocket推送开台状态

#### 1.2 点餐流程
```
顾客点餐 → 创建SaleOrder → 添加SaleOrderProduct → 计算价格 → 实时推送
```

**数据流路径:**
- 创建SaleOrder，关联SaleBill、Member、Desk
- 添加商品到SaleOrderProduct
- 处理商品BOM、属性、套餐组合
- 实时计算订单金额，支持会员折扣
- 通过WebSocket推送订单状态到厨房显示

#### 1.3 结账支付流程
```
请求结账 → 计算费用 → 选择支付方式 → 创建PaymentOrder → 完成支付
```

**数据流路径:**
- 计算SaleBill总费用（商品金额+服务费+税费-折扣）
- 创建PaymentOrder，关联PaymentMethod
- 处理多种支付方式（现金、余额、第三方支付）
- 更新SaleOrder和SaleBill状态
- 创建PrinterLog记录打印操作

#### 1.4 H5扫码点餐流程
```
顾客扫码 → 创建H5Order → 选择商品 → 在线支付 → 同步到POS
```

**数据流路径:**
- 顾客扫码进入H5页面
- 创建H5Order，关联Member、Desk、Company
- 添加H5OrderProduct
- 在线支付创建PaymentOrder
- 支付成功后同步数据到主POS系统

### 2. 会员管理数据流

#### 2.1 会员注册流程
```
顾客注册 → 创建Member → 分配会员等级 → 发放会员卡 → 初始化积分余额
```

**数据流路径:**
- 创建Member记录，设置基础信息
- 关联MemberLevel（默认等级）
- 可选创建MemberCard和MemberCardType
- 初始化积分和余额为0

#### 2.2 会员充值流程
```
会员充值 → 创建MemberRechargeOrder → 支付处理 → 更新余额 → 记录日志
```

**数据流路径:**
- 创建MemberRechargeOrder，关联Member、Staff
- 创建PaymentOrder处理支付
- 支付成功后更新Member余额和赠送余额
- 创建MemberBalanceLog记录充值明细
- 可能赠送积分，创建MemberPointLog

#### 2.3 会员消费流程
```
会员消费 → 冻结余额积分 → 完成订单 → 解冻更新 → 记录日志
```

**数据流路径:**
- 订单支付时冻结Member余额和积分
- 订单完成后更新Member累计消费和积分
- 创建MemberBalanceLog和MemberPointLog
- 检查会员等级升级条件

#### 2.4 会员外卖流程
```
会员下单 → 创建MemberSaleOrder → 分配配送员 → 配送完成 → 结算
```

**数据流路径:**
- 创建MemberSaleOrder，关联Member、Staff（配送员）
- 添加MemberSaleOrderProduct
- 处理支付和配送费用
- 更新配送状态和完成时间

### 3. 支付系统数据流

#### 3.1 支付方式管理
```
配置支付方式 → 设置手续费 → 关联图标 → 启用/禁用
```

**数据流路径:**
- 创建PaymentMethod，设置Code、Name、FeePercent
- 上传Logo和二维码图片到File表
- 配置显示范围（收银机、点餐助手、会员充值）
- 设置ERPNext支付方式映射

#### 3.2 支付处理流程
```
选择支付方式 → 计算手续费 → 调用支付接口 → 更新状态 → 记录交易号
```

**数据流路径:**
- 根据PaymentMethod计算手续费
- 创建PaymentOrder，设置初始状态为未支付
- 调用第三方支付接口（微信、支付宝、连连支付等）
- 更新PaymentOrder状态和交易号
- 处理支付结果和异常情况

#### 3.3 退款流程
```
申请退款 → 创建RefundOrder → 处理退款 → 更新状态 → 记录日志
```

**数据流路径:**
- 创建RefundOrder，关联PaymentOrder和SaleOrder
- 处理退款金额分配（主账户/赠送账户）
- 更新PaymentOrder状态为已退款
- 恢复Member余额或积分
- 记录退款原因和操作日志

### 4. 仓储管理数据流

#### 4.1 仓库管理流程
```
创建仓库 → 设置仓库类型 → 关联ERPNext → 配置商品库存
```

**数据流路径:**
- 创建Warehouse，设置Type、Code、Status
- 关联Company和MultiLanguageName
- 设置ErpCode与ERPNext系统集成
- 创建WarehouseItem记录商品库存

#### 4.2 库存变动流程
```
库存操作 → 更新WarehouseItem → 记录变动日志 → 同步ERPNext
```

**数据流路径:**
- 入库、出库、调拨、盘点等操作
- 更新WarehouseItem的数量和状态
- 创建WarehouseInOutLog记录变动详情
- 同步数据到ERPNext系统

### 5. 员工权限数据流

#### 5.1 员工管理流程
```
创建员工 → 分配角色 → 绑定设备 → 设置权限
```

**数据流路径:**
- 创建Staff，关联Company
- 分配Role，创建StaffRole关联
- 绑定Device，设置BindKey
- 配置用户类型和权限范围

#### 5.2 登录认证流程
```
员工登录 → 验证凭据 → 记录日志 → 更新在线状态
```

**数据流路径:**
- 验证Staff的Username和Password
- 创建StaffLoginLog记录登录信息
- 更新Staff的CashierOnline状态
- 记录登录时间和设备信息

#### 5.3 交班流程
```
开始交班 → 记录现金 → 统计业务 → 生成快照 → 完成交班
```

**数据流路径:**
- 创建StaffShiftLog，记录交班信息
- 统计当班收入、支出、业务数据
- 生成StaffShiftSnapshot保存快照
- 更新Staff的CashierOnline状态

### 6. 设备管理数据流

#### 6.1 设备注册流程
```
设备注册 → 绑定公司 → 关联打印机 → 设置类型
```

**数据流路径:**
- 创建Device，设置DeviceId和Source
- 关联Company和Printer
- 设置设备类型（收银机、平板、厨显等）
- 记录设备IP、版本等信息

#### 6.2 设备状态同步
```
设备连接 → 更新状态 → 实时推送 → 心跳检测
```

**数据流路径:**
- 设备连接时更新FinallyLoginUuid和时间
- 通过WebSocket推送设备状态
- 定期心跳检测设备在线状态
- 处理设备断连和重连

### 7. 打印系统数据流

#### 7.1 打印任务流程
```
触发打印 → 创建PrinterLog → 选择打印机 → 发送任务 → 记录结果
```

**数据流路径:**
- 创建PrinterLog，关联Printer、Company、Staff
- 根据打印类型选择对应打印机
- 发送打印数据到Device或网络打印机
- 更新打印状态和结果

#### 7.2 打印模板管理
```
创建模板 → 设计布局 → 关联打印机 → 测试打印
```

**数据流路径:**
- 创建PrinterTemplate，设计打印布局
- 关联Printer和PrinterCustomize
- 支持多种打印类型（小票、标签、厨房单）
- 测试和优化打印效果

## 系统集成数据流

### 1. ERPNext集成
- **Site配置**: 管理ERPNext多租户连接
- **发票同步**: SaleOrder、MemberRechargeOrder同步到ERPNext
- **仓库同步**: Warehouse与ERPNext仓库映射
- **支付映射**: PaymentMethod与ERPNext支付方式对应

### 2. 第三方支付集成
- **连连支付**: 支持微信、支付宝、二维码支付
- **手续费计算**: 自动计算和收取手续费
- **交易同步**: 实时同步支付状态和交易号
- **退款处理**: 支持原路退款和手动退款

### 3. 消息推送集成
- **WebSocket**: 实时推送订单状态、设备状态
- **厨房显示**: 订单实时推送到KDS设备
- **会员通知**: 充值、消费、积分变动通知
- **管理告警**: 异常情况实时通知管理员

## 数据一致性保证

### 1. 事务处理
- 销售订单创建使用数据库事务保证一致性
- 支付和库存操作采用原子性操作
- 会员余额变动使用冻结机制防止并发问题

### 2. 数据同步
- 关键业务数据通过WebSocket实时推送
- 定时任务同步数据到ERPNext系统
- 设备状态定期心跳同步

### 3. 审计日志
- 完整的操作日志记录（StaffOperationLog）
- 支付和退款详细记录（PaymentOrder、RefundOrder）
- 会员变动完整追踪（MemberBalanceLog、MemberPointLog）

## 性能优化策略

### 1. 数据库优化
- 合理的索引设计提升查询性能
- 分库分表支持大量数据存储
- 读写分离提升并发性能

### 2. 缓存策略
- 商品信息缓存减少数据库查询
- 会员信息缓存提升响应速度
- 配置信息缓存减少系统加载时间

### 3. 实时通信优化
- WebSocket连接池管理
- 消息队列处理高并发推送
- 设备分组推送减少无效通信

这个数据流分析展现了TTPOS系统的完整业务流程，涵盖了从前端到后端、从操作到存储的完整数据链路，为系统设计和优化提供了重要参考。