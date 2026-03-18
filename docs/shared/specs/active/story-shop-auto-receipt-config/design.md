# 品采自动收货配置 技术设计文档

> 本文档定义品采自动收货配置功能的技术设计和实现方案。
> DooTask #40189 - 新管理端-参数设置/品采收货：特定发货仓库自动收货（DN 发送后定时执行）

## 📋 基本信息

| 项目 | 内容 |
|------|------|
| **Spec ID** | story-shop-auto-receipt-config |
| **设计人** | 曾振华 |
| **设计日期** | 2026-03-11 |
| **版本** | V2.20.0 |
| **终端** | shop（新管理端） |

---

## 📖 需求概述

### 背景

当前品采收货依赖人工操作，容易遗漏，且影响日结效率。部分**特定发货仓库**需要在固定时间点由系统自动完成收货。

### 用户故事

作为**总部管理员**，当 DN 已发送且满足配置条件时，我希望系统在设定时间**自动完成该发货仓库的收货**，并在记录中明确标记"自动收货"，以减少人工处理并保证日结及时。

### 功能范围

本文档覆盖以下功能模块：

| 模块 | 说明 |
|------|------|
| **参数设置 - 品采收货配置** | 总部配置自动收货规则（仓库 + 门店 + 延迟天数） |
| **自动收货记录** | 记录自动收货成功的单据，支持按日期筛选 |
| **定时任务** | 每天 00:00 扫描规则并执行自动收货 |
| **品采收货标识** | 自动收货的收货单增加"自动收货"标识，跳过附件上传 |

---

## 🔄 代码复用分析

### 可复用的现有组件

| 组件 | 路径 | 复用方式 |
|------|------|---------|
| **CompanySrv.GetCompanyListWithStoreCode** | `main/app/service/company.go` | 直接复用，获取门店列表 + 排序 |
| **总部仓库列表接口** | `GET /shop/warehouse/headquarter/list` | 直接复用，选择发货仓库 |
| **ProcessStoreCodeForSort** | `main/pkg/utils/string.go` | 直接复用，门店排序逻辑 |
| **erpSrv.GetDeliveryNoteList** | `main/app/service/rpc/erp/delivery_note.go` | 直接复用，通过 SoNo 查 ERPNext DN 列表 |
| **purchaseReceiptOrderSrv** | `main/app/service/purchase_order/receipt_order.go` | 扩展，增加自动收货参数 |
| **DailySalesOutboundSummaryTask** | `main/app/tasks/daily_sales_outbound_summary.go` | 参考，定时任务"扫描主库→遍历门店"模式 |
| **ErpStockEntryTask** | `main/app/tasks/erp_stock_entry_task.go` | 参考，分布式锁 + 重试模式 |
| **CompanyRepo.GetNoDeleteListByHeadquarterUuid** | `main/app/repository/company.go` | 直接复用，查总部下所有门店 |

### 需要新建的组件

| 组件 | 路径 | 说明 |
|------|------|------|
| **AutoReceiptRule Model** | `main/app/model/auto_receipt_rule.go` | 规则主表模型 |
| **AutoReceiptRuleShop Model** | `main/app/model/auto_receipt_rule_shop.go` | 规则门店子表模型 |
| **AutoReceiptLog Model** | `main/app/model/auto_receipt_log.go` | 记录数据模型 |
| **AutoReceiptRuleRepo** | `main/app/repository/auto_receipt_rule_repo.go` | 规则主表 Repository |
| **AutoReceiptRuleShopRepo** | `main/app/repository/auto_receipt_rule_shop_repo.go` | 规则门店 Repository |
| **AutoReceiptLogRepo** | `main/app/repository/auto_receipt_log_repo.go` | 记录 Repository |
| **AutoReceiptSrv** | `main/app/service/setting/auto_receipt.go` | 自动收货配置服务 |
| **AutoReceiptTask** | `main/app/tasks/auto_receipt_task.go` | 定时任务 |
| **DTO Req** | `main/app/dto/req/auto_receipt.go` | 请求参数 |
| **DTO Resp** | `main/app/dto/resp/setting/auto_receipt.go` | 响应结构 |
| **迁移文件** | `admin/database/migrations/` | 建表脚本 |

---

## 🏗️ 架构设计

### 整体架构

```
┌─────────────────────────────────────────────────────────────────┐
│                        新管理端 (Flutter)                        │
├─────────────────────────────────────────────────────────────────┤
│  规则配置 CRUD          │  自动收货记录查看      │  收货详情查看   │
│  warehouse/list (复用)  │  log/list             │  receipt/detail │
│  shop_list (新)         │                       │  (复用)         │
│  rule/* (新)            │                       │                 │
└──────────┬──────────────┴───────────┬───────────┴────────┬──────┘
           │                         │                     │
┌──────────▼──────────────────────────▼─────────────────────▼──────┐
│                         Main 模块 (Go + Gin)                     │
├──────────────────────────────────────────────────────────────────┤
│                                                                  │
│  API Layer: shop_setting.go (新增路由组)                          │
│       │                                                          │
│  Service Layer: setting/auto_receipt.go (新文件)                  │
│       │                                                          │
│  Repository Layer: auto_receipt_rule_repo.go                     │
│                    auto_receipt_rule_shop_repo.go                │
│                    auto_receipt_log_repo.go                      │
│       │                                                          │
│  Model Layer: auto_receipt_rule.go (主表)                        │
│               auto_receipt_rule_shop.go (子表)                   │
│               auto_receipt_log.go                                │
│                                                                  │
│  ┌─────────────────────────────────────┐                         │
│  │  定时任务: auto_receipt_task.go      │                         │
│  │  Cron: 0 0 0 * * * (每天00:00)      │                         │
│  │  分布式锁 + 错误隔离                 │                         │
│  └──────────┬──────────────────────────┘                         │
│             │                                                    │
└─────────────┼────────────────────────────────────────────────────┘
              │
    ┌─────────▼──────────┐          ┌──────────────────────┐
    │   SAAS 主库         │          │   门店库 (shop{uuid}) │
    │                    │          │                      │
    │  auto_receipt_rule │          │                      │
    │  auto_receipt_     │──扫描──▶ │  purchase_receipt    │
    │    rule_shop       │          │  _order (执行收货)    │
    │  auto_receipt_log  │◀──写入── │                      │
    └────────────────────┘          └──────────────────────┘
```

### 数据流

