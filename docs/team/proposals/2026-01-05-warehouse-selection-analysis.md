# 品牌采购流程中门店仓库选择分析

> 本文档分析在品牌采购流程中，门店提交采购申请时是否需要选择总部仓库，以及现有流程设计的原因。

---

## 📋 问题分析

### 核心问题

**从门店视角**：门店提交采购申请时，是否需要进行总部仓库的选择？

**业务场景**：
- 门店创建 Material Request（材料申请单）
- 门店只需要表达"我需要这些物品"
- 门店不需要关心"这些物品从哪里来"（是总部仓库A还是仓库B）

---

## 🎯 结论

### ✅ 门店不需要选择总部仓库

**理由**：

1. **业务职责分离**
   - **门店职责**：提出需求（需要什么物品、需要多少、什么时候需要）
   - **总部职责**：决定供应（从哪个仓库发货、如何分配、如何配送）

2. **信息不对称**
   - 门店不知道总部的库存分布情况
   - 门店不知道总部的仓库优先级
   - 门店不知道总部的批次效期情况
   - 门店无法做出最优的仓库选择决策

3. **灵活性要求**
   - 总部需要根据实时库存情况动态调整发货仓库
   - 总部需要根据分配策略（FIFO、优先级、批次效期）自动分配
   - 如果门店提前选择，会限制总部的灵活性

4. **用户体验**
   - 门店操作越简单越好
   - 减少门店的选择负担
   - 降低操作错误率

---

## 🔍 现有流程分析

### 当前实现

查看代码和文档，发现 Material Request 中有以下字段：

```protobuf
message SaveMaterialRequestReq {
    int64 transaction_date = 1;      // 单据日期,必填
    string company_abbr = 2;         // 公司缩写,必填
    string branch = 3;               // 分支名称 必填
    int64 required_by = 4;           // 需求时间,必填
    string source_warehouse = 5;      // 来源仓库，必填
    string target_warehouse = 6;     // 目标仓库，必填
    string purpose = 7;              // 申请目的,可选 默认 Purchase
    string supplier = 8;               // 供应商名称, purpose 为 Purchases时 必填
    repeated MaterialRequestItem items = 9;  // 物品列表
}
```

**问题**：`source_warehouse` 字段标记为"必填"，要求门店选择总部仓库。

---

## 🤔 为什么现有流程会有这个选择？

### 可能的原因

#### 1. **ERPNext 技术限制**

ERPNext 的 Material Request Item 中，`warehouse` 字段通常是必填的。这个字段在 ERPNext 中的含义是：
- **对于 Purchase 类型的 MR**：`warehouse` 表示"目标仓库"（收货仓库）
- **对于 Material Transfer 类型的 MR**：`warehouse` 可能表示"源仓库"或"目标仓库"

**技术实现**：
- ERPNext 需要知道物品的仓库信息，用于库存计算和单据流转
- 如果不填写，ERPNext 可能无法正常处理

#### 2. **历史设计遗留**

可能之前的设计中：
- 门店需要明确指定从哪个总部仓库要货
- 或者是为了支持某些特殊业务场景（如指定仓库的紧急调拨）

#### 3. **业务场景考虑**

可能存在以下业务场景：
- **场景A**：门店知道某个物品只在特定仓库有库存，需要指定仓库
- **场景B**：门店需要从特定仓库发货（如距离最近的仓库）
- **场景C**：门店需要紧急调拨，指定优先级最高的仓库

但这些场景应该由**总部决策**，而不是门店决策。

---

## 💡 优化方案

### 方案 A：门店只选择目标仓库（推荐）

**设计**：

1. **门店操作**
   - 门店只需要选择 `target_warehouse`（门店自己的仓库，收货仓库）
   - 门店**不需要**选择 `source_warehouse`（总部仓库，发货仓库）

2. **Material Request 字段调整**
   ```protobuf
   message SaveMaterialRequestReq {
       // ... 其他字段 ...
       string target_warehouse = 6;      // 目标仓库（门店仓库），必填
       // source_warehouse 字段移除或改为可选
   }
   ```

3. **总部操作**
   - 总部审批 MR 后，创建 Sales Order 时
   - 总部根据分配策略自动选择 `source_warehouse`（总部仓库）
   - 或者总部手动选择 `source_warehouse`

### 方案 B：使用仓库组（母仓库）作为占位符

**设计**：

1. **Material Request 字段**
   - 门店选择 `warehouse_group`（母仓库，仓库组）
   - 这不是具体的仓库，而是仓库组的标识

2. **Sales Order 创建时**
   - 从 MR 继承 `warehouse_group`
   - 总部在 Sales Order 中分配具体的子仓库

3. **优势**
   - 门店只需要选择一个"仓库组"（类似选择"总部仓库"这个大类）
   - 不需要知道具体的子仓库
   - 符合"门店只选母仓库，总部决定子仓库"的业务需求

---

## 🔍 代码检查结果

### Material Request 中 warehouse 字段的实际使用情况

**检查代码实现**（`ttpos-bmp/app/ttpos-erp/internal/logic/stock/stock.go`）：

