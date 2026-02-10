# story-shop-store-summary-metrics-enhancement 任务清单

## 📊 进度总览

| 项目 | 数值 |
|------|------|
| 总 SP | 5 |
| 总任务数 | 12 |
| 已完成 | 12 |
| 完成率 | 100% |

---

## Phase 1: Resp 结构体扩展

### 1.1 新增 Average 结构体

| 项目 | 内容 |
|------|------|
| File | `main/app/dto/resp/statistics_summary_resp.go` |
| Purpose | 定义营业数据、退款金额、支付方式的平均值结构体 |
| Requirements | Req1, Req2 |
| Leverage | 参考现有 SummaryItem 结构体字段设计 |

**详细任务**:
- 新增 `CompanyBusinessAverage` 结构体（15个字段）
- 新增 `CompanyRefundAverage` 结构体（8个字段）
- 新增 `CompanyPaymentMethodAverage` 结构体（3个字段）

- [x] 完成

### 1.2 扩展 Resp 结构体

| 项目 | 内容 |
|------|------|
| File | `main/app/dto/resp/statistics_summary_resp.go` |
| Purpose | 在三个 Resp 类型中增加 Average 字段 |
| Requirements | Req1, Req2 |
| Leverage | 现有 CompanyBusinessSummaryResp 等结构体 |

**详细任务**:
- CompanyBusinessSummaryResp 增加 `Average *CompanyBusinessAverage`
- CompanyRefundSummaryResp 增加 `Average *CompanyRefundAverage`
- CompanyPaymentMethodSummaryResp 增加 `Average *CompanyPaymentMethodAverage`

- [x] 完成

### 1.3 扩展 RefundSummaryItem 结构体

| 项目 | 内容 |
|------|------|
| File | `main/app/dto/resp/statistics_summary_resp.go` |
| Purpose | 在退款汇总项中增加平均退款额字段 |
| Requirements | Req3 |
| Leverage | 现有 CompanyRefundSummaryItem 结构体 |

**详细任务**:
- 新增 `AvgRefundAmount float64` 字段
- 添加 JSON tag: `json:"avg_refund_amount"`

- [x] 完成

---

## Phase 2: Service 平均值计算逻辑

### 2.1 实现平均值计算辅助函数

| 项目 | 内容 |
|------|------|
| File | `main/app/service/business.go` |
| Purpose | 封装平均值计算逻辑，处理除数为0等边界情况 |
| Requirements | Req1, Req2 |
| Leverage | 无，新增私有函数 |

**详细任务**:
- 新增 `calculateBusinessAverage(items []resp.CompanyBusinessSummaryItem) *resp.CompanyBusinessAverage`
- 新增 `calculateRefundAverage(items []resp.CompanyRefundSummaryItem) *resp.CompanyRefundAverage`
- 新增 `calculatePaymentMethodAverage(items []resp.CompanyPaymentMethodSummaryItem) *resp.CompanyPaymentMethodAverage`
- 使用 decimal 库保证精度，保留2位小数

- [x] 完成

### 2.2 修改 CountCompanyBusinessSummary 方法

| 项目 | 内容 |
|------|------|
| File | `main/app/service/business.go` |
| Purpose | 在返回结果中填充 Average 字段 |
| Requirements | Req1, Req2 |
| Leverage | 现有 CountCompanyBusinessSummary 方法（行号 4152-4630） |

**详细任务**:
- 在营业数据分支中调用 calculateBusinessAverage
- 在支付方式分支中调用 calculatePaymentMethodAverage
- 在退款分支中调用 calculateRefundAverage
- 平均值基于全量数据计算（非分页数据）

- [x] 完成

### 2.3 修改 countCompanyRefundSummary 方法

| 项目 | 内容 |
|------|------|
| File | `main/app/service/business.go` |
| Purpose | 在每条退款记录中计算并填充平均退款额 |
| Requirements | Req3 |
| Leverage | 现有 countCompanyRefundSummary 方法（行号 5066-5240） |

