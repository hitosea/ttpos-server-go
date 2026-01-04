# 在 /assistant/desk/ping 接口中返回已选国旗 ID 需求提案

> 本文档用于需求发起阶段，经团队评审后创建正式 Spec。

---

## 📋 提案信息

| 项目          | 内容                                                                                                         |
| ------------- | ------------------------------------------------------------------------------------------------------------ |
| **提案人**    | weifashi                                                                                                     |
| **日期**      | 2025-11-25                                                                                                   |
| **目标版本**  | v2.11.0                                                                                                      |
| **状态**      | 已创建 Spec                                                                                                  |
| **关联任务**  | -                                                                                                            |
| **关联 Spec** | [story-assistant-desk-ping-nationality](../../../shared/specs/archived/v2.12/story-assistant-desk-ping-nationality/) |

---

## 🎯 背景和动机

### 问题描述

当前 `/assistant/desk/ping` 接口用于定时轮询桌台详情，但返回的数据中不包含已选国籍 ID（nationality_uuid）信息。前端助手端在轮询时需要获取当前桌台订单的国籍信息，以便在 UI 上显示或进行相关业务处理。

**现状**：

- 系统已支持通过 `/assistant/desk/set_nationality` 接口设置桌台订单的国籍
- `SaleBill` 模型中已有 `nationality_uuid` 字段存储国籍信息
- 但 `/assistant/desk/ping` 接口返回的 `DeskPing` 响应中缺少该字段

**痛点**：

- 前端需要通过额外的接口查询才能获取国籍信息
- 增加了不必要的网络请求和复杂度
- 影响用户体验和系统性能

### 业务价值

- **提升前端开发效率**：减少接口调用次数，简化前端逻辑
- **改善用户体验**：前端可以实时显示当前桌台的国籍信息
- **降低系统负载**：减少不必要的接口请求
- **保持数据一致性**：在轮询接口中统一返回桌台相关信息

### 目标用户

- [x] 收银员
- [ ] 商户管理员
- [ ] 厨房人员
- [ ] 顾客
- [x] 其他: 助手端开发人员

---

## 💡 解决方案概述

### 方案描述

在 `/assistant/desk/ping` 接口的响应结构 `DeskPing` 中添加 `nationality_uuid` 字段，返回当前桌台订单的国籍 ID。当桌台未开台或订单未设置国籍时，返回 `0`。

**实现要点**：

1. 在 `resp.DeskPing` 结构体中添加 `NationalityUuid uint64` 字段
2. 在 `service.GetDeskPing` 方法中，从 `desk.SaleBill.NationalityUuid` 获取值并赋值给响应
3. 更新 Swagger 文档，说明新增字段
4. 保持向后兼容，未设置国籍时返回 `0`

### 核心功能点

1. **响应结构扩展**：在 `DeskPing` 中添加 `nationality_uuid` 字段
2. **数据获取**：从销售账单（SaleBill）中读取国籍 UUID
3. **兼容性处理**：未设置国籍时返回 `0`，保持向后兼容
4. **文档更新**：更新 API 文档，说明新增字段含义

### 影响范围

**涉及终端**：

- [ ] POS 收银端
- [ ] Shop 商家管理端
- [ ] KDS 厨显端
- [ ] QDS 排号叫号端
- [x] Assistant 助手端
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
- [ ] 其他: Swagger 文档

---

## 📊 初步评估

### 技术复杂度

- [x] **低**：纯 UI 调整，无业务逻辑变更
- [ ] **中**：需要前后端联调，基础业务逻辑
- [ ] **高**：涉及架构调整、第三方集成、复杂算法

**说明**：此功能为简单的字段扩展，数据已存在于 `SaleBill` 模型中，只需在响应结构中添加字段并赋值即可。

### 工作量预估

[粗略估算，具体 SP 在技术方案评审时确定]

- **预计天数**: 0.5 天
- **预估 SP**: 1（待技术评审确认）

**工作内容**：

1. 修改响应结构体（5 分钟）
2. 修改 Service 层逻辑（10 分钟）
3. 更新 Swagger 文档（5 分钟）
4. 测试验证（30 分钟）

