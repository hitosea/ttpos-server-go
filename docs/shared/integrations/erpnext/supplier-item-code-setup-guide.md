# ERPNext 供应商物品编码设置和使用指南

> 📖 **用途**: 详细说明如何在 ERPNext 中设置和使用供应商物品编码（`supplier_part_no`）

---

## 一、概述

### 1.1 什么是供应商物品编码

供应商物品编码（`supplier_part_no`）是供应商对同一物品使用的编码。例如：
- **系统内部编码**：`ITEM-001`（苹果）
- **供应商A编码**：`SUPPLIER-A-APPLE-001`
- **供应商B编码**：`SUPPLIER-B-APPLE-001`

### 1.2 业务价值

- ✅ **沟通桥梁**：连接系统内部编码与供应商编码
- ✅ **准确性**：减少采购、收货、对账中的编码错误
- ✅ **追溯性**：支持质量追溯与供应商管理
- ✅ **效率**：提高采购与对账效率

---

## 二、在 Item 中设置供应商编码

### 2.1 通过 UI 界面设置

#### 步骤 1：打开物品主数据

1. **登录 ERPNext 系统**
2. **导航路径**：
   ```
   库存（Stock） → 物品（Item） → 新建（New）
   ```
   或者编辑现有物品：`库存 → 物品 → 选择物品`

#### 步骤 2：填写物品基本信息

在物品表单中填写：
- **物品编码**（Item Code）：`ITEM-001`
- **物品名称**（Item Name）：`苹果`
- **物品组**（Item Group）：`水果`
- 其他必要字段...

#### 步骤 3：添加供应商信息

1. **滚动到"供应商"（Suppliers）部分**
   - 位置：物品表单底部，通常在"价格"（Pricing）部分之后

2. **点击"添加行"（Add Row）按钮**

3. **填写供应商信息**：

| 字段 | 说明 | 示例值 |
|------|------|--------|
| **供应商**（Supplier）* | 选择供应商 | `Supplier A - Company` |
| **供应商物品编码**（Supplier Part No） | **供应商对该物品的编码** ⭐ | `SUPPLIER-A-APPLE-001` |
| **供应商物品名称**（Supplier Item Name） | 可选，供应商对该物品的名称 | `苹果（供应商A）` |
| **最小订单数量**（Min Order Qty） | 可选，最小采购数量 | `10` |
| **最后采购价格**（Last Purchase Rate） | 可选，最后一次采购价格 | `10.00` |

4. **继续添加其他供应商**：
   - 再次点击"添加行"
   - 选择另一个供应商，填写对应的供应商编码

**示例配置**：

| 供应商 | 供应商物品编码 | 供应商物品名称 | 最小订单数量 | 最后采购价格 |
|--------|---------------|---------------|-------------|-------------|
| Supplier A - Company | `SUPPLIER-A-APPLE-001` | 苹果（供应商A） | 10 | 10.00 |
| Supplier B - Company | `SUPPLIER-B-APPLE-001` | 苹果（供应商B） | 20 | 9.50 |
| Supplier C - Company | `SUPPLIER-C-APPLE-001` | 苹果（供应商C） | 15 | 10.50 |

#### 步骤 4：保存物品

1. 点击右上角"保存"按钮
2. 系统会验证数据
3. 保存成功后，供应商编码已关联到物品

### 2.2 通过 API 设置

#### 方法 1：创建物品时同时设置供应商

```python
# 创建物品并设置供应商编码
POST /api/resource/Item
{
    "item_code": "ITEM-001",
    "item_name": "苹果",
    "item_group": "水果",
    "stock_uom": "Nos",
    "suppliers": [
        {
            "supplier": "Supplier A - Company",
            "supplier_part_no": "SUPPLIER-A-APPLE-001",
            "supplier_item_name": "苹果（供应商A）",
            "min_order_qty": 10,
            "last_purchase_rate": 10.00
        },
        {
            "supplier": "Supplier B - Company",
            "supplier_part_no": "SUPPLIER-B-APPLE-001",
            "supplier_item_name": "苹果（供应商B）",
            "min_order_qty": 20,
            "last_purchase_rate": 9.50
        }
    ]
}
```

