# TTPOS ERP 集成服务需求文档

## 1. 项目概述

### 1.1 项目简介
ttpos-erp 是 TTPOS 中台工程的核心模块之一，使用 GoFrame v2.x 框架开发的 MonoRepoApp。负责与 ERPNext 系统进行数据同步和业务对接，为 POS 系统提供完整的库存管理、销售管理、采购管理等 ERP 能力支撑。

### 1.2 技术栈
- **框架**: GoFrame v2.x
- **通信协议**: gRPC（对内）+ HTTP REST API（对接 ERPNext）
- **消息队列**: RocketMQ
- **数据库**: MySQL（独立 schema）+ ERPNext Database
- **外部系统**: ERPNext（Frappe 框架）
- **工程结构**: 遵循 GoFrame 标准目录结构

### 1.3 核心特性
- 统一的 ERPNext API 调用封装
- 多站点（Site）支持，支持多租户架构
- 物品（Item）管理与同步
- POS 销售管理（开账、关账、发票）
- 采购管理与供应商管理
- 库存管理与盘点
- 公司与分支机构管理
- 异步消息处理机制

---

## 2. 功能需求

### 2.1 ERPNext 集成服务

#### 2.1.1 Document Service - 文档操作服务

**接口定义**:
```go
type IDocument interface {
    // 查询文档列表
    List(ctx context.Context, req *ErpReq, params *RequestParams) (*gjson.Json, error)
    // 获取单个文档
    Get(ctx context.Context, req *ErpReq, params interface{}) (*gjson.Json, error)
    // 创建文档
    Create(ctx context.Context, doctype string, data interface{}) (*gjson.Json, error)
    // 更新文档
    Update(ctx context.Context, req *ErpReq, data interface{}) (*gjson.Json, error)
    // 删除文档
    Delete(ctx context.Context, req *ErpReq) (*gjson.Json, error)
    // 变更文档状态
    ChangeDocStatus(ctx context.Context, doctype, name, status string) (*gjson.Json, error)
}
```

**功能说明**:
- 封装 ERPNext Frappe REST API 调用
- 支持多站点授权切换
- 统一的请求/响应处理
- 错误检测与包装

#### 2.1.2 RPC Service - 远程方法调用服务

**接口定义**:
```go
type IRpc interface {
    // 执行 ERPNext 远程方法
    Execute(ctx context.Context, req *ErpReq, params interface{}) (*gjson.Json, error)
    // 获取站点编码
    GetSiteCode(ctx context.Context) string
    // 获取站点授权信息
    GetAndProcessSiteAuthorization(ctx context.Context, siteCode string) (*SiteAuth, error)
}
```

**功能说明**:
- 调用 ERPNext 自定义 API 方法
- 站点授权管理
- 收银员身份模拟（Fake User）

---

### 2.2 物品管理功能

#### 2.2.1 Item Service - 物品服务

**gRPC 接口定义**:
```protobuf
service ItemService {
  // 获取物品列表
  rpc GetItemList(GetItemListReq) returns (GetItemListResp);
  // 获取单个物品
  rpc GetItem(GetItemReq) returns (ItemInfo);
  // 保存物品（创建/更新）
  rpc SaveItem(ItemInfo) returns (ItemInfo);
  // 删除物品
  rpc DeleteItem(DeleteItemReq) returns (DeleteItemResp);
  // 获取物品库存
  rpc GetItemStock(GetItemStockReq) returns (GetItemStockResp);
  // 保存 POS 属性物品
  rpc SavePosAttribute(SavePosAttributeReq) returns (ItemInfo);
  // 保存 POS 加料物品
  rpc SavePosAddon(SavePosAddonReq) returns (ItemInfo);
  // 创建多规格商品
  rpc CreateSingleVariantItem(CreateSingleVariantItemReq) returns (string);
}
```

**功能说明**:
- 物品的增删改查操作
- 多规格商品（变体）管理
- POS 特殊物品（属性、加料）管理
- 物品库存查询
- 物品同步到 ERPNext

#### 2.2.2 物品分组

| 物品分组 | 常量 | 编码前缀 | 说明 |
|---------|------|---------|------|
| 原材料 | Raw Material | WPR | 库存物品，需盘点 |
| 商品 | CFG Products | SP | 销售商品，非库存 |
| 套餐 | CFG Products | TC | 组合商品 |
| POS 属性 | Pos Attribute | SXV | 规格属性 |
| POS 加料 | Pos Addon | JLV | 加料选项 |

---

### 2.3 销售管理功能

