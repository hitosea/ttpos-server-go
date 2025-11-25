# 套餐分组类型增强功能（必选和默认选中）- 前端对接文档

> 本文档说明后端新增的套餐分组类型增强功能（必选和默认选中），供前端开发人员对接使用。

**更新日期**: 2025-11-25  
**版本**: v1.0.0

---

## 📋 概述

后端在套餐分组类型和加价功能的基础上，新增了**必选**和**默认选中**两个属性。可选分组中的商品可以设置为"必选"（必须选择才能下单）或"默认选中"（用户选购时默认选中），从而引导顾客选择商家推荐的商品。

**功能特点**：
- 必选商品：必须选择才能下单，自动选中且不可取消
- 默认选中商品：默认选中但可以取消，引导顾客优先选择
- 仅适用于可选分组（`group_type=1`），固定分组不支持

---

## 🔄 API 变更

### 1. 创建套餐接口

**接口地址**: `POST /api/v1/shop/product/add`

**请求参数变更**:

在 `package.groups[].products[]` 数组中，每个商品对象新增以下字段：

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
          "zh": "饮料",
          "th": "Drink",
          "en": "Drink"
        },
        "group_type": 1,              // 可选分组
        "optional_count": 2,          // 3选2
        "products": [
          {
            "bom_uuid": 123456,
            "num": 1,
            "sort": 1,
            "add_price": 0.00,
            "is_required": 1,         // ⭐ 新增：必选 0-不必选 1-必选
            "is_default": 0           // ⭐ 新增：默认选中 0-默认不选中 1-默认选中
          },
          {
            "bom_uuid": 123457,
            "num": 1,
            "sort": 2,
            "add_price": 2.00,
            "is_required": 0,         // 不必选
            "is_default": 1           // 默认选中
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
| `is_required` | int | 否 | 必选：0-不必选，1-必选（仅可选分组有效） | 0 |
| `is_default` | int | 否 | 默认选中：0-不默认选中，1-默认选中（仅可选分组有效） | 0 |

**验证规则**:

1. `is_required` 必须为 0 或 1
2. `is_default` 必须为 0 或 1
3. **固定分组**（`group_type=0`）时，`is_required` 和 `is_default` 必须为0
4. **可选分组**（`group_type=1`）时：
   - 必选数量不能大于可选数量（`optional_count`）
   - 默认数量不能大于可选数量（`optional_count`）

**错误响应示例**:

```json
{
  "code": 0,
  "message": "必选数量不能大于可选数量",
  "data": {}
}
```

```json
{
  "code": 0,
  "message": "固定分组的必选必须为0",
  "data": {}
}
```

---

### 2. 编辑套餐接口

**接口地址**: `POST /api/v1/shop/product/edit`

**请求参数变更**:

在 `package.groups[].products[]` 数组中，每个商品对象新增以下字段（与创建接口相同）：

```json
{
  "uuid": 123456,
  "type": 1,
  "locale_name": {...},
  "package": {
    "price": 99.00,
    "groups": [
      {
        "uuid": 789012,
        "locale_name": {
          "zh": "饮料"
        },
        "group_type": 1,
        "optional_count": 2,
        "products": [
          {
            "uuid": 345678,
            "bom_uuid": 123456,
            "num": 1,
            "sort": 1,
            "add_price": 0.00,
            "is_required": 1,         // ⭐ 新增：必选
            "is_default": 0,          // ⭐ 新增：默认选中
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

---

### 3. 商品详情查询接口

**接口地址**: `GET /api/v1/shop/product/detail?uuid={商品UUID}`

**响应格式变更**:

在套餐商品的响应中，`package_sub_product_groups.list[]` 数组中每个分组对象新增以下字段：

```json
{
  "code": 1,
  "message": "success",
  "data": {
    "package_sub_product_groups": {
      "list": [
        {
          "uuid": 789012,
          "locale_name": {
            "zh": "饮料",
            "th": "Drink"
          },
          "group_type": 1,              // ⭐ 新增：分组类型 0-固定 1-可选
          "optional_count": 2,          // ⭐ 新增：可选数量
          "products": {
            "list": [
              {
                "uuid": 345678,
                "bom_uuid": 123456,
                "product_uuid": 111222,
                "num": 1,
                "price": 0,
                "is_required": 1,    // ⭐ 新增：必选
                "is_default": 0      // ⭐ 新增：默认选中
              }
            ]
          }
        }
      ]
    }
  }
}
```

**字段说明**:

| 字段 | 类型 | 说明 |
|------|------|------|
| `group_type` | int | 分组类型：0-固定，1-可选 |
| `optional_count` | int | 可选数量，表示本组商品中要求选择多少个商品 |
| `is_required` | int | 必选：0-不必选，1-必选 |
| `is_default` | int | 默认选中：0-不默认选中，1-默认选中 |

**注意**: 商品详情接口返回的 `package_sub_product_groups` 现在包含完整的分组信息（分组类型、可选数量、必选和默认选中），可以用于编辑套餐时的数据回显。

---

### 4. 商品列表查询接口（商家管理端）

**接口地址**: `GET /api/v1/shop/product/list`

**响应格式变更**:

在套餐商品的响应中，`package_group_list.list[].products.list[]` 数组中每个商品对象新增以下字段：

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
                "zh": "饮料",
                "th": "Drink"
              },
              "group_type": 1,
              "optional_count": 2,
              "products": {
                "list": [
                  {
                    "detail": {...},
                    "num": 1,
                    "add_price": 5.00,
                    "is_required": 1,    // ⭐ 新增：必选 0-不必选 1-必选
                    "is_default": 0,     // ⭐ 新增：默认选中 0-默认不选中 1-默认选中
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

---

### 5. 商品列表查询接口（收银端）

**接口地址**: `GET /api/v1/{terminal}/product/list`

**终端类型**: `pos`、`assistant`、`tablet`、`mobile`、`member` 等

**响应格式变更**:

在套餐商品的响应中，`package_group_list.list[].products.list[]` 数组中每个商品对象新增以下字段：

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
              "group_type": 1,
              "optional_count": 2,
              "is_full": false,
              "num": 3,
              "products": {
                "list": [
                  {
                    "detail": {
                      "uuid": 111222,
                      "locale_name": {...},
                      "image": "...",
                      "price": 0
                    },
                    "num": 1,
                    "add_price": 5.00,
                    "is_required": 1,    // ⭐ 新增：必选 0-不必选 1-必选
                    "is_default": 0,     // ⭐ 新增：默认选中 0-默认不选中 1-默认选中
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
| `is_required` | int | 必选：0-不必选，1-必选 |
| `is_default` | int | 默认选中：0-不默认选中，1-默认选中 |

---

## 📝 使用示例

### 示例 1: 创建包含必选商品的套餐

```json
{
  "type": 1,
  "locale_name": {
    "zh": "汉堡套餐",
    "th": "Burger Set"
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
        "optional_count": 1,
        "products": [
          {
            "bom_uuid": 1001,
            "num": 1,
            "sort": 1,
            "add_price": 0.00,
            "is_required": 0,         // 固定分组必须为0
            "is_default": 0          // 固定分组必须为0
          }
        ]
      },
      {
        "locale_name": {
          "zh": "饮料",
          "th": "Drink"
        },
        "group_type": 1,              // 可选分组
        "optional_count": 1,          // 3选1
        "products": [
          {
            "bom_uuid": 2001,
            "num": 1,
            "sort": 1,
            "add_price": 0.00,
            "is_required": 1,         // 必选：大杯可乐必须选择
            "is_default": 0
          },
          {
            "bom_uuid": 2002,
            "num": 1,
            "sort": 2,
            "add_price": 0.00,
            "is_required": 0,
            "is_default": 1          // 默认选中：中杯可乐默认选中
          },
          {
            "bom_uuid": 2003,
            "num": 1,
            "sort": 3,
            "add_price": 0.00,
            "is_required": 0,
            "is_default": 0
          }
        ]
      }
    ]
  }
}
```

### 示例 2: 创建包含默认选中商品的套餐

```json
{
  "type": 1,
  "locale_name": {
    "zh": "套餐A"
  },
  "package": {
    "price": 99.00,
    "groups": [
      {
        "locale_name": {
          "zh": "小食",
          "th": "Snack"
        },
        "group_type": 1,              // 可选分组
        "optional_count": 2,          // 5选2
        "products": [
          {
            "bom_uuid": 3001,
            "num": 1,
            "sort": 1,
            "add_price": 0.00,
            "is_required": 0,
            "is_default": 1          // 默认选中：薯条
          },
          {
            "bom_uuid": 3002,
            "num": 1,
            "sort": 2,
            "add_price": 0.00,
            "is_required": 0,
            "is_default": 1          // 默认选中：鸡块
          },
          {
            "bom_uuid": 3003,
            "num": 1,
            "sort": 3,
            "add_price": 0.00,
            "is_required": 0,
            "is_default": 0
          },
          {
            "bom_uuid": 3004,
            "num": 1,
            "sort": 4,
            "add_price": 0.00,
            "is_required": 0,
            "is_default": 0
          },
          {
            "bom_uuid": 3005,
            "num": 1,
            "sort": 5,
            "add_price": 0.00,
            "is_required": 0,
            "is_default": 0
          }
        ]
      }
    ]
  }
}
```

### 示例 3: 编辑套餐必选和默认选中属性

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
        "group_type": 1,
        "optional_count": 1,
        "products": [
          {
            "uuid": 345678,
            "bom_uuid": 123456,
            "num": 1,
            "sort": 1,
            "add_price": 0.00,
            "is_required": 1,         // 修改为必选
            "is_default": 0,
            "is_delete": false
          },
          {
            "uuid": 345679,
            "bom_uuid": 123457,
            "num": 1,
            "sort": 2,
            "add_price": 2.00,
            "is_required": 0,
            "is_default": 1,          // 修改为默认选中
            "is_delete": false
          }
        ],
        "is_delete": false
      }
    ]
  }
}
```

