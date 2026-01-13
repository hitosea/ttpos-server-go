# Opt-260113-001: 统计查询 MySQL FROM_UNIXTIME 时区优化

## 基本信息

| 字段       | 值                    |
| ---------- | --------------------- |
| 优化 ID    | opt-260113-001        |
| 模块       | statistics            |
| 优化类型   | performance           |
| 优先级     | high                  |
| 当前版本   | v2.13                 |
| 提出日期   | 2026-01-13            |
| 提出者     | 王昱                  |
| 状态       | 🟢 规划中             |
| 优化方案   | solution.md           |
| 任务清单   | tasks.md              |

## 优化需求

### 当前问题

统计功能中使用了 MySQL 的 `FROM_UNIXTIME` 函数进行日期格式化，但该函数使用的是 MySQL 服务器的时区，而不是业务时区（商户设置的时区），导致统计结果错误。

**问题表现**：
- 当 MySQL 服务器时区与业务时区不一致时，统计报表的日期分组错误
- 跨日期边界时间（如 23:00-01:00）的订单被错误地归类到错误的日期
- 跨时区商户（如日本、泰国、土耳其等）的统计数据不准确
- 可能导致财务对账错误

**问题根源**：
- `FROM_UNIXTIME(timestamp)` 函数使用 MySQL 服务器的时区进行时间转换
- 代码中虽然 Service 层已获取业务时区，但 Repository 层的 SQL 查询中未使用时区转换
- 约 105 处使用 `FROM_UNIXTIME` 的查询，其中 14 处关键统计查询受影响

### 性能指标（如适用）

- **当前性能**: 统计数据不准确，跨时区商户的数据错误
- **目标性能**: 所有统计查询按业务时区正确分组，数据准确性 100%
- **提升目标**: 
  - 统计数据准确性：从不准确 → 100% 准确
  - 跨时区商户支持：从不支持 → 完全支持

### 影响面

- **影响终端**: 
  - shop（店铺后台统计报表）
  - pos（前台收银统计，如果使用）
- **影响用户**: 
  - 商户管理员（查看统计报表）
  - 财务人员（财务对账）
  - 跨时区商户（日本、泰国、土耳其等）
- **业务价值**: 
  - 确保统计数据的准确性和可靠性
  - 支持跨时区商户的正确统计
  - 避免财务对账错误
  - 提升系统的国际化支持能力

## 触发原因

**现状分析**：
1. **技术债务**：代码中直接使用 `FROM_UNIXTIME`，未考虑时区转换
2. **设计缺陷**：Repository 层方法未接收时区参数，SQL 查询中未使用时区转换
3. **业务需求**：系统需要支持跨时区商户，统计数据必须按商户时区计算

**相关记录**：
- 已发现相关 Bug：`bug-251226-001-statistics-from-unixtime-timezone-statistics-error`
- 已有相关 Spec（已归档）：`story-shop-statistics-merchant-timezone-query`
- 相关提案：`merchant-timezone-based-statistics-query`

**用户反馈**：
- 跨时区商户反映统计数据不准确
- 财务对账时发现日期分组错误

## 初步分析

### 可能原因

1. **MySQL `FROM_UNIXTIME` 函数的时区特性**
   - `FROM_UNIXTIME(timestamp)` 使用 MySQL 服务器的时区进行转换
   - 如果服务器时区与业务时区不一致，会导致日期分组错误

2. **代码设计缺陷**
   - Service 层已获取业务时区：`timezone := ctx.GetCompanySetting().Timezone`
   - Repository 层的方法未接收时区参数
   - SQL 查询中直接使用 `FROM_UNIXTIME`，未考虑时区转换

3. **时区不一致的场景**
   - 跨时区商户（如日本、泰国、土耳其等）
   - MySQL 服务器时区设置为 UTC
   - 业务时区设置为 `Asia/Shanghai`、`Asia/Tokyo` 等

### 优化方向

1. **方案一：使用 MySQL `CONVERT_TZ` 函数**
   - 在数据库层完成时区转换，性能较好
   - 需要将时区名称转换为偏移量（如 `Asia/Shanghai` → `+08:00`）
   - 需要处理夏令时等复杂情况

2. **方案二：在应用层转换时区**
   - 完全控制时区转换逻辑，不依赖 MySQL 时区数据
   - 需要在应用层计算日期分组，SQL 查询逻辑复杂
   - 性能可能不如数据库层转换

3. **方案三：设置 MySQL Session Timezone + `CONVERT_TZ`**
   - 结合两种方案的优点
   - 需要修改数据库连接配置，可能影响其他查询

**推荐方案**：方案一（使用 MySQL `CONVERT_TZ` 函数）
- 性能最优，在数据库层完成转换
- 代码改动相对较小，只需修改 SQL 查询
- 项目已有 `utils.Timezone` 工具类，可以复用

### 预估收益

- **数据准确性**：从不准确 → 100% 准确
- **跨时区支持**：完全支持跨时区商户
- **业务价值**：避免财务对账错误，提升系统可靠性
- **技术债务**：消除约 14 处关键统计查询的时区问题

## 相关链接

- **相关 Bug**: `bug-251226-001-statistics-from-unixtime-timezone-statistics-error`
- **相关 Spec**: `story-shop-statistics-merchant-timezone-query` (已归档)
- **相关提案**: `merchant-timezone-based-statistics-query`
- **问题排查文档**: `docs/shared/troubleshooting/mysql-fromunixtime-timezone.md`
- **影响文件**:
  - `main/app/repository/statistics.go` - 统计仓库层（约 14 处 `FROM_UNIXTIME` 使用）
  - `main/app/service/statistics.go` - 统计服务层
  - `main/app/api/v1/shop/shop_statistics.go` - 统计 API 层

## 受影响的统计功能

1. **综合运营统计** (`CountBusinessSummary`)
   - 按日/按月统计订单金额、支付金额、退款金额等
   - 使用 `FROM_UNIXTIME(sb.finish_time, '%Y-%m-%d')` 或 `FROM_UNIXTIME(sb.finish_time, '%Y-%m')`

2. **支付方式统计** (`CountBusinessPaymentMethod`)
   - 按日/按月统计各支付方式的支付金额和笔数
   - 使用 `FROM_UNIXTIME(po.create_time, '%Y-%m-%d')` 或 `FROM_UNIXTIME(po.create_time, '%Y-%m')`

3. **销售天数统计** (`CountSaleDays`)
   - 使用 `FROM_UNIXTIME(complete_time, '%Y-%m-%d')`

4. **支付天数统计** (`CountPaymentDays`)
   - 使用 `FROM_UNIXTIME(sp.complete_time, '%Y-%m-%d')`

5. **区域统计** (`CountAreaDays`)
   - 使用 `FROM_UNIXTIME(ss.complete_time, '%Y-%m-%d')`

6. **其他使用 `FROM_UNIXTIME` 的统计查询**
   - 代码中搜索到约 105 处 `FROM_UNIXTIME` 的使用，其中 14 处关键统计查询需要修复

## 下一步

1. **技术评估**：评估各方案的可行性和成本
2. **方案设计**：使用 `/optimize-spec` 创建优化方案和任务分解
3. **实施优化**：按方案逐步修复所有受影响的统计查询
