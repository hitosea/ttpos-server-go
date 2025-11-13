# Recharge Order Service 会员充值订单服务说明文档

## 📋 概述

`service/recharge_order.go` 是 TTPOS 系统的会员充值订单管理服务，负责处理会员充值的完整业务流程。该服务支持多种支付方式、充值赠送、反结账、退款等复杂业务场景，并集成了ERP系统、短信通知、打印小票等功能。

**文件路径**: `/home/coder/workspaces/ttpos-server-go/main/app/service/recharge_order.go`  
**文件大小**: 2087行  
**接口定义**: `IRechargeOrderSrv`  
**实现结构**: `rechargeOrderSrv`

---

## 🏗️ 架构设计

### 接口定义 (IRechargeOrderSrv)

```go
type IRechargeOrderSrv interface {
    // 获取进行中的会员充值订单
    GetPendingRechargeOrder(companyUuid uint64) resp.RechargeOrder
    
    // 创建充值订单
    CreateRechargeOrder(ctx context.Context, rechargeReq req.RechargeReq) (resp.RechargeOrder, error)
    
    // 充值订单添加支付方式
    AddPaymentMethod(ctx context.Context, addPaymentMethod req.RechargeOrderAddPaymentMethodReq) (resp.RechargeOrder, error)
    
    // 充值订单撤销支付方式
    CancelPaymentMethod(ctx context.Context, cancelPaymentMethod req.RechargeOrderCancelPaymentMethodReq) (resp.RechargeOrder, error)
    
    // 确认充值订单
    ConfirmRechargeOrder(ctx context.Context, confirmRechargeOrderReq req.ConfirmRechargeOrder) (resp.ConfirmRechargeOrder, error)
    
    // 打印充值订单
    PrintTicket(ctx context.Context, printRechargeOrderReq req.PrintRechargeOrderReq) (*resp.PrinterData, error)
    
    // 充值订单列表
    GetRechargeOrderList(ctx context.Context, listReq req.RechargeOrderListReq) (resp.RechargeOrderList, error)
    
    // 获取充值订单详情
    GetRechargeOrderInfo(ctx context.Context, uuid uint64) (resp.RechargeOrderInfo, error)
    
    // 获取支付方式的二维码信息
    GetRechargeOrderPaymentQrcode(ctx context.Context, req req.RechargeOrderPaymentQrcodeReq) (*resp.RechargeOrderPaymentQrcodeInfoResp, error)
    
    // 取消充值订单
    CancelRechargeOrder(ctx context.Context, uuid uint64) error
    
    // 获取退款信息
    GetRechargeOrderRefundInfo(ctx context.Context, uuid uint64) (resp.RechargeOrderRefundInfo, error)
    
    // 检查反结账信息
    CheckRechargeOrderReverseSettle(ctx context.Context, uuid uint64) (resp.RechargeOrderReverseSettleInfo, error)
    
    // 充值订单反结账
    RechargeOrderReverseSettle(ctx context.Context, uuid uint64) error
    
    // 充值订单退款
    RechargeOrderRefund(ctx context.Context, refundReq req.RechargeOrderRefundReq) error
    
    // 充值订单重新退款
    RechargeOrderReReturnOrder(ctx context.Context, req req.RechargeOrderReReturnReq) error
}
```

### 依赖服务

```go
type rechargeOrderSrv struct {
    dbm              *database.DBManager  // 数据库管理器
    bus              *event.SystemEventBus // 事件总线
    cache            cache.Cache          // 缓存
    paymentMethodSrv IPaymentMethodSrv    // 支付方式服务
    settingSrv       setting.ISrv         // 设置服务
    cashBoxSrv       ICashBoxSrv          // 钱箱服务
    memberSrv        IMemberSrv           // 会员服务
    smsSrv           ISmsSrv              // 短信服务
    lock             lock.Lock            // 并发锁
}
```

### 服务初始化

```go
func NewRechargeOrderSrv(
    dbm *database.DBManager, 
    cache cache.Cache, 
    paymentMethodSrv IPaymentMethodSrv, 
    settingSrv setting.ISrv, 
    cashBoxSrv ICashBoxSrv, 
    memberSrv IMemberSrv, 
    smsSrv ISmsSrv
) IRechargeOrderSrv {
    return &rechargeOrderSrv{
        dbm:              dbm,
        bus:              event.NewSystemBus(),
        cache:            cache,
        paymentMethodSrv: paymentMethodSrv,
        settingSrv:       settingSrv,
        cashBoxSrv:       cashBoxSrv,
        memberSrv:        memberSrv,
        smsSrv:           smsSrv,
        lock:             lock.NewSystemLock(),
    }
}
```

---

## 🎯 核心概念

### 1. 充值订单状态

| 状态 | 常量 | 说明 |
|-----|------|------|
| 待支付 | `RechargeOrderStatusPending` (0) | 订单已创建，待支付 |
| 已支付 | `RechargeOrderStatusPaid` (1) | 已完成充值 |
| 已取消 | `RechargeOrderStatusCanceled` (2) | 订单已取消 |
| 已过期 | `RechargeOrderStatusExp` (3) | 订单已过期 |

### 2. 充值订单业务流程

```
创建充值订单
  ↓
添加支付方式（可多种）
  ├─ 现金
  ├─ 第三方支付（微信、支付宝等）
  └─ 其他支付方式
  ↓
确认充值（完成支付）
  ├─ 更新会员余额
  ├─ 赠送余额
  ├─ 赠送积分
  ├─ 更新钱箱余额
  ├─ 保存ERP发票
  ├─ 打印充值单
  ├─ 发送短信通知
  └─ 发布事件
  ↓
充值完成
```