#### 2.3.1 Selling Service - 销售服务

**gRPC 接口定义**:
```protobuf
service SellingService {
  // 获取 POS 配置文件列表
  rpc GetPosProfileList(PosProfileReq) returns (PosProfileListResp);
  // 创建 POS 配置文件
  rpc CreatePosProfile(CreatePosProfileReq) returns (POSProfile);
  // POS 开账
  rpc OpenPosEntry(OpenPosEntryReq) returns (OpenPosEntryResp);
  // POS 关账
  rpc ClosePosEntry(ClosePosEntryReq) returns (ClosePosEntryResp);
  // 保存 POS 发票
  rpc SavePosInvoice(SavePosInvoiceReq) returns (SavePosInvoiceResp);
  // 退货 POS 发票
  rpc ReturnPosInvoice(ReturnPosInvoiceReq) returns (ReturnPosInvoiceResp);
  // 取消 POS 发票
  rpc CancelPosInvoice(CancelPosInvoiceReq) returns (Empty);
  // 获取支付方式列表
  rpc GetModeOfPaymentList(GetModeOfPaymentListReq) returns (GetModeOfPaymentListResp);
}
```

**功能说明**:
- POS 配置文件管理
- 开账/关账流程
- POS 发票创建与管理
- 退货处理
- 支付方式管理

---

### 2.4 采购管理功能

#### 2.4.1 Buying Service - 采购服务

**gRPC 接口定义**:
```protobuf
service BuyingService {
  // 创建采购订单
  rpc CreatePurchaseOrder(CreatePurchaseOrderReq) returns (CreatePurchaseOrderResp);
  // 更新采购订单
  rpc UpdatePurchaseOrder(UpdatePurchaseOrderReq) returns (UpdatePurchaseOrderResp);
  // 获取供应商列表
  rpc GetSupplierList(GetSupplierListReq) returns (GetSupplierListResp);
  // 创建供应商
  rpc CreateSupplier(CreateSupplierReq) returns (CreateSupplierResp);
}
```

**功能说明**:
- 采购订单的创建与更新
- 供应商管理
- 与总部供应商的关联

---

### 2.5 库存管理功能

#### 2.5.1 Stock Service - 库存服务

**gRPC 接口定义**:
```protobuf
service StockService {
  // 获取仓库列表
  rpc GetWarehouseList(GetWarehouseListReq) returns (GetWarehouseListResp);
  // 获取默认仓库
  rpc GetDefaultWarehouse(GetDefaultWarehouseReq) returns (Warehouse);
  // 创建物料转移
  rpc CreateMaterialTransfer(CreateMaterialTransferReq) returns (CreateMaterialTransferResp);
  // 创建库存盘点
  rpc CreateStockReconciliation(CreateStockReconciliationReq) returns (CreateStockReconciliationResp);
  // 创建发货单
  rpc CreateDeliveryNote(CreateDeliveryNoteReq) returns (CreateDeliveryNoteResp);
}
```

**功能说明**:
- 仓库管理
- 物料转移处理
- 库存盘点
- 发货单管理

---

### 2.6 公司管理功能

#### 2.6.1 Company Service - 公司服务

**gRPC 接口定义**:
```protobuf
service CompanyService {
  // 获取公司列表
  rpc GetCompanyList(GetCompanyListReq) returns (GetCompanyListResp);
  // 根据缩写获取公司
  rpc GetCompanyWithAbbr(GetCompanyWithAbbrReq) returns (CompanyInfo);
  // 根据缩写获取公司名称
  rpc GetCompanyNameWithAbbr(GetCompanyNameWithAbbrReq) returns (string);
}
```

**功能说明**:
- 公司信息查询
- 公司缩写与全名映射
- 分支机构管理

---

### 2.7 其他管理功能

#### 2.7.1 CRM Service - 客户关系管理

- 地址管理（Address）
- 联系人管理（Contact）

#### 2.7.2 Manufacturing Service - 生产管理

- BOM（物料清单）管理

#### 2.7.3 Setup Service - 系统设置

- 时区转换
- 系统参数配置

---

## 3. 数据库设计

### 3.1 本地数据表

