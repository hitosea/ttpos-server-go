# 打印单添加外卖标识 需求文档

> 本文档定义打印单添加外卖标识功能的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | 无（直接来自任务 36917）                                                                                    |
| **创建日期**      | 2025-11-20                                                                                                   |
| **负责人**        | weifashi                                                                                                     |
| **目标 Sprint**   | Sprint v2.10.0                                                                                              |
| **涉及技术栈**    | [x] Go (main/) [ ] Go (ttpos-bmp/) [ ] PHP (admin/) [ ] Vue (admin/views/)                                   |

---

## 📋 概述

在打印单中添加外卖标识，当订单为外卖订单时，在打印单的标题和桌号/序号前显示外卖标识，帮助后厨和收银人员快速识别外卖订单。

## 🎯 产品对齐

提升后厨和收银人员的工作效率，通过视觉标识快速区分外卖订单和堂食订单，减少出错率。

## 📝 用户故事

**作为** 后厨人员/收银员  
**我想** 在打印单上看到外卖标识  
**以便于** 快速识别外卖订单，避免混淆

---

## 功能需求

### Requirement 1: 出菜单添加外卖标识

**用户故事**: 作为后厨人员，我想在出菜单上看到外卖标识，以便于快速识别外卖订单

#### 验收标准

1. **WHEN** 订单为外卖订单 **THEN** 出菜单标题"出菜单"旁边 **SHALL** 显示"(外卖)"标识
2. **WHEN** 订单为外卖订单 **THEN** 桌号/序号前 **SHALL** 显示外卖标识
3. **WHEN** 订单为外卖订单 **THEN** 外卖标识 **SHALL** 支持多语言显示

#### 具体要求

- [ ] 1.1 在 `dishes_img.go` 的 `OutMenuTemplate` 方法中，当 `order.IsOrderSourceTakeout()` 为 true 时，在"出菜单"标题后添加"(外卖)"标识
- [ ] 1.2 在 `dishes_xprinter.go` 的 `OutMenuTemplate` 方法中，当 `order.IsOrderSourceTakeout()` 为 true 时，在"出菜单"标题后添加"(外卖)"标识
- [ ] 1.3 在桌号/序号显示时，如果是外卖订单，在桌号/序号前添加外卖标识（使用 `t.base.Translate("外卖")`）
- [ ] 1.4 外卖标识需要支持多语言（中文、英文、日语、韩语、泰语、缅甸语、土耳其语、老挝语、瑞典语等）

---

### Requirement 2: 一菜一单添加外卖标识

**用户故事**: 作为后厨人员，我想在一菜一单上看到外卖标识，以便于快速识别外卖订单

#### 验收标准

1. **WHEN** 订单为外卖订单 **THEN** 一菜一单的桌号/序号前 **SHALL** 显示外卖标识

#### 具体要求

- [ ] 2.1 在 `dishes_codesoft.go` 的 `OneDishOneOrder` 方法中，桌号/序号前添加外卖标识
- [ ] 2.2 在 `dishes_img.go` 的 `OneDishOneOrder` 方法中（如存在），桌号/序号前添加外卖标识
- [ ] 2.3 在 `dishes_xprinter.go` 的 `OneDishOneOrder` 方法中（如存在），桌号/序号前添加外卖标识

---

### Requirement 3: 整单打印添加外卖标识

**用户故事**: 作为后厨人员，我想在整单打印上看到外卖标识，以便于快速识别外卖订单

#### 验收标准

1. **WHEN** 订单为外卖订单 **THEN** 整单打印的桌号/序号前 **SHALL** 显示外卖标识

#### 具体要求

- [ ] 3.1 在 `dishes_codesoft.go` 的 `CompleteOrder` 方法中，桌号/序号前添加外卖标识
- [ ] 3.2 在 `dishes_img.go` 的 `CompleteOrder` 方法中，桌号/序号前添加外卖标识
- [ ] 3.3 在 `dishes_xprinter.go` 的 `CompleteOrder` 方法中（如存在），桌号/序号前添加外卖标识

---

### Requirement 4: 退菜单添加外卖标识

**用户故事**: 作为收银员，我想在退菜单上看到外卖标识，以便于快速识别外卖订单

