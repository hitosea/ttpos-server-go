# 打印单添加外卖标识 任务分解

> 本文档定义打印单添加外卖标识功能的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-2 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 35  
**已完成**: 19  
**进行中**: -  
**完成率**: 54%

---

## Phase 1: 出菜单模板修改

- [x] 1.1 修改 dishes_img.go 的 OutMenuTemplate 方法 - 标题添加外卖标识

  - File: `main/app/printer/template/dishes_img.go`
  - Purpose: 在桌号/序号前添加外卖标识
  - Requirements: 1.1
  - Leverage: 现有代码第 689 行，使用 `order.IsOrderSourceTakeout()` 和 `t.base.Translate("外卖")`
  - Prompt: Role: Go Developer | Task: 在 dishes_img.go 的 OutMenuTemplate 方法中，当订单为外卖订单时，在"出菜单"标题后添加"(外卖)"标识 | Context: 当前代码在第 689 行 `img.AppendText(t.base.Translate("出菜单"))`，需要判断 `order.IsOrderSourceTakeout()`，如果是外卖订单，则修改为 `img.AppendText(t.base.Translate("出菜单") + "(" + t.base.Translate("外卖") + ")")` | Restrictions: 遵循 .cursor/rules/go-main.mdc，保持代码风格一致 | Success: 外卖订单的出菜单标题显示为"出菜单(外卖)"

- [x] 1.2 修改 dishes_img.go 的 OutMenuTemplate 方法 - 订单号添加外卖标识

  - File: `main/app/printer/template/dishes_img.go`
  - Purpose: 在桌号/序号前添加外卖标识
  - Requirements: 1.3
  - Leverage: 现有代码在打印桌号/序号处（如第 94-111 行附近），使用 `order.SerialNo` 和 `t.base.Translate("外卖")`
  - Prompt: Role: Go Developer | Task: 在 dishes_img.go 的 OutMenuTemplate 方法中，当订单为外卖订单时，在桌号/序号前添加外卖标识 | Context: 当前代码在打印桌号/序号时（如第 94-111 行附近），需要判断 `order.IsOrderSourceTakeout()`，如果是外卖订单，则在桌号/序号前添加 `t.base.Translate("外卖") + " "` | Restrictions: 遵循 .cursor/rules/go-main.mdc，保持代码风格一致 | Success: 外卖订单的桌号/序号显示为"外卖 A01"

- [x] 1.3 修改 dishes_xprinter.go 的 OutMenuTemplate 方法 - 标题添加外卖标识

  - File: `main/app/printer/template/dishes_xprinter.go`
  - Purpose: 在出菜单标题"出菜单"后添加"(外卖)"标识
  - Requirements: 1.2
  - Leverage: 现有代码第 1182 行，使用 `order.IsOrderSourceTakeout()` 和 `t.base.Translate("外卖")`
  - Prompt: Role: Go Developer | Task: 在 dishes_xprinter.go 的 OutMenuTemplate 方法中，当订单为外卖订单时，在"出菜单"标题后添加"(外卖)"标识 | Context: 当前代码在第 1182 行 `printer.AppendText(t.base.Translate("出菜单"))`，需要判断 `order.IsOrderSourceTakeout()`，如果是外卖订单，则修改为 `printer.AppendText(t.base.Translate("出菜单") + "(" + t.base.Translate("外卖") + ")")` | Restrictions: 遵循 .cursor/rules/go-main.mdc，保持代码风格一致 | Success: 外卖订单的出菜单标题显示为"出菜单(外卖)"

- [x] 1.4 修改 dishes_xprinter.go 的 OutMenuTemplate 方法 - 订单号添加外卖标识

  - File: `main/app/printer/template/dishes_xprinter.go`
  - Purpose: 在桌号/序号前添加外卖标识
  - Requirements: 1.3
  - Leverage: 现有代码在打印桌号/序号处（如第 94-111 行附近），使用 `order.SerialNo` 和 `t.base.Translate("外卖")`
  - Prompt: Role: Go Developer | Task: 在 dishes_xprinter.go 的 OutMenuTemplate 方法中，当订单为外卖订单时，在桌号/序号前添加外卖标识 | Context: 当前代码在打印桌号/序号时（如第 94-111 行附近），需要判断 `order.IsOrderSourceTakeout()`，如果是外卖订单，则在桌号/序号前添加 `t.base.Translate("外卖") + " "` | Restrictions: 遵循 .cursor/rules/go-main.mdc，保持代码风格一致 | Success: 外卖订单的桌号/序号显示为"外卖 A01"

