---
name: order-consistency-check
description: 订单一致性验证。对 TTPOS 多终端（收银/助手/H5/平板）订单 API 执行单端验证、跨端一致性和并发安全测试。当用户使用 /order-consistency-check 命令或提到"订单一致性"、"多端测试"、"跨端验证"、"回归测试"时触发。
---

# 订单一致性验证

## 触发条件

- 用户使用 `/order-consistency-check` 命令
- 用户提到"订单一致性"、"多端测试"、"跨端验证"、"回归测试"、"跨端测试"
- 修改共享 Service 方法（如 `buildShopCartFromSaleBill`）后需要验证

## 测试原理

**核心验证逻辑**：mutation 端点（如加购）的响应数据 vs query 端点（如查询购物车）的响应数据，二者应完全一致。

```
mutation 端点返回 (buildShopCartFromSaleBill 内存构建)
        ↕ 对比
query 端点返回 (GetOrderCartInfo DB 查询)
```

详细的端点配对和对比规则见 [references/](./references/) 目录。

## 执行流程

### Phase 1: 测试配置

使用 `AskUserQuestion` 收集测试参数（最多 2 轮）。

#### Round 1: 模式 & 端点

```yaml
Question 1: "选择回归测试模式"
Header: "测试模式"
MultiSelect: false
Options:
  - label: "单端验证"
    description: "一个端点执行 mutation，验证响应与 query 端点是否一致"
  - label: "跨端交互"
    description: "端点 A 执行 mutation → 端点 B 查询，验证跨端数据一致性"
  - label: "并发安全"
    description: "多端同时执行 mutation，验证所有响应和最终状态一致"
  - label: "完整套件"
    description: "依次执行全部三种模式"

Question 2: "选择参与测试的端点"
Header: "测试端点"
MultiSelect: true
Options:
  - label: "收银桌台 (cashier)"
    description: "POST /cashier/desk/order/cart/product/add → GET /cashier/desk/order/cart/info"
  - label: "助手端 (assistant)"
    description: "POST /assistant/desk/order/cart/product/add → GET /assistant/desk/order/cart/info"
  - label: "H5扫码 (h5)"
    description: "POST /h5/order/cart/product/add → GET /h5/desk/ping"
  - label: "平板端 (tablet)"
    description: "POST /tablet/desk/order/cart/product/add → GET /tablet/desk/ping"
```

#### Round 2: 商户 & 环境

**先从数据库动态查询商户列表**，再以选项形式呈现给用户：

```sql
-- Agent 先执行此查询，将结果填入 AskUserQuestion 的选项中
source "$(git rev-parse --show-toplevel)/.env"
mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USERNAME" -p"$DB_PASSWORD" saas -e "
  SELECT uuid, name FROM ttpos_company WHERE delete_time = 0 AND status = 1 ORDER BY id DESC LIMIT 8;
"
```

```yaml
Question 1: "选择测试商户"
Header: "商户"
MultiSelect: false
Options:
  # 从上述 SQL 查询结果动态生成，格式：
  - label: "{name}"
    description: "company_uuid: {uuid}"
  - label: "{name}"
    description: "company_uuid: {uuid}"
  # ... 用户也可选 "Other" 手动输入 company_uuid

Question 2: "测试环境"
Header: "环境"
MultiSelect: false
Options:
  - label: "本地 dev (localhost:8080)"
    description: "连接本地开发服务器，自动选择空闲桌台"
  - label: "自定义"
    description: "手动指定服务器地址和桌台"
```

默认参数：
- HOST: `http://localhost:8080`
- COMPANY: 用户选择的商户 UUID
- DB: 从 `.env` 文件读取
- DESK: 自动从商户 DB 查询空闲桌台

### Phase 2: 环境搭建

按以下步骤搭建测试环境，**每步失败则停止并报告错误**。

#### 2.1 构建 Token 生成工具

```bash
PROJECT_ROOT="$(git rev-parse --show-toplevel)"
cd "$PROJECT_ROOT/tools/gen_token" && go build -o gen_token .
```

