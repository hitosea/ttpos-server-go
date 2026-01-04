# 单品备注预设备注打印 任务分解

> 本文档定义单品备注预设备注打印功能的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 20  
**已完成**: 17  
**进行中**: -  
**完成率**: 85%

---

## Phase 1: CodeSoft 打印模板修改

- [x] 1.1 修改 CodeSoft 模板 - 一菜一单打印

  - File: `main/app/printer/template/dishes_codesoft.go` (line 206-219)
  - Purpose: 在一菜一单打印时，将 `product.Remark` 改为拼接后的备注信息
  - Requirements: 1.1, 1.5-1.9
  - Leverage: 
    - `main/app/model/sale_order_product.go` - `GetOrderItemRemark()` (line 1197-1210)
    - `main/app/model/sale_order_product.go` - `BuildOrderItemRemarkInfo()` (line 1212-1261)
    - 整单备注实现: `dishes_codesoft.go` (line 629-636)
  - Prompt: Role: Go Developer specializing in Printer Template | Task: 修改 CodeSoft 模板的一菜一单打印逻辑，将 product.Remark 改为使用拼接后的备注信息（包含预设备注和自定义备注） | Context: 使用 product.GetOrderItemRemark() 获取预设备注，使用 product.BuildOrderItemRemarkInfo() 构建备注信息，使用 remarkInfo.Remark.GetLocale(t.base.Lang) 获取多语言内容 | Restrictions: 保持原有的行间距和字符大小设置逻辑，参考整单备注的实现方式 | Success: 备注信息正确拼接并显示，多语言支持正确

- [x] 1.2 修改 CodeSoft 模板 - 整单打印

  - File: `main/app/printer/template/dishes_codesoft.go` (line 378-389)
  - Purpose: 在整单打印时，将 `product.Remark` 改为拼接后的备注信息
  - Requirements: 1.1, 1.5-1.9
  - Leverage: Task 1.1 的实现逻辑
  - Success: 备注信息正确拼接并显示

- [x] 1.3 修改 CodeSoft 模板 - 退菜单打印

  - File: `main/app/printer/template/dishes_codesoft.go` (line 570-576)
  - Purpose: 在退菜单打印时，将 `product.Remark` 改为拼接后的备注信息
  - Requirements: 1.1, 1.5-1.9
  - Leverage: Task 1.1 的实现逻辑
  - Success: 备注信息正确拼接并显示

- [x] 1.4 修改 CodeSoft 模板 - 出菜单打印

  - File: `main/app/printer/template/dishes_codesoft.go` (line 718-723)
  - Purpose: 在出菜单打印时，将 `product.Remark` 改为拼接后的备注信息
  - Requirements: 1.1, 1.5-1.9
  - Leverage: Task 1.1 的实现逻辑
  - Success: 备注信息正确拼接并显示

---

## Phase 2: XPrinter 打印模板修改

- [x] 2.1 修改 XPrinter 模板 - 一菜一单打印（模板一）

  - File: `main/app/printer/template/dishes_xprinter.go` (line 229-241)
  - Purpose: 在一菜一单打印时，将 `product.Remark` 改为拼接后的备注信息
  - Requirements: 1.2, 1.5-1.9
  - Leverage: 
    - Task 1.1 的实现逻辑
    - 整单备注实现: `dishes_xprinter.go` (line 719-726)
  - Success: 备注信息正确拼接并显示

- [x] 2.2 修改 XPrinter 模板 - 整单打印（模板一）

  - File: `main/app/printer/template/dishes_xprinter.go` (line 428-440)
  - Purpose: 在整单打印时，将 `product.Remark` 改为拼接后的备注信息
  - Requirements: 1.2, 1.5-1.9
  - Leverage: Task 2.1 的实现逻辑
  - Success: 备注信息正确拼接并显示

- [x] 2.3 修改 XPrinter 模板 - 退菜单打印（模板二）

  - File: `main/app/printer/template/dishes_xprinter.go` (line 652-664)
  - Purpose: 在退菜单打印时，将 `product.Remark` 改为拼接后的备注信息
  - Requirements: 1.2, 1.5-1.9
  - Leverage: Task 2.1 的实现逻辑
  - Success: 备注信息正确拼接并显示

- [x] 2.4 修改 XPrinter 模板 - 出菜单打印（模板二）

  - File: `main/app/printer/template/dishes_xprinter.go` (line 834-840)
  - Purpose: 在出菜单打印时，将 `product.Remark` 改为拼接后的备注信息
  - Requirements: 1.2, 1.5-1.9
  - Leverage: Task 2.1 的实现逻辑
  - Success: 备注信息正确拼接并显示

