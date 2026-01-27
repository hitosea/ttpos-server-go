# ERP 商品供应商字段扩展 需求文档

## 📋 基本信息

| 项目              | 内容                                                                 |
| ----------------- | -------------------------------------------------------------------- |
| **Spec ID**       | story-erp-item-supplier-fields                                       |
| **来源 Proposal** | [erp-item-supplier-fields](../../../team/proposals/2026-01/erp-item-supplier-fields.md) |
| **创建日期**      | 2026-01-26                                                           |
| **负责人**        | rikugun                                                              |
| **目标版本**      | v2.16                                                                |

## 📋 审核状态

| 项目         | 内容       |
| ------------ | ---------- |
| **审核状态** | 🚧 开发中  |
| **审核人**   | rikugun    |
| **审核日期** | 2026-01-26 |

---

## 📝 用户故事

**作为** 采购人员
**我想** 通过 GetItem 和 GetItemList 接口获取商品的供应商直配标识和供应商列表
**以便于** 快速判断商品采购来源，提升采购决策效率，保持与 ERPNext 数据一致性

---

## 功能需求

### Requirement 1: 扩展 ItemInfo Protobuf 定义

**用户故事**: 作为前端开发者，我想在 ItemInfo 中获取供应商相关字段，以便于展示商品供应商信息

#### 验收标准

1. **WHEN** 查看 item.proto **THEN** ItemInfo **SHALL** 包含 `delivered_by_supplier` (bool) 字段
2. **WHEN** 查看 item.proto **THEN** ItemInfo **SHALL** 包含 `supplier_items` (repeated SupplierItem) 字段
3. **WHEN** 查看 item.proto **THEN** SupplierItem **SHALL** 包含 `supplier` (string) 和 `idx` (int32) 字段

### Requirement 2: 修改 GetItem 接口实现

**用户故事**: 作为采购人员，我想在查询单个商品时获取其供应商信息，以便于了解商品采购来源

#### 验收标准

1. **WHEN** 调用 GetItem 接口查询商品 **THEN** 系统 **SHALL** 从 ERPNext 获取 `delivered_by_supplier` 字段并返回
2. **WHEN** 调用 GetItem 接口查询商品 **THEN** 系统 **SHALL** 从 ERPNext 获取 `supplier_items` 列表并返回
3. **WHEN** 商品在 ERPNext 中无供应商信息 **THEN** 系统 **SHALL** 返回 `delivered_by_supplier=false` 和空的 `supplier_items` 数组

### Requirement 3: 修改 GetItemList 接口实现

**用户故事**: 作为采购人员，我想在查询商品列表时批量获取供应商信息，以便于快速筛选和对比

#### 验收标准

1. **WHEN** 调用 GetItemList 接口 **THEN** 系统 **SHALL** 在每个 ItemInfo 中包含 `delivered_by_supplier` 字段
2. **WHEN** 调用 GetItemList 接口 **THEN** 系统 **SHALL** 在每个 ItemInfo 中包含 `supplier_items` 列表
3. **WHEN** 列表中某商品无供应商信息 **THEN** 系统 **SHALL** 返回该商品的 `delivered_by_supplier=false` 和空数组

---

## 非功能需求

### 测试要求

- [ ] logic 层单元测试覆盖率 ≥ 80%
- [ ] 接口集成测试验证字段正确返回
- [ ] 边界条件测试：空数组、null 值处理

### 平台兼容性

- [x] Shop 商家管理端（Web/App）
- [ ] 其他终端按需扩展

### 性能要求

- [ ] GetItemList 批量查询供应商信息响应时间增量 < 50ms
- [ ] 评估是否需要缓存优化（如供应商数据变化不频繁）

---

## 约束条件

### 技术约束

- Go 版本: 1.23+
- 框架: GoFrame v2.x
- 模块: ttpos-bmp/app/ttpos-erp
- 必须遵循 CLAUDE.md 和 ttpos-bmp/.cursor/rules/ 规范
- 禁止修改 dao/entity/do 等自动生成文件

### 涉及文件

| 文件路径 | 变更类型 |
|---------|---------|
| `ttpos-bmp/app/ttpos-erp/manifest/protobuf/item/item.proto` | 新增字段定义 |
| `ttpos-bmp/app/ttpos-erp/internal/logic/item/item.go` | 修改 GetItem 和 GetItemList 实现 |

### 数据结构

```protobuf
message ItemInfo {
  // ... existing fields (1-26) ...
  bool delivered_by_supplier = 27; // 是否由供应商直接配送
  repeated SupplierItem supplier_items = 28; // 关联的供应商列表
}

message SupplierItem {
  string supplier = 1; // 供应商名称
  int32 idx = 2; // 排序索引
}
```

### 资源约束

- Story Point: 3

---

## 风险和缓解

### 风险 1: ERPNext 接口返回格式变更

**影响**: 中
**缓解措施**: 增加字段校验和默认值处理，接口调用失败时不影响其他字段返回

### 风险 2: 批量查询性能影响

**影响**: 中
**缓解措施**: 评估是否需要按需加载（lazy loading）或增加缓存层

---

**版本**: v1.0.0
**创建日期**: 2026-01-26