### 3. 支付方式处理

#### 现金支付特殊规则
- 可以超过充值金额（用于找零）
- 找零金额 = 实付金额 - 充值金额 - 手续费
- 支持多次添加，自动合并

#### 第三方支付
- 不能超过充值金额
- 支持连连支付（微信、支付宝、PromptPay）
- 生成支付二维码
- 异步回调更新支付状态

### 4. 金额计算

```go
// 充值金额：用户要充值的金额
rechargeAmount := 1000.00

// 赠送金额：系统赠送的余额
giftAmount := 100.00

// 支付金额：实际支付的金额
paymentAmount := 1000.00

// 手续费：第三方支付手续费
commissionFee := paymentMethod.CalculatePaymentCommissionFee(paymentAmount)

// 实付金额：支付金额 + 手续费
actualAmount := paymentAmount + commissionFee

// 找零：现金支付时
chargeDue := actualAmount - rechargeAmount - commissionFee

// 应收金额：订单总金额
amount := rechargeAmount + commissionFee

// 充值后余额
balanceRecharged := member.Balance + rechargeAmount + giftAmount
```

---

## 🎯 核心功能

### 1. 获取进行中的充值订单 (GetPendingRechargeOrder)

**功能描述**: 获取当前公司进行中（待支付）的充值订单。

#### 方法签名

```go
func (s *rechargeOrderSrv) GetPendingRechargeOrder(companyUuid uint64) resp.RechargeOrder
```

#### 业务规则

- 同一时间只能有一个待支付的充值订单
- 返回订单的所有支付方式信息
- 计算找零金额

#### 代码实现

```go
func (s *rechargeOrderSrv) GetPendingRechargeOrder(companyUuid uint64) resp.RechargeOrder {
    rechargeOrderRepo := repository.NewMemberRechargeOrderRepo(s.dbm.GetDB(companyUuid))
    
    // 查询待支付订单
    order := rechargeOrderRepo.GetRechargeOrder(
        rechargeOrderRepo.WhereStatus(constant.RechargeOrderStatusPending),
        rechargeOrderRepo.WithPaymentOrders(),
        rechargeOrderRepo.WithPaymentOrderPaymentMethod(),
    )
    
    if order.Uuid == 0 {
        return resp.RechargeOrder{
            PaymentOrders: resp.PaymentInfoList{
                List: make([]resp.PaymentOrder, 0),
            },
        }
    }
    
    // 组装支付订单信息
    respPaymentOrders := make([]resp.PaymentOrder, 0, len(order.PaymentOrders))
    for _, paymentOrder := range order.PaymentOrders {
        var respPaymentOrder resp.PaymentOrder
        copier.Copy(&respPaymentOrder, paymentOrder)
        
        respPaymentOrder.PaymentMethodCode = paymentOrder.PaymentMethod.Code
        respPaymentOrder.PaymentMethodName = paymentOrder.PaymentMethod.PaymentName
        respPaymentOrder.DisabledCancel = slices.Contains([]int{
            constant.PaymentMethodCodeLianLianWechatPay,
            constant.PaymentMethodCodeLianLianAliPay,
            constant.PaymentMethodCodeLianLianQRPromptPay,
        }, paymentOrder.PaymentMethod.Code)
        
        respPaymentOrders = append(respPaymentOrders, respPaymentOrder)
    }
    
    var respRechargeOrder resp.RechargeOrder
    copier.Copy(&respRechargeOrder, order)
    
    // 计算找零
    respRechargeOrder.ChargeDue = s.getChargeDue(order.PaymentOrders)
    respRechargeOrder.PaymentOrders = resp.PaymentInfoList{List: respPaymentOrders}
    
    return respRechargeOrder
}
```

---

### 2. 创建充值订单 (CreateRechargeOrder)

**功能描述**: 创建新的充值订单或修改已有的待支付订单。

#### 方法签名

```go
func (s *rechargeOrderSrv) CreateRechargeOrder(
    ctx context.Context, 
    rechargeReq req.RechargeReq
) (resp.RechargeOrder, error)
```

#### 请求参数

```go
type RechargeReq struct {
    MemberUuid        uint64  `json:"member_uuid"`         // 会员UUID
    RechargeAmount    float64 `json:"recharge_amount"`     // 充值金额
    GiftAmount        float64 `json:"gift_amount"`         // 赠送金额
    GiftPoint         float64 `json:"gift_point"`          // 赠送积分
    RechargeOrderUuid uint64  `json:"recharge_order_uuid"` // 订单UUID（修改时传）
}
```

#### 业务逻辑

```
1. 验证会员是否存在
   ↓
2. 判断是否为修改订单
   ├─ 是：验证充值金额不能小于已支付金额
   │      更新订单信息
   │      添加操作日志
   │
   └─ 否：检查是否已有待支付订单
           创建新订单
           生成订单号
           添加操作日志
   ↓
3. 返回待支付订单
```

#### 订单号生成规则

