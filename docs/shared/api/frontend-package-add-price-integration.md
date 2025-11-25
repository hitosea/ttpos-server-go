# 套餐加价功能 - 前端对接接口文档

> 本文档说明套餐加价功能的前端对接变更，包括请求参数、响应字段和验证逻辑。

**更新日期**: 2025-11-25  
**版本**: v1.0.0  
**影响范围**: POS、Assistant、Tablet、Mobile、Member 终端

---

## 📋 概述

后端新增了套餐商品加价功能，支持为套餐子商品设置加价金额，并在加购时进行分组选择验证。该功能已集成到套餐加购接口中。

### 核心功能

1. **加价金额传递**: 前端在加购套餐时，需要传递每个子商品的 `add_price` 参数
2. **分组选择验证**: 后端会自动验证套餐分组选择是否符合套餐设置（固定分组必须全选，可选分组必须选择指定数量）
3. **价格计算**: 加价金额会自动计入商品价格计算

---

## 🔄 API 变更

### 1. 套餐加购接口

#### 1.1 POS 收银端

**接口地址**: `POST /api/v1/cashier/desk/order/cart/product_package/add`

**接口地址**: `POST /api/v1/cashier/instant/order/cart/product_package/add`

#### 1.2 点餐助手端

**接口地址**: `POST /api/v1/assistant/desk/order/cart/product_package/add`

#### 1.3 平板端

**接口地址**: `POST /api/v1/tablet/desk/order/cart/product_package/add`

#### 1.4 Mobile 扫码端

**接口地址**: `POST /api/v1/h5/order/cart/product_package/add`

#### 1.5 Member 会员端

**接口地址**: `POST /api/v1/member/order/cart/product_package/add`

---

## 📥 请求参数变更

### 请求体结构

```json
{
  "sale_bill_uuid": 1234567890,
  "sale_order_uuid": 1234567891,
  "product_package_uuid": 1234567892,
  "products": [
    {
      "product_package_group_uuid": 1234567893,
      "flavor_uuid": 1234567894,
      "sauce_uuid": [],
      "attribute_uuid": [],
      "num": 1.0,
      "unit_num": 1.0,
      "add_price": 5.00  // ⭐ 新增：加价金额
    }
  ]
}
```

### 字段说明

| 字段 | 类型 | 必填 | 说明 | 默认值 |
|------|------|------|------|--------|
| `sale_bill_uuid` | uint64 | 否 | 销售账单UUID，不填时自动创建 | 0 |
| `sale_order_uuid` | uint64 | 否 | 销售订单UUID | 0 |
| `product_package_uuid` | uint64 | 是 | 套餐UUID | - |
| `products` | array | 是 | 套餐子商品列表 | - |
| `products[].product_package_group_uuid` | uint64 | 是 | 套餐分组UUID | - |
| `products[].flavor_uuid` | uint64 | 是 | 商品规格UUID | - |
| `products[].sauce_uuid` | array | 否 | 小料UUID列表 | [] |
| `products[].attribute_uuid` | array | 否 | 属性UUID列表 | [] |
| `products[].num` | float64 | 是 | 商品数量 | - |
| `products[].unit_num` | float64 | 否 | 单位数量（套餐子商品） | 1.0 |
| `products[].add_price` | float64 | ⭐ 新增 | 加价金额，表示该子商品需要加价多少钱 | 0.00 |

### 重要说明

1. **`add_price` 字段**：
   - 必须从套餐详情接口获取，对应 `product_package_group_item.add_price`
   - 如果商品没有加价，传 `0.00` 或不传（后端默认 `0.00`）
   - 支持小数，精度为 4 位小数

2. **分组选择要求**：
   - **固定分组**（`group_type = 0`）：必须选择分组内所有商品
   - **可选分组**（`group_type = 1`）：必须选择 `optional_count` 个商品

---

## 📤 响应字段变更