---

## 🎨 前端界面建议

### 1. 商家管理端界面

#### 1.1 必选和默认选中复选框

- **显示位置**：仅在可选分组（`group_type=1`）的商品配置中显示
- **固定分组**：隐藏必选和默认选中复选框（固定分组不支持）
- **可选分组**：显示必选和默认选中复选框

**界面布局建议**：

```
[可选分组商品配置]
├── 商品名称：可乐
├── 商品数量：1
├── 加价金额：0.00
├── ☑ 必选          ← 仅在可选分组时显示
└── ☐ 默认选中      ← 仅在可选分组时显示
```

#### 1.2 验证提示

- **必选数量验证**：
  - 实时统计：当前分组中 `is_required=1` 的商品数量
  - 验证规则：必选数量 <= 可选数量（`optional_count`）
  - 错误提示："必选数量不能大于可选数量，当前必选 {X} 个，可选数量为 {Y}"

- **默认数量验证**：
  - 实时统计：当前分组中 `is_default=1` 的商品数量
  - 验证规则：默认数量 <= 可选数量（`optional_count`）
  - 错误提示："默认数量不能大于可选数量，当前默认 {X} 个，可选数量为 {Y}"

- **固定分组限制**：
  - 当 `group_type=0` 时，禁用或隐藏必选和默认选中复选框
  - 如果用户尝试设置，提示："固定分组不支持必选和默认选中"

