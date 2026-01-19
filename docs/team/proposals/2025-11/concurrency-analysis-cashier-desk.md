# 收银端桌台操作并发问题分析

> 本文档分析 `cashier_desk.go` 中所有操作的并发安全性，识别潜在的并发问题。

---

## 📋 分析信息

| 项目       | 内容     |
| ---------- | -------- |
| **分析人** | xiezhihuan   |
| **日期**   | 2025-11-26   |
| **文件**   | `main/app/api/v1/cashier/cashier_desk.go` |
| **状态**   | 待评审   |
| **关联提案** | `table-operations-multi-order-lock.md` |

---

## 🎯 分析目标

识别 `cashier_desk.go` 中所有操作可能存在的并发问题，特别是：
1. 涉及多个订单的操作
2. 涉及桌台状态变更的操作
3. 与转台、并台、转菜操作可能冲突的操作

---

## 📊 操作列表与并发分析

### ✅ 单订单操作（锁机制正确）

以下操作只涉及单个订单，使用 `SaleBillUuid` 锁，锁机制正确：

| 操作 | 方法 | 锁机制 | 状态 |
|------|------|--------|------|
| **加菜** | `OrderCartProductAdd` | `LockUuid(SaleBillUuid)` | ✅ 正确 |
| **修改数量** | `OrderCartProductNum` | `LockUuid(SaleBillUuid)` | ✅ 正确 |
| **送厨** | `OrderCartProductCooking` | `LockUuid(SaleBillUuid)` | ✅ 正确 |
| **退菜** | `OrderCartProductReturning` | `LockUuid(SaleBillUuid)` | ✅ 正确 |
| **取消退菜** | `OrderCartProductCancelReturning` | `LockUuid(SaleBillUuid)` | ✅ 正确 |
| **赠菜** | `OrderCartProductGiving` | `LockUuid(SaleBillUuid)` | ✅ 正确 |
| **取消赠菜** | `OrderCartProductCancelGiving` | `LockUuid(SaleBillUuid)` | ✅ 正确 |
| **删菜** | `OrderProductDelete` | `LockUuid(SaleBillUuid)` | ✅ 正确 |
| **改价** | `OrderProductChangePrice` | `LockUuid(SaleBillUuid)` | ✅ 正确 |
| **打折** | `OrderDiscount` | `LockUuid(SaleBillUuid)` | ✅ 正确 |
| **整单改价** | `OrderAmountChange` | `LockUuid(SaleBillUuid)` | ✅ 正确 |
| **修改人数** | `OrderChangePopulation` | `LockUuid(SaleBillUuid)` | ✅ 正确 |
| **调整自助餐** | `OrderChangeBuffet` | `LockUuid(SaleBillUuid)` | ✅ 正确 |
| **自助餐加钟** | `OrderChangeBuffetClock` | `LockUuid(SaleBillUuid)` | ✅ 正确 |
| **商品备注** | `OrderProductRemark` | `LockUuid(SaleBillUuid)` | ✅ 正确 |
| **整单备注** | `OrderRemark` | `LockUuid(SaleBillUuid)` | ✅ 正确 |
| **取消折扣** | `OrderDiscountCancel` | `LockUuid(SaleBillUuid)` | ✅ 正确 |

---

### ⚠️ 已识别的并发问题

#### 1. **转台操作（ChangeDesk）**

**问题**：
- 当前只锁定源订单 `SaleBillUuid`
- 但会修改目标桌台状态（`desk.SetOpenDesk`）
- 开台操作锁定目标桌台 `DeskUuid`
- **并发场景**：转台操作检查新桌台是否空闲时，开台操作也在检查同一桌台，两者都认为桌台空闲，导致数据不一致

**解决方案**：
- 锁定源订单 + 目标桌台（按 UUID 排序）
- 详见：`table-operations-multi-order-lock.md`

**代码位置**：`main/app/service/desk.go:694-794`

---

#### 2. **并台操作（MergeDesk）**

**问题**：
- 当前同时锁定 `SaleBillUuid` 和 `companyUuid`
- 涉及多个订单（主订单 + 所有被合并的订单）
- 需要锁定所有涉及的订单

**解决方案**：
- 锁定主订单 + 所有被合并的订单（按 UUID 排序）
- 详见：`table-operations-multi-order-lock.md`

**代码位置**：`main/app/service/desk.go:799-1010`

---

#### 3. **转菜操作（OrderCartProductChangeDesk）**

**问题**：
- 当前只锁定源订单 `SaleBillUuid`
- 涉及源订单和目标订单两个订单
- **并发场景**：转菜操作修改源订单和目标订单时，如果只锁定源订单，目标订单可能被其他操作（如加菜）并发修改