```go
func (s *rechargeOrderSrv) generateRechargeOrderNo() string {
    // 格式：日期(8位) + 类型(1位) + 微秒(6位) + 随机(3位)
    // 示例：20240112 + 3 + 123456 + 789
    typeNum := "3"  // 充值订单类型
    now := time.Now()
    datePart := now.Format("20060102")
    microSeconds := fmt.Sprintf("%06d", now.Nanosecond()/1000)
    
    randBytes := make([]byte, 3)
    rand.Read(randBytes)
    randNum := fmt.Sprintf("%03d", int(randBytes[0])+int(randBytes[1])+int(randBytes[2]))
    
    orderNo := datePart + typeNum + microSeconds + randNum
    
    // 使用Redis确保唯一性
    key := "__CREATE_NEW_RC_ORDERNO__" + orderNo
    if _, exits := s.cache.Get(key); exits {
        return s.generateRechargeOrderNo()  // 递归重试
    }
    
    s.cache.Set(key, 1, 5*time.Second)
    return orderNo
}
```

---

### 3. 添加支付方式 (AddPaymentMethod)

**功能描述**: 为充值订单添加支付方式，支持多种支付方式组合。

#### 方法签名

```go
func (s *rechargeOrderSrv) AddPaymentMethod(
    ctx context.Context, 
    addReq req.RechargeOrderAddPaymentMethodReq
) (resp.RechargeOrder, error)
```

#### 请求参数

```go
type RechargeOrderAddPaymentMethodReq struct {
    RechargeOrderUuid uint64                `json:"recharge_order_uuid"` // 充值订单UUID
    PaymentMethodUuid uint64                `json:"payment_method_uuid"` // 支付方式UUID
    PaymentAmount     float64               `json:"payment_amount"`      // 支付金额
    PaymentOrderUuid  uint64                `json:"payment_order_uuid"`  // 支付订单UUID（在线支付）
    CompanySetting    model.CompanySetting  `json:"-"`                   // 公司设置
}
```

#### 业务规则

| 规则 | 说明 |
|-----|------|
| 不能使用余额支付 | 充值场景不支持余额支付 |
| 非现金不能超额 | 非现金支付金额不能超过充值金额 |
| 现金可以超额 | 现金支付可以超额，用于找零 |
| 连连支付唯一 | 同一支付方式只能添加一次 |
| 支付方式必须启用 | 需验证支付方式是否可用 |

#### 执行流程

```
1. 验证充值订单状态
   ↓
2. 验证支付方式
   ├─ 支付方式是否存在
   ├─ 不能使用余额支付
   └─ 支付方式是否启用
   ↓
3. 计算支付手续费
   ↓
4. 验证支付金额
   ├─ 是否已足额
   ├─ 非现金不能超额
   └─ 现金特殊处理
   ↓
5. 创建或更新支付订单
   ├─ 已存在 → 更新金额
   └─ 不存在 → 创建新支付订单
   ↓
6. 返回更新后的订单
```

#### 现金支付特殊处理

```go
if paymentMethod.Code == constant.PaymentMethodCodeCash {
    // 如果现金支付订单存在
    if paymentOrder.Uuid != 0 {
        cashPaidPaymentAmount = paymentOrder.PaymentAmount
    }
    
    // 计算剩余应收金额
    rechargeAmountLeft = utils.DecimalSub(
        order.RechargeAmount, 
        utils.DecimalSub(sumPaymentAmount, cashPaidPaymentAmount),
    )
    
    // 如果支付金额超过剩余应收，自动调整
    if rechargeAmountLeft > 0 && addReq.PaymentAmount > rechargeAmountLeft {
        addReq.PaymentAmount = rechargeAmountLeft
    }
}
```

---

### 4. 撤销支付方式 (CancelPaymentMethod)

**功能描述**: 撤销已添加的支付方式。

#### 方法签名

```go
func (s *rechargeOrderSrv) CancelPaymentMethod(
    ctx context.Context, 
    cancelReq req.RechargeOrderCancelPaymentMethodReq
) (resp.RechargeOrder, error)
```

#### 业务规则

- 连连支付不支持撤销
- 撤销后自动调整现金支付金额
- 软删除支付订单（设置delete_time）

---

### 5. 确认充值订单 (ConfirmRechargeOrder)

**功能描述**: 确认充值订单，完成支付并更新会员余额、积分等。这是充值流程的核心方法。

#### 方法签名

```go
func (s *rechargeOrderSrv) ConfirmRechargeOrder(
    ctx context.Context, 
    confirmReq req.ConfirmRechargeOrder
) (resp.ConfirmRechargeOrder, error)
```

#### 执行流程

```
1. 加锁防止并发
   ↓
2. 验证会员和订单
   ↓
3. 验证支付金额
   ├─ 非现金不能超额
   └─ 总金额必须足额
   ↓
4. 开启事务
   │
   ├─ 更新充值订单状态
   │  ├─ status = Paid
   │  ├─ payment_time = now
   │  ├─ amount = 应收金额
   │  ├─ charge_due = 找零
   │  ├─ balance = 充值前余额
   │  └─ balance_recharged = 充值后余额
   │
   ├─ 处理会员积分
   │  └─ 赠送积分（如果有）
   │
   ├─ 处理会员余额
   │  ├─ 充值余额
   │  └─ 赠送余额
   │
   ├─ 更新钱箱余额
   │  └─ 现金支付时增加钱箱余额
   │
   ├─ 添加操作日志
   │
   └─ 保存ERP发票（如果开启）
   ↓
5. 提交事务
   ↓
6. 异步处理
   ├─ 处理会员升级
   ├─ 打印充值单
   ├─ 发送短信通知
   └─ 发布事件
       ├─ 会员余额变动事件
       ├─ 会员积分变动事件
       └─ 统计事件
   ↓
7. 返回确认结果
```

