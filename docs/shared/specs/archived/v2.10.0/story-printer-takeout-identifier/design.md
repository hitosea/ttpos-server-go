# 打印单添加外卖标识 设计文档

> 本文档定义打印单添加外卖标识功能的技术设计和实现方案。

## 📋 概述

在现有的打印模板中添加外卖标识显示逻辑，当订单为外卖订单时，在打印单标题和桌号/序号前显示外卖标识。涉及所有打印单类型：出菜单、一菜一单、整单打印、退菜单、预结账单、结账单、发票。主要涉及以下打印模板文件：

- **菜品打印模板**: `dishes_img.go`、`dishes_xprinter.go`、`dishes_codesoft.go`
- **订单打印模板**: `statement_order_img.go`、`statement_order_xprinter.go`、`statement_order_codesoft.go`、`statement_order_compax.go`、`statement_order_sunmi.go`
- **发票打印模板**: `invoice_img.go`、`invoice_xprinter.go`、`invoice_codesoft.go`、`invoice_compax.go`、`invoice_sunmi.go`、`invoice_img_58mm.go`

---

## 🎯 规范对齐

### Go Main 规范 (go-main.mdc)

- 保持现有代码结构，不引入新的依赖
- 复用现有的翻译方法 `t.base.Translate()`
- 复用现有的订单判断方法 `order.IsOrderSourceTakeout()`
- 不使用 panic，返回 error

---

## 🔄 代码复用分析

### 可复用的现有组件

- **翻译方法**: `t.base.Translate("外卖")` - 支持多语言翻译
- **订单判断**: `order.IsOrderSourceTakeout()` - 判断是否为外卖订单（订单来源为外卖）
- **打印模板结构**: 现有的 `OutMenuTemplate`、`CompleteOrder`、`OneDishOneOrder` 方法

### 集成点

- **订单模型**: `model.SaleBill` - 提供订单信息
- **打印模板基类**: `printerTemplate` - 提供翻译和格式化方法

---

## 🏗️ 架构设计

### 修改范围

本次修改只涉及打印模板层，不涉及 Service、Repository 或 API 层。

```
打印模板层 (Template)
  ↓ 读取
订单模型 (SaleBill)
  ↓ 使用
翻译服务 (i18n)
```

### 修改文件

#### 菜品打印模板

1. **`main/app/printer/template/dishes_img.go`**
   - `OutMenuTemplate` 方法：在"出菜单"标题后添加外卖标识，在桌号/序号前添加外卖标识
   - `ReturnMenuTemplate` 方法：在"退菜单"标题后添加外卖标识，在桌号/序号前添加外卖标识
   - `CompleteOrder` 方法：在桌号/序号前添加外卖标识
   - `OneDishOneOrder` 方法（如存在）：在桌号/序号前添加外卖标识

2. **`main/app/printer/template/dishes_xprinter.go`**
   - `OutMenuTemplate` 方法：在"出菜单"标题后添加外卖标识，在桌号/序号前添加外卖标识
   - `ReturnMenuTemplate` 方法：在"退菜单"标题后添加外卖标识，在桌号/序号前添加外卖标识
   - `CompleteOrder` 方法（如存在）：在桌号/序号前添加外卖标识
   - `OneDishOneOrder` 方法（如存在）：在桌号/序号前添加外卖标识

3. **`main/app/printer/template/dishes_codesoft.go`**
   - `CompleteOrder` 方法：在桌号/序号前添加外卖标识
   - `OneDishOneOrder` 方法：在桌号/序号前添加外卖标识

#### 订单打印模板（预结账单、结账单、发票）

4. **`main/app/printer/template/statement_order_img.go`**
   - `GetPrintContent` 方法：根据 `printOrderType` 判断类型，在标题后添加外卖标识，在桌号/序号前添加外卖标识

5. **`main/app/printer/template/statement_order_xprinter.go`**
   - `GetPrintContent` 方法：根据 `printType` 判断类型，在标题后添加外卖标识，在桌号/序号前添加外卖标识

6. **`main/app/printer/template/statement_order_codesoft.go`**
   - `GetPrintContent` 方法：根据 `printType` 判断类型，在标题后添加外卖标识，在桌号/序号前添加外卖标识

