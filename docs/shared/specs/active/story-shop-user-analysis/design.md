# 用户分析统计 设计文档

> 本文档定义用户分析统计功能的详细技术方案。

## 📋 概述

实现用户分析统计功能，包括：
1. 在 `ttpos_statistics_sale` 表新增 `nationality_uuid` 字段
2. 在 `SaveSale` 方法中保存 `nationality_uuid`
3. 新增用户分析统计查询 API
4. 新增用户分析统计导出 API

功能提供四个维度的订单统计分析：国籍、点餐方式来源、桌台方式来源、用餐方式。所有统计按订单数升序排序，排除已被数据管理的订单。

---

## 🎯 规范对齐

### Go Main 规范 (go-main.mdc)

- Service 只依赖其他 Service 接口
- Repository 只持有 `*gorm.DB` 实例
- URL 使用 snake_case
- data 字段必须是对象
- 不使用 panic，返回 error

### API 设计规范 (api.mdc)

- URL 使用 snake_case：`/api/v1/shop/statistics/user_analysis`
- 响应格式：`{code, message, data{}}`
- data 不能为 null 或数组

### 数据库规范 (database.mdc)

- 新增字段使用 `bigint(20) unsigned`，默认值 0
- 时间字段使用 int 类型
- 金额字段使用 decimal(14,2)

---

## 🔄 代码复用分析

| 组件 | 路径 | 复用策略 |
| --- | --- | --- |
| `CountSale` | `main/app/repository/statistics.go` | 复用统计查询逻辑，添加数据管理过滤 |
| `SaveSale` | `main/app/service/statistics.go` | 扩展保存逻辑，添加 `nationality_uuid` 字段 |
| `ExportChannelSales` | `main/app/service/business.go` | 参考导出实现，复用 Excel 导出工具 |
| `WhereNotInDataManageSubQuery` | `main/app/repository/common.go` | 复用数据管理过滤逻辑 |
| 时间工具 | `pkg/timeutil/company_time.go` | 使用门店时区工具推导今日范围 |

---

## 🏗️ 架构设计

```
API(shop_statistics.go)
  ↓ 依赖 (interface)
Service(business.go)
  ↓ 依赖
Repository(statistics.go CountUserAnalysis)
  ↓ 查询
Database(ttpos_statistics_sale, ttpos_sale_bill)
```

- **API 层**：新增 `UserAnalysis`、`ExportUserAnalysis` 方法，固定查询今天数据
- **Service 层**：在 `IBusinessSrv` 中扩展用户分析统计方法，负责数据组装、导出格式
- **Repository 层**：在 `statistics.go` 中新增 `CountUserAnalysis`，按四个维度聚合统计

---

## 🗄️ 数据库设计

### 表结构变更

#### ttpos_statistics_sale 表新增字段

```sql
ALTER TABLE `ttpos_statistics_sale` 
ADD COLUMN `nationality_uuid` bigint(20) unsigned NOT NULL DEFAULT 0 COMMENT '国籍UUID（0=未记录）' AFTER `order_source_uuid`;

-- 添加索引
ALTER TABLE `ttpos_statistics_sale` 
ADD INDEX `idx_nationality_uuid` (`nationality_uuid`);
```

**字段说明**:
| 字段 | 类型 | 说明 | 约束 |
|------|------|------|------|
| nationality_uuid | bigint(20) unsigned | 国籍UUID | DEFAULT 0, INDEX |

**迁移文件**: `admin/database/migrations/{YYYYMMDDHHMMSS}_add_nationality_uuid_to_statistics_sale.php`

---

## 📊 数据模型

### Go Model 更新

```go
// main/app/model/statistics.go
type StatisticsSale struct {
    // ... 现有字段 ...
    OrderSourceUuid      uint64  `gorm:"column:order_source_uuid;type:bigint(20) unsigned;default:0;comment:订单来源UUID（0=店内，>0=外卖/渠道）;NOT NULL" json:"order_source_uuid"`
    NationalityUuid      uint64  `gorm:"column:nationality_uuid;type:bigint(20) unsigned;default:0;comment:国籍UUID（0=未记录）;NOT NULL" json:"nationality_uuid"`
    // ... 其他字段 ...
}
```

### DTO 定义

#### Request DTO

```go
// main/app/dto/req/statistics_user_analysis_req.go
// 无请求参数（固定查询今天数据）
```

#### Response DTO

```go
// main/app/dto/resp/statistics_user_analysis_resp.go
type UserAnalysisItem struct {
    Name       string  `json:"name"`        // 名称（国籍名称/来源名称/用餐方式名称）
    OrderCount int64   `json:"order_count"` // 订单数
    Percentage float64 `json:"percentage"` // 占比（%，保留2位小数，从 decimal 转换）
}

type UserAnalysisResp struct {
    Nationality   []UserAnalysisItem `json:"nationality"`    // 国籍统计
    OrderSource   []UserAnalysisItem `json:"order_source"`   // 点餐方式来源统计
    DeskSource    []UserAnalysisItem `json:"desk_source"`    // 桌台方式来源统计
    DiningMethod  []UserAnalysisItem `json:"dining_method"`  // 用餐方式统计
}
```

