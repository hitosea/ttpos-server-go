# 单品备注预设备注打印 需求提案

> 本文档用于需求发起阶段，经团队评审后创建正式 Spec。

---

## 📋 提案信息

| 项目       | 内容     |
| ---------- | -------- |
| **提案人** | 王昱   |
| **日期**   | 2025-12-11   |
| **目标版本** | {版本号} |
| **状态**   | 已创建 Spec   |
| **关联任务** | - |
| **关联 Spec** | [story-main-item-remark-preset-printing](../../../shared/specs/archived/v2.12/story-main-item-remark-preset-printing/requirements.md) |

---

## 🎯 背景和动机

### 问题描述

当前单品备注打印功能已实现，但仅打印手动输入的自定义备注内容，未包含单品预设备注信息。在实际业务场景中，商户会为菜品设置预设备注（如"少盐"、"不要香菜"、"微辣"等），这些预设备注在小票打印时也需要显示，以便后厨人员准确理解制作要求。

**当前痛点**：
- 小票打印时只显示手动输入的自定义备注，不显示预设备注
- 后厨人员无法看到完整的备注信息，可能导致制作错误
- 预设备注的多语言信息未在小票上体现
- 影响整单打印、一菜一单、退菜单、出菜单等多种打印场景

### 业务价值

- 提升后厨制作准确性，减少因备注信息不完整导致的错误
- 完整展示备注信息，提升服务质量
- 支持多语言备注打印，满足国际化需求
- 统一备注展示逻辑，与整单备注保持一致

### 目标用户

- [ ] 收银员
- [ ] 商户管理员
- [x] 厨房人员
- [ ] 顾客
- [x] 其他: 点餐助手操作员、H5/平板用户

---

## 💡 解决方案概述

### 方案描述

在小票打印时，将单品预设备注与自定义备注拼接后一起打印。参考整单备注的实现方式，使用多语言支持，确保在不同语言环境下都能正确显示备注内容。

**实现要点**：
1. 在打印模板中获取订单商品的预设备注列表
2. 将预设备注的多语言内容与自定义备注拼接
3. 使用与整单备注相同的多语言处理逻辑
4. 支持所有打印场景：整单打印、一菜一单、退菜单、出菜单

### 核心功能点

1. **预设备注获取**：从订单商品中获取关联的预设备注列表
2. **备注拼接**：将预设备注的多语言内容与自定义备注拼接
3. **多语言支持**：根据打印语言环境显示对应的备注内容
4. **打印场景覆盖**：支持整单打印、一菜一单、退菜单、出菜单等所有打印场景

### 影响范围

**涉及终端**：
- [x] POS 收银端
- [ ] Shop 商家管理端
- [ ] KDS 厨显端
- [ ] QDS 排号叫号端
- [x] Assistant 助手端
- [x] Tablet 平板端
- [x] Mobile 扫码端
- [ ] Menu 电子菜单端
- [ ] Member 会员端

**涉及模块**：
- [ ] UI 组件
- [ ] API 接口
- [ ] 数据模型
- [x] 业务逻辑
- [ ] 第三方集成
- [x] 其他: 打印模板（CodeSoft、XPrinter、图片模板等）

---

## 📊 初步评估

### 技术复杂度

- [ ] **低**：纯 UI 调整，无业务逻辑变更
- [x] **中**：需要前后端联调，基础业务逻辑
- [ ] **高**：涉及架构调整、第三方集成、复杂算法

### 工作量预估

[粗略估算，具体 SP 在技术方案评审时确定]

- **预计天数**: 3-5 天
- **预估 SP**: 5 SP（待技术评审确认）

**工作量分解**：
- 打印模板修改（CodeSoft、XPrinter、图片模板等）：2-3 天
- 多语言处理逻辑：1 天
- 测试验证（各种打印场景）：1-2 天

### 风险识别

**潜在风险**：
1. 打印模板较多，需要逐一修改和测试
2. 多语言拼接逻辑需要与整单备注保持一致
3. 不同打印机的字符宽度限制可能影响显示效果

**缓解措施**：
1. 参考整单备注的实现方式，复用现有多语言处理逻辑
2. 使用 `BuildOrderItemRemarkInfo` 方法统一构建备注信息
3. 在测试阶段覆盖所有打印场景和打印机类型

---

## 🔗 相关资源

### 参考需求

- 类似功能: [story-order-item-remark-preset](../../../shared/specs/archived/v2.12/story-order-item-remark-preset/requirements.md) - 单品备注预设功能
- 参考实现: 整单备注打印逻辑
  - `main/app/printer/template/dishes_codesoft.go` - CodeSoft 模板
  - `main/app/printer/template/dishes_xprinter.go` - XPrinter 模板
  - `main/app/printer/template/dishes_img.go` - 图片模板

