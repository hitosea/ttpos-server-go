# 收银机-已选订单数据排除 设计文档

> 本文档定义收银机已选订单数据排除功能的技术设计和实现方案。

## 📋 概述

当订单被标记为"已选"（数据管理订单）后，系统需要在所有相关的数据统计和展示场景中，自动排除这些已选订单的数据，确保各项业务数据的准确性和可靠性。

**核心实现思路**：
1. **Handler 层判断数据管理功能是否开启**：在 API Handler 层判断两个条件（公司开启数据管理 + 数据管理功能启用），设置 `ExcludeDataManage` 参数
2. **通过参数传递过滤标志**：使用 `ExcludeDataManage` 参数传递过滤标志，而不是通过 context
3. **Service 层根据参数应用过滤**：在 Service 层根据 `ExcludeDataManage` 参数决定是否添加 `WhereNotInDataManageSubQuery` 过滤条件
4. **Repository 层执行过滤**：Repository 层接收过滤选项（DBOption），执行数据库查询
5. **交班特殊处理**：交班时单独统计已选订单的现金收入，用于调整钱箱余额

**已选订单判断标准**：
- 通过 `ttpos_data_manage` 表判断
- `type = 0` (DataManageTypeOrder)
- `data_uuid` 存储订单的 `sale_bill_uuid`

**数据管理功能开关判断**：
- 在 Handler 层判断两个条件：
  1. `CompanySetting.IsOpenDataManagement()` - 公司是否开启数据管理
  2. `DataManageSetting.IsEnableDataManage` - 数据管理功能是否启用
- 两个条件都满足时，设置 `ExcludeDataManage = true` 参数传递给 Service 层
- 如果任一条件不满足，则 `ExcludeDataManage = false`，不执行过滤逻辑

---

## 🎯 规范对齐

### Go Main 规范 (go-main.mdc)

- Service 只依赖其他 Service 接口
- Repository 只持有 db 实例
- URL 使用 snake_case
- data 字段必须是对象
- 不使用 panic，返回 error
- 所有代码使用中文注释

### API 设计规范 (api.mdc)

- URL 使用 snake_case
- 响应格式统一
- data 不能为 null 或数组

### 数据库规范 (database.mdc)

- 不需要新增表，使用现有的 `ttpos_data_manage` 表
- 已选订单的判断逻辑已存在

---

## 🔄 代码复用分析

### 可复用的现有组件

- **CommonRepo.WhereNotInDataManageSubQuery**: `main/app/repository/common.go` 第 742-751 行
  - 已存在的过滤方法，用于排除数据管理订单
  - 方法签名：`WhereNotInDataManageSubQuery(db *gorm.DB, field string, opts ...DBOption) DBOption`
  - 使用方式：`CommonRepo.WhereNotInDataManageSubQuery(ctx.GetDB(), "sale_bill_uuid", CommonRepo.WhereByType(model.DataManageTypeOrder), CommonRepo.WhereBySoftDelete())`
  - **重要**：需要传入独立的 `db` 参数，避免子查询继承外部查询的上下文

- **订单列表过滤逻辑**: `main/app/repository/order.go` 第 437-444 行
  - `GetCashierOrderListWithPagination` 中已有过滤逻辑示例
  - 参考实现：`WHERE uuid NOT IN (SELECT data_uuid FROM ttpos_data_manage WHERE type = ?)`

- **统计服务**: `main/app/service/statistics.go`
  - `CountSale`: 销售数据统计
  - `CountPayment`: 支付数据统计
  - `CountShiftRefundAmount`: 交班退款金额统计
  - 需要在所有统计方法中添加过滤条件

- **打印记录服务**: `main/app/printer/service/printer_log.go`
  - `GetPrinterLogList`: 获取打印记录列表
  - 需要在查询时排除已选订单

- **交班服务**: `main/app/service/staff_shift.go`
  - `GetShiftInfo`: 获取交班信息
  - `SubmitShift`: 提交交班
  - 需要在统计时排除已选订单

### 集成点

