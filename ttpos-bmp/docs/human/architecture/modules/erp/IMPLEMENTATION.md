# TTPOS ERP 集成服务实现文档

## 📋 实现概览

根据需求文档，TTPOS ERP 集成服务已完成完整实现。本文档记录了实现的详细内容和关键决策。

## ✅ 已完成功能

### Phase 1: 基础架构搭建

- ✅ 创建项目目录结构（遵循 GoFrame 规范）
- ✅ 创建 Protobuf 定义文件
  - `buying/buying.proto` - 采购服务
  - `company/company.proto` - 公司服务
  - `item/item.proto` - 物品服务
  - `selling/selling.proto` - 销售服务
  - `stock/stock.proto` - 库存服务
  - `warehouse/warehouse.proto` - 仓库服务
  - `manufacturing/bom.proto` - BOM 服务
- ✅ 创建配置文件模板 (`manifest/config/config.tpl.yaml`)
- ✅ 创建开发工具脚本 (Makefile)

### Phase 2: ERPNext 集成层

- ✅ ERPNext API 封装
  - Document Service (List/Get/Create/Update/Delete)
  - RPC Service (Execute/GetSiteCode)
  - Report Service (Run)
  - Doctype Service (Count)
  - Token Service (站点授权)
- ✅ 多站点支持
  - Site 配置管理
  - 动态站点切换
  - 收银员身份模拟

### Phase 3: 核心业务功能

- ✅ 物品管理服务
  - GetItemList - 物品列表查询
  - GetItem - 单个物品查询
  - SaveItem - 物品保存（创建/更新）
  - DeleteItem - 物品删除
  - GetItemStock - 库存查询
  - SavePosAttribute - POS 属性保存
  - SavePosAddon - POS 加料保存
  - CreateSingleVariantItem - 多规格商品创建
- ✅ 销售管理服务
  - GetPosProfileList - POS 配置列表
  - CreatePosProfile - 创建 POS 配置
  - OpenPosEntry - POS 开账
  - ClosePosEntry - POS 关账
  - SavePosInvoice - 保存 POS 发票
  - ReturnPosInvoice - 退货处理
  - CancelPosInvoice - 取消发票
  - GetModeOfPaymentList - 支付方式列表
- ✅ 采购管理服务
  - CreatePurchaseOrder - 创建采购订单
  - UpdatePurchaseOrder - 更新采购订单
  - GetSupplierList - 供应商列表
  - CreateSupplier - 创建供应商
- ✅ 库存管理服务
  - GetWarehouseList - 仓库列表
  - GetDefaultWarehouse - 默认仓库
  - CreateMaterialTransfer - 物料转移
  - CreateStockReconciliation - 库存盘点
  - CreateDeliveryNote - 发货单

### Phase 4: 异步处理

- ✅ RocketMQ 集成
- ✅ 物品同步消费者 (`item_sync.go`)
- ✅ 销售记录消费者 (`selling_consumer.go`)

## 📁 核心文件说明

### 1. Protobuf 定义

**目录**: `manifest/protobuf/`

定义了多个 gRPC 服务：
- `item/` - 物品服务接口
- `selling/` - 销售服务接口
- `buying/` - 采购服务接口
- `stock/` - 库存服务接口
- `company/` - 公司服务接口
- `warehouse/` - 仓库服务接口

### 2. 常量定义

**文件**: `internal/consts/consts.go`

定义了所有常量：
- 物品分组（Raw Material/CFG Products/Pos Attribute/Pos Addon）
- 物品编码前缀（SP/WPR/TC/SXV/JLV）
- 站点编码（0-UAT/1-TTPOS/4-Wallace）
- 支付方式（Cash/Balance/Free Meal）
- 默认配置值

### 3. ERPNext 数据传输对象

**目录**: `internal/model/dto/erp/`

定义了 ERPNext DocType 对应的数据结构：
- `Item` - 物品
- `POSProfile` - POS 配置
- `POSOpeningEntry` - POS 开账
- `POSCloseEntry` - POS 关账
- `POSInvoice` - POS 发票
- `PurchaseOrder` - 采购订单
- `Supplier` - 供应商
- `Warehouse` - 仓库
- `ModeOfPayment` - 支付方式

### 4. ERPNext 集成层

**目录**: `internal/logic/erpnext/`

实现核心 ERPNext 集成：

| 文件 | 功能 |
|-----|------|
| `erpnext.go` | RPC 服务，HTTP 客户端封装，站点授权 |
| `document.go` | Document 操作（CRUD） |
| `doctype.go` | Doctype 统计查询 |
| `report.go` | 报表查询 |
| `token.go` | 站点/收银员授权管理 |
| `resource.go` | 资源文件处理 |
| `print_format.go` | 打印格式管理 |

### 5. 业务逻辑层

**目录**: `internal/logic/`