### 相关文档

- 打印模块规范: `.cursor/rules/go-printer.mdc`
- 单品备注预设提案: `docs/team/proposals/2025-12/item-remark-preset.md`
- 数据模型: `main/app/model/sale_order_product.go` - `BuildOrderItemRemarkInfo()` 方法

---

## 🤝 需求评审

### 评审参与人

| 角色         | 姓名   | 签名/日期 |
| ------------ | ------ | --------- |
| 产品经理     | {姓名} |           |
| 技术负责人   | {姓名} |           |
| 开发代表     | {姓名} |           |
| 测试代表     | {姓名} |           |
| UI/UX 设计师 | {姓名} |           |

### 评审结论

- [ ] ✅ **批准**：进入技术方案设计阶段
- [ ] 🔄 **修改后批准**：需补充以下内容
- [ ] ❌ **拒绝**：不符合产品规划或优先级

**评审意见**：

```
[记录评审会议的关键讨论和决策]
```

**下一步行动**：

- [ ] 创建 Spec：`story-main-item-remark-preset-printing`
- [ ] 分配负责人：{姓名}
- [ ] 目标 Sprint：Sprint {N}

---

## 📝 附录

### User Story（初稿）

**作为** 后厨人员  
**我想** 在小票上看到完整的单品备注信息（包括预设备注和自定义备注）  
**以便于** 准确理解制作要求，减少制作错误

### AC 验收标准（初稿）

1. **WHEN** 订单商品包含预设备注 **THEN** 小票打印时 **SHALL** 显示预设备注的多语言内容
2. **WHEN** 订单商品同时包含预设备注和自定义备注 **THEN** 小票打印时 **SHALL** 将两者拼接显示
3. **WHEN** 打印语言为不同语言（中文、英文、泰文、缅甸文等） **THEN** 备注内容 **SHALL** 显示对应语言版本
4. **WHEN** 执行整单打印、一菜一单、退菜单、出菜单等打印操作 **THEN** 备注信息 **SHALL** 在所有打印场景中正确显示
5. **IF** 订单商品没有预设备注 **THEN** 打印逻辑 **SHALL** 保持原有行为（仅显示自定义备注）

### 技术实现要点

1. **参考整单备注实现**：
   - 整单备注使用 `order.GetLatestOrderRemarkRes()` 获取备注
   - 使用 `orderRemark.Remark.GetLocale(t.base.Lang)` 获取多语言内容
   - 格式：`t.base.Translate("整单备注") + ": " + orderRemark.Remark.GetLocale(t.base.Lang)`

2. **单品备注实现方式**：
   - 使用 `product.GetOrderItemRemark()` 获取预设备注列表
   - 使用 `product.BuildOrderItemRemarkInfo()` 构建备注信息（已实现）
   - 拼接预设备注和自定义备注：`预设1;预设2;自定义备注`

3. **打印模板修改位置**：
   - `main/app/printer/template/dishes_codesoft.go` - 多处 `product.Remark` 打印位置
   - `main/app/printer/template/dishes_xprinter.go` - 多处 `product.Remark` 打印位置
   - `main/app/printer/template/dishes_img.go` - 多处 `product.Remark` 打印位置
   - `main/app/printer/template/base.go` - `PrintCompleteOrderImgProducts` 方法

### 线框图/原型（可选）

[附加 UI 线框图或原型链接]

---

## 📄 模板使用说明

### 何时使用此模板

- ✅ 产品经理提出新功能想法
- ✅ 用户反馈需求建议
- ✅ 技术团队提出改进方案
- ✅ 需要团队讨论和评审的需求

### 与 Spec 的区别

| 阶段        | 文档类型      | 详细程度 | 用途               |
| ----------- | ------------- | -------- | ------------------ |
| **需求发起** | Proposal      | 粗略     | 团队评审、决策是否做 |
| **需求确认** | Requirements  | 详细     | User Story + AC，开发依据 |
| **技术设计** | Design        | 详细     | 技术方案，实现指导 |
| **任务分解** | Tasks         | 详细     | 开发执行，进度追踪 |

### 流转路径

```
提案 (Proposal) 
  ↓ 评审批准
需求文档 (Requirements) 
  ↓ 技术评审
设计文档 (Design) 
  ↓ SP 评估 ≤ 5
任务分解 (Tasks)
  ↓
开发实现
```

---

**版本**: v1.0.0  
**创建日期**: 2025-12-11  
**维护者**: 产品组 + Scrum Master  
**相关规范**: `.cursor/rules/scrum_story_point.mdc`, `.cursor/rules/specs.mdc`, `.cursor/rules/go-printer.mdc`