#### 方法 2：更新现有物品，添加供应商

```python
# 更新物品，添加供应商编码
PUT /api/resource/Item/ITEM-001
{
    "suppliers": [
        {
            "supplier": "Supplier A - Company",
            "supplier_part_no": "SUPPLIER-A-APPLE-001",
            "supplier_item_name": "苹果（供应商A）"
        }
    ]
}
```

#### 方法 3：使用 TTPOS 系统同步

在 TTPOS 系统中，通过同步供应商物品列表，系统会自动从 ERPNext 获取供应商物品关联关系：

```go
// 代码位置：main/app/service/supplier.go
// 同步供应商物品列表
erpSupplierItems, err := erpSrv.GetSupplierItemList(ctx, req.GetErpnextSupplierItemListReq{
    SiteCode: companySetting.ErpnextSiteCode,
    Supplier: supplier.ErpCode,
})
```

---

## 三、在采购订单中使用

### 3.1 自动填充机制

当创建采购订单并选择供应商后，ERPNext 会自动从 Item Supplier 子表中获取该供应商对应的 `supplier_part_no`，并填充到采购订单行项目中。

### 3.2 通过 UI 创建采购订单

#### 步骤 1：创建采购订单

1. **导航路径**：
   ```
   采购（Buying） → 采购订单（Purchase Order） → 新建（New）
   ```

2. **填写基本信息**：
   - **供应商**（Supplier）：选择 `Supplier A - Company`
   - **交易日期**（Transaction Date）：`2025-01-15`
   - **计划日期**（Schedule Date）：`2025-01-25`
   - **仓库**（Set Warehouse）：`Stores - Company`

#### 步骤 2：添加物品

1. **在"物品"（Items）部分，点击"添加行"**

2. **选择物品**：
   - **物品编码**（Item Code）：选择 `ITEM-001`（苹果）
   - **数量**（Qty）：`100`
   - **单位**（UOM）：`Nos`

3. **系统自动填充**：
   - ✅ **供应商物品编码**（Supplier Part No）：自动填充为 `SUPPLIER-A-APPLE-001`
   - ✅ **价格**（Rate）：自动填充为 `10.00`（从 Last Purchase Rate 获取）

4. **查看供应商编码**：
   - 在采购订单行项目中，可以看到"供应商物品编码"字段已自动填充
   - 这个编码会显示在采购订单 PDF 中，方便与供应商沟通

#### 步骤 3：保存并提交

1. 点击"保存"按钮
2. 点击"提交"按钮
3. 采购订单创建完成，供应商编码已包含在订单中

### 3.3 通过 API 创建采购订单

```python
# 创建采购订单
POST /api/resource/Purchase Order
{
    "supplier": "Supplier A - Company",
    "transaction_date": "2025-01-15",
    "schedule_date": "2025-01-25",
    "set_warehouse": "Stores - Company",
    "items": [
        {
            "item_code": "ITEM-001",
            "qty": 100,
            "rate": 10.00,
            "uom": "Nos",
            "warehouse": "Stores - Company"
            // 注意：supplier_part_no 会自动从 Item Supplier 子表获取
            // 不需要手动指定
        }
    ]
}

# 提交采购订单
POST /api/resource/Purchase Order/PO-00001
{
    "action": "submit"
}
```

**重要说明**：
- 在 API 中创建采购订单时，不需要手动指定 `supplier_part_no`
- ERPNext 会根据选择的供应商和物品，自动从 Item Supplier 子表中获取对应的 `supplier_part_no`
- 如果物品没有为该供应商设置 `supplier_part_no`，该字段将为空

---

## 四、在采购收货单中使用

### 4.1 自动继承机制

创建采购收货单（Purchase Receipt）时，系统会自动从关联的采购订单中继承 `supplier_part_no`。

### 4.2 通过 UI 创建收货单

#### 步骤 1：从采购订单创建收货单

