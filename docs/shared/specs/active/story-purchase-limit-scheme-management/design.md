# 品牌采购限购方案管理 技术设计文档

> 本文档定义品牌采购限购方案管理功能的技术设计和实现方案（Phase 1: 限购方案 CRUD + 数据迁移）。

## 📋 基本信息

| 项目 | 内容 |
|------|------|
| **Spec ID** | story-purchase-limit-scheme-management |
| **设计人** | weifashi |
| **设计日期** | 2026-01-20 |
| **总 SP** | 3 |
| **预计时间** | 3.5-4 天 |

---

## 🔄 代码复用分析

### 可复用的现有组件

| 组件 | 路径 | 复用方式 |
|------|------|---------|
| **PurchaseOrderService** | `main/app/service/purchase_order/purchase_order.go` | 参考服务结构和初始化模式 |
| **PurchaseQuotaService** | `main/app/service/purchase_order/purchase_quota.go` | 参考限购校验逻辑 |
| **PurchaseQuotaConfigRepo** | `main/app/repository/purchase_quota_config.go` | 参考 Repository 设计模式 |
| **PurchaseHandler** | `main/app/api/v1/shop/shop_purchase.go` | 参考 API Handler 设计模式 |

### 需要新建的组件

| 组件 | 路径 | 说明 |
|------|------|------|
| **PurchaseLimitSchemeService** | `main/app/service/purchase_order/purchase_limit_scheme.go` | 限购方案管理服务 |
| **PurchaseLimitSchemeRepo** | `main/app/repository/purchase_limit_scheme_repo.go` | 限购方案 Repository |
| **PurchaseLimitSchemeItemRepo** | `main/app/repository/purchase_limit_scheme_item_repo.go` | 物品配置 Repository |
| **PurchaseLimitSchemeShopRepo** | `main/app/repository/purchase_limit_scheme_shop_repo.go` | 门店配置 Repository |
| **PurchaseLimitSchemeWeekdayRepo** | `main/app/repository/purchase_limit_scheme_weekday_repo.go` | 星期配置 Repository |
| **4 个 Model 文件** | `main/app/model/purchase_limit_scheme*.go` | 数据模型 |
| **数据迁移脚本** | `admin/database/migrations/{timestamp}_migrate_purchase_quota_to_limit_scheme.php` | 旧表迁移 |

---

## 🏗️ 架构设计

### 分层设计

```
API 层 (shop_purchase.go)
  ↓ 调用
Service 层 (purchase_limit_scheme.go)
  ↓ 调用
Repository 层 (purchase_limit_scheme_repo.go + 3个关联表 Repo)
  ↓ 操作
Database (4个新表)
```

### 模块划分

- **API 层**: `main/app/api/v1/shop/shop_purchase.go` - 新增 5 个限购方案接口
- **Service 层**: `main/app/service/purchase_order/purchase_limit_scheme.go` - 限购方案管理业务逻辑
- **Repository 层**: `main/app/repository/purchase_limit_scheme_*.go` - 数据访问（4 个 Repo）
- **Model 层**: `main/app/model/purchase_limit_scheme*.go` - 数据模型（4 个 Model）
- **DTO 层**: 
  - `main/app/dto/req/purchase_limit_scheme_req.go` - 请求参数
  - `main/app/dto/resp/purchase_limit_scheme_resp.go` - 响应数据

---

## 🗄️ 数据库设计

### 新建表结构

#### 表 1: ttpos_purchase_limit_scheme（限购方案主表）

