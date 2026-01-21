# 品牌采购仓库选择优化方案

> 优化品牌采购流程，门店只需提出采购要求，不关注具体发货仓库；总部在销售订单阶段选择子仓库。

---

## 📋 需求分析

### 核心需求

1. **门店操作简化**
   - 品牌采购时，门店只提出采购要求（需要什么物品、需要多少、什么时候需要）
   - 门店不关注总部由什么仓库发货
   - 门店不需要选择具体的总部仓库

2. **总部操作增强**
   - 在总部对品牌采购自动生成的销售订单进行操作时，进行子仓库的选择
   - 总部可以根据库存情况、分配策略等灵活选择子仓库

3. **系统自动处理**
   - 门店不需要选择仓库，系统自动使用母仓库（仓库组）
   - 门店界面不显示仓库选择字段，简化操作

---

## 📖 用户故事

### 用户故事 1

**作为** 门店采购员  
**我想** 在创建品牌采购申请时，不需要选择总部仓库  
**以便于** 简化操作流程，避免因不了解总部库存分布而做出错误选择

**详细说明**：
- 当前痛点：门店需要从下拉列表中选择"总部仓库A"或"总部仓库B"，但门店不知道哪个仓库有库存，容易选错
- 优化方案：系统自动使用默认母仓库（仓库组），门店界面不显示仓库选择字段
- 预期价值：
  - ✅ 操作步骤减少，从 5 步减少到 4 步
  - ✅ 消除门店的困惑和选择负担
  - ✅ 降低操作错误率
  - ✅ 提升用户体验

---

### 用户故事 2

**作为** 总部采购经理  
**我想** 在处理品牌采购订单时，能够根据实时库存情况灵活选择发货仓库  
**以便于** 提高库存利用率，减少缺货风险，支持多仓库组合发货

**详细说明**：
- 当前痛点：门店已选择仓库（如"总部仓库A"），但该仓库库存不足，需要手动修改或联系门店重新提交
- 优化方案：系统自动设置母仓库（仓库组），总部在 Sales Order 阶段根据库存情况选择具体的子仓库
- 预期价值：
  - ✅ 总部可以根据实时库存情况灵活分配
  - ✅ 支持多仓库组合发货（不同物品从不同仓库发货）
  - ✅ 提高库存利用率
  - ✅ 减少缺货风险

---

### 用户故事 3

**作为** 总部采购经理  
**我想** 在处理大量品牌采购订单时，能够使用自动分配功能快速分配仓库  
**以便于** 大幅提升处理效率，减少人工选择的工作量，确保分配策略的一致性

**详细说明**：
- 场景：门店提交了大量品牌采购申请，需要快速处理并分配仓库
- 优化方案：系统提供"自动分配仓库"功能，根据分配策略（FIFO、优先级、批次效期）自动选择最优子仓库
- 预期价值：
  - ✅ 大幅提升处理效率
  - ✅ 减少人工选择的工作量
  - ✅ 确保分配策略的一致性
  - ✅ 支持批量处理

---

### 用户故事 4

**作为** 门店店长  
**我想** 在紧急采购时，能够快速提交采购申请，不需要选择仓库  
**以便于** 提高响应速度，确保紧急需求能够及时处理

**详细说明**：
- 场景：门店临时需要紧急采购一批食材，今天就要到货
- 优化方案：门店快速提交申请（不需要选择仓库），系统自动创建 Sales Order，总部根据距离和库存选择最优仓库发货
- 预期价值：
  - ✅ 门店操作简单，快速提交申请
  - ✅ 总部可以根据实际情况（距离、库存）灵活选择最优仓库
  - ✅ 提高紧急订单的处理效率

---

### 用户故事 5

**作为** 总部采购经理  
**我想** 在处理多个门店的批量采购申请时，能够使用批量自动分配功能  
**以便于** 提高处理效率，系统自动优化分配方案，降低配送成本

**详细说明**：
- 场景：多个门店同时提交了品牌采购申请，需要统一处理
- 优化方案：系统提供批量自动分配功能，考虑各子仓库的库存情况、各门店的地理位置、配送成本等因素，自动生成最优分配方案
- 预期价值：
  - ✅ 支持批量处理，提高效率
  - ✅ 系统自动优化分配方案
  - ✅ 降低配送成本
  - ✅ 提高库存周转率

---

### 用户故事总结

| 用户故事 | 角色 | 需求 | 价值 |
|---------|------|------|------|
| **故事 1** | 门店采购员 | 不需要选择仓库 | ✅ 操作简单，无需选择仓库 |
| **故事 2** | 总部采购经理 | 灵活分配仓库 | ✅ 可以根据库存情况灵活分配 |
| **故事 3** | 总部采购经理 | 自动分配仓库 | ✅ 支持自动分配和批量处理 |
| **故事 4** | 门店店长 | 快速提交紧急采购 | ✅ 快速提交，总部快速响应 |
| **故事 5** | 总部采购经理 | 批量自动分配 | ✅ 批量处理，自动优化分配 |

---

## 🔍 当前实现分析

### 当前流程

1. **Material Request 创建**
   - Protobuf 定义中，`source_warehouse` 字段标记为**必填**
   - 实际代码中，`source_warehouse` 在创建 Material Request 时**未被使用**
   - Material Request Item 的 `warehouse` 字段使用的是 `target_warehouse`（目标仓库，门店仓库）

2. **Sales Order 创建**
   - 从 Material Request 创建 Sales Order 时，使用 `CreateInnerSaleOrderFromPurchaseOrder` 方法
   - 该方法中，`source_warehouse` 用于设置 Sales Order 的 `SetWarehouse` 字段
   - 如果未提供 `source_warehouse`，则使用默认仓库

### 问题点

1. **Protobuf 定义与实际使用不一致**
   - `source_warehouse` 标记为必填，但创建 Material Request 时未使用
   - 造成前端必须传递该字段，但后端不使用，形成冗余

2. **门店操作负担**
   - 门店需要选择总部仓库，但门店不知道总部的库存分布情况
   - 门店无法做出最优的仓库选择决策
   - 门店不应该承担仓库选择的职责

3. **总部灵活性受限**
   - 如果门店提前选择了仓库，会限制总部的灵活性
   - 总部无法根据实时库存情况动态调整发货仓库

---

## 💡 解决方案

### 方案概述

**核心思路**：
1. **移除门店仓库选择**：门店不需要选择仓库，系统自动使用母仓库（仓库组）作为标识
2. **后端自动处理**：创建 Material Request 时，系统自动获取并设置默认母仓库到自定义字段 `warehouse_group`
3. **继承到 Sales Order**：在创建 Sales Order 时，自动继承母仓库信息到自定义字段 `warehouse_group`
4. **总部分配子仓库**：总部在 Sales Order 阶段选择具体的子仓库（**重要**：母仓库不能直接用于收发，必须选择子仓库）

**⚠️ 重要技术限制**：
- **母仓库（仓库组）不能直接用于库存记账和收发**
- 母仓库只是逻辑分组标识（`is_group = 1`），不存放实物
- 实际库存和收发操作必须在子仓库（`is_group = 0`）中进行
- Sales Order 的 `SetWarehouse` 字段必须设置为具体的子仓库，不能设置为母仓库

### 详细方案

#### 1. Protobuf 定义调整

**当前定义**：
```protobuf
message SaveMaterialRequestReq {
  int64 transaction_date = 1;      // 单据日期,必填
  string company_abbr = 2;        // 公司缩写,必填
  string branch = 3;               // 分支名称 必填
  int64 required_by = 4;           // 需求时间,必填
  string source_warehouse = 5;    // 来源仓库，必填  ⚠️ 问题点
  string target_warehouse = 6;     // 目标仓库，必填
  string purpose = 7;             // 申请目的,可选 默认 Purchase
  string supplier = 8;            // 供应商名称, purpose 为 Purchases时 必填
  repeated MaterialRequestItem items = 9;  // 物品列表
}
```

**调整后定义**：
```protobuf
message SaveMaterialRequestReq {
  int64 transaction_date = 1;      // 单据日期,必填
  string company_abbr = 2;        // 公司缩写,必填
  string branch = 3;              // 分支名称 必填
  int64 required_by = 4;           // 需求时间,必填
  // 移除 source_warehouse 字段，系统自动使用母仓库 ✅
  string target_warehouse = 6;    // 目标仓库，必填
  string purpose = 7;             // 申请目的,可选 默认 Purchase
  string supplier = 8;            // 供应商名称, purpose 为 Purchases时 必填
  repeated MaterialRequestItem items = 9;  // 物品列表
}
```

**变更说明**：
- **移除 `source_warehouse` 字段**：门店不需要选择仓库
- **系统自动处理**：后端自动获取默认母仓库（仓库组）并设置到 Material Request 的自定义字段 `warehouse_group`
- **简化门店操作**：门店界面不再显示仓库选择字段，只需填写物品和数量
- **⚠️ 注意**：`warehouse_group` 只是作为标识，不能用于实际收发，实际收发必须使用子仓库

**技术验证**：
- ✅ ERPNext 的自定义字段 `Link → Warehouse` 类型**可以**存储仓库组（`is_group = 1`）
- ✅ 通过 API 创建 Material Request 时，可以成功设置 `warehouse_group` 为仓库组名称
- ⚠️ **需要验证**：确保设置的仓库名称确实是仓库组（`is_group = 1`），否则会导致后续流程错误

#### 2. Material Request 创建逻辑调整

**当前实现**（`ttpos-bmp/app/ttpos-erp/internal/logic/stock/stock.go`）：

```go
func (s *sStock) CreateMaterialRequest(ctx context.Context, req *stock.SaveMaterialRequestReq) (res *stock.SaveMaterialRequestResp, err error) {
    // ... 现有代码 ...
    // source_warehouse 字段未被使用
}
```

**调整后实现**：

```go
func (s *sStock) CreateMaterialRequest(ctx context.Context, req *stock.SaveMaterialRequestReq) (res *stock.SaveMaterialRequestResp, err error) {
    // ... 现有代码 ...
    
    // 系统自动获取默认母仓库（仓库组）
    warehouseGroup, err := service.Warehouse().GetDefaultWarehouseGroup(ctx, companyName.CompanyName)
    if err != nil {
        return nil, gerror.Wrapf(err, "获取默认仓库组失败")
    }
    
    // ⚠️ 重要验证：确保获取的是仓库组（is_group = 1），不是真实仓库
    if !warehouseGroup.IsGroup {
        return nil, gerror.Newf("默认仓库 '%s' 不是仓库组（is_group = 0），必须是仓库组（is_group = 1）", warehouseGroup.Name)
    }
    
    warehouseGroupName := warehouseGroup.Name
    
    // 将 warehouse_group 存储到 Material Request 的自定义字段中
    // 注意：
    // 1. 需要在 ERPNext 中添加自定义字段 warehouse_group（类型：Link → Warehouse）
    // 2. ERPNext 允许在 Link → Warehouse 类型的自定义字段中存储仓库组（is_group = 1）
    // 3. warehouse_group 只是作为标识，不能用于实际收发
    // 4. Material Request Item 的 warehouse 字段仍然使用 target_warehouse（门店仓库）
    data["warehouse_group"] = warehouseGroupName
}
```

#### 3. Sales Order 创建逻辑调整

**当前实现**（`ttpos-bmp/app/ttpos-erp/internal/logic/buying/buying.go`）：

