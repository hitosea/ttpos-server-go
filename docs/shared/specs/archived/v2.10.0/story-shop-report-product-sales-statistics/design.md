# 新管理端(商家端)-报表中心-商品销售统计 设计文档

> 本文档定义商品销售统计功能的技术设计和实现方案。

## 📋 概述

商品销售统计功能基于现有API实现，通过扩展查询参数和新增导出功能，提供完善的筛选和导出能力。核心实现包括：

- 扩展 `BusinessDataCountProductSalesReq` 请求参数
- 优化 `CountProductSale` 统计查询逻辑
- 新增导出API和异步任务处理

---

## 🎯 规范对齐

### Go Main 规范 (go-main.mdc)

- ✅ 请求参数使用 `form` 和 `json` 标签
- ✅ URL 使用 snake_case: `/shop/statistics/product_sales/export`
- ✅ data 字段返回对象
- ✅ 不使用 panic，返回 error
- ✅ 使用 errors.WithMessage 包装错误

### API 设计规范 (api.mdc)

- ✅ 响应格式统一: `{code, message, data{}}`
- ✅ data 不能为 null 或数组
- ✅ 导出API返回文件下载链接

### 数据库规范 (database.mdc)

- ✅ 复用现有统计表结构
- ✅ 导出记录使用 `ttpos_export_record` 表

---

## 🔄 代码复用分析

### 可复用的现有组件

- **StatisticsService**: `main/app/service/statistics.go` - 复用统计查询逻辑
- **BusinessService**: `main/app/service/business.go` - 复用业务封装逻辑
- **StatisticsRepository**: `main/app/repository/statistics.go` - 复用数据查询逻辑
- **ExportRecord**: `main/app/model/export_record.go` - 复用导出记录模型
- **ExportKitchenProductionDetail**: `main/app/service/business.go` - 参考导出实现

### 集成点

- **统计模块**: 扩展 `CountProductSale` 方法支持新筛选条件
- **导出模块**: 复用导出任务处理机制
- **时间工具**: 复用 `utils.SetTimezone` 处理时区

---

## 🏗️ 架构设计

### 分层设计

```
ProductSalesStatisticsAPI (API 层)
  ↓ 调用
BusinessService.CountProductSales (业务层)
  ↓ 调用
StatisticsService.CountProductSale (统计服务层)
  ↓ 调用
StatisticsRepository.CountProductSale (数据层)
```

**依赖规则**：

- ✅ BusinessService 依赖 StatisticsService 接口
- ✅ StatisticsService 依赖 StatisticsRepository
- ✅ 使用 DBOption 模式支持灵活筛选

---

## 📝 接口设计

### 1. 查询API扩展

**接口路径**: `GET /shop/statistics/product_sales`

**请求参数扩展**:

```go
type BusinessDataCountProductSalesReq struct {
    dto.PageReq
    ProductName    string   `form:"product_name" json:"product_name"`           // 商品名称
    QueryStartTime int64    `form:"query_start_time" json:"query_start_time"`   // 查询开始时间戳
    QueryEndTime   int64    `form:"query_end_time" json:"query_end_time"`       // 查询结束时间戳
    TimeType       int      `form:"time_type" json:"time_type"`                 // 时间类型: 1=今天, 2=昨天, 3=本周, 4=本月, 5=近7天, 6=上月, 7=今年
    AreaUuid       uint64   `form:"area_uuid" json:"area_uuid"`                 // 区域UUID, -1=全部
    CategoryUuids  []uint64  `form:"category_uuids" json:"category_uuids"`      // 分类UUID列表, 空=全部
    OrderType      string   `form:"order_type" json:"order_type"`               // 订单类型: ""=全部, "1"=点餐, "2"=桌台, "3"=外送, 可多选如"1,2,3"
    OrderSource    int      `form:"order_source" json:"order_source"`            // 订单来源: -1=全部, 1=店内, 2=外卖
    SortType       int      `form:"sort_type" json:"sort_type"`                 // 排序类型: 0=默认, 1=按销售数量, 2=按原销售额
    SortDirection  int      `form:"sort_direction" json:"sort_direction"`        // 排序方向: 0=默认, 1=升序, 2=降序
}
```

**响应格式**: 保持不变

```go
type BusinessDataCountProductSalesPagination struct {
    List []BusinessDataCountProductSalesItem `json:"list"`
    Meta dto.PageResponse                    `json:"meta"`
}
```

### 2. 导出API

**接口路径**: `GET /shop/statistics/product_sales/export`

**请求参数**: 与查询API一致

**响应格式**:

```go
// 成功时返回空data，通过异步任务处理
// 失败时返回错误信息
```

---

## 🗄️ 数据模型设计

### 请求参数模型

**文件**: `main/app/dto/req/statistics.go`