---

## 🔌 API 设计

### RESTful API

#### API 1: 用户分析统计查询

**请求**:
- **URL**: `/api/v1/shop/statistics/user_analysis`
- **Method**: `GET`
- **Headers**:
  ```json
  {
    "Authorization": "Bearer {token}",
    "Content-Type": "application/json"
  }
  ```
- **Query**: 无（固定查询今天数据）

**响应**:

```json
{
  "code": 1,
  "message": "success",
  "data": {
    "nationality": [
      {
        "name": "中国",
        "order_count": 100,
        "percentage": 45.23
      },
      {
        "name": "未记录",
        "order_count": 50,
        "percentage": 22.62
      }
    ],
    "order_source": [
      {
        "name": "店内",
        "order_count": 80,
        "percentage": 36.20
      },
      {
        "name": "外卖",
        "order_count": 140,
        "percentage": 63.80
      }
    ],
    "desk_source": [
      {
        "name": "收银机",
        "order_count": 60,
        "percentage": 27.15
      },
      {
        "name": "点餐助手",
        "order_count": 40,
        "percentage": 18.10
      }
    ],
    "dining_method": [
      {
        "name": "打包",
        "order_count": 90,
        "percentage": 40.73
      },
      {
        "name": "店内用餐",
        "order_count": 131,
        "percentage": 59.27
      }
    ]
  }
}
```

#### API 2: 用户分析统计导出

**请求**:
- **URL**: `/api/v1/shop/statistics/user_analysis/export`
- **Method**: `GET`
- **Headers**:
  ```json
  {
    "Authorization": "Bearer {token}",
    "Content-Type": "application/json"
  }
  ```
- **Query**: 无（固定导出今天数据）

**响应**:

```json
{
  "code": 1,
  "message": "success",
  "data": {}
}
```

导出任务异步处理，通过 `export_record` 表记录状态。

---

## 🧩 组件设计

### Repository 层

#### 新增方法

```go
// main/app/repository/statistics.go
import (
    "github.com/shopspring/decimal"
)

// CountUserAnalysis 统计用户分析数据
func (r *StatisticsRepo) CountUserAnalysis(startTime, endTime int64) (*UserAnalysisRepoResult, error) {
    // 1. 按国籍统计
    //    - 先检查是否存在 nationality_uuid > 0 的订单
    //    - 如果所有订单的 nationality_uuid = 0，则返回空数组
    //    - 否则仅统计 nationality_uuid > 0 的订单
    // 2. 按点餐方式来源统计（仅点餐订单）
    // 3. 按桌台方式来源统计（仅桌台订单）
    // 4. 按用餐方式统计（点餐+桌台）
    // 所有统计排除数据管理订单，按订单数升序排序
    // 占比计算使用 decimal：percentage = decimal.NewFromInt(orderCount).Div(decimal.NewFromInt(totalCount)).Mul(decimal.NewFromInt(100)).Round(2)
}

type UserAnalysisRepoResult struct {
    Nationality   []UserAnalysisItemRepo
    OrderSource   []UserAnalysisItemRepo
    DeskSource    []UserAnalysisItemRepo
    DiningMethod  []UserAnalysisItemRepo
}

type UserAnalysisItemRepo struct {
    Name        string
    OrderCount  int64
    Percentage  decimal.Decimal  // 使用 decimal 进行精确计算，保留2位小数
}
```

**统计逻辑**:

1. **国籍统计**：
   - 从 `ttpos_statistics_sale` 表查询
   - **前置检查**：如果查询范围内所有订单的 `nationality_uuid = 0`，则不统计此维度（返回空数组）
   - 仅统计 `nationality_uuid > 0` 的订单
   - 关联 `ttpos_nationality` 表获取国籍名称
   - 排除数据管理订单：`WHERE sale_bill_uuid NOT IN (SELECT data_uuid FROM ttpos_data_manage WHERE type = DataManageTypeOrder)`

2. **点餐方式来源统计**：
   - 仅统计点餐订单（`bill_type = SaleBillTypeInstant`）
   - `order_source_uuid = 0` 归类为"店内"
   - `order_source_uuid > 0` 时，从 `ttpos_order_source` 表关联获取来源名称（通过 `multi_language_name_uuid` 关联多语言名称表）
   - 如果 `order_source` 已被删除或不存在，显示原名称或"未知来源"

3. **桌台方式来源统计**：
   - 仅统计桌台订单（`bill_type = SaleBillTypeDesk`）
   - 从 `ttpos_sale_bill.source` 字段获取来源
   - 映射：`cashier` → "收银机"，`assistant` → "点餐助手"，`tablet` → "平板"，`h5` → "H5"

4. **用餐方式统计**：
   - 统计点餐订单和桌台订单
   - 从 `ttpos_sale_bill.dining_method` 字段获取（0=店内用餐，1=打包）
   - 桌台订单统一归类为"店内用餐"（即使有单品打包，`dining_method` 仍按 0 处理）

### Service 层

#### 新增方法

