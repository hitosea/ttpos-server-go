# 打印模块开发详细指南

> Main 模块打印功能的完整开发文档

---

## 📚 目录

1. [打印模板类型](#打印模板类型)
2. [打印机类型](#打印机类型)
3. [打印单据类型](#打印单据类型)
4. [新增字段完整流程](#新增字段完整流程)
5. [代码示例](#代码示例)
6. [文件结构说明](#文件结构说明)
7. [技术细节](#技术细节)
8. [常见陷阱](#常见陷阱)
9. [排查指南](#排查指南)

---

## 打印模板类型

### 标准模板
- **特点**: 硬编码的 Go 代码
- **位置**: `main/app/printer/template/`
- **优点**: 性能高，完全可控
- **缺点**: 需要重新编译部署

### 自定义模板
- **特点**: JSON 配置驱动
- **位置**: `main/app/printer/pkg/template/`
- **优点**: 用户可在后台自定义
- **缺点**: 功能相对受限

---

## 打印机类型

| 类型 | 文件后缀 | 示例文件 | 说明 |
|------|---------|----------|------|
| 通用图像 | `_img.go` | `statement_order_img.go` | 标准图像模板，适用于大多数打印机 |
| 58mm 图像 | `_img_58mm.go` | `statement_order_img_58mm.go` | 58mm 纸张宽度的图像模板 |
| Compax | `_compax.go` | `statement_order_compax.go` | Compax 品牌打印机专用 |
| Codesoft | `_codesoft.go` | `statement_order_codesoft.go` | Codesoft 品牌打印机专用 |
| XPrinter | `_xprinter.go` | `statement_order_xprinter.go` | XPrinter 品牌打印机专用 |
| Sunmi | `_sunmi.go` | `statement_order_sunmi.go` | Sunmi（商米）品牌打印机专用 |
| 自定义 | `_img_custom.go` | `statement_order_img_custom.go` | JSON 配置驱动的自定义模板 |

---

## 打印单据类型

| 单据类型 | 文件前缀 | 说明 | 使用场景 |
|----------|----------|------|----------|
| 结账单 | `statement_order_` | 正式收银小票 | 顾客结账后打印 |
| 预结账单 | `statement_pre_` | 预结账小票 | 顾客查看账单但未支付 |
| 发票 | `invoice_` | 发票模板 | 正式发票打印 |
| 菜品单 | `dishes_` | 厨房出单 | 厨房制作菜品 |
| 交班单 | `handover_` | 收银交接班 | 班次交接汇总 |
| 充值单 | `recharge_` | 会员充值 | 会员充值记录 |
| 外卖单 | `takeout_order_` | 外卖订单 | 外卖订单打印 |
| 业务数据 | `business_data_` | 经营报表 | 经营数据汇总 |

---

## 新增字段完整流程

### 概览：7 类文件同步修改

当需要在打印模板中新增一个字段（如"活动抵扣"）时，必须同步修改以下文件：

```
数据结构层（1 个文件）
├── printer/template_struct/order.go

标准模板层（6 个文件）
├── printer/template/statement_order_img.go
├── printer/template/statement_order_img_58mm.go
├── printer/template/statement_order_compax.go
├── printer/template/statement_order_codesoft.go
├── printer/template/statement_order_xprinter.go
└── printer/template/statement_order_sunmi.go

自定义模板层（5 个文件）
├── printer/pkg/template/statement_order_tmp.json
├── printer/pkg/template/statement_order_data.json
├── printer/pkg/template/statement_pre_tmp.json（如适用）
├── printer/pkg/template/statement_pre_data.json（如适用）
└── printer/template/statement_order_img_custom.go
```

---

## 代码示例

### 第一步：更新数据结构

**文件**: `printer/template_struct/order.go`

```go
type StatementOrderInfoData struct {
    // ... 其他字段
    CouponExchangeAmount float64 `json:"coupon_exchange_amount"` // 优惠券抵扣金额
    ActivityAmount       float64 `json:"activity_amount"`        // 活动抵扣金额（新增）
    CheckOutZeroFee      float64 `json:"check_out_zero_fee"`     // 手动抹零金额
}
```

### 第二步：更新标准模板（6 个文件）

**示例文件**: `printer/template/statement_order_img.go`

```go
// 优惠券抵扣
if saleOrder.CouponExchangeAmount > 0 {
    img.AppendText(fmt.Sprintf("%s: %s", 
        t.base.Translate("优惠券抵扣"), 
        t.base.GetPriceAndUnit(saleOrder.CouponExchangeAmount)))
    img.LineFeed(1)
}

// 活动抵扣（新增）
if saleOrder.ActivityAmount > 0 {
    img.AppendText(fmt.Sprintf("%s: %s", 
        t.base.Translate("活动抵扣"), 
        t.base.GetPriceAndUnit(saleOrder.ActivityAmount)))
    img.LineFeed(1)
}

// 手动抹零
if saleOrder.CheckOutZeroFee > 0 {
    img.AppendText(fmt.Sprintf("%s: %s", 
        t.base.Translate("手动抹零"), 
        t.base.GetPriceAndUnit(saleOrder.CheckOutZeroFee)))
    img.LineFeed(1)
}
```

**关键点**:
- ✅ 在相同位置插入（优惠券抵扣之后，手动抹零之前）
- ✅ 使用 `if amount > 0` 判断
- ✅ 使用 `Translate()` 支持多语言
- ✅ 使用 `GetPriceAndUnit()` 格式化金额

**需要修改的 6 个文件**:
1. `statement_order_img.go` - 通用图像
2. `statement_order_img_58mm.go` - 58mm
3. `statement_order_compax.go` - Compax
4. `statement_order_codesoft.go` - Codesoft
5. `statement_order_xprinter.go` - XPrinter
6. `statement_order_sunmi.go` - Sunmi

### 第三步：更新自定义模板配置

#### 3.1 默认模板 JSON 配置

**文件**: `printer/pkg/template/statement_order_tmp.json`

在 `coupon_exchange_amount` 配置块之后插入：

```json
[
    {
        "group_block_id": "amount_due_information",
        "block_id": "order.activity_amount",
        "block_type": "label:auto:value",
        "block_label": {
            "zh": "活动抵扣",
            "en": "Activity Deduction",
            "th": "หักลดจากกิจกรรม",
            "ja": "活動控除",
            "ko": "활동 공제",
            "my": "လှုပ်ရှားမှုလျှော့",
            "tr": "Etkinlik indir",
            "de": "Aktivitätsabzug",
            "sv": "Aktivitetavdrag",
            "zhtw": "活動抵扣"
        },
        "block_name": {
            "en": "Activity Deduction",
            "ja": "活動控除",
            "ko": "활동 공제",
            "my": "လှုပ်ရှားမှုလျှော့",
            "th": "หักลดจากกิจกรรม",
            "tr": "Etkinlik Düşümü",
            "zh": "活动抵扣",
            "zhtw": "活動抵扣",
            "de": "Aktivität Abzug",
            "sv": "Aktivitet Avdrag"
        },
        "block_attr": {
            "font_size": 20,
            "align": "right",
            "font_bold": false,
            "dividing_line": false,
            "leading_blank_lines": 0,
            "trailing_blank_lines": 0,
            "width": 60,
            "show_currency_unit": true
        },
        "conditions": [
            {
                "field": "order.activity_amount",
                "operator": "not_empty",
                "value": 0
            }
        ]
    }
]
```

**同样需要修改**: `statement_pre_tmp.json`（预结账单）

#### 3.2 示例数据

**文件**: `printer/pkg/template/statement_order_data.json`

```json
{
    "order": {
        "member_points_discount": 100,
        "coupon_exchange_amount": 10,
        "activity_amount": 50,
        "check_out_zero_fee": "0.92"
    }
}
```

**同样需要修改**: `statement_pre_data.json`（预结账单）

#### 3.3 自定义模板数据填充

**文件**: `printer/template/statement_order_img_custom.go`

在 `GetPrintContent()` 方法中：

```go
StatementOrderInfoData: StatementOrderInfoData{
    // ... 其他字段
    CouponExchangeAmount: saleOrder.CouponExchangeAmount,
    ActivityAmount:       saleOrder.ActivityAmount,  // 新增
    CheckOutZeroFee:      saleOrder.CheckOutZeroFee,
}
```

---

## 文件结构说明

### 标准模板目录

```
main/app/printer/template/
├── base.go                          # 模板基类（翻译、格式化等）
│
├── statement_order_*.go             # 结账单模板（6 个文件）
│   ├── statement_order_img.go           # 通用图像
│   ├── statement_order_img_58mm.go      # 58mm
│   ├── statement_order_compax.go        # Compax
│   ├── statement_order_codesoft.go      # Codesoft
│   ├── statement_order_xprinter.go      # XPrinter
│   ├── statement_order_sunmi.go         # Sunmi
│   └── statement_order_img_custom.go    # 自定义模板
│
├── invoice_*.go                     # 发票模板（6 个文件）
├── dishes_*.go                      # 菜品单模板（4 个文件）
├── handover_*.go                    # 交班单模板（6 个文件）
├── recharge_*.go                    # 充值单模板（5 个文件）
├── takeout_order_*.go               # 外卖单模板（5 个文件）
└── business_data_*.go               # 业务数据模板（4 个文件）
```

### 数据结构目录

```
main/app/printer/template_struct/
└── order.go
    ├── StatementOrderData           # 结账单完整数据
    ├── StatementOrderInfoData       # 订单信息数据
    ├── StatementOrderStoreData      # 店铺信息数据
    └── StatementOrderProductData    # 商品信息数据
```

### 自定义模板配置目录

```
main/app/printer/pkg/template/
├── statement_order_tmp.json         # 结账单默认模板配置
├── statement_order_data.json        # 结账单示例数据
├── statement_order_config.json      # 结账单配置元数据
├── statement_pre_tmp.json           # 预结账单默认模板配置
├── statement_pre_data.json          # 预结账单示例数据
├── statement_pre_config.json        # 预结账单配置元数据
└── README_STATEMENT_ORDER.md        # 模板使用文档
```

---

## 技术细节

### 金额处理

```go
// ✅ 正确：使用 float64 存储
type OrderInfo struct {
    ActivityAmount float64 `json:"activity_amount"`
}

// ✅ 正确：格式化金额
formattedAmount := t.base.GetPriceAndUnit(saleOrder.ActivityAmount)
// 输出: "¥50.00" 或 "$50.00"

// ✅ 正确：判断条件
if saleOrder.ActivityAmount > 0 {
    // 只显示正数
}

// ❌ 错误：使用 != 0 会显示负数
if saleOrder.ActivityAmount != 0 {  // 避免这样做
    // 负数也会显示
}
```

### 多语言支持

**支持的语言（9 种）**:

| 代码 | 语言 | 本地名称 |
|------|------|----------|
| `zh` | 中文简体 | 简体中文 |
| `zhtw` | 中文繁体 | 繁體中文 |
| `en` | 英文 | English |
| `th` | 泰文 | ภาษาไทย |
| `ja` | 日文 | 日本語 |
| `ko` | 韩文 | 한국어 |
| `my` | 缅甸文 | မြန်မာဘာသာ |
| `tr` | 土耳其文 | Türkçe |
| `sv` | 瑞典文 | Svenska |

**标准模板使用**:

```go
// 自动根据系统语言翻译
t.base.Translate("活动抵扣")
```

**JSON 模板配置**:

```json
"block_label": {
    "zh": "活动抵扣",
    "en": "Activity Deduction",
    "th": "หักลดจากกิจกรรม",
    "ja": "活動控除",
    "ko": "활동 공제",
    "my": "လှုပ်ရှားမှုလျှော့",
    "tr": "Etkinlik indir",
    "de": "Aktivitätsabzug",
    "sv": "Aktivitetavdrag",
    "zhtw": "活動抵扣"
}
```

### 条件显示控制

**标准模板**:

```go
// 大于 0 显示
if value > 0 {
    img.AppendText(...)
}

// 不为空显示
if value != nil && value != "" {
    img.AppendText(...)
}
```

**JSON 模板**:

```json
"conditions": [
    {
        "field": "order.activity_amount",
        "operator": "not_empty",  // 或 "gt" 大于
        "value": 0
    }
]
```

**可用的条件操作符**:

| 操作符 | 说明 | 适用类型 | 示例 |
|--------|------|----------|------|
| `eq` | 等于 | 所有 | `"operator": "eq", "value": "active"` |
| `ne` | 不等于 | 所有 | `"operator": "ne", "value": "inactive"` |
| `gt` | 大于 | 数字 | `"operator": "gt", "value": 0` |
| `gte` | 大于等于 | 数字 | `"operator": "gte", "value": 100` |
| `lt` | 小于 | 数字 | `"operator": "lt", "value": 1000` |
| `lte` | 小于等于 | 数字 | `"operator": "lte", "value": 500` |
| `empty` | 为空 | 所有 | `"operator": "empty", "value": 0` |
| `not_empty` | 不为空 | 所有 | `"operator": "not_empty", "value": 0` |
| `contains` | 包含 | 字符串 | `"operator": "contains", "value": "test"` |
| `not_contains` | 不包含 | 字符串 | `"operator": "not_contains", "value": "test"` |
| `in` | 在列表中 | 所有 | `"operator": "in", "value": ["active", "pending"]` |
| `not_in` | 不在列表中 | 所有 | `"operator": "not_in", "value": ["inactive"]` |

---

## 常见陷阱

### 1. 字段位置不一致

❌ **错误示例**:

```go
// statement_order_img.go
优惠券抵扣 → 活动抵扣 → 手动抹零

// statement_order_sunmi.go
优惠券抵扣 → 手动抹零 → 活动抵扣  // ❌ 顺序不一致
```

✅ **正确做法**:

```go
// 所有 6 个标准模板都使用相同顺序
优惠券抵扣 → 活动抵扣 → 手动抹零
```

### 2. 多语言翻译不完整

❌ **错误示例**:

```json
"block_label": {
    "zh": "活动抵扣",
    "en": "Activity Deduction"
    // ❌ 缺少其他 7 种语言
}
```

✅ **正确做法**:

```json
"block_label": {
    "zh": "活动抵扣",
    "en": "Activity Deduction",
    "th": "หักลดจากกิจกรรม",
    "ja": "活動控除",
    "ko": "활动 공제",
    "my": "လှုပ်ရှားမှုလျှော့",
    "tr": "Etkinlik indir",
    "de": "Aktivitätsabzug",
    "sv": "Aktivitetavdrag",
    "zhtw": "活動抵扣"
}
```

### 3. JSON 格式错误

❌ **错误做法**:

```bash
# 直接提交，可能导致运行时错误
git add statement_order_tmp.json
git commit -m "添加字段"
```

✅ **正确做法**:

```bash
# 使用 jq 验证 JSON 格式
jq empty main/app/printer/pkg/template/statement_order_tmp.json
jq empty main/app/printer/pkg/template/statement_pre_tmp.json

# 如果没有输出，说明格式正确
```

### 4. 数据结构未同步

❌ **错误示例**:

```go
// printer/template_struct/order.go - ✅ 已添加
ActivityAmount float64 `json:"activity_amount"`

// printer/template/statement_order_img_custom.go - ❌ 忘记填充
StatementOrderInfoData: StatementOrderInfoData{
    CouponExchangeAmount: saleOrder.CouponExchangeAmount,
    // ActivityAmount: saleOrder.ActivityAmount,  // ❌ 忘记添加
}
```

✅ **正确做法**:

```go
// printer/template/statement_order_img_custom.go
StatementOrderInfoData: StatementOrderInfoData{
    CouponExchangeAmount: saleOrder.CouponExchangeAmount,
    ActivityAmount:       saleOrder.ActivityAmount,  // ✅ 已添加
}
```

### 5. 条件判断边界问题

❌ **错误示例**:

```go
if saleOrder.ActivityAmount != 0 {  // ❌ 负数也会显示
    img.AppendText(...)
}
```

✅ **正确做法**:

```go
if saleOrder.ActivityAmount > 0 {  // ✅ 只显示正数
    img.AppendText(...)
}
```

---

## 排查指南

### 问题：字段不显示

**排查步骤**:

1. **检查数据源**:
```go
fmt.Printf("ActivityAmount: %v\n", saleOrder.ActivityAmount)
```

2. **检查条件判断**:
```go
if saleOrder.ActivityAmount > 0 {  // 确保条件正确
    // ...
}
```

3. **检查自定义模板数据填充**:
```go
// 确保在 statement_order_img_custom.go 中已添加
ActivityAmount: saleOrder.ActivityAmount,
```

4. **检查 JSON 配置**:
```bash
# 查看 JSON 中是否包含该字段
jq '.rows[] | select(.block_id == "order.activity_amount")' \
    main/app/printer/pkg/template/statement_order_tmp.json
```

### 问题：翻译不生效

**排查步骤**:

1. **检查多语言文件**:
```bash
cat main/i18n/languages/zh.json | grep "活动抵扣"
cat main/i18n/languages/en.json | grep "Activity Deduction"
```

2. **检查 JSON 配置**:
```json
// 确保包含所有 9 种语言
"block_label": {
    "zh": "活动抵扣",
    "en": "Activity Deduction",
    // ... 其他语言
}
```

3. **重启服务**:
```bash
# 多语言文件修改后需要重启
systemctl restart ttpos-main
```

### 问题：JSON 模板解析失败

**排查步骤**:

1. **验证 JSON 格式**:
```bash
jq empty main/app/printer/pkg/template/statement_order_tmp.json
```

2. **查看具体错误**:
```bash
# 使用 jq 格式化查看
jq . main/app/printer/pkg/template/statement_order_tmp.json
```

3. **检查常见语法错误**:
- 缺少逗号
- 多余的逗号
- 引号不匹配
- 括号不匹配

4. **查看运行时日志**:
```bash
tail -f /path/to/log/file.log | grep "template"
```

### 问题：不同打印机显示不一致

**排查步骤**:

1. **检查 6 个标准模板**:
```bash
# 检查字段在所有模板中的位置是否一致
grep -n "活动抵扣" main/app/printer/template/statement_order_*.go
```

2. **比对代码逻辑**:
```bash
# 确保条件判断一致
grep -A 3 "ActivityAmount" main/app/printer/template/statement_order_*.go
```

3. **测试每个打印机**:
- 在测试环境逐个测试不同打印机类型
- 比对打印结果

---

## 验证清单

完成修改后，请逐项检查：

### 数据结构层 ✓
- [ ] `printer/template_struct/order.go` - 添加字段定义
- [ ] 字段添加 `json` tag
- [ ] 字段添加中文注释

### 标准模板层（6 个文件）✓
- [ ] `statement_order_img.go` - 通用图像模板
- [ ] `statement_order_img_58mm.go` - 58mm 模板
- [ ] `statement_order_compax.go` - Compax 模板
- [ ] `statement_order_codesoft.go` - Codesoft 模板
- [ ] `statement_order_xprinter.go` - XPrinter 模板
- [ ] `statement_order_sunmi.go` - Sunmi 模板
- [ ] 所有模板字段顺序一致
- [ ] 添加条件判断（`> 0`）
- [ ] 使用 `Translate()` 支持多语言
- [ ] 使用 `GetPriceAndUnit()` 格式化金额

### 自定义模板层 ✓
- [ ] `statement_order_tmp.json` - 添加配置块
- [ ] `statement_pre_tmp.json` - 添加配置块（如适用）
- [ ] `statement_order_data.json` - 添加示例数据
- [ ] `statement_pre_data.json` - 添加示例数据（如适用）
- [ ] `statement_order_img_custom.go` - 填充数据
- [ ] JSON 包含所有 9 种语言翻译
- [ ] JSON 格式通过 `jq` 验证

### 代码质量 ✓
- [ ] 运行 `go fmt` 格式化代码
- [ ] 运行 `go vet` 静态检查
- [ ] JSON 文件通过 `jq empty` 验证

### 功能测试 ✓
- [ ] 创建测试订单（金额 > 0）
- [ ] 验证字段正确显示
- [ ] 测试金额 = 0 的情况
- [ ] 测试多语言环境
- [ ] 回归测试其他字段

---

**最后更新**: 2025-11-25  
**维护者**: TTPOS Team