```
创建规则:  前端 → API → Service → RuleRepo → saas 主库
查看记录:  前端 → API → Service → LogRepo  → saas 主库 (列表)
                                           → 门店库 (详情，复用现有接口)
定时收货:  Cron → Task → 读 saas 主库规则
                       → 遍历门店库执行收货
                       → 写 saas 主库日志
```

---

## 🗄️ 数据库设计

### 存储位置决策

| 决策点 | 结论 | 理由 |
|--------|------|------|
| 规则表位置 | **saas 主库** | 总部级配置，定时任务集中扫描效率高 |
| 记录表位置 | **saas 主库** | 总部统一查看，避免聚合多门店库 |
| 租户隔离 | **headquarter_company_uuid** | 区分不同集团连锁，总部 UUID 做隔离键 |

### 数据方案决策

| 决策点 | 结论 | 理由 |
|--------|------|------|
| 规则存储结构 | **2 层主从结构**（主表=规则卡片，子表=门店明细） | 避免 warehouse_name 等字段重复存储，更新/删除操作更自然，语义更清晰 |
| 列表展示分组 | **按主表记录分组** | 每条主表记录 = 一个独立卡片，天然分组无需应用层处理 |
| 门店唯一约束 | **业务层校验** | 同一总部+同一仓库下，每个门店只能出现在一条规则中（跨规则校验） |
| 记录表内容 | **仅记录成功**，异常跳过仅写 WARN 日志 | 简化设计，减少数据量 |

### 表 1: ttpos_auto_receipt_rule（自动收货规则主表）

**所属库**: saas 主库（迁移 TARGET='main'）