#### 验收标准

1. **WHEN** 订单为外卖订单 **THEN** 退菜单标题"退菜单"旁边 **SHALL** 显示"(外卖)"标识
2. **WHEN** 订单为外卖订单 **THEN** 退菜单的桌号/序号前 **SHALL** 显示外卖标识

#### 具体要求

- [ ] 4.1 在 `dishes_img.go` 的 `ReturnMenuTemplate` 方法中，当 `order.IsOrderSourceTakeout()` 为 true 时，在"退菜单"标题后添加"(外卖)"标识
- [ ] 4.2 在 `dishes_xprinter.go` 的 `ReturnMenuTemplate` 方法中，当 `order.IsOrderSourceTakeout()` 为 true 时，在"退菜单"标题后添加"(外卖)"标识
- [ ] 4.3 在 `dishes_img.go` 的 `ReturnMenuTemplate` 方法中，桌号/序号前添加外卖标识
- [ ] 4.4 在 `dishes_xprinter.go` 的 `ReturnMenuTemplate` 方法中，桌号/序号前添加外卖标识

---

### Requirement 5: 预结账单添加外卖标识

**用户故事**: 作为收银员，我想在预结账单上看到外卖标识，以便于快速识别外卖订单

#### 验收标准

1. **WHEN** 订单为外卖订单 **THEN** 预结账单标题"预结账单"旁边 **SHALL** 显示"(外卖)"标识
2. **WHEN** 订单为外卖订单 **THEN** 预结账单的桌号/序号前 **SHALL** 显示外卖标识

#### 具体要求

- [ ] 5.1 在 `statement_order_img.go` 的 `GetPrintContent` 方法中，当 `printOrderType == constant.PrinterTemplatePreBilling` 且订单为外卖订单时，在"预结账单"标题后添加"(外卖)"标识
- [ ] 5.2 在 `statement_order_xprinter.go` 的 `GetPrintContent` 方法中，当 `printType == constant.PrinterTemplatePreBilling` 且订单为外卖订单时，在"预结账单"标题后添加"(外卖)"标识
- [ ] 5.3 在 `statement_order_codesoft.go` 的 `GetPrintContent` 方法中，当 `printType == constant.PrinterTemplatePreBilling` 且订单为外卖订单时，在"预结账单"标题后添加"(外卖)"标识
- [ ] 5.4 在所有预结账单模板中，桌号/序号前添加外卖标识

---

### Requirement 6: 结账单添加外卖标识

**用户故事**: 作为收银员，我想在结账单上看到外卖标识，以便于快速识别外卖订单

#### 验收标准

1. **WHEN** 订单为外卖订单 **THEN** 结账单标题"结账单"旁边 **SHALL** 显示"(外卖)"标识
2. **WHEN** 订单为外卖订单 **THEN** 结账单的桌号/序号前 **SHALL** 显示外卖标识

#### 具体要求

- [ ] 6.1 在 `statement_order_img.go` 的 `GetPrintContent` 方法中，当 `printOrderType == constant.PrinterTemplateBilling` 且订单为外卖订单时，在"结账单"标题后添加"(外卖)"标识
- [ ] 6.2 在 `statement_order_xprinter.go` 的 `GetPrintContent` 方法中，当 `printType == constant.PrinterTemplateBilling` 且订单为外卖订单时，在"结账单"标题后添加"(外卖)"标识
- [ ] 6.3 在 `statement_order_codesoft.go` 的 `GetPrintContent` 方法中，当 `printType == constant.PrinterTemplateBilling` 且订单为外卖订单时，在"结账单"标题后添加"(外卖)"标识
- [ ] 6.4 在所有结账单模板中，桌号/序号前添加外卖标识

---

### Requirement 7: 发票添加外卖标识

**用户故事**: 作为收银员，我想在发票上看到外卖标识，以便于快速识别外卖订单

#### 验收标准

1. **WHEN** 订单为外卖订单 **THEN** 发票标题"发票"旁边 **SHALL** 显示"(外卖)"标识
2. **WHEN** 订单为外卖订单 **THEN** 发票的桌号/序号前 **SHALL** 显示外卖标识

#### 具体要求

