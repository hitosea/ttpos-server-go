# ERP 商品供应商字段扩展 需求提案

## 📋 提案信息

| 项目          | 内容                |
| ------------- | ------------------- |
| **提案人**    | rikugun             |
| **日期**      | 2026-01-26          |
| **目标版本**  | v2.16               |
| **状态**      | ✅ 已通过           |
| **关联 Spec** | story-erp-item-supplier-fields |

---

## 🎯 背景和动机

### 问题描述

当前 TTPOS 的 ItemInfo 数据结构缺少供应商相关字段，具体问题包括：

1. **缺少供应商直配标识**：无法区分哪些商品由供应商直接配送
2. **缺少商品供应商关联**：无法查看商品关联的供应商列表
3. **ERPNext 数据未同步**：ERPNext 已有 `delivered_by_supplier` 和 `supplier_items` 字段，但 TTPOS 未同步这些数据

### 业务价值

- **效率提升**：减少手动查询 ERPNext 的操作，一站式查看商品供应商信息
- **数据一致性**：确保 TTPOS 与 ERPNext 供应商数据同步，避免数据不一致导致的采购问题
- **采购流程优化**：便于判断商品采购来源和配送方式，优化采购决策

### 目标用户

- [x] 店长
- [x] 商户管理员
- [x] 采购人员

---

## 💡 解决方案概述

### 方案描述

在 `ItemInfo` protobuf 消息中新增两个字段：`delivered_by_supplier`（布尔值，标识是否由供应商直接配送）和 `supplier_items`（数组，包含商品关联的供应商信息）。同时修改 `GetItem` 和 `GetItemList` 接口的 logic 层实现，从 ERPNext 查询接口获取这些字段数据并填充返回。

### 核心功能点

1. **Protobuf 字段扩展**：在 `item.proto` 的 `ItemInfo` 消息中新增 `delivered_by_supplier` 和 `supplier_items` 字段
2. **GetItem 接口扩展**：修改 `sItem.GetItem` 方法，查询并返回供应商相关字段
3. **GetItemList 接口扩展**：修改 `sItem.GetItemList` 方法，批量查询并返回供应商相关字段

### 影响范围

**涉及终端**：
- [x] Shop 商家管理端

**涉及模块**：
- [x] API 接口（protobuf 定义）
- [x] 数据模型（ItemInfo 扩展）
- [x] 业务逻辑（logic 层实现）
- [x] 其他: ERPNext 查询接口对接

---

## 📊 初步评估

### 技术复杂度

- [x] **中**：需要前后端联调,基础业务逻辑

### 工作量预估

- **预估 SP**: 3（待技术评审确认）

### 拆分预估

**是否需要拆分**：
- [x] **否**：单终端，SP ≤ 5，可直接创建 1 个 Spec

### 风险识别

**潜在风险**：
1. ERPNext 接口返回数据格式变更可能导致解析失败
2. 批量查询 supplier_items 可能影响 GetItemList 接口性能

**缓解措施**：
1. 与 ERPNext 团队确认接口稳定性，增加字段校验和默认值处理
2. 评估是否需要按需加载（lazy loading）或缓存优化

---

## 🤝 需求评审

### 评审参与人

| 角色       | 姓名   | 签名/日期 |
| ---------- | ------ | --------- |
| 产品经理   |        |           |
| 技术负责人 |        |           |
| 开发代表   |        |           |
| 测试代表   |        |           |

### 评审结论

- [x] ✅ **批准**：进入技术方案设计阶段
- [ ] 🔄 **修改后批准**：需补充以下内容
- [ ] ❌ **拒绝**：不符合产品规划或优先级

**评审意见**：

```
已通过评审，进入 Spec 阶段。
```

**下一步行动**：

- [x] 创建 Spec：`story-erp-item-supplier-fields`
- [x] 分配负责人：rikugun
- [ ] 目标 Sprint：待定

---

## 📝 附录

### User Story（初稿）

**作为** 采购人员或店长
**我想** 在查看商品信息时能看到供应商直配标识和关联的供应商列表
**以便于** 快速判断商品的采购来源和配送方式，优化采购决策

### AC 验收标准（初稿）

1. **WHEN** 调用 GetItem 接口 **THEN** 系统 **SHALL** 返回 `delivered_by_supplier` 和 `supplier_items` 字段
2. **WHEN** 调用 GetItemList 接口 **THEN** 系统 **SHALL** 在每个 ItemInfo 中包含供应商相关字段
3. **WHEN** ERPNext 中商品无供应商信息 **THEN** 系统 **SHALL** 返回 `delivered_by_supplier=false` 和空的 `supplier_items` 数组

### 技术参考

**涉及文件**：
- `ttpos-bmp/app/ttpos-erp/manifest/protobuf/item/item.proto` - 新增字段定义
- `ttpos-bmp/app/ttpos-erp/internal/logic/item/item.go` - 修改 GetItem 和 GetItemList 实现

**新增字段设计**：
```protobuf
message ItemInfo {
  // ... existing fields ...
  bool delivered_by_supplier = 27; // 是否由供应商直接配送
  repeated SupplierItem supplier_items = 28; // 关联的供应商列表
}

message SupplierItem {
  string supplier = 1; // 供应商名称
  int32 idx = 2; // 排序索引
}
```

**ERPNext 原始数据结构参考**：
```json
{
  "supplier_items": [
    {
      "name": "q32pfrfrhl",
      "owner": "qx123@123.com",
      "creation": "2025-09-29 14:30:00.231359",
      "modified": "2025-10-30 20:17:01.009826",
      "modified_by": "qx123@123.com",
      "docstatus": 0,
      "idx": 1,
      "supplier": "123123",
      "parent": "WPR3689157545295873",
      "parentfield": "supplier_items",
      "parenttype": "Item",
      "doctype": "Item Supplier"
    }
  ]
}
```

> 注：仅映射业务必需字段（supplier, idx），ERPNext 内部 ID（name）及元数据字段不传递给前端。

---

**版本**: v1.0.0