#### 代码实现（核心部分）

```go
func (s *rechargeOrderSrv) ConfirmRechargeOrder(ctx context.Context, confirmReq req.ConfirmRechargeOrder) (resp.ConfirmRechargeOrder, error) {
    // 1. 加锁
    if ctx.NoLock() {
        s.lock.LockUuid(confirmReq.RechargeOrderUuid)
        defer s.lock.UnlockUuid(confirmReq.RechargeOrderUuid)
        ctx.AddLock()
    }
    
    // 2. 验证
    member := memberRepo.GetMember(memberRepo.WhereUuid(confirmReq.MemberUuid))
    if member.Uuid == 0 {
        return confirmResp, errors.New("会员不存在")
    }
    
    order := rechargeOrderRepo.GetRechargeOrder(...)
    if order.Uuid == 0 || order.Status != constant.RechargeOrderStatusPending {
        return confirmResp, errors.New("充值订单不存在")
    }
    
    // 验证金额
    sumPaymentAmount := s.sumPaymentAmount(order.PaymentOrders)
    if sumPaymentAmount < order.RechargeAmount {
        return confirmResp, errors.New("未足额支付")
    }
    
    // 3. 开启事务
    err := db.Transaction(func(tx *gorm.DB) error {
        ctx.SetDB(tx)
        
        // 更新充值订单
        err := repository.NewMemberRechargeOrderRepo(tx).Update(order.Uuid, map[string]any{
            "amount":            amount,
            "status":            constant.RechargeOrderStatusPaid,
            "payment_time":      paymentTime,
            "charge_due":        chargeDue,
            "balance":           member.GetBalanceAll(),
            "balance_recharged": utils.DecimalAdd(member.GetBalanceAll(), order.RechargeAmount, order.GiftAmount),
        })
        
        // 处理会员积分
        if order.GiftPoint > 0 {
            if err := s.memberSrv.HandleMemberPoints(ctx, MemberPointsChangeReq{
                Uuid:     order.MemberUuid,
                Points:   order.GiftPoint,
                Scene:    constant.MemberPointLogSceneRechargeGive,
                Describe: fmt.Sprintf("收银机管理员充值赠送操作 [%s]", ctx.GetStaff().RealName),
            }); err != nil {
                return errors.WithMessage(err)
            }
            memberPointsChanged = true
        }
        
        // 处理会员余额
        if order.RechargeAmount > 0 || order.GiftAmount > 0 {
            if err := s.memberSrv.HandleMemberBalance(ctx, MemberBalanceChangeReq{
                MemberUuid:  member.Uuid,
                Money:       order.RechargeAmount,
                GiftMoney:   order.GiftAmount,
                Scene:       constant.MemberBalanceLogRecharge,
                Describe:    fmt.Sprintf("收银机管理员操作 [%s]", ctx.GetStaff().RealName),
                RelatedUuid: order.Uuid,
            }); err != nil {
                return errors.WithMessage(err)
            }
        }
        
        // 更新钱箱
        for _, paymentOrder := range order.PaymentOrders {
            if paymentOrder.PaymentMethod.Code == constant.PaymentMethodCodeCash {
                if err := s.cashBoxSrv.UpdateBalance(ctx, UpdateCashBalanceParam{
                    Amount:    utils.DecimalSub(sumPaymentAmount, sumPaymentAmountExcludeCash),
                    Scene:     constant.CashBoxLogSceneRecharge,
                    OrderUuid: order.Uuid,
                }); err != nil {
                    return errors.WithMessage(err)
                }
            }
        }
        
        // 保存ERP发票
        if company.IsOpenErpPhase3() && companySetting.ErpnextSiteCode != "" {
            invoiceResp, err := s.SavePosInvoice(ctx, &order, tx)
            if err != nil {
                return errors.WithMessage(err)
            }
            order.ErpProductsInvoiceName = invoiceResp.ProductsInvoiceName
        }
        
        return nil
    })
    
    // 4. 异步处理
    if memberPointsChanged {
        utils.Go(func() {
            s.memberSrv.HandleMemberUpgrade(companyUuid, member.Uuid)
        })
    }
    
    // 打印充值单
    utils.Go(func() {
        order := rechargeOrderRepo.GetRechargeOrder(...)
        printer.NewPrinterRepo(ctx).PrintingRechargeOrder(order, 0)
    })
    
    // 发送短信
    utils.Go(func() {
        smsReq := sms.MemberRechargeRequest{
            Company:       ctx.GetCompany().Name,
            Recharge:      order.RechargeAmount,
            BonusMoney:    order.GiftAmount,
            BonusPoints:   order.GiftPoint,
            Balance:       member.GetBalanceAll(),
            PointsBalance: member.GetPoints(),
        }
        s.smsSrv.SendMemberRechargeSMS(ctx, member.Phone, &smsReq)
    })
    
    // 发布事件
    utils.Go(func() {
        s.bus.PublishChangeMemberBalanceEvent(...)
        s.bus.PublishChangeMemberPointsEvent(...)
        s.bus.PublishStatisticsMemberEvent(...)
    })
    
    return s.confirmRechargeOrderResp(companyUuid, order.Uuid), nil
}
```

---

### 6. 反结账 (RechargeOrderReverseSettle)

**功能描述**: 撤销已完成的充值订单，回退会员余额和积分。