#### 2.2 读取数据库配置

```bash
source "$PROJECT_ROOT/.env"
# DB_HOST, DB_PORT, DB_USERNAME, DB_PASSWORD 来自 .env
# 商户数据库: shop${COMPANY}
```

#### 2.3 验证服务器连通性

```bash
curl -sf "$HOST/api/v1/cashier/desk/list" \
  -H "Authorization: Bearer $(./gen_token -source cashier -company $COMPANY 2>/dev/null)" \
  > /dev/null
```

失败时提示检查服务器是否启动。

#### 2.4 查询有效 Staff

gen_token 的默认 staff 可能不在目标商户中，需要先查询：

```sql
SELECT uuid FROM ttpos_staff WHERE delete_time = 0 LIMIT 1;
```

后续 gen_token 命令加上 `-staff $STAFF_UUID`。

#### 2.5 桌台准备

**查询活跃桌台**：

```sql
SELECT d.uuid, d.desk_no, sb.uuid as sale_bill_uuid, so.uuid as sale_order_uuid
FROM ttpos_desk d
JOIN ttpos_sale_bill sb ON sb.desk_uuid = d.uuid AND sb.status = 0 AND sb.delete_time = 0
JOIN ttpos_sale_order so ON so.sale_bill_uuid = sb.uuid AND so.delete_time = 0
WHERE d.delete_time = 0
ORDER BY sb.create_time DESC
LIMIT 1;
```

**无活跃桌台时**：通过 cashier API 开桌：

```bash
curl -s -X POST "$HOST/api/v1/cashier/desk/open" \
  -H "Authorization: Bearer $CASHIER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"desk_uuid": $DESK_UUID, "meal_num": 2}'
```

#### 2.5 生成各端 Token

根据选择的测试端点生成对应 Token：

```bash
GEN_TOKEN="$PROJECT_ROOT/tools/gen_token/gen_token"

# 收银端
CASHIER_TOKEN=$("$GEN_TOKEN" -source cashier -company $COMPANY 2>/dev/null)

# 助手端（自动绑定收银设备）
ASSISTANT_TOKEN=$("$GEN_TOKEN" -source assistant -company $COMPANY 2>/dev/null)

# 平板端
TABLET_TOKEN=$("$GEN_TOKEN" -source tablet -company $COMPANY 2>/dev/null)

# H5端（需要 desk_uuid）
H5_TOKEN=$("$GEN_TOKEN" -source h5 -company $COMPANY -desk-uuid $DESK_UUID 2>/dev/null)
```

> **注意**：H5 使用 DeskToken 格式（非标准 JWT），assistant 的 JWT 包含双重设备绑定。详见 `tools/gen_token/main.go`。

#### 2.6 查询测试商品

```sql
SELECT pb.uuid as bom_uuid, pb.price, SUBSTRING(pf.name, 1, 60) as flavor_name
FROM ttpos_product_bom pb
JOIN ttpos_product_flavor pf ON pb.product_flavor_uuid = pf.uuid
WHERE pb.delete_time = 0 AND pf.delete_time = 0
ORDER BY pb.uuid
LIMIT 5;
```

选择 2-3 个不同价格的商品用于测试。

#### 2.7 保存环境变量

将所有环境信息保存到临时文件，供后续 Phase 复用：

```bash
cat > /tmp/ttpos_regression_env.sh << EOF
export HOST="$HOST"
export COMPANY="$COMPANY"
export DESK_UUID=$DESK_UUID
export BILL=$SALE_BILL_UUID
export ORDER=$SALE_ORDER_UUID
export CASHIER_TOKEN="$CASHIER_TOKEN"
export ASSISTANT_TOKEN="$ASSISTANT_TOKEN"
export TABLET_TOKEN="$TABLET_TOKEN"
export H5_TOKEN="$H5_TOKEN"
export TEST_PRODUCT_1=$BOM_UUID_1
export TEST_PRODUCT_2=$BOM_UUID_2
EOF
```

#### 2.8 平板端桌台绑定（如需测试平板）