| 模块 | 文件 | 功能 |
|-----|------|------|
| stock | `item.go` | 物品管理 |
| stock | `warehouse.go` | 仓库管理 |
| stock | `stock.go` | 库存管理 |
| stock | `material_transfer.go` | 物料转移 |
| stock | `stock_reconciliation.go` | 库存盘点 |
| stock | `delivery_note.go` | 发货单 |
| selling | `selling.go` | POS 销售主逻辑 |
| selling | `sale_order.go` | 销售订单 |
| selling | `selling_customer.go` | 客户管理 |
| selling | `async_selling.go` | 异步销售处理 |
| buying | `buying.go` | 采购管理 |
| buying | `supplier.go` | 供应商管理 |
| company | `company.go` | 公司管理 |
| manufacturing | `bom.go` | BOM 管理 |
| crm | `address.go` | 地址管理 |
| crm | `contact.go` | 联系人管理 |

### 6. gRPC 控制器层

**目录**: `internal/controller/rpc/`

实现 gRPC 接口：
- `item/` - 物品服务控制器
- `selling/` - 销售服务控制器
- `buying/` - 采购服务控制器
- `stock/` - 库存服务控制器
- `company/` - 公司服务控制器
- `warehouse/` - 仓库服务控制器
- `manufacturing/` - 生产服务控制器

### 7. 消费者

**目录**: `internal/consumer/`

| 文件 | Topic | 功能 |
|-----|-------|------|
| `item_sync.go` | item-sync | 物品同步处理 |
| `selling_consumer.go` | selling-* | 销售记录异步处理 |

## 🎯 技术亮点

### 1. 多站点支持

通过 gRPC Metadata 传递站点编码，动态切换 ERPNext 站点：

```go
func GetClient(ctx context.Context) *gclient.Client {
    var c = g.Client()
    m := grpcx.Ctx.IncomingMap(ctx)
    if m.Contains(consts.ContextSiteCode) {
        serviceAuthorization, err := Rpc.GetAndProcessSiteAuthorization(ctx, m.GetVar(consts.ContextSiteCode).String())
        if err == nil {
            c.SetPrefix(serviceAuthorization.SiteUrl)
            c.SetHeader("Authorization", serviceAuthorization.Authorization)
        }
    }
    return c
}
```

### 2. 收银员身份模拟

支持以收银员身份执行 ERPNext 操作：

```go
func SetFakeUser(ctx context.Context, userEmail string) context.Context {
    ctx = context.WithValue(ctx, consts.ContextFakeUser, userEmail)
    return ctx
}
```

### 3. 统一错误处理

ERPNext 响应错误统一检测和转换：

```go
func detectError(resp *gvar.Var) (*gjson.Json, error) {
    if resp == nil || resp.IsEmpty() {
        return nil, gerror.New("调用erp接口返回空")
    }
    if j, err := gjson.DecodeToJson(resp); err == nil {
        if j.Contains("exc_type") {
            return nil, gerror.Newf("调用erp接口返回异常,exc_type:%s", j.Get("exc_type").String())
        }
        if j.Contains("errors") {
            // 解析错误列表
            return nil, gerror.Newf("调用erp接口返回异常:%s", errorMessages)
        }
        return j, nil
    }
    return nil, gerror.Wrapf(err, "调用erp接口返回解析异常")
}
```

### 4. 物品编码自动生成

根据物品类型自动生成带前缀的编码：

```go
func (s *sItem) generateItemCode(ctx context.Context, req *item.ItemInfo) (string, error) {
    switch req.ItemGroup {
    case item.ItemGroup_RawMaterial:
        return utility.GenItemCode(consts.ItemCodePrefixRawMaterial), nil  // WPR
    case item.ItemGroup_Package:
        return utility.GenItemCode(consts.ItemCodePrefixPackage), nil      // TC
    case item.ItemGroup_PosAttribute:
        return utility.GenItemCode(consts.ItemCodePrefixPosAttribute), nil // SXV
    case item.ItemGroup_PosAddon:
        return utility.GenItemCode(consts.ItemCodePrefixPosAddon), nil     // JLV
    default:
        return utility.GenItemCode(consts.ItemCodePrefixProduct), nil      // SP
    }
}
```

### 5. 多规格商品支持

创建模板商品后，可基于属性创建变体商品：

```go
func (s *sItem) CreateSingleVariantItem(ctx context.Context, req *erp.CreateSingleVariantItemReq, templateItemInfo *item.ItemInfo) (string, error) {
    // 调用 ERPNext 创建变体
    resp, err := service.Rpc().Execute(ctx, &erp.ErpReq{
        Method: erp.ApiMethodCreateVariantItem,
    }, g.Map{
        "item": req.TemplateItem,
        "args": gjson.MustEncodeString(req.Args),
    })
    // 生成变体编码并创建
    itemCode, err = s.generateItemCodeWithTemplate(ctx, req.TemplateItem)
    // ...
}
```

