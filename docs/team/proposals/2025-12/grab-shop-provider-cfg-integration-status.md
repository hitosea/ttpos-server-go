# Grab 店铺集成状态落库与旅程记录 需求提案

> 本文档用于需求发起阶段，经团队评审后创建正式 Spec。

---

## 📋 提案信息

| 项目       | 内容     |
| ---------- | -------- |
| **提案人** | rikugun   |
| **日期**   | 2025-12-11   |
| **目标版本** | vNext |
| **状态**   | 待评审   |
| **关联任务** | - |
| **关联 Spec** | [story-takeout-grab-shop-provider-cfg](../../shared/specs/active/story-takeout-grab-shop-provider-cfg/requirements.md)      |

---

## 🎯 背景和动机

### 问题描述

- Grab webhook 仅在内存/流程中处理 integrationStatus，未落库，无法追踪历史与排查。
- CreateSelfServeJourney 创建时未与门店配置表关联，后续状态查询无据可依。
- 缺少统一的 shop_provider_cfg 表存储第三方状态，导致不同入口状态不一致。

### 业务价值

- 将 Grab 集成状态标准化落库，便于运营和技术排查。
- 支持后续看板/监控对门店集成健康度的统计与预警。
- 减少人工校对，提升门店上线与异常恢复效率。

### 目标用户

- [ ] 收银员
- [x] 商户管理员
- [ ] 厨房人员
- [ ] 顾客
- [ ] 其他: 运营/技术支持

---

## 💡 解决方案概述

### 方案描述

- 新增 `shop_provider_cfg` 表，记录门店与第三方（Grab 等）的集成配置与状态，包括状态枚举 `INACTIVE/ACTIVE/SYNCING/FAILED` 以及基础主键、UUID、时间戳字段。
- 在 Grab 通知回调处理 integrationStatus 时，按 shop_uuid + provider_name 定位记录，更新 `provider_shop_status`，未存在则按必要字段插入。
- 在 `CreateSelfServeJourney` 流程中，当创建 Grab 店铺自助旅程成功后，写入/更新 `shop_provider_cfg` 以保证状态源头一致。

### 核心功能点

1. DDL：创建 `shop_provider_cfg`（id, uuid, shop_uuid, provider_name, provider_merchant_id, provider_shop_status enum，创建/更新/删除时间）。
2. Webhook：Grab integrationStatus 通知映射到 `provider_shop_status`（含失败/同步中/激活/未激活）。
3. 业务流程：CreateSelfServeJourney 成功后落库/更新 `shop_provider_cfg`，保持状态闭环。

### 影响范围

**涉及终端**：
- [ ] POS 收银端
- [x] Shop 商家管理端
- [ ] KDS 厨显端
- [ ] QDS 排号叫号端
- [ ] Assistant 助手端
- [ ] Tablet 平板端
- [ ] Mobile 扫码端
- [ ] Menu 电子菜单端
- [ ] Member 会员端

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

- **预计天数**: 3 天
- **预估 SP**: 5（待技术评审确认）

### 风险识别

**潜在风险**：
1. Grab integrationStatus 与内部枚举映射不一致导致状态漂移。
2. 同步/失败状态重复通知可能引发并发更新冲突。
3. 现有查询/缓存未落库读取，新增表后需补齐读取逻辑。

**缓解措施**：
1. 定义明确的状态映射表，落在代码常量与文档。
2. 更新语句使用乐观检查/幂等逻辑（按 shop_uuid + provider_name upsert）。
3. 在设计/开发阶段梳理读路径，必要时补充缓存刷新或查询接口。

---

## 🔗 相关资源

### 参考需求

- Grab 官方 integrationStatus webhook 文档
- 现有 Grab 外卖集成相关 Spec：`docs/shared/specs/active/task-takeout-push-grab-menu-webhook/requirements.md`

### 相关文档

- 产品需求文档 (PRD): -
- 用户调研报告: -
- 技术预研文档: -

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

- [ ] 创建 Spec：`story-{app}-{feature}`
- [ ] 分配负责人：{姓名}
- [ ] 目标 Sprint：Sprint {N}

---

## 📝 附录

### User Story（初稿）

**作为** 商户管理员/运营  
**我想** 在 Grab 集成创建或状态变化时有可靠的门店状态记录  
**以便于** 及时了解集成健康度并快速排查问题

### AC 验收标准（初稿）

1. **WHEN** Grab 推送 integrationStatus=ACTIVE **THEN** 系统将对应门店 `provider_shop_status` 更新为 ACTIVE 并可查询。
2. **WHEN** integrationStatus=FAILED 且包含错误信息 **THEN** 系统落库 FAILED 状态，保留更新时间；重复推送不产生重复记录。
3. **WHEN** CreateSelfServeJourney 创建成功 **THEN** 系统写入/更新 `shop_provider_cfg`，默认状态为 SYNCING 或 ACTIVE（按业务定义）。

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
**创建日期**: 2025-11-16  
**维护者**: 产品组 + Scrum Master  
**相关规范**: `.cursor/rules/scrum_story_point.mdc`, `.cursor/rules/specs.mdc`
