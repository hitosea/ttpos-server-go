> ⚠️ **已归档** - 此 Spec 已随 v2.12 发布。
>
> - 归档时间: 2026-01-04
> - 归档人: weifashi

# 单品备注预设备注打印 需求文档

> 本文档定义单品备注预设备注打印功能的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2025-12/item-remark-preset-printing.md](../../../../team/proposals/2025-12/item-remark-preset-printing.md) |
| **创建日期**      | 2025-12-11                                                                                                 |
| **负责人**        | {姓名}                                                                                                       |
| **目标 Sprint**   | Sprint {N}                                                                                                   |
| **涉及技术栈**    | [x] Go (main/) [ ] Go (ttpos-bmp/) [ ] PHP (admin/) [ ] Vue (admin/views/)                                   |

## 📋 审核状态

| 项目         | 内容                     |
| ------------ | ------------------------ |
| **审核状态** | 待审核 |
| **审核人**   | {审核人姓名}             |
| **审核日期** | {YYYY-MM-DD}             |
| **审核意见** | {审核意见，如有}         |

---

## 📋 概述

在小票打印时，将单品预设备注与自定义备注拼接后一起打印。参考整单备注的实现方式，使用多语言支持，确保在不同语言环境下都能正确显示备注内容。

**功能范围**：
- ✅ 在打印模板中获取订单商品的预设备注列表
- ✅ 将预设备注的多语言内容与自定义备注拼接
- ✅ 使用与整单备注相同的多语言处理逻辑
- ✅ 支持所有打印场景：整单打印、一菜一单、退菜单、出菜单
- ✅ 支持所有打印模板：CodeSoft、XPrinter、图片模板等

**参考实现**：整单备注打印逻辑（`order.GetLatestOrderRemarkRes()` 和 `orderRemark.Remark.GetLocale(t.base.Lang)`）

## 🎯 产品对齐

支持后厨人员在小票上看到完整的单品备注信息（包括预设备注和自定义备注），准确理解制作要求，减少制作错误。提升后厨制作准确性，完整展示备注信息，支持多语言备注打印，满足国际化需求。

## 📝 用户故事

**作为** 后厨人员  
**我想** 在小票上看到完整的单品备注信息（包括预设备注和自定义备注）  
**以便于** 准确理解制作要求，减少制作错误

---

## 功能需求

### Requirement 1: 打印模板中拼接单品预设备注和自定义备注

**用户故事**: 作为后厨人员，我想在小票上看到完整的单品备注信息，以便于准确理解制作要求

#### 验收标准

1. **WHEN** 订单商品包含预设备注 **THEN** 小票打印时 **SHALL** 显示预设备注的多语言内容

2. **WHEN** 订单商品同时包含预设备注和自定义备注 **THEN** 小票打印时 **SHALL** 将两者拼接显示（格式：`预设1;预设2;自定义备注`）

3. **WHEN** 打印语言为不同语言（中文、英文、泰文、缅甸文等） **THEN** 备注内容 **SHALL** 显示对应语言版本

4. **WHEN** 执行整单打印 **THEN** 备注信息 **SHALL** 正确显示

5. **WHEN** 执行一菜一单打印 **THEN** 备注信息 **SHALL** 正确显示

6. **WHEN** 执行退菜单打印 **THEN** 备注信息 **SHALL** 正确显示

7. **WHEN** 执行出菜单打印 **THEN** 备注信息 **SHALL** 正确显示

8. **IF** 订单商品没有预设备注 **THEN** 打印逻辑 **SHALL** 保持原有行为（仅显示自定义备注）

9. **IF** 订单商品没有自定义备注 **THEN** 打印逻辑 **SHALL** 仅显示预设备注

10. **IF** 订单商品既没有预设备注也没有自定义备注 **THEN** 打印逻辑 **SHALL** 不显示备注信息

#### 具体要求