1. **打开采购订单**：
   ```
   采购（Buying） → 采购订单（Purchase Order） → 选择订单
   ```

2. **点击"创建"（Create）按钮**，选择"采购收货单"（Purchase Receipt）

3. **系统自动填充**：
   - ✅ 供应商信息
   - ✅ 物品信息
   - ✅ **供应商物品编码**（自动从采购订单继承）

#### 步骤 2：核对收货信息

1. **检查供应商编码**：
   - 在收货单行项目中，确认"供应商物品编码"是否正确
   - 与实际收到的货物包装上的编码进行核对

2. **核对数量**：
   - 检查实际收货数量是否与订单一致
   - 如有差异，修改数量

#### 步骤 3：提交收货单

1. 点击"保存"按钮
2. 点击"提交"按钮
3. 库存自动更新，供应商编码已记录在收货单中

### 4.3 通过 API 创建收货单

```python
# 从采购订单创建收货单
POST /api/resource/Purchase Receipt
{
    "supplier": "Supplier A - Company",
    "purchase_order": "PO-00001",
    "posting_date": "2025-01-20",
    "posting_time": "10:00:00",
    "set_warehouse": "Stores - Company",
    "items": [
        {
            "item_code": "ITEM-001",
            "qty": 100,
            "uom": "Nos",
            "warehouse": "Stores - Company",
            "purchase_order_item": "PO-ITEM-001"
            // supplier_part_no 会自动从采购订单继承
        }
    ]
}

# 提交收货单
POST /api/resource/Purchase Receipt/PR-00001
{
    "action": "submit"
}
```

---

## 五、在采购发票中使用

### 5.1 自动继承机制

创建采购发票（Purchase Invoice）时，系统会自动从关联的采购订单或收货单中继承 `supplier_part_no`。

### 5.2 通过 UI 创建发票

#### 步骤 1：从采购订单或收货单创建发票

1. **打开采购订单或收货单**

2. **点击"创建"（Create）按钮**，选择"采购发票"（Purchase Invoice）

3. **系统自动填充**：
   - ✅ 供应商信息
   - ✅ 物品信息
   - ✅ **供应商物品编码**（自动继承）

#### 步骤 2：核对发票信息

1. **检查供应商编码**：
   - 确认发票中的供应商编码与供应商提供的发票一致
   - 如有差异，需要核对

2. **核对价格和数量**：
   - 检查发票价格是否与订单一致
   - 检查发票数量是否与收货一致

#### 步骤 3：提交发票

1. 点击"保存"按钮
2. 点击"提交"按钮
3. 发票创建完成，供应商编码已记录

### 5.3 通过 API 创建发票

```python
# 从采购订单创建发票
POST /api/resource/Purchase Invoice
{
    "supplier": "Supplier A - Company",
    "purchase_order": "PO-00001",
    "posting_date": "2025-01-25",
    "posting_time": "10:00:00",
    "set_warehouse": "Stores - Company",
    "items": [
        {
            "item_code": "ITEM-001",
            "qty": 100,
            "rate": 10.00,
            "uom": "Nos",
            "warehouse": "Stores - Company",
            "purchase_order_item": "PO-ITEM-001"
            // supplier_part_no 会自动从采购订单继承
        }
    ]
}

# 提交发票
POST /api/resource/Purchase Invoice/PI-00001
{
    "action": "submit"
}
```

---

## 六、查询和报表

### 6.1 查询物品的供应商编码

#### 通过 UI 查询

1. **打开物品主数据**
2. **查看"供应商"部分**：
   - 可以看到该物品关联的所有供应商
   - 每个供应商对应的 `supplier_part_no`

#### 通过 API 查询

```python
# 获取物品信息（包含供应商编码）
GET /api/resource/Item/ITEM-001

# 响应示例
{
    "item_code": "ITEM-001",
    "item_name": "苹果",
    "suppliers": [
        {
            "supplier": "Supplier A - Company",
            "supplier_part_no": "SUPPLIER-A-APPLE-001",
            "supplier_item_name": "苹果（供应商A）",
            "last_purchase_rate": 10.00
        },
        {
            "supplier": "Supplier B - Company",
            "supplier_part_no": "SUPPLIER-B-APPLE-001",
            "supplier_item_name": "苹果（供应商B）",
            "last_purchase_rate": 9.50
        }
    ]
}
```