- [ ] 7.1 在 `statement_order_img.go` 的 `GetPrintContent` 方法中，当 `printOrderType == constant.PrinterTemplateInvoice` 且订单为外卖订单时，在"发票"标题后添加"(外卖)"标识
- [ ] 7.2 在 `statement_order_xprinter.go` 的 `GetPrintContent` 方法中，当 `printType == constant.PrinterTemplateInvoice` 且订单为外卖订单时，在"发票"标题后添加"(外卖)"标识
- [ ] 7.3 在 `statement_order_codesoft.go` 的 `GetPrintContent` 方法中，当 `printType == constant.PrinterTemplateInvoice` 且订单为外卖订单时，在"发票"标题后添加"(外卖)"标识
- [ ] 7.4 在 `invoice_img.go`、`invoice_xprinter.go`、`invoice_codesoft.go` 等发票专用模板中，标题和桌号/序号前添加外卖标识
- [ ] 7.5 在所有发票模板中，桌号/序号前添加外卖标识

---

### Requirement 8: 交班单添加外卖标识

**用户故事**: 作为店长，我想在交班单上看到外卖标识，以便于区分外卖和堂食的统计数据

#### 验收标准

1. **WHEN** 交班单数据包含外卖订单 **THEN** 交班单标题"交班单"旁边 **SHALL** 显示"(外卖)"标识
2. **WHEN** 交班单数据包含外卖订单 **THEN** 交班单 **SHALL** 根据订单来源区分数据

#### 具体要求

- [ ] 8.1 在 `handover_img.go` 的 `GetPrintContent` 方法中，当数据包含外卖订单时，在"交班单"标题后添加"(外卖)"标识
- [ ] 8.2 在 `handover_xprinter.go` 的 `GetPrintContent` 方法中，当数据包含外卖订单时，在"交班单"标题后添加"(外卖)"标识
- [ ] 8.3 在 `handover_codesoft.go` 的 `GetPrintContent` 方法中，当数据包含外卖订单时，在"交班单"标题后添加"(外卖)"标识
- [ ] 8.4 在 `handover_compax.go` 的 `GetPrintContent` 方法中（如需要），当数据包含外卖订单时，在"交班单"标题后添加"(外卖)"标识
- [ ] 8.5 在 `handover_sunmi.go` 的 `GetPrintContent` 方法中（如需要），当数据包含外卖订单时，在"交班单"标题后添加"(外卖)"标识
- [ ] 8.6 在 `handover_img_58mm.go` 的 `GetPrintContent58mm` 方法中，当数据包含外卖订单时，在"交班单"标题后添加"(外卖)"标识

---

### Requirement 9: 营业数据单添加外卖标识

**用户故事**: 作为店长，我想在营业数据单上看到外卖标识，以便于区分外卖和堂食的统计数据

#### 验收标准

1. **WHEN** 营业数据包含外卖订单 **THEN** 营业数据单标题"营业数据"旁边 **SHALL** 显示"(外卖)"标识
2. **WHEN** 营业数据包含外卖订单 **THEN** 营业数据单 **SHALL** 根据订单来源区分数据

#### 具体要求

- [ ] 9.1 在 `business_data_img.go` 的 `GetPrintContent` 方法中，当数据包含外卖订单时，在"营业数据"标题后添加"(外卖)"标识
- [ ] 9.2 在 `business_data_xprinter.go` 的 `GetPrintContent` 方法中，当数据包含外卖订单时，在"营业数据"标题后添加"(外卖)"标识
- [ ] 9.3 在 `business_data_img_58mm.go` 的 `GetPrintContent` 方法中（如需要），当数据包含外卖订单时，在"营业数据"标题后添加"(外卖)"标识
- [ ] 9.4 在 `business_data_sunmi.go` 的 `GetPrintContent` 方法中（如需要），当数据包含外卖订单时，在"营业数据"标题后添加"(外卖)"标识

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 严格遵循现有打印模板结构
- **单一职责原则**: 每个模板文件负责自己的打印逻辑
- **模块化设计**: 复用现有的翻译和多语言支持
- **遵循规范**:
  - `.cursor/rules/go-main.mdc` - Go Main 开发规范

### 国际化要求

