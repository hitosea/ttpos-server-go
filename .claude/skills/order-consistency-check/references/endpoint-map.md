# 端点注册表 (Endpoint Map)

本文件定义所有可测试的 mutation/query 端点配对。Agent 执行测试时查阅此表获取请求模板。

## 购物车操作

### Mutation 端点（写操作）

| Key | Source | Path | Method | 说明 |
|-----|--------|------|--------|------|
| cashier-desk-add | cashier | /cashier/desk/order/cart/product/add | POST | 收银桌台加购 |
| cashier-desk-num | cashier | /cashier/desk/order/cart/product/num | POST | 收银桌台改数量 |
| cashier-instant-add | cashier | /cashier/instant/order/cart/product/add | POST | 收银点餐加购 |
| cashier-instant-num | cashier | /cashier/instant/order/cart/product/num | POST | 收银点餐改数量 |
| assistant-add | assistant | /assistant/desk/order/cart/product/add | POST | 助手端加购 |
| assistant-num | assistant | /assistant/desk/order/cart/product/num | POST | 助手端改数量 |
| h5-add | h5 | /h5/order/cart/product/add | POST | H5扫码加购 |
| h5-num | h5 | /h5/order/cart/product/num | POST | H5扫码改数量 |
| tablet-add | tablet | /tablet/desk/order/cart/product/add | POST | 平板加购 |
| tablet-add-cooking | tablet | /tablet/desk/order/cart/product/add_and_cooking | POST | 平板加购并送厨 |

### Query 端点（读操作，用于验证）

| Key | Source | Path | Method | Response Type | 说明 |
|-----|--------|------|--------|---------------|------|
| cashier-desk-query | cashier | /cashier/desk/order/cart/info | GET | ShopCart | 收银桌台购物车 |
| cashier-instant-query | cashier | /cashier/instant/order/cart/info | GET | ShopCart | 收银点餐购物车 |
| assistant-cart-query | assistant | /assistant/desk/order/cart/info | GET | ShopCart | 助手端购物车 |
| assistant-ping | assistant | /assistant/desk/ping | GET | DeskPing | 助手端轮询 |
| h5-ping | h5 | /h5/desk/ping | GET | H5DeskPing | H5轮询 |
| tablet-ping | tablet | /tablet/desk/ping | GET | DeskPing | 平板轮询 |

### Mutation → Query 配对

| Mutation | Query（同源验证） | Query（跨源验证） |
|----------|------------------|------------------|
| cashier-desk-add | cashier-desk-query | assistant-cart-query, h5-ping |
| cashier-desk-num | cashier-desk-query | assistant-cart-query, h5-ping |
| assistant-add | assistant-cart-query | cashier-desk-query, h5-ping |
| assistant-num | assistant-cart-query | cashier-desk-query, h5-ping |
| h5-add | h5-ping | cashier-desk-query, assistant-cart-query |
| h5-num | h5-ping | cashier-desk-query, assistant-cart-query |
| tablet-add | tablet-ping | cashier-desk-query |

## 请求体模板

### OrderCartProductAddReq (cashier / assistant)

```json
{
  "sale_bill_uuid": {{BILL}},
  "sale_order_uuid": {{ORDER}},
  "flavor_uuid": {{BOM_UUID}},
  "num": 1
}
```

字段说明：
- `sale_bill_uuid`: 销售账单 UUID（从 Phase 2 获取）
- `sale_order_uuid`: 销售订单 UUID（从 Phase 2 获取）
- `flavor_uuid`: 商品规格 UUID（即 `ttpos_product_bom.uuid`）
- `num`: 数量，默认 1

### OrderCartProductAddReq (h5)

```json
{
  "flavor_uuid": {{BOM_UUID}},
  "num": 1
}
```

> H5 不需要 `sale_bill_uuid` 和 `sale_order_uuid`，从 DeskToken 中的 desk_uuid 自动解析。

### OrderCartProductNumReq (cashier / assistant / h5)

```json
{
  "sale_bill_uuid": {{BILL}},
  "sale_order_uuid": {{ORDER}},
  "sale_order_product_uuid": {{PRODUCT_UUID}},
  "num": {{NEW_NUM}}
}
```

> H5 的 num 端点也需要 sale_bill_uuid（在 handler 中自动填充）。

### Query 参数

| Endpoint | 参数 |
|----------|------|
| cashier-desk-query | `?sale_bill_uuid={{BILL}}` |
| cashier-instant-query | `?sale_bill_uuid={{BILL}}` |
| assistant-cart-query | `?sale_bill_uuid={{BILL}}` |
| assistant-ping | `?desk_uuid={{DESK_UUID}}` |
| h5-ping | `?desk_uuid={{DESK_UUID}}` |
| tablet-ping | 无参数（从 JWT 的 device_uuid 关联桌台） |

## Token 生成命令

```bash
GEN_TOKEN="$(git rev-parse --show-toplevel)/tools/gen_token/gen_token"

# 收银端
CASHIER_TOKEN=$("$GEN_TOKEN" -source cashier -company $COMPANY 2>/dev/null)

# 助手端（自动检测设备绑定）
ASSISTANT_TOKEN=$("$GEN_TOKEN" -source assistant -company $COMPANY 2>/dev/null)

# 平板端
TABLET_TOKEN=$("$GEN_TOKEN" -source tablet -company $COMPANY 2>/dev/null)

# H5端（需要 desk-uuid）
H5_TOKEN=$("$GEN_TOKEN" -source h5 -company $COMPANY -desk-uuid $DESK_UUID 2>/dev/null)
```

## 端点特殊行为

### tablet-add-cooking
- 返回空响应 `{}`，不经过 `buildShopCartFromSaleBill`
- 验证需使用 tablet-ping 作为 query

### h5-add / h5-num
- mutation 内部先走 `buildShopCartFromSaleBill`，再由 handler 转换为 `H5DeskPing` 格式
- mutation 响应结构与 h5-ping query 响应结构相同

### assistant-add
- mutation 响应为 `ShopCart` 格式（与 cashier 相同）
- assistant-ping query 响应为 `DeskPing` 格式（比 ShopCart 多了 unsent_kitchen 等字段）
- 建议同源验证使用 assistant-cart-query（ShopCart 格式），跨源验证使用 assistant-ping