### 6.2 按供应商编码查询采购记录

```python
# 查询包含特定供应商编码的采购订单
GET /api/resource/Purchase Order
{
    "filters": [
        ["items", "supplier_part_no", "=", "SUPPLIER-A-APPLE-001"]
    ]
}
```

### 6.3 供应商绩效报表

可以通过供应商编码分析：
- 不同供应商的采购频率
- 不同供应商的价格对比
- 不同供应商的交付质量

---

## 七、常见问题

### 7.1 为什么设置了供应商物品编码，但没有生效？⚠️

这是最常见的问题，以下是详细的排查步骤和解决方法：

#### 🔍 排查步骤 1：检查物品主数据配置

**问题**：物品主数据中未正确设置供应商编码

**检查方法**：
1. 打开物品主数据（`库存 → 物品 → 选择物品`）
2. 滚动到"供应商"（Suppliers）部分
3. 检查是否已添加供应商行
4. 检查"供应商物品编码"（Supplier Part No）字段是否有值

**常见错误**：
- ❌ 只填写了"供应商"字段，但忘记填写"供应商物品编码"
- ❌ 供应商编码字段留空
- ❌ 保存时未点击"保存"按钮，数据未保存

**解决方法**：
1. 确保在"供应商"子表中，**必须填写"供应商物品编码"字段**
2. 保存物品后，再次打开确认数据已保存

#### 🔍 排查步骤 2：检查供应商名称是否完全匹配

**问题**：采购订单中的供应商名称与物品主数据中的供应商名称不一致

**检查方法**：
1. 打开物品主数据，查看"供应商"部分中的供应商名称
2. 打开采购订单，查看"供应商"字段的值
3. 对比两者是否**完全一致**（包括公司后缀）

**常见错误**：
- ❌ 物品中设置的是：`Supplier A - Company`
- ❌ 采购订单中选择的是：`Supplier A`（缺少公司后缀）
- ❌ 或者相反：物品中是 `Supplier A`，采购订单中是 `Supplier A - Company`

**解决方法**：
1. 确保采购订单中的供应商名称与物品主数据中的供应商名称**完全一致**
2. 建议在物品主数据中使用完整的供应商名称（包含公司后缀），如：`Supplier A - Company`

#### 🔍 排查步骤 3：检查是否在添加物品前选择了供应商

**问题**：在创建采购订单时，先添加了物品，后选择了供应商

**ERPNext 的自动填充机制**：
- ERPNext 在**选择物品时**，会根据**当前已选择的供应商**，自动从 Item Supplier 子表中查找对应的 `supplier_part_no`
- 如果先添加物品，后选择供应商，系统无法自动填充编码

**正确的操作顺序**：
1. ✅ **先选择供应商**（在采购订单头部）
2. ✅ **再添加物品**（在物品行中）
3. ✅ 系统会自动填充供应商编码

**错误的操作顺序**：
1. ❌ 先添加物品
2. ❌ 后选择供应商
3. ❌ 系统无法自动填充编码

**解决方法**：
1. 删除已添加的物品行
2. 先选择供应商
3. 重新添加物品
4. 系统会自动填充供应商编码

#### 🔍 排查步骤 4：检查系统缓存和刷新

**问题**：系统缓存导致新设置的数据未立即生效

**解决方法**：
1. **刷新页面**：按 `F5` 或 `Ctrl+R` 刷新浏览器
2. **清除浏览器缓存**：按 `Ctrl+Shift+Delete` 清除缓存
3. **重新登录**：退出系统后重新登录
4. **清除 ERPNext 缓存**（管理员操作）：
   ```
   设置（Settings） → 系统设置（System Settings） → 清除缓存（Clear Cache）
   ```

