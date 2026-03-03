---
name: main-debug
description: 问题诊断与调试。当用户描述线上问题、Bug、异常行为、数据不一致等问题时触发。分析根因、引导完善问题、必要时查询数据库验证。
---

# 问题诊断与调试

## 触发条件

当用户描述以下场景时触发：

- 线上 Bug、异常行为、报错信息
- 数据不一致、数据丢失、数据异常
- 接口返回异常、状态码错误
- 性能问题、超时、慢查询
- 业务逻辑不符合预期
- 用户使用 `/main-debug` 命令

## 执行流程

### Phase 1: 问题收集与澄清

#### 1.1 解析用户输入

从用户描述中提取关键信息：

| 信息项 | 说明 | 示例 |
|--------|------|------|
| **现象** | 发生了什么 | "订单金额为 0"、"接口返回 500" |
| **商户** | company_uuid 或商户名 | `123123243435000` |
| **时间** | 何时发生 | "今天下午"、"2026-03-01" |
| **终端** | 哪个终端 | pos / shop / kds / mobile / menu |
| **频率** | 偶发还是必现 | "每次都这样"、"偶尔出现" |
| **接口/页面** | 哪个功能 | "/api/v1/order_info"、"结账页面" |

#### 1.2 判断信息完整度

**信息充分** — 可直接进入 Phase 2 的条件：
- 至少有明确的**现象描述**
- 能定位到**模块/功能范围**（通过接口路径、功能名称、错误信息等推断）

**信息不足** — 需要引导提问：

使用 `AskUserQuestion` 工具进行**最多 2 轮**引导提问，每轮最多 2 个问题。

引导策略：
- **缺少商户信息** → 询问 company_uuid 或商户名称
- **缺少复现条件** → 询问操作步骤和频率
- **现象不清晰** → 询问具体的错误信息、截图、日志
- **范围太大** → 询问具体的终端和功能页面

```yaml
# 引导提问示例
问题: "能提供以下信息帮助定位问题吗？"
选项:
  - 商户 company_uuid 或名称
  - 具体的错误信息或接口路径
  - 操作步骤和复现频率
```

> **注意**：不要过度提问。如果通过代码分析就能缩小范围，优先分析代码。

### Phase 2: 问题调查

根据问题类型选择合适的调查手段，**按优先级排列**：

#### 2.1 检索历史经验（优先）

先查询 Graphiti 是否有类似问题的解决记录：

```yaml
工具: mcp__Graphiti__search_memory_facts
参数:
  query: "{问题关键词} bug solution"
  group_id: "ttpos-troubleshooting"
  max_facts: 5
```

如果找到匹配记录，优先参考历史解决方案，避免重复排查。

#### 2.2 代码分析

根据问题定位相关代码：

**定位路由和接口**：
```yaml
# 如果知道接口路径，搜索路由注册
工具: Grep
参数:
  pattern: "{接口路径关键词}"
  path: "main/app/api/"
```

**追踪调用链**：
```yaml
# 使用 Serena 语义分析追踪符号引用
工具: mcp__Serena__find_symbol
参数:
  symbol_name: "{函数或结构体名}"
  include_body: true
```

```yaml
# 查找引用关系
工具: mcp__Serena__find_referencing_symbols
参数:
  symbol_name: "{函数名}"
```

**分析分层架构**（按 API → Service → Repository → Model 逐层深入）：

| 层级 | 搜索路径 | 关注点 |
|------|----------|--------|
| API | `main/app/api/v1/` | 参数校验、绑定方式 |
| Service | `main/app/service/` | 业务逻辑、条件判断 |
| Repository | `main/app/repository/` | 查询条件、SQL 构造 |
| Model | `main/app/model/` | 字段定义、类型映射 |
| DTO | `main/app/dto/` | 请求/响应结构 |

#### 2.3 数据库查询（按需）

**何时需要查询数据库**：
- 数据不一致问题
- 需要验证数据状态
- 需要确认记录是否存在
- 需要检查关联数据完整性

**数据库连接**：

```bash
# 读取环境变量（动态获取项目根目录，适配 worktree 等场景）
source "$(git rev-parse --show-toplevel)/.env"

# 方式1：通过 source .env 读取变量（双引号展开变量）
mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USERNAME" -p"$DB_PASSWORD" "$DB_DATABASE" -e "
  SELECT uuid, name FROM ttpos_company WHERE uuid = '{company_uuid}';
"

# 方式2：直接填写连接信息（用单引号包裹，避免特殊字符如 ! @ $ 被 shell 解释）
mysql -h'主机' -P'端口' -u'用户名' -p'密码' 'shop{company_uuid}' -e "
  {业务查询SQL}
"
```

**查询安全规范**：
- **只使用 SELECT**，严禁 INSERT/UPDATE/DELETE
- 加 `LIMIT` 限制返回行数（默认 `LIMIT 20`）
- 敏感数据（密码、token）不展示完整值
- 大表查询必须带 WHERE 条件，禁止全表扫描
- 查询前先确认表是否存在：`SHOW TABLES LIKE '{table_name}';`

**常用排查查询模板**：

