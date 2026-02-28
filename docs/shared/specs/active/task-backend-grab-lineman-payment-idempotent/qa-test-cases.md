# Grab/LINE MAN 支付方式幂等性优化 - QA 测试用例

## 📋 基本信息

| 项目 | 内容 |
|------|------|
| **Spec ID** | task-backend-grab-lineman-payment-idempotent |
| **测试范围** | `SaveGrabPaymentMethod`、`SaveLineManPaymentMethod`、`createPaymentFromERP` 幂等性 |
| **创建日期** | 2026-02-28 |
| **测试人员** | - |

---

## 🎯 测试目标

1. 验证 `SaveGrabPaymentMethod` 和 `SaveLineManPaymentMethod` 的幂等性（ERP 已存在时复用）
2. 验证从 ERP 同步支付方式时，能够正确识别 Grab/LINE MAN 系统默认支付方式

---

## 📝 前置条件

1. 商户已开启 ERP 功能（`is_enable_erp = 1`）
2. 商户已配置 `erpnext_company_abbr`（如 "TEST"）
3. ERP 中已创建对应的支付方式

---

## 🧪 测试用例

### 模块一：SaveGrabPaymentMethod / SaveLineManPaymentMethod 幂等性

---

### TC-001: SaveGrabPaymentMethod - 首次创建（ERP 不存在）

**优先级**: P0

**前置条件**:
- 商户已开启 ERP 功能
- TTPOS 中不存在 Grab 支付方式
- ERP 中不存在 Grab 支付方式

**测试步骤**:
1. 调用开启 Grab 外卖接口（会触发 `SaveGrabPaymentMethod`）
2. 查询 TTPOS 中 Grab 支付方式
3. 查询 ERP 中 Grab 支付方式

**预期结果**:
- TTPOS 创建 Grab 支付方式：
  - `code` = `91100`
  - `source` = `0`（系统默认）
  - `erpnext_payment` = `Grab-0000 - {company_abbr}`
  - `erpnext_payment_id` = ERP 返回的 PaymentID
- ERP 创建对应支付方式

---

### TC-002: SaveGrabPaymentMethod - ERP 已存在时复用

**优先级**: P0

**前置条件**:
- 商户已开启 ERP 功能
- TTPOS 中不存在 Grab 支付方式
- **ERP 中已存在** `Grab-0000 - {company_abbr}` 支付方式

**测试步骤**:
1. 调用开启 Grab 外卖接口
2. 查询 TTPOS 中 Grab 支付方式
3. 查询 ERP 中 Grab 支付方式数量

**预期结果**:
- TTPOS 创建 Grab 支付方式，`erpnext_payment` 和 `erpnext_payment_id` 关联到 ERP 已有数据
- **ERP 不重复创建**，仍然只有 1 条 Grab 支付方式
- 日志显示 `ensureERPPaymentMethod-AlreadyExists`

---

### TC-003: SaveGrabPaymentMethod - TTPOS 已存在时跳过

**优先级**: P0

**前置条件**:
- TTPOS 中已存在 Grab 支付方式（`code = 91100`）

**测试步骤**:
1. 调用开启 Grab 外卖接口
2. 查询 TTPOS 中 Grab 支付方式数量

**预期结果**:
- **不重复创建**，TTPOS 中仍然只有 1 条 Grab 支付方式
- 方法正常返回，无错误

---

### TC-004: SaveGrabPaymentMethod - ERP 创建失败但实际已创建

**优先级**: P1

**前置条件**:
- 商户已开启 ERP 功能
- TTPOS 中不存在 Grab 支付方式
- 模拟 ERP 创建接口返回超时错误，但实际已创建成功

**测试步骤**:
1. 调用开启 Grab 外卖接口
2. 观察日志和最终结果

**预期结果**:
- 系统重新查询 ERP 确认支付方式已创建
- TTPOS 正常创建 Grab 支付方式并关联 ERP 数据
- 日志显示 `ensureERPPaymentMethod-CreateFailed-ButExists`

---

### TC-005: SaveLineManPaymentMethod - 首次创建（ERP 不存在）

**优先级**: P0

**前置条件**:
- 商户已开启 ERP 功能
- TTPOS 中不存在 LINE MAN 支付方式
- ERP 中不存在 LINE MAN 支付方式