```go
// main/app/service/business.go

// CountUserAnalysis 统计用户分析数据
func (s *businessSrv) CountUserAnalysis(ctx context.Context) (*resp.UserAnalysisResp, error) {
    // 1. 获取今天时间范围（使用门店时区）
    // 2. 调用 Repository.CountUserAnalysis（返回 decimal.Decimal 类型的 Percentage）
    // 3. 转换响应格式：将 decimal.Decimal 转换为 float64（使用 .InexactFloat64() 或 .Float64()）
    // 4. Repository 层已使用 decimal 计算占比，Service 层只需转换类型
}

// ExportUserAnalysis 导出用户分析统计
func (s *businessSrv) ExportUserAnalysis(ctx context.Context) error {
    // 1. 检查是否有正在导出的任务
    // 2. 获取统计数据
    // 3. 创建导出任务（ExportRecord）
    // 4. 异步处理导出（生成 Excel）
}
```

#### 更新 SaveSale 方法

```go
// main/app/service/statistics.go

func (s *statisticsSrv) SaveSale(ctx context.Context, req SaveSaleReq) error {
    // ... 现有逻辑 ...
    
    sale := model.StatisticsSale{
        // ... 现有字段 ...
        OrderSourceUuid: saleBill.OrderSourceUuid,
        NationalityUuid: saleBill.NationalityUuid, // 新增
    }
    // ... 保存逻辑 ...
}
```

### API 层

```go
// main/app/api/v1/shop/shop_statistics.go

// UserAnalysis 用户分析统计
func (h *statisticsHandler) UserAnalysis(c *gin.Context) {
    ctx := helper.GetContext(c)
    resp, err := h.businessSrv.CountUserAnalysis(ctx)
    if err != nil {
        helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
        return
    }
    helper.Success(c, resp)
}

// ExportUserAnalysis 导出用户分析统计
func (h *statisticsHandler) ExportUserAnalysis(c *gin.Context) {
    ctx := helper.GetContext(c)
    err := h.businessSrv.ExportUserAnalysis(ctx)
    if err != nil {
        helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
        return
    }
    helper.Success(c, nil)
}
```

---

## ⚡ 缓存设计

暂不实现缓存，统计查询固定查询今天数据，数据量可控。

---

## 🚨 错误处理

### 错误场景

#### 场景 1: 统计数据为空

- **处理方式**: 返回空数组，不抛错
- **用户影响**: 前端显示"暂无数据"

#### 场景 2: 导出任务已存在

- **处理方式**: 返回错误提示"正在导出,请稍后再操作"
- **用户影响**: 用户看到错误提示

#### 场景 3: 数据库查询异常

- **处理方式**: 记录错误日志，返回系统错误
- **用户影响**: 用户看到"系统错误，请稍候再试"

---

## 🔒 安全设计

### 身份验证

- **JWT Token**: 所有 API 需要 Token 验证
- **店铺权限**: 仅能查询当前店铺数据

### 数据安全

- **SQL 注入防护**: 使用参数化查询
- **数据管理过滤**: 排除已被数据管理的订单

---

## 🧪 测试策略

### 单元测试

**目标覆盖率**:
- Service 层: 70%+
- Repository 层: 80%+

**测试内容**:
- Repository 统计逻辑（四个维度）
- Service 数据转换和占比计算
- 数据管理过滤逻辑

### API 测试

**测试内容**:
- 查询接口返回格式
- 导出接口任务创建
- 权限校验

### 集成测试

**测试流程**:
- 端到端统计查询
- 导出文件生成和下载

---

## 📈 性能优化

### 优化策略

1. **数据库优化**:
   - 添加索引：`nationality_uuid`、`order_source_uuid`、`source`
   - 使用聚合查询，减少数据库往返

2. **查询优化**:
   - 固定查询今天数据，数据量可控
   - 使用 `GROUP BY` 和聚合函数

### 性能指标

- 查询响应时间: < 500ms
- 导出文件生成: 异步处理，不阻塞用户

---

## 📚 实现清单

### Phase 1: 数据库和模型

- [ ] 创建数据库迁移文件（新增 `nationality_uuid` 字段）
- [ ] 执行数据库迁移
- [ ] 更新 Go Model（`StatisticsSale`）

### Phase 2: 核心实现

- [ ] 更新 `SaveSale` 方法（保存 `nationality_uuid`）
- [ ] 实现 Repository `CountUserAnalysis` 方法
- [ ] 实现 Service `CountUserAnalysis` 方法
- [ ] 实现 Service `ExportUserAnalysis` 方法
- [ ] 创建 DTO 定义

### Phase 3: API 层

- [ ] 实现 API Handler
- [ ] 注册路由

### Phase 4: 导出功能

- [ ] 实现 Excel 导出模板
- [ ] 实现导出任务异步处理

### Phase 5: 测试

- [ ] Repository 单元测试
- [ ] Service 单元测试
- [ ] API 集成测试

**详细任务**: 参见 `tasks.md`

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 在技术方案评审、关键架构决策或踩坑总结后，立即补充 Episode 并在设计文档尾部互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-11-26  
**作者**: 后端开发组  
**审核者**: {审核者}