```go
func (*sBuying) CreateInnerSaleOrderFromPurchaseOrder(ctx context.Context, req *dto.CreateInnerSaleOrderFromPurchaseOrderReq) (res *erp.SaleOrder, err error) {
    // ... 现有代码 ...
    
    //设置来源仓库
    if req.SourceWarehouse != "" {
        salesOrder.SetWarehouse = req.SourceWarehouse
    } else {
        warehouse, err := service.Warehouse().GetDefaultWarehouse(ctx, salesOrder.Company, "")
        if err != nil {
            return nil, gerror.Wrapf(err, "查询默认仓库失败")
        }
        salesOrder.SetWarehouse = warehouse.Name
    }
}
```

**调整后实现**：

```go
func (*sBuying) CreateInnerSaleOrderFromPurchaseOrder(ctx context.Context, req *dto.CreateInnerSaleOrderFromPurchaseOrderReq) (res *erp.SaleOrder, err error) {
    // ... 现有代码 ...
    
    // 从 Material Request 获取 warehouse_group
    var warehouseGroupName string
    if req.MaterialRequestName != "" {
        mr, err := service.Document().Get(ctx, erp.DocTypeMaterialRequest, req.MaterialRequestName)
        if err == nil {
            warehouseGroupName = mr.Get("warehouse_group").String()
        }
    }
    
    // 设置仓库组（母仓库）到自定义字段
    if warehouseGroupName != "" {
        // 在 Sales Order 中设置 warehouse_group 自定义字段（仅作为标识）
        // 注意：需要在 ERPNext 中添加自定义字段 warehouse_group
        salesOrder.WarehouseGroup = warehouseGroupName
        
        // ⚠️ 重要：SetWarehouse 必须设置为具体的子仓库，不能设置为母仓库
        // 母仓库（仓库组）不能直接用于库存记账和收发
        childWarehouses, err := service.Warehouse().GetChildWarehouses(ctx, warehouseGroupName)
        if err == nil && len(childWarehouses) > 0 {
            // 如果有子仓库，优先使用第一个子仓库作为默认值
            // 总部后续可以根据实际情况修改为其他子仓库
            salesOrder.SetWarehouse = childWarehouses[0].Name
        } else {
            // 如果没有子仓库，使用默认仓库（必须是真实仓库，不能是仓库组）
            warehouse, err := service.Warehouse().GetDefaultWarehouse(ctx, salesOrder.Company, "")
            if err != nil {
                return nil, gerror.Wrapf(err, "查询默认仓库失败")
            }
            // 验证默认仓库不是仓库组
            if warehouse.IsGroup {
                return nil, gerror.New("默认仓库不能是仓库组，必须是真实仓库")
            }
            salesOrder.SetWarehouse = warehouse.Name
        }
    } else {
        // 如果没有 warehouse_group，使用默认仓库（必须是真实仓库）
        warehouse, err := service.Warehouse().GetDefaultWarehouse(ctx, salesOrder.Company, "")
        if err != nil {
            return nil, gerror.Wrapf(err, "查询默认仓库失败")
        }
        // 验证默认仓库不是仓库组
        if warehouse.IsGroup {
            return nil, gerror.New("默认仓库不能是仓库组，必须是真实仓库")
        }
        salesOrder.SetWarehouse = warehouse.Name
    }
}
```

#### 4. ERPNext 自定义字段配置

**Material Request 自定义字段**：

1. **Customize Form** → Material Request
2. **添加字段**：

| 字段名 | 字段类型 | 标签 | 说明 |
|--------|---------|------|------|
| `warehouse_group` | Link → Warehouse | 仓库组（母仓库） | 系统自动设置，门店不可见，过滤条件：`is_group = 1` |

**字段配置**：
```python
{
    "fieldname": "warehouse_group",
    "fieldtype": "Link",
    "label": "仓库组（母仓库）",
    "options": "Warehouse",
    "filters": [
        ["is_group", "=", 1]
    ],
    "reqd": 0,  # 系统自动设置，不需要门店填写
    "read_only": 1,  # 只读，门店不可见
    "hidden": 1  # 隐藏，门店界面不显示
}
```

**Sales Order 自定义字段**：

1. **Customize Form** → Sales Order
2. **添加字段**：

| 字段名 | 字段类型 | 标签 | 说明 |
|--------|---------|------|------|
| `warehouse_group` | Link → Warehouse | 仓库组（母仓库） | 从 Material Request 继承 |
| `source_warehouse` | Link → Warehouse | 发货仓库（子仓库） | 总部选择，过滤条件：`is_group = 0` 且 `parent_warehouse = warehouse_group` |

**字段配置**：
```python
# warehouse_group 字段配置
{
    "fieldname": "warehouse_group",
    "fieldtype": "Link",
    "label": "仓库组（母仓库）",
    "options": "Warehouse",
    "filters": [
        ["is_group", "=", 1]
    ],
    "reqd": 0,
    "read_only": 1  # 从 MR 继承，只读
}

# source_warehouse 字段配置
{
    "fieldname": "source_warehouse",
    "fieldtype": "Link",
    "label": "发货仓库（子仓库）",
    "options": "Warehouse",
    "filters": [
        ["is_group", "=", 0],
        ["parent_warehouse", "=", "warehouse_group"]  # 动态过滤
    ],
    "reqd": 0,  # 总部填写
    "read_only": 0
}
```

**Sales Order Item 自定义字段**：

| 字段名 | 字段类型 | 标签 | 说明 |
|--------|---------|------|------|
| `source_warehouse_item` | Link → Warehouse | 发货仓库（子仓库） | 总部选择，过滤条件：`is_group = 0` 且 `parent_warehouse = warehouse_group` |

#### 5. 前端界面调整

**门店创建 Material Request 时**：

1. **移除仓库选择字段**
   - **完全移除** `source_warehouse` 或 `warehouse_group` 字段
   - 门店界面不显示任何仓库选择相关字段
   - 系统自动使用默认母仓库，门店无需关心

2. **简化操作流程**
   - 门店只需填写：物品、数量、需求时间
   - 仓库信息由系统自动处理，门店无需操作
   - 提示信息（可选）："仓库由系统自动分配，具体发货仓库由总部决定。"

**总部操作 Sales Order 时**：

1. **显示仓库组信息**
   - 显示从 Material Request 继承的 `warehouse_group`（只读）

2. **子仓库选择**
   - 在 Sales Order Item 行上显示 `source_warehouse_item` 字段
   - 下拉选项只显示该仓库组下的子仓库
   - 支持手动选择或自动分配

3. **自动分配按钮**（可选）
   - 添加"自动分配仓库"按钮
   - 根据分配策略（FIFO、优先级、批次效期）自动分配子仓库

---

## ⚠️ 风险评估

### 1. 兼容性风险

**风险描述**：
- 修改 Protobuf 定义后，需要重新生成代码
- 前端需要同步更新，如果前端未更新，可能导致字段不匹配

**影响程度**：**高**

**缓解措施**：
1. **向后兼容处理**
   - 旧版本前端如果传递 `source_warehouse` 字段，后端**忽略该字段**，直接使用默认母仓库
   - 新版本前端不传递任何仓库字段，后端自动使用默认母仓库
   - 确保新旧版本都能正常工作

2. **渐进式迁移**
   - 第一阶段：后端忽略 `source_warehouse` 字段，统一使用默认母仓库
   - 第二阶段：前端完全移除 `source_warehouse` 字段传递

3. **API 版本控制**（可选）
   - 如果使用 API 版本控制，旧版本接口忽略 `source_warehouse`，新版本接口不接收该字段
   - 统一使用默认母仓库逻辑

**实施建议**：
```go
// 兼容处理代码示例
func (s *sStock) CreateMaterialRequest(ctx context.Context, req *stock.SaveMaterialRequestReq) (res *stock.SaveMaterialRequestResp, err error) {
    // 无论前端是否传递 source_warehouse 或 warehouse_group，都忽略
    // 统一使用默认母仓库
    
    // 系统自动获取默认母仓库（仓库组）
    warehouseGroup, err := service.Warehouse().GetDefaultWarehouseGroup(ctx, companyName.CompanyName)
    if err != nil {
        return nil, gerror.Wrapf(err, "获取默认仓库组失败")
    }
    warehouseGroupName := warehouseGroup.Name
    
    // 将 warehouse_group 存储到 Material Request 的自定义字段中
    data["warehouse_group"] = warehouseGroupName
    
    // ... 后续处理 ...
}
```

### 2. ERPNext 自定义字段配置风险

**风险描述**：
- 需要在 ERPNext 中添加自定义字段
- 如果 ERPNext 环境不同步，可能导致字段不存在

**影响程度**：**中**

**缓解措施**：
1. **字段存在性检查**
   - 在创建单据前，检查自定义字段是否存在
   - 如果不存在，记录警告日志，但不影响主流程

2. **配置文档**
   - 提供详细的 ERPNext 配置文档
   - 包含字段配置的 JSON 导出文件
   - 提供配置脚本，一键导入配置

3. **环境同步**
   - 确保开发、测试、生产环境的 ERPNext 配置一致
   - 使用 ERPNext 的配置导出/导入功能

**实施建议**：
```go
// 字段存在性检查
func checkCustomFieldExists(ctx context.Context, doctype, fieldname string) bool {
    // 调用 ERPNext API 检查字段是否存在
    // 如果不存在，记录警告日志
    return true
}
```

### 3. 数据迁移风险

**风险描述**：
- 现有 Material Request 可能已经填写了 `source_warehouse`
- 需要将现有数据迁移到新的 `warehouse_group` 字段

**影响程度**：**中**

**缓解措施**：
1. **数据迁移脚本**
   - 编写数据迁移脚本，将现有 `source_warehouse` 转换为 `warehouse_group`
   - 如果 `source_warehouse` 是子仓库，查找其父仓库组
   - 如果找不到父仓库组，使用默认母仓库

2. **分批迁移**
   - 先迁移已完成的历史数据
   - 再迁移进行中的数据
   - 最后处理新创建的数据

3. **回滚方案**
   - 保留原始数据，不删除 `source_warehouse` 字段
   - 如果迁移失败，可以回滚

**实施建议**：
```python
# ERPNext 数据迁移脚本
def migrate_material_request_warehouse():
    """
    将 Material Request 的 source_warehouse 迁移到 warehouse_group
    """
    mrs = frappe.get_all("Material Request", filters={"warehouse_group": ""}, fields=["name", "source_warehouse"])
    
    for mr in mrs:
        if mr.source_warehouse:
            # 查找仓库组
            warehouse = frappe.get_doc("Warehouse", mr.source_warehouse)
            
            if warehouse.is_group:
                # 如果本身就是仓库组，直接使用
                warehouse_group = mr.source_warehouse
            elif warehouse.parent_warehouse:
                # 如果有父仓库，检查是否为仓库组
                parent = frappe.get_doc("Warehouse", warehouse.parent_warehouse)
                if parent.is_group:
                    warehouse_group = warehouse.parent_warehouse
                else:
                    # 使用默认母仓库
                    warehouse_group = get_default_warehouse_group()
            else:
                # 使用默认母仓库
                warehouse_group = get_default_warehouse_group()
            
            # 更新 Material Request
            frappe.db.set_value("Material Request", mr.name, "warehouse_group", warehouse_group)
    
    frappe.db.commit()
```

### 4. 业务逻辑风险

**风险描述**：
- 如果门店未选择仓库组，系统默认使用母仓库
- 如果默认母仓库配置错误，可能导致业务异常

**影响程度**：**中**