- **数据管理表**: `ttpos_data_manage` - 存储已选订单信息
- **设置服务**: `main/app/service/setting/setting.go` - `GetDataManageSetting` 方法获取数据管理开关
- **统计服务**: `main/app/service/statistics.go` - 所有统计方法需要统一过滤
- **打印记录**: `main/app/printer/service/printer_log.go` - 打印记录查询需要过滤
- **交班统计**: `main/app/service/staff_shift.go` - 交班统计需要过滤

---

## 🏗️ 架构设计

### 分层设计原则

**Go Main 三层架构**:

```
API 层 (Controller/API)
  ↓ 依赖
业务层 (Service)
  ↓ 依赖
数据层 (Repository)
```

**依赖规则**:
- ✅ 上层可依赖下层
- ❌ 禁止下层依赖上层
- ❌ 禁止跨层调用
- ❌ Service 不能依赖 Repository
- ✅ Service 可以依赖其他 Service 接口

### 模块划分

#### Go Main 模块

- **Service 层**: `main/app/service/` - 业务逻辑、统计数据计算
  - `statistics.go` - 统计服务，需要修改所有统计方法
  - `staff_shift.go` - 交班服务，需要修改交班统计
  - `business.go` - 营业数据服务，需要修改营业数据统计
- **Repository 层**: `main/app/repository/` - 数据访问、数据库操作
  - `statistics.go` - 统计 Repository，需要修改所有统计查询
  - `print_log.go` - 打印记录 Repository，需要修改打印记录查询
- **Printer Service 层**: `main/app/printer/service/` - 打印服务
  - `printer_log.go` - 打印记录服务，需要修改打印记录查询

---

## 🗄️ 数据库设计

### 数据表设计

**不需要新增表**，使用现有的 `ttpos_data_manage` 表：

