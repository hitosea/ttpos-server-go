# Stock Entry 库存扣减流程

## 概览

Stock Entry 是 TTPOS 与 ERPNext 之间的库存扣减机制。每日定时任务将所有待扣减订单（堂食 + 外卖）的原材料按 `item_code` 合并，提交一次 Stock Entry（Material Issue）到 ERPNext 扣减库存。

---

## 核心数据结构

| 结构 | 说明 |
|------|------|
| `mergeMap` | `map[item_code] → {ItemCode, Qty}` — 按 item_code 合并后的待扣减数量 |
| `deductedSet` | `map["orderUuid:erpCode"] → bool` — 标记某订单的某 item 是否"已处理" |
| `orderRequiredItems` | `map[orderUuid] → map[erpCode]bool` — 每个订单需要扣减哪些 item |
| `excludedSet` | `map[item_code] → bool` — 本次提交中被排除的问题 item |
| `stock_deduction_log` | 持久化表 — 记录实际提交到 ERPNext 的 item（反结账依赖此表回滚） |

---

## 1. TriggerStockEntryDeduction 主流程

```mermaid
flowchart TD
    START([定时任务触发]) --> LOCK[获取分布式锁<br/>stock_entry_deduction_companyUuid]
    LOCK --> CHECK_ERP{ErpnextSiteCode<br/>是否配置?}
    CHECK_ERP -- 未配置 --> END_NIL([return nil])
    CHECK_ERP -- 已配置 --> PHASE1

    subgraph PHASE1 ["阶段一：数据准备"]
        direction TB
        Q1[查询堂食订单<br/>erp_stock_deducted=0<br/>status=已结账<br/>erp_sales_invoice_name 非空] --> Q2[查询外卖订单<br/>erp_stock_deducted=0<br/>order_state=已完成<br/>erp_pos_invoice_resp 非空]
        Q2 --> Q3[查询已有 stock_deduction_log<br/>构建 deductedSet<br/>避免重复扣减]
        Q3 --> Q4[从 sale_order_material 表<br/>读取堂食原材料]
        Q4 --> Q5[从 takeout_order_material 表<br/>读取外卖原材料]
        Q5 --> Q6[按 item_code 合并到 mergeMap<br/>跳过 deductedSet 中已扣减的]
        Q6 --> Q7[构建 orderRequiredItems<br/>无原材料的订单标记 __none__]
    end

    PHASE1 --> CHECK_EMPTY1{订单列表为空?}
    CHECK_EMPTY1 -- 是 --> END_NIL
    CHECK_EMPTY1 -- 否 --> CHECK_EMPTY2{mergeMap 为空?<br/>全部已扣减}
    CHECK_EMPTY2 -- 是 --> MARK0[markFullyDeductedOrders]
    MARK0 --> END_NIL
    CHECK_EMPTY2 -- 否 --> PHASE2

    subgraph PHASE2 ["阶段二：整体提交"]
        direction TB
        BUILD[构建 Stock Entry 请求<br/>合并所有 mergeMap items] --> SUBMIT[提交 ERPNext<br/>SubmitStockEntry<br/>Material Issue]
    end

    PHASE2 --> SUCCESS1{提交成功?}
    SUCCESS1 -- 成功 --> PHASE2_OK

    subgraph PHASE2_OK ["整体提交成功处理"]
        direction TB
        LOG1[writeDeductionLogs<br/>写入 stock_deduction_log<br/>记录每个 item 的订单归属和数量] --> SET1[所有 allDetails<br/>加入 deductedSet]
        SET1 --> MARK1[markFullyDeductedOrders<br/>标记 erp_stock_deducted=1]
    end

    PHASE2_OK --> END_NIL
    SUCCESS1 -- 失败 --> PHASE3

    subgraph PHASE3 ["阶段三：重试排除循环（最多5次）"]
        direction TB
        PARSE[parseStockEntryErrorItemCodes<br/>从错误消息中提取 item_code] --> PARSE_OK{解析到<br/>item_code?}
        PARSE_OK -- 否 --> ERR1([return error<br/>无法解析失败item])
        PARSE_OK -- 是 --> EXCLUDE[新 item_code<br/>加入 excludedSet]
        EXCLUDE --> BUILD_RETRY[构建 retryItems<br/>从 mergeMap 中排除 excludedSet]
        BUILD_RETRY --> CHECK_RETRY_EMPTY{retryItems<br/>为空?}

        CHECK_RETRY_EMPTY -- 是 --> ALL_SKIP[所有 item 均被排除<br/>excluded items → deductedSet<br/>markFullyDeductedOrders]
        ALL_SKIP --> END_NIL2([return nil])

        CHECK_RETRY_EMPTY -- 否 --> RETRY_SUBMIT[重新提交 ERPNext<br/>排除问题 item 后的请求]
        RETRY_SUBMIT --> RETRY_OK{成功?}
        RETRY_OK -- 失败 --> CHECK_MAX{达到最大<br/>重试次数?}
        CHECK_MAX -- 否 --> PARSE
        RETRY_OK -- 成功 --> RETRY_SUCCESS
    end

    subgraph RETRY_SUCCESS ["排除后提交成功处理"]
        direction TB
        LOG2[writeDeductionLogs<br/>仅写入成功 item 的日志] --> SET2[成功 item → deductedSet]
        SET2 --> SET3[excluded item → deductedSet<br/>视为已处理 但无日志记录]
        SET3 --> MARK2[markFullyDeductedOrders]
    end

    RETRY_SUCCESS --> END_NIL3([return nil])

    CHECK_MAX -- 是 --> EXHAUSTED

    subgraph EXHAUSTED ["重试耗尽"]
        direction TB
        LAST_PARSE[解析最后一次错误<br/>加入 excludedSet] --> LAST_SET[excluded items → deductedSet]
        LAST_SET --> LAST_MARK[markFullyDeductedOrders]
    end

    EXHAUSTED --> END_ERR([return error<br/>重试N次后仍失败])

    style START fill:#4CAF50,color:#fff
    style END_NIL fill:#8BC34A,color:#fff
    style END_NIL2 fill:#8BC34A,color:#fff
    style END_NIL3 fill:#8BC34A,color:#fff
    style ERR1 fill:#f44336,color:#fff
    style END_ERR fill:#FF9800,color:#fff
```