#### 🔍 排查步骤 5：检查字段是否在表单中显示

**问题**：供应商编码字段在采购订单表单中未显示

**检查方法**：
1. 打开采购订单表单
2. 在物品行中，查看是否有"供应商物品编码"（Supplier Part No）列
3. 如果没有，需要自定义表单显示该字段

**解决方法**（详细步骤）：

**步骤 1：打开自定义表单**
1. 登录 ERPNext 系统（需要管理员权限）
2. 导航路径（根据 ERPNext 版本可能略有不同）：
   - **方法 A**：`设置（Settings） → 自定义（Customize） → 自定义表单（Customize Form）`
   - **方法 B**：`主页（Home） → 设置（Settings） → 自定义（Customize） → 表单定制器（Form Customizer）`
   - **方法 C**：在搜索框中输入 `Customize Form` 或 `自定义表单`

**步骤 2：选择表单类型**
1. 在"表单类型"（Form Type）或"文档类型"（DocType）下拉菜单中
2. 选择 **`Purchase Order Item`**（采购订单物品）
3. 点击"自定义"（Customize）或"编辑"（Edit）按钮

**步骤 3：查找并配置字段**
1. 在字段列表中找到 `supplier_part_no` 字段
   - 如果字段列表很长，可以使用搜索框搜索 `supplier` 或 `供应商`
2. 点击该字段，展开其详细设置
3. 在字段属性中，找到并勾选以下选项（选项名称可能因版本而异）：
   - ✅ **"在列表中显示"**（In List View / Show in List）
   - ✅ **"在表格中显示"**（In Grid View / Show in Table / Show in Table View）
   - ✅ **"在表格中可见"**（Visible in Grid）
   
   **注意**：不同版本的 ERPNext 选项名称可能不同：
   - 英文版：`In List View`、`In Grid View`、`Show in Table`
   - 中文版：`在列表中显示`、`在表格中显示`、`表格中可见`
   
4. 如果找不到这些选项，可以尝试：
   - 勾选 **"可见"**（Visible）或 **"显示"**（Show）
   - 设置 **"在表格中"**（In Table）为 `是`（Yes）

**步骤 4：保存更改**
1. 点击页面顶部的 **"保存"**（Save）按钮
2. 系统会提示"表单已更新"或类似消息
3. 刷新采购订单页面，字段应该已经显示

**如果仍然找不到字段或选项**：

**方法 1：检查字段是否存在**
- 在自定义表单中，如果找不到 `supplier_part_no` 字段，可能需要先添加该字段
- 点击"添加字段"（Add Field）或"添加自定义字段"（Add Custom Field）
- 选择 `supplier_part_no` 字段并添加到表单

**方法 2：使用列表视图设置**
- 导航到：`采购（Buying） → 采购订单（Purchase Order）`
- 在列表视图中，点击右上角的"列"（Columns）或"设置列"（Set Columns）
- 查找并勾选"供应商物品编码"（Supplier Part No）

**方法 3：检查用户权限**
- 确认当前用户有查看和编辑采购订单的权限
- 导航到：`设置 → 用户和权限 → 角色权限管理器`
- 检查角色是否有"采购订单"的查看和编辑权限

**方法 4：清除缓存后重试**
- 清除浏览器缓存：`Ctrl+Shift+Delete`
- 清除 ERPNext 缓存：`设置 → 系统设置 → 清除缓存`
- 重新登录系统

#### 🔍 排查步骤 6：检查 ERPNext 版本和自定义代码

**问题**：ERPNext 版本过旧或自定义代码影响了自动填充功能

**检查方法**：
1. 查看 ERPNext 版本：`设置 → 关于（About）`
2. 检查是否有自定义的 Python 脚本或 JavaScript 客户端脚本影响了采购订单

**解决方法**：
1. **更新 ERPNext**：升级到最新版本
2. **检查自定义代码**：临时禁用自定义脚本，测试是否正常
3. **联系技术支持**：如果问题持续，联系 ERPNext 社区或技术支持

#### 🔍 排查步骤 7：通过 API 验证数据

