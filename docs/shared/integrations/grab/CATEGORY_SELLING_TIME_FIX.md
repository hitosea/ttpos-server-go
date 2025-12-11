# 分类 sellingTimeID 字段添加

## 问题
根据 Grab 菜单格式规范（参考 `1.json`），分类对象缺少 `sellingTimeID` 字段。

## 解决方案

### 修改位置
文件：`main/app/modules/takeout/infrastructure/adapter/grab/grab_converter.go`

### 修改内容
在 `convertTTPOSCategory` 方法中添加：

```go
// 设置售卖时段 ID（默认使用全天）
categoryVO.SellingTimeID = "SELLINGTIME-01"
```

### Grab 菜单结构
根据 `1.json` 的标准格式，分类对象应该包含以下字段：

```json
{
  "id": "CATEGORY-01",
  "name": "Savoury Pancakes",
  "sequence": 1,
  "availableStatus": "AVAILABLE",
  "sellingTimeID": "SELLINGTIME-01",  // ← 必需字段
  "nameTranslation": {...},
  "items": [...]
}
```

### 关联关系
- **分类** (`Category`) 通过 `sellingTimeID` 关联到 **售卖时段** (`SellingTime`)
- **商品** (`MenuItem`) 也通过 `sellingTimeID` 关联到 **售卖时段**
- 这样可以控制不同分类和商品在不同时段的可见性

### 当前实现
- 默认所有分类使用 `"SELLINGTIME-01"`（全天可售）
- 未来可以根据业务需求，从数据库读取分类的实际售卖时段配置

## 验证
编译后，分类的 JSON 输出应该包含 `sellingTimeID` 字段：

```json
{
  "id": "CAT-123",
  "name": "Manager's Recommendation",
  "sequence": 1,
  "availableStatus": "AVAILABLE",
  "sellingTimeID": "SELLINGTIME-01",
  "items": [...]
}
```

---

**修复日期**: 2025-12-10
**修复文件**: `main/app/modules/takeout/infrastructure/adapter/grab/grab_converter.go`

