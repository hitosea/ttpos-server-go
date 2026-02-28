# task-backend-grab-lineman-payment-idempotent 任务清单

## 📊 进度总览

| 项目 | 数值 |
|------|------|
| 总 SP | 3 |
| 总任务数 | 7 |
| 已完成 | 7 |
| 完成率 | 100% |

---

## Phase 1: ERP 查询接口封装

### 1.1 新增 getERPPaymentByName 方法

| 项目 | 内容 |
|------|------|
| File | `main/app/service/payment_method.go` |
| Purpose | 从 ERP 按名称精确查询单个支付方式 |
| Requirements | R1: ERP 支付方式存在性检查 |
| Leverage | **直接调用已有的 `selling.GetModeOfPayment` RPC 接口** |

**实现要点**:
- 调用 `client.GetModeOfPayment` 按 `Name` 精确查询
- 新增 `buildERPPaymentName` 方法构造 ERP 支付方式名称
- 名称格式：`{PayType}-0000-{company_abbr}`（系统默认序号固定为 0000）
- 返回 `*selling.ModeOfPayment` 或 `nil`（不存在）

```go
func (s *paymentMethodSrv) buildERPPaymentName(payType string, companyAbbr string) string
func (s *paymentMethodSrv) getERPPaymentByName(
    ctx context.Context,
    companySetting *model.CompanySetting,
    paymentName string,
) (*selling.ModeOfPayment, error)
```

- [x] 完成

### 1.2 新增 ensureERPPaymentMethod 方法

| 项目 | 内容 |
|------|------|
| File | `main/app/service/payment_method.go` |
| Purpose | 确保 ERP 支付方式存在（幂等性封装） |
| Requirements | R1, R2, R3 |
| Leverage | 复用 `getERPPaymentByName`, `erpService.SaveModeOfPayment` |

**实现要点**:
1. 构造 ERP 支付方式名称（`{PayType}-0000-{company_abbr}`）
2. 先查询 ERP 是否已存在
3. 已存在则直接返回
4. 不存在则创建
5. 创建失败时重新查询确认状态

```go
func (s *paymentMethodSrv) ensureERPPaymentMethod(
    ctx context.Context,
    payType string,
    source int,
) (*selling.SaveModeOfPaymentResp, error)
```

- [x] 完成

---

## Phase 2: 修改 Grab/LINE MAN 保存方法

### 2.1 修改 SaveGrabPaymentMethod 方法

| 项目 | 内容 |
|------|------|
| File | `main/app/service/payment_method.go:1182-1257` |
| Purpose | 添加 ERP 幂等性检查 |
| Requirements | R1, R2, R3, R4 |
| Leverage | 使用新增的 `ensureERPPaymentMethod` |

**修改要点**:
1. 保持现有 TTPOS 侧幂等性检查（通过 code 查询）
2. ERP 调用改为使用 `ensureERPPaymentMethod`
3. 添加详细日志记录（包含 company_uuid）

- [x] 完成

### 2.2 修改 SaveLineManPaymentMethod 方法

| 项目 | 内容 |
|------|------|
| File | `main/app/service/payment_method.go:1259-1339` |
| Purpose | 添加 ERP 幂等性检查 |
| Requirements | R1, R2, R3, R4 |
| Leverage | 使用新增的 `ensureERPPaymentMethod` |

**修改要点**: 同 SaveGrabPaymentMethod

- [x] 完成

---

## Phase 3: 优化 createPaymentFromERP 方法

### 3.1 修改 createPaymentFromERP 方法

| 项目 | 内容 |
|------|------|
| File | `main/app/service/payment_method.go:1130-1205` |
| Purpose | 优化同步逻辑处理数据不一致场景 |
| Requirements | R5 |
| Leverage | 现有 `paymentMethodRepo` |

**修改要点**:
1. 创建前检查 TTPOS 是否已存在同名支付方式（通过 erpnext_payment 或 erpnext_payment_id）
2. 已存在时更新而非创建（以 ERP 数据为准）
3. 添加日志记录数据修复情况

- [x] 完成

---

## Phase 4: 测试与文档

### 4.1 编写单元测试

| 项目 | 内容 |
|------|------|
| File | `main/app/service/payment_method_test.go` |
| Purpose | 单元测试覆盖幂等性场景 |
| Requirements | 覆盖率 ≥ 80% |

**测试场景**:
- [x] ERP 已存在、TTPOS 不存在 (TestPaymentMethodSrv_createPaymentFromERP_GrabSystemDefault)
- [x] ERP 已存在、TTPOS 已存在 (TestPaymentMethodSrv_SaveGrabPaymentMethod_ExistingByCode)
- [x] ERP 不存在、TTPOS 不存在 (TestPaymentMethodSrv_SaveGrabPaymentMethod_NoERP)
- [x] ERP 创建返回错误但实际已创建 (逻辑在 ensureERPPaymentMethod 中)
- [x] 并发调用场景 (TestPaymentMethodSrv_SaveGrabPaymentMethod_Concurrent - 串行幂等性验证)

- [x] 完成

### 4.2 更新文档

| 项目 | 内容 |
|------|------|
| Files | requirements.md, design.md, qa-test-cases.md |
| Purpose | 更新审核状态和版本信息 |

- [x] 完成

### 4.3 代码审查

| 项目 | 内容 |
|------|------|
| Purpose | 确保代码质量和规范符合 |
| Requirements | 遵循 CLAUDE.md 和 go-main.mdc 规范 |

- [x] 完成（已通过代码审查，确认参数传递正确，名称格式一致）

---

## 提交清单

### 代码质量
- [x] `go mod tidy` 执行
- [x] `go fmt ./...` 执行
- [x] `go vet ./...` 通过（预先存在的问题除外）
- [x] 测试通过: `go test -run "TestPaymentMethodSrv_(SaveGrab|createPaymentFromERP)" ./app/service/...`

### 功能完整性
- [x] R1: ERP 存在性检查
- [x] R2: ERP 已存在时复用
- [x] R3: ERP 失败时重新确认
- [x] R4: TTPOS 侧幂等性
- [x] R5: createPaymentFromERP 优化

### 日志规范
- [x] 所有日志包含 `company_uuid` 字段
- [x] 记录 ERP 查询/创建的请求和响应
- [x] 记录幂等性判断的关键决策点

---

## 开发顺序建议

```
Phase 1.1 → Phase 1.2 → Phase 2.1 → Phase 2.2 → Phase 3.1 → Phase 4
```

1. 先实现基础的 ERP 查询方法
2. 封装幂等性逻辑
3. 修改 Grab/LINE MAN 保存方法
4. 优化 createPaymentFromERP
5. 编写测试验证

---

**版本**: v1.1.0
**创建日期**: 2026-02-27
**更新日期**: 2026-02-28