**缓解措施**：
1. **默认仓库组配置**
   - 在系统配置中明确设置默认母仓库
   - 提供配置界面，方便管理员修改

2. **配置验证**
   - 在系统启动时验证默认母仓库配置
   - 如果配置不存在或无效，记录错误日志并阻止系统启动

3. **降级方案**
   - 如果找不到默认母仓库，使用第一个可用的母仓库
   - 如果都没有，使用默认仓库（兼容旧逻辑）

**实施建议**：
```go
// 获取默认仓库组
func (s *sWarehouse) GetDefaultWarehouseGroup(ctx context.Context, companyName string) (*erp.Warehouse, error) {
    // 1. 从系统配置获取默认母仓库
    defaultWarehouseGroup, err := s.getConfigDefaultWarehouseGroup(ctx, companyName)
    if err == nil && defaultWarehouseGroup != nil {
        return defaultWarehouseGroup, nil
    }
    
    // 2. 查找公司的第一个母仓库
    warehouseGroups, err := s.getWarehouseGroups(ctx, companyName)
    if err == nil && len(warehouseGroups) > 0 {
        return warehouseGroups[0], nil
    }
    
    // 3. 降级：使用默认仓库
    defaultWarehouse, err := s.GetDefaultWarehouse(ctx, companyName, "")
    if err != nil {
        return nil, gerror.Wrapf(err, "获取默认仓库组失败")
    }
    
    // 如果默认仓库不是仓库组，查找其父仓库组
    if !defaultWarehouse.IsGroup && defaultWarehouse.ParentWarehouse != "" {
        parentWarehouse, err := s.GetWarehouse(ctx, defaultWarehouse.ParentWarehouse)
        if err == nil && parentWarehouse.IsGroup {
            return parentWarehouse, nil
        }
    }
    
    return nil, gerror.New("未找到默认仓库组")
}
```

### 5. 性能风险

**风险描述**：
- 在创建 Sales Order 时，需要查询 Material Request 的 `warehouse_group`
- 如果查询频繁，可能影响性能

**影响程度**：**低**

**缓解措施**：
1. **缓存机制**
   - 缓存 Material Request 的 `warehouse_group` 信息
   - 减少数据库查询次数

2. **批量查询**
   - 如果同时处理多个 Material Request，使用批量查询

3. **索引优化**
   - 在 ERPNext 的 Material Request 表上为 `warehouse_group` 字段添加索引

**实施建议**：
```go
// 使用缓存
var materialRequestCache = sync.Map{}

func getMaterialRequestWarehouseGroup(ctx context.Context, mrName string) (string, error) {
    // 先从缓存获取
    if cached, ok := materialRequestCache.Load(mrName); ok {
        return cached.(string), nil
    }
    
    // 查询数据库
    mr, err := service.Document().Get(ctx, erp.DocTypeMaterialRequest, mrName)
    if err != nil {
        return "", err
    }
    
    warehouseGroup := mr.Get("warehouse_group").String()
    
    // 存入缓存
    materialRequestCache.Store(mrName, warehouseGroup)
    
    return warehouseGroup, nil
}
```

---

## 📅 实施计划

### 阶段一：准备阶段（1-2天）

1. **需求确认**
   - 与业务方确认需求细节
   - 确认默认母仓库配置

2. **技术方案评审**
   - 技术方案评审
   - 风险评估和缓解措施确认

3. **ERPNext 配置准备**
   - 准备 ERPNext 自定义字段配置
   - 准备配置脚本

### 阶段二：开发阶段（3-5天）

1. **后端开发**
   - 修改 Protobuf 定义
   - 调整 Material Request 创建逻辑
   - 调整 Sales Order 创建逻辑
   - 实现兼容性处理
   - 添加默认仓库组获取逻辑

2. **ERPNext 配置**
   - 在 ERPNext 中添加自定义字段
   - 配置字段过滤条件
   - 测试字段功能

3. **数据迁移脚本**
   - 编写数据迁移脚本
   - 测试迁移脚本

### 阶段三：测试阶段（2-3天）

1. **单元测试**
   - 测试 Material Request 创建逻辑
   - 测试 Sales Order 创建逻辑
   - 测试兼容性处理

2. **集成测试**
   - 测试完整流程：门店创建 MR → 总部创建 SO → 总部分配子仓库
   - 测试默认值处理
   - 测试数据迁移

3. **性能测试**
   - 测试缓存机制
   - 测试批量查询性能

### 阶段四：部署阶段（1-2天）

1. **ERPNext 配置部署**
   - 在测试环境部署 ERPNext 配置
   - 验证配置正确性

2. **代码部署**
   - 部署后端代码
   - 验证功能正常

3. **数据迁移**
   - 执行数据迁移脚本
   - 验证迁移结果

4. **前端更新**
   - 更新前端代码
   - 验证前端功能

### 阶段五：验证阶段（1-2天）

1. **功能验证**
   - 验证门店创建 MR 流程
   - 验证总部创建 SO 流程
   - 验证总部分配子仓库流程

2. **兼容性验证**
   - 验证新旧版本兼容性
   - 验证数据迁移结果

3. **性能验证**
   - 验证性能指标
   - 验证缓存效果

---

## ✅ 验收标准

### 功能验收

1. **门店操作**
   - ✅ 门店创建 Material Request 时，**不需要选择仓库**，系统自动使用默认母仓库
   - ✅ 门店界面不显示仓库选择字段，操作更简单
   - ✅ Material Request 可以正常保存和提交

2. **总部操作**
   - ✅ 从 Material Request 创建 Sales Order 时，自动继承仓库组信息到自定义字段 `warehouse_group`
   - ✅ Sales Order 的 `SetWarehouse` 字段必须设置为具体的子仓库（不能是母仓库）
   - ✅ 总部可以在 Sales Order 中选择具体的子仓库
   - ✅ 总部可以手动选择或自动分配子仓库
   - ✅ 验证：不能将母仓库设置为 `SetWarehouse`，系统会报错

3. **系统自动处理**
   - ✅ 系统自动使用默认母仓库，无需门店选择
   - ✅ 默认母仓库配置正确，系统能正确获取

4. **兼容性**
   - ✅ 旧版本前端传递 `source_warehouse` 字段时，系统自动忽略，使用默认母仓库
   - ✅ 新版本前端不传递仓库字段，系统自动使用默认母仓库

### 性能验收

1. **响应时间**
   - Material Request 创建响应时间 < 1秒
   - Sales Order 创建响应时间 < 2秒

2. **并发性能**
   - 支持 100+ 并发创建 Material Request
   - 支持 50+ 并发创建 Sales Order

### 数据验收

1. **数据完整性**
   - 所有 Material Request 都有 `warehouse_group` 字段（存储母仓库标识）
   - 所有 Sales Order 都正确继承 `warehouse_group`（存储母仓库标识）
   - 所有 Sales Order 的 `SetWarehouse` 字段都是具体的子仓库（不是母仓库）
   - 所有 Sales Order Item 的 `warehouse` 字段都是具体的子仓库（不是母仓库）

2. **数据迁移**
   - 历史数据迁移成功率 > 99%
   - 迁移后数据正确性验证通过

---

## 📝 注意事项

### 1. ERPNext 限制

**⚠️ 重要：母仓库不能直接用于库存记账和收发**

1. **仓库组（母仓库）的技术限制**
   - 母仓库（`is_group = 1`）只是逻辑分组标识，**不存放实物**
   - ERPNext 的 `Bin` 表（库存台账）只能关联真实仓库（`is_group = 0`）
   - **不能直接使用母仓库进行库存记账、收发操作**

2. **Material Request 和 Sales Order 的 warehouse 字段限制**
   - Material Request 和 Sales Order 的 `warehouse` 字段**不能选择仓库组**
   - 必须选择具体的子仓库（`is_group = 0`）才能进行库存操作
   - 需要通过自定义字段 `warehouse_group` 存储母仓库信息（仅作为标识）

3. **Sales Order 创建时的处理**
   - Sales Order 的 `SetWarehouse` 字段必须设置为具体的子仓库
   - 不能设置为母仓库，否则会导致库存操作失败
   - 如果母仓库下只有一个子仓库，可以自动设置为该子仓库
   - 如果母仓库下有多个子仓库，必须由总部分配具体的子仓库

4. **Delivery Note 仓库继承**
   - Delivery Note 的 `warehouse` 字段会从 Sales Order Item 的 `warehouse` 字段继承
   - 必须确保 Sales Order Item 的 `warehouse` 字段已设置为具体的子仓库
   - 不能继承母仓库，否则会导致发货失败

5. **自定义字段 `warehouse_group` 设置验证**
   - ✅ **可以设置**：ERPNext 的自定义字段 `Link → Warehouse` 类型**可以**存储仓库组（`is_group = 1`）
   - ✅ **API 支持**：通过 API 创建 Material Request 时，可以成功设置 `warehouse_group` 为仓库组名称
   - ⚠️ **必须验证**：在设置 `warehouse_group` 之前，必须验证该仓库确实是仓库组（`is_group = 1`）
   - ⚠️ **字段配置**：自定义字段 `warehouse_group` 必须配置过滤条件 `["is_group", "=", 1]`，确保只能选择仓库组
   - ⚠️ **用途限制**：`warehouse_group` 仅作为标识，不能用于实际的库存操作（如 `SetWarehouse`、`warehouse` 字段）

### 2. 配置要求

- 必须在 ERPNext 中正确配置仓库层级结构
- 必须正确配置自定义字段
- 必须正确配置默认母仓库

### 3. 版本兼容

- 在过渡期内，需要同时支持新旧字段
- 前端需要逐步迁移到新字段
- 后端需要保持向后兼容

---

## 🔄 库存同步性与数据一致性分析

### 1. 库存同步时机

#### 1.1 品牌采购流程中的库存流转

**品牌采购分为两部分：集采和直采，在途仓库的使用时机不同**

##### 1.1.1 集采部分（总部发货给门店）

**完整流程**：

```
1. 门店创建 Material Request
   - TTPOS：创建采购申请
   - ERPNext：创建 Material Request（warehouse_group = 母仓库）
   - ⚠️ 此时：库存无变化
   
2. MR 审批后，创建 Sales Order
   - ERPNext：创建 Sales Order（warehouse_group = 母仓库，SetWarehouse = 子仓库）
   - ⚠️ 此时：库存无变化
   
3. 总部创建 Delivery Note 并提交
   - ERPNext：从总部子仓库扣减库存 ✅
   - TTPOS：需要同步库存变化 ⚠️
   - ⚠️ 关键：TTPOS 如何知道 Delivery Note 已提交？
   
4. 物品进入在途仓库（TTPOS）
   - TTPOS：库存进入门店的在途仓库 ✅
   - ⚠️ 关键：何时触发？Delivery Note 提交时？还是门店确认收货时？
   
5. 门店确认收货
   - TTPOS：从在途仓库转入目标仓库 ✅
   - ERPNext：库存已在 Delivery Note 提交时扣减 ✅
```

##### 1.1.2 直采部分（外部供应商发货给门店）

**完整流程**：