7. **`main/app/printer/template/statement_order_compax.go`**（如需要）
   - `GetPrintContent` 方法：在标题和桌号/序号前添加外卖标识

8. **`main/app/printer/template/statement_order_sunmi.go`**（如需要）
   - `GetPrintContent` 方法：在标题和桌号/序号前添加外卖标识

#### 发票专用模板

9. **`main/app/printer/template/invoice_img.go`**
   - 在"发票"标题后添加外卖标识，在桌号/序号前添加外卖标识

10. **`main/app/printer/template/invoice_xprinter.go`**
    - 在"发票"标题后添加外卖标识，在桌号/序号前添加外卖标识

11. **`main/app/printer/template/invoice_codesoft.go`**（如需要）
    - 在"发票"标题后添加外卖标识，在桌号/序号前添加外卖标识

12. **`main/app/printer/template/invoice_compax.go`**（如需要）
    - 在"发票"标题后添加外卖标识，在桌号/序号前添加外卖标识

13. **`main/app/printer/template/invoice_sunmi.go`**（如需要）
    - 在"发票"标题后添加外卖标识，在桌号/序号前添加外卖标识

14. **`main/app/printer/template/invoice_img_58mm.go`**
    - 在"发票"标题后添加外卖标识，在桌号/序号前添加外卖标识

#### 交班单模板

15. **`main/app/printer/template/handover_img.go`**
    - `GetPrintContent` 方法：在"交班单"标题后添加外卖标识（当数据包含外卖订单时）

16. **`main/app/printer/template/handover_xprinter.go`**
    - `GetPrintContent` 方法：在"交班单"标题后添加外卖标识（当数据包含外卖订单时）

17. **`main/app/printer/template/handover_codesoft.go`**
    - `GetPrintContent` 方法：在"交班单"标题后添加外卖标识（当数据包含外卖订单时）

18. **`main/app/printer/template/handover_compax.go`**（如需要）
    - `GetPrintContent` 方法：在"交班单"标题后添加外卖标识（当数据包含外卖订单时）

19. **`main/app/printer/template/handover_sunmi.go`**（如需要）
    - `GetPrintContent` 方法：在"交班单"标题后添加外卖标识（当数据包含外卖订单时）

20. **`main/app/printer/template/handover_img_58mm.go`**
    - `GetPrintContent58mm` 方法：在"交班单"标题后添加外卖标识（当数据包含外卖订单时）

#### 营业数据单模板

21. **`main/app/printer/template/business_data_img.go`**
    - `GetPrintContent` 方法：在"营业数据"标题后添加外卖标识（当数据包含外卖订单时）

22. **`main/app/printer/template/business_data_xprinter.go`**
    - `GetPrintContent` 方法：在"营业数据"标题后添加外卖标识（当数据包含外卖订单时）

23. **`main/app/printer/template/business_data_img_58mm.go`**（如需要）
    - `GetPrintContent` 方法：在"营业数据"标题后添加外卖标识（当数据包含外卖订单时）

24. **`main/app/printer/template/business_data_sunmi.go`**（如需要）
    - `GetPrintContent` 方法：在"营业数据"标题后添加外卖标识（当数据包含外卖订单时）

---

## 📊 数据模型

无需修改数据模型，使用现有的 `model.SaleBill` 和 `order.IsOrderSourceTakeout()` 方法。

---

## 🔌 实现设计

### 方案 1: 出菜单标题旁添加外卖标识

**位置**: `dishes_img.go` 和 `dishes_xprinter.go` 的 `OutMenuTemplate` 方法

**实现**:

```go
// 当前代码（第 689 行）
img.AppendText(t.base.Translate("出菜单"))

// 修改后
menuTitle := t.base.Translate("出菜单")
if order.IsOrderSourceTakeout() {
    menuTitle += "(" + t.base.Translate("外卖") + ")"
}
img.AppendText(menuTitle)
```

### 方案 2: 桌号/序号前添加外卖标识

**位置**: 所有打印模板的桌号/序号显示处

**实现**:

