# 盘点单 TTPOS 与 ERPNext 数据同步机制

> 详细说明盘点单在 TTPOS 和 ERPNext 之间的数据同步机制

---

## 一、同步概览

### 1.1 同步方向

**单向同步：TTPOS → ERPNext**

- TTPOS 是数据源，ERPNext 是数据接收方
- 只在公司开启 ERP 功能时才会同步
- ERPNext 的数据变更不会回写到 TTPOS

### 1.2 同步时机

盘点单在 TTPOS 中有两个关键操作会触发 ERPNext 同步：

| TTPOS 操作 | ERPNext 操作 | 同步时机 |
|------------|--------------|----------|
| **提交盘点单** | 创建盘点单（保存状态） | 用户点击"提交"时 |
| **审核盘点单** | 提交盘点单（提交状态） | 用户点击"审核通过"时 |

### 1.3 同步条件

同步仅在以下条件同时满足时才会执行：

1. ✅ 公司开启了 ERP 功能（`company.is_open_erp = true`）
2. ✅ 盘点单状态符合要求（提交时：已保存；审核时：已提交）
3. ✅ 盘点单有有效的 ERP 编码（审核时需要）

---

## 二、同步流程详解

### 2.1 提交盘点单时的同步流程

```
TTPOS 用户操作
    ↓
点击"提交盘点单"
    ↓
TTPOS 业务校验
    ├─ 物品状态检查
    ├─ 仓库状态检查
    └─ 物品明细清理（移除禁用/删除物品）
    ↓
构建 ERP 请求数据
    ├─ 公司缩写（CompanyAbbr）
    ├─ 分支名称（Branch）
    ├─ 过账日期（PostingDate）
    ├─ 过账时间（PostingTime）
    ├─ 仓库编码（Warehouse ERP Code）
    └─ 物品明细列表
        ├─ 物品编码（ItemCode）
        └─ 实盘数量（Qty）
    ↓
调用 BMP 模块 gRPC 接口
    ↓
BMP 模块处理
    ├─ 查询 ERPNext 公司信息
    ├─ 查询仓库信息（如果未指定）
    ├─ 查询物品库存（过滤无差异物品）
    └─ 构建 ERPNext 盘点单数据
    ↓
调用 ERPNext API
    ├─ 创建盘点单（状态：Draft）
    └─ 返回盘点单号（MAT-RECO-YYYY-XXXXX）
    ↓
更新 TTPOS 盘点单
    ├─ 保存 ERP 盘点单号（erp_code）
    ├─ 更新状态为"已提交"
    └─ 记录提交时间
```

**关键代码位置**：
- TTPOS Main 模块：`main/app/service/stock_reconciliation.go::submitStockReconciliation()`
- TTPOS ERP 服务：`main/app/service/rpc/erp/stock.go::SubmitStockReconciliation()`
- BMP 模块：`ttpos-bmp/app/ttpos-erp/internal/logic/stock/stock_reconciliation.go::SaveStockReconciliation()`

### 2.2 审核盘点单时的同步流程

```
TTPOS 用户操作
    ↓
点击"审核通过"
    ↓
TTPOS 业务校验
    ├─ 盘点单状态检查（必须是已提交）
    ├─ 仓库状态检查
    ├─ 物品状态检查（如果禁用，返回错误）
    └─ 物品删除检查（自动移除已删除物品）
    ↓
更新 TTPOS 库存
    ├─ 更新仓库物品库存为实盘数量
    └─ 生成盘盈盘亏出入库记录
    ↓
调用 BMP 模块 gRPC 接口
    ├─ 传递 ERP 盘点单号
    └─ 请求提交 ERPNext 盘点单
    ↓
BMP 模块处理
    └─ 调用 ERPNext API 提交盘点单
    ↓
ERPNext 处理
    ├─ 更新盘点单状态为 Submitted
    └─ 更新 ERPNext 库存
    ↓
更新 TTPOS 盘点单状态
    └─ 状态更新为"已审核"
```

**关键代码位置**：
- TTPOS Main 模块：`main/app/service/stock_reconciliation.go::ApproveStockReconciliation()`
- TTPOS ERP 服务：`main/app/service/rpc/erp/stock.go::ApproveStockReconciliation()`
- BMP 模块：`ttpos-bmp/app/ttpos-erp/internal/logic/stock/stock_reconciliation.go::SubmitStockReconciliation()`