```
1. 门店创建 Material Request
   - TTPOS：创建采购申请
   - ERPNext：创建 Material Request
   - ⚠️ 此时：库存无变化
   
2. MR 审批后，创建 Purchase Order
   - ERPNext：创建 Purchase Order（外部供应商）
   - ⚠️ 此时：库存无变化
   - ⚠️ 当前问题：代码在此时就添加到在途仓库（过早）
   
3. 外部供应商发货，Purchase Receipt 提交
   - ERPNext：Purchase Receipt 提交，库存增加（门店仓库）✅
   - TTPOS：需要同步库存变化 ⚠️
   - ⚠️ 关键：TTPOS 如何知道 Purchase Receipt 已提交？
   
4. 物品进入在途仓库（TTPOS）
   - TTPOS：库存进入门店的在途仓库 ✅
   - ⚠️ 关键：应该在 Purchase Receipt 提交时触发，而不是 Purchase Order 创建时
   
5. 门店确认收货
   - TTPOS：从在途仓库转入目标仓库 ✅
   - ERPNext：库存已在 Purchase Receipt 提交时增加 ✅
```

**直采部分在途仓库使用时机对比**：

| 时机 | 当前实现 | 应该实现 | 说明 |
|------|---------|---------|------|
| **Purchase Order 创建** | ✅ 添加到在途仓库 | ✅ 添加到在途仓库 | 表示"预期在途"，供应商发货前就记录预期库存 |
| **外部供应商发货** | ❌ 无法知道 | ❌ 无法知道 | 外部供应商发货是线下操作，不在 ERPNext 中记录 |
| **门店创建收货单** | ❌ 未处理 | ✅ 可选：添加到在途仓库 | 如果货物还在运输途中，可以添加到在途仓库 |
| **门店确认收货** | ✅ 从在途仓库转入目标仓库 | ✅ 从在途仓库转入目标仓库 | 收货时从在途仓库扣减，目标仓库增加 |

**⚠️ 关键问题**：

1. **外部供应商发货时间未知**：
   - 外部供应商发货是线下操作，不会在 ERPNext 中记录
   - 无法知道具体发货时间
   - 当门店在 TTPOS 中创建 Purchase Receipt 时，货物可能已经到达门店

2. **在途仓库的意义**：
   - 如果货物已经到达门店，在 Purchase Receipt 提交时添加到在途仓库已经没有意义
   - **建议**：保持当前实现，在创建 Purchase Order 时就添加到在途仓库，表示"预期在途"
   - 这样可以在供应商发货前就记录预期库存，便于库存管理

#### 1.2 在途仓库使用时机

**当前实现分析**（`main/app/service/purchase_order/helper.go`）：

```go
// 在 handleInternalPurchaseErp 中
// 1. 减少总部库存
err := s.helper.reduceHeadquarterStockAndLog(ctx, subDb, tx, purchaseOrder)

// 2. 添加到门店在途仓库
if transitWarehouse != nil {
    err := s.helper.AddToTransitWarehouse(tx, transitWarehouse, purchaseOrder, supplierUuid, &item, actualNum)
}
```

**问题分析**：

1. **集采部分时机问题**：
   - 当前代码在创建 Material Request 时就添加到在途仓库
   - 但此时 Delivery Note 还未创建，总部还未发货
   - **应该**：在 Delivery Note 提交后，才添加到在途仓库

2. **直采部分时机问题**：
   - 当前代码在创建 Purchase Order 时就添加到在途仓库（`handleExternalPurchaseErp`）
   - ⚠️ **问题**：外部供应商发货是线下操作，无法知道具体发货时间
   - ⚠️ **问题**：当门店创建 Purchase Receipt 时，货物可能已经到达门店
   - **解决方案**：
     - **方案 A（推荐）**：保持当前实现，在创建 Purchase Order 时就添加到在途仓库，表示"预期在途"
     - **方案 B**：不使用在途仓库，直采部分直接从供应商到门店仓库（简化流程）
     - **方案 C**：在门店创建收货单时添加到在途仓库（如果货物还在运输途中）

3. **同步问题**：
   - Delivery Note 提交后，ERPNext 从总部仓库扣减库存
   - Purchase Receipt 提交后，ERPNext 增加门店仓库库存
   - TTPOS 需要同步这些变化
   - **需要**：监听 Delivery Note 和 Purchase Receipt 提交事件，或通过轮询同步

#### 1.3 正确的库存流转时机

##### 1.3.1 集采部分（总部发货给门店）

**推荐流程**：

```
1. 门店创建 Material Request
   - TTPOS：创建采购申请
   - ERPNext：创建 Material Request
   - 库存：无变化 ✅

2. MR 审批后，创建 Sales Order
   - ERPNext：创建 Sales Order
   - 库存：无变化 ✅

3. 总部创建 Delivery Note 并提交
   - ERPNext：从总部子仓库扣减库存 ✅
   - TTPOS：同步库存变化（通过 Webhook 或轮询）✅
   - TTPOS：库存进入门店在途仓库 ✅
   - ⚠️ 关键：此时物品在运输途中

4. 门店确认收货
   - TTPOS：从在途仓库转入目标仓库 ✅
   - ERPNext：库存已在 Delivery Note 提交时扣减 ✅
```

##### 1.3.2 直采部分（外部供应商发货给门店）

**推荐流程（方案 A：保持当前实现）**：

```
1. 门店创建 Material Request
   - TTPOS：创建采购申请
   - ERPNext：创建 Material Request
   - 库存：无变化 ✅

2. MR 审批后，创建 Purchase Order
   - ERPNext：创建 Purchase Order（外部供应商）
   - TTPOS：库存进入门店在途仓库（预期在途）✅
   - ⚠️ 说明：此时供应商还未发货，但记录"预期在途"库存，便于库存管理

3. 外部供应商发货（线下操作，不在 ERPNext 中记录）
   - ⚠️ 关键：无法知道具体发货时间
   - ⚠️ 关键：货物可能在运输途中，也可能已经到达门店

4. 门店创建 Purchase Receipt（收货单）
   - TTPOS：门店创建收货单，记录实际收货数量
   - ⚠️ 关键：此时货物可能已经到达门店，也可能还在运输途中
   - ⚠️ 关键：实际收货数量可能与 Purchase Order 数量不一致（供应商可能修改数量）

5. 门店确认收货
   - TTPOS：从在途仓库转入目标仓库 ✅
   - ERPNext：Purchase Receipt 提交，库存增加（门店仓库）✅
   - ⚠️ 关键：使用实际收货数量，而不是 Purchase Order 数量
```

**替代方案（方案 B：不使用在途仓库）**：

```
1. 门店创建 Material Request
   - TTPOS：创建采购申请
   - ERPNext：创建 Material Request
   - 库存：无变化 ✅

2. MR 审批后，创建 Purchase Order
   - ERPNext：创建 Purchase Order（外部供应商）
   - TTPOS：不添加到在途仓库 ✅
   - ⚠️ 说明：简化流程，直采部分不使用在途仓库

3. 外部供应商发货（线下操作，不在 ERPNext 中记录）
   - ⚠️ 关键：无法知道具体发货时间

4. 门店创建 Purchase Receipt（收货单）
   - TTPOS：门店创建收货单，记录实际收货数量
   - ⚠️ 关键：实际收货数量可能与 Purchase Order 数量不一致

5. 门店确认收货
   - TTPOS：直接增加到目标仓库 ✅
   - ERPNext：Purchase Receipt 提交，库存增加（门店仓库）✅
   - ⚠️ 关键：使用实际收货数量，而不是 Purchase Order 数量
```

**直采部分代码修正建议**：

```go
// 方案 A（推荐）：保持当前实现，在创建 Purchase Order 时就添加到在途仓库
// 优点：表示"预期在途"，便于库存管理
// 缺点：如果供应商未发货，在途库存会一直存在
func (s *purchaseOrderSrv) handleExternalPurchaseErp(...) {
    // ... 创建 Purchase Order ...
    
    // ✅ 保持：在创建 Purchase Order 时就添加到在途仓库（预期在途）
    if transitWarehouse != nil {
        err := s.helper.AddToTransitWarehouse(tx, transitWarehouse, purchaseOrder, supplierUuid, &item, actualNum)
    }
}

// 方案 B：不使用在途仓库，直采部分简化流程
func (s *purchaseOrderSrv) handleExternalPurchaseErp(...) {
    // ... 创建 Purchase Order ...
    
    // ❌ 移除：直采部分不使用在途仓库
    // if transitWarehouse != nil {
    //     err := s.helper.AddToTransitWarehouse(...)
    // }
}

// 门店确认收货时，处理数量差异
func (s *purchaseReceiptOrderSrv) updateMaterialStock(...) {
    // 1. 从在途仓库转入目标仓库（方案 A）
    //    或直接增加到目标仓库（方案 B）
    // 2. 使用实际收货数量，而不是 Purchase Order 数量
    // 3. 处理数量差异（多收或少收）
    for _, item := range receiptOrder.Items {
        actualNum := item.GetUnitsTotalConversionRateNum()  // 实际收货数量
        // ... 更新库存 ...
    }
}
```

### 2. 数据一致性保障

#### 2.1 ERPNext 与 TTPOS 库存同步机制

**当前同步方式**：

1. **主动同步**（TTPOS → ERPNext）
   - TTPOS 创建单据时，主动调用 ERPNext API
   - 例如：创建 Material Request、创建 Purchase Order

2. **被动同步**（ERPNext → TTPOS）
   - **问题**：Delivery Note 提交后，TTPOS 如何知道？
   - **方案 A**：ERPNext Webhook（推荐）
     - ERPNext 提交 Delivery Note 时，触发 Webhook
     - TTPOS 接收 Webhook，同步库存变化
   - **方案 B**：定时轮询
     - TTPOS 定时查询 ERPNext 的 Delivery Note 状态
     - 发现已提交的 Delivery Note，同步库存变化

#### 2.2 库存同步实现建议

**方案 A：ERPNext Webhook（推荐）**

**需要配置的 Webhook 列表**：

| Webhook 事件 | 单据类型 | 业务场景 | 库存影响 | 优先级 |
|-------------|---------|---------|---------|--------|
| **Delivery Note 提交** | Delivery Note | 集采发货 | 总部仓库扣减 → 门店在途仓库增加 | ⚠️ **高** |
| **Purchase Receipt 提交** | Purchase Receipt | 直采收货 | 门店在途仓库增加（如果使用在途仓库） | ⚠️ **高** |
| **Sales Invoice (Credit Note) 提交** | Sales Invoice | 集采退货 | 门店仓库扣减 → 总部仓库增加 | ⚠️ **高** |
| **Purchase Invoice (Debit Note) 提交** | Purchase Invoice | 直采退货 | 门店仓库扣减 | ⚠️ **高** |
| **换货 Delivery Note 提交** | Delivery Note | 集采换货 | 总部仓库扣减 → 门店在途仓库增加 | ⚠️ **中** |
| **Sales Order 取消** | Sales Order | 集采订单取消 | 需要回滚库存 | ⚠️ **中** |
| **Purchase Order 取消** | Purchase Order | 直采订单取消 | 需要清理在途库存 | ⚠️ **中** |

**Webhook 配置**：

