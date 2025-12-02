# 套餐分组类型和加价功能 - 前端对接文档

> 本文档说明后端新增的套餐分组类型和加价功能，供前端开发人员对接使用。

**更新日期**: 2025-11-24  
**版本**: v1.0.0

---

## 📋 概述

后端新增了套餐分组类型和商品加价功能，支持设置固定分组或可选分组，并为每个分组商品设置加价金额。该功能已集成到现有的套餐创建、编辑和查询接口中。

---

## 🔄 API 变更

### 1. 创建套餐接口

**接口地址**: `POST /api/v1/shop/product/add`

**请求参数变更**:

在 `package.groups[]` 数组中，每个分组对象新增以下字段：

```json
{
  "type": 1,
  "locale_name": {...},
  "category_uuid": 123,
  "package": {
    "price": 99.00,
    "internal_code": "PKG001",
    "groups": [
      {
        "locale_name": {
          "zh": "主餐",
          "th": "Main Course",
          "en": "Main Course"
        },
        "group_type": 0,              // ⭐ 新增：分组类型 0-固定 1-可选，默认0
        "optional_count": 1,          // ⭐ 新增：可选数量（可选分组时有效），默认1
        "products": [
          {
            "bom_uuid": 123456,
            "num": 1,
            "sort": 1,
            "add_price": 0.00         // ⭐ 新增：加价金额，默认0
          }
        ]
      }
    ]
  }
}
```

**字段说明**:

| 字段 | 类型 | 必填 | 说明 | 默认值 |
|------|------|------|------|--------|
| `group_type` | int | 否 | 分组类型：0-固定（分组内所有商品必选），1-可选（分组内商品可选择性） | 0 |
| `optional_count` | int | 否 | 可选数量，表示本组商品中要求选择多少个商品（仅当 `group_type=1` 时有效） | 1 |
| `add_price` | float | 否 | 加价金额，表示该商品需要加价多少钱 | 0.00 |

**验证规则**:

1. `group_type` 必须为 0 或 1
2. `optional_count` 必须 >= 1（当 `group_type=1` 时）
3. `optional_count` 不能大于分组内商品总数
4. `add_price` 必须 >= 0
5. 固定分组（`group_type=0`）时，`optional_count` 会自动设置为分组内商品总数

**响应格式**:

```json
{
  "code": 1,
  "message": "保存成功",
  "data": {}
}
```

**错误响应示例**:

```json
{
  "code": 0,
  "message": "可选数量不能大于分组内商品总数",
  "data": {}
}
```

---

### 2. 编辑套餐接口

**接口地址**: `POST /api/v1/shop/product/edit`

**请求参数变更**:

在 `package.groups[]` 数组中，每个分组对象新增以下字段（与创建接口相同）：

```json
{
  "uuid": 123456,
  "type": 1,
  "locale_name": {...},
  "package": {
    "price": 99.00,
    "internal_code": "PKG001",
    "groups": [
      {
        "uuid": 789012,               // 编辑时必填
        "locale_name": {
          "zh": "主餐",
          "th": "Main Course",
          "en": "Main Course"
        },
        "group_type": 1,              // ⭐ 新增：分组类型
        "optional_count": 2,          // ⭐ 新增：可选数量
        "products": [
          {
            "uuid": 345678,           // 编辑时必填
            "bom_uuid": 123456,
            "num": 1,
            "sort": 1,
            "add_price": 5.00,         // ⭐ 新增：加价金额
            "is_delete": false
          }
        ],
        "is_delete": false
      }
    ]
  }
}
```

**字段说明**: 与创建接口相同

**验证规则**: 与创建接口相同

**响应格式**: 与创建接口相同

---

### 3. 商品详情查询接口

**接口地址**: `GET /api/v1/shop/product/detail?uuid={商品UUID}`

**响应格式变更**:

商品详情接口主要用于查看商品信息。对于套餐商品，分组信息会在商品列表接口中返回（见下方说明）。

**注意**: 商品详情接口返回的 `package_sub_product_groups` 是套餐子商品分组列表，用于展示，不包含分组类型和可选数量字段。编辑套餐时请使用商品列表接口或创建/编辑接口返回的数据。

---

### 4. 商品列表查询接口（套餐分组信息）

**接口地址**: `GET /api/v1/shop/product/list`

**响应格式变更**:

在套餐商品的响应中，`package_group_list.list[]` 数组中每个分组对象新增以下字段：

```json
{
  "code": 1,
  "message": "success",
  "data": {
    "list": [
      {
        "uuid": 123456,
        "product_type": 1,
        "locale_name": {...},
        "package_group_list": {
          "list": [
            {
              "uuid": 789012,
              "locale_name": {
                "zh": "主餐",
                "th": "Main Course",
                "en": "Main Course"
              },
              "group_type": 1,              // ⭐ 新增：分组类型 0-固定 1-可选
              "optional_count": 2,          // ⭐ 新增：可选数量
              "is_full": false,
              "num": 3,
              "products": {
                "list": [
                  {
                    "detail": {...},
                    "num": 1,
                    "add_price": 5.00,     // ⭐ 新增：加价金额
                    "can_edit": true
                  }
                ]
              }
            }
          ]
        }
      }
    ]
  }
}
```

**字段说明**:

| 字段 | 类型 | 说明 |
|------|------|------|
| `group_type` | int | 分组类型：0-固定，1-可选 |
| `optional_count` | int | 可选数量，表示本组商品中要求选择多少个商品 |
| `add_price` | float | 加价金额，表示该商品需要加价多少钱 |

---

## 📝 使用示例

### 示例 1: 创建固定分组套餐

