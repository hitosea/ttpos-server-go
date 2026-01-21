# story-erp-item-sales-uom 任务清单

## 📊 进度总览

| 项目     | 数值 |
| -------- | ---- |
| 总 SP    | 2    |
| 总任务数 | 5    |
| 已完成   | 5    |
| 完成率   | 100% |

---

## Phase 1: Proto + DTO 定义

### 1.1 修改 item.proto - 新增 sales_uom 字段

| 项目         | 内容                                                                 |
| ------------ | -------------------------------------------------------------------- |
| File         | `ttpos-bmp/app/ttpos-erp/manifest/protobuf/item/item.proto`          |
| Purpose      | 在 ItemInfo 消息中新增 optional string sales_uom 字段                |
| Requirements | AC1: ItemInfo 包含 sales_uom 字段                                    |
| Leverage     | 参考现有 purchase_uom 字段定义                                       |

**修改内容**:
```protobuf
message ItemInfo {
  // ... 现有字段 ...
  optional string sales_uom = 26; // 销售单位，可选
}
```

- [x] 完成

### 1.2 修改 erp/item.go - 新增 SalesUom 字段

| 项目         | 内容                                                      |
| ------------ | --------------------------------------------------------- |
| File         | `ttpos-bmp/app/ttpos-erp/internal/model/dto/erp/item.go`  |
| Purpose      | 在 Item 结构体中新增 SalesUom 字段                        |
| Requirements | 支持从 ERP 读取和写入 sales_uom 字段                      |
| Leverage     | 参考现有 PurchaseUom 字段定义                             |

**修改内容**:
```go
type Item struct {
    // ... 在 PurchaseUom 后新增 ...
    SalesUom string `json:"sales_uom,omitempty"` // 销售单位
}
```

- [x] 完成

---

## Phase 2: Logic 层实现

### 2.1 修改 queryItemList - 查询返回 sales_uom

| 项目         | 内容                                                        |
| ------------ | ----------------------------------------------------------- |
| File         | `ttpos-bmp/app/ttpos-erp/internal/logic/stock/item.go`      |
| Purpose      | GetItemList 返回结果中包含 sales_uom 字段                   |
| Requirements | AC2: GetItemList 返回 sales_uom；AC3: GetItem 返回 sales_uom |
| Leverage     | 参考 PurchaseUom 的读取方式                                 |

**修改内容**:
1. 在 `queryItemList()` 中的返回 ItemInfo 添加 `SalesUom` 字段映射
2. `GetItem()` 已返回完整 erp.Item，DTO 新增字段后自动包含

- [x] 完成

### 2.2 修改 buildUpdateItemData - 更新支持 sales_uom

| 项目         | 内容                                                   |
| ------------ | ------------------------------------------------------ |
| File         | `ttpos-bmp/app/ttpos-erp/internal/logic/stock/item.go` |
| Purpose      | SaveItem 更新时支持保存 sales_uom 字段                 |
| Requirements | AC4: SaveItem 正确保存 sales_uom；AC5: 未传时不影响    |
| Leverage     | 参考 purchase_uom 的更新逻辑                           |

**修改内容**:
```go
func (s *sItem) buildUpdateItemData(req *item.ItemInfo) g.Map {
    // ... 现有代码 ...
    // 新增 sales_uom 处理
    if req.SalesUom != nil {
        itemForUpdate["sales_uom"] = req.GetSalesUom()
    }
    // ...
}
```

- [x] 完成

### 2.3 修改 buildNewItemData - 创建支持 sales_uom

| 项目         | 内容                                                   |
| ------------ | ------------------------------------------------------ |
| File         | `ttpos-bmp/app/ttpos-erp/internal/logic/stock/item.go` |
| Purpose      | SaveItem 创建时支持保存 sales_uom 字段                 |
| Requirements | AC4: SaveItem 正确保存 sales_uom                       |
| Leverage     | 参考 purchase_uom 的创建逻辑                           |

**修改内容**:
```go
func (s *sItem) buildNewItemData(...) (g.Map, error) {
    newItem := g.Map{
        // ... 现有字段 ...
        "sales_uom": req.GetSalesUom(), // 新增
    }
    // ...
}
```

- [x] 完成

---

## Phase 3: 代码生成与验证

### 3.1 执行 make pb 生成 Go 代码

| 项目         | 内容                                      |
| ------------ | ----------------------------------------- |
| File         | `ttpos-bmp/app/ttpos-erp/`                |
| Purpose      | 根据修改后的 proto 文件生成 Go 代码       |
| Requirements | 生成的代码包含 SalesUom 字段              |
| Command      | `cd ttpos-bmp/app/ttpos-erp && make pb`   |

- [x] 完成

---

## 提交清单

### 代码质量
- [x] `go mod tidy` 执行
- [x] `go fmt ./...` 执行
- [x] `go vet ./...` 通过

### 功能完整性
- [x] AC1: ItemInfo 包含 optional string sales_uom 字段
- [x] AC2: GetItemList 返回 sales_uom 字段（如果有值）
- [x] AC3: GetItem 返回 sales_uom 字段（如果有值）
- [x] AC4: SaveItem 正确保存 sales_uom 字段
- [x] AC5: SaveItem 未传 sales_uom 时不影响其他字段

### 兼容性
- [x] 向后兼容：使用 optional 关键字，不破坏现有调用方

---

**版本**: v1.0.0
