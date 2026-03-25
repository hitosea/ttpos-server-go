# story-erp-reverse-settle-refund 任务清单

## 进度总览

| 项目 | 数值 |
|------|------|
| 总 SP | 5 |
| 总任务数 | 10 |
| 已完成 | 0 |
| 完成率 | 0% |

---

## Phase 1: Protobuf 和 DTO

### 1.1 定义 Cancel/Return SalesInvoice Protobuf

| 项目 | 内容 |
|------|------|
| File | `ttpos-bmp/app/ttpos-erp/manifest/protobuf/selling/selling.proto` |
| Purpose | 定义 CancelSalesInvoiceReq/Resp, ReturnSalesInvoiceReq/Resp |
| Leverage | 现有 CancelPosInvoiceReq/ReturnPosInvoiceReq 参考 |

- [ ] 完成

### 1.2 Credit Note DTO

| 项目 | 内容 |
|------|------|
| File | `ttpos-bmp/app/ttpos-erp/internal/model/dto/erp/credit_note.go` |
| Purpose | ERPNext Credit Note API 结构 |

- [ ] 完成

### 1.3 退款 Payment Entry DTO

| 项目 | 内容 |
|------|------|
| File | `ttpos-bmp/app/ttpos-erp/internal/model/dto/erp/payment_entry.go` |
| Purpose | 退款方向的 PE 结构（payment_type=Pay） |

- [ ] 完成

---

## Phase 2: BMP 核心逻辑

### 2.1 取消 Sales Invoice 逻辑

| 项目 | 内容 |
|------|------|
| File | `ttpos-bmp/app/ttpos-erp/internal/logic/selling/cancel_sales_invoice.go` |
| Purpose | 先取消 PE → 再取消 SI |
| Requirements | 顺序保证、幂等、Stock Entry 联动 |

- [ ] 完成

### 2.2 退款 Credit Note 逻辑

| 项目 | 内容 |
|------|------|
| File | `ttpos-bmp/app/ttpos-erp/internal/logic/selling/return_sales_invoice.go` |
| Purpose | 创建 Credit Note + 退款 PE |
| Requirements | update_stock=0、全部/部分退款、幂等 |

- [ ] 完成

### 2.3 反向 Stock Entry 逻辑

| 项目 | 内容 |
|------|------|
| File | `ttpos-bmp/app/ttpos-erp/internal/logic/stock/stock_entry_merge.go` |
| Purpose | 反结账后回增库存（如已通过 Stock Entry 扣减） |
| Requirements | 生成反向 Stock Entry |

- [ ] 完成

### 2.4 MQ Consumer 实现

| 项目 | 内容 |
|------|------|
| File | `ttpos-bmp/app/ttpos-erp/internal/consumer/selling/` |
| Purpose | cancel-sales-invoice / return-sales-invoice 消费者 |
| Leverage | 现有 Consumer 模式 |

- [ ] 完成

---

## Phase 3: Main 模块集成

### 3.1 ERP RPC 客户端扩展

| 项目 | 内容 |
|------|------|
| File | `main/app/service/rpc/erp/selling.go` |
| Purpose | 新增 CancelSalesInvoice / ReturnSalesInvoice gRPC 调用 |

- [ ] 完成

### 3.2 反结账服务集成

| 项目 | 内容 |
|------|------|
| File | `main/app/service/order.go` |
| Purpose | 反结账时调用 CancelSalesInvoice，ErpReverseCount++ |
| Requirements | 单据号后缀递增 |

- [ ] 完成

### 3.3 退款服务集成

| 项目 | 内容 |
|------|------|
| File | `main/app/service/order.go` 或退款相关服务 |
| Purpose | 退款时调用 ReturnSalesInvoice |
| Requirements | 全部/部分退款、按支付方式拆分 |

- [ ] 完成

---

## 提交清单

### 代码质量
- [ ] `go mod tidy` 执行
- [ ] `go fmt ./...` 执行
- [ ] `go vet ./...` 通过
- [ ] 测试通过

### 功能完整性
- [ ] 反结账：先取消 PE 再取消 SI
- [ ] 反结账后重新下发：单据号递增
- [ ] 全部退款：Credit Note 所有行 + 退款 PE
- [ ] 部分退款：Credit Note 退款行 + 退款 PE
- [ ] Credit Note update_stock=0
- [ ] 反结账后 Stock Entry 联动回增
- [ ] 所有操作幂等

### 迁移同步
- [ ] BMP: receive_cancel_sales_invoice 表（如需独立表）