- [x] 2.5 修改 XPrinter 模板 - 其他打印场景（模板三）

  - File: `main/app/printer/template/dishes_xprinter.go` (line 1078-1084, 1349-1355)
  - Purpose: 在其他打印场景时，将 `product.Remark` 改为拼接后的备注信息
  - Requirements: 1.2, 1.5-1.9
  - Leverage: Task 2.1 的实现逻辑
  - Success: 备注信息正确拼接并显示

---

## Phase 3: 图片打印模板修改

- [x] 3.1 修改图片模板 - 一菜一单打印

  - File: `main/app/printer/template/dishes_img.go` (line 290-301)
  - Purpose: 在一菜一单打印时，将 `product.Remark` 改为拼接后的备注信息
  - Requirements: 1.3, 1.5-1.9
  - Leverage: 
    - Task 1.1 的实现逻辑
    - 整单备注实现: `dishes_img.go` (line 79-86)
  - Success: 备注信息正确拼接并显示

- [x] 3.2 修改图片模板 - 整单打印

  - File: `main/app/printer/template/dishes_img.go` (line 431-442)
  - Purpose: 在整单打印时，将 `product.Remark` 改为拼接后的备注信息
  - Requirements: 1.3, 1.5-1.9
  - Leverage: Task 3.1 的实现逻辑
  - Success: 备注信息正确拼接并显示

- [x] 3.3 修改图片模板 - 退菜单打印

  - File: `main/app/printer/template/dishes_img.go` (line 615-619)
  - Purpose: 在退菜单打印时，将 `product.Remark` 改为拼接后的备注信息
  - Requirements: 1.3, 1.5-1.9
  - Leverage: Task 3.1 的实现逻辑
  - Success: 备注信息正确拼接并显示

- [x] 3.4 修改图片模板 - 出菜单打印

  - File: `main/app/printer/template/dishes_img.go` (line 817-821)
  - Purpose: 在出菜单打印时，将 `product.Remark` 改为拼接后的备注信息
  - Requirements: 1.3, 1.5-1.9
  - Leverage: Task 3.1 的实现逻辑
  - Success: 备注信息正确拼接并显示

- [x] 3.5 修改基础模板 - PrintCompleteOrderImgProducts 方法

  - File: `main/app/printer/template/base.go` (line 707-714)
  - Purpose: 在基础打印方法中，将 `product.Remark` 改为拼接后的备注信息
  - Requirements: 1.4, 1.5-1.9
  - Leverage: Task 3.1 的实现逻辑
  - Success: 备注信息正确拼接并显示

---

## Phase 4: 测试验证

- [ ] 4.1 单元测试 - CodeSoft 模板

  - File: `main/app/printer/template/dishes_codesoft_test.go` (如需要)
  - Purpose: 测试 CodeSoft 模板的备注打印逻辑
  - Requirements: 测试要求
  - Leverage: 现有测试文件（如有）
  - Success: 测试通过，覆盖所有修改位置

- [ ] 4.2 单元测试 - XPrinter 模板

  - File: `main/app/printer/template/dishes_xprinter_test.go` (如需要)
  - Purpose: 测试 XPrinter 模板的备注打印逻辑
  - Requirements: 测试要求
  - Leverage: 现有测试文件（如有）
  - Success: 测试通过，覆盖所有修改位置

- [ ] 4.3 单元测试 - 图片模板

  - File: `main/app/printer/template/dishes_img_test.go` (如需要)
  - Purpose: 测试图片模板的备注打印逻辑
  - Requirements: 测试要求
  - Leverage: 现有测试文件（如有）
  - Success: 测试通过，覆盖所有修改位置

- [ ] 4.4 集成测试 - 整单打印

  - File: -
  - Purpose: 测试整单打印场景下的备注显示
  - Requirements: 验收标准 4
  - Success: 整单打印时，备注信息正确显示

- [ ] 4.5 集成测试 - 一菜一单打印

  - File: -
  - Purpose: 测试一菜一单打印场景下的备注显示
  - Requirements: 验收标准 5
  - Success: 一菜一单打印时，备注信息正确显示

- [ ] 4.6 集成测试 - 退菜单打印

  - File: -
  - Purpose: 测试退菜单打印场景下的备注显示
  - Requirements: 验收标准 6
  - Success: 退菜单打印时，备注信息正确显示

