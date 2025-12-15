# grab-api-response-wrapper 需求提案

> 本文档用于需求发起阶段，经团队评审后创建正式 Spec。

---

## 📋 提案信息

| 项目       | 内容     |
| ---------- | -------- |
| **提案人** | rikugun   |
| **日期**   | 2025-12-15   |
| **目标版本** | v2.10.0 |
| **状态**   | 已创建 Spec   |
| **关联任务** | - |
| **关联 Spec** | [docs/shared/specs/active/story-takeout-grab-api-response-wrapper/requirements.md](../../../../shared/specs/active/story-takeout-grab-api-response-wrapper/requirements.md)      |

---

## 🎯 背景和动机

### 问题描述

当前 `ttpos-bmp/app/ttpos-takeout/manifest/protobuf/grab/grab.proto` 中的 `CreateSelfServeJourney` 和 `GetShopProviderCfg` 方法返回值使用了自定义的响应消息类型，与项目中其他服务（如 `menu.proto`）的统一响应格式不一致。

- `CreateSelfServeJourney` 返回 `CreateSelfServeJourneyResp`
- `GetShopProviderCfg` 返回 `GetShopProviderCfgResp`

而 `menu.proto` 中的方法都统一使用 `takeout.ApiResponse` 作为返回值包装器。

### 业务价值

- **提升 API 一致性**：统一响应格式，便于前端和客户端统一处理
- **简化错误处理**：使用标准化的 ApiResponse 可以统一处理成功/失败状态
- **维护效率提升**：减少重复的响应消息定义，降低维护成本
- **开发体验改善**：前端开发者无需适配不同的响应格式

### 目标用户

- [x] 商户管理员（需要配置外卖渠道集成）
- [x] 技术开发团队（需要维护 API 接口）
- [ ] 其他: 前端开发团队

---

## 💡 解决方案概述

### 方案描述

修改 `grab.proto` 文件中的两个 gRPC 方法，将返回值从自定义响应消息改为使用统一的 `takeout.ApiResponse` 包装器。

具体修改：
1. `CreateSelfServeJourney` 方法返回值从 `CreateSelfServeJourneyResp` 改为 `takeout.ApiResponse`
2. `GetShopProviderCfg` 方法返回值从 `GetShopProviderCfgResp` 改为 `takeout.ApiResponse`
3. 在响应消息中通过 `data` 字段返回具体的数据内容
4. 保持请求消息不变

参考 `menu.proto` 的实现方式，确保 API 响应格式的一致性。

### 核心功能点

1. 修改 `CreateSelfServeJourney` 方法返回值类型
2. 修改 `GetShopProviderCfg` 方法返回值类型
3. 更新对应的服务实现逻辑，适配新的响应格式
4. 重新生成 protobuf Go 代码
5. 验证前后端集成正常工作

### 影响范围

**涉及终端**：
- [ ] POS 收银端
- [ ] Shop 商家管理端
- [ ] KDS 厨显端
- [ ] QDS 排号叫号端
- [ ] Assistant 助手端
- [ ] Tablet 平板端
- [ ] Mobile 扫码端
- [ ] Menu 电子菜单端
- [ ] Member 会员端

**涉及模块**：
- [x] API 接口（protobuf 定义）
- [x] 业务逻辑（服务实现）
- [ ] UI 组件
- [ ] 数据模型
- [ ] 第三方集成

---

## 📊 初步评估

### 技术复杂度

- [x] **低**：纯接口定义调整，无业务逻辑变更
- [ ] **中**：需要前后端联调，基础业务逻辑
- [ ] **高**：涉及架构调整、第三方集成、复杂算法

### 工作量预估

[粗略估算，具体 SP 在技术方案评审时确定]

- **预计天数**: 1 天
- **预估 SP**: 2（待技术评审确认）

### 风险识别

**潜在风险**：
1. 前端客户端需要适配新的响应格式
2. 现有集成测试可能需要更新

**缓解措施**：
1. 提供详细的 API 变更说明文档
2. 编写迁移指南，帮助前端团队适配

---

## 🔗 相关资源

### 参考需求

- 类似功能: `ttpos-bmp/app/ttpos-takeout/manifest/protobuf/menu/menu.proto` 中的统一响应格式

### 相关文档

- 产品需求文档 (PRD): -
- 用户调研报告: -
- 技术预研文档: `ttpos-bmp/.cursor/rules/proto-rules.mdc`

---

## 🤝 需求评审

### 评审参与人

| 角色         | 姓名   | 签名/日期 |
| ------------ | ------ | --------- |
| 产品经理     | rikugun |           |
| 技术负责人   |        |           |
| 开发代表     |        |           |
| 测试代表     |        |           |
| UI/UX 设计师 |        |           |

### 评审结论

- [ ] ✅ **批准**：进入技术方案设计阶段
- [ ] 🔄 **修改后批准**：需补充以下内容
- [ ] ❌ **拒绝**：不符合产品规划或优先级

**评审意见**：

```
[记录评审会议的关键讨论和决策]
```

**下一步行动**：

- [ ] 创建 Spec：`story-takeout-grab-api-response-wrapper`
- [ ] 分配负责人：rikugun
- [ ] 目标 Sprint：Sprint 当前

---

## 📝 附录

### User Story（初稿）

**作为** 技术开发人员
**我想** 统一 Grab 服务 API 的响应格式
**以便于** 提高代码维护性和 API 一致性

### AC 验收标准（初稿）

1. **WHEN** 调用 `CreateSelfServeJourney` 接口 **THEN** 系统 **SHALL** 返回 `takeout.ApiResponse` 格式的响应
2. **WHEN** 调用 `GetShopProviderCfg` 接口 **THEN** 系统 **SHALL** 返回 `takeout.ApiResponse` 格式的响应
3. **IF** 接口调用成功 **THEN** 系统 **SHALL** 在 `data` 字段中返回具体的业务数据
4. **IF** 接口调用失败 **THEN** 系统 **SHALL** 在 `code` 和 `message` 字段中返回错误信息

### 线框图/原型（可选）

[无 UI 变更，无需线框图]

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