---

## 2. markFullyDeductedOrders 订单标记逻辑

```mermaid
flowchart LR
    INPUT[遍历所有订单 UUID] --> CHECK{该订单所有<br/>requiredItems<br/>都在 deductedSet?}
    CHECK -- 否 --> SKIP[跳过，不标记]
    CHECK -- 是 --> TYPE{订单类型?}
    TYPE -- 堂食 --> SALE[sale_order 表<br/>erp_stock_deducted = 1]
    TYPE -- 外卖 --> TAKEOUT[takeout_order 表<br/>erp_stock_deducted = 1]
```

**关键**：`__none__` 是无原材料订单的占位符，始终视为"已完成"，因此没有原材料的订单会直接被标记。

---

## 3. 错误解析正则

`parseStockEntryErrorItemCodes` 从 ERPNext 的错误消息中提取 item_code：

| 错误类型 | ERPNext 错误消息模式 | 正则 |
|---------|---------------------|------|
| 库存不足 | `Item Code: <strong>WPR123</strong>` | `Item Code: <strong>([^<]+)</strong>` |
| 非库存物料 | `SP123 is not a stock Item` | `(\S+) is not a stock Item` |
| 物品已禁用 | `Item WPR123 is disabled` | `Item (\S+) is disabled` |

---

## 4. deductedSet vs stock_deduction_log 的区别

这是理解整个流程的核心：

```
┌─────────────────────────────────────────────────────────────────┐
│                        deductedSet (内存)                        │
│  用途: 判断订单是否可标记 erp_stock_deducted=1                     │
│  包含: 成功扣减的 item ✅ + 被排除的 item ✅                        │
│  生命周期: 仅当次执行有效                                          │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                   stock_deduction_log (持久化)                    │
│  用途: 记录实际提交到 ERPNext 的扣减明细                             │
│  包含: 仅成功扣减的 item ✅（被排除的 ❌ 不记录）                     │
│  生命周期: 持久化，反结账时依赖此表回滚                               │
└─────────────────────────────────────────────────────────────────┘
```

**影响**：

| 场景 | deductedSet | stock_deduction_log | 订单标记 | 反结账回滚 |
|------|------------|--------------------|---------|-----------|
| item 成功扣减 | ✅ 标记 | ✅ 写入 | erp_stock_deducted=1 | 创建反向 Stock Entry 恢复库存 |
| item 被排除（disabled/not_stock/insufficient） | ✅ 标记 | ❌ 无记录 | erp_stock_deducted=1 | 不回滚（无实际扣减） |
| item 未处理（解析失败等） | ❌ 未标记 | ❌ 无记录 | 不标记，下次重试 | N/A |

---

## 5. 反结账库存恢复流程