---

## 三、数据映射关系

### 3.1 提交时的数据映射

| TTPOS 字段 | ERPNext 字段 | 说明 | 示例 |
|------------|--------------|------|------|
| `company.erpnext_company_abbr` | `company` | 公司缩写 | `Company A` |
| `company.erpnext_branch_name` | `branch` | 分支名称 | `Branch 1` |
| `warehouse.erp_code` | `set_warehouse` | 仓库编码 | `WH-001` |
| `now().Format("2006-01-02")` | `posting_date` | 过账日期（商家时区） | `2025-01-16` |
| `now().Format("15:04:05")` | `posting_time` | 过账时间（商家时区） | `14:30:00` |
| `material.code` | `item_code` | 物品编码 | `MAT-001` |
| `item.counted_quantity` | `qty` | 实盘数量（基准单位） | `100.000` |
| - | `purpose` | 盘点目的（默认：Stock Reconciliation） | `Stock Reconciliation` |
| - | `naming_series` | 编号系列 | `MAT-RECO-.YYYY.-` |

**注意**：
- 物品名称（`item_name`）也会传递，但 ERPNext 会根据 `item_code` 自动填充
- 估值价格（`valuation_rate`）获取优先级：
  1. 请求值（如果 > 0）
  2. Item.ValuationRate（如果 > 0）
  3. Item.StandardRate（如果 > 0）
  4. Item.LastPurchaseRate（如果 > 0）
  5. 如果所有价格都是 0，返回 0 并设置 `allow_zero_valuation_rate = 1`（允许零估值率）
  6. 如果获取 Item 信息失败，使用默认值 1.0（保证向后兼容）
- **所有 TTPOS 中的物品都会同步到 ERPNext**，包括盘0和与账面相同库存的物品，确保数据完全一致
- **支持零估值率**：如果物品的所有价格都是 0，系统会自动设置 `allow_zero_valuation_rate = 1`，以便 ERPNext 接受该盘点记录

### 3.2 审核时的数据映射

| TTPOS 字段 | ERPNext 字段 | 说明 |
|------------|--------------|------|
| `stock_reconciliation.erp_code` | `stock_reconciliation_name` | ERP 盘点单号 |

审核时只传递盘点单号，ERPNext 根据盘点单号找到对应的盘点单并提交。

---

## 四、同步的数据内容

### 4.1 提交时同步的数据

**主表数据**：
```json
{
  "company": "Company A",
  "branch": "Branch 1",
  "posting_date": "2025-01-16",
  "posting_time": "14:30:00",
  "set_warehouse": "WH-001",
  "purpose": "Stock Reconciliation",
  "naming_series": "MAT-RECO-.YYYY.-"
}
```

**明细数据**：
```json
{
  "items": [
    {
      "item_code": "MAT-001",
      "item_name": "大米",
      "qty": 100.000,
      "warehouse": "WH-001",
      "valuation_rate": 1.0
    },
    {
      "item_code": "MAT-002",
      "item_name": "面粉",
      "qty": 50.000,
      "warehouse": "WH-001",
      "valuation_rate": 1.0
    },
    {
      "item_code": "MAT-003",
      "item_name": "免费样品",
      "qty": 10.000,
      "warehouse": "WH-001",
      "valuation_rate": 0.0,
      "allow_zero_valuation_rate": 1
    }
  ]
}
```

**注意**：如果物品的估值率为 0，会自动设置 `allow_zero_valuation_rate = 1`，以便 ERPNext 接受该盘点记录。

### 4.2 审核时同步的数据

```json
{
  "stock_reconciliation_name": "MAT-RECO-2025-00001"
}
```

---

## 五、数据过滤规则

### 5.1 TTPOS 端过滤

在提交到 ERPNext 之前，TTPOS 会过滤以下物品：

1. **已禁用的物品**：`material.status = false`
   - 自动从盘点单中删除（设置 `delete_time`）
   - 不会传递到 ERPNext

2. **已删除的物品**：`item.delete_time > 0`
   - 不会传递到 ERPNext

3. **待盘点物品**：所有单位明细都没有数量
   - 自动从盘点单中删除
   - 不会传递到 ERPNext

### 5.2 ERPNext 端过滤

