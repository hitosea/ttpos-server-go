# 渠道营业统计 API 设计文档

> 本文档定义渠道营业统计查询与导出的详细技术方案。

## 📋 概述

- 实现两个 Go Main 模块 API：`GET /api/v1/shop/statistics/channel_sales` 与 `GET /api/v1/shop/statistics/channel_sales/export`。
- 复用 `repository.CountSale` 的 SQL/聚合逻辑，抽取公共方法 `CountChannelSale`（暂命名）以按渠道返回合计、堂食、点餐-店内、点餐-外卖、外送的数据。
- 查询 API 返回 JSON 数据用于首页及报表中心；导出 API 调用同一 Service 输出 Excel。

---

## 🎯 规范对齐

- **Go Main**：Controller 只做参数校验，Service 依赖接口，Repository 仅持有 `*gorm.DB`，禁止 panic。
- **API 规范**：URL snake_case，响应 `{code,message,data{}}`，data 不为 null；导出流响应保持 `application/octet-stream`。
- **安全规范**：复用现有店铺权限中间件校验店铺身份；导出操作写入操作日志。

---

## 🔄 代码复用分析

| 组件 | 路径 | 复用策略 |
| --- | --- | --- |
| `CountSale` | `main/app/repository/statistics.go` | 拆出渠道聚合逻辑，产出 map/struct 供 Service 调用 |
| 时间工具 | `pkg/timeutil/company_time.go` (现有) | 继续使用门店时区工具推导今日范围 |
| Excel 工具 | `pkg/excel/exporter.go`（现有导出模块） | 复用通用导出器生成渠道报表 |
| DTO | `main/app/dto/resp/statistics.go` | 若已有渠道结构体则复用，否则新增 `ChannelSalesResp` |

---

## 🏗️ 架构设计

```
API(shop_statistics.go)
  ↓ 依赖 (interface)
Service(business.go)
  ↓ 依赖
Repository(statistics.go CountChannelSale)
```

- **API 层**：新增 `ChannelSales`、`ExportChannelSales` 方法，接收 query 参数并调用 Service。
- **Service 层**：沿用 `IBusinessSrv` 在现有实现中扩展渠道营业统计方法，负责时间默认值、渠道补齐、导出格式组装。
- **Repository 层**：在 `statistics.go` 中新增 `CountChannelSale(ctx context.Context, shopId uint64, startTime, endTime int64) (*ChannelSaleRepoResult, error)`，内部复用 `CountSale` SQL，按 `order_source`/`sale_mode` 聚合。

### 数据结构

```go
type ChannelSaleItem struct {
    TotalOrderNum      int64           `json:"total_order_num"`
    MinOrderAmount     decimal.Decimal `json:"min_order_amount"`
    MaxOrderAmount     decimal.Decimal `json:"max_order_amount"`
    AvgOrderAmount     decimal.Decimal `json:"avg_order_amount"`
    TotalDeskNum       int64           `json:"total_desk_num"`
    TotalMealNum       int64           `json:"total_meal_num"`
}

type ChannelSalesResp struct {
    Summary         ChannelSaleItem `json:"summary"`
    Table           ChannelSaleItem `json:"table"`
    DineIn          ChannelSaleItem `json:"dine_in"`
    TakeoutShop     ChannelSaleItem `json:"takeout_shop"`
    TakeoutDelivery ChannelSaleItem `json:"takeout_delivery"`
    Meta struct {
        StartTime int64 `json:"start_time"`
        EndTime   int64 `json:"end_time"`
    } `json:"meta"`
}
```

Repository 返回 map[string]ChannelSaleItem，由 Service 根据配置补齐，字段命名遵循 `CountSale` 里的 `total_*` / `min/max/avg_*` 约定（例如 `total_order_num`, `min_order_amount`, `max_order_amount`, `avg_order_amount`, `total_desk_num`, `total_meal_num`），以便与现有前端/模型复用同一结构。

---

## 🧩 组件设计

### DTO

- `main/app/dto/req/statistics_channel_req.go`
  ```go
  type ChannelSalesReq struct {
      StartTime int64 `form:"start_time"`
      EndTime   int64 `form:"end_time"`
  }
  ```

- `main/app/dto/resp/statistics_channel_resp.go`
  ```go
  type ChannelSalesResp struct {
      Summary         *ChannelSalesBlock `json:"summary"`
      Table           *ChannelSalesBlock `json:"table"`
      DineIn          *ChannelSalesBlock `json:"dine_in"`
      TakeoutShop     *ChannelSalesBlock `json:"takeout_shop"`
      TakeoutDelivery *ChannelSalesBlock `json:"takeout_delivery"`
      Meta            *ChannelSalesMeta  `json:"meta"`
  }
  ```

### Service