---

## Phase 2: 一菜一单和整单打印模板修改

- [x] 2.1 修改 dishes_codesoft.go 的 CompleteOrder 方法 - 桌号/序号添加外卖标识

  - File: `main/app/printer/template/dishes_codesoft.go`
  - Purpose: 在整单打印的桌号/序号前添加外卖标识
  - Requirements: 3.1
  - Leverage: 现有代码第 81-95 行附近，使用 `order.SerialNo` 和 `t.base.Translate("外卖")`
  - Prompt: Role: Go Developer | Task: 在 dishes_codesoft.go 的 CompleteOrder 方法中，当订单为外卖订单时，在桌号/序号前添加外卖标识 | Context: 当前代码在打印桌号/序号时（如第 81-95 行附近），需要判断 `order.IsOrderSourceTakeout()`，如果是外卖订单，则在桌号/序号前添加 `t.base.Translate("外卖") + " "` | Restrictions: 遵循 .cursor/rules/go-main.mdc，保持代码风格一致 | Success: 外卖订单的桌号/序号显示为"外卖 A01"

- [x] 2.2 修改 dishes_codesoft.go 的 OneDishOneOrder 方法 - 桌号/序号添加外卖标识

  - File: `main/app/printer/template/dishes_codesoft.go`
  - Purpose: 在一菜一单的桌号/序号前添加外卖标识
  - Requirements: 2.1
  - Leverage: 现有代码，使用 `order.SerialNo` 和 `t.base.Translate("外卖")`
  - Prompt: Role: Go Developer | Task: 在 dishes_codesoft.go 的 OneDishOneOrder 方法中，当订单为外卖订单时，在桌号/序号前添加外卖标识 | Context: 需要找到桌号/序号显示的位置，判断 `order.IsOrderSourceTakeout()`，如果是外卖订单，则在桌号/序号前添加 `t.base.Translate("外卖") + " "` | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 外卖订单的桌号/序号显示为"外卖 A01"

- [x] 2.3 修改 dishes_img.go 的 CompleteOrder 方法 - 桌号/序号添加外卖标识

  - File: `main/app/printer/template/dishes_img.go`
  - Purpose: 在整单打印的桌号/序号前添加外卖标识
  - Requirements: 3.2
  - Leverage: 现有代码，使用 `order.SerialNo` 和 `t.base.Translate("外卖")`
  - Success: 外卖订单的桌号/序号显示为"外卖 A01"

---

## Phase 3: 退菜单模板修改

- [x] 3.1 修改 dishes_img.go 的 ReturnMenuTemplate 方法 - 标题添加外卖标识

  - File: `main/app/printer/template/dishes_img.go`
  - Purpose: 在退菜单标题"退菜单"后添加"(外卖)"标识
  - Requirements: 4.1
  - Leverage: 现有代码第 496 行，使用 `order.IsOrderSourceTakeout()` 和 `t.base.Translate("外卖")`
  - Prompt: Role: Go Developer | Task: 在 dishes_img.go 的 ReturnMenuTemplate 方法中，当订单为外卖订单时，在"退菜单"标题后添加"(外卖)"标识 | Context: 当前代码在第 496 行 `img.AppendText(t.base.Translate("退菜单"))`，需要判断 `order.IsOrderSourceTakeout()`，如果是外卖订单，则修改为 `img.AppendText(t.base.Translate("退菜单") + "(" + t.base.Translate("外卖") + ")")` | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 外卖订单的退菜单标题显示为"退菜单(外卖)"

- [x] 3.2 修改 dishes_img.go 的 ReturnMenuTemplate 方法 - 订单号添加外卖标识

  - File: `main/app/printer/template/dishes_img.go`
  - Purpose: 在退菜单的桌号/序号前添加外卖标识
  - Requirements: 4.3
  - Leverage: 现有代码在打印桌号/序号处，使用 `order.SerialNo` 和 `t.base.Translate("外卖")`
  - Success: 外卖订单的桌号/序号显示为"外卖 A01"