```go
// ERPNext 配置 Webhook
// 1. Delivery Note 提交后（集采发货）
//    Webhook URL：https://ttpos-api.com/api/erpnext/webhook/delivery-note-submitted
//    Condition: docstatus = 1 AND is_return = 0
//
// 2. Purchase Receipt 提交后（直采收货）
//    Webhook URL：https://ttpos-api.com/api/erpnext/webhook/purchase-receipt-submitted
//    Condition: docstatus = 1 AND is_return = 0
//
// 3. Sales Invoice (Credit Note) 提交后（集采退货）
//    Webhook URL：https://ttpos-api.com/api/erpnext/webhook/sales-invoice-credit-note-submitted
//    Condition: docstatus = 1 AND is_return = 1 AND is_credit_note = 1
//
// 4. Purchase Invoice (Debit Note) 提交后（直采退货）
//    Webhook URL：https://ttpos-api.com/api/erpnext/webhook/purchase-invoice-debit-note-submitted
//    Condition: docstatus = 1 AND is_return = 1 AND is_debit_note = 1
//
// 5. Delivery Note (换货) 提交后（集采换货）
//    Webhook URL：https://ttpos-api.com/api/erpnext/webhook/delivery-note-exchange-submitted
//    Condition: docstatus = 1 AND is_return = 0 AND 关联原始 Sales Order
//
// 6. Sales Order 取消后（集采订单取消）
//    Webhook URL：https://ttpos-api.com/api/erpnext/webhook/sales-order-cancelled
//    Condition: docstatus = 2
//
// 7. Purchase Order 取消后（直采订单取消）
//    Webhook URL：https://ttpos-api.com/api/erpnext/webhook/purchase-order-cancelled
//    Condition: docstatus = 2
```

**TTPOS 处理 Webhook 实现**：

```go
// 1. 集采发货：Delivery Note 提交后
func HandleDeliveryNoteSubmitted(ctx context.Context, req *DeliveryNoteWebhookReq) error {
    // 1. 验证 Webhook 签名
    // 2. 获取 Delivery Note 信息
    // 3. 获取关联的 Sales Order
    // 4. 同步库存变化：
    //    - 减少总部仓库库存（如果 TTPOS 维护总部库存）
    //    - 增加门店在途仓库库存（集采部分）
    // 5. 记录同步日志
}

// 2. 直采收货：Purchase Receipt 提交后
func HandlePurchaseReceiptSubmitted(ctx context.Context, req *PurchaseReceiptWebhookReq) error {
    // 1. 验证 Webhook 签名
    // 2. 获取 Purchase Receipt 信息
    // 3. 获取关联的 Purchase Order
    // 4. 同步库存变化：
    //    - 增加门店在途仓库库存（直采部分，如果使用在途仓库）
    //    - ERPNext 已增加门店仓库库存
    // 5. 记录同步日志
}

// 3. 集采退货：Sales Invoice (Credit Note) 提交后
func HandleSalesInvoiceCreditNoteSubmitted(ctx context.Context, req *SalesInvoiceWebhookReq) error {
    // 1. 验证 Webhook 签名
    // 2. 获取 Sales Invoice (Credit Note) 信息
    // 3. 获取关联的 Delivery Note
    // 4. 同步库存变化：
    //    - 减少门店仓库库存
    //    - 增加总部仓库库存（如果 TTPOS 维护总部库存）
    // 5. 记录同步日志
}

// 4. 直采退货：Purchase Invoice (Debit Note) 提交后
func HandlePurchaseInvoiceDebitNoteSubmitted(ctx context.Context, req *PurchaseInvoiceWebhookReq) error {
    // 1. 验证 Webhook 签名
    // 2. 获取 Purchase Invoice (Debit Note) 信息
    // 3. 获取关联的 Purchase Receipt
    // 4. 同步库存变化：
    //    - 减少门店仓库库存
    //    - 可能需要处理在途仓库库存（如果之前添加到在途仓库）
    // 5. 记录同步日志
}

// 5. 集采换货：换货 Delivery Note 提交后
func HandleDeliveryNoteExchangeSubmitted(ctx context.Context, req *DeliveryNoteWebhookReq) error {
    // 1. 验证 Webhook 签名
    // 2. 获取换货 Delivery Note 信息
    // 3. 获取关联的原始 Sales Order
    // 4. 同步库存变化：
    //    - 减少总部仓库库存（如果 TTPOS 维护总部库存）
    //    - 增加门店在途仓库库存（换货部分）
    // 5. 记录同步日志
}

// 6. 集采订单取消：Sales Order 取消后
func HandleSalesOrderCancelled(ctx context.Context, req *SalesOrderWebhookReq) error {
    // 1. 验证 Webhook 签名
    // 2. 获取 Sales Order 信息
    // 3. 检查是否有已提交的 Delivery Note
    // 4. 如果有，可能需要回滚库存变化
    // 5. 记录同步日志
}

// 7. 直采订单取消：Purchase Order 取消后
func HandlePurchaseOrderCancelled(ctx context.Context, req *PurchaseOrderWebhookReq) error {
    // 1. 验证 Webhook 签名
    // 2. 获取 Purchase Order 信息
    // 3. 清理在途仓库库存（如果之前添加到在途仓库）
    // 4. 记录同步日志
}
```

**方案 B：定时轮询**

```go
// 定时任务：同步 Delivery Note 和 Purchase Receipt 状态
func SyncErpDocumentStatus(ctx context.Context) error {
    // 1. 同步 Delivery Note（集采部分）
    //    - 查询 ERPNext 中已提交但未同步的 Delivery Note
    //    - 对于每个 Delivery Note：
    //      - 同步库存变化（减少总部库存，增加门店在途库存）
    //      - 标记为已同步
    
    // 2. 同步 Purchase Receipt（直采部分）
    //    - 查询 ERPNext 中已提交但未同步的 Purchase Receipt
    //    - 对于每个 Purchase Receipt：
    //      - 同步库存变化（增加门店在途库存）
    //      - 标记为已同步
    
    // 3. 记录同步日志
}
```

#### 2.3 在途仓库使用时机修正

##### 2.3.1 集采部分：Delivery Note 提交后添加到在途仓库

**修正后的实现**：

```go
// 在 Delivery Note 提交后（通过 Webhook 或轮询触发）
func HandleDeliveryNoteSubmitted(ctx context.Context, deliveryNoteName string) error {
    // 1. 获取 Delivery Note 信息
    deliveryNote, err := erpService.GetDeliveryNote(ctx, deliveryNoteName)
    if err != nil {
        return err
    }
    
    // 2. 获取关联的 Sales Order
    salesOrder, err := erpService.GetSalesOrder(ctx, deliveryNote.AgainstSalesOrder)
    if err != nil {
        return err
    }
    
    // 3. 获取门店信息（从 Sales Order 的 Customer 获取）
    storeCompany, err := getStoreCompanyFromCustomer(ctx, salesOrder.Customer)
    if err != nil {
        return err
    }
    
    // 4. 获取门店在途仓库
    transitWarehouse, err := getTransitWarehouse(ctx, storeCompany.Uuid)
    if err != nil {
        return err
    }
    
    // 5. 为每个物品添加到在途仓库（集采部分）
    for _, item := range deliveryNote.Items {
        // 添加到门店在途仓库
        err := addToTransitWarehouse(ctx, transitWarehouse, item, "集采")
        if err != nil {
            return err
        }
    }
    
    return nil
}
```

##### 2.3.2 直采部分：Purchase Receipt 提交后添加到在途仓库

**修正后的实现**：

```go
// 在 Purchase Receipt 提交后（通过 Webhook 或轮询触发）
func HandlePurchaseReceiptSubmitted(ctx context.Context, purchaseReceiptName string) error {
    // 1. 获取 Purchase Receipt 信息
    purchaseReceipt, err := erpService.GetPurchaseReceipt(ctx, purchaseReceiptName)
    if err != nil {
        return err
    }
    
    // 2. 获取关联的 Purchase Order
    purchaseOrder, err := erpService.GetPurchaseOrder(ctx, purchaseReceipt.PurchaseOrder)
    if err != nil {
        return err
    }
    
    // 3. 获取门店信息（从 Purchase Order 的 Company 获取）
    storeCompany, err := getStoreCompanyFromErpCode(ctx, purchaseOrder.Company)
    if err != nil {
        return err
    }
    
    // 4. 获取门店在途仓库
    transitWarehouse, err := getTransitWarehouse(ctx, storeCompany.Uuid)
    if err != nil {
        return err
    }
    
    // 5. 为每个物品添加到在途仓库（直采部分）
    for _, item := range purchaseReceipt.Items {
        // 添加到门店在途仓库
        err := addToTransitWarehouse(ctx, transitWarehouse, item, "直采")
        if err != nil {
            return err
        }
    }
    
    return nil
}
```

**关键修正点**：

1. **方案 A（推荐）：保持 Purchase Order 创建时的在途仓库添加**
   - 在 `handleExternalPurchaseErp` 中，保持 `AddToTransitWarehouse` 调用
   - 表示"预期在途"，便于库存管理
   - ⚠️ 注意：如果供应商未发货，在途库存会一直存在

2. **方案 B：不使用在途仓库**
   - 在 `handleExternalPurchaseErp` 中，移除 `AddToTransitWarehouse` 调用
   - 简化流程，直采部分不使用在途仓库
   - 门店收货时，直接增加到目标仓库

3. **数量差异处理**：
   - 使用实际收货数量，而不是 Purchase Order 数量
   - 从在途仓库扣减时，使用实际收货数量
   - 如果实际收货数量 > Purchase Order 数量，需要处理在途仓库库存不足的情况

### 3. 数据一致性检查点

#### 3.1 关键检查点

| 检查点 | ERPNext 状态 | TTPOS 状态 | 一致性要求 |
|--------|-------------|-----------|-----------|
| **Material Request 创建** | Material Request 已创建 | 采购申请已创建 | ✅ 单据一致 |
| **Sales Order 创建（集采）** | Sales Order 已创建，warehouse_group 已设置 | 集采订单已创建 | ✅ 单据一致 |
| **Purchase Order 创建（直采）** | Purchase Order 已创建 | 直采订单已创建，在途仓库库存已增加（方案 A） | ✅ 单据一致 |
| **外部供应商发货（直采）** | ❌ 无法知道 | ❌ 无法知道 | ⚠️ **线下操作，无法同步** |
| **门店创建收货单（直采）** | ❌ 未创建 | 收货单已创建，记录实际收货数量 | ⚠️ **数量可能不一致** |
| **Delivery Note 提交（集采）** | 总部仓库库存已扣减 | 门店在途仓库库存应增加 | ⚠️ **需要同步** |
| **Purchase Receipt 提交（直采）** | 门店仓库库存已增加（使用实际收货数量） | 在途仓库库存转入目标仓库（使用实际收货数量） | ✅ 库存一致 |
| **门店确认收货（集采）** | 库存已在 Delivery Note 提交时扣减 | 在途仓库库存转入目标仓库 | ✅ 库存一致 |
| **门店确认收货（直采）** | 库存已在 Purchase Receipt 提交时增加（实际收货数量） | 在途仓库库存转入目标仓库（实际收货数量） | ✅ 库存一致 |

#### 3.2 数量差异处理

**问题**：外部供应商可能会修改 Purchase Order 的数量，导致实际到货数量与采购申请数量不一致。

**当前实现**（`main/app/service/purchase_order/receipt_order.go`）：

```go
// 门店创建收货单时，记录实际收货数量
func (s *purchaseReceiptOrderSrv) CreatePurchaseReceiptOrder(...) {
    // 1. 验证收货数量（允许与采购订单数量不一致）
    reqNum := s.validator.validateReceiptQuantityNew(ctx, orderItem, itemReq.UnitList)
    
    // 2. 更新采购申请明细的到货数量
    newArrivalNum := orderItem.ArrivalNum + reqNum
    
    // 3. 创建收货明细（使用实际收货数量）
    receiptItem := &model.PurchaseReceiptOrderItem{
        Num: reqNum,  // 实际收货数量
        // ...
    }
}

// 门店确认收货时，使用实际收货数量更新库存
func (s *purchaseReceiptOrderSrv) updateMaterialStock(...) {
    for _, item := range receiptOrder.Items {
        actualNum := item.GetUnitsTotalConversionRateNum()  // 实际收货数量
        // 从在途仓库扣减（使用实际收货数量）
        // 增加到目标仓库（使用实际收货数量）
    }
}
```