**问题**：通过 UI 设置的数据可能未正确保存

**验证方法**：

```python
# 1. 查询物品的供应商编码
GET /api/resource/Item/ITEM-001

# 检查响应中的 suppliers 数组
{
    "item_code": "ITEM-001",
    "suppliers": [
        {
            "supplier": "Supplier A - Company",
            "supplier_part_no": "SUPPLIER-A-APPLE-001"  # 确认这个字段有值
        }
    ]
}

# 2. 创建测试采购订单
POST /api/resource/Purchase Order
{
    "supplier": "Supplier A - Company",  # 确保与物品中的供应商名称完全一致
    "transaction_date": "2025-01-15",
    "items": [
        {
            "item_code": "ITEM-001",
            "qty": 1
        }
    ]
}

# 3. 查询创建的采购订单，检查 supplier_part_no 是否自动填充
GET /api/resource/Purchase Order/PO-00001

# 检查响应中的 items 数组
{
    "items": [
        {
            "item_code": "ITEM-001",
            "supplier_part_no": "SUPPLIER-A-APPLE-001"  # 确认这个字段有值
        }
    ]
}
```

#### 📋 快速检查清单

在排查问题时，按以下清单逐一检查：

- [ ] **物品主数据中已设置供应商编码**
  - [ ] 供应商子表中有数据行
  - [ ] "供应商物品编码"字段有值（不为空）
  - [ ] 已保存物品

- [ ] **供应商名称完全匹配**
  - [ ] 物品中的供应商名称：`Supplier A - Company`
  - [ ] 采购订单中的供应商名称：`Supplier A - Company`
  - [ ] 两者完全一致（包括大小写和公司后缀）

- [ ] **操作顺序正确**
  - [ ] 先选择供应商（在采购订单头部）
  - [ ] 再添加物品（在物品行中）

- [ ] **字段已显示**
  - [ ] 采购订单物品行中可以看到"供应商物品编码"列
  - [ ] 如果看不到，需要自定义表单显示该字段

- [ ] **系统缓存已清除**
  - [ ] 已刷新页面
  - [ ] 已重新登录系统

#### 🎯 最可能的原因（按概率排序）

1. **80% 概率**：操作顺序错误（先添加物品，后选择供应商）
2. **10% 概率**：供应商名称不匹配（缺少公司后缀或大小写不一致）
3. **5% 概率**：物品主数据中未填写供应商编码字段
4. **3% 概率**：系统缓存问题
5. **2% 概率**：字段未在表单中显示或系统版本问题

#### ✅ 标准解决方案

**如果以上排查都正常，但仍未生效，请按以下步骤操作**：

1. **重新设置物品供应商编码**：
   - 打开物品主数据
   - 删除现有的供应商行
   - 重新添加供应商行，确保填写"供应商物品编码"
   - 保存物品

2. **创建新的测试采购订单**：
   - 先选择供应商（确保名称与物品中的完全一致）
   - 再添加物品
   - 检查是否自动填充供应商编码

3. **如果仍不生效，通过 API 直接设置**：
   ```python
   # 在创建采购订单时，手动指定 supplier_part_no
   POST /api/resource/Purchase Order
   {
       "supplier": "Supplier A - Company",
       "items": [
           {
               "item_code": "ITEM-001",
               "supplier_part_no": "SUPPLIER-A-APPLE-001",  # 手动指定
               "qty": 100
           }
       ]
   }
   ```

---

### 7.2 为什么采购订单中没有显示供应商编码？

**可能原因**：
1. 物品没有为该供应商设置 `supplier_part_no`
2. 采购订单中的供应商与物品关联的供应商不一致

**解决方法**：
1. 检查物品的"供应商"部分，确认是否已为该供应商设置编码
2. 如果未设置，按照"二、在 Item 中设置供应商编码"的步骤添加
3. 参考"7.1 为什么设置了供应商物品编码，但没有生效？"的详细排查步骤

### 7.3 如何批量设置供应商编码？

**方法 1：通过 Excel 导入**