- [ ] 支持 10 种语言（中文、英文、日语、韩语、泰语、缅甸语、土耳其语、老挝语、瑞典语等）
- [ ] 所有文案使用 `t.base.Translate("外卖")` 实现多语言
- [ ] 参考: `main/i18n/` - 国际化配置

### 测试要求

- [ ] 单元测试覆盖打印模板修改
- [ ] 集成测试验证打印单输出格式
- [ ] 多语言测试验证翻译正确性

---

## 验收标准

### 功能验收

1. **出菜单外卖标识**: 外卖订单的出菜单标题旁显示"(外卖)"标识，桌号/序号前显示外卖标识
2. **一菜一单外卖标识**: 外卖订单的一菜一单桌号/序号前显示外卖标识
3. **整单打印外卖标识**: 外卖订单的整单打印桌号/序号前显示外卖标识
4. **退菜单外卖标识**: 外卖订单的退菜单标题旁显示"(外卖)"标识，桌号/序号前显示外卖标识
5. **预结账单外卖标识**: 外卖订单的预结账单标题旁显示"(外卖)"标识，桌号/序号前显示外卖标识
6. **结账单外卖标识**: 外卖订单的结账单标题旁显示"(外卖)"标识，桌号/序号前显示外卖标识
7. **发票外卖标识**: 外卖订单的发票标题旁显示"(外卖)"标识，桌号/序号前显示外卖标识
8. **交班单外卖标识**: 包含外卖订单数据的交班单标题旁显示"(外卖)"标识
9. **营业数据单外卖标识**: 包含外卖订单数据的营业数据单标题旁显示"(外卖)"标识
10. **多语言支持**: 外卖标识在所有支持的语言中正确显示
11. **非外卖订单**: 非外卖订单不显示外卖标识

### 测试验收

1. **单元测试**: 打印模板方法测试通过
2. **集成测试**: 实际打印输出格式正确
3. **多语言测试**: 所有语言的外卖标识显示正确

---

## 约束条件

### 技术约束

#### Go Main 模块

- 必须使用现有的打印模板结构
- 复用现有的翻译方法 `t.base.Translate()`
- 复用现有的订单判断方法 `order.IsOrderSourceTakeout()`
- 不使用 panic，返回 error

### 业务约束

- 只影响打印单显示，不影响业务逻辑
- 保持现有打印单格式和布局

### 资源约束

- 开发时间: 3-4 天
- Story Point: 3 (必须 ≤ 5)

---

## 依赖关系

### 技术依赖

- `main/app/printer/template/` - 打印模板模块
- `main/app/model/` - 订单模型（SaleBill）
- `main/i18n/` - 国际化支持

### 业务依赖

- 订单模型需要支持 `IsOrderSourceTakeout()` 方法（已存在）

---

## 风险和缓解

### 风险 1: 打印格式混乱

**影响**: 中  
**概率**: 低  
**缓解措施**:

- 仔细测试各种打印机类型
- 保持现有格式和布局

### 风险 2: 多语言翻译缺失

**影响**: 低  
**概率**: 低  
**缓解措施**:

- 检查现有翻译文件是否包含"外卖"翻译
- 如缺失，添加翻译

---

## 时间表

- **Phase 1 - 出菜单模板修改**: 0.5 天
- **Phase 2 - 一菜一单和整单打印模板修改**: 0.5 天
- **Phase 3 - 退菜单模板修改**: 0.5 天
- **Phase 4 - 预结账单、结账单、发票模板修改**: 1 天
- **Phase 5 - 交班单和营业数据单模板修改**: 0.5 天
- **Phase 6 - 测试和验证**: 0.5 天
- **总计**: 3.5 天（SP = 3）

---

## 参考资料

### 核心规范

- `.cursor/rules/go-main.mdc` - Go Main 核心约束
- `.cursor/rules/api.mdc` - API 设计规范

### 相关代码

- `main/app/printer/template/dishes_img.go` - 图片打印模板
- `main/app/printer/template/dishes_xprinter.go` - XPrinter 打印模板
- `main/app/printer/template/dishes_codesoft.go` - Codesoft 打印模板
- `main/app/model/sale_bill.go` - 订单模型

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

