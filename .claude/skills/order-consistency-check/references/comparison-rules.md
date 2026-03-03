# 响应对比规则 (Comparison Rules)

本文件定义各种响应类型的可比字段提取规则和跨端归一化逻辑。

## 响应类型

### 1. ShopCart

**来源**: cashier mutation/query, assistant mutation/query

**Go 结构**: `resp.ShopCart` (`main/app/dto/resp/shop_cart.go:159`)

**可比字段提取** (JSON path):

```
顶层:
  data.sale_bill_uuid           → 账单 UUID
  data.update_time              → 更新时间戳

订单级 (data.sale_order_list[]):
  [i].uuid                      → 子单 UUID
  [i].product_num               → 商品总数量（关键对比字段）
  [i].amount_info               → 金额信息（关键对比字段）
  [i].product_list              → 商品列表

商品级 (data.sale_order_list[i].product_list[]):
  [j].uuid                      → 商品 UUID（唯一标识）
  [j].num                       → 数量
  [j].price                     → 小计金额 (num × unit_price)
  [j].product_price             → 商品单价
  [j].sale_price                → 销售单价
  [j].status                    → 状态 (0=未送厨, 1=已送厨)
  [j].total_price               → 总价
  [j].discount_fee              → 折扣金额
  [j].product_package_uuid      → 商品包 UUID

金额信息 (data.sale_order_list[i].amount_info):
  .product_amount               → 商品总金额
  .discount_amount              → 折扣总金额
  .receivable_amount            → 应收金额
  .service_fee                  → 服务费
  .tax_fee                      → 税费
```

### 2. DeskPing

**来源**: assistant desk/ping, tablet desk/ping

**Go 结构**: `resp.DeskPing` (`main/app/dto/resp/desk.go:102`)

**与 ShopCart 共有字段**:
- `data.sale_order_list` → 结构完全相同
- `data.update_time`

**DeskPing 额外字段**:
```
data.unsent_kitchen             → 未送厨信息 (UnsentKitchen)
  .products.list[]              → 未送厨商品列表
  .amount_info                  → 未送厨金额
    .product_num                → 未送厨商品数量
    .product_amount             → 未送厨商品金额

data.sent_kitchen               → 已送厨信息 (SentKitchen)
  .groups.list[]                → 已送厨分组列表
    [k].send_kitchen_time       → 送厨时间
    [k].products.list[]         → 该组商品列表

data.desk_info                  → 桌台信息
  .uuid                         → 桌台 UUID
  .desk_no                      → 桌台编号
  .meal_num                     → 就餐人数
```

**对比策略**: 当 mutation 返回 ShopCart、query 返回 DeskPing 时，使用 `sale_order_list` 作为公共比较基准。

### 3. H5DeskPing

**来源**: h5 mutation/query (h5-add, h5-num, h5-ping)

**Go 结构**: `resp.H5DeskPing` (`main/app/dto/resp/desk.go:117`)

**可比字段提取**:
```
未送厨 (data.unsent_kitchen):
  .products.list[]              → 未下单商品列表
    [j].uuid                    → 商品 UUID
    [j].num                     → 数量
    [j].price                   → 金额
    [j].locale_name             → 多语言名称
  .amount_info                  → 金额信息
    .product_num                → 未下单商品总数量（关键对比字段）
    .product_amount             → 未下单商品总金额（关键对比字段）

已送厨 (data.sent_kitchen):
  .groups.list[]                → 已下单分组列表
    [k].products.list[]         → 该组商品列表
    [k].accept_time             → 接单时间
    [k].is_accept               → 是否接单

通用:
  data.update_time              → 更新时间
  data.desk_info                → 桌台信息
```

> **注意**: H5DeskPing 没有 `sale_order_list`。H5 的商品视图按 unsent/sent 分组，而非按 SaleOrder 分组。

## 同源对比规则

同一端点的 mutation 响应与 query 响应对比。

### ShopCart vs ShopCart (cashier, assistant)