#### 方法签名

```go
func (s *rechargeOrderSrv) RechargeOrderReverseSettle(
    ctx context.Context, 
    uuid uint64
) error
```

#### 业务规则

| 规则 | 说明 |
|-----|------|
| 同员工同班次 | 只能反结账本人本班次的订单 |
| 未退款 | 已退款的订单不能反结账 |
| 余额充足 | 会员余额、积分必须充足 |
| 无待支付订单 | 当前无其他待支付充值订单 |

#### 执行流程

```
1. 验证订单状态
   ├─ 订单已支付
   ├─ 未退款
   └─ 会员存在
   ↓
2. 检查会员余额
   ├─ 主账户余额 >= 充值金额
   ├─ 赠送账户余额 >= 赠送金额
   └─ 积分 >= 赠送积分
   ↓
3. 开启事务
   │
   ├─ 扣除会员积分
   ├─ 扣除会员余额
   ├─ 标记支付订单为退款
   ├─ 创建退款单
   ├─ 更新钱箱余额
   ├─ 取消ERP发票
   └─ 恢复订单为待支付状态
   ↓
4. 发送短信通知
   ↓
5. 发布事件
```

---

### 7. 退款 (RechargeOrderRefund)

**功能描述**: 对已完成的充值订单进行退款，支持整单退款和部分退款。

#### 方法签名

```go
func (s *rechargeOrderSrv) RechargeOrderRefund(
    ctx context.Context, 
    refundReq req.RechargeOrderRefundReq
) error
```

#### 请求参数

```go
type RechargeOrderRefundReq struct {
    Uuid        uint64  `json:"uuid"`         // 充值订单UUID
    RefundType  uint    `json:"refund_type"`  // 退款类型 1-部分退 2-整单退
    RefundMoney float64 `json:"refund_money"` // 退款金额
    BankCode    string  `json:"bank_code"`    // 银行代码（PromptPay）
    AccountNo   string  `json:"account_no"`   // 账号
    AccountName string  `json:"account_name"` // 账户名
}
```

#### 退款类型

| 类型 | 常量 | 说明 |
|-----|------|------|
| 部分退款 | `ReturnOrderRefundTypePartial` (1) | 退部分金额 |
| 整单退款 | `ReturnOrderRefundTypeTotal` (2) | 退全部金额 |

#### 退款逻辑

```
1. 验证订单和会员
   ↓
2. 计算退款金额
   ├─ 整单退：退款金额 = 订单金额 - 已退款金额
   └─ 部分退：使用指定金额
   ↓
3. 计算可退款金额
   ├─ 遍历所有支付方式
   ├─ 减去已退款金额
   └─ 得到每个支付方式的可退金额
   ↓
4. 验证会员余额
   ├─ 计算从主账户扣除的金额
   └─ 验证主账户余额充足
   ↓
5. 分配退款金额
   ├─ 按支付方式顺序退款
   ├─ 现金 → 退现金
   ├─ 第三方支付 → 原路退回
   └─ 创建退款单
   ↓
6. 开启事务
   │
   ├─ 更新充值订单退款金额
   ├─ 添加操作日志
   ├─ 创建退款单记录
   ├─ 创建退款金额记录
   ├─ 发起第三方退款
   ├─ 扣除会员余额
   ├─ 更新钱箱余额
   └─ 保存ERP退款发票
   ↓
7. 发送短信通知
   ↓
8. 发布统计事件
```

#### 退款金额分配算法

```go
// 按支付方式顺序分配退款金额
var returnOrderAmounts []model.ReturnOrderAmount
var refundCashMoney float64

for _, record := range paymentRecords {
    if record.RefundableAmount == 0 {
        continue
    }
    
    // 如果可退金额小于退款金额，全部退完
    if record.RefundableAmount < refundMoney {
        returnOrderAmounts = append(returnOrderAmounts, model.ReturnOrderAmount{
            PaymentMethodUuid: record.PaymentMethodUuid,
            Amount:            record.RefundableAmount,
            PaymentOrderUuid:  record.PaymentOrderUuid,
        })
        refundMoney = utils.DecimalSub(refundMoney, record.RefundableAmount)
        
        if record.PaymentMethodCode == constant.PaymentMethodCodeCash {
            refundCashMoney = record.RefundableAmount
        }
    } else {
        // 如果可退金额大于等于退款金额，退部分
        returnOrderAmounts = append(returnOrderAmounts, model.ReturnOrderAmount{
            PaymentMethodUuid: record.PaymentMethodUuid,
            Amount:            refundMoney,
            PaymentOrderUuid:  record.PaymentOrderUuid,
        })
        
        if record.PaymentMethodCode == constant.PaymentMethodCodeCash {
            refundCashMoney = refundMoney
        }
        break  // 退款完成
    }
}
```

---

### 8. 获取充值订单列表 (GetRechargeOrderList)

**功能描述**: 查询充值订单列表，支持多种筛选条件和分页。

#### 方法签名

```go
func (s *rechargeOrderSrv) GetRechargeOrderList(
    ctx context.Context, 
    listReq req.RechargeOrderListReq
) (resp.RechargeOrderList, error)
```

#### 请求参数

