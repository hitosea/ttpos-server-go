# story-erp-sales-invoice-pipeline 任务清单

## 进度总览

| 项目 | 数值 |
|------|------|
| 总 SP | 5 |
| 总任务数 | 12 |
| 已完成 | 0 |
| 完成率 | 0% |

---

## Phase 1: Protobuf 和数据模型

### 1.1 定义 Sales Invoice Protobuf

| 项目 | 内容 |
|------|------|
| File | `ttpos-bmp/app/ttpos-erp/manifest/protobuf/selling/selling.proto` |
| Purpose | 定义 SaveSalesInvoiceReq/Resp 等 message |
| Requirements | 包含订单信息、商品/物品明细、支付方式、仓库映射 |
| Leverage | 现有 SavePosInvoiceReq 结构参考 |

- [ ] 完成

### 1.2 SaleOrder 新增 ERP 字段

| 项目 | 内容 |
|------|------|
| File | `main/app/model/sale_order.go`, `admin/database/migrations/` |
| Purpose | 新增 erp_sales_invoice_name, erp_payment_entry_names, erp_sync_status, erp_reverse_count, erp_stock_deducted |
| Requirements | 迁移文件 + shop_01.sql 同步 |

- [ ] 完成

### 1.3 BMP 新增 receive_sales_invoice / receive_payment_entry 表

| 项目 | 内容 |
|------|------|
| File | `ttpos-bmp/app/ttpos-erp/internal/model/entity/`, `ttpos-bmp/manifest/sql/` |
| Purpose | SI/PE 异步记录存储 |
| Requirements | 幂等唯一索引 |

- [ ] 完成

---

## Phase 2: BMP 核心逻辑

### 2.1 Sales Invoice DTO

| 项目 | 内容 |
|------|------|
| File | `ttpos-bmp/app/ttpos-erp/internal/model/dto/erp/sales_invoice.go` |
| Purpose | Sales Invoice ERPNext API 请求/响应结构 |
| Leverage | 现有 POSInvoice DTO 参考 |

- [ ] 完成

### 2.2 Payment Entry DTO

| 项目 | 内容 |
|------|------|
| File | `ttpos-bmp/app/ttpos-erp/internal/model/dto/erp/payment_entry.go` |
| Purpose | Payment Entry ERPNext API 请求/响应结构 |

- [ ] 完成

### 2.3 Sales Invoice 同步逻辑

| 项目 | 内容 |
|------|------|
| File | `ttpos-bmp/app/ttpos-erp/internal/logic/selling/sales_invoice.go` |
| Purpose | 创建 SI → 提交 SI → 创建 PE 的完整流程 |
| Requirements | 幂等、重试、状态更新 |
| Leverage | 现有 async_selling.go 模式 |

- [ ] 完成

### 2.4 Payment Entry 同步逻辑

| 项目 | 内容 |
|------|------|
| File | `ttpos-bmp/app/ttpos-erp/internal/logic/selling/payment_entry.go` |
| Purpose | 为每个支付方式创建 PE 并关联 SI |
| Requirements | 每个支付方式 1 张 PE，幂等 |

- [ ] 完成

### 2.5 MQ Consumer 实现

| 项目 | 内容 |
|------|------|
| File | `ttpos-bmp/app/ttpos-erp/internal/consumer/selling/sales_invoice_consumer.go` |
| Purpose | 消费 save-sales-invoice topic，执行 SI+PE 创建 |
| Requirements | FIFO 顺序处理，失败重试 |
| Leverage | 现有 SavePosInvoiceConsumer 模式 |

- [ ] 完成

---

## Phase 3: Main 模块集成

### 3.1 ERP RPC 客户端扩展

| 项目 | 内容 |
|------|------|
| File | `main/app/service/rpc/erp/selling.go` |
| Purpose | 新增 SaveSalesInvoice gRPC 调用 |
| Leverage | 现有 SavePosInvoice 实现 |

- [ ] 完成

### 3.2 订单服务 SaveSalesInvoice

| 项目 | 内容 |
|------|------|
| File | `main/app/service/order.go` |
| Purpose | 结账后构建 SI 请求并调用 RPC |
| Requirements | 商品/物品分离、行级仓库映射、外卖字段 |
| Leverage | 现有 SavePosInvoice 逻辑参考 |

- [ ] 完成

### 3.3 外卖 ERP 同步切换

| 项目 | 内容 |
|------|------|
| File | `main/app/modules/takeout/domain/service/takeout_erp_sync_service.go` |
| Purpose | 外卖订单从 POS Invoice 切换到 Sales Invoice |
| Requirements | 包含外卖特有字段映射（order_source, takeout_order_no 等） |

- [ ] 完成

---

## Phase 4: 测试

### 4.1 编写 BMP 单元测试

| 项目 | 内容 |
|------|------|
| File | `ttpos-bmp/app/ttpos-erp/internal/logic/selling/sales_invoice_test.go` |
| Purpose | SI/PE 创建逻辑测试 |
| Requirements | 覆盖率 >= 70% |

- [ ] 完成

### 4.2 编写 Main 单元测试

| 项目 | 内容 |
|------|------|
| File | `main/app/service/order_test.go` |
| Purpose | SaveSalesInvoice 逻辑测试 |
| Requirements | 覆盖率 >= 80% |

- [ ] 完成

---

## 提交清单

### 代码质量
- [ ] `go mod tidy` 执行
- [ ] `go fmt ./...` 执行
- [ ] `go vet ./...` 通过
- [ ] 测试通过

### 功能完整性
- [ ] 结账后 30s 内生成 SI + PE
- [ ] 幂等：重复触发只生成 1 次
- [ ] 失败重试：5min x 3次
- [ ] 外卖订单字段正确映射
- [ ] 商品和物品使用各自仓库映射

### 迁移同步
- [ ] Main: SaleOrder 新增字段迁移
- [ ] BMP: receive_sales_invoice 表迁移
- [ ] BMP: receive_payment_entry 表迁移
- [ ] shop_01.sql 已更新
