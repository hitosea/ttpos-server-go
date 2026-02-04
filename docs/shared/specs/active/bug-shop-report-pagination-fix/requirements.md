# 门店统计报表分页数据不一致修复 需求文档

## 📋 基本信息

| 项目              | 内容                                                                 |
| ----------------- | -------------------------------------------------------------------- |
| **Spec ID**       | bug-shop-report-pagination-fix                                       |
| **来源 Proposal** | [shop-report-pagination-fix](../../../../team/proposals/2026-02/shop-report-pagination-fix.md) |
| **创建日期**      | 2026-02-04                                                           |
| **负责人**        | 王昱                                                                 |
| **目标 Sprint**   | 待定                                                                 |

## 📋 审核状态

| 项目         | 内容   |
| ------------ | ------ |
| **审核状态** | 待审核 |
| **审核人**   | -      |
| **审核日期** | -      |

---

## 📝 用户故事

**作为** 店长、商户管理员或运营人员
**我想** 在门店统计报表中切换分页时看到稳定一致的数据
**以便于** 准确了解业务情况并做出正确的经营决策

---

## 缺陷描述

### 问题现象

新管理端（Shop）的门店汇总统计报表在切换分页时，显示的数据与预期不符：
- 同一页多次请求返回不同数据
- 相同日期的门店记录顺序随机变化
- 快速翻页时数据出现重复或遗漏

### 根因分析

经代码排查，定位到 `main/app/service/business.go` 中的三个核心问题：

| 位置 | 问题 | 严重度 |
|------|------|--------|
| 行 4448-4450, 4520-4522 | 排序仅按 `Date`，缺乏二级排序字段，相同日期记录顺序不稳定 | 🔴 严重 |
| 行 4174-4330 | 并发查询（goroutine）结果收集顺序取决于协程完成顺序，不确定 | 🔴 严重 |
| 行 4356-4445 | Map 遍历顺序随机（Go 语言特性） | 🟠 中等 |

### 受影响方法

1. `CountCompanyBusinessSummary` (行 4098-4558)
2. `countCompanyPaymentMethodSummary` (行 4560-5001)
3. `countCompanyRefundSummary` (行 5003-5358)

---

## 功能需求

### Requirement 1: 添加二级排序字段（店铺编号）

**用户故事**: 作为用户，我想在分页时数据顺序稳定，以便于准确追踪数据

#### 排序规则

二级排序字段使用**店铺编号（CompanyCode）**，排序规则如下：

1. **未设置店铺编号的记录排在最前面**
2. **数字优先**：0-9 排在字母前面
3. **字母次之**：a-z（不区分大小写）
4. **混合编号**：按字符逐位比较（如 `1a` < `1b` < `2a`）

**排序示例**：
```
(空) → 001 → 01 → 1 → 10 → 1a → 2 → a1 → a2 → b1
```

#### 验收标准

1. **WHEN** 排序时存在相同日期的记录 **THEN** 系统 **SHALL** 使用店铺编号（CompanyCode）作为二级排序字段
2. **IF** 店铺编号为空 **THEN** 系统 **SHALL** 将该记录排在最前面
3. **WHEN** 比较两个店铺编号 **THEN** 系统 **SHALL** 按"数字优先、字母次之"的规则排序
4. **WHEN** 多次请求同一页数据 **THEN** 系统 **SHALL** 返回完全相同的记录顺序

### Requirement 2: Map 遍历前排序 key

**用户故事**: 作为用户，我想数据聚合后顺序一致，以便于分页结果稳定

#### 验收标准

1. **WHEN** 遍历 `dateDecimalMap` 生成结果列表 **THEN** 系统 **SHALL** 先提取并排序 key，再按排序后的 key 顺序遍历
2. **IF** Map 中有 N 个日期 **THEN** 系统 **SHALL** 按日期升序输出

### Requirement 3: 并发结果保持顺序

**用户故事**: 作为用户，我想多门店数据查询后顺序稳定，以便于分页一致

#### 验收标准

