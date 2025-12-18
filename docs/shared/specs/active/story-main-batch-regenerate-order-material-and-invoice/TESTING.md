# 单个订单测试操作指南

> 本文档说明如何使用4个独立的命令单独对某个订单进行测试操作，用于验证功能或排查问题。

## 📋 概述

批量重新生成工具内部使用了4个独立的命令，这些命令也可以单独使用来测试单个订单。本文档说明如何使用这些命令进行测试。

## 🎯 4个独立命令

### 命令1: 重新生成订单材料消耗

**命令名称**: `regenerate-order-material`

**功能**: 重新计算并保存订单的材料消耗记录

**参数**:
- `--company-uuid` (必填): 门店UUID
- `--sale-order-uuid` (必填): 销售订单UUID
- `--dry-run` (可选): 预览模式，仅预览不实际执行

**使用示例**:

```bash
# 预览模式（推荐先使用）
./ttpos-server-go regenerate-order-material \
  --company-uuid 7709131161600000 \
  --sale-order-uuid 202512151410092620 \
  --dry-run

# 实际执行
./ttpos-server-go regenerate-order-material \
  --company-uuid 7709131161600000 \
  --sale-order-uuid 202512151410092620
```

**数据影响**: 
- 更新 `sale_order_material` 表（软删除旧记录，插入新记录）

**幂等性**: ✅ 可视为幂等操作，重复执行仅会多出一些软删除的记录

---

### 命令2: 重新生成订单材料出库

**命令名称**: `regenerate-sale-order-material-outbound`

**功能**: 重新生成订单的材料出库记录

**参数**:
- `--company-uuid` (必填): 门店UUID
- `--sale-order-uuid` (必填): 销售订单UUID
- `--dry-run` (可选): 预览模式，仅预览不实际执行

**使用示例**:

```bash
# 预览模式（推荐先使用）
./ttpos-server-go regenerate-sale-order-material-outbound \
  --company-uuid 7709131161600000 \
  --sale-order-uuid 202512151410092620 \
  --dry-run

# 实际执行
./ttpos-server-go regenerate-sale-order-material-outbound \
  --company-uuid 7709131161600000 \
  --sale-order-uuid 202512151410092620
```

**数据影响**: 
- 更新 `warehouse_out_form_item` 表（软删除旧记录，创建新记录）
- 更新 `warehouse_item` 表（退回库存时增加库存数量，扣减库存时减少库存数量）

**幂等性**: ✅ 可视为幂等操作，重复执行仅会多出一些软删除的记录

---

### 命令3: 重新生成订单POS发票

**命令名称**: `regenerate-order-pos-invoice`

**功能**: 重新生成订单的POS发票数据

**参数**:
- `--company-uuid` (必填): 门店UUID
- `--sale-order-uuid` (必填): 销售订单UUID
- `--open-pos-entry-name` (必填): OpenPosEntryName（用于生成POS发票）
- `--dry-run` (可选): 预览模式，仅预览不实际执行

**使用示例**:

```bash
# 预览模式（推荐先使用）
./ttpos-server-go regenerate-order-pos-invoice \
  --company-uuid 7709131161600000 \
  --sale-order-uuid 202512151410092620 \
  --open-pos-entry-name "POS-ENTRY-001" \
  --dry-run

# 实际执行（需要输入 'yes' 确认）
./ttpos-server-go regenerate-order-pos-invoice \
  --company-uuid 7709131161600000 \
  --sale-order-uuid 202512151410092620 \
  --open-pos-entry-name "POS-ENTRY-001"
```

**数据影响**: 
- 在ERP系统中创建新的POS发票
- 更新订单表中的发票信息字段（`ErpProductsInvoiceName`、`ErpMaterialInvoiceName`）

**幂等性**: ⚠️ **非幂等操作**，每次执行都会创建新的发票，重复执行会导致同一订单产生多张发票

**注意**: 
- 执行前需要输入 `yes` 确认
- 旧的发票由ERP系统自己去删除或作废，本命令不处理旧发票的清理

---

### 命令4: 重新生成销售出库汇总（日期级别）

**命令名称**: `regenerate-sales-outbound`

**功能**: 重新生成指定日期的销售出库汇总数据

**参数**:
- `--company-uuid` (必填): 门店UUID
- `--date` (必填): 日期，格式：YYYY-MM-DD
- `--dry-run` (可选): 预览模式，仅预览不实际执行

**使用示例**:

```bash
# 预览模式（推荐先使用）
./ttpos-server-go regenerate-sales-outbound \
  --company-uuid 7709131161600000 \
  --date 2025-12-15 \
  --dry-run

# 实际执行（需要输入 'yes' 确认）
./ttpos-server-go regenerate-sales-outbound \
  --company-uuid 7709131161600000 \
  --date 2025-12-15
```

**数据影响**: 
- 更新 `warehouse_in_out_log` 表（软删除旧记录，创建新记录）
- 更新 `sale_order_material` 表（更新 `is_summarized` 字段，标记为已统计）

**幂等性**: ✅ 可视为幂等操作，重复执行仅会多出一些软删除的记录

**注意**: 
- 这是日期级别的操作，会处理该日期所有订单的汇总数据
- 执行前需要输入 `yes` 确认

---

## 🔄 完整测试流程

### 场景1: 测试单个订单的完整流程

按照以下顺序执行命令，模拟批量工具的处理流程：