```go
type RechargeOrderListReq struct {
    PageNo            int    `json:"page_no"`             // 页码
    PageSize          int    `json:"page_size"`           // 每页大小
    OrderNo           string `json:"order_no"`            // 订单号（模糊搜索）
    Status            int    `json:"status"`              // 状态（-1=全部）
    DateType          int    `json:"date_type"`           // 日期类型（0=今天 1=昨天 2=本周 -1=全部）
    QueryStartTime    int64  `json:"query_start_time"`    // 查询开始时间
    QueryEndTime      int64  `json:"query_end_time"`      // 查询结束时间
    EnableCreateTime  bool   `json:"enable_create_time"`  // 启用创建时间筛选
    EnablePaymentTime bool   `json:"enable_payment_time"` // 启用支付时间筛选
}
```

#### 筛选条件

| 条件 | 说明 |
|-----|------|
| 订单号 | 模糊搜索 |
| 状态 | 待支付、已支付、已取消 |
| 日期类型 | 今天、昨天、本周 |
| 时间范围 | 创建时间或支付时间 |

#### 返回数据

```go
type RechargeOrderList struct {
    List []RechargeOrderItem     `json:"list"` // 订单列表
    Meta RechargeOrderListMeta   `json:"meta"` // 元数据
}

type RechargeOrderListMeta struct {
    PageResponse dto.PageResponse `json:",inline"` // 分页信息
    UnpaidNum    int64             `json:"unpaid_num"`   // 待支付数量
    CancelNum    int64             `json:"cancel_num"`   // 已取消数量
    CompleteNum  int64             `json:"complete_num"` // 已完成数量
}

type RechargeOrderItem struct {
    Uuid           uint64                      `json:"uuid"`
    OrderNo        string                      `json:"order_no"`
    Status         int                         `json:"status"`
    PaymentTime    int64                       `json:"payment_time"`
    RechargeAmount float64                     `json:"recharge_amount"`
    Amount         float64                     `json:"amount"`         // 实付金额（减去退款）
    PaymentMethods []string                    `json:"payment_methods"`
    GiftAmount     float64                     `json:"gift_amount"`
    GiftPoint      float64                     `json:"gift_point"`
    MemberUuid     uint64                      `json:"member_uuid"`
    RefundMoney    float64                     `json:"refund_money"`
    Cashier        RechargeOrderCashier        `json:"cashier"`
    Extra          RechargeOrderItemExtra      `json:"extra"`
}
```

---

### 9. 获取支付二维码 (GetRechargeOrderPaymentQrcode)

**功能描述**: 为第三方支付生成支付二维码。

#### 方法签名

```go
func (s *rechargeOrderSrv) GetRechargeOrderPaymentQrcode(
    ctx context.Context, 
    req req.RechargeOrderPaymentQrcodeReq
) (*resp.RechargeOrderPaymentQrcodeInfoResp, error)
```

#### 支持的支付方式

- 连连微信支付
- 连连支付宝支付
- 连连PromptPay支付

#### 返回数据

```go
type RechargeOrderPaymentQrcodeInfoResp struct {
    PaymentOrderUuid uint64  `json:"payment_order_uuid"` // 支付订单UUID
    QrCode           string  `json:"qr_code"`            // 二维码内容
    QrCodeExpireSec  int     `json:"qr_code_expire_sec"` // 二维码过期时间（秒）
    Status           int     `json:"status"`             // 支付状态
    PaymentAmount    float64 `json:"payment_amount"`     // 支付金额
}
```

---

## 🔄 辅助方法

### 金额计算方法

```go
// 计算支付订单初始金额（不含手续费）
func (s *rechargeOrderSrv) sumPaymentAmount(paymentOrders []model.PaymentOrder) float64

// 计算应收金额（含手续费）
func (s *rechargeOrderSrv) getRechargeOrderAmount(paymentOrders []model.PaymentOrder) float64

// 计算找零
func (s *rechargeOrderSrv) getChargeDue(paymentOrders []model.PaymentOrder) float64

// 计算实收金额
func (s *rechargeOrderSrv) getActualAmount(paymentOrders []model.PaymentOrder) float64

// 计算手续费总额
func (s *rechargeOrderSrv) getPayFee(paymentOrders []model.PaymentOrder) float64

// 计算非现金支付金额
func (s *rechargeOrderSrv) sumPaymentAmountExcludeCash(paymentOrders []model.PaymentOrder) float64
```

### ERP集成方法

```go
// 保存发票到ERP
func (s *rechargeOrderSrv) SavePosInvoice(
    ctx context.Context, 
    memberRechargeOrder *model.MemberRechargeOrder, 
    db *gorm.DB
) (*selling.SavePosInvoiceResp, error)

// 退款发票到ERP
func (s *rechargeOrderSrv) ReturnPosInvoice(
    ctx context.Context, 
    memberRechargeOrder *model.MemberRechargeOrder, 
    returnOrder *model.ReturnOrder, 
    db *gorm.DB, 
    returnType uint
) (*selling.ReturnPosInvoiceResp, error)
```

---

## 📊 数据模型

### MemberRechargeOrder 模型

```go
type MemberRechargeOrder struct {
    BaseModel                             // Uuid, CreateTime, UpdateTime, DeleteTime
    OrderNo                string         // 订单号
    DutyNo                 uint64         // 班次号
    RechargeAmount         float64        // 充值金额
    GiftAmount             float64        // 赠送金额
    GiftPoint              float64        // 赠送积分
    Amount                 float64        // 应收金额
    ChargeDue              float64        // 找零
    MemberUuid             uint64         // 会员UUID
    StaffUuid              uint64         // 员工UUID
    Status                 int            // 状态
    PaymentTime            int64          // 支付时间
    Balance                float64        // 充值前余额
    BalanceRecharged       float64        // 充值后余额
    RefundMoney            float64        // 退款金额
    RefundAmount           float64        // 已退金额（从余额扣除）
    ErpProductsInvoiceName string         // ERP发票名称
    
    // 关联
    Member                 *Member
    Staff                  *Staff
    PaymentOrders          []PaymentOrder
    ReturnOrders           []ReturnOrder
    RechargeOrderOperationLogs []MemberRechargeOrderOperationLog
}
```