### 6. POS 开关账流程

完整的 POS 开账/关账生命周期管理：

```
开账流程:
1. 验证配置文件存在
2. 获取公司信息
3. 创建 POS Opening Entry（草稿）
4. 提交开账记录（状态变更为 Submitted）

关账流程:
1. 获取开账记录
2. 查询期间所有已提交的 POS Invoice
3. 创建 POS Closing Entry（草稿）
4. 关联期间发票
5. 提交关账记录
```

### 7. 异步销售记录

支持将销售记录异步推送到 ERPNext：

```go
// 同步模式
resp, err := s.SavePosInvoiceStep(ctx, req, openingEntry, false)

// 异步模式（通过消息队列）
queue.Push(string(consts.TopicSellingSave), req)
```

## 🔧 开发流程

### 1. 生成代码

```bash
# 进入模块目录
cd app/ttpos-erp

# 生成 Protobuf 代码
make proto

# 生成 DAO 代码
make dao

# 生成 Service 接口
make service
```

### 2. 配置服务

```bash
# 复制配置模板
cp manifest/config/config.tpl.yaml manifest/config/config.yaml

# 编辑配置文件，设置：
# - 数据库连接
# - ERPNext URL
# - RocketMQ 地址（可选）
# - Nacos 配置（可选）
```

### 3. 初始化数据

```bash
# 执行 ERPNext 迁移脚本
make migrate-all --siteCode 1 --dirBase ./manifest/erp-migrate/
```

### 4. 运行服务

```bash
# 开发模式运行
make run
```

## 📊 数据流程

### 物品同步流程

```
1. 管理员在后台创建/修改物品
   ↓
2. 调用 ttpos-erp SaveItem 接口
   ↓
3. 业务逻辑层处理
   - 参数验证
   - 公司信息获取
   - 物品编码生成（如新建）
   ↓
4. 调用 ERPNext Document API
   - Create 或 Update
   ↓
5. 返回结果给调用方
   ↓
6. 推送同步消息到队列（延迟）
   ↓
7. 消费者处理同步任务
```

### POS 销售流程

```
1. 收银员开账
   - 调用 OpenPosEntry
   - 创建 POS Opening Entry
   ↓
2. 收银员结账
   - 调用 SavePosInvoice
   - 创建商品发票
   - 创建物品发票（如有原材料）
   ↓
3. 退货处理（如需要）
   - 调用 ReturnPosInvoice
   - 创建退货发票
   ↓
4. 收银员关账
   - 调用 ClosePosEntry
   - 查询期间发票
   - 创建 POS Closing Entry
```

## 🔍 关键决策

### 1. 为什么不直接修改 ERPNext？

- 避免 ERPNext 升级时的冲突
- 保持 ERPNext 的标准性
- 通过中间层封装业务差异
- 便于多租户隔离

### 2. 为什么使用 gRPC？

- 高性能的二进制协议
- 强类型的接口定义
- 双向流支持
- 与 GoFrame 集成良好

### 3. 为什么物品和商品分开发票？

- ERPNext 物品发票需要更新库存
- 商品发票不需要更新库存
- 分开处理便于错误回滚

### 4. 为什么需要收银员身份模拟？

- ERPNext 权限控制基于用户
- POS 发票需要绑定创建者
- 便于审计和追踪

## 🚀 后续优化

### 短期优化

1. 物品缓存机制
2. ERPNext 连接池
3. 批量操作支持
4. 错误重试机制

### 长期优化

1. 离线模式支持
2. 数据同步对账
3. 性能监控面板
4. API 限流保护

## 📝 注意事项

1. **ERPNext 版本**：确保 ERPNext 版本兼容（推荐 v14+）
2. **站点配置**：erp_site 表需要正确配置站点信息
3. **权限配置**：确保 API Key 有足够权限
4. **时区处理**：注意 ERPNext 时区与本地时区差异
5. **并发处理**：注意 ERPNext API 的并发限制

## 🎉 总结

ttpos-erp 模块作为 TTPOS 与 ERPNext 的桥梁，提供了完整的 ERP 功能支撑。

### 技术特点

- **多站点支持**：动态切换 ERPNext 站点
- **身份模拟**：支持收银员身份执行操作
- **统一封装**：标准化的 ERPNext API 调用
- **异步处理**：消息队列支持高并发

### 设计优势

- **解耦合**：POS 系统与 ERPNext 解耦
- **可扩展**：易于添加新的业务功能
- **可维护**：清晰的分层架构
- **高可用**：支持水平扩展

### 业务价值

- **一体化**：POS + ERP 无缝集成
- **实时性**：销售数据实时同步
- **可追溯**：完整的操作审计
- **灵活性**：支持多种业务场景

