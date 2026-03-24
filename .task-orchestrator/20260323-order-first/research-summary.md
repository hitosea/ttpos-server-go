# 调研摘要：会员端堂食"先下单后付"模式

## 1. Shop端手机点餐配置

- **API**: GET/POST `/shop/setting/store_scan_order`
- **API Handler**: `main/app/api/v1/shop/shop_setting.go` (lines 1219-1262)
- **Service**: `main/app/service/setting/setting.go` (lines 2606-2667)
- **DTO**: `main/app/dto/req/store_scan_order_setting.go`
- **存储**: `ttpos_setting` 表, key=`store_scan_order`, value=JSON
- **会员堂食开关**: `is_open_member_instant` 字段在 `ttpos_company_setting` 表

## 2. 会员端堂食下单流程

- **API**: `main/app/api/v1/member/member_order.go`
  - POST `/member/order/dine_in/create` → CreateDineInOrder (line 77-101)
  - POST `/member/order/dine_in/pay` → PayDineInOrder (line 211-233)
  - GET `/member/order/dine_in/list` → GetMemberDineInOrderList (line 648-664)
  - GET `/member/order/dine_in/detail` → GetMemberDineInOrderDetail (line 677-693)
- **Service**: `main/app/service/order_member.go` (3600+ lines)
  - CreateMemberDineInOrder (line ~2761)
  - createDineInOrder (line ~2786)
  - GetMemberDineInOrderList (line ~3128)
  - GetMemberDineInOrderDetail (line ~3369)
  - 订单状态判断逻辑 (line ~3628-3690)
- **支付完成事件**: `main/app/event/member/member_dine_in_order_pay_finish_event_handler.go`
  - createH5OrderForMemberDineIn (line 35-191) — 创建H5订单
  - markMemberDineInOrderComplete (line 335-492) — 标记完成
  - autoAcceptMemberDineInOrder (line 196-331) — 自动接单

### 当前流程
创建订单(SaleBill status=0) → 支付 → 支付回调 → 创建H5订单 → 标记完成(status=1) → 收银接单

### 订单状态常量 (main/app/constant/order.go)
- OrderSourceMemberDineIn = "member_dine_in"
- SaleBillTypeInstant = 1
- SaleBillStatusPending = 0, Complete = 1, Canceled = 2
- H5OrderStatusOrder = 1(未接单), Accepted = 2(已接单), Rejected = 3
- MemberDineInDetailStatus: unpaid, pending, preparing, completed, partial_refund, full_refund, cancelled, rejected

## 3. 收银端接单逻辑

- **API**: `main/app/api/v1/cashier/cashier_h5_order.go`
  - POST `/cashier/h5_order/accept` → AcceptH5Order (line 113-134)
  - POST `/cashier/h5_order/reject` → RejectH5Order (line 84-101)
- **Service**: `main/app/service/order_h5.go`
  - AcceptH5Order (line ~274-444) — 接单核心逻辑
- **送厨**: `main/app/service/order_action.go`
  - ActionCooking (line 140-251) — 送厨逻辑

### 接单流程
AcceptH5Order → 验证H5订单 → 标记已接单 → ActionCooking(送厨) → 发布事件

## 4. 即时点餐挂单/取单机制

- **挂单标志**: `SaleBill.hide_bill_time > 0` = 挂单, `= 0` = 取单
- **挂单API**: POST `/cashier/instant/order/hide` → HideOrder
  - Service: order_base.go line 446-497
  - 设置 hide_bill_time = time.Now().Unix()
- **取单API**: POST `/cashier/instant/order/show` → ShowOrder
  - Service: order_base.go line 500-569
  - 设置 hide_bill_time = 0, device_uuid = 当前设备
- **挂单列表API**: GET `/cashier/instant/order/list`
  - Service: order_base.go line 571-690
  - 查询条件: hide_bill_time > 0 AND status = 0
- **Model方法**:
  - `sale_bill.go:495` IsShowSaleBill() — hide_bill_time == 0
  - `sale_bill_ext_getset.go:819` SetHideSaleBill() — hide_bill_time = now
  - `sale_bill_ext_getset.go:813` SetShowSaleBill(deviceUuid) — hide_bill_time = 0

### 挂单后的即时点餐操作
取单后完全复用即时点餐流程：加购(ActionAdd)、送厨(ActionCooking)、结账(InstantOrderPaymentFinish)、反结账(ReverseSettle)

## 5. 数据回显机制

- 会员端和收银端共享同一 SaleBill/SaleOrder/SaleOrderProduct 表
- 会员端每次查询(GetMemberDineInOrderDetail)直接从数据库读取最新数据
- 收银端修改通过 CalcAndSaveSaleBill() 持久化到表中
- **无需额外同步机制**，会员端刷新即可看到最新数据

## 6. 新功能设计方案

### 新增配置字段
在 `store_scan_order_setting.go` 的 DTO 中新增:
- `IsOrderFirstPayLater int` — 0:先付后下单(默认), 1:先下单后付

### 新流程（先下单后付模式）
1. 会员端 CreateMemberDineInOrder → 创建 SaleBill(status=0, bill_type=1)
2. 直接创建 H5 Order(status=1/未接单) — 不经过支付
3. SaleBill 保持 status=0（未付款）
4. 收银端 AcceptH5Order → 标记H5已接单 → 设置 SaleBill.hide_bill_time > 0（挂单）
5. 收银端即时点餐取单 → 正常即时点餐流程（加购/送厨/结账/反结账）
6. 会员端查询详情 → 自动获取最新数据

### 会员端状态映射
- 创建后(H5 status=1): 待接单(pending)
- 接单后(H5 status=2): 备餐中(preparing)
- 收银结账后(SaleBill status=1): 已完成(completed)