```go
// BusinessDataCountProductSalesReq 营业数据商品销售统计列表请求
type BusinessDataCountProductSalesReq struct {
    dto.PageReq
    ProductName    string   `form:"product_name" json:"product_name"`
    QueryStartTime int64    `form:"query_start_time" json:"query_start_time"`
    QueryEndTime   int64    `form:"query_end_time" json:"query_end_time"`
    TimeType       int      `form:"time_type" json:"time_type"`                 // 新增
    AreaUuid       uint64   `form:"area_uuid" json:"area_uuid"`
    CategoryUuid   uint64 `form:"category_uuid" json:"category_uuid"`       
    CategoryUuids  []uint64  `form:"category_uuids" json:"category_uuids"`      // 新增
    OrderType      string   `form:"order_type" json:"order_type"`               // 新增
    OrderSource    int      `form:"order_source" json:"order_source"`            // 新增
    SortType       int      `form:"sort_type" json:"sort_type"`
    SortDirection  int      `form:"sort_direction" json:"sort_direction"`
}
```

### 统计查询参数模型

**文件**: `main/app/service/statistics.go`

```go
// CountReq 统计请求
type CountReq struct {
    QueryStartTime int64
    QueryEndTime   int64
    TimeType       int      // 新增
    RankType       int
    RankDirection  int
    PageNo         int
    PageSize       int
    AreaUuid       uint64
    CategoryUuid   uint64
    CategoryUuids  []uint64  // 新增
    ProductName    string
    OrderTypes     []uint   // 新增：订单类型列表
    OrderSource    int      // 新增：订单来源
}
```

### Repository查询参数模型

**文件**: `main/app/repository/statistics.go`

```go
// CountProductSaleRepoReq 统计商品销售请求
type CountProductSaleRepoReq struct {
    PageNo        int
    PageSize      int
    RankType      int
    RankDirection int
    Language      string
    AreaUuid      uint64
    CategoryUuid uint64
    CategoryUuids []uint64  // 新增
    ProductName   string
    OrderTypes    []uint    // 新增：订单类型列表
    OrderSource   int       // 新增：订单来源
}
```

---

## 🔧 核心实现逻辑

### 1. 时间类型处理

**位置**: `main/app/service/statistics.go` - `buildCountOpts`

```go
// 处理时间范围
if req.TimeType > 0 && req.TimeType <= 7 {
    timezone := ctx.GetCompanySetting().Timezone
    timezoneUtil := utils.SetTimezone(timezone)
    
    switch req.TimeType {
    case 1: // 今天
        queryStartTime, queryEndTime = timezoneUtil.TodayStartEndUnix()
    case 2: // 昨天
        queryStartTime, queryEndTime = timezoneUtil.YesterdayStartEndUnix()
    case 3: // 本周
        queryStartTime, queryEndTime = timezoneUtil.WeekStartEndUnix()
    case 4: // 本月
        queryStartTime, queryEndTime = timezoneUtil.MonthStartEndUnix()
    case 5: // 近7天
        queryStartTime, queryEndTime = timezoneUtil.Last7DaysStartEndUnix() // 需要新增方法
    case 6: // 上月
        queryStartTime, queryEndTime = timezoneUtil.LastMonthStartEndUnix() // 需要新增方法
    case 7: // 今年
        queryStartTime, queryEndTime = timezoneUtil.YearStartEndUnix() // 需要新增方法
    }
}

// 如果同时提供了自定义时间范围，优先使用自定义时间范围
if req.QueryStartTime > 0 && req.QueryEndTime > 0 {
    queryStartTime = req.QueryStartTime
    queryEndTime = req.QueryEndTime
}
```

### 2. 订单类型筛选

**位置**: `main/app/repository/statistics.go` - `CountProductSale`

```go
// 订单类型筛选
if len(req.OrderTypes) > 0 {
    // 映射订单类型到SaleBillType
    var billTypes []uint
    for _, orderType := range req.OrderTypes {
        switch orderType {
        case 1: // 点餐订单
            billTypes = append(billTypes, constant.SaleBillTypeInstant)
        case 2: // 桌台订单
            billTypes = append(billTypes, constant.SaleBillTypeDesk)
        case 3: // 外送订单
            billTypes = append(billTypes, constant.SaleBillTypeTakeout)
        }
    }
    
    if len(billTypes) > 0 {
        // 关联销售账单表进行筛选
        saleBillTable := prefix + "sale_bill as sb"
        db.Joins("LEFT JOIN " + saleBillTable + " ON sp.sale_bill_uuid = sb.uuid")
        db.Where("sb.bill_type IN (?)", billTypes)
    }
}
```

### 3. 订单来源筛选

**位置**: `main/app/repository/statistics.go` - `CountProductSale`

```go
// 订单来源筛选（仅在订单类型包含点餐订单时生效）
if req.OrderSource > 0 && containsOrderType(req.OrderTypes, 1) {
    // 需要关联订单表或销售账单表
    // 订单来源映射：1=店内, 2=外卖
    // 具体实现需要根据数据模型确定
    // 可能需要关联 order 表或 sale_bill 表的 order_source 字段
}
```