**测试步骤**:
1. 调用开启 LINE MAN 外卖接口
2. 查询 TTPOS 中 LINE MAN 支付方式

**预期结果**:
- TTPOS 创建 LINE MAN 支付方式：
  - `code` = `91200`
  - `source` = `0`
  - `erpnext_payment` = `LINE MAN-0000 - {company_abbr}`
- ERP 创建对应支付方式

---

### TC-006: SaveLineManPaymentMethod - ERP 已存在时复用

**优先级**: P0

**前置条件**:
- 商户已开启 ERP 功能
- TTPOS 中不存在 LINE MAN 支付方式
- **ERP 中已存在** `LINE MAN-0000 - {company_abbr}` 支付方式

**测试步骤**:
1. 调用开启 LINE MAN 外卖接口
2. 查询 ERP 中 LINE MAN 支付方式数量

**预期结果**:
- **ERP 不重复创建**
- TTPOS 创建并关联到 ERP 已有数据

---

### TC-007: SaveLineManPaymentMethod - TTPOS 已存在时跳过

**优先级**: P0

**前置条件**:
- TTPOS 中已存在 LINE MAN 支付方式（`code = 91200`）

**测试步骤**:
1. 调用开启 LINE MAN 外卖接口
2. 查询 TTPOS 中 LINE MAN 支付方式数量

**预期结果**:
- **不重复创建**，仍然只有 1 条

---

### TC-008: SaveGrabPaymentMethod - 未开启 ERP 时仅创建 TTPOS

**优先级**: P1

**前置条件**:
- 商户**未开启** ERP 功能（`is_enable_erp = 0`）
- TTPOS 中不存在 Grab 支付方式

**测试步骤**:
1. 调用开启 Grab 外卖接口
2. 查询 TTPOS 中 Grab 支付方式

**预期结果**:
- TTPOS 创建 Grab 支付方式
- `erpnext_payment` 和 `erpnext_payment_id` 为空
- 不调用 ERP 接口

---

### TC-009: 并发调用 SaveGrabPaymentMethod

**优先级**: P1

**前置条件**:
- TTPOS 中不存在 Grab 支付方式

**测试步骤**:
1. 同时发起 2 个开启 Grab 外卖请求

**预期结果**:
- TTPOS 中只有 1 条 Grab 支付方式（事务保证）
- 两个请求都正常返回，无报错

---

### 模块二：createPaymentFromERP（ERP 同步到 TTPOS）

---

### TC-010: Grab 系统默认支付方式创建

**优先级**: P0

**前置条件**:
- ERP 中存在支付方式，名称为 `Grab-0000 - {company_abbr}`（如 `Grab-0000 - TEST`）
- TTPOS 中不存在 Grab 支付方式

**测试步骤**:
1. 触发 ERP 支付方式同步（调用 `syncFromERP`）
2. 查询 TTPOS 中新创建的支付方式

**预期结果**:
| 字段 | 预期值 |
|------|--------|
| `name` | `Grab` |
| `payment_name` | `Grab` |
| `code` | `91100` |
| `source` | `0`（系统默认） |
| `is_show_cashier` | `0` |
| `is_show_assistant` | `0` |
| `is_show_member_recharge` | `0` |
| `is_show_kiosk` | `0` |
| `default_img` | 空字符串 |
| `erpnext_payment` | `Grab-0000 - TEST`（原始 ERP 名称） |
| `erpnext_payment_id` | ERP 返回的 PaymentID |

---

### TC-011: LINE MAN 系统默认支付方式创建（syncFromERP）

**优先级**: P0

**前置条件**:
- ERP 中存在支付方式，名称为 `LINE MAN-0000 - {company_abbr}`（如 `LINE MAN-0000 - TEST`）
- TTPOS 中不存在 LINE MAN 支付方式

**测试步骤**:
1. 触发 ERP 支付方式同步
2. 查询 TTPOS 中新创建的支付方式