**详细任务**:
- 遍历 items，计算 `AvgRefundAmount = RefundAmount / RefundNum`
- 处理 RefundNum = 0 的情况，返回 0

- [x] 完成

---

## Phase 3: 导出功能扩展

### 3.1 修改 exportBusinessSummaryToExcel 方法

| 项目 | 内容 |
|------|------|
| File | `main/app/service/business.go` |
| Purpose | 在 Excel 明细表和汇总表末尾增加平均值行 |
| Requirements | Req4 |
| Leverage | 现有 exportBusinessSummaryToExcel 方法 |

**详细任务**:
- 获取平均值数据
- 在数据行后追加一行，第一列为"平均值"（多语言）
- 后续列填充对应的平均值
- 可选：设置平均值行粗体样式

- [x] 完成

### 3.2 修改 exportRefundSummaryToExcel 方法

| 项目 | 内容 |
|------|------|
| File | `main/app/service/business.go` |
| Purpose | 在 Excel 中增加平均退款额列，并追加平均值行 |
| Requirements | Req3, Req4 |
| Leverage | 现有 exportRefundSummaryToExcel 方法 |

**详细任务**:
- 在表头增加"平均退款额"列（多语言）
- 在数据行中填充 AvgRefundAmount 字段
- 在末尾追加平均值行

- [x] 完成

### 3.3 修改 exportPaymentMethodSummaryToExcel 方法

| 项目 | 内容 |
|------|------|
| File | `main/app/service/business.go` |
| Purpose | 在 Excel 末尾增加平均值行 |
| Requirements | Req4 |
| Leverage | 现有 exportPaymentMethodSummaryToExcel 方法 |

**详细任务**:
- 获取平均值数据
- 在数据行后追加平均值行

- [x] 完成

---

## Phase 4: 测试与文档

### 4.1 编写单元测试

| 项目 | 内容 |
|------|------|
| File | `main/app/service/business_summary_test.go` |
| Purpose | 覆盖平均值计算和边界情况 |
| Requirements | 覆盖率 ≥ 80% |
| Leverage | 现有测试文件（245行） |

**测试用例**:
- [x] `TestCalculateBusinessAverage_Normal` - 正常数据
- [x] `TestCalculateBusinessAverage_Empty` - 空数据返回0
- [x] `TestCalculateRefundAverage_Normal` - 正常退款数据
- [x] `TestCalculateRefundAverage_Empty` - 空数据返回0
- [x] `TestAvgRefundAmountCalculation` - 平均退款额计算（含除数为0场景）
- [x] `TestDeliveryAmountCalculation` - 外卖业绩计算（含外卖平台订单）
- [x] `TestInStoreAmountCalculation` - 到店业绩计算
- [x] `TestCalculatePaymentMethodAverage_Normal` - 支付方式平均值正常数据
- [x] `TestCalculatePaymentMethodAverage_Empty` - 支付方式空数据返回0

- [x] 完成

---

## Phase 5: 业绩分类字段扩展（补充需求）

### 5.1 扩展 BusinessSummaryItem 结构体

| 项目 | 内容 |
|------|------|
| File | `main/app/dto/resp/statistics_summary_resp.go` |
| Purpose | 在营业数据汇总项中增加业绩分类字段 |
| Requirements | Req5 |
| Leverage | 现有 CompanyBusinessSummaryItem 结构体 |

**详细任务**:
- 新增 `InStoreAmount float64` 字段（到店业绩）
- 新增 `InStoreOrderNum int64` 字段（到店订单数）
- 新增 `DeliveryAmount float64` 字段（外卖业绩）
- 新增 `DeliveryOrderNum int64` 字段（外卖订单数）
- 字段位置：在 CashAC 后面

- [x] 完成

### 5.2 扩展 BusinessAverage 结构体