#### 1.3 交互逻辑

- **分组类型切换**：
  - 从"可选"切换到"固定"时，自动将所有商品的 `is_required` 和 `is_default` 设置为 0
  - 从"固定"切换到"可选"时，保持默认值（0）

- **必选和默认选中互斥**：
  - 建议：必选商品通常不需要设置为默认选中（因为必选商品会自动选中）
  - 允许同时设置：技术上允许，但业务上不推荐

---

### 2. 收银端界面（POS/助手/平板/会员/自助点餐机）

#### 2.1 必选商品显示

- **标识显示**：
  - 必选商品显示"必选"标签或图标
  - 建议使用醒目的颜色（如红色）标识必选商品

- **选择状态**：
  - 必选商品自动选中，且不可取消
  - 必选商品的数量至少为1份，不能减少到0

- **交互限制**：
  - 禁用取消选择按钮
  - 禁用减少数量按钮（当数量为1时）
  - 提示："必选商品不可删除"

**界面示例**：

```
[饮料分组 - 3选1]
├── ☑ 大杯可乐 [必选] ← 自动选中，不可取消
├── ☐ 中杯可乐 [默认] ← 默认选中，可以取消
└── ☐ 雪碧
```

#### 2.2 默认选中商品显示

- **标识显示**：
  - 默认选中商品显示"推荐"或"默认"标签
  - 建议使用温和的颜色（如蓝色）标识默认选中商品

- **选择状态**：
  - 默认选中商品默认选中，但可以取消
  - 用户可以自由增加或减少数量

- **交互逻辑**：
  - 初始状态：默认选中商品自动选中（数量为1）
  - 用户可以取消选择
  - 用户可以增加数量

