# story-erp-item-supplier-fields 任务清单

## 📊 进度总览

| 项目 | 数值 |
|------|------|
| 总 SP | 3 |
| 总任务数 | 6 |
| 已完成 | 0 |
| 完成率 | 0% |

---

## Phase 1: Protobuf 和 DTO 层

### 1.1 扩展 ItemInfo Protobuf 定义

| 项目 | 内容 |
|------|------|
| File | `ttpos-bmp/app/ttpos-erp/manifest/protobuf/item/item.proto` |
| Purpose | 新增供应商相关字段 |
| Requirements | Requirement 1 |
| Leverage | 现有 ItemInfo message 结构 |

**变更内容**:
```protobuf
// 在 ItemInfo message 末尾添加
bool delivered_by_supplier = 27; // 是否由供应商直接配送
repeated SupplierItem supplier_items = 28; // 关联的供应商列表

// 新增 message
message SupplierItem {
  string supplier = 1; // 供应商名称
  int32 idx = 2;       // 排序索引
}
```

- [ ] 完成

### 1.2 生成 Protobuf Go 代码

| 项目 | 内容 |
|------|------|
| File | `ttpos-bmp/app/ttpos-erp/api/item/` |
| Purpose | 自动生成 Go 代码 |
| Command | `cd ttpos-bmp/app/ttpos-erp && make pb` |

- [ ] 完成

### 1.3 修改 DTO ItemSupplier 结构体

| 项目 | 内容 |
|------|------|
| File | `ttpos-bmp/app/ttpos-erp/internal/model/dto/erp/item.go` |
| Purpose | 定义 ItemSupplier 类型，修改 SupplierItems 字段类型 |
| Requirements | - |
| Leverage | 现有 Item 结构体 |

**变更内容**:
```go
// 新增结构体（在 Item 结构体后面）
type ItemSupplier struct {
    Name       string `json:"name,omitempty"`
    Supplier   string `json:"supplier,omitempty"`
    Idx        int    `json:"idx,omitempty"`
    Parent     string `json:"parent,omitempty"`
    Parenttype string `json:"parenttype,omitempty"`
    Doctype    string `json:"doctype,omitempty"`
}

// 修改 Item 结构体中的 SupplierItems 字段类型
SupplierItems []ItemSupplier `json:"supplier_items,omitempty"` // 从 []interface{} 改为 []ItemSupplier
```

- [ ] 完成

---

## Phase 2: Logic 层实现

### 2.1 修改 queryItemList 方法

| 项目 | 内容 |
|------|------|
| File | `ttpos-bmp/app/ttpos-erp/internal/logic/stock/item.go` |
| Purpose | 在构建 ItemInfo 时添加供应商字段映射 |
| Requirements | Requirement 2, Requirement 3 |
| Leverage | 现有 `queryItemList` 方法 |
| Line | 约 191-212 行 |

**变更内容**:
在 `itemList = append(itemList, &item.ItemInfo{...})` 中添加：
```go
// 供应商直配标识
DeliveredBySupplier: itemInfo.DeliveredBySupplier == 1,
// 供应商列表
SupplierItems: convertSupplierItems(itemInfo.SupplierItems),
```

**新增辅助函数**:
```go
// convertSupplierItems 将 DTO 供应商列表转换为 Protobuf 格式
func convertSupplierItems(dtoItems []erp.ItemSupplier) []*item.SupplierItem {
    if len(dtoItems) == 0 {
        return make([]*item.SupplierItem, 0)
    }
    result := make([]*item.SupplierItem, 0, len(dtoItems))
    for _, s := range dtoItems {
        result = append(result, &item.SupplierItem{
            Supplier: s.Supplier,
            Idx:      int32(s.Idx),
        })
    }
    return result
}
```

- [ ] 完成

---

## Phase 3: 测试与验证

### 3.1 编写单元测试

| 项目 | 内容 |
|------|------|
| File | `ttpos-bmp/app/ttpos-erp/internal/logic/stock/item_test.go` |
| Purpose | 测试字段映射逻辑 |
| Requirements | 覆盖率 ≥ 80% |

**测试用例**:
1. `TestConvertSupplierItems_Empty` - 空数组返回空数组
2. `TestConvertSupplierItems_Normal` - 正常转换
3. `TestGetItemList_WithSupplierFields` - 列表包含供应商字段

- [ ] 完成

### 3.2 集成测试验证

| 项目 | 内容 |
|------|------|
| Purpose | 验证接口返回值正确 |
| Requirements | 所有验收标准 |

**验证步骤**:
1. 启动服务：`cd ttpos-bmp/app/ttpos-erp && make run`
2. 调用 GetItem 接口，验证返回值包含 `delivered_by_supplier` 和 `supplier_items`
3. 调用 GetItemList 接口，验证每个 ItemInfo 包含供应商字段
4. 测试无供应商商品，验证返回 `delivered_by_supplier=false` 和空数组

- [ ] 完成

---

## 提交清单

### 代码质量
- [ ] `go mod tidy` 执行
- [ ] `gofmt -w .` 执行
- [ ] `go vet ./...` 通过
- [ ] 测试通过: `go test ./internal/logic/stock/...`

### 功能完整性
- [ ] GetItem 接口返回供应商字段
- [ ] GetItemList 接口返回供应商字段
- [ ] 无供应商时返回默认值（false 和空数组）

### BMP 代码生成
- [ ] `make pb` 执行（Protobuf 代码生成）
- [ ] `make service` 执行（如有服务接口变更）

---

**版本**: v1.0.0
**创建日期**: 2026-01-26