### 风险识别

**潜在风险**：

1. **数据一致性**：确保从 `SaleBill` 中正确读取 `NationalityUuid`
2. **向后兼容**：未设置国籍时返回 `0`，前端需要处理该情况

**缓解措施**：

1. 在 Service 层添加空值检查，确保 `desk.SaleBill` 不为 `nil` 时再读取
2. 在 API 文档中明确说明：`nationality_uuid` 为 `0` 时表示未设置国籍
3. 前端已有处理 `0` 值的逻辑（参考其他接口）

---

## 🔗 相关资源

### 参考需求

- 类似功能: `/assistant/desk/set_nationality` - 设置桌台订单国籍接口
- 相关 Spec: `story-order-source-nationality` - 订单来源和国籍功能

### 相关文档

- API 文档: `main/docs/swagger.yaml`
- 响应结构: `main/app/dto/resp/desk.go`
- Service 实现: `main/app/service/desk.go`
- 数据模型: `main/app/model/sale_bill.go`

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

- [x] 创建 Spec：`story-assistant-desk-ping-nationality`
- [ ] 分配负责人：{姓名}
- [ ] 目标 Sprint：Sprint {N}

---

## 📝 附录

### User Story（初稿）

**作为** 助手端开发人员  
**我想** 在 `/assistant/desk/ping` 接口响应中获取当前桌台订单的国籍 ID  
**以便于** 在前端 UI 中显示国籍信息，无需额外接口调用

### AC 验收标准（初稿）

1. **WHEN** 调用 `/assistant/desk/ping` 接口 **THEN** 响应中 **SHALL** 包含 `nationality_uuid` 字段
2. **IF** 桌台已开台且订单已设置国籍 **THEN** `nationality_uuid` **SHALL** 返回对应的国籍 UUID（大于 0）
3. **IF** 桌台未开台或订单未设置国籍 **THEN** `nationality_uuid` **SHALL** 返回 `0`
4. **WHEN** 通过 `/assistant/desk/set_nationality` 设置国籍后 **THEN** 下次轮询 `/assistant/desk/ping` **SHALL** 返回更新后的 `nationality_uuid`

### 技术实现细节（初稿）

**修改文件**：

1. `main/app/dto/resp/desk.go` - 在 `DeskPing` 结构体中添加字段
2. `main/app/service/desk.go` - 在 `GetDeskPing` 方法中赋值
3. `main/docs/swagger.yaml` - 更新 API 文档

**代码示例**：

```go
// resp/desk.go
type DeskPing struct {
    // ... 现有字段 ...
    NationalityUuid uint64 `json:"nationality_uuid"` // 国籍UUID（0=未设置）
}

// service/desk.go
func (s *deskSrv) GetDeskPing(...) (resp.DeskPing, error) {
    // ... 现有逻辑 ...

    // 设置国籍UUID
    if desk.SaleBill != nil {
        res.NationalityUuid = desk.SaleBill.NationalityUuid
    }

    return res, nil
}
```

### 线框图/原型（可选）

无需 UI 变更，仅 API 响应字段扩展。

---

## 📄 模板使用说明

### 何时使用此模板

- ✅ 产品经理提出新功能想法
- ✅ 用户反馈需求建议
- ✅ 技术团队提出改进方案
- ✅ 需要团队讨论和评审的需求

### 与 Spec 的区别

| 阶段         | 文档类型     | 详细程度 | 用途                      |
| ------------ | ------------ | -------- | ------------------------- |
| **需求发起** | Proposal     | 粗略     | 团队评审、决策是否做      |
| **需求确认** | Requirements | 详细     | User Story + AC，开发依据 |
| **技术设计** | Design       | 详细     | 技术方案，实现指导        |
| **任务分解** | Tasks        | 详细     | 开发执行，进度追踪        |

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
**创建日期**: 2025-11-25  
**维护者**: 产品组 + Scrum Master  
**相关规范**: `.cursor/rules/scrum_story_point.mdc`, `.cursor/rules/specs.mdc`