平板通过 `ttpos_desk.device_uuid` 字段绑定桌台。需要：

1. 从 token 获取平板设备的 device_uuid
2. 将测试桌台的 `device_uuid` 更新为平板设备 UUID
3. **测试结束后务必恢复原值**

```sql
-- 绑定
UPDATE ttpos_desk SET device_uuid = $TABLET_DEVICE_UUID WHERE uuid = $DESK_UUID;
-- 测试后恢复
UPDATE ttpos_desk SET device_uuid = 0 WHERE uuid = $DESK_UUID;
```

### Phase 3: 测试执行

参考 [endpoint-map.md](./references/endpoint-map.md) 获取每个端点的请求模板。
参考 [comparison-rules.md](./references/comparison-rules.md) 获取响应对比规则。

#### 3.1 单端验证模式

对每个选中的端点执行：

```
步骤 1: 执行 mutation（加购商品）
步骤 2: 保存 mutation 响应到 /tmp/ttpos_test_{source}_mutation.json
步骤 3: 执行 query（查询购物车）
步骤 4: 保存 query 响应到 /tmp/ttpos_test_{source}_query.json
步骤 5: 按 comparison-rules.md 对比两个响应
步骤 6: 记录对比结果（PASS/FAIL + 差异字段）
```

可选附加测试（如果端点支持）：
- 改数量（OrderCartProductNum）
- 减为 0
- 送厨后再加购

#### 3.2 跨端交互模式

生成端点两两组合的测试用例矩阵：

```
对于选中的端点列表 [A, B, C, ...]：
  对每对 (X, Y) where X != Y：
    步骤 1: 端点 X 执行 mutation（加购）
    步骤 2: 保存 X 的 mutation 响应
    步骤 3: 端点 Y 执行 query（查询）
    步骤 4: 保存 Y 的 query 响应
    步骤 5: 按 comparison-rules.md 归一化后对比
    步骤 6: 记录结果
```

**关键跨端验证点**：
- cashier 加购后，H5 的 unordered 列表不应受影响
- H5 加购后，cashier 的 product_list 不应包含 H5 未确认商品
- assistant 加购后，cashier 和 assistant 看到的商品列表应一致
- 任一端改数量后，其他端的金额应正确反映

#### 3.3 并发安全模式

```
步骤 1: 记录初始状态（各端 query）
步骤 2: 多端同时 mutation（使用 bash & + wait）
步骤 3: 检查所有 mutation 响应 code=0
步骤 4: 各端 query 获取最终状态
步骤 5: 验证最终状态一致性：
  - 所有 accepted 端看到相同的 product 数量
  - H5 看到正确的 unordered 数量
  - 无重复商品 UUID
步骤 6: 记录结果
```

并发实现模式：

```bash
source /tmp/ttpos_regression_env.sh

# 并发执行
curl -s -X POST "$HOST/api/v1/cashier/desk/order/cart/product/add" \
  -H "Authorization: Bearer $CASHIER_TOKEN" \
  -H "Content-Type: application/json" \
  -d "$CASHIER_BODY" -o /tmp/ttpos_test_concurrent_cashier.json &

curl -s -X POST "$HOST/api/v1/assistant/desk/order/cart/product/add" \
  -H "Authorization: Bearer $ASSISTANT_TOKEN" \
  -H "Content-Type: application/json" \
  -d "$ASSISTANT_BODY" -o /tmp/ttpos_test_concurrent_assistant.json &

curl -s -X POST "$HOST/api/v1/h5/order/cart/product/add" \
  -H "Authorization: Bearer $H5_TOKEN" \
  -H "Content-Type: application/json" \
  -d "$H5_BODY" -o /tmp/ttpos_test_concurrent_h5.json &

wait
```

### Phase 4: 对比 & 报告

#### 4.1 响应对比

使用 python3 内联脚本解析 JSON 并执行字段级对比：