**注意**：自 2025-01-17 起，BMP 模块不再过滤任何物品，所有 TTPOS 中的物品都会同步到 ERPNext，包括：

- ✅ **盘0的物品**：账面库存为0且实盘数量为0的物品
- ✅ **与账面相同库存的物品**：账面库存与实盘数量一致的物品

这样可以确保 TTPOS 和 ERPNext 的盘点单数据完全一致，避免数据差异。

---

## 六、错误处理机制

### 6.1 同步失败处理

#### 提交时失败

**场景 1：仓库被禁用**
```
错误信息：Disabled Warehouse
处理方式：返回错误，提示"仓库状态已关闭，请修改仓库状态"
TTPOS 状态：保持"已保存"状态，不更新
```

**场景 2：物品被禁用**
```
错误信息：Item XXX is disabled
处理方式：返回错误，提示"物品XXX状态已关闭，请修改物品状态"
TTPOS 状态：保持"已保存"状态，不更新
```

**场景 3：其他错误**
```
错误信息：ERP API 返回的错误信息
处理方式：返回通用错误"提交盘点单失败"
TTPOS 状态：保持"已保存"状态，不更新
```

#### 审核时失败

**场景 1：仓库被禁用**
```
错误信息：Disabled Warehouse
处理方式：回滚事务，返回错误"仓库状态已关闭，请修改仓库状态"
TTPOS 状态：保持"已提交"状态，库存不更新
```

**场景 2：其他错误**
```
错误信息：ERP API 返回的错误信息
处理方式：回滚事务，返回错误"审核盘点单失败"
TTPOS 状态：保持"已提交"状态，库存不更新
```

### 6.2 事务保护

- **提交时**：使用数据库事务，确保 ERP 同步成功后才更新 TTPOS 状态
- **审核时**：使用数据库事务，确保 ERP 同步成功后才更新库存和状态

如果 ERP 同步失败，整个事务回滚，TTPOS 数据保持不变。

---

## 七、ERP 盘点单号管理

### 7.1 ERP 盘点单号格式

ERPNext 生成的盘点单号格式：
```
MAT-RECO-YYYY-XXXXX
```

其中：
- `MAT-RECO`：固定前缀
- `YYYY`：年份（4位）
- `XXXXX`：序列号（5位，从 00001 开始）

示例：`MAT-RECO-2025-00001`

### 7.2 TTPOS 单据编号格式

TTPOS 生成的单据编号格式：
```
ST + YYYYMMDD + 4位序列号
```

示例：`ST202501160001`

### 7.3 关联关系

- TTPOS 盘点单的 `erp_code` 字段存储 ERPNext 盘点单号
- 两个编号相互独立，通过 `erp_code` 关联
- 列表页和详情页都会显示两个编号

---

## 八、同步状态对应关系

### 8.1 状态映射表

| TTPOS 状态 | TTPOS 状态值 | ERPNext 状态 | ERPNext 状态值 | 说明 |
|------------|-------------|--------------|---------------|------|
| 已保存 | 0 | Draft | 0 | 草稿状态，未同步 |
| 已提交 | 1 | Draft | 0 | 已同步创建，但未提交 |
| 已审核 | 2 | Submitted | 1 | 已同步提交 |
| 已驳回 | 3 | Cancelled | 2 | 如果驳回，可同步取消 |

### 8.2 状态流转图

```
TTPOS 状态流转：
已保存(0) → 已提交(1) → 已审核(2)
              ↓
           已驳回(3)

ERPNext 状态流转：
Draft(0) → Submitted(1)
    ↓
Cancelled(2)
```

**注意**：
- TTPOS 的"已提交"对应 ERPNext 的"Draft"（已创建但未提交）
- TTPOS 的"已审核"对应 ERPNext 的"Submitted"（已提交）

---

## 九、数据一致性保障

### 9.1 同步时机的一致性

- **提交时**：TTPOS 提交 → ERPNext 创建（Draft）
- **审核时**：TTPOS 审核 → ERPNext 提交（Submitted）

确保两个系统的状态流转保持一致。

### 9.2 数据内容的一致性

- **物品编码**：使用 TTPOS 的 `material.code`（与 ERPNext 的 Item Code 一致）
- **仓库编码**：使用 TTPOS 的 `warehouse.erp_code`（与 ERPNext 的 Warehouse 一致）
- **数量**：使用 TTPOS 的实盘数量（`counted_quantity`），已换算为基准单位