```go
// 当前代码（桌号/序号显示处，如第 94-111 行）
if order.DeskUuid > 0 {
    printer.AppendText(t.base.Translate("桌号") + ": " + order.SerialNo + mealNumStr)
} else if order.IsTakeoutBill() {
    printer.AppendText(t.base.Translate("外送") + ": " + order.SerialNo + "\n")
} else {
    printer.AppendText(t.base.Translate("取单号") + ": " + order.SerialNo + "\n")
}

// 修改后
serialNoText := order.SerialNo
if order.IsOrderSourceTakeout() {
    serialNoText = t.base.Translate("外卖") + " " + serialNoText
}
if order.DeskUuid > 0 {
    printer.AppendText(t.base.Translate("桌号") + ": " + serialNoText + mealNumStr)
} else if order.IsTakeoutBill() {
    printer.AppendText(t.base.Translate("外送") + ": " + serialNoText + "\n")
} else {
    printer.AppendText(t.base.Translate("取单号") + ": " + serialNoText + "\n")
}
```

### 方案 3: 交班单和营业数据单标题添加外卖标识

**位置**: `handover_*.go` 和 `business_data_*.go` 的 `GetPrintContent` 方法

**实现**:

```go
// 当前代码（handover_img.go 第 64 行）
img.AppendText(t.base.Translate("交班单"))

// 修改后 - 需要判断数据是否包含外卖订单
// 方式1: 通过 businessData 中的订单来源判断
// 方式2: 通过参数传递标识
handoverTitle := t.base.Translate("交班单")
if hasTakeoutOrders {
    handoverTitle += "(" + t.base.Translate("外卖") + ")"
}
img.AppendText(handoverTitle)
```

**判断逻辑**:
- 交班单和营业数据单是统计类打印单，需要判断统计数据中是否包含外卖订单
- 可以通过检查 `businessData` 中的订单来源（`OrderSourceUuid > 0`）来判断
- 或者通过调用方传递参数标识是否包含外卖订单数据

### 多语言支持

使用现有的 `t.base.Translate("外卖")` 方法，该方法会自动根据当前语言返回对应的翻译。

**支持的翻译**（根据任务描述）:
- 中文: 外卖
- 泰语: อาหาร
- 繁体中文: 外賣
- 英语: Food Delivery
- 日语: 出前
- 韩语: 배달
- 土耳其语: Paket Servis
- 缅甸语: အစားအစာ ပို့ဆောင်ခြင်း
- 瑞典语: Takeaway
- 老挝语: ອາຫານສັ່ງໃຫ້ມາສົ່ງ

---

## 🧩 具体实现

### dishes_img.go - OutMenuTemplate

**修改点 1**: 标题添加外卖标识（约第 689 行）

```go
// 修改前
img.AppendText(t.base.Translate("出菜单"))

// 修改后
menuTitle := t.base.Translate("出菜单")
if order.IsOrderSourceTakeout() {
    menuTitle += "(" + t.base.Translate("外卖") + ")"
}
img.AppendText(menuTitle)
```

**修改点 2**: 桌号/序号前添加外卖标识（约第 94-111 行附近）

```go
// 修改前
if order.DeskUuid > 0 {
    printer.AppendText(t.base.Translate("桌号") + ": " + order.SerialNo + mealNumStr)
} else if order.IsTakeoutBill() {
    printer.AppendText(t.base.Translate("外送") + ": " + order.SerialNo + "\n")
} else {
    printer.AppendText(t.base.Translate("取单号") + ": " + order.SerialNo + "\n")
}

// 修改后
serialNoText := order.SerialNo
if order.IsOrderSourceTakeout() {
    serialNoText = t.base.Translate("外卖") + " " + serialNoText
}
if order.DeskUuid > 0 {
    printer.AppendText(t.base.Translate("桌号") + ": " + serialNoText + mealNumStr)
} else if order.IsTakeoutBill() {
    printer.AppendText(t.base.Translate("外送") + ": " + serialNoText + "\n")
} else {
    printer.AppendText(t.base.Translate("取单号") + ": " + serialNoText + "\n")
}
```

### dishes_xprinter.go - OutMenuTemplate

**修改点 1**: 标题添加外卖标识（约第 1182 行）