- [ ] 1.1 修改 CodeSoft 打印模板（`dishes_codesoft.go`），在所有打印 `product.Remark` 的位置，改为打印拼接后的备注信息
- [ ] 1.2 修改 XPrinter 打印模板（`dishes_xprinter.go`），在所有打印 `product.Remark` 的位置，改为打印拼接后的备注信息
- [ ] 1.3 修改图片打印模板（`dishes_img.go`），在所有打印 `product.Remark` 的位置，改为打印拼接后的备注信息
- [ ] 1.4 修改基础打印模板（`base.go`），在 `PrintCompleteOrderImgProducts` 方法中，改为打印拼接后的备注信息
- [ ] 1.5 使用 `product.GetOrderItemRemark()` 获取预设备注列表
- [ ] 1.6 使用 `product.BuildOrderItemRemarkInfo()` 构建备注信息（已实现，包含多语言支持）
- [ ] 1.7 根据打印语言环境（`t.base.Lang`）获取对应语言的备注内容
- [ ] 1.8 拼接格式：预设备注用分号分隔，最后拼接自定义备注（如：`预设1;预设2;自定义备注`）
- [ ] 1.9 保持与整单备注相同的多语言处理逻辑和显示格式

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 严格遵循打印模板层的修改，不涉及 Service/Repository 层
- **单一职责原则**: 每个打印模板文件负责各自的打印逻辑
- **模块化设计**: 复用现有的 `BuildOrderItemRemarkInfo` 方法
- **依赖管理**: 打印模板依赖 Model 层的方法
- **遵循规范**:
  - `.cursor/rules/go-main.mdc` - Go Main 开发规范
  - `.cursor/rules/go-printer.mdc` - 打印模块开发规范
  - `.cursor/rules/structs.mdc` - 项目结构规范

### API 设计要求

- [ ] 不涉及 API 接口（纯打印模板修改）

### 数据库设计要求

- [ ] 不涉及数据库变更（数据模型已存在）

### 性能要求

- [ ] 打印性能不受影响（仅修改显示内容，不增加查询）
- [ ] 备注信息拼接逻辑高效（使用已实现的 `BuildOrderItemRemarkInfo` 方法）

### 测试要求

- [ ] 单元测试：打印模板逻辑测试
- [ ] 集成测试：端到端打印测试（整单打印、一菜一单、退菜单、出菜单）
- [ ] 手动测试：不同打印机类型测试（CodeSoft、XPrinter、图片打印）
- [ ] 多语言测试：不同语言环境下的备注显示测试
- [ ] 参考: `.cursor/rules/go-main.mdc` - 测试规范

### 国际化要求

- [x] 支持 10 种语言（中文、英文、日语、韩语、泰文、缅甸文等）
- [x] 备注内容使用多语言实现（通过 `BuildOrderItemRemarkInfo` 方法）
- [x] 参考: `main/i18n/` - 国际化配置

### 安全要求

- [ ] 不涉及安全相关功能

### 可靠性要求

- [ ] 错误处理：如果预设备注获取失败，降级为仅显示自定义备注
- [ ] 向后兼容：如果订单商品没有预设备注，保持原有打印行为

---

## 验收标准

### 功能验收

1. **整单打印**: 订单商品包含预设备注时，小票上显示拼接后的备注信息
2. **一菜一单打印**: 每个商品的备注信息正确显示
3. **退菜单打印**: 退菜商品的备注信息正确显示
4. **出菜单打印**: 出菜商品的备注信息正确显示
5. **多语言支持**: 不同语言环境下，备注内容显示对应语言版本
6. **拼接格式**: 预设备注和自定义备注正确拼接（格式：`预设1;预设2;自定义备注`）
7. **向后兼容**: 没有预设备注的订单商品，保持原有打印行为

### 测试验收

1. **单元测试**: 打印模板逻辑测试通过
2. **集成测试**: 所有打印场景测试通过
3. **手动测试**: 不同打印机类型测试通过
4. **多语言测试**: 不同语言环境测试通过

### 文档验收

1. **代码注释**: 所有修改的代码有中文注释
2. **技术文档**: design.md 完整且准确（待创建）

---

## 约束条件

### 技术约束

#### Go Main 模块