```python
# 对比逻辑伪代码
def compare_shopcart(mutation_resp, query_resp):
    """对比两个 ShopCart 响应"""
    mismatches = []
    m_products = extract_products(mutation_resp)  # data.sale_order_list[].product_list[]
    q_products = extract_products(query_resp)

    # 1. 商品 UUID 集合是否一致
    # 2. 每个商品的 num, price, status, total_price, discount_fee
    # 3. SaleOrder 级别的 product_num, amount_info
    return mismatches

def compare_h5_deskping(mutation_resp, query_resp):
    """对比两个 H5DeskPing 响应"""
    # unsent_kitchen.products.list[] 的商品列表
    # unsent_kitchen.amount_info 的 product_num, product_amount
    ...

def cross_compare(source_a_resp, source_b_resp, type_a, type_b):
    """跨端对比（归一化后）"""
    # 根据 comparison-rules.md 提取可比字段
    ...
```

具体对比规则见 [comparison-rules.md](./references/comparison-rules.md)。

#### 4.2 生成报告

输出结构化 Markdown 报告：

```markdown
## 回归测试报告

### 测试配置
- **模式**: {single/cross/concurrent/full}
- **端点**: {cashier, assistant, h5, tablet}
- **商户**: {company_uuid}
- **桌台**: {desk_no} (uuid: {desk_uuid})
- **账单**: {sale_bill_uuid}
- **测试商品**: {bom_uuid_1} (¥{price_1}), {bom_uuid_2} (¥{price_2})
- **时间**: {timestamp}

### 结果总览

| # | 测试用例 | 状态 | 详情 |
|---|---------|------|------|
| 1 | Cashier 单端: add → cart/info | PASS | product_num=3, amount=660 |
| 2 | H5→Cashier 跨端: H5 add → Cashier query | PASS | H5 商品已隔离 |
| 3 | 并发 3 端同时 add | PASS | 无竞态，锁串行化正确 |

**通过**: {pass_count}/{total_count}
**失败**: {fail_count}/{total_count}

### 详细对比（仅失败项）

#### FAIL: {test_case_name}
**Mutation 响应** ({mutation_endpoint}):
- product_num: 3
- total_amount: 660.00

**Query 响应** ({query_endpoint}):
- product_num: 2         ← MISMATCH
- total_amount: 440.00   ← MISMATCH

**差异字段**:
| 字段 | Mutation | Query |
|------|----------|-------|
| product_num | 3 | 2 |
| total_amount | 660.00 | 440.00 |

### 发现的问题
{如有问题，列出可能的根因和涉及文件}
```

### Phase 5: 清理

测试完成后执行清理：

1. **恢复平板绑定**（如修改过）：
   ```sql
   UPDATE ttpos_desk SET device_uuid = 0 WHERE uuid = $DESK_UUID;
   ```
2. **临时文件清理**（可选）：
   ```bash
   rm -f /tmp/ttpos_test_*.json /tmp/ttpos_regression_env.sh
   ```
3. **不自动关桌/删商品** — 保留测试数据供人工复查

## 注意事项

### 安全规范
- **只读数据库操作**：除平板绑定外，所有 DB 操作均为 SELECT
- **平板绑定必须恢复**：测试后恢复 `device_uuid = 0`
- **不操作生产环境**：默认 dev 环境，生产环境需用户二次确认

### 已知行为
- assistant 的 `GetDeskPing` 的 `product_num` 包含 H5 未确认商品（预存行为，非 bug）
- tablet 的 `add_and_cooking` 返回空响应，不走 `buildShopCartFromSaleBill`
- H5 使用 DeskToken 而非 JWT，token 格式和生成方式不同
- API 成功响应 `code=0`（非 CLAUDE.md 中的 `code=1`）

### 扩展新端点

1. 在 [endpoint-map.md](./references/endpoint-map.md) 添加新的 mutation/query 配对
2. 如有新的响应类型，在 [comparison-rules.md](./references/comparison-rules.md) 添加提取规则
3. Phase 3 的测试逻辑自动适配新端点

## 相关 Skill

- [main-debug](../main-debug/) — 测试发现问题后的根因排查
- [managing-knowledge](../managing-knowledge/) — 记录测试中发现的模式和问题