**处理逻辑**：

1. **允许数量差异**：
   - 收货数量可以大于或小于采购订单数量
   - 系统记录实际收货数量，而不是采购订单数量

2. **更新到货数量**：
   - 每次收货时，更新采购申请明细的 `ArrivalNum`（到货数量）
   - 支持多次收货，累计到货数量

3. **库存更新**：
   - 从在途仓库扣减时，使用实际收货数量
   - 增加到目标仓库时，使用实际收货数量
   - 确保库存数量与实际收货数量一致

4. **ERPNext 同步**：
   - Purchase Receipt 提交时，使用实际收货数量
   - 确保 ERPNext 和 TTPOS 的库存数量一致

**数量差异场景**：

| 场景 | 处理方式 | 说明 |
|------|---------|------|
| **实际收货数量 = 采购订单数量** | ✅ 正常处理 | 数量一致，直接更新库存 |
| **实际收货数量 < 采购订单数量** | ✅ 正常处理 | 少收部分，在途仓库会有剩余库存，需要后续处理 |
| **实际收货数量 > 采购订单数量** | ✅ 正常处理 | 多收部分，在途仓库库存不足，需要从预期库存中补充 |
| **多次收货** | ✅ 支持 | 每次收货时累计到货数量，直到全部收货完成 |

#### 3.2.1 场景1：实际收货数量 < 采购订单数量（少收）

**问题描述**：
- 采购订单数量：100
- 实际收货数量：80
- 在途仓库库存：100（创建 Purchase Order 时添加）
- 少收数量：20

**当前实现**（`main/app/service/purchase_order/receipt_order.go`）：

```go
// 从在途仓库扣减实际收货数量
actualNum := item.GetUnitsTotalConversionRateNum()  // 80
warehouseItem, err := warehouseItemRepo.GetByWarehouseAndMaterial(transitWarehouse.Uuid, item.MaterialUuid)
if warehouseItem.Stock < actualNum {
    actualNum = warehouseItem.Stock  // 如果库存不足，限制为库存数量
}
// 减少在途仓库库存：100 - 80 = 20（剩余库存）
err = warehouseItemRepo.ReduceStock(warehouseItem.Uuid, actualNum)
```

**处理结果**：
- ✅ 在途仓库剩余库存：20
- ✅ 目标仓库增加库存：80
- ⚠️ **问题**：在途仓库有剩余库存，需要后续处理

**后续处理方案**：

**方案 A：等待后续收货**
- 如果供应商后续补发，可以再次创建收货单
- 使用剩余的在途仓库库存（20）
- 直到全部收货完成

**方案 B：清理剩余库存**
- 如果供应商确认不再发货，需要清理剩余库存
- 提供"取消剩余收货"功能，将剩余在途库存清零
- 更新采购订单状态为"部分收货"或"收货完成"

**方案 C：自动清理（推荐）**
- 设置超时时间（如 30 天）
- 如果超过超时时间仍未收货，自动清理剩余在途库存
- 记录清理日志，更新采购订单状态

**代码实现建议**：

```go
// 处理少收场景
func (s *purchaseReceiptOrderSrv) handlePartialReceipt(ctx context.Context, receiptOrder *model.PurchaseReceiptOrder) error {
    for _, item := range receiptOrder.Items {
        actualNum := item.GetUnitsTotalConversionRateNum()  // 实际收货数量
        orderItemNum := item.PurchaseOrderItem.GetUnitsTotalConversionRateNum()  // 采购订单数量
        
        // 从在途仓库扣减实际收货数量
        err := s.reduceTransitStock(ctx, item, actualNum)
        if err != nil {
            return err
        }
        
        // 如果少收，记录剩余库存
        if actualNum < orderItemNum {
            remainingNum := orderItemNum - actualNum
            // 记录剩余库存信息，用于后续处理
            s.recordRemainingStock(ctx, item, remainingNum)
        }
    }
    return nil
}

// 清理剩余库存（手动或自动）
func (s *purchaseReceiptOrderSrv) clearRemainingStock(ctx context.Context, purchaseOrderUuid uint64) error {
    // 1. 查询剩余在途库存
    // 2. 清理剩余库存
    // 3. 记录清理日志
    // 4. 更新采购订单状态
}
```

#### 3.2.2 场景2：实际收货数量 > 采购订单数量（多收）

**问题描述**：
- 采购订单数量：100
- 实际收货数量：120
- 在途仓库库存：100（创建 Purchase Order 时添加）
- 多收数量：20

**当前实现**（`main/app/service/purchase_order/receipt_order.go`）：

```go
// 从在途仓库扣减，但限制为库存数量
actualNum := item.GetUnitsTotalConversionRateNum()  // 120
warehouseItem, err := warehouseItemRepo.GetByWarehouseAndMaterial(transitWarehouse.Uuid, item.MaterialUuid)
if warehouseItem.Stock < actualNum {
    actualNum = warehouseItem.Stock  // 限制为 100（在途仓库库存）
}
// 减少在途仓库库存：100 - 100 = 0
err = warehouseItemRepo.ReduceStock(warehouseItem.Uuid, actualNum)
```

**处理结果**：
- ✅ 在途仓库库存：0（全部扣减）
- ✅ 目标仓库增加库存：100（限制为在途仓库库存）
- ⚠️ **问题**：多收的 20 个物品没有增加到目标仓库

**处理方案**：

**方案 A：允许多收，补充在途库存（推荐）**
- 如果实际收货数量 > 在途仓库库存，允许多收
- 自动补充在途仓库库存（从预期库存中补充）
- 然后从在途仓库扣减全部实际收货数量

**方案 B：限制为在途仓库库存**
- 如果实际收货数量 > 在途仓库库存，限制为在途仓库库存
- 多收部分不处理，需要人工处理

**方案 C：拒绝多收**
- 如果实际收货数量 > 采购订单数量，拒绝收货
- 要求供应商按订单数量发货

**代码实现建议（方案 A）**：

```go
// 处理多收场景
func (s *purchaseReceiptOrderSrv) handleOverReceipt(ctx context.Context, receiptOrder *model.PurchaseReceiptOrder) error {
    for _, item := range receiptOrder.Items {
        actualNum := item.GetUnitsTotalConversionRateNum()  // 实际收货数量：120
        orderItemNum := item.PurchaseOrderItem.GetUnitsTotalConversionRateNum()  // 采购订单数量：100
        
        // 获取在途仓库库存
        warehouseItem, err := warehouseItemRepo.GetByWarehouseAndMaterial(transitWarehouse.Uuid, item.MaterialUuid)
        if err != nil {
            return err
        }
        
        transitStock := warehouseItem.Stock  // 在途仓库库存：100
        
        // 如果实际收货数量 > 在途仓库库存，需要补充在途库存
        if actualNum > transitStock {
            // 补充在途仓库库存（多收部分）
            additionalNum := actualNum - transitStock  // 20
            err := s.addToTransitWarehouse(ctx, item, additionalNum)
            if err != nil {
                return err
            }
            // 记录多收日志
            s.recordOverReceiptLog(ctx, item, additionalNum)
        }
        
        // 从在途仓库扣减全部实际收货数量
        err = warehouseItemRepo.ReduceStock(warehouseItem.Uuid, actualNum)
        if err != nil {
            return err
        }
        
        // 增加到目标仓库（使用实际收货数量）
        err = s.addToTargetWarehouse(ctx, item, actualNum)
        if err != nil {
            return err
        }
    }
    return nil
}

// 补充在途仓库库存（多收部分）
func (s *purchaseReceiptOrderSrv) addToTransitWarehouse(ctx context.Context, item *model.PurchaseReceiptOrderItem, additionalNum float64) error {
    // 1. 添加到在途仓库
    // 2. 记录在途入库日志（场景：多收补充）
    // 3. 记录多收信息，用于后续对账
}
```

**处理结果（方案 A）**：
- ✅ 在途仓库库存：100 + 20 - 120 = 0
- ✅ 目标仓库增加库存：120（全部实际收货数量）
- ✅ 多收部分已处理，库存数量与实际收货数量一致

#### 3.2.3 综合处理方案

**推荐实现**：

```go
// 统一的收货库存处理逻辑
func (s *purchaseReceiptOrderSrv) updateMaterialStock(ctx context.Context, receiptOrder *model.PurchaseReceiptOrder) error {
    for _, item := range receiptOrder.Items {
        actualNum := item.GetUnitsTotalConversionRateNum()  // 实际收货数量
        orderItemNum := item.PurchaseOrderItem.GetUnitsTotalConversionRateNum()  // 采购订单数量
        
        // 获取在途仓库库存
        warehouseItem, err := warehouseItemRepo.GetByWarehouseAndMaterial(transitWarehouse.Uuid, item.MaterialUuid)
        if err != nil {
            return err
        }
        
        transitStock := warehouseItem.Stock  // 在途仓库当前库存
        
        // 场景1：实际收货数量 < 在途仓库库存（少收）
        if actualNum < transitStock {
            // 从在途仓库扣减实际收货数量
            err := warehouseItemRepo.ReduceStock(warehouseItem.Uuid, actualNum)
            if err != nil {
                return err
            }
            // 记录剩余库存信息
            remainingNum := transitStock - actualNum
            s.recordRemainingStock(ctx, item, remainingNum)
        }
        
        // 场景2：实际收货数量 > 在途仓库库存（多收）
        if actualNum > transitStock {
            // 补充在途仓库库存（多收部分）
            additionalNum := actualNum - transitStock
            err := s.addToTransitWarehouse(ctx, item, additionalNum)
            if err != nil {
                return err
            }
            // 从在途仓库扣减全部实际收货数量
            err = warehouseItemRepo.ReduceStock(warehouseItem.Uuid, actualNum)
            if err != nil {
                return err
            }
            // 记录多收日志
            s.recordOverReceiptLog(ctx, item, additionalNum)
        }
        
        // 场景3：实际收货数量 = 在途仓库库存（正常）
        if actualNum == transitStock {
            // 从在途仓库扣减全部库存
            err := warehouseItemRepo.ReduceStock(warehouseItem.Uuid, actualNum)
            if err != nil {
                return err
            }
        }
        
        // 增加到目标仓库（使用实际收货数量）
        err = s.addToTargetWarehouse(ctx, item, actualNum)
        if err != nil {
            return err
        }
    }
    return nil
}
```

**关键点**：
1. ✅ **使用实际收货数量**：无论多收还是少收，都使用实际收货数量更新库存
2. ✅ **处理数量差异**：自动处理多收和少收的情况
3. ✅ **记录差异日志**：记录多收和少收的详细信息，便于后续对账
4. ✅ **库存一致性**：确保库存数量与实际收货数量一致

#### 3.3 数据一致性验证

**验证方法**：

1. **定期对账**
   - 对比 ERPNext 和 TTPOS 的库存数据
   - 发现不一致时，记录差异并告警

2. **关键节点验证**
   - Delivery Note 提交后，验证 TTPOS 在途仓库库存是否正确
   - 门店收货后，验证库存是否正确转入目标仓库
   - 验证实际收货数量与库存数量是否一致

3. **异常处理**
   - 如果同步失败，记录错误日志
   - 提供手动同步功能，允许管理员手动触发同步
   - 处理数量差异导致的在途仓库库存异常

### 4. 风险评估与缓解措施

