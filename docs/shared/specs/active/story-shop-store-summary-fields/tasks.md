# story-shop-store-summary-fields 任务清单

## 📊 进度总览

| 项目 | 数值 |
|------|------|
| 总 SP | 3 |
| 总任务数 | 8 |
| 已完成 | 8 |
| 完成率 | 100% |

---

## Phase 1: 核心实现

### 1.1 修改响应 DTO 字段

| 项目 | 内容 |
|------|------|
| File | `main/app/dto/resp/statistics_summary_resp.go` |
| Purpose | 新增现金统计字段、店铺编号字段 |
| Requirements | R2 新增现金统计字段, R4 店铺名称格式优化 |
| Leverage | 现有 CompanyBusinessSummaryItem 结构体 |

**变更内容**:
- 新增字段: cash_tc, cash_amount, cash_ac (到 CompanyBusinessSummaryItem)
- 新增字段: store_code (到 CompanySummaryItem)
- 注：保留原 JSON Key，避免 Breaking Change

- [x] 完成

### 1.2 实现现金统计计算逻辑

| 项目 | 内容 |
|------|------|
| File | `main/app/service/business.go` |
| Purpose | 计算现金TC、现金金额、现金AC |
| Requirements | R2 新增现金统计字段 |
| Leverage | 现有 CountBusinessPaymentMethod 方法 |

**实现要点**:
- 现金TC: 使用 PaymentMethodNames: []string{"Cash"} 查询现金支付订单数
- 现金金额: SUM(amount WHERE payment_method='Cash')
- 现金AC: 现金金额/现金TC，保留2位小数，除零返回0.00

- [x] 完成

### 1.3 实现店铺名称格式化

| 项目 | 内容 |
|------|------|
| File | `main/app/service/business.go` |
| Purpose | 格式化店铺名称为"编号 名称"格式 |
| Requirements | R4 店铺名称格式优化 |
| Leverage | 现有 GetStoreSetting 方法获取 store_code |

**实现要点**:
- 有编号: `{编号} {名称}`
- 无编号: `{名称}`

- [x] 完成

### 1.4 实现店铺排序逻辑

| 项目 | 内容 |
|------|------|
| File | `main/app/service/business.go` |
| Purpose | 按规则排序店铺列表 |
| Requirements | R4 店铺排序规则 |
| Leverage | Go sort.Slice 函数 |

**排序规则**:
1. 无编号优先
2. 数字(0-9)优先
3. 字母(a-z)其次

- [x] 完成

---

## Phase 2: API 层集成

### 2.1 更新门店汇总统计接口

| 项目 | 内容 |
|------|------|
| File | `main/app/api/v1/shop/shop_statistics.go` |
| Purpose | 集成新的响应结构 |
| Requirements | R1-R4 |
| Leverage | 现有 API Handler (pass-through) |

**变更内容**:
- API Handler 直接传递 Service 层响应，无需修改
- 新字段通过 DTO 结构体自动包含在 JSON 响应中

- [x] 完成

### 2.2 更新导出功能

| 项目 | 内容 |
|------|------|
| File | `main/app/service/business.go` (exportBusinessSummaryToExcel) |
| Purpose | 同步调整 Excel 导出字段 |
| Requirements | R5 导出功能同步调整 |
| Leverage | 现有 excelize 导出逻辑 |

**变更内容**:
- 表头名称同步修改 (门店名称→店铺名称, 订单金额→总营业额, 实付金额→实收金额, 订单量→TC, 订单金额单均→AC)
- 新增现金统计列 (现金TC, 现金金额, 现金AC)
- 字段顺序与需求一致 (营业日、店铺名称、总营业额、实收金额、TC、现金TC、现金金额、现金AC...)

- [x] 完成

---

## Phase 3: 测试与文档

### 3.1 编写单元测试

| 项目 | 内容 |
|------|------|
| File | `main/app/service/business_summary_test.go` |
| Purpose | 单元测试覆盖 |
| Requirements | 覆盖排序、格式化、计算逻辑 |

**测试用例**:
- [x] 现金统计计算（有现金订单）
- [x] 现金统计计算（无现金订单，除零场景）
- [x] 店铺名称格式化（有编号）
- [x] 店铺名称格式化（无编号）
- [x] 店铺排序规则验证（无编号优先、数字优先于字母）

- [x] 完成

### 3.2 更新 API 文档

| 项目 | 内容 |
|------|------|
| File | Swagger/API 文档 (DTO 注释) |
| Purpose | 更新接口文档 |
| Requirements | 文档与实现一致 |

- [x] 完成 (DTO 注释已更新，运行 swag init 即可生成)

---

## 提交清单

### 代码质量
- [x] `go mod tidy` 执行
- [x] `go fmt ./...` 执行
- [x] `go vet ./...` 通过 (除预先存在的警告)
- [x] 测试通过: `go test ./app/service/...`

### 功能完整性
- [x] 所有验收标准通过
- [x] API 响应格式正确（data 为对象）
- [x] 导出字段与页面一致

### 前后端协同
- [ ] 与前端确认字段变更 (待完成)
- [ ] 灰度发布验证 (待完成)

---

## 修改的文件清单

| 文件 | 变更类型 | 说明 |
|------|----------|------|
| `main/app/dto/resp/statistics_summary_resp.go` | 修改 | 新增 CashTC, CashAmount, CashAC, StoreCode 字段 |
| `main/app/service/business.go` | 修改 | 新增现金统计查询、店铺名称格式化、排序逻辑、导出列更新 |
| `main/app/service/business_summary_test.go` | 新增 | 单元测试文件 |

---

**版本**: v1.0.0
**创建日期**: 2026-02-03
**完成日期**: 2026-02-03