```mermaid
flowchart TD
    REVERSE([反结账触发]) --> RESET[saleOrder.ErpStockDeducted = 0<br/>重置扣减状态]
    RESET --> CHECK_SI{有 Sales Invoice?}
    CHECK_SI -- 否 --> DONE([结束])
    CHECK_SI -- 是 --> CANCEL_SI[取消 Sales Invoice + Payment Entry]
    CANCEL_SI --> LOCK2[获取分布式锁<br/>防止与定时任务竞态]
    LOCK2 --> CHECK_DEDUCTED{erp_stock_deducted<br/>之前 = 1?}

    CHECK_DEDUCTED -- 否 --> CLEAN[清理可能存在的<br/>部分扣减日志]
    CLEAN --> UNLOCK([释放锁])

    CHECK_DEDUCTED -- 是 --> QUERY_LOG[查询 stock_deduction_log<br/>获取实际扣减明细]
    QUERY_LOG --> HAS_LOG{有扣减日志?}
    HAS_LOG -- 否 --> UNLOCK
    HAS_LOG -- 是 --> RECEIPT[创建反向 Stock Entry<br/>Material Receipt<br/>按日志中的 item + qty 回滚]
    RECEIPT --> RECEIPT_OK{成功?}
    RECEIPT_OK -- 成功 --> DELETE_LOG[删除 stock_deduction_log<br/>避免重新结账后二次扣减]
    DELETE_LOG --> UNLOCK
    RECEIPT_OK -- 失败 --> KEEP_LOG[保留 deduction log<br/>不阻塞反结账<br/>日志用于后续重试]
    KEEP_LOG --> UNLOCK

    style REVERSE fill:#FF9800,color:#fff
    style RECEIPT fill:#2196F3,color:#fff
```

**关键**：被排除的 item 没有 `stock_deduction_log` 记录，所以反结账时不会为它们创建反向 Stock Entry — 这是正确的行为，因为这些 item 从未实际扣减过库存。

---

## 6. GenerateStocktakeSnapshot 盘点快照流程

```mermaid
flowchart TD
    START2([盘点快照触发]) --> Q_SALE[查询堂食订单<br/>erp_stock_deducted=0<br/>已结账 + 已开发票]
    Q_SALE --> Q_TAKE[查询外卖订单<br/>erp_stock_deducted=0<br/>已完成 + 已开发票]
    Q_TAKE --> MERGE[从原材料表读取<br/>按 item_code 合并数量]
    MERGE --> CHECK_EMPTY3{mergeMap 为空?}
    CHECK_EMPTY3 -- 是 --> END_NIL4([return nil])
    CHECK_EMPTY3 -- 否 --> DEDUCT_PARTIAL[查询 stock_deduction_log<br/>扣除已部分扣减的数量]
    DEDUCT_PARTIAL --> CALC[mergeMap.Qty -= log.Qty<br/>Qty ≤ 0 则删除该 item]
    CALC --> CHECK_EMPTY4{mergeMap 为空?}
    CHECK_EMPTY4 -- 是 --> END_NIL4
    CHECK_EMPTY4 -- 否 --> SAVE[保存 stocktake_snapshot<br/>每个 item_code 一条记录]
    SAVE --> END_NIL4

    style START2 fill:#9C27B0,color:#fff
```

**与 Stock Entry 的一致性**：快照使用相同数据源（`erp_stock_deducted=0` 的订单），并扣除 `stock_deduction_log` 中已扣减的数量，确保快照反映的是"还需要扣减"的实际数量。

---

## 7. 完整生命周期示例

### 场景：3个订单，5种 item，其中 item_C 被禁用

```
订单 A: item_A(3), item_B(2), item_C(1)
订单 B: item_A(1), item_D(4)
订单 C: item_C(2), item_E(5)

合并后 mergeMap:
  item_A: 4, item_B: 2, item_C: 3, item_D: 4, item_E: 5
```

**执行过程**：

```
1. 整体提交 → ERPNext 返回 "Item item_C is disabled"
2. 解析出 item_C → excludedSet = {item_C}
3. 排除 item_C 后重试 → 提交 item_A(4), item_B(2), item_D(4), item_E(5) → 成功 ✅

4. writeDeductionLogs: 写入 item_A, item_B, item_D, item_E 的扣减日志
5. deductedSet 标记:
   - 成功: A:item_A ✅, A:item_B ✅, B:item_A ✅, B:item_D ✅, C:item_E ✅
   - 排除: A:item_C ✅, C:item_C ✅ （视为已处理）
6. markFullyDeductedOrders:
   - 订单 A: item_A ✅ item_B ✅ item_C ✅ → 全部完成 → erp_stock_deducted=1
   - 订单 B: item_A ✅ item_D ✅ → 全部完成 → erp_stock_deducted=1
   - 订单 C: item_C ✅ item_E ✅ → 全部完成 → erp_stock_deducted=1
```

**反结账订单 A**：

```
1. saleOrder.ErpStockDeducted = 0
2. 查询 stock_deduction_log → 找到 item_A(3), item_B(2)（item_C 无记录）
3. 创建反向 Stock Entry (Material Receipt): item_A(3), item_B(2)
4. item_C 不回滚 — 因为从未实际扣减过 ✅
```