### 4. 商品分类多选

**位置**: `main/app/repository/statistics.go` - `CountProductSale`

```go
// 商品分类筛选（支持多选）
if len(req.CategoryUuids) > 0 {
    // 查询所有子分类
    var allCategoryUuids []uint64
    for _, categoryUuid := range req.CategoryUuids {
        allCategoryUuids = append(allCategoryUuids, categoryUuid)
        // 查询子分类
        var subCategoryUuids []uint64
        r.db.Table(productCategoryTable).
            Select("uuid").
            Where("parent_uuid = ?", categoryUuid).
            Pluck("uuid", &subCategoryUuids)
        allCategoryUuids = append(allCategoryUuids, subCategoryUuids...)
    }
    
    db.Where("pp.category_uuid IN (?)", allCategoryUuids)
}
```

### 5. 导出功能实现

**位置**: `main/app/service/business.go`

```go
// ExportProductSales 导出商品销售统计
func (s *businessSrv) ExportProductSales(ctx context.Context, req req.BusinessDataCountProductSalesReq) error {
    // 1. 检查数据量
    req.PageNo = 1
    req.PageSize = 1000
    result, err := s.CountProductSales(ctx, req)
    if err != nil {
        return err
    }
    if result.Meta.Total > 1000 {
        return errors.WithMessage(errors.New("请选择具体时间段，最多可导出1000条以下的数据"))
    }
    if result.Meta.Total == 0 {
        return errors.WithMessage(errors.New("没有数据需要导出"))
    }
    
    // 2. 创建导出任务
    params, err := json.Marshal(req)
    if err != nil {
        return err
    }
    
    fileNameMul := model.MultiLanguageName{
        EnName:   "Product Sales Statistics",
        ZhName:   "商品销售统计",
        ZhTwName: "商品銷售統計",
        // ... 其他语言
    }
    
    // 使用当前时间戳作为文件名后缀
    timestamp := time.Now().Unix()
    fileName := fmt.Sprintf("%s_%d.xlsx", fileNameMul.GetNameByLang(ctx.GetLanguage()), timestamp)
    
    uuid, _ := utils.GetID()
    record := &model.ExportRecord{
        BaseModel:    model.BaseModel{Uuid: uuid},
        ExportType:   model.ExportTypeProductSales, // 需要新增导出类型
        ExportName:   fileName,
        FileUuid:     0,
        Status:       model.ExportStatusPending,
        ErrorMsg:     "",
        ExportParams: string(params),
        StaffUuid:    ctx.GetStaffUuid(),
    }
    
    db := ctx.GetDB()
    if err := repository.NewExportRecordRepo(db).Create(record); err != nil {
        return err
    }
    
    // 3. 异步处理导出任务
    utils.Go(func() {
        _, err := s.ExportProductSalesTask(ctx, req, record.Uuid)
        // 处理错误和状态更新
    })
    
    return nil
}
```

---

## 📊 数据库查询优化

### SQL查询优化

1. **索引使用**: 确保 `statistics_product` 表的 `sale_bill_uuid`, `product_package_uuid`, `desk_uuid` 字段有索引
2. **JOIN优化**: 使用 LEFT JOIN 避免数据丢失
3. **WHERE条件**: 将筛选条件放在WHERE子句中，利用索引

### 性能考虑

- 时间范围筛选：使用索引字段 `complete_time`
- 分类筛选：使用 `IN` 查询，避免多次查询
- 订单类型筛选：关联 `sale_bill` 表，使用 `bill_type` 索引

---

## 🔐 安全设计

### 权限控制

- 导出功能需要商户管理员权限
- 数据隔离：仅能查询当前商户的数据

### 数据验证

- 时间范围验证：不能选择未来时间
- 数据量限制：导出最多1000条数据
- 参数验证：使用 binding 标签验证请求参数

---

## 📦 文件清单

### 需要修改的文件

1. `main/app/dto/req/statistics.go` - 扩展请求参数
2. `main/app/service/statistics.go` - 扩展统计服务
3. `main/app/service/business.go` - 扩展业务服务和新增导出方法
4. `main/app/repository/statistics.go` - 扩展数据查询逻辑
5. `main/app/api/v1/shop/shop_statistics.go` - 新增导出API
6. `main/app/model/export_record.go` - 新增导出类型常量
7. `main/pkg/utils/time.go` - 新增时间工具方法

### 需要新增的文件

1. `main/app/service/business_export_product_sales.go` - 导出任务处理（可选，如果代码量大）

---

## 🧪 测试要点

### 单元测试

- 时间类型计算准确性
- 订单类型筛选逻辑
- 商品分类多选逻辑
- 导出数据格式验证

### 集成测试

- API接口功能测试
- 导出任务流程测试
- 数据准确性验证

---

**版本**: v1.0.0  
**创建日期**: 2025-11-19  
**维护者**: 开发组