| 项目 | 内容 |
|------|------|
| File | `main/app/dto/resp/statistics_summary_resp.go` |
| Purpose | 在营业数据平均值中增加业绩分类字段 |
| Requirements | Req5 |
| Leverage | 现有 CompanyBusinessAverage 结构体 |

**详细任务**:
- 新增 `InStoreAmount float64` 字段
- 新增 `InStoreOrderNum float64` 字段
- 新增 `DeliveryAmount float64` 字段
- 新增 `DeliveryOrderNum float64` 字段

- [x] 完成

### 5.3 修改 Service 计算逻辑

| 项目 | 内容 |
|------|------|
| File | `main/app/service/business.go` |
| Purpose | 在查询和计算逻辑中填充新字段 |
| Requirements | Req5 |
| Leverage | 现有 countCompanyBusinessSummary 方法 |

**详细任务**:
- 计算到店业绩 = 桌台订单金额(DeskOrderAmount) + 点餐订单金额(InstantOrderAmount)
- 计算到店订单数 = 桌台订单数(DeskNum) + 点餐订单数(InstantOrderNum)
- 计算外卖业绩 = 外送订单金额(TakeoutOrderAmount) + 第三方外卖订单金额(InstantOrderTakeawayAmount)
- 计算外卖订单数 = 外送订单数(TakeoutOrderNum) + 第三方外卖订单数(InstantOrderTakeawayNum)
- 更新 calculateBusinessAverage 函数包含新字段

- [x] 完成

### 5.4 修改 Excel 导出

| 项目 | 内容 |
|------|------|
| File | `main/app/service/business.go` |
| Purpose | 在 Excel 导出中增加新列 |
| Requirements | Req5 |
| Leverage | 现有 exportBusinessSummaryToExcel 方法 |

**详细任务**:
- 在表头"现金AC"后增加4个新列（到店业绩、到店订单数、外卖业绩、外卖订单数）
- 在数据行中填充对应字段值
- 在平均值行中填充对应平均值
- 更新所有9种语言的表头（zh, en, th, zhtw, ja, ko, my, tr, sv）

- [x] 完成

---

## 📝 多语言文案

需要添加的多语言 Key:

| Key | 中文 | English | 泰文 |
|-----|------|---------|------|
| average | 平均值 | Average | ค่าเฉลี่ย |
| avg_refund_amount | 平均退款额 | Avg Refund Amount | จำนวนคืนเงินเฉลี่ย |
| in_store_amount | 到店业绩 | In-store Amount | ยอดขายหน้าร้าน |
| in_store_order_num | 到店订单数 | In-store Orders | จำนวนออเดอร์หน้าร้าน |
| delivery_amount | 外卖业绩 | Delivery Amount | ยอดขายเดลิเวอรี่ |
| delivery_order_num | 外卖订单数 | Delivery Orders | จำนวนออเดอร์เดลิเวอรี่ |

---

## 提交清单

### 代码质量
- [x] `go mod tidy` 执行
- [x] `go fmt ./...` 执行
- [x] `go vet ./...` 通过（现有警告来自其他文件）
- [x] 测试通过: `go test ./...`

### 功能完整性
- [x] 营业数据汇总表底部显示平均值行
- [x] 退款金额汇总表底部显示平均值行
- [x] 支付方式汇总表底部显示平均值行
- [x] 退款表包含平均退款额列
- [x] 平均值基于筛选全量数据计算
- [x] 导出 Excel 包含平均值行
- [x] API 响应格式正确（data 为对象）
- [x] 营业数据表包含到店业绩、到店订单数、外卖业绩、外卖订单数四列
- [x] 导出 Excel 包含业绩分类四列

### 迁移同步
- [x] 无需数据库迁移（仅修改响应结构）

---

**版本**: v1.1.0
**创建日期**: 2026-02-09
**更新日志**:
- v1.1.0 (2026-02-09): 新增 Phase 5 业绩分类字段扩展（4个新任务）