### PaymentOrder 模型

```go
type PaymentOrder struct {
    BaseModel
    PaymentMethodName    string         // 支付方式名称
    PaymentMethodUuid    uint64         // 支付方式UUID
    PaymentFeePercent    float64        // 手续费百分比
    RelatedType          int            // 关联类型
    RelatedUuid          uint64         // 关联UUID
    CurrencyUnit         string         // 货币单位
    PaymentAmount        float64        // 支付金额
    PaymentCommissionFee float64        // 手续费
    Amount               float64        // 实付金额（PaymentAmount + CommissionFee）
    Status               int            // 状态
    
    // 关联
    PaymentMethod        *PaymentMethod
    MemberRechargeOrder  *MemberRechargeOrder
}
```

---

## 🎯 使用场景

### 1. 会员充值

```go
// 1. 创建充值订单
createReq := req.RechargeReq{
    MemberUuid:     12345,
    RechargeAmount: 1000.00,
    GiftAmount:     100.00,   // 赠送余额
    GiftPoint:      50.00,    // 赠送积分
}
order, _ := rechargeOrderSrv.CreateRechargeOrder(ctx, createReq)

// 2. 添加支付方式
addReq := req.RechargeOrderAddPaymentMethodReq{
    RechargeOrderUuid: order.Uuid,
    PaymentMethodUuid: 1,      // 现金支付
    PaymentAmount:     1000.00,
}
order, _ = rechargeOrderSrv.AddPaymentMethod(ctx, addReq)

// 3. 确认充值
confirmReq := req.ConfirmRechargeOrder{
    RechargeOrderUuid: order.Uuid,
    MemberUuid:        12345,
}
result, _ := rechargeOrderSrv.ConfirmRechargeOrder(ctx, confirmReq)
```

### 2. 多种支付方式组合

```go
// 1. 创建订单（充值1000元）
order, _ := rechargeOrderSrv.CreateRechargeOrder(ctx, req.RechargeReq{
    MemberUuid:     12345,
    RechargeAmount: 1000.00,
})

// 2. 添加微信支付500元
rechargeOrderSrv.AddPaymentMethod(ctx, req.RechargeOrderAddPaymentMethodReq{
    RechargeOrderUuid: order.Uuid,
    PaymentMethodUuid: 10,  // 微信
    PaymentAmount:     500.00,
})

// 3. 添加现金支付600元（找零100元）
rechargeOrderSrv.AddPaymentMethod(ctx, req.RechargeOrderAddPaymentMethodReq{
    RechargeOrderUuid: order.Uuid,
    PaymentMethodUuid: 1,   // 现金
    PaymentAmount:     600.00,
})

// 4. 确认充值
// 实际收款：500 + 600 = 1100元
// 充值金额：1000元
// 找零：100元
```

### 3. 反结账

```go
// 检查是否可以反结账
info, _ := rechargeOrderSrv.CheckRechargeOrderReverseSettle(ctx, orderUuid)
if info.Status == 0 {
    // 可以反结账
    err := rechargeOrderSrv.RechargeOrderReverseSettle(ctx, orderUuid)
} else {
    // 不可反结账
    fmt.Println(info.Message)
}
```

### 4. 退款

```go
// 整单退款
refundReq := req.RechargeOrderRefundReq{
    Uuid:       orderUuid,
    RefundType: constant.ReturnOrderRefundTypeTotal,
}
rechargeOrderSrv.RechargeOrderRefund(ctx, refundReq)

// 部分退款
refundReq := req.RechargeOrderRefundReq{
    Uuid:        orderUuid,
    RefundType:  constant.ReturnOrderRefundTypePartial,
    RefundMoney: 500.00,
}
rechargeOrderSrv.RechargeOrderRefund(ctx, refundReq)
```

---

## 🎨 API接口示例

### 1. 创建充值订单

#### 请求

```http
POST /api/v1/recharge_order/create
Authorization: Bearer {token}
Content-Type: application/json

{
  "member_uuid": 12345,
  "recharge_amount": 1000.00,
  "gift_amount": 100.00,
  "gift_point": 50.00
}
```

#### 响应

```json
{
  "code": 1,
  "message": "success",
  "data": {
    "uuid": 67890,
    "order_no": "202401123123456789",
    "recharge_amount": 1000.00,
    "gift_amount": 100.00,
    "gift_point": 50.00,
    "charge_due": 0,
    "payment_orders": {
      "list": []
    }
  }
}
```

### 2. 添加支付方式

#### 请求

```http
POST /api/v1/recharge_order/add_payment_method
Authorization: Bearer {token}
Content-Type: application/json

{
  "recharge_order_uuid": 67890,
  "payment_method_uuid": 1,
  "payment_amount": 1000.00
}
```

#### 响应