- 接口扩展：`main/app/service/i_business_srv.go`
  ```go
  type IBusinessSrv interface {
      ...
      CountChannelSales(ctx *gin.Context, shopId uint64, req *dto_req.ChannelSalesReq) (*dto_resp.ChannelSalesResp, error)
      ExportChannelSales(ctx *gin.Context, shopId uint64, req *dto_req.ChannelSalesReq) (*bytes.Buffer, string, error)
  }
  ```

- 实现：`main/app/service/business.go`
  - 依赖 `repository.IStatisticsRepo`（新增方法）与 `service.IShopSettingSrv`（获取启用渠道）。
  - 处理默认时间：若 `req.StartTime/EndTime` 为空，调用 `timeutil.GetShopTodayRange(ctx)`。
  - 统一调用 `repo.CountChannelSale`，将缺失渠道填 0；统计方法统一使用 `Count*` 前缀，导出统一使用 `Export*` 前缀，方便与现有业务命名保持一致。
  - 导出：调用查询结果后交给 `pkg/excel` 生成表格（列顺序与示例一致），实现细节对齐 `ExportBusinessPaymentMethod`（沿用相同导出 helper、header 写法与日志记录），保证体验一致。

### Repository

- 接口新增：
  ```go
  type IStatisticsRepo interface {
      CountChannelSale(ctx context.Context, shopId uint64, startTime, endTime int64) (map[string]*ChannelSaleRepoResult, error)
  }
  ```

- 实现：
  - 在 `statistics.go` 内增加 SQL 构建，复用 `CountSale` 的 join/where 条件，仅在 select 中增加 `order_source`/`sale_mode` 分组。
  - 将结果写入 map key：`summary`, `table`, `dine_in`, `takeout_shop`, `takeout_delivery`。

### API

- `main/app/api/v1/shop/shop_statistics.go`
  - 新增 handler `ChannelSales`：
    ```go
    func (api *ShopStatisticsAPI) ChannelSales(c *gin.Context) {
        var req dto_req.ChannelSalesReq
        if err := c.ShouldBindQuery(&req); err != nil { ... }
        resp, err := api.businessSrv.CountChannelSales(c, ctxShopID, &req)
        helper.Success(c, resp)
    }
    ```
  - 新增 handler `ChannelSalesExport`：参数校验后调用 `businessSrv.ExportChannelSales`，设置头 `Content-Disposition`，写 `buffer.Bytes()`。

---

## 🔌 API 设计

### 1. 渠道营业统计查询

- **Method**: `GET /api/v1/shop/statistics/channel_sales`
- **Query**:
  ```json
  {
    "start_time": 1732550400,
    "end_time": 1732636799
  }
  ```
- **Response**:
  ```json
  {
    "code": 1,
    "message": "success",
    "data": {
      "summary": { "total_order_num": 88, "min_order_amount": "88.48", "max_order_amount": "8869.10", "avg_order_amount": "889.15" },
      "table": { "total_desk_num": 12, "total_meal_num": 48, "avg_order_amount": "689.10", ... },
      "dine_in": { ... },
      "takeout_shop": { ... },
      "takeout_delivery": { ... },
      "meta": { "start_time": 1732550400, "end_time": 1732636799 }
    }
  }
  ```

### 2. 渠道营业统计导出

- **Method**: `GET /api/v1/shop/statistics/channel_sales/export`
- **Query**: 与查询接口一致，可选 `format=csv|xlsx`（默认 `xlsx`）。
- **Response**: HTTP 200，Headers:
  ```
  Content-Type: application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
  Content-Disposition: attachment; filename="channel_sales_10001_20251126.xlsx"
  ```
  Body: Excel 二进制流。

---

## 🚨 错误处理

| 场景 | 处理 |
| --- | --- |
| `start_time > end_time` | 返回 `CodeInvalidParam` |
| Repository 查询异常 | 记录 `logger`，返回 `CodeFail` |
| 导出生成失败 | 返回 `CodeFail` 且写业务告警 |

---

## 🔒 安全设计

- API 继续使用 `ShopAuthMiddleware` 校验店铺身份。
- 导出接口写入 `operation_log`（调用现有操作日志 Service）。
- 对导出接口增加限速（可复用 `rate_limiter` 中间件，默认 5 次/分钟）。

---

## 🧪 测试策略

- Repository：使用真实 test DB，覆盖多渠道组合 / 无数据 / 仅单渠道。
- Service：mock repo & shopSetting，测试默认时间、补齐渠道、导出文件生成。
- API：集成测试验证权限、参数、响应结构，导出接口验证返回头信息。

---

## 📈 性能

- Repository SQL 添加 `shop_id + finish_time` 组合索引（已有则复用），确保查询 < 200ms。
- 导出限制最大时间跨度（可在 Service 中校验 ≤ 31 天，超出报错）。

---

## Graphiti & 活动日志

- 如统计口径有新增规范，补充 Graphiti Episode 并在此文档尾部互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-11-26  
**作者**: 技术团队（代）  
**审核者**: 待定


