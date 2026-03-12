# story-all-erp-mobile-order-sync 任务清单

## 📊 进度总览

| 项目 | 数值 |
|------|------|
| 总 SP | 3 |
| 总任务数 | 6 |
| 已完成 | 0 |
| 完成率 | 0% |

---

## Phase 1: 核心实现

### 1.1 扩展 SaveSalesInvoiceReq 增加订单来源字段

| 项目 | 内容 |
|------|------|
| File | `main/app/dto/req/erpnext.go` |
| Purpose | 在 SI 请求中传递 order_source_uuid 和 order_source_name |
| Requirements | Req4: Sales Invoice 订单来源标识 |
| Leverage | 现有 SaveSalesInvoiceReq 结构体 |

**变更内容**:
- 在 `SaveSalesInvoiceReq` 结构体中新增 `OrderSourceUuid` 和 `OrderSourceName` 字段

- [ ] 完成

### 1.2 SaveSalesInvoice 透传订单来源

| 项目 | 内容 |
|------|------|
| File | `main/app/service/order_erp_sales_invoice.go` |
| Purpose | 构建 SI 请求时从 SaleBill 读取来源信息并填入请求 |
| Requirements | Req4: Sales Invoice 订单来源标识 |
| Leverage | `SaleBill.OrderSourceUuid` + `SaleBill.OrderSourceName` |

**变更内容**:
- 在 `SaveSalesInvoice` 方法中，构建 `SaveSalesInvoiceReq` 时填入 `OrderSourceUuid` 和 `OrderSourceName`

- [ ] 完成

### 1.3 AcceptH5Order 接单后触发 ERP 推送

| 项目 | 内容 |
|------|------|
| File | `main/app/service/order_h5.go` |
| Purpose | 会员/扫码点餐订单接单成功后，若已结账则推送 ERP |
| Requirements | Req1: 接单后推送 ERP |
| Leverage | `order_pay.go` 的 ERP 推送判断逻辑模式 |

**变更内容**:
- 在 `AcceptH5Order` 接单事务完成后，增加 ERP 推送逻辑
- 判断条件：`company.IsOpenErpPhase3()` + `companySetting.ErpnextSiteCode != ""`
- 判断新旧方案：`companySetting.IsErpSalesInvoiceMode()` + `currentShift.IsNewShiftVersion()`
- **前提**：订单已结账（有 PaymentOrders），未结账的订单由结账时触发
- 调用 `SaveSalesInvoice` 或 `SavePosInvoice`
- ERP 推送失败不影响接单结果（异步容错）

- [ ] 完成

### 1.4 BMP 透传订单来源到 ERPNext

| 项目 | 内容 |
|------|------|
| File | `ttpos-bmp/app/ttpos-erp/api/selling/` + `ttpos-bmp/app/ttpos-erp/internal/logic/selling/sales_invoice.go` |
| Purpose | BMP 端接收并透传 order_source 字段到 ERPNext Sales Invoice |
| Requirements | Req4: Sales Invoice 订单来源标识 |
| Leverage | 现有 protobuf + SaveSalesInvoice 逻辑 |

**变更内容**:
- protobuf 定义中增加 `order_source_uuid` 和 `order_source_name` 字段
- `SaveSalesInvoice` logic 中将字段传入 ERPNext 文档的自定义字段
- 执行 `make pb` 重新生成代码

- [ ] 完成

---

## Phase 2: 测试验证

### 2.1 端到端测试

| 项目 | 内容 |
|------|------|
| Purpose | 验证会员/扫码订单在接单和结账两种场景下正确推送 ERP |
| Requirements | Req1 + Req2 + Req3 全部验收标准 |

**测试场景**:
1. 开启接单 + 新班次 → 接单后走 Sales Invoice
2. 开启接单 + 旧班次 → 接单后走 POS Invoice
3. 未开启接单 + 新班次 → 结账后走 Sales Invoice（已有逻辑）
4. 未开启接单 + 旧班次 → 结账后走 POS Invoice（已有逻辑）
5. 接单时未结账 → 不推 ERP，等结账时触发
6. 验证 SI 包含正确的 order_source_uuid 和 order_source_name

- [ ] 完成

### 2.2 代码质量检查

| 项目 | 内容 |
|------|------|
| Purpose | 确保代码符合项目规范 |

- [ ] `go mod tidy` 执行
- [ ] `go fmt ./...` 执行
- [ ] `go vet ./...` 通过
- [ ] 测试通过: `cd main && go test ./app/service/...`

- [ ] 完成

---

## 提交清单

### 代码质量
- [ ] `go mod tidy` 执行
- [ ] `go fmt ./...` 执行
- [ ] `go vet ./...` 通过
- [ ] 测试通过

### 功能完整性
- [ ] AC1: 开启接单 → 接单后推送 ERP ✓
- [ ] AC2: 自动接单 → 同样触发 ERP 推送 ✓
- [ ] AC3: 未开启接单 → 结账后推送 ERP ✓
- [ ] AC4: 旧班次 → 走旧方案 ✓
- [ ] AC5: 新班次 → 走新方案 ✓
- [ ] AC6: 切换不可逆 ✓
- [ ] AC7: SI 包含 order_source_uuid 和 order_source_name ✓