**解决方案**：
- 锁定源订单和目标订单（按 UUID 排序）
- 详见：`table-operations-multi-order-lock.md`

**代码位置**：`main/app/service/order_product.go:1060-1220`

---

#### 4. **开台操作（CreateDeskOrder）**

**问题**：
- 当前锁定目标桌台 `DeskUuid`
- 与转台操作存在并发冲突（见转台操作分析）

**解决方案**：
- 转台操作需要锁定目标桌台，确保与开台操作串行执行
- 详见：`table-operations-multi-order-lock.md`

**代码位置**：`main/app/service/order_base.go:95-215`

---

### 🔍 需要进一步分析的操作

### ✅ 子订单操作（锁机制正确）

以下操作涉及子订单，但所有子订单都属于同一个 `SaleBill`，使用 `SaleBillUuid` 锁是正确的：

| 操作 | 方法 | 锁机制 | 状态 |
|------|------|--------|------|
| **创建子订单** | `InstantOrderSaleOrderCreate` | `LockUuid(SaleBillUuid)` | ✅ 正确 |
| **删除子订单** | `InstantOrderSaleOrderDelete` | `LockUuid(SaleBillUuid)` | ✅ 正确 |
| **撤销拆单** | `InstantOrderSaleOrderDeleteAll` | `LockUuid(SaleBillUuid)` | ✅ 正确 |
| **移动商品到子订单** | `SaleOrderMoveProduct` | `LockUuid(SaleBillUuid)` | ✅ 正确 |

**分析**：
- 所有子订单都属于同一个 `SaleBill`
- 使用 `SaleBillUuid` 锁可以保证同一 `SaleBill` 下的所有操作串行执行
- 不存在跨 `SaleBill` 的并发问题

**代码位置**：
- `main/app/service/order_base.go:671-922`（创建子订单）
- `main/app/service/order_base.go:924-1086`（删除子订单）
- `main/app/service/order_base.go:1087-1125`（撤销拆单）
- `main/app/service/order_base.go:782-831`（移动商品到子订单）

---

## 📝 总结

### 已确认的并发问题

1. ✅ **转台操作**：需要锁定源订单 + 目标桌台
2. ✅ **并台操作**：需要锁定主订单 + 所有被合并的订单
3. ✅ **转菜操作**：需要锁定源订单 + 目标订单
4. ✅ **开台操作**：与转台操作存在并发冲突（已通过转台操作锁定目标桌台解决）

### 单订单操作（锁机制正确）

所有只涉及单个订单的操作，使用 `SaleBillUuid` 锁，锁机制正确，不存在并发问题。

### 子订单操作（锁机制正确）

所有涉及子订单的操作（创建、删除、移动商品），虽然涉及多个子订单，但所有子订单都属于同一个 `SaleBill`，使用 `SaleBillUuid` 锁可以保证同一 `SaleBill` 下的所有操作串行执行，锁机制正确。

### 结账相关操作（锁机制正确）

以下操作涉及订单支付，使用 `SaleBillUuid` 锁，锁机制正确：

| 操作 | 方法 | 锁机制 | 状态 |
|------|------|--------|------|
| **获取结账信息** | `OrderPaymentInfo` | 无锁（只读） | ✅ 正确 |
| **选择优惠券** | `OrderPaymentCoupon` | `LockUuid(SaleBillUuid)` | ✅ 正确 |
| **选择满减活动** | `OrderPaymentActivity` | `LockUuid(SaleBillUuid)` | ✅ 正确 |
| **设置积分抵扣** | `OrderPaymentPoints` | `LockUuid(SaleBillUuid)` | ✅ 正确 |
| **创建支付单** | `OrderPaymentCreate` | `LockUuid(SaleBillUuid)` | ✅ 正确 |
| **撤销支付单** | `OrderPaymentCancel` | `LockUuid(SaleBillUuid)` | ✅ 正确 |
| **完成结账** | `OrderPaymentFinish` | `LockUuid(SaleBillUuid)` | ✅ 正确 |
| **免单** | `OrderFree` | `LockUuid(SaleBillUuid)` | ✅ 正确 |

**分析**：
- 所有结账相关操作都只涉及单个订单
- 使用 `SaleBillUuid` 锁可以保证同一订单的结账操作串行执行
- 不存在并发问题

---

## 🔗 相关文档

- 提案：`table-operations-multi-order-lock.md`
- 代码位置：
  - 收银端 API：`main/app/api/v1/cashier/cashier_desk.go`
  - 桌台服务：`main/app/service/desk.go`
  - 订单服务：`main/app/service/order_product.go`
  - 订单基础服务：`main/app/service/order_base.go`

---

**版本**: v1.0.0  
**创建日期**: 2025-11-26  
**维护者**: 技术团队