```python
def compare_shopcart(mutation, query):
    """
    mutation: POST /cashier/desk/order/cart/product/add 响应
    query:    GET  /cashier/desk/order/cart/info 响应
    """
    fields_to_compare = []

    # 1. 商品 UUID 集合
    m_uuids = {p['uuid'] for so in mutation['sale_order_list'] for p in so['product_list']}
    q_uuids = {p['uuid'] for so in query['sale_order_list'] for p in so['product_list']}
    if m_uuids != q_uuids:
        fields_to_compare.append(('product_uuids', m_uuids, q_uuids))

    # 2. 逐商品对比
    for uuid in m_uuids & q_uuids:
        m_prod = find_product(mutation, uuid)
        q_prod = find_product(query, uuid)
        for key in ['num', 'price', 'status', 'product_price', 'sale_price',
                     'total_price', 'discount_fee']:
            if str(m_prod.get(key)) != str(q_prod.get(key)):
                fields_to_compare.append((f'product.{uuid}.{key}', m_prod[key], q_prod[key]))

    # 3. SaleOrder 级别对比
    for i, (m_so, q_so) in enumerate(zip(mutation['sale_order_list'], query['sale_order_list'])):
        if m_so.get('product_num') != q_so.get('product_num'):
            fields_to_compare.append((f'sale_order[{i}].product_num', ...))
        # amount_info 逐字段对比
        for key in ['product_amount', 'discount_amount', 'receivable_amount']:
            ...

    return fields_to_compare  # 空列表 = PASS
```

### H5DeskPing vs H5DeskPing (h5)

```python
def compare_h5(mutation, query):
    """
    mutation: POST /h5/order/cart/product/add 响应 (经 GetH5DeskPing 转换)
    query:    GET  /h5/desk/ping 响应
    """
    # 1. unsent_kitchen 对比
    m_unsent = mutation.get('unsent_kitchen', {})
    q_unsent = query.get('unsent_kitchen', {})

    m_ai = m_unsent.get('amount_info', {})
    q_ai = q_unsent.get('amount_info', {})

    for key in ['product_num', 'product_amount']:
        if m_ai.get(key) != q_ai.get(key):
            fields.append((f'unsent_kitchen.amount_info.{key}', ...))

    # 2. 未送厨商品列表对比
    m_prods = m_unsent.get('products', {}).get('list', [])
    q_prods = q_unsent.get('products', {}).get('list', [])

    m_uuids = {p['uuid'] for p in m_prods}
    q_uuids = {p['uuid'] for p in q_prods}
    if m_uuids != q_uuids:
        fields.append(('unsent_products_uuids', ...))

    # 3. 逐商品 num/price 对比
    ...
```

### DeskPing vs DeskPing (tablet)

```python
def compare_deskping(mutation, query):
    """与 ShopCart vs ShopCart 类似，使用 sale_order_list 对比"""
    # sale_order_list 结构相同
    return compare_shopcart_fields(mutation, query)
```

## 跨端对比规则

不同端点之间的数据一致性验证。

### 核心原则：商品可见性隔离

```
┌─────────────────────────────────────────────────┐
│ DB: ttpos_sale_order_product                     │
│                                                  │
│  is_accept_order = 1 (已接受)                     │
│  → 对 cashier, assistant, tablet 可见             │
│  → 对 H5 不可见（属于其他端的商品）                  │
│                                                  │
│  is_accept_order = 0 (H5未确认)                   │
│  → 仅对 H5 可见（在 unsent_kitchen 中显示）         │
│  → 对 cashier, assistant, tablet 不可见            │
└─────────────────────────────────────────────────┘
```

### 规则 1: Cashier 加购 → H5 验证

**预期**: H5 的 `unsent_kitchen` 不应包含 cashier 新加的商品。

```python
def cross_cashier_to_h5(cashier_mutation, h5_query):
    # cashier 新加的商品 UUID
    new_product_uuid = cashier_mutation['new_product_uuid']

    # H5 unsent 商品列表不应包含该 UUID
    h5_unsent_uuids = extract_h5_unsent_uuids(h5_query)
    assert new_product_uuid not in h5_unsent_uuids
```