- 必须使用 Gin 框架（打印模板层）
- 复用现有的 `BuildOrderItemRemarkInfo` 方法
- 参考整单备注的实现方式
- 不使用 panic，返回 error（如果涉及错误处理）

### 业务约束

- 必须与整单备注的多语言处理逻辑保持一致
- 必须支持所有打印场景（整单打印、一菜一单、退菜单、出菜单）
- 必须支持所有打印模板（CodeSoft、XPrinter、图片模板等）

### 资源约束

- 开发时间: 3-5 天
- Story Point: 5 SP

---

## 依赖关系

### 技术依赖

- `main/app/model/sale_order_product.go` - `GetOrderItemRemark()` 方法（已存在）
- `main/app/model/sale_order_product.go` - `BuildOrderItemRemarkInfo()` 方法（已存在）
- `main/app/printer/template/dishes_codesoft.go` - CodeSoft 打印模板
- `main/app/printer/template/dishes_xprinter.go` - XPrinter 打印模板
- `main/app/printer/template/dishes_img.go` - 图片打印模板
- `main/app/printer/template/base.go` - 基础打印模板

### 服务依赖

- 无外部服务依赖

### 业务依赖

- 依赖单品备注预设功能（`story-order-item-remark-preset`）
- 依赖订单商品备注预设关联功能（`story-main-order-item-remark-reason-management`）

---

## 风险和缓解

### 风险 1: 打印模板较多，需要逐一修改和测试

**影响**: 中  
**概率**: 高  
**缓解措施**:
- 统一修改逻辑，使用相同的方法获取和拼接备注信息
- 先修改一个模板，验证逻辑正确后再修改其他模板
- 建立测试清单，确保所有模板都测试覆盖

### 风险 2: 多语言拼接逻辑需要与整单备注保持一致

**影响**: 中  
**概率**: 中  
**缓解措施**:
- 复用现有的 `BuildOrderItemRemarkInfo` 方法（已实现多语言支持）
- 参考整单备注的实现方式（`orderRemark.Remark.GetLocale(t.base.Lang)`）
- 在测试阶段验证多语言显示正确性

### 风险 3: 不同打印机的字符宽度限制可能影响显示效果

**影响**: 低  
**概率**: 低  
**缓解措施**:
- 保持与现有备注打印相同的格式和行间距设置
- 在测试阶段覆盖不同打印机类型
- 如果出现显示问题，可以调整行间距或换行逻辑

---

## 时间表

- **Phase 1 - 打印模板修改**: 2-3 天
  - CodeSoft 模板修改
  - XPrinter 模板修改
  - 图片模板修改
  - 基础模板修改
- **Phase 2 - 多语言处理**: 1 天
  - 验证多语言逻辑
  - 测试不同语言环境
- **Phase 3 - 测试验证**: 1-2 天
  - 单元测试
  - 集成测试
  - 手动测试（所有打印场景和打印机类型）
- **总计**: 3-5 天（SP = 5）

---

## 参考资料

### 核心规范

- `.cursor/rules/go-main.mdc` - Go Main 核心约束
- `.cursor/rules/go-printer.mdc` - 打印模块开发规范
- `.cursor/rules/structs.mdc` - 项目结构规范

### 参考实现

- **整单备注打印**: `main/app/printer/template/dishes_codesoft.go` (line 629-636)
- **整单备注打印**: `main/app/printer/template/dishes_xprinter.go` (line 719-726)
- **整单备注打印**: `main/app/printer/template/dishes_img.go` (line 79-86)
- **备注信息构建**: `main/app/model/sale_order_product.go` - `BuildOrderItemRemarkInfo()` (line 1212-1261)

### 相关文档

- 单品备注预设提案: `docs/team/proposals/2025-12/item-remark-preset.md`
- 单品备注预设 Spec: `docs/shared/specs/active/story-order-item-remark-preset/requirements.md`
- 订单商品备注预设关联 Spec: `docs/shared/specs/active/story-main-order-item-remark-reason-management/requirements.md`

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 提醒：需求评审或范围调整若形成经验，应同步更新 Episode 并在 Proposal/Spec 互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-12-11  
**作者**: 王昱  
**审核者**: {审核者}