```bash
# 步骤1: 重新生成订单材料消耗
./ttpos-server-go regenerate-order-material \
  --company-uuid 7709131161600000 \
  --sale-order-uuid 202512151410092620

# 步骤2: 重新生成订单材料出库
./ttpos-server-go regenerate-sale-order-material-outbound \
  --company-uuid 7709131161600000 \
  --sale-order-uuid 202512151410092620

# 步骤3: 重新生成订单POS发票（可选，注意非幂等）
./ttpos-server-go regenerate-order-pos-invoice \
  --company-uuid 7709131161600000 \
  --sale-order-uuid 202512151410092620

# 步骤4: 重新生成日期级别的销售出库汇总（需要等待该日期所有订单处理完成）
./ttpos-server-go regenerate-sales-outbound \
  --company-uuid 7709131161600000 \
  --date 2025-12-15
```

### 场景2: 仅测试材料相关功能（不生成发票）

如果只需要测试材料消耗和出库功能，可以跳过步骤3：

```bash
# 步骤1: 重新生成订单材料消耗
./ttpos-server-go regenerate-order-material \
  --company-uuid 7709131161600000 \
  --sale-order-uuid 202512151410092620

# 步骤2: 重新生成订单材料出库
./ttpos-server-go regenerate-sale-order-material-outbound \
  --company-uuid 7709131161600000 \
  --sale-order-uuid 202512151410092620

# 步骤3: 重新生成日期级别的销售出库汇总
./ttpos-server-go regenerate-sales-outbound \
  --company-uuid 7709131161600000 \
  --date 2025-12-15
```

### 场景3: 使用预览模式验证

在执行实际操作前，建议先使用 `--dry-run` 参数预览：

```bash
# 预览所有步骤
./ttpos-server-go regenerate-order-material \
  --company-uuid 7709131161600000 \
  --sale-order-uuid 202512151410092620 \
  --dry-run

./ttpos-server-go regenerate-sale-order-material-outbound \
  --company-uuid 7709131161600000 \
  --sale-order-uuid 202512151410092620 \
  --dry-run

./ttpos-server-go regenerate-order-pos-invoice \
  --company-uuid 7709131161600000 \
  --sale-order-uuid 202512151410092620 \
  --dry-run

./ttpos-server-go regenerate-sales-outbound \
  --company-uuid 7709131161600000 \
  --date 2025-12-15 \
  --dry-run
```

---

## ⚠️ 注意事项

### 1. 执行顺序

- **推荐按顺序执行**：步骤1 → 步骤2 → 步骤3 → 步骤4
- **可以跳过不需要的步骤**：如果确保某个步骤的数据已经是正确的，可以跳过该步骤
- **依赖关系说明**：
  - 步骤2依赖步骤1的结果（材料消耗记录），如果步骤1的数据正确，可以跳过步骤1直接执行步骤2
  - 步骤3**不依赖**步骤1和步骤2，可以独立执行，仅需要订单信息即可生成POS发票
  - 步骤4依赖该日期所有订单的步骤1和步骤2都完成，如果该日期所有订单的材料数据都正确，可以跳过步骤1和步骤2直接执行步骤4

### 2. 日期级别操作

- 步骤4（`regenerate-sales-outbound`）是日期级别的操作
- 需要等待该日期所有订单的步骤1和步骤2都完成后才能执行
- 如果只测试单个订单，建议先完成该订单的步骤1和步骤2，再执行步骤4

### 3. POS发票生成

- ⚠️ **步骤3不是幂等的**，每次执行都会创建新的发票
- 建议仅在必要时执行，避免重复创建发票
- 执行前会要求输入 `yes` 确认

### 4. 幂等性说明

- ✅ 步骤1、步骤2、步骤4可以重复执行，最终结果一致
- ⚠️ 步骤3不能重复执行，每次都会创建新发票

### 5. 数据备份

- 建议在执行前备份相关数据
- 特别是执行步骤3（POS发票）前，因为会产生新的发票记录

---

## 🐛 故障排查

### 问题1: 命令执行失败

**检查项**:
- 确认门店UUID和订单UUID是否正确
- 确认订单是否存在且已完成结账（步骤3需要）
- 查看日志文件：`logs/batch-regenerate-{timestamp}.log`

### 问题2: 步骤2执行失败（库存不足）

**原因**: 退回库存时发现库存不足

**解决方案**:
- 检查 `warehouse_item` 表的库存数据
- 确认是否有其他操作同时修改了库存
- 检查订单的材料出库记录是否正确

### 问题3: 步骤3执行失败（ERP配置问题）

**原因**: ERP系统未配置或未启用

**检查项**:
- 确认公司是否启用了 ERP Phase3
- 确认 `ErpnextSiteCode` 是否配置
- 检查ERP系统连接是否正常

### 问题4: 步骤4执行失败（日期格式错误）

**原因**: 日期格式不正确

**解决方案**:
- 确认日期格式为 `YYYY-MM-DD`（例如：`2025-12-15`）
- 确认日期在订单的创建日期范围内

---

## 📚 相关文档

- **批量工具文档**: `DELIVERY.md` - 批量重新生成工具的完整文档
- **需求文档**: `requirements.md` - 功能需求详细说明
- **设计文档**: `design.md` - 技术设计和实现方案

---

**最后更新**: 2025-12-17

