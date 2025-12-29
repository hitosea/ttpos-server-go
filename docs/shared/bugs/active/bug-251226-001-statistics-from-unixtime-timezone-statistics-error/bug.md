# Bug-251226-001: FROM_UNIXTIME时区问题导致统计错误

## 基本信息

| 字段       | 值                    |
| ---------- | --------------------- |
| Bug ID     | bug-251226-001        |
| 模块       | statistics            |
| 严重程度   | high                  |
| 发现版本   | 待确认                |
| 发现日期   | 2025-12-26            |
| 发现者     | 王昱                  |
| 状态       | 🟡 规划中             |

## 问题描述

### 现象

统计功能中使用了 MySQL 的 `FROM_UNIXTIME` 函数进行日期格式化，但该函数使用的是 MySQL 服务器的时区，而不是业务时区，导致统计结果错误。

### 复现步骤

1. 设置商户时区为 `Asia/Shanghai`（UTC+8）
2. 在跨日期的边界时间（如 23:00-01:00）创建订单
3. 查询统计报表（按日统计）
4. 观察统计结果，发现日期分组错误

### 预期行为

统计结果应该按照业务时区（商户设置的时区）进行日期分组，例如：
- 订单完成时间戳 `1735142400`（2024-12-26 00:00:00 UTC+8）
- 应该被统计到 `2024-12-26` 这一天

### 实际行为

如果 MySQL 服务器时区与业务时区不一致，`FROM_UNIXTIME` 会使用服务器时区进行转换，导致：
- 订单完成时间戳 `1735142400`（2024-12-26 00:00:00 UTC+8）
- 如果服务器时区是 UTC，会被转换为 `2024-12-25 16:00:00 UTC`
- 统计结果错误地归到 `2024-12-25` 这一天

## 环境信息

- **技术栈**: Go Main 模块
- **数据库**: MySQL 8.0+
- **影响文件**:
  - `main/app/repository/statistics.go` - 统计仓库层
  - `main/app/service/statistics.go` - 统计服务层
  - `main/app/api/v1/shop/shop_statistics.go` - 统计 API 层

## 影响范围

### 受影响的统计功能

1. **综合运营统计** (`CountBusinessSummary`)
   - 按日/按月统计订单金额、支付金额、退款金额等
   - 使用 `FROM_UNIXTIME(sb.finish_time, '%Y-%m-%d')` 或 `FROM_UNIXTIME(sb.finish_time, '%Y-%m')`

2. **支付方式统计** (`CountBusinessPaymentMethod`)
   - 按日/按月统计各支付方式的支付金额和笔数
   - 使用 `FROM_UNIXTIME(po.create_time, '%Y-%m-%d')` 或 `FROM_UNIXTIME(po.create_time, '%Y-%m')`

3. **其他使用 `FROM_UNIXTIME` 的统计查询**
   - 代码中搜索到 14 处 `FROM_UNIXTIME` 的使用

### 影响的终端

- **shop**: 店铺后台统计报表
- **pos**: 前台收银统计（如果使用）

### 业务影响

- 统计报表数据不准确
- 跨时区商户的数据错误
- 可能导致财务对账错误

## 初步分析

### 问题根源

MySQL 的 `FROM_UNIXTIME` 函数使用服务器时区进行时间转换，而不是业务时区。代码中虽然传递了 `Timezone` 参数，但在 SQL 查询中未使用。

### 相关代码位置

```1241:1276:main/app/repository/statistics.go
totalQuery := fmt.Sprintf(`
	SELECT COUNT(DISTINCT FROM_UNIXTIME(sb.finish_time, '%s'))
	FROM ttpos_sale_bill AS sb
	...
`, dateFormat)

query := fmt.Sprintf(`
	SELECT 
		date,
		...
	FROM (
		SELECT 
			FROM_UNIXTIME(sb.finish_time, '%s') AS date,
			...
	`, dateFormat)
```

```1413:1422:main/app/repository/statistics.go
countQuery := fmt.Sprintf(`
	SELECT COUNT(DISTINCT CONCAT(FROM_UNIXTIME(po.create_time, '%s'), '_', po.payment_method_uuid))
	%s
`, dateFormat, baseQuery)

dataQuery := fmt.Sprintf(`
	SELECT 
		FROM_UNIXTIME(po.create_time, '%s') AS date,
		...
`, dateFormat, baseQuery)
```

### 解决方案思路

1. **方案一**: 使用 `CONVERT_TZ` 函数
   - `CONVERT_TZ(FROM_UNIXTIME(timestamp), @@session.time_zone, 'Asia/Shanghai')`
   - 需要设置 MySQL session time_zone

2. **方案二**: 在应用层转换时区
   - 使用 Go 的 `time` 包进行时区转换
   - 在查询前将时间戳转换为目标时区的日期字符串

3. **方案三**: 使用 MySQL 的时区转换函数
   - `DATE_FORMAT(CONVERT_TZ(FROM_UNIXTIME(timestamp), '+00:00', '+08:00'), '%Y-%m-%d')`

### 参考文档

- `docs/shared/troubleshooting/mysql-fromunixtime-timezone.md` - 时区问题排查文档（待完善）

## 相关链接

- 相关文件: `main/app/repository/statistics.go`
- 相关文件: `main/app/service/statistics.go`
- 相关文件: `main/app/api/v1/shop/shop_statistics.go`
- 排查文档: `docs/shared/troubleshooting/mysql-fromunixtime-timezone.md`

## 下一步

1. ✅ Bug 报告已创建
2. ✅ 技术分析（确定最佳解决方案）
3. ✅ 使用 `/bug-spec` 创建修复方案和任务分解
4. ⏳ 实施修复（参考 `tasks.md`）
5. ⏳ 测试验证
6. ⏳ 使用 `/bug-archive` 归档

## 修复方案

- 修复方案文档: `solution.md`
- 任务分解清单: `tasks.md`

