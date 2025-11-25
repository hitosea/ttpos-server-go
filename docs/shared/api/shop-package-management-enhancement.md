# 旧管理端-商品管理-套餐功能增强 API 文档

> 本文档描述旧管理端商品管理中套餐功能的增强 API 变更。

## 📋 概述

本次更新增强了套餐商品管理功能，支持更灵活的分组和选择机制，包括分组类型（固定/可选）、可选数量控制、必选/默认选项配置、价格加价模式等功能。

**版本**: v2.10.0  
**更新日期**: 2025-11-25  
**关联 Spec**: [story-shop-package-management](../../specs/story-shop-package-management/requirements.md)

---

## 🔄 API 变更

### 1. 保存套餐组接口

**接口**: `POST /api/shop/product.store.product/edit`

#### 请求参数变更

**新增字段**（在 `package_group` 数组中）：

| 字段 | 类型 | 必填 | 说明 | 默认值 |
|------|------|------|------|--------|
| `group_type` | int | 否 | 分组类型：0-固定，1-可选 | 0 |
| `optional_count` | int | 否 | 可选数量（当 group_type=1 时有效） | 0 |

**新增字段**（在 `package_group[].product_list` 数组中）：

| 字段 | 类型 | 必填 | 说明 | 默认值 |
|------|------|------|------|--------|
| `add_price` | decimal | 否 | 加价金额 | 0 |
| `is_required` | int | 否 | 是否必选：0-否，1-是 | 0 |
| `is_default` | int | 否 | 是否默认：0-否，1-是 | 0 |

**字段变更**：

| 字段 | 变更 | 说明 |
|------|------|------|
| `num` | 默认值变更 | 从 0 改为 1 |

#### 请求示例

```json
{
  "product_id": 123456,
  "type": 30,
  "package_price": 100.00,
  "package_group": [
    {
      "group_id": 789012,
      "group_name": "主菜",
      "group_type": 1,
      "optional_count": 2,
      "product_list": [
        {
          "item_id": 345678,
          "product_id": 456789,
          "num": 1,
          "sort": 0,
          "add_price": 5.00,
          "is_required": 1,
          "is_default": 0
        }
      ]
    }
  ]
}
```

#### 响应格式

**成功响应**:

```json
{
  "code": 1,
  "message": "更新成功",
  "data": {}
}
```

**错误响应**:

```json
{
  "code": 0,
  "message": "必选不可大于可选数量",
  "data": {}
}
```

#### 数据校验规则

1. **必选数量校验**：
   - 当 `group_type` 为 1（可选）时，该组中 `is_required=1` 的商品数量不能大于 `optional_count`
   - 如果违反此规则，返回错误："必选不可大于可选数量"

---

### 2. 商品详情接口

**接口**: `GET /api/shop/product.store.product/edit?product_id={product_id}`

#### 响应数据变更

**新增字段**（在 `package_group` 数组中）：

| 字段 | 类型 | 说明 |
|------|------|------|
| `group_type` | int | 分组类型：0-固定，1-可选 |
| `optional_count` | int | 可选数量 |

**新增字段**（在 `package_group[].product_list` 数组中）：

| 字段 | 类型 | 说明 |
|------|------|------|
| `add_price` | decimal | 加价金额 |
| `is_required` | int | 是否必选：0-否，1-是 |
| `is_default` | int | 是否默认：0-否，1-是 |

#### 响应示例

```json
{
  "code": 1,
  "message": "success",
  "data": {
    "model": {
      "uuid": 123456,
      "package": {
        "package_price": 100.00,
        "package_stock": 1000,
        "is_open_stock": 1,
        "package_group": [
          {
            "group_id": 789012,
            "group_name": "主菜",
            "group_name_text": "主菜",
            "group_type": 1,
            "optional_count": 2,
            "product_list": [
              {
                "item_id": 345678,
                "product_id": 456789,
                "product_name_text": "宫保鸡丁",
                "spec_name_text": "标准",
                "product_price": 25.00,
                "stock_num": 100,
                "num": 1,
                "sort": 0,
                "add_price": 5.00,
                "is_required": 1,
                "is_default": 0
              }
            ]
          }
        ]
      }
    }
  }
}
```

---

## 📝 字段说明

### group_type（分组类型）

- **0 - 固定**：该组商品必须全部选择，不可选
- **1 - 可选**：该组商品可以选择，需要配合 `optional_count` 使用

### optional_count（可选数量）

- 当 `group_type=1` 时有效
- 表示该组中要求选择多少个商品
- 默认值为 0，建议设置为 1 或以上

### is_required（是否必选）

- **0 - 否**：该商品不是必选的
- **1 - 是**：该商品是必选的
- 当 `group_type=1` 时，必选商品数量不能大于 `optional_count`

### is_default（是否默认）

- **0 - 否**：该商品不是默认选择的
- **1 - 是**：该商品是默认选择的
- 用于前端展示默认选中的商品

### add_price（加价金额）

- 该商品在套餐中的加价金额
- 默认值为 0
- 用于支持商品加价模式

### num（商品数量）

- 商品数量
- **默认值从 0 改为 1**

---

## 🔍 业务逻辑说明

### 可选分组配置示例

假设有一个套餐组：
- `group_type = 1`（可选）
- `optional_count = 2`（可选2个）
- 包含3个商品：
  - 商品A：`is_required = 1`（必选）
  - 商品B：`is_required = 0`（可选）
  - 商品C：`is_required = 0`（可选）

**配置结果**：
- 用户必须选择商品A（必选）
- 用户可以从商品B和商品C中选择1个（因为可选数量是2，必选1个，剩余可选1个）

### 数据校验示例

**错误配置**：
- `group_type = 1`
- `optional_count = 1`
- 包含2个必选商品（`is_required = 1`）

**结果**：返回错误 "必选不可大于可选数量"

---

## 🧪 测试场景

### 场景1：固定分组

```json
{
  "group_type": 0,
  "optional_count": 0,
  "product_list": [
    {"is_required": 0, "is_default": 0}
  ]
}
```

**预期**：保存成功，所有商品必须选择

### 场景2：可选分组（正常）

```json
{
  "group_type": 1,
  "optional_count": 2,
  "product_list": [
    {"is_required": 1, "is_default": 0},
    {"is_required": 0, "is_default": 1}
  ]
}
```

**预期**：保存成功

### 场景3：可选分组（必选数量大于可选数量）

```json
{
  "group_type": 1,
  "optional_count": 1,
  "product_list": [
    {"is_required": 1, "is_default": 0},
    {"is_required": 1, "is_default": 0}
  ]
}
```

**预期**：返回错误 "必选不可大于可选数量"

### 场景4：商品详情回显

**请求**：`GET /api/shop/product.store.product/edit?product_id=123456`

**预期**：返回的 `package_group` 包含所有新字段

---

## 📚 相关文档

- [需求文档](../../specs/story-shop-package-management/requirements.md)
- [设计文档](../../specs/story-shop-package-management/design.md)
- [任务分解](../../specs/story-shop-package-management/tasks.md)

---

**最后更新**: 2025-11-25  
**维护者**: 开发组