1. **Material Request Item 的 `warehouse` 字段**
   ```go
   itemList = append(itemList, g.MapStrAny{
       "item_code":     item.ItemCode,
       "qty":           item.Qty,
       "uom":           item.Uom,
       "schedule_date": service.Setup().MustGetLocalDateTime(ctx, gtime.New(req.RequiredBy)).Format("Y-m-d"),
       "warehouse":     targetWarehouseName,  // 使用的是 target_warehouse（目标仓库）
   })
   ```
   - **使用的是 `target_warehouse`（目标仓库）**，不是 `source_warehouse`（来源仓库）
   - 这个字段在 ERPNext 中**不是必填的**（根据 ERPNext 文档）

2. **`source_warehouse` 字段的使用**
   - 在 `CreateMaterialRequest` 函数中，`source_warehouse` **没有被使用**
   - `source_warehouse` 是在**后续创建 Sales Order 时使用**的（`CreateInnerSaleOrderFromPurchaseOrder`）

3. **Protobuf 定义 vs 实际使用**
   ```protobuf
   string source_warehouse = 5;  // 来源仓库，必填
   ```
   - Protobuf 中标记为"必填"，但**在 Material Request 创建时并没有被使用**
   - 这个字段是在后续流程（创建 Sales Order）中使用的

### 结论

1. **Material Request Item 的 `warehouse` 字段**
   - 在 ERPNext 中**不是必填的**（默认情况下）
   - 我们的代码中设置了 `warehouse = target_warehouse`（目标仓库）
   - **不需要 `source_warehouse`（发货仓库）**

2. **`source_warehouse` 字段**
   - 在 Material Request 创建时**没有被使用**
   - 是在后续创建 Sales Order 时使用的
   - 可以改为可选字段，或者改为 `warehouse_group`（仓库组）

3. **优化建议**
   - 将 `source_warehouse` 改为 `warehouse_group`（仓库组）
   - 或者将 `source_warehouse` 改为可选字段
   - Material Request Item 的 `warehouse` 字段保持使用 `target_warehouse`（目标仓库）

---

## 📊 对比分析

| 方案 | 门店操作 | 总部操作 | 灵活性 | 用户体验 |
|------|---------|---------|--------|---------|
| **现有方案** | 选择具体总部仓库 | 使用门店选择的仓库 | ❌ 低（门店已选择，总部无法调整） | ❌ 差（门店需要了解总部仓库） |
| **方案A** | 只选择门店仓库 | 总部决定发货仓库 | ✅ 高（总部完全控制） | ✅ 好（门店操作简单） |
| **方案B** | 选择仓库组（母仓库） | 总部分配子仓库 | ✅ 高（总部完全控制） | ✅ 好（门店只需选择大类） |

---

## ✅ 推荐方案

### 推荐：方案 B（仓库组方案）

**理由**：

1. **符合业务需求**
   - 门店只需要选择"母仓库"（仓库组），表达"从总部仓库组要货"
   - 总部根据策略分配"子仓库"，决定具体从哪个仓库发货

2. **技术实现简单**
   - 使用自定义字段 `warehouse_group`（仓库组）
   - 不需要修改 ERPNext 标准字段
   - 兼容现有流程

3. **用户体验好**
   - 门店只需要选择一个"仓库组"（如"总部仓库组"）
   - 不需要了解具体的子仓库
   - 操作简单直观

4. **灵活性高**
   - 总部可以根据实时情况动态分配子仓库
   - 支持多种分配策略（FIFO、优先级、批次效期）
   - 不受门店选择的限制

---

## 🔧 实施建议

### 1. Material Request 字段调整

**当前**：
```protobuf
string source_warehouse = 5;  // 来源仓库，必填
```

**调整后**：
```protobuf
string warehouse_group = 5;   // 仓库组（母仓库），必填
// 或者
string source_warehouse = 5;  // 来源仓库，可选（如果填写，则使用仓库组）
```

### 2. 前端界面调整

**门店创建 Material Request 时**：
- 显示"仓库组（母仓库）"字段
- 下拉选项只显示仓库组（`is_group = 1`）
- 不显示具体的子仓库
- 提示："选择总部仓库组，具体发货仓库由总部决定"

### 3. 后端逻辑调整

**创建 Material Request 时**：
- 如果传入 `warehouse_group`，使用仓库组
- 如果传入 `source_warehouse`，检查是否为仓库组，如果是则使用，如果不是则报错

**创建 Sales Order 时**：
- 从 MR 继承 `warehouse_group`
- 总部根据分配策略自动分配子仓库
- 或者总部手动选择子仓库

---

## 📝 总结

### 核心观点

1. **门店不需要选择具体的总部仓库**
   - 门店只需要表达需求（需要什么、需要多少、什么时候需要）
   - 总部负责决定供应（从哪个仓库发货、如何分配）

2. **现有流程的问题**
   - `source_warehouse` 字段要求门店选择具体仓库
   - 限制了总部的灵活性
   - 增加了门店的操作负担

3. **优化方向**
   - 使用仓库组（母仓库）作为门店的选择
   - 总部在 Sales Order 阶段分配具体的子仓库
   - 符合"门店只选母仓库，总部决定子仓库"的业务需求

---

**版本**: v1.0.0  
**创建日期**: 2026-01-05  
**维护者**: TTPOS Team