```sql
CREATE TABLE `ttpos_data_manage` (
    `id` bigint unsigned NOT NULL AUTO_INCREMENT,
    `uuid` bigint unsigned NOT NULL DEFAULT 0,
    `type` int(10) NOT NULL DEFAULT 0 COMMENT '数据类型 0订单',
    `data_uuid` bigint(20) unsigned NOT NULL DEFAULT 0 COMMENT '数据UUID',
    `staff_uuid` bigint(20) unsigned NOT NULL DEFAULT 0 COMMENT '员工UUID',
    `create_time` int NOT NULL DEFAULT 0,
    `update_time` int NOT NULL DEFAULT 0,
    `delete_time` int NOT NULL DEFAULT 0,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_uuid` (`uuid`),
    KEY `idx_data_uuid` (`data_uuid`),
    KEY `idx_type` (`type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='数据管理表';
```

**字段说明**:
| 字段 | 类型 | 说明 | 约束 |
|------|------|------|------|
| type | int | 数据类型，0=订单 | DEFAULT 0 |
| data_uuid | bigint unsigned | 订单的 sale_bill_uuid | NOT NULL |
| staff_uuid | bigint unsigned | 操作员工UUID | NOT NULL |

---

## 📊 数据模型

### 已选订单判断逻辑

```go
// main/app/model/data_manage.go
const DataManageTypeOrder = 0 // 数据类型 0订单

// 判断订单是否为已选订单
// 通过查询 ttpos_data_manage 表，type = 0 且 data_uuid = sale_bill_uuid
```

### 过滤方法

```go
// main/app/repository/common.go
// WhereNotInDataManageSubQuery 根据DataManage子查询不包含
func (r *commonRepo) WhereNotInDataManageSubQuery(db2 *gorm.DB, field string, opts ...DBOption) DBOption {
    return func(db *gorm.DB) *gorm.DB {
        // 使用独立的 db2 参数构建子查询，避免继承外部查询的上下文
        subQuery := db2.Model(&model.DataManage{}).Select("data_uuid")
        for _, opt := range opts {
            subQuery = opt(subQuery)
        }
        return db.Where(field+" NOT IN (?)", subQuery)
    }
}

// 使用示例（Service 层）：
if req.ExcludeDataManage {
    opts = append(opts, repository.CommonRepo.WhereNotInDataManageSubQuery(
        ctx.GetDB(),
        "sale_bill_uuid",
        repository.CommonRepo.WhereByType(model.DataManageTypeOrder),
        repository.CommonRepo.WhereBySoftDelete(),
    ))
}
```

### 数据管理功能开关判断（Handler 层）

```go
// main/app/api/v1/cashier/cashier_statistics.go
// 在 Handler 层判断数据管理功能是否开启
func (h *statisticsHandler) CountBusiness(c *gin.Context) {
    ctx := helper.GetContext(c)
    var countReq req.BusinessDataCountReq
    // ... 参数绑定 ...
    
    // 判断数据管理功能是否开启
    companySetting := ctx.GetCompanySetting()
    dataSetting := setting.NewSrvImpl(...).GetDataManageSetting(ctx)
    countReq.ExcludeDataManage = companySetting.IsOpenDataManagement() && dataSetting.IsEnableDataManage
    
    // 调用 Service 层
    businessData, err := h.businessSrv.CountBusiness(ctx, countReq)
    // ...
}
```

### Service 层实现示例

```go
// main/app/service/statistics.go
// 在 Service 层根据参数应用过滤
func (s *statisticsSrv) CountSale(ctx context.Context, req CountReq) CountSaleResp {
    db := ctx.GetDB()
    opts := s.buildCountOpts(ctx, req)
    
    // 根据参数决定是否需要排除数据管理订单
    if req.ExcludeDataManage {
        opts = append(opts, repository.CommonRepo.WhereNotInDataManageSubQuery(
            db,
            "sale_bill_uuid",
            repository.CommonRepo.WhereByType(model.DataManageTypeOrder),
            repository.CommonRepo.WhereBySoftDelete(),
        ))
    }
    
    saleData := repository.NewStatisticsRepo(db).CountSale(opts...)
    // ...
}
```

---

## 🔌 API 设计

### 不需要新增 API

本功能是对现有 API 的增强，不需要新增 API 接口。需要修改的现有 API：

1. **打印记录查询 API**
   - 路径: `/api/v1/printer/log/list` (推测)
   - 修改: 在查询时排除已选订单

2. **交班信息查询 API**
   - 路径: `/api/v1/cashier/shift/info` (推测)
   - 修改: 在统计时排除已选订单

3. **交班提交 API**
   - 路径: `/api/v1/cashier/shift/submit` (推测)
   - 修改: 在统计时排除已选订单

4. **营业数据统计 API**
   - 路径: `/api/v1/cashier/business/data` (推测)
   - 修改: 在所有统计维度中排除已选订单

---

## 🧩 组件和接口

### Service 层修改

#### Statistics Service

**文件**: `main/app/service/statistics.go`

**需要修改的方法**:
- `CountSale`: 销售数据统计
- `CountPayment`: 支付数据统计
- `CountShiftRefundAmount`: 交班退款金额统计
- `CountBusinessTimePeriod`: 营业时间段统计
- `CountBusinessSummary`: 营业汇总统计
- `CountCategory`: 商品分类统计
- `CountTax`: 税费统计
- `CountProduct`: 商品统计
- `CountMemberNum`: 会员数量统计
- `CountUnpaidOrder`: 未结订单统计

**修改方式**: 
1. Service 层接收 `CountReq` 参数，其中包含 `ExcludeDataManage` 字段
2. 如果 `req.ExcludeDataManage = true`，则在统计查询中添加 `WhereNotInDataManageSubQuery` 过滤条件
3. 如果 `req.ExcludeDataManage = false`，则跳过过滤逻辑，提高性能
4. **重要**：调用 `WhereNotInDataManageSubQuery` 时需要传入独立的 `db` 参数（如 `ctx.GetDB()`），避免子查询继承外部查询上下文

**实现示例**:
```go
func (s *statisticsSrv) CountSale(ctx context.Context, req CountReq) CountSaleResp {
    db := ctx.GetDB()
    opts := s.buildCountOpts(ctx, req)
    
    // 根据参数决定是否需要排除数据管理订单
    if req.ExcludeDataManage {
        opts = append(opts, repository.CommonRepo.WhereNotInDataManageSubQuery(
            db,  // 传入独立的 db 参数
            "sale_bill_uuid",
            repository.CommonRepo.WhereByType(model.DataManageTypeOrder),
            repository.CommonRepo.WhereBySoftDelete(),
        ))
    }
    
    saleData := repository.NewStatisticsRepo(db).CountSale(opts...)
    // ...
}
```

#### Staff Shift Service

**文件**: `main/app/service/staff_shift.go`

**需要修改的方法**:
- `GetShiftInfo`: 获取交班信息
  - 调用 `CountSale` 时需要排除已选订单
  - 调用 `CountPayment` 时需要排除已选订单
  - 调用 `CountShiftRefundAmount` 时需要排除已选订单
- `SubmitShift`: 提交交班
  - 调用 `CountSale` 时需要排除已选订单
  - 调用 `CountPayment` 时需要排除已选订单

#### Business Service

**文件**: `main/app/service/business.go`

**需要修改的方法**:
- `CountBusiness`: 统计营业数据 ✅ 已完成
  - 接收 `ExcludeDataManage` 参数
  - 调用 `CountSale` 时传递 `ExcludeDataManage` 参数
  - 调用 `BuildPaymentMethodIncome` 时传递 `ExcludeDataManage` 参数
  - 调用 `CountMemberNum` 时传递 `ExcludeDataManage` 参数
  - 调用 `CountUnpaidOrder` 时传递 `ExcludeDataManage` 参数
  - 调用 `BuildCategoryList` 时传递 `ExcludeDataManage` 参数
  - 调用 `CountTax` 时传递 `ExcludeDataManage` 参数
- `CountHome`: 统计首页 ✅ 已完成
  - 接收 `ExcludeDataManage` 参数并传递给统计方法
- `Printer`: 打印 ✅ 已完成
  - 判断数据管理功能是否开启，传递 `ExcludeDataManage` 参数
- `BuildPaymentMethodIncome`: 构建支付方式收入 ✅ 已完成
  - 接收 `ExcludeDataManage` 参数并传递给 `CountPayment` 和 `CountFreePayment`
- `BuildCategoryList`: 构建分类列表 ✅ 已完成
  - 接收 `ExcludeDataManage` 参数并传递给 `CountCategory`

#### Printer Log Service

**文件**: `main/app/printer/service/printer_log.go`

**需要修改的方法**:
- `GetPrinterLogList`: 获取打印记录列表 ✅ 已完成
  - 判断数据管理功能是否开启（公司开启 + 功能启用）
  - 在查询选项中添加过滤条件，使用子查询排除已选订单的打印记录
  - 实现方式：`WHERE related_uuid NOT IN (SELECT data_uuid FROM ttpos_data_manage WHERE type = ? AND delete_time = 0)`

### Repository 层修改

#### Statistics Repository

**文件**: `main/app/repository/statistics.go`

**需要修改的方法**:
- 所有统计查询方法都需要支持可选的 `WhereNotInDataManageSubQuery` 过滤条件
- 确保在查询 `ttpos_sale_bill` 时排除已选订单

**修改方式**:
- Repository 层接收 Service 层传递的过滤选项（DBOption）
- Service 层根据 `ExcludeDataManage` 参数决定是否添加过滤条件
- Repository 层只负责执行查询，不判断开关状态
- **重要**：`WhereNotInDataManageSubQuery` 方法需要传入独立的 `db` 参数，避免子查询继承外部查询上下文

**修改示例**:

```go
// Repository 层：接收过滤选项
func (r *StatisticsRepo) CountSale(opts ...DBOption) model.StatisticsSaleData {
    var result model.StatisticsSaleData
    db := r.db
    
    // 应用所有过滤选项（包括可选的已选订单过滤）
    for _, opt := range opts {
        db = opt(db)
    }
    
    // 执行统计查询...
    return result
}

// Service 层：根据参数传递过滤条件
func (s *statisticsSrv) CountSale(ctx context.Context, req CountReq) CountSaleResp {
    db := ctx.GetDB()
    opts := s.buildCountOpts(ctx, req)
    
    // 根据参数决定是否需要排除数据管理订单
    if req.ExcludeDataManage {
        opts = append(opts, repository.CommonRepo.WhereNotInDataManageSubQuery(
            db,  // 传入独立的 db 参数
            "sale_bill_uuid",
            repository.CommonRepo.WhereByType(model.DataManageTypeOrder),
            repository.CommonRepo.WhereBySoftDelete(),
        ))
    }
    
    // 调用 Repository，传递过滤选项
    saleData := repository.NewStatisticsRepo(db).CountSale(opts...)
    // ...
}
```

#### Print Log Repository

**文件**: `main/app/repository/print_log.go`

**需要修改的方法**:
- `PaginateGet`: 分页获取打印记录
- `GetPrinterLogList`: 获取打印记录列表

**修改方式**: 
1. Service 层判断数据管理功能是否开启（公司开启 + 功能启用）
2. 如果开启，在查询选项中添加过滤条件，排除已选订单的打印记录
3. 如果未开启，跳过过滤逻辑
4. Repository 层接收过滤选项并执行查询

**实现示例**:
```go
// Service 层
func (s *printerLogSrv) GetPrinterLogList(ctx context.Context, req req.PrinterListReq) (*resp.PrinterListPaginationResp, error) {
    // 判断数据管理功能是否开启
    companySetting := ctx.GetCompanySetting()
    dataSetting := s.settingSrv.GetDataManageSetting(ctx)
    excludeDataManage := companySetting.IsOpenDataManagement() && dataSetting.IsEnableDataManage
    
    queryOpts := []repository.DBOption{
        printerLogRepo.WithPrinter(),
        printerLogRepo.WithPrinterPrinterType(),
        printerLogRepo.WithSaleOrder(),
        printerLogRepo.WithSaleBill(),
    }
    
    // 只有开启数据管理功能时才添加过滤条件
    if excludeDataManage {
        queryOpts = append(queryOpts, func(db *gorm.DB) *gorm.DB {
            return db.Where("related_uuid NOT IN (SELECT data_uuid FROM ttpos_data_manage WHERE type = ? AND delete_time = 0)", 
                model.DataManageTypeOrder)
        })
    }
    
    // 调用 Repository
    printerLogList, total, err := printerLogRepo.PaginateGet(req.PageNo, req.PageSize, queryOpts...)
    // ...
}
```

---

## ⚡ 缓存设计

### 不需要新增缓存

本功能主要是数据过滤，不需要新增缓存机制。

---

## 🚨 错误处理

### 错误场景

#### 场景 1: 统计查询失败

- **处理方式**: 记录错误日志，返回错误信息
- **用户影响**: 统计数据显示错误或为空
- **代码示例**:
  ```go
  if err != nil {
      logger.Logger.Error("统计查询失败", zap.Error(err))
      return nil, errors.WithMessage(err, "统计查询失败")
  }
  ```

#### 场景 2: 已选订单数据不一致

- **处理方式**: 确保所有统计场景使用统一的过滤逻辑
- **用户影响**: 数据统计不准确
- **缓解措施**: 统一使用 `WhereNotInDataManageSubQuery` 方法

---

## 🔒 安全设计

### 身份验证

- **JWT Token**: 所有 API 需要 Token 验证（现有机制）

### 权限控制

- **数据管理权限**: 只有有权限的员工才能标记订单为已选（现有机制）

### 数据安全

- **SQL 注入防护**: 使用参数化查询（GORM 自动处理）
- **数据一致性**: 确保所有统计场景使用统一的过滤逻辑

---

## 🧪 测试策略

### 单元测试

**目标覆盖率**:
- main/app/service: 70%+
- main/app/repository: 80%+
- **Payment/Order 相关: 100%**（高风险）

**测试内容**:
- Service 业务逻辑（统计计算）
- Repository 数据访问（过滤逻辑）
- 已选订单过滤正确性

**测试用例**:
1. 有已选订单时，统计结果应排除已选订单
2. 无已选订单时，统计结果不变
3. 交班统计应排除已选订单
4. 打印记录应排除已选订单
5. 营业数据各维度统计应排除已选订单

### API 测试

**测试内容**:
- API 接口调用
- 响应格式
- 数据准确性

### 集成测试

**测试流程**:
- 端到端业务流程
- 数据一致性验证
- 各种业务场景测试

---

## 📈 性能优化

### 优化策略

1. **功能开关判断优化**:
   - **在 Handler 层判断数据管理功能是否开启**：判断两个条件（公司开启数据管理 + 数据管理功能启用）
   - 如果未开启，则 `ExcludeDataManage = false`，Service 层跳过过滤逻辑，避免不必要的子查询
   - 通过 `SettingService.GetDataManageSetting(ctx)` 获取设置（可能有缓存）
   - 只有在两个条件都满足时才设置 `ExcludeDataManage = true`
   
2. **子查询上下文隔离**:
   - `WhereNotInDataManageSubQuery` 方法使用独立的 `db` 参数构建子查询，避免继承外部查询的上下文
   - 这样可以确保子查询构建正确，不会因为外部查询的表别名、JOIN 等导致 SQL 错误

2. **数据库优化**:
   - 确保 `ttpos_data_manage` 表有索引（`idx_data_uuid`, `idx_type`）
   - 使用子查询优化，避免多次 JOIN

3. **查询优化**:
   - 使用 `NOT IN` 子查询，性能优于 `LEFT JOIN ... WHERE ... IS NULL`
   - 确保子查询使用索引
   - 当数据管理功能未开启时，不执行子查询，提高性能

4. **统计查询优化**:
   - 在统计查询中统一添加过滤条件，避免重复查询
   - 通过功能开关判断，避免不必要的过滤开销

### 性能指标

- 本地响应时间: < 200ms
- 数据库查询: < 50ms
- 统计查询: < 100ms
- **数据管理功能未开启时**: 查询性能不受影响（无额外过滤开销）

---

## 🌐 浏览器兼容性

### 前端兼容性（Vue）

- Chrome 90+
- Safari 14+
- Firefox 88+
- Edge 90+

**注意**: 本功能主要是后端修改，前端不需要修改。

---

## 📚 实现清单

### Phase 1: API Handler 层修改

- [x] 修改 `cashier_statistics.go` - 在 Handler 层判断数据管理功能并设置参数
  - `CountBusiness` - ✅ 已完成
  - `CountPaymentMethod` - ✅ 已完成
  - `CountProductCategory` - ✅ 已完成
  - `CountProduct` - ✅ 已完成

### Phase 2: DTO 层修改

- [x] 修改 `BusinessDataCountReq` - 添加 `ExcludeDataManage` 字段 ✅ 已完成
- [x] 修改 `CountReq` - 添加 `ExcludeDataManage` 和 `OnlyDataManage` 字段 ✅ 已完成

### Phase 3: Repository 层修改

- [x] 修改 `WhereNotInDataManageSubQuery` - 添加 `db` 参数，避免上下文污染 ✅ 已完成
- [x] 修改 `WhereInDataManageSubQuery` - 添加 `db` 参数 ✅ 已完成
- [x] 添加辅助方法：`WhereInSaleBillUuids`, `WhereByRelatedOrderType`, `WhereNotInRelatedOrderUuids` ✅ 已完成
- [x] 添加 `DataManageRepo.GetDataUuids` 方法 ✅ 已完成
- [x] 添加 `OrderRepo.GetSaleOrderUuids` 方法 ✅ 已完成

### Phase 4: Service 层修改（统计服务）

- [x] 修改 `StatisticsService.CountSale` - 添加已选订单过滤 ✅ 已完成
- [x] 修改 `StatisticsService.CountPayment` - 添加已选订单过滤（支持 `ExcludeDataManage` 和 `OnlyDataManage`）✅ 已完成
- [x] 修改 `StatisticsService.CountTax` - 添加已选订单过滤 ✅ 已完成
- [x] 修改 `StatisticsService.CountCategory` - 添加已选订单过滤 ✅ 已完成
- [x] 修改 `StatisticsService.CountUnpaidOrder` - 添加已选订单过滤 ✅ 已完成
- [ ] 修改 `StatisticsService.CountShiftRefundAmount` - 添加已选订单过滤（已在 Service 层实现）
- [ ] 修改 `StatisticsService.CountBusinessTimePeriod` - 添加已选订单过滤
- [ ] 修改 `StatisticsService.CountBusinessSummary` - 添加已选订单过滤
- [ ] 修改 `StatisticsService.CountProduct` - 添加已选订单过滤（已在 `buildCountOpts` 中实现）
- [ ] 修改 `StatisticsService.CountMemberNum` - 添加已选订单过滤（已在 `buildCountOpts` 中实现）

### Phase 5: Service 层修改（营业数据服务）

- [x] 修改 `BusinessService.CountBusiness` - 传递 `ExcludeDataManage` 参数 ✅ 已完成
- [x] 修改 `BusinessService.CountHome` - 传递 `ExcludeDataManage` 参数 ✅ 已完成
- [x] 修改 `BusinessService.Printer` - 传递 `ExcludeDataManage` 参数 ✅ 已完成
- [x] 修改 `BusinessService.BuildPaymentMethodIncome` - 传递 `ExcludeDataManage` 参数 ✅ 已完成
- [x] 修改 `BusinessService.BuildCategoryList` - 传递 `ExcludeDataManage` 参数 ✅ 已完成

### Phase 6: Service 层修改（交班服务）

- [x] 修改 `StaffShiftService.GetShiftInfo` - 添加数据管理过滤，特殊处理已选订单现金收入 ✅ 已完成
- [x] 修改 `StaffShiftService.SubmitShift` - 添加数据管理过滤 ✅ 已完成

### Phase 7: Service 层修改（打印服务）

- [x] 修改 `PrinterLogService.GetPrinterLogList` - 添加已选订单过滤 ✅ 已完成

### Phase 8: Service 层修改（数据管理服务）

- [x] 修改 `DataManageService.GetDataManage` - 更新方法调用方式 ✅ 已完成

### Phase 5: 测试

- [ ] 单元测试：统计服务测试
- [ ] 单元测试：Repository 测试
- [ ] 集成测试：交班流程测试
- [ ] 集成测试：营业数据统计测试
- [ ] 集成测试：打印记录查询测试

**详细任务**: 参见 `tasks.md`

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 在技术方案评审、关键架构决策或踩坑总结后，立即补充 Episode 并在设计文档尾部互链。

---

**版本**: v1.1.0  
**创建日期**: 2025-12-09  
**最后更新**: 2025-12-10  
**作者**: 王昱  
**审核者**: {审核者}

---

## 📝 实现记录

### 2025-12-10 实现总结

**实现方式**：
- ✅ 采用参数传递方式，而非 context 传递
- ✅ Handler 层判断数据管理功能是否开启（公司开启 + 功能启用）
- ✅ Service 层根据 `ExcludeDataManage` 参数应用过滤
- ✅ Repository 层方法签名修改，添加独立的 `db` 参数避免上下文污染

**已完成功能**：
- ✅ API Handler 层：4 个统计 API 已实现
- ✅ DTO 层：添加 `ExcludeDataManage` 字段
- ✅ Repository 层：修改方法签名，添加辅助方法
- ✅ Service 层：核心统计方法已实现（CountSale, CountPayment, CountTax, CountCategory, CountUnpaidOrder）
- ✅ 交班服务：已实现，包括已选订单现金收入特殊处理
- ✅ 打印服务：已实现
- ✅ 营业数据服务：已实现

**关键实现细节**：
1. **交班特殊处理**：`GetShiftInfo` 方法中单独统计已选订单的现金收入，计算钱箱余额时减去已选订单现金收入
2. **子查询上下文隔离**：`WhereNotInDataManageSubQuery` 使用独立的 `db` 参数，避免继承外部查询上下文
3. **性能优化**：仅在数据管理功能开启时执行过滤，未开启时无额外开销

**测试状态**（2025-12-10）：
- ✅ 手动测试已完成
- ✅ 统计功能：已选订单正确排除，统计结果准确
- ✅ 交班功能：交班统计正确排除已选订单，钱箱金额计算正确
- ✅ 打印记录：已选订单的打印记录正确排除
- ✅ 营业数据：各维度统计正确排除已选订单