1. **WHEN** 并发查询多个门店数据 **THEN** 系统 **SHALL** 按原始门店列表顺序收集结果
2. **IF** 使用 goroutine 并发查询 **THEN** 系统 **SHALL** 使用带索引的结果收集机制，而非依赖 channel 接收顺序

---

## 非功能需求

### 测试要求

- [ ] 单元测试覆盖率 ≥ 80%
- [ ] 分页一致性测试：同一页连续请求 10 次，结果完全一致
- [ ] 边界测试：空数据、单条数据、跨页边界

### 性能要求

- [ ] 修复后性能不低于修复前（排序操作不应显著影响响应时间）

### 平台兼容性

- [x] Shop 商家管理端（Web）

---

## 约束条件

### 技术约束

- Go 版本: 1.23+
- 框架: Gin + GORM
- 分层架构: API → Service → Repository → Model
- 必须遵循 CLAUDE.md 和 .cursor/rules/go-main.mdc 规范

### 资源约束

- Story Point: 2（修复范围明确，代码改动集中）

---

## 修复方案概要

### 1. 添加二级排序字段（店铺编号）

```go
// 修复前
sort.Slice(finalList, func(i, j int) bool {
    return finalList[i].Date < finalList[j].Date
})

// 修复后
sort.Slice(finalList, func(i, j int) bool {
    if finalList[i].Date != finalList[j].Date {
        return finalList[i].Date < finalList[j].Date
    }
    // 二级排序：店铺编号
    return compareCompanyCode(finalList[i].CompanyCode, finalList[j].CompanyCode)
})

// compareCompanyCode 店铺编号排序比较函数
// 规则：1. 空值排最前 2. 数字优先(0-9) 3. 字母次之(a-z)
func compareCompanyCode(a, b string) bool {
    // 空值排最前
    if a == "" && b == "" {
        return false
    }
    if a == "" {
        return true
    }
    if b == "" {
        return false
    }

    // 逐字符比较：数字 < 字母
    minLen := len(a)
    if len(b) < minLen {
        minLen = len(b)
    }

    for i := 0; i < minLen; i++ {
        ca, cb := a[i], b[i]
        aIsDigit := ca >= '0' && ca <= '9'
        bIsDigit := cb >= '0' && cb <= '9'

        if aIsDigit && !bIsDigit {
            return true // 数字优先
        }
        if !aIsDigit && bIsDigit {
            return false
        }

        // 同类型字符比较（不区分大小写）
        caLower := strings.ToLower(string(ca))[0]
        cbLower := strings.ToLower(string(cb))[0]
        if caLower != cbLower {
            return caLower < cbLower
        }
    }

    return len(a) < len(b)
}
```

### 2. Map 遍历前排序 key

```go
// 修复前
for date, dateDecimal := range dateDecimalMap {
    finalList = append(finalList, ...)
}

// 修复后
dateKeys := make([]string, 0, len(dateDecimalMap))
for date := range dateDecimalMap {
    dateKeys = append(dateKeys, date)
}
sort.Strings(dateKeys)

for _, date := range dateKeys {
    dateDecimal := dateDecimalMap[date]
    finalList = append(finalList, ...)
}
```

### 3. 并发结果保持顺序

```go
// 修复前：依赖 channel 接收顺序
for result := range resultChan {
    allItems = append(allItems, result.items...)
}

// 修复后：使用带索引的结果收集
results := make([]resultItem, len(companyList))
var wg sync.WaitGroup
for idx, companyItem := range companyList {
    wg.Add(1)
    go func(i int, company CompanyItem) {
        defer wg.Done()
        results[i] = queryCompanyData(company)
    }(idx, companyItem)
}
wg.Wait()
```

---

## 风险和缓解

### 风险 1: 修复可能影响其他报表模块

**影响**: 中
**缓解措施**: 修复后进行回归测试，覆盖所有使用相同排序逻辑的报表

### 风险 2: 排序操作可能影响性能

**影响**: 低
**缓解措施**: 排序复杂度为 O(n log n)，对于报表数据量（通常 < 1000 条）影响可忽略

---

**版本**: v1.0.0
**创建日期**: 2026-02-04