```sql
-- 检查订单状态
SELECT uuid, order_no, status, total_amount, pay_status,
       FROM_UNIXTIME(create_time) as created
FROM ttpos_order
WHERE uuid = '{order_uuid}'
LIMIT 1;

-- 检查商品数据
SELECT uuid, name, price, status, delete_time
FROM ttpos_product
WHERE uuid = '{product_uuid}'
LIMIT 1;

-- 检查最近的异常数据
SELECT uuid, FROM_UNIXTIME(create_time) as created
FROM {table_name}
WHERE create_time > UNIX_TIMESTAMP('{date}')
ORDER BY create_time DESC
LIMIT 20;

-- 检查表结构
DESC {table_name};

-- 检查关联数据完整性
SELECT a.uuid, b.uuid as related_uuid
FROM {table_a} a
LEFT JOIN {table_b} b ON a.{fk_field} = b.uuid
WHERE b.uuid IS NULL AND a.delete_time = 0
LIMIT 20;
```

#### 2.4 日志与配置检查

```bash
# 检查应用日志（默认目录 main/log/，文件名格式 YYYY-MM-DD.log）
PROJECT_ROOT="$(git rev-parse --show-toplevel)"
tail -100 "$PROJECT_ROOT/main/log/$(date +%F).log" 2>/dev/null

# 检查 SQL 日志
tail -100 "$PROJECT_ROOT/main/log/$(date +%F).sql.log" 2>/dev/null

# 检查 Go 编译是否有问题
cd "$PROJECT_ROOT/main" && go vet ./app/...
```

### Phase 3: 根因分析

完成调查后，按以下结构输出分析结果：

```markdown
## 🔍 问题诊断报告

### 问题现象
{一句话描述问题}

### 影响范围
- **终端**: {受影响的终端}
- **商户**: {受影响的商户范围}
- **严重程度**: 高 / 中 / 低

### 根本原因
{详细解释根本原因，引用具体代码位置}

关键代码：`{file_path}:{line_number}`

### 证据
{支撑根因结论的证据列表}
1. {代码逻辑分析}
2. {数据查询结果}
3. {日志记录}

### 涉及文件
| 文件路径 | 行号 | 说明 |
|----------|------|------|
| `{relative_path}` | {line} | {角色说明：Bug 所在 / 调用方 / 数据层等} |

### 修复建议
{具体的修复方案}

### 预防建议
{如何避免类似问题再次发生}
```

**严重程度判定标准**：

| 级别 | 标准 |
|------|------|
| **高** | 影响核心流程（下单/支付/结账），数据丢失，全部商户受影响 |
| **中** | 影响非核心功能，部分商户受影响，有 workaround |
| **低** | 界面显示问题，边缘场景，不影响业务数据 |

### Phase 4: 经验记录

如果问题是**非平凡问题**（排查过程涉及多步分析或发现了隐藏的 Bug），自动记录到 Graphiti：

```yaml
工具: mcp__Graphiti__add_memory
参数:
  name: "qa-{issue-keyword}-{YYYY-MM}"
  group_id: "ttpos-troubleshooting"
  content: |
    问题: {一句话描述}
    原因: {根本原因}
    解决方案:
    1. {步骤1}
    2. {步骤2}
    关键代码: {file_path}:{line_number}
    预防措施: {如何避免}
```

## 问题类型速查

### 按模块定位

| 关键词 | 模块 | 入口代码 |
|--------|------|----------|
| 订单/下单/结账 | Main | `main/app/service/order_*.go` |
| 商品/菜品/分类 | Main | `main/app/service/product_*.go` |
| 会员/积分/充值 | Main | `main/app/service/member_*.go` |
| 桌位/桌台 | Main | `main/app/service/table_*.go` |
| 支付/收款 | Main | `main/app/service/payment_*.go` |
| 外卖/外送 | BMP | `ttpos-bmp/app/ttpos-takeout/` |
| 报表/统计 | Main | `main/app/service/report_*.go` |
| 打印/小票 | Main | `main/app/service/print_*.go` |
| 用户/登录/权限 | Main | `main/app/service/staff_*.go` |
| ERP/采购/库存 | BMP | `ttpos-bmp/app/ttpos-erp/` |

### 按错误类型定位

| 错误现象 | 常见原因 | 排查方向 |
|----------|----------|----------|
| 接口 500 | panic / nil pointer / DB 错误 | 检查 Service 层 error 处理 |
| 接口 401/403 | Token 过期 / 权限不足 | 检查 auth 中间件 |
| 数据为空 | 查询条件错误 / delete_time 过滤 | 检查 Repository 查询条件 |
| 金额异常 | 精度丢失 / 计算逻辑错误 | 检查 decimal 类型使用 |
| 多语言缺失 | LocaleResponse 未正确填充 | 检查 DTO 转换逻辑 |
| 并发异常 | 缺少锁 / 事务隔离级别 | 检查分布式锁和事务使用 |
| 数据不同步 | 事件未触发 / MQ 消费失败 | 检查 eventbus / rocketmq |

## 调试原则

1. **先搜后问**：先利用代码分析和历史经验缩小范围，再向用户提问
2. **最小查询**：数据库查询只查必要字段，避免 `SELECT *`
3. **只读操作**：严禁对数据库执行任何写操作
4. **保护隐私**：不展示用户密码、token、手机号等敏感信息
5. **逐层深入**：按 API → Service → Repository 层级逐步排查，不要跳层
6. **证据驱动**：每个结论都要有代码或数据支撑，不做无根据的猜测

## 相关 Skill

- [managing-knowledge](../managing-knowledge/) — 记录排查经验
- [bmp-autogen](../bmp-autogen/) — BMP 模块代码生成问题排查