- [x] 3.3 修改 dishes_xprinter.go 的 ReturnMenuTemplate 方法 - 标题添加外卖标识

  - File: `main/app/printer/template/dishes_xprinter.go`
  - Purpose: 在退菜单标题"退菜单"后添加"(外卖)"标识
  - Requirements: 4.2
  - Leverage: 现有代码第 919 行，使用 `order.IsOrderSourceTakeout()` 和 `t.base.Translate("外卖")`
  - Success: 外卖订单的退菜单标题显示为"退菜单(外卖)"

- [x] 3.4 修改 dishes_xprinter.go 的 ReturnMenuTemplate 方法 - 订单号添加外卖标识

  - File: `main/app/printer/template/dishes_xprinter.go`
  - Purpose: 在退菜单的桌号/序号前添加外卖标识
  - Requirements: 4.4
  - Leverage: 现有代码在打印桌号/序号处，使用 `order.SerialNo` 和 `t.base.Translate("外卖")`
  - Success: 外卖订单的桌号/序号显示为"外卖 A01"

---

## Phase 4: 预结账单、结账单、发票模板修改

- [x] 4.1 修改 statement_order_img.go 的 GetPrintContent 方法 - 预结账单标题和订单号添加外卖标识

  - File: `main/app/printer/template/statement_order_img.go`
  - Purpose: 在预结账单标题和桌号/序号前添加外卖标识
  - Requirements: 5.1, 5.4
  - Leverage: 现有代码第 75 行和第 117 行，使用 `saleBill.IsOrderSourceTakeout()` 和 `t.base.Translate("外卖")`
  - Prompt: Role: Go Developer | Task: 在 statement_order_img.go 的 GetPrintContent 方法中，当 printOrderType == constant.PrinterTemplatePreBilling 且订单为外卖订单时，在"预结账单"标题后添加"(外卖)"标识，在桌号/序号前添加外卖标识 | Context: 需要判断 `printOrderType == constant.PrinterTemplatePreBilling` 和 `saleBill.IsOrderSourceTakeout()` | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 外卖订单的预结账单标题显示为"预结账单(外卖)"，桌号/序号前显示"外卖"标识

- [x] 4.2 修改 statement_order_img.go 的 GetPrintContent 方法 - 结账单标题和订单号添加外卖标识

  - File: `main/app/printer/template/statement_order_img.go`
  - Purpose: 在结账单标题和桌号/序号前添加外卖标识
  - Requirements: 6.1, 6.4
  - Leverage: 现有代码，使用 `saleBill.IsOrderSourceTakeout()` 和 `t.base.Translate("外卖")`
  - Success: 外卖订单的结账单标题显示为"结账单(外卖)"，桌号/序号前显示"外卖"标识

- [x] 4.3 修改 statement_order_img.go 的 GetPrintContent 方法 - 发票标题和订单号添加外卖标识

  - File: `main/app/printer/template/statement_order_img.go`
  - Purpose: 在发票标题和桌号/序号前添加外卖标识
  - Requirements: 7.1, 7.5
  - Leverage: 现有代码，使用 `saleBill.IsOrderSourceTakeout()` 和 `t.base.Translate("外卖")`
  - Success: 外卖订单的发票标题显示为"发票(外卖)"，桌号/序号前显示"外卖"标识

- [x] 4.4 修改 statement_order_xprinter.go 的 GetPrintContent 方法 - 预结账单、结账单、发票标题和订单号添加外卖标识

  - File: `main/app/printer/template/statement_order_xprinter.go`
  - Purpose: 在预结账单、结账单、发票标题和桌号/序号前添加外卖标识
  - Requirements: 5.2, 6.2, 7.2
  - Leverage: 现有代码，使用 `saleBill.IsOrderSourceTakeout()` 和 `t.base.Translate("外卖")`
  - Success: 所有类型的打印单标题和桌号/序号前都显示外卖标识

- [x] 4.5 修改 statement_order_codesoft.go 的 GetPrintContent 方法 - 预结账单、结账单、发票标题和订单号添加外卖标识

  - File: `main/app/printer/template/statement_order_codesoft.go`
  - Purpose: 在预结账单、结账单、发票标题和桌号/序号前添加外卖标识
  - Requirements: 5.3, 6.3, 7.3
  - Leverage: 现有代码，使用 `saleBill.IsOrderSourceTakeout()` 和 `t.base.Translate("外卖")`
  - Success: 所有类型的打印单标题和桌号/序号前都显示外卖标识