#### 4.1 库存同步失败风险

**风险描述**：
- Delivery Note 提交后，如果 TTPOS 同步失败，会导致库存不一致
- 门店在途仓库库存未增加，但 ERPNext 已扣减总部库存
- Purchase Receipt 提交后，如果同步失败，会导致库存不一致

**影响程度**：**高**

**同步失败原因分析**：

| 失败原因 | 发生频率 | 影响范围 | 处理难度 |
|---------|---------|---------|---------|
| **网络超时** | 高 | 单个单据 | 低（可重试） |
| **服务不可用** | 中 | 所有同步 | 中（需要告警） |
| **数据格式错误** | 低 | 单个单据 | 高（需要修复） |
| **业务逻辑错误** | 低 | 单个单据 | 高（需要修复） |
| **并发冲突** | 中 | 单个单据 | 中（需要重试） |
| **数据库异常** | 低 | 所有同步 | 高（需要修复） |

**缓解措施**：

1. **重试机制**
   - 同步失败时，自动重试（最多 3 次）
   - 重试失败后，记录错误日志并告警

2. **补偿机制**
   - 提供手动同步功能
   - 定期对账，发现不一致时自动补偿

3. **监控告警**
   - 监控同步成功率
   - 同步失败率超过阈值时，发送告警

---

### 5. 同步失败解决方案

#### 5.1 重试机制

##### 5.1.1 自动重试策略

**重试策略**：

```go
// Webhook 处理函数中的重试逻辑
func HandleWebhookWithRetry(ctx context.Context, req *WebhookReq, handler func(context.Context, *WebhookReq) error) error {
    maxRetries := 3
    baseDelay := time.Second * 2
    
    var lastErr error
    for attempt := 0; attempt < maxRetries; attempt++ {
        // 执行同步操作
        if err := handler(ctx, req); err == nil {
            // 成功，记录成功日志
            logSyncSuccess(ctx, req)
            return nil
        } else {
            lastErr = err
            
            // 判断是否可重试
            if !isRetryableError(err) {
                // 不可重试的错误（如数据格式错误），直接返回
                logSyncFailure(ctx, req, err, attempt+1, false)
                return err
            }
            
            // 可重试的错误，等待后重试
            if attempt < maxRetries-1 {
                // 指数退避 + 随机抖动
                delay := baseDelay * time.Duration(math.Pow(2, float64(attempt)))
                jitter := time.Duration(rand.Intn(1000)) * time.Millisecond
                delay += jitter
                
                logSyncRetry(ctx, req, err, attempt+1, delay)
                time.Sleep(delay)
            }
        }
    }
    
    // 重试失败，记录到重试队列
    logSyncFailure(ctx, req, lastErr, maxRetries, true)
    return enqueueRetryTask(ctx, req, lastErr)
}

// 判断错误是否可重试
func isRetryableError(err error) bool {
    // 网络超时、服务不可用、数据库连接失败等可重试
    if errors.Is(err, context.DeadlineExceeded) {
        return true
    }
    if errors.Is(err, context.Canceled) {
        return true
    }
    // 数据库连接错误
    if strings.Contains(err.Error(), "connection") {
        return true
    }
    // 数据格式错误、业务逻辑错误等不可重试
    if strings.Contains(err.Error(), "validation") {
        return false
    }
    if strings.Contains(err.Error(), "business logic") {
        return false
    }
    return true
}
```

**重试参数配置**：

| 参数 | 默认值 | 说明 |
|------|--------|------|
| **最大重试次数** | 3 | 最多重试 3 次 |
| **基础延迟** | 2 秒 | 第一次重试延迟 |
| **指数退避** | 2^n | 每次重试延迟翻倍 |
| **随机抖动** | 0-1 秒 | 避免同时重试 |

##### 5.1.2 异步重试队列

**重试队列实现**：

```go
// 重试任务表结构
type SyncRetryTask struct {
    Uuid          uint64    `gorm:"primaryKey"`
    WebhookType   string    // webhook 类型（delivery_note, purchase_receipt 等）
    DocumentName  string    // ERPNext 单据名称
    CompanyUuid   uint64    // 公司 UUID
    RequestData   string    // 请求数据（JSON）
    ErrorMessage  string    // 错误信息
    RetryCount    int       // 已重试次数
    MaxRetries    int       // 最大重试次数
    NextRetryTime int64     // 下次重试时间
    Status        int       // 状态（0-待重试，1-重试中，2-成功，3-失败）
    CreatedAt     int64
    UpdatedAt     int64
}

// 加入重试队列
func enqueueRetryTask(ctx context.Context, req *WebhookReq, err error) error {
    task := &SyncRetryTask{
        WebhookType:   req.WebhookType,
        DocumentName:  req.DocumentName,
        CompanyUuid:   req.CompanyUuid,
        RequestData:   req.ToJSON(),
        ErrorMessage:  err.Error(),
        RetryCount:    0,
        MaxRetries:    5,  // 队列中最多重试 5 次
        NextRetryTime: time.Now().Add(time.Minute * 5).Unix(),  // 5 分钟后重试
        Status:        0,  // 待重试
    }
    return syncRetryTaskRepo.Create(task)
}

// 定时任务：处理重试队列
func ProcessRetryQueue(ctx context.Context) error {
    // 1. 查询待重试的任务（NextRetryTime <= 当前时间）
    tasks, err := syncRetryTaskRepo.GetPendingTasks(time.Now().Unix())
    if err != nil {
        return err
    }
    
    for _, task := range tasks {
        // 2. 更新状态为"重试中"
        task.Status = 1
        task.RetryCount++
        syncRetryTaskRepo.Update(task)
        
        // 3. 执行重试
        if err := retrySyncTask(ctx, task); err != nil {
            // 重试失败
            if task.RetryCount >= task.MaxRetries {
                // 超过最大重试次数，标记为失败
                task.Status = 3
                task.ErrorMessage = err.Error()
                syncRetryTaskRepo.Update(task)
                // 发送告警
                sendAlert(ctx, task, err)
            } else {
                // 更新下次重试时间（指数退避）
                delay := time.Duration(math.Pow(2, float64(task.RetryCount))) * time.Minute
                task.NextRetryTime = time.Now().Add(delay).Unix()
                task.Status = 0
                syncRetryTaskRepo.Update(task)
            }
        } else {
            // 重试成功
            task.Status = 2
            syncRetryTaskRepo.Update(task)
        }
    }
    
    return nil
}
```

**重试队列配置**：

| 参数 | 默认值 | 说明 |
|------|--------|------|
| **队列最大重试次数** | 5 | 队列中最多重试 5 次 |
| **初始重试延迟** | 5 分钟 | 第一次重试延迟 |
| **重试间隔** | 指数退避 | 2^n 分钟 |
| **处理频率** | 每 1 分钟 | 定时任务处理频率 |

#### 5.2 补偿机制

##### 5.2.1 定时对账

**对账任务实现**：

```go
// 定时对账任务（每 30 分钟执行一次）
func ReconcileStockSync(ctx context.Context) error {
    // 1. 查询最近 24 小时内的 ERPNext 单据
    //    - Delivery Note（已提交）
    //    - Purchase Receipt（已提交）
    //    - Sales Invoice (Credit Note)（已提交）
    //    - Purchase Invoice (Debit Note)（已提交）
    
    // 2. 对比 TTPOS 和 ERPNext 的库存变化
    //    - 检查 TTPOS 是否已同步
    //    - 检查库存数量是否一致
    
    // 3. 发现不一致时，自动补偿
    //    - 记录差异日志
    //    - 自动同步库存变化
    //    - 发送告警通知
    
    return nil
}

// 对账逻辑
func reconcileDeliveryNote(ctx context.Context, deliveryNoteName string) error {
    // 1. 从 ERPNext 获取 Delivery Note 信息
    deliveryNote, err := erpService.GetDeliveryNote(ctx, deliveryNoteName)
    if err != nil {
        return err
    }
    
    // 2. 检查 TTPOS 是否已同步
    syncRecord, err := syncRecordRepo.GetByDocumentName(deliveryNoteName)
    if err != nil || syncRecord == nil {
        // 未同步，需要补偿
        return compensateSync(ctx, deliveryNote)
    }
    
    // 3. 检查库存是否一致
    if !isStockConsistent(ctx, deliveryNote) {
        // 库存不一致，需要补偿
        return compensateStock(ctx, deliveryNote)
    }
    
    return nil
}
```

**对账配置**：

| 参数 | 默认值 | 说明 |
|------|--------|------|
| **对账频率** | 每 30 分钟 | 定时任务执行频率 |
| **对账范围** | 最近 24 小时 | 检查最近 24 小时内的单据 |
| **差异阈值** | 0 | 允许的库存差异（0 表示不允许差异） |

##### 5.2.2 自动补偿

**补偿逻辑**：

```go
// 自动补偿同步
func compensateSync(ctx context.Context, document interface{}) error {
    // 1. 根据单据类型选择对应的处理函数
    switch doc := document.(type) {
    case *DeliveryNote:
        return handleDeliveryNoteSubmitted(ctx, doc.Name)
    case *PurchaseReceipt:
        return handlePurchaseReceiptSubmitted(ctx, doc.Name)
    case *SalesInvoice:
        if doc.IsCreditNote {
            return handleSalesInvoiceCreditNoteSubmitted(ctx, doc.Name)
        }
    case *PurchaseInvoice:
        if doc.IsDebitNote {
            return handlePurchaseInvoiceDebitNoteSubmitted(ctx, doc.Name)
        }
    }
    
    return nil
}

// 补偿库存差异
func compensateStock(ctx context.Context, deliveryNote *DeliveryNote) error {
    // 1. 计算库存差异
    differences := calculateStockDifferences(ctx, deliveryNote)
    
    // 2. 自动补偿差异
    for _, diff := range differences {
        if diff.ExpectedStock > diff.ActualStock {
            // 库存不足，需要增加
            err := addStock(ctx, diff.Warehouse, diff.Material, diff.ExpectedStock-diff.ActualStock)
            if err != nil {
                return err
            }
        } else if diff.ExpectedStock < diff.ActualStock {
            // 库存过多，需要扣减
            err := reduceStock(ctx, diff.Warehouse, diff.Material, diff.ActualStock-diff.ExpectedStock)
            if err != nil {
                return err
            }
        }
    }
    
    // 3. 记录补偿日志
    logCompensation(ctx, deliveryNote, differences)
    
    return nil
}
```

#### 5.3 监控告警

##### 5.3.1 监控指标

**关键指标**：

```go
// 监控指标定义
type SyncMetrics struct {
    // 同步成功率
    SuccessRate float64 `json:"success_rate"`
    
    // 同步失败率
    FailureRate float64 `json:"failure_rate"`
    
    // 平均同步延迟（毫秒）
    AvgLatency int64 `json:"avg_latency"`
    
    // 重试队列大小
    RetryQueueSize int `json:"retry_queue_size"`
    
    // 对账差异数量
    ReconciliationDifferences int `json:"reconciliation_differences"`
    
    // 补偿次数
    CompensationCount int `json:"compensation_count"`
}

// 记录监控指标
func recordSyncMetrics(ctx context.Context, metrics *SyncMetrics) {
    // 1. 记录到监控系统（Prometheus、Grafana 等）
    // 2. 更新数据库指标表
    // 3. 检查告警阈值
    checkAlertThresholds(ctx, metrics)
}
```

**监控指标配置**：