**界面示例**：

```
[小食分组 - 5选2]
├── ☑ 薯条 [推荐] ← 默认选中，可以取消
├── ☑ 鸡块 [推荐] ← 默认选中，可以取消
├── ☐ 鸡翅
├── ☐ 洋葱圈
└── ☐ 玉米杯
```

#### 2.3 选择验证

- **必选商品验证**：
  - 当用户尝试删除必选商品时，提示："必选商品不可删除"
  - 当用户尝试将必选商品数量减少到0时，提示："必选商品至少需要选择1份"

- **分组选择验证**：
  - 可选分组：验证已选数量是否等于 `optional_count`
  - 必选商品：自动计入已选数量
  - 默认选中商品：如果选中，计入已选数量

**验证提示示例**：

```
[饮料分组 - 3选1，已选0份]
提示："该分组需要选择 1 个商品，当前已选 0 个，还差 1 个"

[饮料分组 - 3选1，已选1份（必选商品）]
提示：无（已选满）

[小食分组 - 5选2，已选1份]
提示："该分组需要选择 2 个商品，当前已选 1 个，还差 1 个"
```

---

## 🔍 字段映射表

### 请求字段映射

| 前端字段名 | API 字段名 | 类型 | 说明 |
|-----------|-----------|------|------|
| `isRequired` | `is_required` | int | 必选：0-不必选，1-必选 |
| `isDefault` | `is_default` | int | 默认选中：0-不默认选中，1-默认选中 |

### 响应字段映射

| API 字段名 | 前端字段名 | 类型 | 说明 |
|-----------|-----------|------|------|
| `is_required` | `isRequired` | int | 必选：0-不必选，1-必选 |
| `is_default` | `isDefault` | int | 默认选中：0-不默认选中，1-默认选中 |

---

## ⚠️ 注意事项

### 1. 向后兼容

- **默认值**：
  - 如果前端不传 `is_required` 和 `is_default`，后端会使用默认值（`is_required=0`，`is_default=0`）
  - 现有套餐数据会自动设置为不必选、不默认选中

- **固定分组**：
  - 固定分组（`group_type=0`）时，`is_required` 和 `is_default` 必须为0
  - 后端会验证并拒绝固定分组设置必选或默认选中

### 2. 数据验证

- **后端验证**：
  - 所有验证在后端进行，前端也需要进行前端验证以提升用户体验
  - 固定分组时，后端会自动将 `is_required` 和 `is_default` 设置为0

- **前端验证**：
  - 实时验证必选数量和默认数量
  - 分组类型切换时，自动重置必选和默认选中状态

### 3. 业务规则

- **必选商品**：
  - 必选商品必须选择才能下单
  - 必选商品在顾客选择时自动选中且不可取消
  - 必选商品必须保持至少1份

- **默认选中商品**：
  - 默认选中商品默认选中，但可以取消
  - 用户可以自由增加或减少数量
  - 默认选中商品不强制选择

- **数量限制**：
  - 必选数量不能大于可选数量
  - 默认数量不能大于可选数量
  - 必选数量 + 默认数量可以大于可选数量（但业务上不推荐）

### 4. 收银端选择逻辑

- **必选商品处理**：
  - 自动选中必选商品（数量为1）
  - 必选商品计入已选数量
  - 必选商品不能取消或删除

- **默认选中商品处理**：
  - 初始状态：默认选中商品自动选中（数量为1）
  - 用户可以取消选择
  - 如果选中，计入已选数量

- **选择验证**：
  - 可选分组：验证已选数量（包括必选和默认选中）是否等于 `optional_count`
  - 验证失败时提示："该分组需要选择 {optional_count} 个商品，当前已选 {实际数量} 个，还差 {差值} 个"

---

## 📚 相关文档

- 需求文档: `docs/shared/specs/story-shop-package-group-type-enhancement/requirements.md`
- 设计文档: `docs/shared/specs/story-shop-package-group-type-enhancement/design.md`
- 套餐分组类型功能文档: `docs/shared/api/frontend-changes-package-group-type.md`
- API 文档: Swagger UI (`/apidoc/index.html`)

---

## 🔄 变更历史

| 版本 | 日期 | 变更内容 |
|------|------|----------|
| v1.0.0 | 2025-11-25 | 初始版本，新增必选和默认选中功能 |

---

**版本**: v1.0.0  
**创建日期**: 2025-11-25  
**维护者**: 后端开发组