- [x] 4.6 修改 invoice_img.go - 发票标题和订单号添加外卖标识

  - File: `main/app/printer/template/invoice_img.go`
  - Purpose: 在发票标题和桌号/序号前添加外卖标识
  - Requirements: 7.4
  - Leverage: 现有代码，使用 `saleBill.IsOrderSourceTakeout()` 和 `t.base.Translate("外卖")`
  - Success: 外卖订单的发票标题显示为"发票(外卖)"，桌号/序号前显示"外卖"标识

- [x] 4.7 修改 invoice_xprinter.go - 发票标题和订单号添加外卖标识

  - File: `main/app/printer/template/invoice_xprinter.go`
  - Purpose: 在发票标题和桌号/序号前添加外卖标识
  - Requirements: 7.4
  - Leverage: 现有代码，使用 `saleBill.IsOrderSourceTakeout()` 和 `t.base.Translate("外卖")`
  - Success: 外卖订单的发票标题显示为"发票(外卖)"，桌号/序号前显示"外卖"标识

- [x] 4.8 修改 invoice_img_58mm.go - 发票标题和订单号添加外卖标识

  - File: `main/app/printer/template/invoice_img_58mm.go`
  - Purpose: 在发票标题和桌号/序号前添加外卖标识
  - Requirements: 7.4
  - Leverage: 现有代码第 96 行和第 248 行，使用 `saleBill.IsOrderSourceTakeout()` 和 `t.base.Translate("外卖")`
  - Success: 外卖订单的发票标题显示为"发票(外卖)"，桌号/序号前显示"外卖"标识

---

## Phase 5: 交班单和营业数据单模板修改

- [ ] 5.1 修改 handover_img.go 的 GetPrintContent 方法 - 标题添加外卖标识

  - File: `main/app/printer/template/handover_img.go`
  - Purpose: 在交班单标题"交班单"后添加"(外卖)"标识
  - Requirements: 8.1
  - Leverage: 现有代码第 64 行和第 368 行，需要判断数据是否包含外卖订单
  - Prompt: Role: Go Developer | Task: 在 handover_img.go 的 GetPrintContent 方法中，当数据包含外卖订单时，在"交班单"标题后添加"(外卖)"标识 | Context: 需要判断 businessData 中是否包含外卖订单数据，可以通过检查订单来源或添加参数来判断 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 包含外卖订单数据的交班单标题显示为"交班单(外卖)"

- [ ] 5.2 修改 handover_xprinter.go 的 GetPrintContent 方法 - 标题添加外卖标识

  - File: `main/app/printer/template/handover_xprinter.go`
  - Purpose: 在交班单标题"交班单"后添加"(外卖)"标识
  - Requirements: 8.2
  - Leverage: 现有代码，需要判断数据是否包含外卖订单
  - Success: 包含外卖订单数据的交班单标题显示为"交班单(外卖)"

- [ ] 5.3 修改 handover_codesoft.go 的 GetPrintContent 方法 - 标题添加外卖标识

  - File: `main/app/printer/template/handover_codesoft.go`
  - Purpose: 在交班单标题"交班单"后添加"(外卖)"标识
  - Requirements: 8.3
  - Leverage: 现有代码，需要判断数据是否包含外卖订单
  - Success: 包含外卖订单数据的交班单标题显示为"交班单(外卖)"

- [ ] 5.4 修改 handover_img_58mm.go 的 GetPrintContent58mm 方法 - 标题添加外卖标识

  - File: `main/app/printer/template/handover_img_58mm.go`
  - Purpose: 在交班单标题"交班单"后添加"(外卖)"标识
  - Requirements: 8.6
  - Leverage: 现有代码，需要判断数据是否包含外卖订单
  - Success: 包含外卖订单数据的交班单标题显示为"交班单(外卖)"

- [ ] 5.5 修改 business_data_img.go 的 GetPrintContent 方法 - 标题添加外卖标识

  - File: `main/app/printer/template/business_data_img.go`
  - Purpose: 在营业数据单标题"营业数据"后添加"(外卖)"标识
  - Requirements: 9.1
  - Leverage: 现有代码第 57 行，需要判断数据是否包含外卖订单
  - Prompt: Role: Go Developer | Task: 在 business_data_img.go 的 GetPrintContent 方法中，当数据包含外卖订单时，在"营业数据"标题后添加"(外卖)"标识 | Context: 需要判断 businessData 中是否包含外卖订单数据，可以通过检查订单来源或添加参数来判断 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 包含外卖订单数据的营业数据单标题显示为"营业数据(外卖)"

