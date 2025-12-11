# 叫号系统返回配置信息 需求提案

> 本文档用于需求发起阶段，经团队评审后创建正式 Spec。

---

## 📋 提案信息

| 项目       | 内容     |
| ---------- | -------- |
| **提案人** | 王昱   |
| **日期**   | 2025-12-11   |
| **目标版本** | - |
| **状态**   | 待评审   |
| **关联任务** | - |
| **关联 Spec** | [story-callboard-data-config](../../../shared/specs/active/story-callboard-data-config/requirements.md)      |

---

## 🎯 背景和动机

### 问题描述

当前 `/callboard/data` 接口只返回队列数据（等待队列和取餐队列），但叫号展示设备端还需要获取配置信息（如系统名称、背景图片、超时限制、语音叫号开关、叫号次数等）来正确展示界面和功能。

目前这些配置信息只在商家管理端的设备列表接口 `/shop/callboard/device/list` 中返回，设备端无法获取，导致设备端需要额外调用其他接口或无法获取完整的配置信息。

**示例**：
> 叫号展示设备在展示队列数据时，需要根据配置的背景图片、系统名称等信息来渲染界面，但目前 `/callboard/data` 接口无法提供这些配置信息。

### 业务价值

- 减少设备端接口调用次数，提升性能
- 简化设备端逻辑，统一从 `/callboard/data` 接口获取所需数据
- 提升用户体验，确保设备端能正确展示配置的界面和功能
- 降低前后端耦合度，配置信息统一管理

### 目标用户

- [ ] 收银员
- [ ] 商户管理员
- [ ] 厨房人员
- [ ] 顾客
- [x] 其他: 叫号展示设备端

---

## 💡 解决方案概述

### 方案描述

扩展 `/callboard/data` 接口的响应结构，在现有队列数据基础上，增加返回叫号系统配置信息字段：
- `name`: 叫号系统名称
- `background_image_url`: 背景图片 URL
- `timeout_limit`: 超时限制（分钟）
- `voice_call_enabled`: 语音叫号开关
- `call_count`: 叫号次数

这些配置信息已存储在 Redis 的 `DeviceBindInfo` 中，只需在 `GetQueueData` 方法中读取并返回即可。

### 核心功能点

1. 扩展 `QueueDataResp` 响应结构，新增配置信息字段（必返字段）
2. 修改 `GetQueueData` 服务方法，从 Redis 读取配置信息并填充到响应中
3. 配置信息缺失时使用默认值，确保字段始终返回

### 影响范围

**涉及终端**：
- [ ] POS 收银端
- [ ] Shop 商家管理端
- [ ] KDS 厨显端
- [x] QDS 排号叫号端
- [ ] Assistant 助手端
- [ ] Tablet 平板端
- [ ] Mobile 扫码端
- [ ] Menu 电子菜单端
- [ ] Member 会员端

**涉及模块**：
- [ ] UI 组件
- [x] API 接口
- [ ] 数据模型
- [x] 业务逻辑
- [ ] 第三方集成
- [ ] 其他: ________

---

## 📊 初步评估

### 技术复杂度

- [ ] **低**：纯 UI 调整，无业务逻辑变更
- [x] **中**：需要前后端联调，基础业务逻辑
- [ ] **高**：涉及架构调整、第三方集成、复杂算法

### 工作量预估

[粗略估算，具体 SP 在技术方案评审时确定]

- **预计天数**: 0.5 天
- **预估 SP**: 2（待技术评审确认）

### 风险识别

**潜在风险**：
1. 接口响应结构变更可能影响现有设备端代码
2. 配置信息读取失败时的降级处理

**缓解措施**：
1. 新增字段为可选字段，保持向后兼容
2. 配置信息读取失败时使用默认值，不影响队列数据返回

---

## 🔗 相关资源

### 参考需求

- 类似功能: `/shop/callboard/device/list` 接口已返回配置信息
- 相关 Spec: `story-shop-callboard-settings`（叫号系统配置管理）

### 相关文档

- API 文档: `/callboard/data` 接口文档
- 代码位置: `main/app/api/v1/callboard/handler.go`, `main/app/service/callboard/service.go`

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

- [ ] 创建 Spec：`story-callboard-data-config`
- [ ] 分配负责人：{姓名}
- [ ] 目标 Sprint：Sprint {N}

---

## 📝 附录

### User Story（初稿）

**作为** 叫号展示设备端  
**我想** 从 `/callboard/data` 接口获取配置信息（系统名称、背景图片、超时限制等）  
**以便于** 正确展示界面和功能，无需额外调用其他接口

### AC 验收标准（初稿）

1. **WHEN** 设备端调用 `/callboard/data` 接口 **THEN** 响应中 **SHALL** 必须包含 `name`、`background_image_url`、`timeout_limit`、`voice_call_enabled`、`call_count` 字段（必返字段）
2. **IF** 配置信息不存在 **THEN** 系统 **SHALL** 返回默认值（name 默认为 "WALLACE"，background_image_url 默认为空字符串，timeout_limit 默认为 0，voice_call_enabled 默认为 false，call_count 默认为 1）
3. **IF** 配置信息读取失败 **THEN** 系统 **SHALL** 仍返回队列数据，配置字段使用默认值

### 技术实现要点

1. **响应结构扩展**：
   ```go
   type QueueDataResp struct {
       Lang1              string   `json:"lang1"`
       Lang2              string   `json:"lang2"`
       UpdateTime         int64    `json:"update_time"`
       PreparingQueue     []string `json:"preparing_queue"`
       PreparedQueue      []string `json:"prepared_queue"`
       // 新增配置字段（必返）
       Name               string   `json:"name"`
       BackgroundImageUrl string   `json:"background_image_url"`
       TimeoutLimit       *int     `json:"timeout_limit"`
       VoiceCallEnabled   *bool    `json:"voice_call_enabled"`
       CallCount          int      `json:"call_count"`
   }
   ```

2. **服务层修改**：
   - 在 `GetQueueData` 方法中，从 `bindInfo` 读取配置信息
   - 设置默认值（name 为空时默认为 "WALLACE"）
   - 填充到响应结构中

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
**相关规范**: `.cursor/rules/scrum_story_point.mdc`, `.cursor/rules/specs.mdc`