```sql
CREATE TABLE IF NOT EXISTS `ttpos_purchase_limit_scheme` (
    `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `uuid` bigint unsigned NOT NULL DEFAULT 0 COMMENT '唯一标识',
    `name` varchar(50) NOT NULL DEFAULT '' COMMENT '方案名称',
    `status` tinyint NOT NULL DEFAULT 1 COMMENT '状态：0=关闭，1=开启',
    `apply_to_all_shops` tinyint NOT NULL DEFAULT 1 COMMENT '是否适用所有门店：0=否，1=是',
    `daily_limit` int NOT NULL DEFAULT 0 COMMENT '每日申请次数限制（0=不限制）',
    `create_time` int NOT NULL DEFAULT 0 COMMENT '创建时间',
    `update_time` int NOT NULL DEFAULT 0 COMMENT '更新时间',
    `delete_time` int NOT NULL DEFAULT 0 COMMENT '删除时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_uuid` (`uuid`),
    KEY `idx_status` (`status`, `delete_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='限购方案表';
```

#### 表 2: ttpos_purchase_limit_scheme_item（物品配置表）

```sql
CREATE TABLE IF NOT EXISTS `ttpos_purchase_limit_scheme_item` (
    `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `uuid` bigint unsigned NOT NULL DEFAULT 0 COMMENT '唯一标识',
    `scheme_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT '限购方案ID',
    `material_code` varchar(50) NOT NULL DEFAULT '' COMMENT '物品编码',
    `unit_code` varchar(20) NOT NULL DEFAULT '' COMMENT '单位编码',
    `quota_limit` decimal(20,8) NOT NULL DEFAULT 0 COMMENT '限购数量（0=不限制）',
    `create_time` int NOT NULL DEFAULT 0 COMMENT '创建时间',
    `update_time` int NOT NULL DEFAULT 0 COMMENT '更新时间',
    `delete_time` int NOT NULL DEFAULT 0 COMMENT '删除时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_uuid` (`uuid`),
    KEY `idx_scheme_material` (`scheme_id`, `material_code`, `delete_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='限购方案物品配置表';
```

#### 表 3: ttpos_purchase_limit_scheme_shop（门店配置表）

```sql
CREATE TABLE IF NOT EXISTS `ttpos_purchase_limit_scheme_shop` (
    `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `uuid` bigint unsigned NOT NULL DEFAULT 0 COMMENT '唯一标识',
    `scheme_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT '限购方案ID',
    `company_uuid` bigint unsigned NOT NULL DEFAULT 0 COMMENT '门店UUID',
    `create_time` int NOT NULL DEFAULT 0 COMMENT '创建时间',
    `update_time` int NOT NULL DEFAULT 0 COMMENT '更新时间',
    `delete_time` int NOT NULL DEFAULT 0 COMMENT '删除时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_uuid` (`uuid`),
    KEY `idx_scheme_company` (`scheme_id`, `company_uuid`, `delete_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='限购方案门店配置表';
```

#### 表 4: ttpos_purchase_limit_scheme_weekday（星期配置表）

```sql
CREATE TABLE IF NOT EXISTS `ttpos_purchase_limit_scheme_weekday` (
    `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `uuid` bigint unsigned NOT NULL DEFAULT 0 COMMENT '唯一标识',
    `scheme_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT '限购方案ID',
    `weekday` tinyint NOT NULL DEFAULT 0 COMMENT '星期：1=周一，7=周日',
    `create_time` int NOT NULL DEFAULT 0 COMMENT '创建时间',
    `update_time` int NOT NULL DEFAULT 0 COMMENT '更新时间',
    `delete_time` int NOT NULL DEFAULT 0 COMMENT '删除时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_uuid` (`uuid`),
    KEY `idx_scheme_weekday` (`scheme_id`, `weekday`, `delete_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='限购方案星期配置表';
```

### 数据迁移脚本

**迁移文件**: `admin/database/migrations/{timestamp}_migrate_purchase_quota_to_limit_scheme.php`

**迁移逻辑**：

1. **创建 4 个新表**
2. **迁移数据**：
   - 从 `ttpos_purchase_quota_config` 迁移到 `ttpos_purchase_limit_scheme` 和 `ttpos_purchase_limit_scheme_item`
   - 从 `ttpos_purchase_quota_config_shop` 迁移到 `ttpos_purchase_limit_scheme_shop`
   - 默认周期：全周（周一到周日）
3. **删除旧表**：
   - 删除 `ttpos_purchase_quota_config`
   - 删除 `ttpos_purchase_quota_config_shop`
4. **使用事务**保证原子性
5. **备份数据**到日志文件

---

## 📊 数据模型

### Go Model

```go
// main/app/model/purchase_limit_scheme.go
type PurchaseLimitScheme struct {
    Id                uint64 `gorm:"column:id;primaryKey" json:"id"`
    Uuid              uint64 `gorm:"column:uuid;uniqueIndex" json:"uuid"`
    Name              string `gorm:"column:name" json:"name"`
    Status            int8   `gorm:"column:status" json:"status"`
    ApplyToAllShops   int8   `gorm:"column:apply_to_all_shops" json:"apply_to_all_shops"`
    DailyLimit        int    `gorm:"column:daily_limit" json:"daily_limit"`
    CreateTime        int64  `gorm:"column:create_time" json:"create_time"`
    UpdateTime        int64  `gorm:"column:update_time" json:"update_time"`
    DeleteTime        int64  `gorm:"column:delete_time;index" json:"delete_time"`
}

func (*PurchaseLimitScheme) TableName() string {
    return "ttpos_purchase_limit_scheme"
}
```

```go
// main/app/model/purchase_limit_scheme_item.go
type PurchaseLimitSchemeItem struct {
    Id           uint64  `gorm:"column:id;primaryKey" json:"id"`
    Uuid         uint64  `gorm:"column:uuid;uniqueIndex" json:"uuid"`
    SchemeId     uint64  `gorm:"column:scheme_id" json:"scheme_id"`
    MaterialCode string  `gorm:"column:material_code" json:"material_code"`
    UnitCode     string  `gorm:"column:unit_code" json:"unit_code"`
    QuotaLimit   float64 `gorm:"column:quota_limit" json:"quota_limit"`
    CreateTime   int64   `gorm:"column:create_time" json:"create_time"`
    UpdateTime   int64   `gorm:"column:update_time" json:"update_time"`
    DeleteTime   int64   `gorm:"column:delete_time;index" json:"delete_time"`
}

func (*PurchaseLimitSchemeItem) TableName() string {
    return "ttpos_purchase_limit_scheme_item"
}
```

```go
// main/app/model/purchase_limit_scheme_shop.go
type PurchaseLimitSchemeShop struct {
    Id          uint64 `gorm:"column:id;primaryKey" json:"id"`
    Uuid        uint64 `gorm:"column:uuid;uniqueIndex" json:"uuid"`
    SchemeId    uint64 `gorm:"column:scheme_id" json:"scheme_id"`
    CompanyUuid uint64 `gorm:"column:company_uuid" json:"company_uuid"`
    CreateTime  int64  `gorm:"column:create_time" json:"create_time"`
    UpdateTime  int64  `gorm:"column:update_time" json:"update_time"`
    DeleteTime  int64  `gorm:"column:delete_time;index" json:"delete_time"`
}

func (*PurchaseLimitSchemeShop) TableName() string {
    return "ttpos_purchase_limit_scheme_shop"
}
```

```go
// main/app/model/purchase_limit_scheme_weekday.go
type PurchaseLimitSchemeWeekday struct {
    Id         uint64 `gorm:"column:id;primaryKey" json:"id"`
    Uuid       uint64 `gorm:"column:uuid;uniqueIndex" json:"uuid"`
    SchemeId   uint64 `gorm:"column:scheme_id" json:"scheme_id"`
    Weekday    int8   `gorm:"column:weekday" json:"weekday"`
    CreateTime int64  `gorm:"column:create_time" json:"create_time"`
    UpdateTime int64  `gorm:"column:update_time" json:"update_time"`
    DeleteTime int64  `gorm:"column:delete_time;index" json:"delete_time"`
}

func (*PurchaseLimitSchemeWeekday) TableName() string {
    return "ttpos_purchase_limit_scheme_weekday"
}
```

### DTO 定义

#### Request DTO

```go
// main/app/dto/req/purchase_limit_scheme_req.go
package req

type PurchaseLimitSchemeCreateReq struct {
    Name            string                         `json:"name" binding:"required,max=50"`
    Status          int8                           `json:"status"`
    ApplyToAllShops int8                           `json:"apply_to_all_shops"`
    DailyLimit      int                            `json:"daily_limit"`
    Weekdays        []int8                         `json:"weekdays" binding:"required,min=1"`
    Items           []PurchaseLimitSchemeItemReq   `json:"items" binding:"required,min=1"`
    Shops           []uint64                       `json:"shops"`
}

type PurchaseLimitSchemeItemReq struct {
    MaterialCode string  `json:"material_code" binding:"required"`
    UnitCode     string  `json:"unit_code" binding:"required"`
    QuotaLimit   float64 `json:"quota_limit"`
}

type PurchaseLimitSchemeUpdateReq struct {
    Uuid            uint64                         `json:"uuid" binding:"required"`
    Name            string                         `json:"name" binding:"required,max=50"`
    Status          int8                           `json:"status"`
    ApplyToAllShops int8                           `json:"apply_to_all_shops"`
    DailyLimit      int                            `json:"daily_limit"`
    Weekdays        []int8                         `json:"weekdays" binding:"required,min=1"`
    Items           []PurchaseLimitSchemeItemReq   `json:"items" binding:"required,min=1"`
    Shops           []uint64                       `json:"shops"`
}

type PurchaseLimitSchemeDetailReq struct {
    Uuid uint64 `json:"uuid" form:"uuid" binding:"required"`
}

type PurchaseLimitSchemeListReq struct {
    PageNo   int    `json:"page_no" form:"page_no" binding:"required,min=1"`
    PageSize int    `json:"page_size" form:"page_size" binding:"required,min=1,max=100"`
    Status   *int8  `json:"status" form:"status"`
    Name     string `json:"name" form:"name"`
}

type PurchaseLimitSchemeDeleteReq struct {
    Uuid uint64 `json:"uuid" form:"uuid" binding:"required"`
}
```

#### Response DTO

```go
// main/app/dto/resp/purchase_limit_scheme_resp.go
package resp

type PurchaseLimitSchemeResp struct {
    Uuid            uint64                           `json:"uuid"`
    Name            string                           `json:"name"`
    Status          int8                             `json:"status"`
    ApplyToAllShops int8                             `json:"apply_to_all_shops"`
    DailyLimit      int                              `json:"daily_limit"`
    Weekdays        []int8                           `json:"weekdays"`
    Items           []PurchaseLimitSchemeItemResp    `json:"items"`
    Shops           []uint64                         `json:"shops"`
    CreateTime      int64                            `json:"create_time"`
    UpdateTime      int64                            `json:"update_time"`
}

type PurchaseLimitSchemeItemResp struct {
    MaterialCode string  `json:"material_code"`
    UnitCode     string  `json:"unit_code"`
    QuotaLimit   float64 `json:"quota_limit"`
}

type PurchaseLimitSchemeListResp struct {
    List []PurchaseLimitSchemeSummaryResp `json:"list"`
    Meta PageMeta                          `json:"meta"`
}

type PurchaseLimitSchemeSummaryResp struct {
    Uuid       uint64 `json:"uuid"`
    Name       string `json:"name"`
    Status     int8   `json:"status"`
    WeekdayStr string `json:"weekday_str"` // "周一、周三、周五"
    ShopCount  int    `json:"shop_count"`  // 门店数量
    ItemCount  int    `json:"item_count"`  // 物品数量
    DailyLimit int    `json:"daily_limit"`
    CreateTime int64  `json:"create_time"`
    UpdateTime int64  `json:"update_time"`
}
```

---

## 🔌 API 设计

### API 1: 获取限购方案列表

- **URL**: `/shop/purchase/limit_scheme/list`
- **Method**: `GET`
- **请求参数**: `PurchaseLimitSchemeListReq`
  ```json
  {
    "page_no": 1,
    "page_size": 20,
    "status": 1,
    "name": "周末限购"
  }
  ```
- **响应**:
  ```json
  {
    "code": 1,
    "message": "success",
    "data": {
      "list": [
        {
          "uuid": 123456,
          "name": "周末限购方案",
          "status": 1,
          "weekday_str": "周六、周日",
          "shop_count": 10,
          "item_count": 45,
          "daily_limit": 5,
          "create_time": 1705708800,
          "update_time": 1705708800
        }
      ],
      "meta": {
        "page_no": 1,
        "page_size": 20,
        "total": 5
      }
    }
  }
  ```

### API 2: 获取限购方案详情

- **URL**: `/shop/purchase/limit_scheme/detail`
- **Method**: `GET`
- **请求参数**: `uuid=123456`
- **响应**:
  ```json
  {
    "code": 1,
    "message": "success",
    "data": {
      "uuid": 123456,
      "name": "周末限购方案",
      "status": 1,
      "apply_to_all_shops": 0,
      "daily_limit": 5,
      "weekdays": [6, 7],
      "items": [
        {
          "material_code": "ITEM001",
          "unit_code": "箱",
          "quota_limit": 10.0
        }
      ],
      "shops": [111, 222, 333],
      "create_time": 1705708800,
      "update_time": 1705708800
    }
  }
  ```

### API 3: 创建限购方案

- **URL**: `/shop/purchase/limit_scheme/create`
- **Method**: `POST`
- **请求Body**: `PurchaseLimitSchemeCreateReq`
  ```json
  {
    "name": "周末限购方案",
    "status": 1,
    "apply_to_all_shops": 0,
    "daily_limit": 5,
    "weekdays": [6, 7],
    "items": [
      {
        "material_code": "ITEM001",
        "unit_code": "箱",
        "quota_limit": 10.0
      }
    ],
    "shops": [111, 222, 333]
  }
  ```
- **响应**:
  ```json
  {
    "code": 1,
    "message": "创建成功",
    "data": {
      "uuid": 123456
    }
  }
  ```

### API 4: 更新限购方案

- **URL**: `/shop/purchase/limit_scheme/update`
- **Method**: `POST`
- **请求Body**: `PurchaseLimitSchemeUpdateReq`（同创建，增加 uuid 字段）
- **响应**:
  ```json
  {
    "code": 1,
    "message": "更新成功",
    "data": {}
  }
  ```

### API 5: 删除限购方案

- **URL**: `/shop/purchase/limit_scheme/delete`
- **Method**: `DELETE`
- **请求参数**: `uuid=123456`
- **响应**:
  ```json
  {
    "code": 1,
    "message": "删除成功",
    "data": {}
  }
  ```

---

## 🧩 组件和接口

### Service 层

```go
// main/app/service/purchase_order/purchase_limit_scheme.go
type IPurchaseLimitSchemeSrv interface {
    Create(ctx context.Context, req req.PurchaseLimitSchemeCreateReq) (uint64, error)
    Update(ctx context.Context, req req.PurchaseLimitSchemeUpdateReq) error
    GetByUuid(ctx context.Context, uuid uint64) (*resp.PurchaseLimitSchemeResp, error)
    GetList(ctx context.Context, req req.PurchaseLimitSchemeListReq) (*resp.PurchaseLimitSchemeListResp, error)
    Delete(ctx context.Context, uuid uint64) error
}

type purchaseLimitSchemeSrv struct {
    dbm *database.DBManager
}

func NewPurchaseLimitSchemeSrv(dbm *database.DBManager) IPurchaseLimitSchemeSrv {
    return &purchaseLimitSchemeSrv{dbm: dbm}
}
```

### Repository 层

```go
// main/app/repository/purchase_limit_scheme_repo.go
type IPurchaseLimitSchemeRepo interface {
    Create(scheme *model.PurchaseLimitScheme) error
    Update(scheme *model.PurchaseLimitScheme) error
    GetByUuid(uuid uint64) (*model.PurchaseLimitScheme, error)
    GetList(options ...DBOption) ([]*model.PurchaseLimitScheme, int64, error)
    Delete(uuid uint64) error
    
    // 选项方法
    WhereCompanyUuid(companyUuid uint64) DBOption
    WhereStatus(status int8) DBOption
    WhereName(name string) DBOption
    Paginate(pageNo, pageSize int) DBOption
}

// 其他 3 个 Repo 类似
```

### API 层

在 `main/app/api/v1/shop/shop_purchase.go` 中新增方法和路由：

```go
func (h *PurchaseHandler) GetLimitSchemeList(c *gin.Context) {
    // 实现
}

func (h *PurchaseHandler) GetLimitSchemeDetail(c *gin.Context) {
    // 实现
}

func (h *PurchaseHandler) CreateLimitScheme(c *gin.Context) {
    // 实现
}

func (h *PurchaseHandler) UpdateLimitScheme(c *gin.Context) {
    // 实现
}

func (h *PurchaseHandler) DeleteLimitScheme(c *gin.Context) {
    // 实现
}

// 在路由注册中添加
func (h *PurchaseHandler) RegisterRoutes(router *gin.RouterGroup) {
    privateApi := router.Group("", middleware.Auth(authSrv, dbm))
    {
        // ... 现有路由 ...
        
        // 限购方案管理
        privateApi.GET("/purchase/limit_scheme/list", wrapper.GetLimitSchemeList)
        privateApi.GET("/purchase/limit_scheme/detail", wrapper.GetLimitSchemeDetail)
        privateApi.POST("/purchase/limit_scheme/create", wrapper.CreateLimitScheme)
        privateApi.POST("/purchase/limit_scheme/update", wrapper.UpdateLimitScheme)
        privateApi.DELETE("/purchase/limit_scheme/delete", wrapper.DeleteLimitScheme)
    }
}
```

---

## 🚨 错误处理

### 场景 1: 方案名称重复

- **处理方式**: 创建前检查同一总部下是否存在同名方案
- **错误提示**: "限购方案名称已存在，请使用其他名称"
- **HTTP Code**: 400

### 场景 2: 周期未选择

- **处理方式**: 参数验证失败
- **错误提示**: "至少需选择一个星期"
- **HTTP Code**: 400

### 场景 3: 物品未选择

- **处理方式**: 参数验证失败
- **错误提示**: "至少需选择一个物品"
- **HTTP Code**: 400

### 场景 4: 数据迁移失败

- **处理方式**: 事务回滚，保留旧表
- **错误提示**: "数据迁移失败，请联系管理员"
- **日志记录**: 详细的错误堆栈

---

## 🔒 安全设计

- **身份验证**: 所有 API 需要 JWT Token
- **权限控制**: 只有总部用户可以管理限购方案
- **SQL 注入防护**: 使用 GORM 参数化查询
- **软删除**: 使用 delete_time 字段，不物理删除

---

## 🧪 测试策略

### 单元测试

**目标覆盖率**:
- Service 层: 70%+
- Repository 层: 80%+

**测试内容**:
- 限购方案 CRUD
- 多表关联操作
- 数据迁移逻辑

### API 测试

**测试内容**:
- API 接口调用
- 参数验证
- 响应格式
- 错误处理

### 数据迁移测试

**测试内容**:
- 旧表数据完整性
- 迁移后数据一致性
- 事务回滚机制

---

## 📈 性能优化

1. **数据库优化**:
   - 添加索引（company_uuid, status, scheme_id等）
   - 使用连接池

2. **缓存优化**:
   - Redis 缓存限购方案配置（TTL: 5分钟）
   - Key 格式: `ttpos:purchase:limit_scheme:{uuid}`

3. **并发控制**:
   - 使用事务保证数据一致性
   - 数据迁移使用分布式锁

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 活动日志：`docs/team/activities/weifashi/2026-01/2026-01-20.md`

---

**版本**: v1.0.0  
**创建日期**: 2026-01-20  
**作者**: weifashi  
**Story Point**: 3