### 响应体结构

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "sale_bill_uuid": 1234567890,
    "desk": {...},
    "sale_order_list": [
      {
        "uuid": 1234567891,
        "product_list": [
          {
            "uuid": 1234567895,
            "product_type": 1,
            "add_price": 10.00,  // ⭐ 新增：套餐主商品的加价总和
            "package_product_list": {
              "list": [
                {
                  "uuid": 1234567896,
                  "locale_name": {...},
                  "num": 1.0,
                  "unit_num": 1.0,
                  "add_price": 5.00  // ⭐ 新增：子商品的加价金额
                }
              ]
            }
          }
        ]
      }
    ]
  }
}
```

### 字段说明

#### Product（套餐主商品）

| 字段 | 类型 | 说明 |
|------|------|------|
| `add_price` | float64 | ⭐ 新增：加价金额（套餐主商品的加价总和 = Σ(子商品加价 × 子商品数量)） |

#### PackageProduct（套餐子商品）

| 字段 | 类型 | 说明 |
|------|------|------|
| `add_price` | float64 | ⭐ 新增：加价金额（单个子商品的加价金额） |

---

## ✅ 验证逻辑

### 1. 固定分组验证

**规则**: 固定分组（`group_type = 0`）必须选择分组内所有商品

**错误示例**:
```json
{
  "code": 400,
  "message": "固定分组「主餐」必须选择所有商品，请重新选择"
}
```

**前端处理**:
- 在套餐选择界面，固定分组应禁用部分商品选择
- 或提示用户必须选择所有商品

### 2. 可选分组验证

**规则**: 可选分组（`group_type = 1`）必须选择 `optional_count` 个商品

**错误示例**:
```json
{
  "code": 400,
  "message": "该分组「配菜」需要选择 2 个商品，当前已选 1 个，还差 1 个"
}
```

或

```json
{
  "code": 400,
  "message": "该分组「配菜」最多选择 2 个商品，当前已选 3 个，请删除多余商品"
}
```

**前端处理**:
- 在套餐选择界面，显示可选数量提示（如"请选择 2 个"）
- 选择数量不足时，禁用"确定"按钮
- 选择数量超过时，提示用户删除多余商品

---

## 💡 前端实现建议

### 1. 获取套餐详情

在加购套餐前，先调用套餐详情接口获取分组信息和加价金额：

**接口**: `GET /api/v1/{terminal}/order/product/package/detail`

**响应示例**:
```json
{
  "code": 200,
  "data": {
    "uuid": 1234567892,
    "locale_name": {...},
    "groups": [
      {
        "uuid": 1234567893,
        "locale_name": {...},
        "group_type": 0,  // 0-固定 1-可选
        "optional_count": 1,  // 可选数量
        "items": [
          {
            "bom_uuid": 1234567894,
            "locale_name": {...},
            "add_price": 5.00  // ⭐ 加价金额
          }
        ]
      }
    ]
  }
}
```

### 2. 构建请求参数

```javascript
// 示例：构建套餐加购请求
function buildPackageAddRequest(packageDetail, selectedItems) {
  const products = selectedItems.map(item => ({
    product_package_group_uuid: item.groupUuid,
    flavor_uuid: item.bomUuid,
    sauce_uuid: item.sauceUuidList || [],
    attribute_uuid: item.attributeUuidList || [],
    num: item.num || 1.0,
    unit_num: item.unitNum || 1.0,
    add_price: item.addPrice || 0.00  // ⭐ 从套餐详情中获取
  }));

  return {
    product_package_uuid: packageDetail.uuid,
    products: products
  };
}
```

### 3. 验证分组选择

```javascript
// 示例：验证分组选择是否符合要求
function validateGroupSelection(packageDetail, selectedItems) {
  const groupMap = new Map();
  
  // 按分组统计已选商品
  selectedItems.forEach(item => {
    const groupUuid = item.groupUuid;
    if (!groupMap.has(groupUuid)) {
      groupMap.set(groupUuid, []);
    }
    groupMap.get(groupUuid).push(item);
  });

  // 验证每个分组
  for (const group of packageDetail.groups) {
    const selected = groupMap.get(group.uuid) || [];
    const selectedCount = selected.reduce((sum, item) => sum + (item.num || 1), 0);

    if (group.group_type === 0) {
      // 固定分组：必须选择所有商品
      const validItems = group.items.filter(item => !item.is_delete);
      if (selected.length !== validItems.length) {
        return {
          valid: false,
          message: `固定分组「${group.locale_name.zh}」必须选择所有商品，请重新选择`
        };
      }
    } else {
      // 可选分组：必须选择 optional_count 个
      if (selectedCount !== group.optional_count) {
        const diff = group.optional_count - selectedCount;
        return {
          valid: false,
          message: diff > 0
            ? `该分组「${group.locale_name.zh}」需要选择 ${group.optional_count} 个商品，当前已选 ${selectedCount} 个，还差 ${diff} 个`
            : `该分组「${group.locale_name.zh}」最多选择 ${group.optional_count} 个商品，当前已选 ${selectedCount} 个，请删除多余商品`
        };
      }
    }
  }

  return { valid: true };
}
```

### 4. 显示加价金额

```javascript
// 示例：在商品列表中显示加价金额
function renderProductItem(product) {
  const addPrice = product.add_price || 0;
  const displayPrice = product.unit_price + addPrice;  // 单价 + 加价
  
  return `
    <div class="product-item">
      <span class="product-name">${product.locale_name.zh}</span>
      <span class="product-price">
        ¥${displayPrice.toFixed(2)}
        ${addPrice > 0 ? `<span class="add-price">(+¥${addPrice.toFixed(2)})</span>` : ''}
      </span>
    </div>
  `;
}
```

---

## 🚨 错误处理

### 常见错误码

| 错误码 | 说明 | 处理建议 |
|--------|------|----------|
| 400 | 分组选择不符合要求 | 显示错误提示，引导用户重新选择 |
| 400 | 参数验证失败 | 检查请求参数格式 |
| 404 | 套餐不存在 | 提示套餐已下架或不存在 |
| 500 | 服务器错误 | 提示系统错误，稍后重试 |

### 错误响应示例

```json
{
  "code": 400,
  "message": "固定分组「主餐」必须选择所有商品，请重新选择",
  "data": null
}
```

---

## 📝 兼容性说明

### 向后兼容

- `add_price` 字段为可选参数，不传时默认为 `0.00`
- 旧版本前端不传 `add_price` 时，接口仍可正常调用
- 响应中的 `add_price` 字段始终存在，旧版本前端可忽略该字段

### 升级建议

1. **立即升级**（推荐）：
   - 更新套餐详情接口调用，获取 `add_price` 字段
   - 在加购请求中传递 `add_price` 参数
   - 更新 UI 显示加价金额

2. **渐进式升级**：
   - 先更新请求参数，传递 `add_price`（从套餐详情获取，无加价时传 `0.00`）
   - 后续再更新 UI 显示加价金额

---

## 🔗 相关接口

### 1. 获取套餐详情

**接口**: `GET /api/v1/{terminal}/order/product/package/detail`

**参数**:
- `sale_bill_uuid` (可选): 销售账单UUID
- `sale_order_uuid` (可选): 销售订单UUID
- `product_package_uuid` (必填): 套餐UUID

**响应**: 包含分组信息、商品列表和加价金额

### 2. 获取购物车详情

**接口**: `GET /api/v1/{terminal}/desk/order/cart/info`

**响应**: 包含购物车商品列表，其中套餐商品包含 `add_price` 字段

---

## 📚 参考文档

- [套餐加购业务逻辑文档](./package-add-to-cart-business-logic.md)
- [套餐分组类型和加价功能文档](./frontend-changes-package-group-type.md)
- [功能规格文档](../specs/story-pos-package-selection/requirements.md)

---

## 📞 技术支持

如有问题，请联系后端开发团队或查看相关代码：

- 服务层实现: `main/app/service/order_product.go`
- 请求 DTO: `main/app/dto/req/shop_cart.go`
- 响应 DTO: `main/app/dto/resp/shop_cart.go`

---

**最后更新**: 2025-11-25  
**文档维护**: TTPOS Backend Team

