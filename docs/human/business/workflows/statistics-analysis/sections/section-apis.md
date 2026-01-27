# API 设计与规范

## 一、接口设计风格

### 1.1 命名规范

| 前缀 | 语义 | 示例 |
|-----|------|------|
| Count* | 统计计算，返回聚合数据 | CountSale, CountBusiness |
| Export* | 导出操作，生成文件 | ExportBusinessSummary |
| Rank* | 排名统计，返回有序列表 | RankProduct |
| Save* | 数据持久化（内部方法）| SaveSale |

### 1.2 接口职责划分

**IBusinessSrv**（业务统计服务）：
- 面向终端用户的业务数据统计
- 包含数据格式化和多语言处理
- 支持打印、导出等复合功能

**IStatisticsSrv**（核心统计服务）：
- 提供底层统计计算能力
- 返回原始聚合数据
- 供 BusinessSrv 复用

## 二、请求参数设计

### 2.1 时间参数优先级

```
TimeType > StartTime/EndTime > QueryStartDate/QueryEndDate
```

### 2.2 过滤参数

| 参数 | 类型 | 用途 | 默认 |
|-----|------|------|-----|
| StaffUuid | uint64 | 按员工过滤 | 不过滤 |
| AreaUuid | uint64 | 按区域过滤 | 不过滤 |
| DutyNo | string | 按班次过滤 | 不过滤 |
| ExcludeDataManage | bool | 排除数据管理订单 | true |

## 三、响应结构设计

### 3.1 金额字段规范

| 命名模式 | 含义 |
|---------|------|
| total_xxx_amount | 汇总金额 |
| avg_xxx_amount | 平均金额 |
| min_xxx_amount | 最小金额 |
| max_xxx_amount | 最大金额 |

精度统一为两位小数，通过 `Decimal.Round(2).InexactFloat64()` 转换。

### 3.2 多语言支持

响应字段使用 `LocaleResponse` 结构，支持 9 种语言（zh/en/th/zhtw/ja/ko/my/tr/sv）。

## 四、版本兼容性

### 4.1 参数向后兼容

```go
// 优先使用新参数，回退到旧参数
startTime := req.StartTime
if req.QueryStartDate != "" && startTime == 0 {
    startTime, _ = timeUtil.FormatDateTimeToUnix(req.QueryStartDate)
}
```

### 4.2 响应字段扩展

新增响应字段采用**追加而非修改**策略，确保旧版客户端不受影响。

## 五、设计最佳实践

1. **单一职责**：Count/Export/Rank 方法各司其职
2. **参数可选**：通过默认值减少必填参数
3. **响应一致**：统一的字段命名和数据类型
4. **向后兼容**：新增而非修改
5. **多语言支持**：服务端处理本地化
