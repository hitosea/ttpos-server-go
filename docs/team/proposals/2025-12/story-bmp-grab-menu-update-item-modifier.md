# GrabFood 菜单项和修饰符更新功能 需求提案

> 本文档用于需求发起阶段，经团队评审后创建正式 Spec。

---

## 📋 提案信息

| 项目       | 内容     |
| ---------- | -------- |
| **提案人** | rikugun   |
| **日期**   | 2025-12-15   |
| **目标版本** | v2.1.0 |
| **状态**   | 待评审   |
| **关联任务** | - |
| **关联 Spec** | [docs/shared/specs/archived/v2.12/story-bmp-grab-menu-update-item-modifier/requirements.md](../../../../shared/specs/archived/v2.12/story-bmp-grab-menu-update-item-modifier/requirements.md) |

---

## 🎯 背景和动机

### 问题描述

目前 TTPOS BMP 模块中的 GrabFood 菜单服务仅实现了菜单的获取和同步功能，但在实际运营中，经常需要对菜单中的具体商品和修饰符进行实时更新操作，例如：

- 商品价格调整
- 商品可用状态变更
- 修饰符选项的增删改
- 商品库存状态更新

现有实现缺少对 GrabFood API `update-menu-record` 接口的对接，导致无法实时同步这些变更到 GrabFood 平台，影响外卖业务的运营效率。

### 业务价值

- 提升外卖菜单管理的实时性
- 减少因菜单不同步导致的订单错误
- 提高商户运营效率
- 增强与 GrabFood 平台的集成深度

### 目标用户

- [x] 商户管理员
- [ ] 收银员
- [ ] 厨房人员
- [ ] 顾客
- [ ] 其他: 运营人员

---

## 💡 解决方案概述

### 方案描述

在现有的 `ttpos-bmp/app/ttpos-takeout/internal/logic/grab_menu/grab_menu.go` 文件中增加对接 GrabFood API `update-menu-record` 接口的实现，将商品更新和修饰符更新分为两个独立方法：

1. `UpdateMenuItem()` - 处理单个商品的更新操作
2. `UpdateMenuModifier()` - 处理单个修饰符的更新操作

两个方法都基于 GrabFood API v1.1.3 的 update-menu-record 接口规范，通过调用第三方 SDK 实现与 GrabFood 平台的实时同步。

### 核心功能点

1. 商品信息更新（价格、状态、库存等）
2. 修饰符选项更新（价格、可用状态等）
3. 错误处理和状态回调
4. 菜单同步日志记录
5. 与现有菜单快照机制的集成

### 影响范围

**涉及终端**：
- [x] Shop 商家管理端

**涉及模块**：
- [ ] UI 组件
- [x] API 接口
- [x] 数据模型
- [x] 业务逻辑
- [x] 第三方集成
- [ ] 其他: ________

---

## 📊 初步评估

### 技术复杂度

- [ ] **低**：纯 UI 调整，无业务逻辑变更
- [x] **中**：需要前后端联调，基础业务逻辑
- [ ] **高**：涉及架构调整、第三方集成、复杂算法

### 工作量预估

[粗略估算，具体 SP 在技术方案评审时确定]

- **预计天数**: 5 天
- **预估 SP**: 8（待技术评审确认）

### 风险识别

**潜在风险**：
1. GrabFood API 接口变更导致兼容性问题
2. 网络超时导致的同步失败处理
3. 并发更新导致的数据一致性问题

**缓解措施**：
1. 基于官方 SDK 开发，确保 API 兼容性
2. 实现重试机制和错误恢复逻辑
3. 使用数据库事务确保数据一致性

---

## 🔗 相关资源

### 参考需求

- 现有菜单同步功能实现
- GrabFood API 文档: https://developer.grab.com/docs/grabfood/api/v1-1-3/#tag/update-menu-record/operation/update-menu

### 相关文档

- 技术预研文档: ttpos-bmp/.cursor/rules/go-rules.mdc
- API 设计规范: .cursor/rules/api.mdc

---

## 🤝 需求评审

### 评审参与人

| 角色         | 姓名   | 签名/日期 |
| ------------ | ------ | --------- |
| 产品经理     | rikugun |           |
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

- [ ] 创建 Spec：`story-bmp-grab-menu-update-item-modifier`
- [ ] 分配负责人：{姓名}
- [ ] 目标 Sprint：Sprint {N}

---

## 📝 附录

### User Story（初稿）

**作为** 商户管理员
**我想** 在 TTPOS 中更新商品或修饰符信息时，自动同步到 GrabFood 平台
**以便于** 确保外卖菜单信息的实时一致性

### AC 验收标准（初稿）

1. **WHEN** 商户在 TTPOS 中修改商品价格 **THEN** 系统 **SHALL** 调用 GrabFood API 更新对应商品价格
2. **WHEN** 商户在 TTPOS 中修改修饰符选项 **THEN** 系统 **SHALL** 调用 GrabFood API 更新对应修饰符信息
3. **IF** GrabFood API 调用失败 **THEN** 系统 **SHALL** 记录错误日志并支持重试机制

### 线框图/原型（可选）

[暂无]

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
**创建日期**: 2025-11-16  
**维护者**: 产品组 + Scrum Master  
**相关规范**: `.cursor/rules/scrum_story_point.mdc`, `.cursor/rules/specs.mdc`