### 规则 2: H5 加购 → Cashier 验证

**预期**: Cashier 的 `product_list` 不应包含 H5 未确认的商品。

```python
def cross_h5_to_cashier(h5_mutation, cashier_query):
    # H5 新加的商品在 cashier 视图中不应出现
    cashier_uuids = extract_shopcart_uuids(cashier_query)
    h5_new_uuids = extract_h5_new_uuids(h5_mutation)
    assert h5_new_uuids.isdisjoint(cashier_uuids)
```

### 规则 3: Cashier 加购 → Assistant 验证

**预期**: Assistant 看到的商品列表应包含 cashier 新加的商品。

```python
def cross_cashier_to_assistant(cashier_mutation, assistant_query):
    # assistant 使用 /assistant/desk/order/cart/info (ShopCart 格式)
    # 应与 cashier 的商品列表完全一致
    cashier_products = extract_shopcart_products(cashier_mutation)
    assistant_products = extract_shopcart_products(assistant_query)
    compare_product_lists(cashier_products, assistant_products)
```

### 规则 4: Assistant 加购 → Cashier 验证

**预期**: 与规则 3 对称。Cashier 应能看到 assistant 新加的商品。

### 规则 5: 任一端改数量 → 其他同类端验证

**预期**: 改数量后所有 accepted 端（cashier/assistant/tablet）看到的商品数量和金额应一致。

```python
def cross_num_change(source_mutation, *other_queries):
    changed_uuid = source_mutation['changed_product_uuid']
    new_num = source_mutation['new_num']

    for query in other_queries:
        product = find_product_in_response(query, changed_uuid)
        assert product['num'] == new_num
```

## 并发对比规则

### 最终状态一致性

并发 mutation 完成后，所有端 query 的最终状态应满足：

```python
def verify_concurrent_final_state(endpoint_queries):
    """
    endpoint_queries: dict of {source: query_response}
    """
    # 1. 所有 accepted 端（cashier/assistant/tablet）看到相同的商品集
    accepted_views = []
    for source in ['cashier', 'assistant', 'tablet']:
        if source in endpoint_queries:
            products = extract_accepted_products(endpoint_queries[source])
            accepted_views.append(set(p['uuid'] for p in products))

    # 所有 accepted 端的商品 UUID 集合应完全相同
    if len(accepted_views) > 1:
        assert all(v == accepted_views[0] for v in accepted_views)

    # 2. H5 只看到自己的未确认商品
    if 'h5' in endpoint_queries:
        h5_unsent_uuids = extract_h5_unsent_uuids(endpoint_queries['h5'])
        for v in accepted_views:
            assert h5_unsent_uuids.isdisjoint(v)

    # 3. 无重复商品 UUID（每次 add 应创建新 UUID）
    all_uuids = []
    for products in accepted_views:
        all_uuids.extend(products)
    assert len(all_uuids) == len(set(all_uuids))
```

### 串行化验证

由于分布式锁，并发请求实际串行执行。验证点：

- 每个 mutation 响应包含的商品数应递增（先完成的看到更少商品）
- 最后完成的 mutation 响应应与最终 query 一致

## 对比忽略字段

以下字段在对比时应忽略（允许不一致）：

| 字段 | 原因 |
|------|------|
| `update_time` | mutation 和 query 的时间戳可能微秒级不同 |
| `desk.duration` | 随时间变化 |
| `desk.start_time` | mutation 和 query 间可能跨秒 |
| `buffet.remaining_seconds` | 随时间变化 |
| `buffet.remaining_ordering_time` | 随时间变化 |
| `must_plans` | 状态可能因加购而变化 |
| `product.locale_name` | 格式可能略有不同（dict vs string） |

## 判定标准

| 结果 | 条件 |
|------|------|
| **PASS** | 所有可比字段完全一致，无 mismatch |
| **WARN** | 仅忽略字段存在差异 |
| **FAIL** | 关键字段（uuid/num/price/status/amount_info）存在差异 |