```json
{
  "type": 1,
  "locale_name": {
    "zh": "汉堡套餐",
    "th": "Burger Set",
    "en": "Burger Set"
  },
  "package": {
    "price": 59.00,
    "groups": [
      {
        "locale_name": {
          "zh": "主餐",
          "th": "Main Course"
        },
        "group_type": 0,              // 固定分组
        "optional_count": 1,           // 固定分组时，会自动设置为商品总数
        "products": [
          {
            "bom_uuid": 1001,
            "num": 1,
            "sort": 1,
            "add_price": 0.00
          }
        ]
      }
    ]
  }
}
```

### 示例 2: 创建可选分组套餐

```json
{
  "type": 1,
  "locale_name": {
    "zh": "套餐A",
    "th": "Set A"
  },
  "package": {
    "price": 99.00,
    "groups": [
      {
        "locale_name": {
          "zh": "饮料",
          "th": "Drink"
        },
        "group_type": 1,              // 可选分组
        "optional_count": 1,           // 3选1
        "products": [
          {
            "bom_uuid": 2001,
            "num": 1,
            "sort": 1,
            "add_price": 0.00          // 可乐，不加价
          },
          {
            "bom_uuid": 2002,
            "num": 1,
            "sort": 2,
            "add_price": 2.00          // 大杯可乐，加价2元
          },
          {
            "bom_uuid": 2003,
            "num": 1,
            "sort": 3,
            "add_price": 0.00          // 雪碧，不加价
          }
        ]
      },
      {
        "locale_name": {
          "zh": "小食",
          "th": "Snack"
        },
        "group_type": 1,              // 可选分组
        "optional_count": 2,           // 5选2
        "products": [
          {
            "bom_uuid": 3001,
            "num": 1,
            "sort": 1,
            "add_price": 0.00
          },
          {
            "bom_uuid": 3002,
            "num": 1,
            "sort": 2,
            "add_price": 0.00
          },
          {
            "bom_uuid": 3003,
            "num": 1,
            "sort": 3,
            "add_price": 0.00
          },
          {
            "bom_uuid": 3004,
            "num": 1,
            "sort": 4,
            "add_price": 0.00
          },
          {
            "bom_uuid": 3005,
            "num": 1,
            "sort": 5,
            "add_price": 0.00
          }
        ]
      }
    ]
  }
}
```

### 示例 3: 编辑套餐分组类型

```json
{
  "uuid": 123456,
  "type": 1,
  "package": {
    "groups": [
      {
        "uuid": 789012,
        "locale_name": {
          "zh": "饮料"
        },
        "group_type": 1,              // 从固定改为可选
        "optional_count": 1,
        "products": [...],
        "is_delete": false
      }
    ]
  }
}
```

---

## 🎨 前端界面建议

### 1. 分组类型选择器

- **固定分组**（`group_type=0`）：
  - 显示所有商品，全部必选
  - 隐藏"可选数量"输入框
  - 商品列表不可取消选择

- **可选分组**（`group_type=1`）：
  - 显示"可选数量"输入框
  - 商品列表可选择性
  - 显示"X选Y"提示（如"3选1"、"5选2"）

### 2. 可选数量输入框

- 仅在"可选分组"时显示
- 最小值：1
- 最大值：分组内商品总数
- 实时验证：不能大于商品总数

### 3. 加价金额输入框

- 每个商品显示加价输入框
- 默认值：0.00
- 支持小数，精度2位
- 最小值：0（不能为负数）

### 4. 验证提示

- **可选数量验证失败**：显示"可选数量不能大于分组内商品总数"
- **加价金额验证失败**：显示"加价金额不能为负数"
- **分组类型验证失败**：显示"分组类型必须为0（固定）或1（可选）"

---

## 🔍 字段映射表

### 请求字段映射

| 前端字段名 | API 字段名 | 类型 | 说明 |
|-----------|-----------|------|------|
| `groupType` | `group_type` | int | 分组类型 |
| `optionalCount` | `optional_count` | int | 可选数量 |
| `addPrice` | `add_price` | float | 加价金额 |

### 响应字段映射

| API 字段名 | 前端字段名 | 类型 | 说明 |
|-----------|-----------|------|------|
| `group_type` | `groupType` | int | 分组类型 |
| `optional_count` | `optionalCount` | int | 可选数量 |
| `add_price` | `addPrice` | float | 加价金额 |

---

## ⚠️ 注意事项

1. **向后兼容**：
   - 如果前端不传 `group_type` 和 `optional_count`，后端会使用默认值（`group_type=0`，`optional_count=1`）
   - 如果前端不传 `add_price`，后端会使用默认值 `0.00`
   - 现有套餐数据会自动设置为固定分组（`group_type=0`）

2. **数据验证**：
   - 所有验证在后端进行，前端也需要进行前端验证以提升用户体验
   - 固定分组时，`optional_count` 会自动设置为商品总数，前端无需手动设置

3. **价格计算**：
   - 加价金额会在套餐总价基础上累加
   - 前端需要正确计算和显示最终价格

4. **可选分组逻辑**：
   - 可选分组中，顾客需要选择指定数量的商品（`optional_count`）
   - 前端需要控制选择数量，确保不超过 `optional_count`

---

## 📚 相关文档

- 需求文档: `docs/shared/specs/active/story-shop-package-group-type/requirements.md`
- 设计文档: `docs/shared/specs/active/story-shop-package-group-type/design.md`
- API 文档: Swagger UI (`/apidoc/index.html`)

---

**版本**: v1.0.0  
**创建日期**: 2025-11-24  
**维护者**: 后端开发组