### 9.3 库存更新的一致性

- **TTPOS**：审核时直接更新为实盘数量
- **ERPNext**：提交盘点单后，ERPNext 会根据盘点单自动更新库存

两个系统的库存更新逻辑一致，确保数据同步。

---

## 十、特殊场景处理

### 10.1 公司未开启 ERP

如果公司未开启 ERP（`is_open_erp = false`）：

- ✅ 可以正常创建、编辑、提交、审核盘点单
- ✅ TTPOS 库存正常更新
- ❌ 不会同步到 ERPNext
- ❌ `erp_code` 字段为空

### 10.2 ERP 盘点单号为空

如果盘点单的 `erp_code` 为空：

- ✅ 可以正常审核（如果公司未开启 ERP）
- ❌ 如果公司开启了 ERP，审核时会跳过 ERP 同步（不会报错）

### 10.3 物品在 ERPNext 中不存在

如果物品编码在 ERPNext 中不存在：

- ❌ ERPNext API 会返回错误
- ❌ TTPOS 提交/审核失败
- ✅ 用户需要先在 ERPNext 中创建该物品

### 10.4 仓库在 ERPNext 中不存在

如果仓库编码在 ERPNext 中不存在：

- ❌ ERPNext API 会返回错误
- ❌ TTPOS 提交/审核失败
- ✅ 用户需要先在 ERPNext 中创建该仓库

---

## 十一、技术实现细节

### 11.1 调用链路

```
TTPOS Main 模块
    ↓ (gRPC)
BMP 模块（ttpos-erp）
    ↓ (HTTP API)
ERPNext 系统
```

### 11.2 通信协议

- **TTPOS → BMP**：gRPC（Protocol Buffers）
- **BMP → ERPNext**：HTTP REST API（JSON）

### 11.3 数据转换

- **TTPOS → BMP**：Go struct → Protobuf message
- **BMP → ERPNext**：Protobuf message → JSON

### 11.4 时区处理

- **过账日期和时间**：使用商家时区（`company_setting.timezone`）
- **格式**：
  - 日期：`2006-01-02`（YYYY-MM-DD）
  - 时间：`15:04:05`（HH:MM:SS）

---

## 十二、监控与日志

### 12.1 日志记录

**成功日志**：
- TTPOS：记录 ERP 盘点单号
- BMP：记录 ERPNext API 调用成功

**失败日志**：
- TTPOS：记录错误信息（包含物品名称、仓库名称等）
- BMP：记录 ERPNext API 错误响应

### 12.2 错误追踪

- 所有 ERP 同步错误都会记录到日志系统
- 错误信息包含完整的调用链路信息
- 支持通过 `erp_code` 追踪对应的 ERPNext 盘点单

---

## 十三、总结

### 13.1 同步特点

1. **单向同步**：TTPOS → ERPNext，不反向同步
2. **条件同步**：只在公司开启 ERP 时同步
3. **事务保护**：使用数据库事务，确保数据一致性
4. **错误处理**：完善的错误处理和回滚机制
5. **数据过滤**：自动过滤无效数据，确保同步质量

### 13.2 数据一致性

- ✅ 状态流转一致
- ✅ 物品编码一致
- ✅ 仓库编码一致
- ✅ 数量数据一致
- ✅ 库存更新一致

### 13.3 注意事项

1. **物品和仓库必须在 ERPNext 中存在**，否则同步会失败
2. **物品和仓库必须处于启用状态**，否则同步会失败
3. **ERP 盘点单号一旦生成，不会变更**，即使驳回重新提交也不会重新生成
4. **如果 ERP 同步失败，TTPOS 操作也会失败**，需要先解决 ERP 问题
5. **支持零估值率物品盘点**：如果物品的所有价格都是 0，系统会自动设置 `allow_zero_valuation_rate = 1`，以便 ERPNext 接受该盘点记录
6. **估值率获取逻辑**：优先使用请求值（如果 > 0），如果未设置或为 0，则从 Item 获取；如果 Item 的所有价格都是 0，返回 0 并允许零估值率

---

**最后更新**：2025-01-16  
**维护者**：TTPOS Team