```go
// 修改前
printer.AppendText(t.base.Translate("出菜单"))

// 修改后
menuTitle := t.base.Translate("出菜单")
if order.IsOrderSourceTakeout() {
    menuTitle += "(" + t.base.Translate("外卖") + ")"
}
printer.AppendText(menuTitle)
```

**修改点 2**: 桌号/序号前添加外卖标识（约第 94-111 行附近）

```go
// 修改前
if order.DeskUuid > 0 {
    printer.AppendText(t.base.Translate("桌号") + ": " + order.SerialNo + mealNumStr)
} else if order.IsTakeoutBill() {
    printer.AppendText(t.base.Translate("外送") + ": " + order.SerialNo + "\n")
} else {
    printer.AppendText(t.base.Translate("取单号") + ": " + order.SerialNo + "\n")
}

// 修改后
serialNoText := order.SerialNo
if order.IsOrderSourceTakeout() {
    serialNoText = t.base.Translate("外卖") + " " + serialNoText
}
if order.DeskUuid > 0 {
    printer.AppendText(t.base.Translate("桌号") + ": " + serialNoText + mealNumStr)
} else if order.IsTakeoutBill() {
    printer.AppendText(t.base.Translate("外送") + ": " + serialNoText + "\n")
} else {
    printer.AppendText(t.base.Translate("取单号") + ": " + serialNoText + "\n")
}
```

### dishes_codesoft.go（可选）

**修改点**: CompleteOrder 和 OneDishOneOrder 方法的桌号/序号显示处

```go
// 修改前
if order.DeskUuid > 0 {
    printer.AppendText(t.base.Translate("桌号") + ": " + order.SerialNo + mealNumStr)
} else if order.IsTakeoutBill() {
    printer.AppendText(t.base.Translate("外送") + ": " + order.SerialNo + "\n")
} else {
    printer.AppendText(t.base.Translate("取单号") + ": " + order.SerialNo + "\n")
}

// 修改后
serialNoText := order.SerialNo
if order.IsOrderSourceTakeout() {
    serialNoText = t.base.Translate("外卖") + " " + serialNoText
}
if order.DeskUuid > 0 {
    printer.AppendText(t.base.Translate("桌号") + ": " + serialNoText + mealNumStr)
} else if order.IsTakeoutBill() {
    printer.AppendText(t.base.Translate("外送") + ": " + serialNoText + "\n")
} else {
    printer.AppendText(t.base.Translate("取单号") + ": " + serialNoText + "\n")
}
```

---

## 🧪 测试策略

### 单元测试

**测试内容**:
- 外卖订单的打印模板输出包含外卖标识
- 非外卖订单的打印模板输出不包含外卖标识
- 多语言翻译正确

**测试文件**:
- `main/app/printer/template/dishes_img_test.go`（如需要）
- `main/app/printer/template/dishes_xprinter_test.go`（如需要）

### 集成测试

**测试流程**:
- 创建外卖订单，打印出菜单，验证外卖标识显示
- 创建堂食订单，打印出菜单，验证不显示外卖标识
- 切换不同语言，验证翻译正确

---

## 📈 性能优化

无需特殊优化，只是字符串拼接操作，性能影响可忽略。

---

## 📚 实现清单

### Phase 1: 出菜单模板修改

- [ ] 修改 `dishes_img.go` 的 `OutMenuTemplate` 方法
- [ ] 修改 `dishes_xprinter.go` 的 `OutMenuTemplate` 方法
- [ ] 测试出菜单打印输出

### Phase 2: 其他打印模板修改（可选）

- [ ] 修改 `dishes_codesoft.go` 的 `CompleteOrder` 方法
- [ ] 修改 `dishes_codesoft.go` 的 `OneDishOneOrder` 方法
- [ ] 测试其他打印单输出

### Phase 3: 测试和验证

- [ ] 单元测试（如需要）
- [ ] 集成测试
- [ ] 多语言测试

**详细任务**: 参见 `tasks.md`

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`

---

**版本**: v1.0.0  
**创建日期**: 2025-11-20  
**作者**: weifashi  
**审核者**: 待审核