- [ ] 5.6 修改 business_data_xprinter.go 的 GetPrintContent 方法 - 标题添加外卖标识

  - File: `main/app/printer/template/business_data_xprinter.go`
  - Purpose: 在营业数据单标题"营业数据"后添加"(外卖)"标识
  - Requirements: 9.2
  - Leverage: 现有代码第 66 行，需要判断数据是否包含外卖订单
  - Success: 包含外卖订单数据的营业数据单标题显示为"营业数据(外卖)"

---

## Phase 6: 测试和验证

- [ ] 5.1 测试出菜单打印输出

  - File: -
  - Purpose: 验证外卖订单的出菜单正确显示外卖标识
  - Requirements: Requirement 1
  - Command: 创建外卖订单，打印出菜单，检查输出
  - Success: 外卖订单的出菜单标题显示"(外卖)"，桌号/序号前显示"外卖"标识

- [ ] 5.2 测试一菜一单和整单打印输出

  - File: -
  - Purpose: 验证外卖订单的一菜一单和整单打印正确显示外卖标识
  - Requirements: Requirement 2, 3
  - Command: 创建外卖订单，打印一菜一单和整单打印，检查输出
  - Success: 外卖订单的桌号/序号前显示"外卖"标识

- [ ] 5.3 测试退菜单打印输出

  - File: -
  - Purpose: 验证外卖订单的退菜单正确显示外卖标识
  - Requirements: Requirement 4
  - Command: 创建外卖订单，打印退菜单，检查输出
  - Success: 外卖订单的退菜单标题显示"(外卖)"，桌号/序号前显示"外卖"标识

- [ ] 6.4 测试预结账单、结账单、发票打印输出

  - File: -
  - Purpose: 验证外卖订单的预结账单、结账单、发票正确显示外卖标识
  - Requirements: Requirement 5, 6, 7
  - Command: 创建外卖订单，打印预结账单、结账单、发票，检查输出
  - Success: 所有类型的打印单标题显示"(外卖)"，桌号/序号前显示"外卖"标识

- [ ] 6.5 测试交班单和营业数据单打印输出

  - File: -
  - Purpose: 验证包含外卖订单数据的交班单和营业数据单正确显示外卖标识
  - Requirements: Requirement 8, 9
  - Command: 生成包含外卖订单数据的交班单和营业数据单，检查输出
  - Success: 包含外卖订单数据的交班单和营业数据单标题显示"(外卖)"标识

- [ ] 6.6 测试多语言支持

  - File: -
  - Purpose: 验证外卖标识在所有支持的语言中正确显示
  - Requirements: 1.4
  - Command: 切换不同语言，打印外卖订单的各种打印单，检查翻译
  - Success: 所有语言的外卖标识翻译正确

- [ ] 6.7 测试非外卖订单

  - File: -
  - Purpose: 验证非外卖订单不显示外卖标识
  - Requirements: 功能验收
  - Command: 创建堂食订单，打印各种打印单，检查输出
  - Success: 非外卖订单的所有打印单都不显示外卖标识

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 `go fmt` 和 `go vet`
- [ ] 代码风格与现有代码一致

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成

### 测试验证

- [ ] 外卖订单的出菜单正确显示外卖标识
- [ ] 非外卖订单的出菜单不显示外卖标识
- [ ] 多语言翻译正确

### 规范遵循

- [ ] 遵循 `.cursor/rules/go-main.mdc`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/story-printer-takeout-identifier/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/story-printer-takeout-identifier/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/story-printer-takeout-identifier/tasks.md
```

### 执行流程

1. **选择任务**: 选择下一个未完成任务
2. **阅读需求**: 查看 requirements.md 中的关联需求
3. **查看复用**: 检查 Leverage 中的可复用代码
4. **使用 AI**: 复制 Prompt 模板，让 AI 生成代码
5. **实现代码**: 按照规范实现功能
6. **运行检查**: `go fmt`, `go vet`
7. **标记完成**: 将 `[ ]` 改为 `[x]`
8. **提交代码**: Git commit（参考 `.cursor/rules/version.mdc`）

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`

---

**模板版本**: v1.0.0  
**最后更新**: 2025-11-20  
**维护者**: weifashi