- [ ] 4.7 集成测试 - 出菜单打印

  - File: -
  - Purpose: 测试出菜单打印场景下的备注显示
  - Requirements: 验收标准 7
  - Success: 出菜单打印时，备注信息正确显示

- [x] 4.8 手动测试 - 多语言支持

  - File: -
  - Purpose: 测试不同语言环境下的备注显示
  - Requirements: 验收标准 3
  - Success: 中文、英文、泰文、缅甸文等语言环境下，备注内容显示对应语言版本 ✅

- [x] 4.9 手动测试 - 向后兼容

  - File: -
  - Purpose: 测试没有预设备注的商品，保持原有打印行为
  - Requirements: 验收标准 8
  - Success: 没有预设备注的商品，仅显示自定义备注（保持原有行为） ✅

- [x] 4.10 手动测试 - 不同打印机类型

  - File: -
  - Purpose: 测试不同打印机类型下的备注显示
  - Requirements: 测试要求
  - Success: CodeSoft、XPrinter、图片打印等所有打印机类型下，备注信息正确显示 ✅

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 `go fmt` 和 `go vet`
- [ ] 所有修改的代码有中文注释
- [ ] 代码风格与现有打印模板保持一致

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成
- [ ] 所有打印场景已测试

### 文档同步

- [ ] 代码注释完整
- [ ] design.md 准确反映实现
- [ ] 如有问题或经验，更新 Graphiti Episode

### 规范遵循

- [ ] 遵循 `.cursor/rules/go-main.mdc`
- [ ] 遵循 `.cursor/rules/go-printer.mdc`
- [ ] 复用现有的 Model 方法
- [ ] 参考整单备注的实现方式

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/story-main-item-remark-preset-printing/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-main-item-remark-preset-printing/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-main-item-remark-preset-printing/tasks.md
```

### 执行流程

1. **选择任务**: 选择下一个未完成任务（建议按 Phase 顺序执行）
2. **阅读需求**: 查看 requirements.md 中的关联需求
3. **查看复用**: 检查 Leverage 中的可复用代码
4. **查看设计**: 参考 design.md 中的实现示例
5. **实现代码**: 按照设计文档中的示例实现
6. **运行检查**: `go fmt`, `go vet`
7. **测试验证**: 手动测试打印效果
8. **标记完成**: 将 `[ ]` 改为 `[x]`
9. **提交代码**: Git commit（参考 `.cursor/rules/version.mdc`）

---

## 附录：标准 Prompt 模板

### 打印模板修改

```
Role: Go Developer specializing in Printer Template

Task: 修改 {模板名称} 的 {打印场景} 逻辑，将 product.Remark 改为使用拼接后的备注信息（包含预设备注和自定义备注）

Context:
- Current file: {文件路径}
- Line range: {行号范围}
- Leverage code: 
  - main/app/model/sale_order_product.go - GetOrderItemRemark() (line 1197-1210)
  - main/app/model/sale_order_product.go - BuildOrderItemRemarkInfo() (line 1212-1261)
  - {参考实现路径} - 整单备注实现
- Requirements: 1.{需求编号}, 1.5-1.9
- Project specs: 遵循 .cursor/rules/go-main.mdc, .cursor/rules/go-printer.mdc

Implementation Steps:
1. 使用 product.GetOrderItemRemark() 获取预设备注列表
2. 使用 product.BuildOrderItemRemarkInfo(orderItemRemarkList, product.Remark) 构建备注信息
3. 使用 remarkInfo.Remark.GetLocale(t.base.Lang) 获取多语言内容
4. 将原来的 product.Remark 替换为 remarkText
5. 保持原有的行间距和字符大小设置逻辑

Restrictions:
- 保持与现有打印模板一致的代码风格
- 保持原有的行间距和字符大小设置
- 参考整单备注的实现方式
- 不使用 panic，保持错误处理逻辑

Success Criteria:
- 备注信息正确拼接（预设备注 + 自定义备注）
- 多语言支持正确（根据 t.base.Lang 显示对应语言）
- 代码通过 go fmt 和 go vet
- 打印效果与原有格式一致
```

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 在执行任务过程中若总结出经验或规避策略，请记录 Episode，并在 tasks.md 尾部更新 `Related Episode`。

---

**模板版本**: v1.0.0  
**最后更新**: 2025-12-11  
**维护者**: 后端开发组