**预期结果**:
| 字段 | 预期值 |
|------|--------|
| `name` | `LINE MAN` |
| `payment_name` | `LINE MAN` |
| `code` | `91200` |
| `source` | `0`（系统默认） |
| `is_show_cashier` | `0` |
| `is_show_assistant` | `0` |
| `is_show_member_recharge` | `0` |
| `default_img` | 空字符串 |
| `erpnext_payment` | `LINE MAN-0000 - TEST` |

---

### TC-012: 普通支付方式创建（非 Grab/LINE MAN）

**优先级**: P0

**前置条件**:
- ERP 中存在普通支付方式，名称为 `Cash-0001 - TEST`
- TTPOS 中不存在该支付方式

**测试步骤**:
1. 触发 ERP 支付方式同步
2. 查询 TTPOS 中新创建的支付方式

**预期结果**:
| 字段 | 预期值 |
|------|--------|
| `name` | `Cash-0001 - TEST`（使用 ERP 原始名称） |
| `payment_name` | `Cash-0001 - TEST` |
| `code` | `>= 20000`（自动生成） |
| `source` | `1`（手动添加） |
| `is_show_cashier` | `1` |
| `is_show_assistant` | `1` |
| `is_show_member_recharge` | `1` |
| `default_img` | `/image/pay/ja_pay.png` |

---

### TC-013: Grab 序号不匹配（非 0000）

**优先级**: P1

**前置条件**:
- ERP 中存在支付方式，名称为 `Grab-0001 - TEST`（序号为 0001，非 0000）

**测试步骤**:
1. 触发 ERP 支付方式同步
2. 查询 TTPOS 中新创建的支付方式

**预期结果**:
- 应按**普通支付方式**处理
- `name` = `Grab-0001 - TEST`
- `source` = `1`（手动添加）
- `code` >= 20000（自动生成）
- `is_show_cashier` = `1`

---

### TC-014: 公司缩写不匹配

**优先级**: P1

**前置条件**:
- 当前商户 `erpnext_company_abbr` = `TEST`
- ERP 中存在支付方式，名称为 `Grab-0000 - OTHER`（公司缩写不匹配）

**测试步骤**:
1. 触发 ERP 支付方式同步
2. 查询 TTPOS 中新创建的支付方式

**预期结果**:
- 应按**普通支付方式**处理
- `name` = `Grab-0000 - OTHER`
- `source` = `1`（手动添加）
- `code` >= 20000

---

### TC-015: 类似前缀但名称不同

**优先级**: P1

**前置条件**:
- ERP 中存在支付方式，名称为 `GrabFood-0000 - TEST`（前缀类似但不是 "Grab"）

**测试步骤**:
1. 触发 ERP 支付方式同步
2. 查询 TTPOS 中新创建的支付方式

**预期结果**:
- 应按**普通支付方式**处理
- `name` = `GrabFood-0000 - TEST`
- `source` = `1`

---

### TC-016: Grab 禁用状态同步

**优先级**: P1

**前置条件**:
- ERP 中存在支付方式 `Grab-0000 - TEST`，`enabled = false`

**测试步骤**:
1. 触发 ERP 支付方式同步
2. 查询 TTPOS 中新创建的支付方式

**预期结果**:
- `status` = `0`（禁用）
- 其他字段仍符合 Grab 系统默认配置（source=0, code=91100 等）

---

### TC-017: syncFromERP 已存在时不重复创建

**优先级**: P0

**前置条件**:
- TTPOS 中已存在 Grab 支付方式（通过 `erpnext_payment_id` 匹配）
- ERP 返回相同的 Grab 支付方式数据

**测试步骤**:
1. 触发 ERP 支付方式同步
2. 查询 TTPOS 中 Grab 支付方式数量

**预期结果**:
- 不创建新记录，只有 1 条 Grab 支付方式
- 可能更新 `status` 字段（根据 ERP 返回的 enabled 状态）

---

### TC-018: 公司缩写为空时的处理

**优先级**: P2

**前置条件**:
- 商户 `erpnext_company_abbr` 为空字符串
- ERP 中存在支付方式 `Grab-0000 - `（末尾为空）

**测试步骤**:
1. 触发 ERP 支付方式同步

**预期结果**:
- 如果 ERP 名称精确匹配 `Grab-0000 - `，则按系统默认处理
- 实际场景中 ERP 不太可能返回这种格式，应按普通支付方式处理

---