#### 3.1.1 Site 表 - 站点配置
```sql
CREATE TABLE `erp_site` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `site_code` varchar(10) NOT NULL COMMENT '站点编码',
  `site_name` varchar(100) NOT NULL COMMENT '站点名称',
  `site_url` varchar(200) NOT NULL COMMENT '站点URL',
  `api_key` varchar(200) NOT NULL COMMENT 'API密钥',
  `api_secret` varchar(200) NOT NULL COMMENT 'API密钥',
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '状态',
  `created_at` int NOT NULL DEFAULT 0,
  `updated_at` int NOT NULL DEFAULT 0,
  `deleted_at` int NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_site_code` (`site_code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='ERP站点配置表';
```

#### 3.1.2 Item 表 - 物品缓存
```sql
CREATE TABLE `erp_item` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `item_code` varchar(100) NOT NULL COMMENT '物品编码',
  `item_name` varchar(200) NOT NULL COMMENT '物品名称',
  `item_group` varchar(100) NOT NULL COMMENT '物品分组',
  `stock_uom` varchar(50) NOT NULL COMMENT '库存单位',
  `company` varchar(200) DEFAULT NULL COMMENT '所属公司',
  `branch` varchar(100) DEFAULT NULL COMMENT '所属分支',
  `disabled` tinyint NOT NULL DEFAULT 0 COMMENT '是否禁用',
  `sync_time` int NOT NULL DEFAULT 0 COMMENT '同步时间',
  `created_at` int NOT NULL DEFAULT 0,
  `updated_at` int NOT NULL DEFAULT 0,
  `deleted_at` int NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_item_code` (`item_code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='ERP物品缓存表';
```

### 3.2 ERPNext DocType 映射

| DocType | 说明 | 主要用途 |
|---------|------|---------|
| Item | 物品 | 商品/原材料管理 |
| Item Group | 物品分组 | 物品分类 |
| Item Attribute | 物品属性 | 多规格商品 |
| POS Profile | POS 配置 | 收银配置 |
| POS Opening Entry | POS 开账 | 开账记录 |
| POS Closing Entry | POS 关账 | 关账记录 |
| POS Invoice | POS 发票 | 销售记录 |
| Purchase Order | 采购订单 | 采购管理 |
| Supplier | 供应商 | 供应商管理 |
| Warehouse | 仓库 | 仓库管理 |
| Stock Entry | 库存条目 | 库存变动 |
| Stock Reconciliation | 库存盘点 | 盘点管理 |
| Delivery Note | 发货单 | 发货管理 |
| Company | 公司 | 公司管理 |
| Mode of Payment | 支付方式 | 支付配置 |
| BOM | 物料清单 | 生产配方 |

---

## 4. 技术架构

### 4.1 架构图

```
┌─────────────────────┐
│    POS Terminal     │
│    (ttpos-main)     │
└──────────┬──────────┘
           │ gRPC
           ▼
┌─────────────────────────────────────────┐
│              ttpos-erp                  │
│  ┌─────────────────────────────────┐    │
│  │       gRPC Controller           │    │
│  │  (Item/Selling/Buying/Stock)    │    │
│  └─────────────┬───────────────────┘    │
│                │                        │
│  ┌─────────────▼───────────────────┐    │
│  │         Logic Layer             │    │
│  │   ┌─────────────────────────┐   │    │
│  │   │    ERPNext Service      │   │    │
│  │   │  (Document/RPC/Report)  │   │    │
│  │   └───────────┬─────────────┘   │    │
│  └───────────────┼─────────────────┘    │
│                  │                      │
│  ┌───────────────▼─────────────────┐    │
│  │        HTTP Client              │    │
│  │   (Site Authorization)          │    │
│  └───────────────┬─────────────────┘    │
└──────────────────┼──────────────────────┘
                   │ HTTP REST API
                   ▼
┌─────────────────────────────────────────┐
│             ERPNext                     │
│  ┌─────────────────────────────────┐    │
│  │      Frappe REST API            │    │
│  │   /api/resource/{doctype}       │    │
│  │   /api/method/{method}          │    │
│  └─────────────────────────────────┘    │
└─────────────────────────────────────────┘
```

### 4.2 核心组件

#### 4.2.1 gRPC 服务层
- 位置：`internal/controller/rpc/`
- 职责：接收 gRPC 请求，参数验证，调用业务逻辑层

#### 4.2.2 业务逻辑层
- 位置：`internal/logic/`
- 子模块：
  - `erpnext/` - ERPNext API 封装
  - `stock/` - 物品与库存管理
  - `selling/` - 销售管理
  - `buying/` - 采购管理
  - `company/` - 公司管理
  - `manufacturing/` - 生产管理

#### 4.2.3 数据传输对象
- 位置：`internal/model/dto/erp/`
- 职责：ERPNext 数据结构映射

#### 4.2.4 消息队列消费者
- 位置：`internal/consumer/`
- 职责：处理异步任务（物品同步、销售记录等）

---

## 5. 配置管理

### 5.1 配置文件示例

```yaml
# manifest/config/config.tpl.yaml
server:
  address: ":14021"
  logPath: "./log"

# gRPC 服务配置
grpc:
  name: "ttpos-erp"
  address: ":14022"
  logPath: "./log"

# 数据库配置
database:
  default:
    link: "mysql:$DB_USERNAME:$DB_PASSWORD@tcp($DB_HOST:$DB_PORT)/ttpos_erp"
    maxIdle: 10
    maxOpen: 100
    maxLifetime: 30

# ERPNext 配置
app:
  erpnext:
    serviceUrl: "$ERPNEXT_URL"
    dump: false

# RocketMQ 配置
rocketmq:
  nameServers:
    - "$ROCKETMQ_HOST:9876"
  groupName: "ttpos-erp-group"

# Nacos 配置
nacos:
  serverAddr: "$NACOS_SERVER_ADDR"
  namespace: "$NACOS_NAMESPACE"
  group: "DEFAULT_GROUP"
  dataId: "ttpos-erp.yaml"

# 日志配置
logger:
  level: "all"
  stdout: true
```

---

## 6. 开发计划

### 6.1 Phase 1 - 基础架构（已完成）
- ✅ 项目目录结构
- ✅ ERPNext API 封装
- ✅ 多站点支持
- ✅ gRPC protobuf 定义

### 6.2 Phase 2 - 核心功能（已完成）
- ✅ 物品管理服务
- ✅ 销售管理服务
- ✅ 采购管理服务
- ✅ 库存管理服务

### 6.3 Phase 3 - 异步处理（已完成）
- ✅ RocketMQ 集成
- ✅ 物品同步消费者
- ✅ 销售记录消费者

### 6.4 Phase 4 - 持续优化
- [ ] 缓存机制优化
- [ ] 批量操作支持
- [ ] 性能监控
- [ ] 错误重试机制

---

## 7. 非功能性需求

### 7.1 性能要求
- gRPC 接口响应时间 < 500ms
- 支持 100 TPS 并发请求
- 物品同步延迟 < 30s

### 7.2 可用性要求
- 服务可用性 > 99.9%
- 支持水平扩展
- ERPNext 连接失败自动重试

### 7.3 安全要求
- API Key/Secret 加密存储
- 站点隔离，租户数据隔离
- 支持收银员身份模拟审计

### 7.4 可维护性
- 完善的日志记录
- 清晰的错误信息
- 代码注释完整
- 遵循 GoFrame 开发规范

---

## 8. 依赖服务

### 8.1 必需服务
- MySQL 数据库
- ERPNext 系统
- Nacos 服务注册与配置中心

### 8.2 可选服务
- RocketMQ 消息队列（异步处理）

---

## 9. 接口调用示例

### 9.1 获取物品列表

```go
// 调用方代码示例
req := &item.GetItemListReq{
    CompanyAbbr: "CFG",
    Branch:      "Wallace Burger (CFG)",
    ItemGroup:   item.ItemGroup_Products,
}

resp, err := itemClient.GetItemList(ctx, req)
if err != nil {
    g.Log().Error(ctx, "获取物品列表失败", err)
    return
}

for _, item := range resp.ItemList {
    g.Log().Infof(ctx, "物品: %s - %s", item.ItemCode, item.ItemName)
}
```

### 9.2 POS 开账

```go
req := &selling.OpenPosEntryReq{
    PosProfileName:  "Wallace Burger POS",
    CashierEmail:    "cashier@example.com",
    CompanyAbbr:     "CFG",
    PeriodStartDate: time.Now().Unix(),
    OpenPosEntryDetail: []*selling.OpenPosEntryDetail{
        {ModeOfPayment: "Cash", OpeningAmount: 1000},
    },
}

resp, err := sellingClient.OpenPosEntry(ctx, req)
if err != nil {
    g.Log().Error(ctx, "POS开账失败", err)
    return
}

g.Log().Infof(ctx, "开账成功: %s", resp.OpenPosEntryInfo.OpenPosEntryName)
```

---

## 10. 附录

### 10.1 站点编码说明
- `0` - UAT 测试站点
- `1` - TTPOS 正式站点
- `4` - Wallace 站点

### 10.2 文档状态说明
- `0` - Draft（草稿）
- `1` - Submitted（已提交）
- `2` - Cancelled（已取消）

### 10.3 ERPNext API 文档
- 官方文档：https://frappeframework.com/docs/
- REST API：https://frappeframework.com/docs/user/en/api/rest

