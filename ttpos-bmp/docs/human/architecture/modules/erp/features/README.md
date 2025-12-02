# TTPOS ERP 功能模块索引

## 📋 模块概览

ttpos-erp 模块提供与 ERPNext 系统的完整集成能力，包含以下核心功能模块：

## 🔧 核心服务

| 模块 | 文档 | 说明 |
|-----|------|------|
| ERPNext 集成 | [erpnext-service.md](./erpnext-service.md) | ERPNext API 封装、多站点支持、身份模拟 |
| 库存管理 | [stock-service.md](./stock-service.md) | 物品 CRUD、库存查询、多规格商品、仓库管理、物料转移 |
| 销售管理 | [selling-service.md](./selling-service.md) | POS 开关账、发票管理、退货处理、销售订单、客户管理 |
| 采购管理 | [buying-service.md](./buying-service.md) | 采购订单、供应商管理、采购收货、内部交易 |
| 公司管理 | [company-service.md](./company-service.md) | 公司信息查询、子公司管理 |
| 设置服务 | [setup-service.md](./setup-service.md) | 店铺初始化、POS Profile 创建、用户管理、文档初始化 |
| 制造管理 | [manufacturing-service.md](./manufacturing-service.md) | BOM（物料清单）管理 |
| 权限管理 | [permission-service.md](./permission-service.md) | POS 权限规则管理 |
| 核心服务 | [core-service.md](./core-service.md) | 用户管理、价格表管理 |
| CRM 服务 | [crm-service.md](./crm-service.md) | 联系人管理、地址管理 |

## 📊 模块依赖关系

```
                    ┌─────────────────┐
                    │  gRPC Gateway   │
                    └────────┬────────┘
                             │
        ┌────────────────────┼────────────────────┐
        │                    │                    │
        ▼                    ▼                    ▼
┌───────────────┐   ┌───────────────┐   ┌───────────────┐
│ Stock Service │   │Selling Service│   │Buying Service │
│ (Item/Warehouse│   │               │   │               │
│  /Stock)      │   │               │   │               │
└───────┬───────┘   └───────┬───────┘   └───────┬───────┘
        │                   │                   │
        └───────────────────┼───────────────────┘
                            │
        ┌───────────────────┼───────────────────┐
        │                   │                   │
        ▼                   ▼                   ▼
┌───────────────┐   ┌───────────────┐   ┌───────────────┐
│Company Service│   │ Setup Service │   │Manufacturing  │
│               │   │               │   │  (BOM)        │
└───────┬───────┘   └───────┬───────┘   └───────┬───────┘
        │                   │                   │
        └───────────────────┼───────────────────┘
                            │
                            ▼
                ┌───────────────────────┐
                │   ERPNext Service     │
                │  (Document/RPC/Report)│
                └───────────┬───────────┘
                            │
                            ▼
                ┌───────────────────────┐
                │      ERPNext          │
                │   (Frappe REST API)   │
                └───────────────────────┘
```

## 🎯 业务流程

### POS 销售完整流程

```
1. 初始化
   ├── 创建 POS Profile
   ├── 创建支付方式账户
   └── 绑定可用用户

2. 日常运营
   ├── 收银员开账 (OpenPosEntry)
   ├── 销售结账 (SavePosInvoice)
   │   ├── 商品发票
   │   └── 物品发票 (如有原材料)
   ├── 退货处理 (ReturnPosInvoice)
   └── 收银员关账 (ClosePosEntry)

3. 数据同步
   ├── 物品同步到 ERPNext
   └── 销售数据同步
```

### 物品管理流程

```
1. 物品创建
   ├── 创建模板商品 (HasVariants=true)
   └── 创建变体商品 (CreateSingleVariantItem)

2. 物品分类
   ├── 商品 (CFG Products)
   ├── 原材料 (Raw Material)
   ├── 套餐 (Package)
   ├── POS 属性 (Pos Attribute)
   └── POS 加料 (Pos Addon)

3. 库存管理
   ├── 库存查询 (GetItemStock)
   ├── 盘点调整 (StockReconciliation)
   └── 物料转移 (MaterialTransfer)
```

## 📁 代码目录结构

```
app/ttpos-erp/
├── api/                    # Protobuf 生成的 API 定义
│   ├── item/              # 物品服务 API
│   ├── selling/           # 销售服务 API
│   ├── buying/            # 采购服务 API
│   ├── stock/             # 库存服务 API
│   ├── company/           # 公司服务 API
│   ├── warehouse/         # 仓库服务 API
│   ├── manufacturing/     # 制造服务 API
│   ├── setup/             # 设置服务 API
│   └── permission/        # 权限服务 API
├── internal/
│   ├── controller/rpc/    # gRPC 控制器
│   ├── logic/             # 业务逻辑层
│   │   ├── erpnext/       # ERPNext 集成
│   │   │   ├── erpnext.go    # RPC 服务
│   │   │   ├── document.go   # Document CRUD
│   │   │   ├── doctype.go    # DocType 统计
│   │   │   ├── report.go     # 报表查询
│   │   │   ├── token.go      # 授权管理
│   │   │   ├── resource.go   # 资源文件
│   │   │   └── print_format.go # 打印格式
│   │   ├── stock/         # 库存/物品逻辑
│   │   │   ├── item.go        # 物品管理
│   │   │   ├── stock.go        # 库存查询
│   │   │   ├── warehouse.go   # 仓库管理
│   │   │   ├── item_group.go  # 物品分组
│   │   │   ├── product.go     # 商品管理
│   │   │   ├── uom.go         # 计量单位
│   │   │   ├── delivery_note.go # 发货单
│   │   │   ├── material_transfer.go # 物料转移
│   │   │   └── stock_reconciliation.go # 库存盘点
│   │   ├── selling/       # 销售逻辑
│   │   │   ├── selling.go      # POS 销售
│   │   │   ├── sale_order.go   # 销售订单
│   │   │   ├── selling_customer.go # 客户管理
│   │   │   └── async_selling.go # 异步销售
│   │   ├── buying/        # 采购逻辑
│   │   │   ├── buying.go       # 采购订单
│   │   │   ├── supplier.go     # 供应商管理
│   │   │   └── buying_create_update.go
│   │   ├── company/       # 公司逻辑
│   │   ├── setup/          # 设置逻辑
│   │   ├── manufacturing/  # 制造逻辑（BOM）
│   │   ├── permission/     # 权限逻辑
│   │   ├── core/           # 核心服务
│   │   │   ├── user.go         # 用户管理
│   │   │   └── pos_price_list.go # 价格表
│   │   └── crm/            # CRM 逻辑
│   ├── model/dto/erp/     # ERPNext 数据结构
│   └── consumer/          # 消息消费者
└── manifest/
    ├── protobuf/          # Protobuf 定义
    └── config/            # 配置文件
```

## 🔗 相关文档

- [需求文档](../requirement.md)
- [实现文档](../IMPLEMENTATION.md)
- [实体文档](../entities/)