| 指标 | 阈值 | 告警级别 | 说明 |
|------|------|---------|------|
| **同步失败率** | > 5% | Warning | 5 分钟内失败率超过 5% |
| **同步失败率** | > 10% | Critical | 5 分钟内失败率超过 10% |
| **重试队列大小** | > 100 | Warning | 重试队列积压超过 100 |
| **重试队列大小** | > 500 | Critical | 重试队列积压超过 500 |
| **对账差异数量** | > 0 | Warning | 发现库存差异 |
| **平均同步延迟** | > 5 秒 | Warning | 平均同步延迟超过 5 秒 |
| **补偿次数** | > 10/小时 | Warning | 1 小时内补偿次数超过 10 |

##### 5.3.2 告警规则

**告警实现**：

```go
// 检查告警阈值
func checkAlertThresholds(ctx context.Context, metrics *SyncMetrics) {
    alerts := []Alert{}
    
    // 1. 同步失败率告警
    if metrics.FailureRate > 0.1 {
        alerts = append(alerts, Alert{
            Level:   "Critical",
            Message: fmt.Sprintf("同步失败率过高: %.2f%%", metrics.FailureRate*100),
            Metric:  "failure_rate",
        })
    } else if metrics.FailureRate > 0.05 {
        alerts = append(alerts, Alert{
            Level:   "Warning",
            Message: fmt.Sprintf("同步失败率较高: %.2f%%", metrics.FailureRate*100),
            Metric:  "failure_rate",
        })
    }
    
    // 2. 重试队列大小告警
    if metrics.RetryQueueSize > 500 {
        alerts = append(alerts, Alert{
            Level:   "Critical",
            Message: fmt.Sprintf("重试队列积压严重: %d", metrics.RetryQueueSize),
            Metric:  "retry_queue_size",
        })
    } else if metrics.RetryQueueSize > 100 {
        alerts = append(alerts, Alert{
            Level:   "Warning",
            Message: fmt.Sprintf("重试队列积压: %d", metrics.RetryQueueSize),
            Metric:  "retry_queue_size",
        })
    }
    
    // 3. 对账差异告警
    if metrics.ReconciliationDifferences > 0 {
        alerts = append(alerts, Alert{
            Level:   "Warning",
            Message: fmt.Sprintf("发现库存差异: %d 处", metrics.ReconciliationDifferences),
            Metric:  "reconciliation_differences",
        })
    }
    
    // 4. 发送告警
    for _, alert := range alerts {
        sendAlert(ctx, alert)
    }
}

// 发送告警
func sendAlert(ctx context.Context, alert Alert) {
    // 1. 记录告警日志
    logger.Logger.Warn("同步告警", zap.String("level", alert.Level), zap.String("message", alert.Message))
    
    // 2. 发送到告警系统（邮件、短信、钉钉、企业微信等）
    alertService.Send(ctx, alert)
    
    // 3. 记录到数据库
    alertRepo.Create(&model.Alert{
        Level:   alert.Level,
        Message: alert.Message,
        Metric:  alert.Metric,
    })
}
```

#### 5.4 手动同步功能

##### 5.4.1 管理后台手动同步

**手动同步接口**：

```go
// 手动同步接口
POST /api/admin/sync/manual
{
    "document_type": "delivery_note",  // 单据类型
    "document_name": "DN-00001",       // 单据名称
    "company_uuid": 12345              // 公司 UUID
}

// 实现
func ManualSync(ctx context.Context, req *ManualSyncReq) error {
    // 1. 验证权限（仅管理员可操作）
    if !ctx.IsAdmin() {
        return errors.New("权限不足")
    }
    
    // 2. 根据单据类型选择处理函数
    switch req.DocumentType {
    case "delivery_note":
        return handleDeliveryNoteSubmitted(ctx, req.DocumentName)
    case "purchase_receipt":
        return handlePurchaseReceiptSubmitted(ctx, req.DocumentName)
    case "sales_invoice":
        return handleSalesInvoiceCreditNoteSubmitted(ctx, req.DocumentName)
    case "purchase_invoice":
        return handlePurchaseInvoiceDebitNoteSubmitted(ctx, req.DocumentName)
    }
    
    return errors.New("不支持的单据类型")
}
```

##### 5.4.2 批量同步功能

**批量同步接口**：

```go
// 批量同步接口
POST /api/admin/sync/batch
{
    "document_type": "delivery_note",
    "date_range": {
        "start": "2025-01-01",
        "end": "2025-01-31"
    },
    "company_uuid": 12345
}

// 实现
func BatchSync(ctx context.Context, req *BatchSyncReq) error {
    // 1. 查询指定日期范围内的单据
    documents, err := erpService.GetDocuments(ctx, req.DocumentType, req.DateRange)
    if err != nil {
        return err
    }
    
    // 2. 批量同步
    successCount := 0
    failCount := 0
    for _, doc := range documents {
        if err := manualSyncDocument(ctx, doc); err != nil {
            failCount++
            logger.Logger.Error("批量同步失败", zap.String("document", doc.Name), zap.Error(err))
        } else {
            successCount++
        }
    }
    
    // 3. 返回结果
    return &BatchSyncResp{
        Total:   len(documents),
        Success: successCount,
        Failed:  failCount,
    }
}
```

#### 5.5 数据一致性检查

##### 5.5.1 一致性检查任务

**检查实现**：

```go
// 数据一致性检查任务（每 1 小时执行一次）
func CheckDataConsistency(ctx context.Context) error {
    // 1. 检查最近 24 小时内的单据
    // 2. 对比 TTPOS 和 ERPNext 的库存数据
    // 3. 发现不一致时，记录差异并告警
    
    inconsistencies := []Inconsistency{}
    
    // 检查 Delivery Note
    deliveryNotes, _ := erpService.GetDeliveryNotes(ctx, time.Now().Add(-24*time.Hour))
    for _, dn := range deliveryNotes {
        if inconsistency := checkDeliveryNoteConsistency(ctx, dn); inconsistency != nil {
            inconsistencies = append(inconsistencies, *inconsistency)
        }
    }
    
    // 检查 Purchase Receipt
    purchaseReceipts, _ := erpService.GetPurchaseReceipts(ctx, time.Now().Add(-24*time.Hour))
    for _, pr := range purchaseReceipts {
        if inconsistency := checkPurchaseReceiptConsistency(ctx, pr); inconsistency != nil {
            inconsistencies = append(inconsistencies, *inconsistency)
        }
    }
    
    // 记录不一致数据
    if len(inconsistencies) > 0 {
        logInconsistencies(ctx, inconsistencies)
        sendConsistencyAlert(ctx, inconsistencies)
    }
    
    return nil
}
```

##### 5.5.2 差异报告

**报告格式**：

```go
// 差异报告
type ConsistencyReport struct {
    CheckTime      time.Time        `json:"check_time"`
    TotalChecked   int              `json:"total_checked"`
    Inconsistencies []Inconsistency `json:"inconsistencies"`
    Summary        ReportSummary    `json:"summary"`
}

type Inconsistency struct {
    DocumentType   string  `json:"document_type"`
    DocumentName   string  `json:"document_name"`
    MaterialCode   string  `json:"material_code"`
    Warehouse      string  `json:"warehouse"`
    ExpectedStock  float64 `json:"expected_stock"`  // ERPNext 库存
    ActualStock    float64 `json:"actual_stock"`    // TTPOS 库存
    Difference     float64 `json:"difference"`      // 差异
    Severity       string  `json:"severity"`        // 严重程度
}
```

#### 5.6 死信队列处理

##### 5.6.1 死信队列机制

**死信队列实现**：

```go
// 死信队列：超过最大重试次数后，移入死信队列
type DeadLetterQueue struct {
    Uuid          uint64
    WebhookType   string
    DocumentName  string
    RequestData   string
    ErrorMessage  string
    RetryCount    int
    CreatedAt     int64
    ProcessedAt   int64
    Status        int  // 0-待处理，1-已处理，2-已忽略
}

// 移入死信队列
func moveToDeadLetterQueue(ctx context.Context, task *SyncRetryTask) error {
    deadLetter := &DeadLetterQueue{
        WebhookType:  task.WebhookType,
        DocumentName: task.DocumentName,
        RequestData:  task.RequestData,
        ErrorMessage: task.ErrorMessage,
        RetryCount:   task.RetryCount,
        Status:       0,  // 待处理
    }
    return deadLetterQueueRepo.Create(deadLetter)
}

// 处理死信队列（管理员手动处理）
func ProcessDeadLetterQueue(ctx context.Context, uuid uint64) error {
    deadLetter, err := deadLetterQueueRepo.GetByUuid(uuid)
    if err != nil {
        return err
    }
    
    // 1. 分析错误原因
    // 2. 修复数据问题
    // 3. 重新同步
    // 4. 标记为已处理
    
    return nil
}
```

#### 5.7 异常处理流程

**完整异常处理流程**：

```
Webhook 接收
    ↓
同步处理
    ↓
失败？
    ├─ 否 → 记录成功日志 ✅
    └─ 是 → 判断错误类型
            ├─ 可重试错误
            │    ↓
            │   自动重试（最多 3 次）
            │    ↓
            │   成功？
            │    ├─ 是 → 记录成功日志 ✅
            │    └─ 否 → 加入重试队列
            │            ↓
            │           定时任务重试（最多 5 次）
            │            ↓
            │           成功？
            │           ├─ 是 → 记录成功日志 ✅
            │           └─ 否 → 移入死信队列
            │                   ↓
            │                   管理员手动处理
            │
            └─ 不可重试错误
                 ↓
                记录错误日志
                发送告警
                移入死信队列
                管理员手动处理
```

**处理优先级**：

1. **自动重试**：网络超时、服务不可用等临时错误
2. **重试队列**：需要延迟重试的错误
3. **定时对账**：兜底机制，确保数据一致性
4. **手动同步**：管理员手动触发同步
5. **死信队列**：需要人工处理的错误

#### 4.2 在途仓库库存异常风险

**风险描述**：
- 如果物品在运输途中丢失，在途仓库库存会一直存在
- 需要定期清理异常的在途库存

**影响程度**：**中**

**缓解措施**：
1. **超时清理**
   - 在途库存超过一定时间（如 30 天）未转入目标仓库，自动标记为异常
   - 需要管理员手动处理

2. **异常告警**
   - 在途库存超过阈值时，发送告警
   - 提醒管理员检查运输状态

### 5. 实施建议

#### 5.1 短期方案（过渡期）

1. **保持现有逻辑**
   - 暂时保持现有的在途仓库添加逻辑
   - 确保功能可用

2. **添加同步机制**
   - 实现 Delivery Note 状态同步（Webhook 或轮询）
   - 确保库存同步及时

#### 5.2 长期方案（优化后）

1. **优化在途仓库使用时机**
   - 在 Delivery Note 提交后，才添加到在途仓库
   - 确保库存流转的准确性

2. **完善监控和告警**
   - 实现库存同步监控
   - 实现异常告警机制

---

## 🔗 相关文档

- [仓库组（母仓库）子仓库分配方案](./2026-01-05-warehouse-group-allocation.md)
- [品牌采购流程中门店仓库选择分析](./2026-01-05-warehouse-selection-analysis.md)
- [品牌采购流程 ERPNext 实现方案](../shared/specs/brand-procurement-erpnext-implementation.md)
- [ERPNext 仓库层级管理实现方案](../shared/integrations/erpnext-warehouse-hierarchy.md)

---

**版本**: v1.0.0  
**创建日期**: 2026-01-15  
**维护者**: TTPOS Team

