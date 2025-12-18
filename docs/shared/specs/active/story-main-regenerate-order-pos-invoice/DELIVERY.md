# 重新生成订单POS发票 - 交付文档

> 本文档用于交付重新生成订单POS发票工具，说明工具功能、使用方法和处理流程。

## 📋 工具概述

**命令名称**: `regenerate-order-pos-invoice`

**功能描述**: 重新生成指定销售订单的POS发票。当订单的POS发票因ERP系统异常、网络问题等原因未能正确生成或保存时，可以通过此工具快速重新生成发票，无需重新走完整结账流程。

**使用场景**:
- ERP系统异常导致发票未能正确生成
- 网络问题导致发票保存失败
- 发票数据需要修复或重新生成
- 批量处理前的单订单测试

## 🎯 处理流程

工具会执行以下步骤：

1. **读取订单信息**: 从数据库读取订单（`saleOrder`）和账单（`saleBill`）信息
2. **验证订单状态**: 检查订单是否已完成结账（`FinishTime != 0`）
3. **验证ERP配置**: 检查ERP Phase3是否启用，SiteCode是否配置
4. **获取班次信息**: 获取订单关联的班次记录（`shiftLog`）
5. **生成POS发票**: 调用 `SavePosInvoice` 方法生成发票到ERP系统
6. **更新订单信息**: 更新订单的发票信息字段（`ErpProductsInvoiceName`、`ErpMaterialInvoiceName`）

**数据影响**: 
- 在ERP系统中创建新的POS发票
- 更新订单表中的发票信息字段（`ErpProductsInvoiceName`、`ErpMaterialInvoiceName`）

**重要说明**: 
- ⚠️ **非幂等性操作**：该命令每调用一次，就会为这个订单创建一次新的发票。重复调用会导致同一订单产生多张发票。
- 旧的发票由ERP系统自己去删除或作废，本工具不处理旧发票的清理
- 建议在执行前使用 `--dry-run` 预览模式确认订单信息

## 📝 使用方法

### 基本用法

```bash
./ttpos-server-go regenerate-order-pos-invoice \
  --company-uuid 7709131161600000 \
  --sale-order-uuid 202512151410092620 \
  --open-pos-entry-name "POS-ENTRY-001"
```

### 参数说明

| 参数 | 必填 | 说明 | 示例 |
|------|------|------|------|
| `--company-uuid` | ✅ | 门店UUID | `7709131161600000` |
| `--sale-order-uuid` | ✅ | 销售订单UUID | `202512151410092620` |
| `--open-pos-entry-name` | ✅ | OpenPosEntryName（用于生成POS发票） | `"POS-ENTRY-001"` |
| `--dry-run` | ❌ | 预览模式，仅预览不实际执行 | - |

### 常用场景

#### 1. 预览模式（推荐先使用）

在执行实际操作前，建议先使用预览模式查看订单信息：

```bash
./ttpos-server-go regenerate-order-pos-invoice \
  --company-uuid 7709131161600000 \
  --sale-order-uuid 202512151410092620 \
  --open-pos-entry-name "POS-ENTRY-001" \
  --dry-run
```

预览模式会显示：
- 订单号
- 订单金额
- ERP SiteCode
- ERP CompanyAbbr
- 将执行的操作说明

#### 2. 实际执行

确认无误后，执行实际操作：

```bash
./ttpos-server-go regenerate-order-pos-invoice \
  --company-uuid 7709131161600000 \
  --sale-order-uuid 202512151410092620 \
  --open-pos-entry-name "POS-ENTRY-001"
```

执行时会要求输入 `yes` 确认，确认后才会执行发票生成操作。

执行成功后会显示：
- 商品发票名称（`ProductsInvoiceName`）
- 材料发票名称（`MaterialInvoiceName`）
- 执行耗时（`DurationMs`）

## 🔒 安全机制

1. **参数验证**: 所有必填参数都会进行验证，缺失或无效时会提示错误并退出
2. **订单状态验证**: 只处理已完成结账的订单，未完成结账的订单会提示错误
3. **ERP配置验证**: 检查ERP Phase3是否启用，SiteCode是否配置
4. **用户确认**: 实际执行前需要输入 `yes` 确认，避免误操作
5. **预览模式**: 支持 `--dry-run` 预览模式，可以先查看订单信息再决定是否执行

## ⚠️ 注意事项

1. **非幂等性**: 每次执行都会创建新的发票，重复执行会导致同一订单产生多张发票
2. **旧发票处理**: 旧的发票由ERP系统自己去删除或作废，本工具不处理旧发票的清理
3. **班次检查**: 如果订单关联的班次已交班，会提示错误并退出
4. **数据备份**: 建议在执行操作前备份相关数据

## 🐛 故障排查

### 问题1: 订单未完成结账

**错误信息**: `订单未完成结账，无法生成发票`

**原因**: 订单的 `FinishTime` 为 0，表示订单尚未完成结账

**解决方案**: 确认订单是否已完成结账，或使用其他方式完成结账后再执行

### 问题2: ERP Phase3未启用

**错误信息**: `ERP Phase3未启用，无法生成发票`

**原因**: 门店的ERP Phase3功能未启用

**解决方案**: 检查门店的ERP配置，确认ERP Phase3是否已启用

### 问题3: ERP SiteCode未配置

**错误信息**: `ERP SiteCode未配置，无法生成发票`

**原因**: 门店的ERP SiteCode未配置

**解决方案**: 检查门店的ERP配置，配置正确的SiteCode

### 问题4: 获取 shiftLog 失败

**错误信息**: `获取 shiftLog 失败,当前门店没有当班记录`

**原因**: 订单关联的收银员没有当班记录，或班次记录不存在

**解决方案**: 确认订单关联的收银员是否有当班记录，或联系技术支持

### 问题5: 班次已交班

**错误信息**: `当前班次已交班，无法保存发票`

**原因**: 订单关联的班次已经交班

**解决方案**: 联系技术支持，可能需要手动处理或使用其他方式

## 📚 相关文档

- **需求文档**: `docs/shared/specs/active/story-main-regenerate-order-pos-invoice/requirements.md`
- **设计文档**: `docs/shared/specs/active/story-main-regenerate-order-pos-invoice/design.md`
- **任务清单**: `docs/shared/specs/active/story-main-regenerate-order-pos-invoice/tasks.md`

## 📅 交付信息

| 项目 | 内容 |
|------|------|
| **交付日期** | 2025-12-17 |
| **版本** | v1.0.0 |
| **负责人** | xiezhihuan |
| **状态** | ✅ 已完成 |

---

**最后更新**: 2025-12-17