1. 导出物品数据（包含供应商信息）
2. 在 Excel 中批量填写 `supplier_part_no`
3. 导入更新

**方法 2：通过 API 批量更新**

```python
# 批量更新多个物品的供应商编码
items = ["ITEM-001", "ITEM-002", "ITEM-003"]
for item_code in items:
    PUT /api/resource/Item/{item_code}
    {
        "suppliers": [
            {
                "supplier": "Supplier A - Company",
                "supplier_part_no": f"SUPPLIER-A-{item_code}"
            }
        ]
    }
```

### 7.4 供应商编码可以修改吗？

**可以修改**，但需要注意：
- 修改后，新的采购订单会使用新的编码
- 历史订单中的编码不会改变（保持历史记录）
- 建议在修改前通知相关人员

### 7.5 一个物品可以关联多少个供应商？

**理论上没有限制**，但建议：
- 每个物品关联 3-5 个主要供应商
- 过多的供应商会增加管理复杂度

---

## 八、最佳实践

### 8.1 编码规范

- ✅ **统一格式**：建议使用统一的编码格式，如 `{供应商简称}-{物品类别}-{序号}`
- ✅ **唯一性**：确保同一供应商的物品编码唯一
- ✅ **可读性**：编码应具有一定的可读性，便于识别

### 8.2 维护建议

- ✅ **定期更新**：当供应商更换编码时，及时更新
- ✅ **验证机制**：在创建采购订单前，验证供应商编码是否存在
- ✅ **文档记录**：记录编码变更历史，便于追溯

### 8.3 与供应商沟通

- ✅ **明确编码**：在首次合作时，明确供应商的物品编码
- ✅ **书面确认**：将编码信息写入采购合同或协议
- ✅ **定期核对**：定期与供应商核对编码，确保一致性

---

## 九、TTPOS 系统集成

### 9.1 同步供应商物品列表

在 TTPOS 系统中，可以通过以下方式同步供应商物品关联关系：

```go
// 代码位置：main/app/service/supplier.go
// 同步供应商物品列表
func (s *supplierSrv) SyncSupplierItemList(ctx context.Context, req req.SyncSupplierItemListReq) error {
    // 获取供应商列表
    supplierList, err := s.supplierRepo.GetList(...)
    
    // 遍历每个供应商
    for _, supplier := range supplierList {
        // 从 ERPNext 获取供应商物品列表
        erpSupplierItems, err := erpSrv.GetSupplierItemList(ctx, req.GetErpnextSupplierItemListReq{
            SiteCode: companySetting.ErpnextSiteCode,
            Supplier: supplier.ErpCode,
        })
        
        // 同步到 TTPOS 数据库
        for _, supplierErp := range erpSupplierItems {
            materialSupplierRepo.Create(&model.MaterialSupplier{
                MaterialUuid:    material.Uuid,
                MaterialCode:    material.Code,
                SupplierUuid:    supplier.Uuid,
                SupplierErpCode: supplier.ErpCode,
            })
        }
    }
}
```

### 9.2 在采购订单中使用

在 TTPOS 系统中创建采购订单时，系统会自动从 ERPNext 获取供应商编码：

```go
// 创建采购订单时，系统会：
// 1. 从 ERPNext 获取物品的供应商编码
// 2. 填充到采购订单行项目中
// 3. 在打印 PDF 时显示供应商编码
```

---

## 十、总结

供应商物品编码（`supplier_part_no`）是 ERPNext 中重要的业务字段，通过正确设置和使用，可以：

1. ✅ **提高准确性**：减少采购、收货、对账中的编码错误
2. ✅ **提升效率**：加快与供应商的沟通速度
3. ✅ **增强追溯性**：支持质量追溯和供应商管理
4. ✅ **优化流程**：简化采购和财务流程

**关键操作步骤**：
1. 在 Item 中设置供应商编码（UI 或 API）
2. 创建采购订单时自动填充
3. 收货单和发票自动继承
4. 定期维护和更新

---

**最后更新**: 2025-01-15  
**维护者**: TTPOS Team