### TC-019: 同时同步 Grab 和 LINE MAN

**优先级**: P1

**前置条件**:
- ERP 中同时存在：
  - `Grab-0000 - TEST`
  - `LINE MAN-0000 - TEST`
  - `Cash-0001 - TEST`

**测试步骤**:
1. 触发 ERP 支付方式同步
2. 查询 TTPOS 中所有新创建的支付方式

**预期结果**:
| 支付方式 | code | source | is_show_cashier |
|----------|------|--------|-----------------|
| Grab | 91100 | 0 | 0 |
| LINE MAN | 91200 | 0 | 0 |
| Cash-0001 - TEST | >= 20000 | 1 | 1 |

---

## 📊 测试矩阵

### 模块一：SaveGrabPaymentMethod / SaveLineManPaymentMethod

| 测试场景 | 用例 | P0 | P1 |
|----------|------|----|----|
| SaveGrab - 首次创建（ERP 不存在） | TC-001 | ✅ | |
| SaveGrab - ERP 已存在时复用 | TC-002 | ✅ | |
| SaveGrab - TTPOS 已存在时跳过 | TC-003 | ✅ | |
| SaveGrab - ERP 创建失败但实际已创建 | TC-004 | | ✅ |
| SaveLineMan - 首次创建 | TC-005 | ✅ | |
| SaveLineMan - ERP 已存在时复用 | TC-006 | ✅ | |
| SaveLineMan - TTPOS 已存在时跳过 | TC-007 | ✅ | |
| SaveGrab - 未开启 ERP | TC-008 | | ✅ |
| SaveGrab - 并发调用 | TC-009 | | ✅ |

### 模块二：createPaymentFromERP (syncFromERP)

| 测试场景 | 用例 | P0 | P1 | P2 |
|----------|------|----|----|-----|
| Grab 系统默认创建 | TC-010 | ✅ | | |
| LINE MAN 系统默认创建 | TC-011 | ✅ | | |
| 普通支付方式创建 | TC-012 | ✅ | | |
| 序号不匹配 | TC-013 | | ✅ | |
| 公司缩写不匹配 | TC-014 | | ✅ | |
| 前缀类似不匹配 | TC-015 | | ✅ | |
| 禁用状态同步 | TC-016 | | ✅ | |
| 已存在时不重复创建 | TC-017 | ✅ | | |
| 公司缩写为空 | TC-018 | | | ✅ |
| 同时同步多个支付方式 | TC-019 | | ✅ | |

---

## 🔧 测试数据准备

### ERP 侧测试数据

```json
[
  {
    "name": "Grab-0000 - TEST",
    "payment_id": "PID001",
    "enabled": true
  },
  {
    "name": "LINE MAN-0000 - TEST",
    "payment_id": "PID002",
    "enabled": true
  },
  {
    "name": "Cash-0001 - TEST",
    "payment_id": "PID003",
    "enabled": true
  },
  {
    "name": "Grab-0001 - TEST",
    "payment_id": "PID004",
    "enabled": true
  }
]
```

### TTPOS 侧验证 SQL

```sql
-- 查询支付方式详情
SELECT
    uuid,
    name,
    payment_name,
    code,
    source,
    is_show_cashier,
    is_show_assistant,
    is_show_member_recharge,
    default_img,
    erpnext_payment,
    erpnext_payment_id,
    status
FROM ttpos_payment_method
WHERE delete_time = 0
ORDER BY code;
```

---

## ✅ 验收标准

1. 所有 P0 测试用例通过（11 个）
2. 所有 P1 测试用例通过（7 个）
3. ERP 侧无重复支付方式创建
4. TTPOS 侧无重复支付方式创建
5. 日志中包含 `company_uuid` 字段

---

## 📋 测试用例总览

| 分类 | P0 | P1 | P2 | 总计 |
|------|----|----|-----|------|
| SaveGrabPaymentMethod | 3 | 2 | 0 | 5 |
| SaveLineManPaymentMethod | 3 | 0 | 0 | 3 |
| createPaymentFromERP | 4 | 5 | 1 | 10 |
| **总计** | **10** | **7** | **1** | **18** |

---

**版本**: v1.0.0
**创建日期**: 2026-02-28