```sql
CREATE TABLE IF NOT EXISTS `ttpos_auto_receipt_rule` (
    `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `uuid` bigint unsigned NOT NULL DEFAULT 0 COMMENT '唯一标识',
    `headquarter_company_uuid` bigint unsigned NOT NULL DEFAULT 0 COMMENT '总部company_uuid（租户隔离）',
    `warehouse_erp_code` varchar(100) NOT NULL DEFAULT '' COMMENT '发货仓库ERP编码',
    `warehouse_name` varchar(500) NOT NULL DEFAULT '' COMMENT '仓库名称（多语言JSON，冗余）',
    `delay_days` int NOT NULL DEFAULT 0 COMMENT 'DN发送后N天自动收货（0=当天24:00）',
    `status` tinyint NOT NULL DEFAULT 1 COMMENT '状态：1=启用，0=禁用',
    `create_time` int NOT NULL DEFAULT 0 COMMENT '创建时间',
    `update_time` int NOT NULL DEFAULT 0 COMMENT '更新时间',
    `delete_time` int NOT NULL DEFAULT 0 COMMENT '删除时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_uuid` (`uuid`),
    KEY `idx_hq_status` (`headquarter_company_uuid`, `status`, `delete_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='品采自动收货规则主表';
```

**说明**: 同一总部下允许多条规则配置同一仓库（不同门店组或不同延迟天数），每条规则对应前端一张卡片。

### 表 2: ttpos_auto_receipt_rule_shop（自动收货规则门店子表）

**所属库**: saas 主库（迁移 TARGET='main'）

```sql
CREATE TABLE IF NOT EXISTS `ttpos_auto_receipt_rule_shop` (
    `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `uuid` bigint unsigned NOT NULL DEFAULT 0 COMMENT '唯一标识',
    `rule_uuid` bigint unsigned NOT NULL DEFAULT 0 COMMENT '关联规则主表UUID',
    `shop_company_uuid` bigint unsigned NOT NULL DEFAULT 0 COMMENT '门店company_uuid',
    `create_time` int NOT NULL DEFAULT 0 COMMENT '创建时间',
    `update_time` int NOT NULL DEFAULT 0 COMMENT '更新时间',
    `delete_time` int NOT NULL DEFAULT 0 COMMENT '删除时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_uuid` (`uuid`),
    KEY `idx_rule` (`rule_uuid`, `delete_time`),
    KEY `idx_shop` (`shop_company_uuid`, `delete_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='品采自动收货规则门店子表';
```

**门店唯一性约束**: 同一总部 + 同一仓库下，每个门店只能出现在一条规则中。此约束通过**业务层校验**实现（创建时查询 `rule_shop JOIN rule WHERE warehouse_erp_code = ? AND shop_company_uuid = ?`），因为约束跨两张表，无法用单表唯一索引表达。

### 表 3: ttpos_auto_receipt_log（自动收货记录表）

**所属库**: saas 主库（迁移 TARGET='main'）

```sql
CREATE TABLE IF NOT EXISTS `ttpos_auto_receipt_log` (
    `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `uuid` bigint unsigned NOT NULL DEFAULT 0 COMMENT '唯一标识',
    `headquarter_company_uuid` bigint unsigned NOT NULL DEFAULT 0 COMMENT '总部company_uuid（租户隔离）',
    `shop_company_uuid` bigint unsigned NOT NULL DEFAULT 0 COMMENT '门店company_uuid（关联查门店名称）',
    `receipt_order_uuid` bigint unsigned NOT NULL DEFAULT 0 COMMENT '收货单UUID',
    `receipt_order_no` varchar(100) NOT NULL DEFAULT '' COMMENT '收货单号',
    `receipt_erp_order_no` varchar(255) NOT NULL DEFAULT '' COMMENT '收货单ERP单号',
    `receipt_time` int NOT NULL DEFAULT 0 COMMENT '收货日期（当天0点时间戳，筛选用）',
    `create_time` int NOT NULL DEFAULT 0 COMMENT '创建时间',
    `update_time` int NOT NULL DEFAULT 0 COMMENT '更新时间',
    `delete_time` int NOT NULL DEFAULT 0 COMMENT '删除时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_uuid` (`uuid`),
    KEY `idx_hq_date` (`headquarter_company_uuid`, `receipt_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='品采自动收货记录表';
```

**设计说明**:
- 仅记录自动收货**成功**的单据，异常跳过的仅写应用 WARN 日志
- `shop_company_uuid` 用于关联查询门店名称（不冗余存储，实时查）
- `receipt_time` 为当天 0 点时间戳，便于按日期筛选
- 查看收货详情时，通过 `shop_company_uuid` + `receipt_order_uuid` 到门店库查收货单

---

## 📊 数据模型

### Go Model

```go
// main/app/model/auto_receipt_rule.go
type AutoReceiptRule struct {
    Id                      uint64 `gorm:"column:id;primaryKey" json:"id"`
    Uuid                    uint64 `gorm:"column:uuid;uniqueIndex" json:"uuid"`
    HeadquarterCompanyUuid  uint64 `gorm:"column:headquarter_company_uuid" json:"headquarter_company_uuid"`
    WarehouseErpCode        string `gorm:"column:warehouse_erp_code" json:"warehouse_erp_code"`
    WarehouseName           string `gorm:"column:warehouse_name" json:"warehouse_name"`
    DelayDays               int    `gorm:"column:delay_days" json:"delay_days"`
    Status                  int    `gorm:"column:status" json:"status"`
    CreateTime              int64  `gorm:"column:create_time" json:"create_time"`
    UpdateTime              int64  `gorm:"column:update_time" json:"update_time"`
    DeleteTime              int64  `gorm:"column:delete_time" json:"delete_time"`
}

func (*AutoReceiptRule) TableName() string {
    return "ttpos_auto_receipt_rule"
}
```

```go
// main/app/model/auto_receipt_rule_shop.go
type AutoReceiptRuleShop struct {
    Id              uint64 `gorm:"column:id;primaryKey" json:"id"`
    Uuid            uint64 `gorm:"column:uuid;uniqueIndex" json:"uuid"`
    RuleUuid        uint64 `gorm:"column:rule_uuid" json:"rule_uuid"`
    ShopCompanyUuid uint64 `gorm:"column:shop_company_uuid" json:"shop_company_uuid"`
    CreateTime      int64  `gorm:"column:create_time" json:"create_time"`
    UpdateTime      int64  `gorm:"column:update_time" json:"update_time"`
    DeleteTime      int64  `gorm:"column:delete_time" json:"delete_time"`
}

func (*AutoReceiptRuleShop) TableName() string {
    return "ttpos_auto_receipt_rule_shop"
}
```

```go
// main/app/model/auto_receipt_log.go
type AutoReceiptLog struct {
    Id                      uint64 `gorm:"column:id;primaryKey" json:"id"`
    Uuid                    uint64 `gorm:"column:uuid;uniqueIndex" json:"uuid"`
    HeadquarterCompanyUuid  uint64 `gorm:"column:headquarter_company_uuid" json:"headquarter_company_uuid"`
    ShopCompanyUuid         uint64 `gorm:"column:shop_company_uuid" json:"shop_company_uuid"`
    ReceiptOrderUuid        uint64 `gorm:"column:receipt_order_uuid" json:"receipt_order_uuid"`
    ReceiptOrderNo          string `gorm:"column:receipt_order_no" json:"receipt_order_no"`
    ReceiptErpOrderNo       string `gorm:"column:receipt_erp_order_no" json:"receipt_erp_order_no"`
    ReceiptTime             int64  `gorm:"column:receipt_time" json:"receipt_time"`
    CreateTime              int64  `gorm:"column:create_time" json:"create_time"`
    UpdateTime              int64  `gorm:"column:update_time" json:"update_time"`
    DeleteTime              int64  `gorm:"column:delete_time" json:"delete_time"`
}

func (*AutoReceiptLog) TableName() string {
    return "ttpos_auto_receipt_log"
}
```

---

## 🧩 API 接口设计

### 路由总览

所有接口挂在 `/shop/setting/auto_receipt/` 路由组下，仅总部可访问。

| 方法 | 路径 | 说明 | 备注 |
|------|------|------|------|
| POST | `/shop/setting/auto_receipt/rule/create` | 创建规则 | 一个仓库 + 多个门店 = 一张卡片 |
| POST | `/shop/setting/auto_receipt/rule/update` | 更新规则 | 按规则 UUID 修改 delay_days / status |
| POST | `/shop/setting/auto_receipt/rule/delete` | 删除规则 | 删除整条规则（含门店） |
| POST | `/shop/setting/auto_receipt/rule/shop/delete` | 删除规则门店 | 从规则卡片中移除指定门店 |
| GET | `/shop/setting/auto_receipt/rule/list` | 规则列表 | 每条规则 = 一个卡片 |
| GET | `/shop/setting/auto_receipt/shop_list` | 可选门店列表 | 已配置的门店标记 disabled |
| GET | `/shop/setting/auto_receipt/log/list` | 自动收货记录 | 默认今天，分页 |
| GET | `/shop/setting/auto_receipt/log/detail` | 收货单详情 | 通过日志 UUID 跨门店查询 |

仓库选择复用现有接口 `GET /shop/warehouse/headquarter/list`，无需新建。

### 接口 1: 创建规则

**POST** `/shop/setting/auto_receipt/rule/create`

**请求**:
```go
// main/app/dto/req/auto_receipt.go
type CreateAutoReceiptRuleReq struct {
    WarehouseErpCode string   `json:"warehouse_erp_code" binding:"required"`
    WarehouseName    string   `json:"warehouse_name" binding:"required"`     // 多语言JSON
    ShopUuids        []uint64 `json:"shop_uuids" binding:"required,min=1"`
    DelayDays        int      `json:"delay_days" binding:"gte=0"`            // 0=当天24:00
}
```

**处理逻辑**:
1. 校验当前用户为总部管理员
2. 获取 `headquarter_company_uuid`
3. 查重：检查每个 shop_uuid 在该仓库下是否已被其他规则配置（跨规则校验）
4. 事务写入：
   - 创建 `ttpos_auto_receipt_rule` 主表记录（一条）
   - 批量创建 `ttpos_auto_receipt_rule_shop` 子表记录（N 条，关联主表 UUID）

**响应**: 标准成功响应 `{"code": 1, "message": "success", "data": {}}`

### 接口 2: 更新规则

**POST** `/shop/setting/auto_receipt/rule/update`

**请求**:
```go
type UpdateAutoReceiptRuleReq struct {
    Uuid      uint64 `json:"uuid" binding:"required"`              // 规则主表 UUID
    DelayDays *int   `json:"delay_days" binding:"omitempty,gte=0"` // 新延迟天数
    Status    *int   `json:"status"`                               // 可选，修改状态
}
```

**处理逻辑**:
- 按 `headquarter + uuid` 更新主表记录（仅更新主表，子表不受影响）

### 接口 3: 删除规则（整条卡片）

**POST** `/shop/setting/auto_receipt/rule/delete`

**请求**:
```go
type DeleteAutoReceiptRuleReq struct {
    Uuids []uint64 `json:"uuids" binding:"required,min=1"`  // 规则主表 UUID 列表
}
```

**处理逻辑**: 事务中同时软删除主表记录和关联的子表记录

### 接口 3.5: 删除规则门店（从卡片中移除门店）

**POST** `/shop/setting/auto_receipt/rule/shop/delete`

**请求**:
```go
type DeleteAutoReceiptRuleShopReq struct {
    Uuids []uint64 `json:"uuids" binding:"required,min=1"`  // 子表 UUID 列表
}
```

**处理逻辑**:
1. 按子表 UUID 软删除门店记录
2. 如果删除后该规则下无门店，自动软删除主表记录（空卡片自动清理）

### 接口 4: 规则列表

**GET** `/shop/setting/auto_receipt/rule/list`

**请求**:
```go
type AutoReceiptRuleListReq struct {
    WarehouseErpCode string `form:"warehouse_erp_code"`  // 可选，按仓库筛选
}
```

**处理逻辑**:
1. 查 `ttpos_auto_receipt_rule` 主表（`headquarter_company_uuid = ?`，可选 `warehouse_erp_code` 筛选）
2. 查 `ttpos_auto_receipt_rule_shop` 子表，按 `rule_uuid` 关联
3. 批量通过 `shop_company_uuid` 关联查询门店名称和店铺编号（复用 `GetCompanyListWithStoreCode`）
4. 每条规则内按门店排序（复用现有排序逻辑）

**响应**:
```go
type AutoReceiptRuleListResp struct {
    List []AutoReceiptRuleGroup `json:"list"`
}

type AutoReceiptRuleGroup struct {
    Uuid                uint64                `json:"uuid"`                  // 规则主表 UUID（卡片标识）
    WarehouseErpCode    string                `json:"warehouse_erp_code"`
    LocaleWarehouseName dto.LocaleResponse    `json:"locale_warehouse_name"`
    DelayDays           int                   `json:"delay_days"`
    Status              int                   `json:"status"`
    ShopCount           int                   `json:"shop_count"`
    Shops               []AutoReceiptRuleShop `json:"shops"`
}

type AutoReceiptRuleShop struct {
    Uuid     uint64 `json:"uuid"`       // 子表 UUID，用于单条删除门店
    ShopUuid uint64 `json:"shop_uuid"`  // 门店 company_uuid
    ShopCode string `json:"shop_code"`  // 实时关联查询
    ShopName string `json:"shop_name"`  // 实时关联查询
}
```

**响应示例**:
```json
{
  "code": 1,
  "data": {
    "list": [
      {
        "uuid": 1001,
        "warehouse_erp_code": "W1",
        "locale_warehouse_name": {"zh-CN": "中央仓库", "en": "Central Warehouse"},
        "delay_days": 0,
        "status": 1,
        "shop_count": 2,
        "shops": [
          {"uuid": 2001, "shop_uuid": 82673045, "shop_code": "001", "shop_name": "朝阳店"},
          {"uuid": 2002, "shop_uuid": 82673046, "shop_code": "002", "shop_name": "海淀店"}
        ]
      },
      {
        "uuid": 1002,
        "warehouse_erp_code": "W1",
        "locale_warehouse_name": {"zh-CN": "中央仓库", "en": "Central Warehouse"},
        "delay_days": 0,
        "status": 1,
        "shop_count": 1,
        "shops": [
          {"uuid": 2003, "shop_uuid": 82673047, "shop_code": "003", "shop_name": "西城店"}
        ]
      }
    ]
  }
}
```

> 注意：上面两个卡片的 `warehouse_erp_code` 和 `delay_days` 相同，但因为是两条独立规则（分两次创建），所以展示为两个独立卡片。

### 接口 5: 可选门店列表

**GET** `/shop/setting/auto_receipt/shop_list`

**请求**:
```go
type AutoReceiptShopListReq struct {
    WarehouseErpCode string `form:"warehouse_erp_code" binding:"required"`
}
```

**处理逻辑**:
1. 调用现有 `CompanySrv.GetCompanyListWithStoreCode()` 获取总部下所有门店
2. 查 `ttpos_auto_receipt_rule_shop JOIN ttpos_auto_receipt_rule` 中该仓库已配置的 `shop_company_uuid` 集合
3. 已配置的门店标记 `disabled=true`，并附带 `disabled_reason`

**响应**:
```go
type AutoReceiptShopListResp struct {
    List []AutoReceiptShopItem `json:"list"`
}

type AutoReceiptShopItem struct {
    Uuid           uint64 `json:"uuid"`             // 门店 company_uuid
    Name           string `json:"name"`             // 门店名称
    StoreCode      string `json:"store_code"`       // 店铺编号
    Status         int    `json:"status"`           // 1=启用, 0=禁用
    Disabled       bool   `json:"disabled"`         // 是否已配置（置灰）
    DisabledReason string `json:"disabled_reason"`  // 置灰原因
}
```

**响应示例**:
```json
{
  "code": 1,
  "data": {
    "list": [
      {"uuid": 82673045, "name": "未编号门店A", "store_code": "",    "status": 1, "disabled": false, "disabled_reason": ""},
      {"uuid": 82673046, "name": "001-朝阳店",  "store_code": "001", "status": 1, "disabled": true,  "disabled_reason": "已配置（当天收货）"},
      {"uuid": 82673047, "name": "002-海淀店",  "store_code": "002", "status": 1, "disabled": false, "disabled_reason": ""},
      {"uuid": 82673048, "name": "A01-西城店",  "store_code": "A01", "status": 0, "disabled": true,  "disabled_reason": "已配置（3天后收货）"}
    ]
  }
}
```

**门店排序规则**（复用现有逻辑）:
1. `store_code` 为空 → 排最前，按 `company_uuid` 升序
2. `store_code` 非空 → 去掉 "No." 前缀
   - 纯数字：按数字大小排序（去掉前导0）
   - 非数字：按字母排序（不区分大小写）
   - 数字 vs 非数字：数字优先
3. `store_code` 相同 → 按 `company_uuid` 升序

### 接口 6: 自动收货记录列表

**GET** `/shop/setting/auto_receipt/log/list`

**请求**:
```go
type AutoReceiptLogListReq struct {
    req.PageReq
    ReceiptDate int64 `form:"receipt_time"`  // 收货日期（默认今天0点时间戳）
}
```

**处理逻辑**:
1. 按 `headquarter_company_uuid + receipt_time` 查询记录
2. 批量通过 `shop_company_uuid` 关联查询门店名称

**响应**:
```go
type AutoReceiptLogListResp struct {
    List []AutoReceiptLogItem `json:"list"`
    Meta resp.PageResponse    `json:"meta"`
}

type AutoReceiptLogItem struct {
    Uuid              uint64 `json:"uuid"`
    ShopCompanyUuid   uint64 `json:"shop_company_uuid"`
    ShopName          string `json:"shop_name"`              // 实时关联查询
    ReceiptOrderUuid  uint64 `json:"receipt_order_uuid"`
    ReceiptOrderNo    string `json:"receipt_order_no"`
    ReceiptErpOrderNo string `json:"receipt_erp_order_no"`
    ReceiptDate       int64  `json:"receipt_time"`
}
```

### 接口 7: 自动收货记录详情（跨门店查询）

**GET** `/shop/setting/auto_receipt/log/detail`

> **为什么不能复用现有接口？** 现有 `GET /shop/purchase/receipt/detail` 通过 `ctx.GetDB()` 获取数据库连接，依赖登录用户的 `company_uuid`。总部管理员的 `company_uuid` 对应总部库，而收货单存储在门店库（如 `shop8267304538112000`），无法直接查询。

**请求**:
```go
type AutoReceiptLogDetailReq struct {
    Uuid uint64 `form:"uuid" binding:"required"`  // 日志记录 UUID
}
```

**处理逻辑**:
1. 按 `headquarter_company_uuid + uuid` 查日志记录，获取 `shop_company_uuid` 和 `receipt_order_uuid`
2. 校验 `shop_company_uuid` 属于当前总部（租户隔离）
3. 通过 `dbm.GetDB(shopCompanyUuid)` 获取门店库连接
4. 复用 `purchaseReceiptOrderSrv` 的查询逻辑，传入门店库 db 查收货单详情

**响应**: 复用现有 `resp.PurchaseReceiptOrderDetailResp`（含收货明细 Items + 附件 Files）

**前端交互**: 在自动收货记录列表点击"查看详情"，传该记录的 `uuid` 即可。

---

## ⏰ 定时任务设计

### 任务注册

```go
// main/command/root.go initializeTimers()
_, _ = c.AddFunc("0 0 * * * *", func() { // 每小时整点执行，覆盖所有时区的午夜
    tasks.NewAutoReceiptTask(dbm, cacheInstance).Execute()
})
```

> **为什么每小时执行一次？** 自动收货的语义是"商户本地午夜触发"，但不同商户可能在不同时区（UTC+3 ~ UTC+9）。如果只在服务器 00:00 执行一次，时区偏移较大的商户会延迟一天。每小时执行一次 + 时区前置过滤，确保每个商户在自己的本地午夜被处理，同时避免重复收货。

### 分布式锁

```go
// main/pkg/lock/system_lock.go 新增常量
const AutoReceiptLock = N  // 下一个可用编号
```

### 任务结构

```go
// main/app/tasks/auto_receipt_task.go
type AutoReceiptTask struct {
    dbm   *database.DBManager
    cache cache.Cache
}

func NewAutoReceiptTask(dbm *database.DBManager, cache cache.Cache) *AutoReceiptTask {
    return &AutoReceiptTask{dbm: dbm, cache: cache}
}
```

### 执行流程

```
Execute()
│
├── defer recover (panic 恢复)
├── 获取分布式锁 (AutoReceiptLock)
├── 判断是否已被其他节点处理（>1s 则跳过）
│
├── Step 1: 查 saas 主库所有启用规则（JOIN 主表+子表）
│   SELECT r.warehouse_erp_code, r.delay_days, s.shop_company_uuid
│   FROM ttpos_auto_receipt_rule r
│   JOIN ttpos_auto_receipt_rule_shop s ON r.uuid = s.rule_uuid
│   WHERE r.status = 1 AND r.delete_time = 0 AND s.delete_time = 0
│   → 构建规则索引: map[shop_company_uuid][]{ warehouse_erp_code, delay_days }
│
├── Step 2: 按 shop_company_uuid 分组
│
├── Step 3: FOR EACH 门店
│   ├── 获取门店库 db := dbm.GetDB(shopCompanyUuid)
│   ├── 获取该门店的 CompanySetting（含 ErpnextHeadquarterAbbr、Timezone）
│   │
│   ├── ★ 时区前置过滤: 商户本地时间是否在 [00:00, 01:00) 区间？
│   │   localNow := time.Now().In(loc)
│   │   localNow.Hour() != 0 → skip 整个门店（不是该门店的午夜时段）
│   │
│   ├── 查该门店所有「审核通过待收货」的品采采购单
│   │   WHERE erp_sale_order_no != ''       -- 有 ERP 销售单号
│   │     AND purchase_type = 2             -- 品采
│   │     AND status = 2                    -- 已通过（4=全部收货，无需处理）
│   │     AND delete_time = 0
│   │
│   ├── FOR EACH 采购单
│   │   ├── 调用 ERPNext API 查询 DN 列表:
│   │   │   erpSrv.GetDeliveryNoteList(ctx, {
│   │   │       CompanyAbbr:  companySetting.ErpnextHeadquarterAbbr,
│   │   │       SoNo:         purchaseOrder.ErpSaleOrderNo,
│   │   │       IncludeItems: true,
│   │   │   })
│   │   │
│   │   ├── FOR EACH DN → 执行 4 层过滤（详见「DN 收货判断」）
│   │   │   │
│   │   │   ├── 过滤 1: DN.Status == "To Bill" → 否则 skip
│   │   │   ├── 过滤 2: DN.SetWarehouse 匹配规则 → 否则 skip
│   │   │   ├── 过滤 3: shouldAutoReceipt(PostingDate, delay_days, timezone) → 否则 skip
│   │   │   ├── 过滤 4: 计算待收数量（详见「待收数量计算逻辑」）→ 全部=0 则 skip
│   │   │   │
│   │   │   ├── 组装 CreatePurchaseReceiptOrderReq:
│   │   │   │       PurchaseOrderUuid: purchaseOrder.Uuid
│   │   │   │       DeliveryNoteNo:    DN.Name
│   │   │   │       IsConfirm=1, FileUuids=空
│   │   │   │       Items[i].UnitList = 各单位待收数量
│   │   │   │
│   │   │   ├── ④ 调用 CreatePurchaseReceiptOrder (复用现有逻辑):
│   │   │   │   ├── validateDNReceipt: DN物品匹配、单位校验
│   │   │   │   ├── validateReceiptMaterialStatus: 物料状态校验
│   │   │   │   ├── 更新 PurchaseOrderItem.ArrivalNum
│   │   │   │   ├── 库存转移: transit仓 → target仓
│   │   │   │   └── ERP 同步: SavePurchaseReceipt
│   │   │   │
│   │   │   ├── 异常 → logger.Warn + skip (不入log表)
│   │   │   └── 成功 → 写 ttpos_auto_receipt_log + 操作日志标记"自动收货"
│   │   │
│   │   └── 单条 DN 失败不影响其他 DN (continue)
│   │
│   └── 单个采购单失败不影响其他采购单 (continue)
│
└── Step 4: 记录统计日志
    logger.Info("自动收货任务完成",
        total_po, total_dn, success, skip)
```

### 时间判断逻辑

时间判断分**两层**：第一层在门店级别快速过滤，第二层在 DN 级别精确判断。

#### 第一层：时区前置过滤（门店级别）

每小时执行一次，但只处理本地时间刚好在 `[00:00, 01:00)` 区间的门店。这样每个门店每天只会被实际处理一次。

```go
func isShopMidnight(timezone string) bool {
    loc, err := time.LoadLocation(timezone)
    if err != nil {
        loc, _ = time.LoadLocation("Asia/Shanghai")
    }
    localNow := time.Now().In(loc)
    return localNow.Hour() == 0  // 本地时间在 00:xx
}
```

**各时区触发时间**（服务器 UTC+8）:
```
商户时区              UTC偏移   服务器触发时刻（UTC+8）   商户本地时间
────────────────────────────────────────────────────────────────
Asia/Tokyo            +9       23:00                    00:00 ✅
Asia/Shanghai         +8       00:00                    00:00 ✅
Asia/Bangkok          +7       01:00                    00:00 ✅
Asia/Yangon           +6:30    01:30(整点01:00触发)      00:30 ✅
Europe/Istanbul       +3       05:00                    00:00 ✅
```

#### 第二层：DN 到期判断（DN 级别）

DN 返回两个时间字段：`PostingDate`（过账日期，如 `"2026-03-10"`）和 `PostingTime`（过账时间，如 `"14:30:00"`）。
仅使用 `PostingDate` 做日期级别判断，`PostingTime` 不参与计算。

```go
// delay_days=0: DN 过账当天 24:00（商户本地时间）执行
// delay_days=1: DN 过账后第 1 天 24:00 执行
// delay_days=N: DN 过账后第 N 天 24:00 执行

func shouldAutoReceipt(dnPostingDate string, delayDays int, timezone string) bool {
    loc, err := time.LoadLocation(timezone)
    if err != nil {
        loc, _ = time.LoadLocation("Asia/Shanghai") // fallback 默认时区
    }

    // 将 DN 过账日期解析为商户时区的 00:00:00
    dnDate, _ := time.ParseInLocation("2006-01-02", dnPostingDate, loc)

    // deadline = 过账日期 + delay_days + 1 天的 00:00:00（商户时区）
    // +1 因为当天24:00 = 次日00:00
    deadline := dnDate.AddDate(0, 0, delayDays+1)

    // 当前时间转为商户时区后比较
    now := time.Now().In(loc)
    return !now.Before(deadline)
}
```

**时区处理要点**:

| 步骤 | 处理 | 说明 |
|------|------|------|
| DN 日期解析 | `time.ParseInLocation(..., loc)` | 以商户时区解析，避免 UTC 偏差 |
| 当前时间 | `time.Now().In(loc)` | 转为商户本地时间 |
| 时区来源 | `CompanySetting.GetTimezone()` | 每个门店独立时区配置 |
| Fallback | `Asia/Shanghai` | 时区解析失败时使用默认值 |

**示例**（商户时区 `Asia/Tokyo` UTC+9）:
```
DN.PostingDate = "2026-03-10", delay_days = 0
deadline = 2026-03-11 00:00:00 JST (= 2026-03-10 15:00:00 UTC)

服务器时间(UTC)           商户本地时间(JST)       判断结果
──────────────────────────────────────────────────────────
2026-03-10 14:59 UTC  →  2026-03-10 23:59 JST   ❌ 不收货
2026-03-10 15:00 UTC  →  2026-03-11 00:00 JST   ✅ 收货
```

### DN 收货判断（4 层过滤）

每个 DN 需通过以下 4 层过滤，全部通过才执行收货：

```
ERPNext 返回 DN 列表
│
├── 过滤 1: DN 状态
│   DN.Status == "To Bill" ?
│   └── "To Bill" = ERPNext docstatus=1（已提交/已发货），可收货
│   └── 其他状态（Draft/Cancelled/Completed）→ skip
│
├── 过滤 2: 仓库匹配规则
│   DN.SetWarehouse 在该门店的自动收货规则中？
│   └── 未命中 → skip（该发货仓库没配置自动收货）
│   └── 命中 → 取出对应的 delay_days
│
├── 过滤 3: 到期时间
│   shouldAutoReceipt(DN.PostingDate, delay_days, timezone) ?
│   └── 未到期 → skip
│
└── 过滤 4: 有待收数量
    计算 DN 各物品的待收数量 > 0 ?
    ├── 全部待收=0 → skip（已被人工或上次自动收完）
    └── 有待收 → 执行收货
```

### 待收数量计算逻辑

自动收货采用**全量收货**策略——将 DN 中所有未收货的数量一次性全部收完。

#### 数据来源

```
┌─────────────────────────────────┐     ┌──────────────────────────────────┐
│  ERPNext DN（API 返回）           │     │  门店库（本地查询）                │
│                                 │     │                                  │
│  DN.Items[] 每项:               │     │  已确认的收货单:                   │
│    ItemCode  = "MAT-001"        │     │    PurchaseReceiptOrder           │
│    Qty       = 100              │     │      WHERE purchase_order_uuid=PO │
│    Uom       = "Kg"             │     │        AND delivery_note_no=DN    │
│                                 │     │        AND is_confirm=1           │
│  → DN 的发货数量（仓库发了多少）   │     │        AND delete_time=0          │
│                                 │     │                                  │
│                                 │     │  收货明细 Items → Units:           │
│                                 │     │    MaterialCode + ErpnextUom      │
│                                 │     │    → 汇总已收数量                  │
└─────────────────────────────────┘     └──────────────────────────────────┘
```

#### 计算流程

```
输入: DN (来自ERPNext), PurchaseOrder (来自门店库)

Step 1: 查该 DN 下所有已确认的收货单
  receipts = PurchaseReceiptOrder
    WHERE purchase_order_uuid = PO.Uuid
      AND delivery_note_no = DN.Name    -- 按 DN 单号匹配
      AND is_confirm = 1                -- 只算已确认的
      AND delete_time = 0

Step 2: 汇总已收数量（按 ItemCode + Uom 维度）
  receivedMap = map[ItemCode:Uom] → 已收总数

  FOR EACH receipt IN receipts:
    FOR EACH item IN receipt.Items:
      FOR EACH unit IN item.Units:
        key = item.MaterialCode + ":" + unit.ErpnextUom
        receivedMap[key] += unit.Num

Step 3: 逐项计算待收数量
  pendingItems = []

  FOR EACH dnItem IN DN.Items:
    key = dnItem.ItemCode + ":" + dnItem.Uom
    received = receivedMap[key]          -- 已收数量（无记录则为0）
    pending  = dnItem.Qty - received     -- 待收数量

    IF pending <= 0:
      continue                           -- 该物品已收完，跳过

    pendingItems = append(pendingItems, {
      ItemCode: dnItem.ItemCode,
      Uom:      dnItem.Uom,
      Pending:  pending,
    })

Step 4: 判断结果
  IF len(pendingItems) == 0:
    → skip 该 DN（全部已收完，无需自动收货）
  ELSE:
    → 组装收货请求，收取所有 pendingItems
```

#### 计算示例

```
DN "DN-2026-001" 发货内容:
  ├── 面粉(MAT-001)  100 Kg
  ├── 糖(MAT-002)     50 Kg
  └── 面粉(MAT-001)   20 Bag    ← 同物品不同单位，独立计算

已有确认收货单（人工收货）:
  └── Receipt#1:
        ├── 面粉(MAT-001)  60 Kg
        └── 糖(MAT-002)    50 Kg

汇总已收: { "MAT-001:Kg"=60, "MAT-002:Kg"=50 }

计算待收:
  面粉 Kg:  100 - 60 = 40 ✅ 待收
  糖 Kg:     50 - 50 =  0 ❌ 已收完，跳过
  面粉 Bag:  20 -  0 = 20 ✅ 待收（不同单位独立计算）

自动收货请求: 收取 面粉 40Kg + 面粉 20Bag
```

#### 组装收货请求

```go
// 将 pendingItems 映射回采购单物品，组装收货请求
req := CreatePurchaseReceiptOrderReq{
    PurchaseOrderUuid:  PO.Uuid,
    DeliveryNoteNo:     DN.Name,
    IsFromDeliveryNote: 1,
    IsConfirm:          1,        // 直接确认，触发入库+ERP同步
    Items: []ReceiptItem{
        {
            PurchaseOrderItemUuid: poItem.Uuid,  // 通过 MaterialCode 匹配采购单物品
            UnitList: []ReceiptItemUnit{
                {
                    UomUuid:    unit.Uuid,        // 从采购单物品的 Units 中匹配
                    Num:        pending,           // 待收数量
                    ErpnextUom: dnItem.Uom,
                },
            },
        },
    },
    // FileUuids 为空，自动收货无附件
}
```

**关键匹配关系**:

| DN 字段 | 采购单字段 | 说明 |
|---------|-----------|------|
| `DN.Name` | `ReceiptOrder.DeliveryNoteNo` | 确定是哪个 DN 的收货 |
| `DNItem.ItemCode` | `POItem.MaterialCode` | 匹配物品 |
| `DNItem.Uom` | `POItemUnit.ErpnextUom` | 匹配单位 |
| `DNItem.Qty` | — | DN 发货数量（计算基准） |

**场景处理**:

| 场景 | 处理方式 |
|------|---------|
| DN 已部分人工收货 | 自动扣减已收数量，只收剩余部分 |
| DN 已全部收完 | 待收数量为 0，skip 该 DN |
| 同物品多单位（如 Kg + Bag） | 按 `ItemCode + Uom` 维度分别计算 |
| 超收 | 不会发生，严格按 DN 待收数量收货 |
| DN 物品不在采购单中 | 跳过该物品（validateDNReceipt 会校验） |
| 一个采购单对应多个 DN | 每个 DN 独立收货，互不影响 |

### 异常处理策略

| 异常类型 | 处理方式 | 日志级别 |
|----------|---------|---------|
| ERPNext API 调用失败 | 跳过该采购单 | ERROR |
| DN 物品与采购单不匹配 | 跳过该 DN | WARN |
| 单位/物料配置缺失 | 跳过该 DN | WARN |
| 门店库连接失败 | 跳过该门店 | ERROR |
| 收货逻辑执行失败 | 跳过该 DN | ERROR |
| ERP 同步失败 | 跳过该 DN | ERROR |
| 任务整体 panic | recover 恢复 | ERROR |

异常单据跳过后不阻塞其他单据处理，保留人工收货能力。

---

## 🔗 与现有收货逻辑的集成

### 扩展点

在现有 `purchaseReceiptOrderSrv` 的收货方法中增加 `isAutoReceipt` 参数：

| 行为 | 手动收货 | 自动收货 (isAutoReceipt=true) |
|------|---------|------------------------------|
| 触发方式 | API 请求 | 定时任务 |
| 附件上传 | 需要 | **跳过** |
| 操作人 | 当前登录用户 | 系统（标记"自动收货"） |
| 异常处理 | 返回错误给前端 | 跳过 + WARN 日志 |
| 收货确认 | 用户确认 | **自动确认** |
| 操作日志 | 记录操作人 | 标记"自动收货" |

### 收货数量计算逻辑

自动收货采用**全量收货**策略——将 DN 中所有未收货的数量一次性全部收完。

**数据来源**:

```
DN 数据来自 ERPNext API（非本地数据库）:
  erpSrv.GetDeliveryNoteList(SoNo = purchaseOrder.ErpSaleOrderNo)
  → DeliveryNote { Name, PostingDate, SetWarehouse, Status, Items[] }
  → DeliveryNoteItem { ItemCode, Qty, Uom }
```

**计算流程**:

```
1. 从 ERPNext 获取 DN 物品列表
   ├── DN Item.ItemCode → 匹配采购单 PurchaseOrderItem.MaterialCode
   ├── DN Item.Uom → 匹配 PurchaseOrderItem.ErpnextUom
   └── DN Item.Qty → DN 发货数量（按单位）

2. 查已有的确认收货（门店库）
   ├── 查 PurchaseReceiptOrder WHERE purchase_order_uuid = PO.uuid
   │     AND delivery_note_no = DN.Name
   │     AND status = 已收货(is_confirm=1)
   ├── 按「MaterialCode + ErpnextUom」维度汇总已收数量
   └── 已收数量 = SUM(receipt_item.num 或 receipt_item_unit.num)

3. 计算待收数量
   └── 待收数量 = DN Item.Qty - 已收数量（按 ItemCode:Uom 维度）

4. 组装请求
   req := CreatePurchaseReceiptOrderReq{
       PurchaseOrderUuid:  purchaseOrder.Uuid,
       DeliveryNoteNo:     dn.Name,        // ERPNext DN 单号
       IsFromDeliveryNote: 1,
       IsConfirm:          1,              // 直接确认，触发入库
       Items: [{
           PurchaseOrderItemUuid: poItem.Uuid,
           UnitList: [{
               UomUuid:    unitUuid,
               Num:        待收数量,       // DN Qty - 已收数量
               ErpnextUom: dnItem.Uom,
           }],
       }],
       // FileUuids 为空，自动收货无附件
   }
```

**关键场景处理**:

| 场景 | 处理方式 |
|------|---------|
| DN 已部分人工收货 | 自动扣减已收数量，只收剩余部分 |
| DN 已全部收完 | 待收数量为 0，跳过该 DN（不重复收货） |
| 多单位物品 | 按 `ItemCode + Uom` 维度分别计算待收数量 |
| 超收 | 不会发生，严格按 DN 待收数量收货 |
| DN 物品不在采购单中 | 跳过该物品（validateDNReceipt 会校验） |
| 一个采购单对应多个 DN | 每个 DN 独立收货，互不影响 |

### 收货单标识

自动收货完成的收货单，在操作日志（`ttpos_purchase_order_log`）中明确标记"自动收货"，便于后续追溯。

---

## 📋 前端交互流程

### 创建规则流程

```
1. 进入"参数设置 → 品采收货配置"
2. 点击"新增规则"
3. 选择发货仓库  → 调用 GET /shop/warehouse/headquarter/list (复用)
4. 选择门店      → 调用 GET /shop/setting/auto_receipt/shop_list
   - 已配置的门店置灰，显示原因
   - 支持多选
5. 设置延迟天数  → 输入 N（0=当天24:00）
6. 确认创建      → 调用 POST /shop/setting/auto_receipt/rule/create
```

### 规则列表展示

每条规则 = 一个独立卡片（按主表 UUID 区分），即使仓库和延迟天数相同也展示为不同卡片：

```
品采收货配置
├── [规则 #1001] 中央仓库 - 当天收货
│   ├── 001-朝阳店 [×]
│   └── 002-海淀店 [×]
├── [规则 #1002] 中央仓库 - 当天收货    ← 同仓库同delay，但不同规则
│   └── 003-西城店 [×]
└── [规则 #1003] 华南仓 - 1天后收货
    ├── A01-广州店 [×]
    └── A02-深圳店 [×]

[×] = 可从卡片中单独移除该门店
```

### 自动收货记录

```
自动收货记录  [日期筛选: 2026-03-11 ▼]
┌──────────┬────────────────┬──────────────────┬────────────────┐
│ 门店名称  │ 收货单号        │ ERP单号           │ 操作           │
├──────────┼────────────────┼──────────────────┼────────────────┤
│ 朝阳店   │ PR-2026-0001   │ PREC-00123       │ [查看详情]      │
│ 海淀店   │ PR-2026-0002   │ PREC-00124       │ [查看详情]      │
└──────────┴────────────────┴──────────────────┴────────────────┘
```

---

## ✅ 验收标准

| # | 条件 | 预期行为 |
|---|------|---------|
| AC1 | 发货仓库已配置自动收货 AND DN 已发送 | 系统在设定时间（默认当天24:00）自动完成收货 |
| AC2 | 配置了 DN 发送后 N 天自动收货 | 系统在第 N 天的 00:00 执行自动收货 |
| AC3 | 收货单据异常（物品不匹配/数量不一致/物料缺失） | 跳过该单据，记录 WARN 日志，不阻塞其他单据 |
| AC4 | 发货仓库未配置自动收货 | 不对该仓库的 DN 执行自动收货 |
| AC5 | 自动收货执行完成 | 操作日志标记"自动收货" |
| AC6 | 自动收货成功 | 记录写入 ttpos_auto_receipt_log |
| AC7 | 人工收货入口存在 | 异常单据可人工收货补录 |
| AC8 | 同一门店同一仓库 | 不允许重复配置规则 |
| AC9 | 选择门店时 | 已配置的门店置灰并显示原因 |
| AC10 | 多集团场景 | 不同总部的规则完全隔离 |

---

## 📂 文件变更清单

### 新建文件

| 文件 | 说明 |
|------|------|
| `main/app/model/auto_receipt_rule.go` | 规则主表 Model |
| `main/app/model/auto_receipt_rule_shop.go` | 规则门店子表 Model |
| `main/app/model/auto_receipt_log.go` | 记录 Model |
| `main/app/repository/auto_receipt_rule_repo.go` | 规则主表 Repository |
| `main/app/repository/auto_receipt_rule_shop_repo.go` | 规则门店 Repository |
| `main/app/repository/auto_receipt_log_repo.go` | 记录 Repository |
| `main/app/service/setting/auto_receipt.go` | 配置服务（规则CRUD + 门店列表 + 记录查询） |
| `main/app/tasks/auto_receipt_task.go` | 定时任务 |
| `main/app/dto/req/auto_receipt.go` | 请求 DTO |
| `main/app/dto/resp/setting/auto_receipt.go` | 响应 DTO |
| `admin/database/migrations/{ts}_create_auto_receipt_tables.php` | 迁移文件（TARGET='main'） |

### 修改文件

| 文件 | 修改内容 |
|------|---------|
| `main/app/api/v1/shop/shop_setting.go` | 新增自动收货路由组 |
| `main/router/router.go` | 注册新路由 |
| `main/command/root.go` | 注册定时任务 |
| `main/pkg/lock/system_lock.go` | 新增 AutoReceiptLock 常量 |
| `main/app/service/purchase_order/receipt_order.go` | 收货方法增加 isAutoReceipt 参数 |
| `admin/database/seeds/saas.sql` | 同步更新建表语句 |
