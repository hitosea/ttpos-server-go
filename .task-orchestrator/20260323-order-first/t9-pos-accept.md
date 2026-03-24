# t9: 收银端接单后将"先下单后付"订单转为即时点餐挂单

## 修改文件

- `main/app/service/order_h5.go` — AcceptH5Order 方法

## 变更说明

在 `AcceptH5Order` 方法中，获取 saleBill 之后、送厨逻辑之前，新增"先下单后付"模式的判断分支：

### 判断条件
1. `h5Order.OrderType == constant.H5OrderTypeMemberDineIn` — 会员端堂食订单
2. `storeScanOrderSetting.IsOrderFirstPayLater == constant.OrderFirstPayLaterYes` — 配置开启先下单后付
3. `saleBill.Status == constant.SaleBillStatusPending` — 订单尚未付款（status=0）

### 先下单后付分支处理
当三个条件同时满足时：
1. 将 H5 订单商品插入到销售订单中（`InsertSaleOrderProduct`）
2. 重新计算账单金额（`CalcAll`）
3. 设置挂单状态（`SetHideSaleBill`，设置 `hide_bill_time = time.Now().Unix()`）
4. 在事务中保存：H5订单、H5订单商品、SaleBill、SaleOrder
5. 直接返回，**不执行送厨（ActionCooking）**，**不标记账单完成（SetFinishSaleBill）**

### 非先下单后付分支
保持原有逻辑不变：送厨 → 标记完成 → 保存 → 发布事件 → 同步ERP

### 关键设计决策
- 通过 `s.settingSrv.GetStoreScanOrderSetting(ctx)` 获取配置，复用已有的设置服务
- 获取配置失败时仅记录 Warn 日志，按默认流程（先付后下单）处理，不阻断接单
- H5 订单仍标记为已接单（status=2），会员端显示"备餐中"
- SaleBill 保持 status=0（未付款），挂单后出现在即时点餐挂单列表（查询条件：hide_bill_time > 0 AND status = 0）
- 收银员通过取单功能取回后，完全复用即时点餐现有流程（加购/送厨/结账）