```json
{
  "code": 1,
  "message": "success",
  "data": {
    "uuid": 67890,
    "order_no": "202401123123456789",
    "recharge_amount": 1000.00,
    "charge_due": 0,
    "payment_orders": {
      "list": [
        {
          "uuid": 11111,
          "payment_method_uuid": 1,
          "payment_method_code": 2,
          "payment_method_name": "现金",
          "payment_amount": 1000.00,
          "amount": 1000.00,
          "payment_commission_fee": 0,
          "disabled_cancel": false
        }
      ]
    }
  }
}
```

### 3. 确认充值

#### 请求

```http
POST /api/v1/recharge_order/confirm
Authorization: Bearer {token}
Content-Type: application/json

{
  "recharge_order_uuid": 67890,
  "member_uuid": 12345
}
```

#### 响应

```json
{
  "code": 1,
  "message": "success",
  "data": {
    "amount": 1000.00,
    "actual_amount": 1000.00,
    "charge_due": 0,
    "payment_methods": ["现金"]
  }
}
```

---

## ⚡ 性能优化

### 1. 并发控制

```go
// ✅ 使用UUID锁防止并发
if ctx.NoLock() {
    s.lock.LockUuid(confirmReq.RechargeOrderUuid)
    defer s.lock.UnlockUuid(confirmReq.RechargeOrderUuid)
    ctx.AddLock()
}
```

### 2. 异步处理

```go
// ✅ 异步打印、发送短信、发布事件
utils.Go(func() {
    printer.NewPrinterRepo(ctx).PrintingRechargeOrder(order, 0)
})

utils.Go(func() {
    s.smsSrv.SendMemberRechargeSMS(ctx, member.Phone, &smsReq)
})

utils.Go(func() {
    s.bus.PublishChangeMemberBalanceEvent(...)
})
```

### 3. 订单号唯一性

```go
// ✅ 使用Redis缓存确保订单号唯一
key := "__CREATE_NEW_RC_ORDERNO__" + orderNo
if _, exits := s.cache.Get(key); exits {
    return s.generateRechargeOrderNo()  // 递归重试
}
s.cache.Set(key, 1, 5*time.Second)
```

---

## 🛡️ 最佳实践

### 1. 事务使用

```go
// ✅ 正确：使用事务确保数据一致性
err := db.Transaction(func(tx *gorm.DB) error {
    ctx.SetDB(tx)
    
    // 更新订单
    repository.NewMemberRechargeOrderRepo(tx).Update(...)
    
    // 更新会员余额
    s.memberSrv.HandleMemberBalance(ctx, ...)
    
    // 更新钱箱
    s.cashBoxSrv.UpdateBalance(ctx, ...)
    
    return nil
})
```

### 2. 金额验证

```go
// ✅ 正确：验证支付金额
sumPaymentAmount := s.sumPaymentAmount(order.PaymentOrders)
if sumPaymentAmount < order.RechargeAmount {
    return errors.New("未足额支付")
}

if paymentMethod.Code != constant.PaymentMethodCodeCash && sumPaymentAmountAddCash > order.RechargeAmount {
    return errors.New("非现金支付不能大于应收")
}
```

### 3. 异步处理

```go
// ✅ 正确：非核心业务异步处理
utils.Go(func() {
    // 打印、短信、事件等
})

// ❌ 错误：核心业务同步处理
printer.PrintingRechargeOrder(order, 0)  // 会阻塞主流程
```

### 4. 错误处理

```go
// ✅ 正确：详细的错误提示
if order.Status != constant.RechargeOrderStatusPending {
    return errors.New("充值订单不存在")
}

if sumPaymentAmount < order.RechargeAmount {
    return errors.New("未足额支付")
}
```

---

## ⚠️ 注意事项

### 1. 并发安全

- 确认充值时必须加锁
- 同一订单不能并发确认
- 使用UUID锁而非全局锁

### 2. 金额计算

- 注意浮点数精度问题
- 使用decimal包处理金额
- 手续费计算要准确

### 3. 支付方式限制

- 余额支付不能用于充值
- 非现金支付不能超额
- 连连支付不支持撤销

### 4. 反结账限制

- 同员工同班次
- 会员余额必须充足
- 已退款不能反结账

### 5. ERP集成

- 需要验证班次状态
- 交班后不能操作
- 失败时事务回滚

### 6. 异步操作

- 打印失败不影响充值
- 短信发送失败不影响充值
- 事件发布异步处理

---

## 📚 相关文档

- [Member Service](member.md) - 会员服务
- [支付方式服务](payment_method.md) - 支付方式服务
- [Cash Box Service](./cash_box_service.md) - 钱箱服务
- [短信服务](sms.md) - 短信服务

---

## 📊 服务特点总结

| 特点 | 说明 |
|-----|------|
| 复杂 | 2087行代码，业务逻辑非常复杂 |
| 完整 | 涵盖充值、退款、反结账全流程 |
| 灵活 | 支持多种支付方式组合 |
| 安全 | 并发控制、事务保证 |
| 集成 | ERP、短信、打印、事件 |
| 可靠 | 详细的验证和错误处理 |

---

## 📄 更新日志

| 日期 | 版本 | 说明 |
|-----|------|-----|
| 2025-11-12 | 1.0 | 初始文档创建 |

---

## 👥 维护者

- 开发团队：Backend Team
- 文档维护：AI Assistant

---

**注意**: 本文档基于代码自动生成，如有代码变更，请及时更新文档。充值订单服务是会员系统的核心，涉及金额处理，修改时需格外谨慎并充分测试。

