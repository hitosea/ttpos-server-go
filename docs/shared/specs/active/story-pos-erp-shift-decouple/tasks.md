# story-pos-erp-shift-decouple 任务清单

## 进度总览

| 项目 | 数值 |
|------|------|
| 总 SP | 3 |
| 总任务数 | 7 |
| 已完成 | 6 |
| 完成率 | 86% |

---

## Phase 1: 班次服务改造

### 1.1 添加 ShiftVersion 字段和迁移

| 项目 | 内容 |
|------|------|
| File | `main/app/model/staff.go`, `main/app/constant/staff.go`, `admin/database/migrations/` |
| Purpose | 区分新旧班次版本 |
| Requirements | ShiftVersion=1 旧版, 2 新版（默认） |
| Leverage | 现有 StaffShiftLog 模型 |

- [x] 完成

### 1.2 开班流程移除 OpenPosEntry

| 项目 | 内容 |
|------|------|
| File | `main/app/service/staff_shift.go` |
| Purpose | 开班时不再调用 ERP OpenPosEntry |
| Requirements | ShiftVersion=2 时跳过 ERP 调用 |
| Leverage | 现有开班逻辑 |

- [x] 完成

### 1.3 交班流程移除 ClosePosEntry

| 项目 | 内容 |
|------|------|
| File | `main/app/service/staff_shift.go` |
| Purpose | 交班时不再调用 ERP ClosePosEntry |
| Requirements | ShiftVersion=2 时跳过 ERP 调用，内部完成汇总 |
| Leverage | 现有交班逻辑（needErpClosePos 自动跳过） |

- [x] 完成

---

## Phase 2: 移除支付方式限制

### 2.1 移除结账支付方式班次校验

| 项目 | 内容 |
|------|------|
| File | `main/app/service/staff_shift.go` (ValidatePaymentMethod) |
| Purpose | 结账时允许使用本班新增的支付方式 |
| Requirements | ShiftVersion=2 时 ValidatePaymentMethod 直接返回 true |
| Leverage | order_pay.go 调用者自动受益 |

- [x] 完成

### 2.2 移除充值支付方式班次校验

| 项目 | 内容 |
|------|------|
| File | `main/app/service/staff_shift.go` (ValidatePaymentMethod) |
| Purpose | 充值时允许使用本班新增的支付方式 |
| Requirements | 同 2.1，recharge_order.go 调用者自动受益 |

- [x] 完成

### 2.3 移除接单支付方式班次校验

| 项目 | 内容 |
|------|------|
| File | `main/app/service/staff_shift.go` (ValidatePaymentMethod) |
| Purpose | 手动/自动接单时允许使用本班新增的支付方式 |
| Requirements | 同 2.1，takeout_help.go 调用者自动受益 |

- [x] 完成

---

## Phase 3: 测试与验证

### 3.1 编写单元测试

| 项目 | 内容 |
|------|------|
| File | `main/app/service/staff_shift_test.go` |
| Purpose | 覆盖开班/交班/支付方式校验变更 |
| Requirements | 覆盖率 >= 80% |

- [ ] 完成

---

## 提交清单

### 代码质量
- [x] `go fmt ./...` 执行
- [x] `go vet ./...` 通过
- [x] `go build ./...` 编译通过
- [ ] `go mod tidy` 执行
- [ ] 测试通过: `go test ./...`

### 功能完整性
- [x] 新班次不创建 Opening Entry
- [x] 新班次不创建 Closing Entry（needErpClosePos 自动为 false）
- [x] 本班新增支付方式可立即使用
- [x] 旧班次（ShiftVersion=1）仍走旧流程

### 迁移同步
- [x] 迁移文件已创建（shift_version 字段）
- [x] shop_01.sql 已更新
