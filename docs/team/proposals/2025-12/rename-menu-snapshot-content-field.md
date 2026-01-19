# 调整 GetMenuSnapshotResp.content 为 GetMenuSnapshotResp.menu_data 需求提案

> 本文档用于需求发起阶段，经团队评审后创建正式 Spec。

---

## 📋 提案信息

| 项目       | 内容     |
| ---------- | -------- |
| **提案人** | rikugun   |
| **日期**   | 2025-12-15   |
| **目标版本** | - |
| **状态**   | 待评审   |
| **关联任务** | - |
| **关联 Spec** | [task-bmp-menu-snapshot-field-rename](../../../shared/specs/archived/v2.12/task-bmp-menu-snapshot-field-rename/requirements.md)      |

---

## 🎯 背景和动机

### 问题描述

在 `ttpos-bmp/app/ttpos-takeout/manifest/protobuf/menu/menu.proto` 中，`GetMenuSnapshotResp` 消息使用 `content` 字段存储菜单数据，而 `SaveMenuSnapshotReq` 消息使用 `menu_data` 字段存储菜单数据。这种命名不一致会导致：

1. **API 语义不清晰**：同一业务概念（菜单数据）使用不同字段名，增加理解成本
2. **代码可读性差**：开发者需要记住两个不同的字段名
3. **维护成本高**：未来扩展时容易混淆

**当前状态**：
```protobuf
message GetMenuSnapshotResp {
  string content = 2;          // Provider 侧原始菜单 JSON
  int64 updated_at = 3;
  string sync_state = 4;
}

message SaveMenuSnapshotReq {
  string provider_name = 1;
  string shop_uuid = 2;
  string menu_data = 3;     // 菜单数据 JSON 字符串
  string request_id = 4;
}
```

### 业务价值

- **提升代码一致性**：统一字段命名，降低理解成本
- **改善 API 设计**：遵循 RESTful 和 gRPC 最佳实践，保持命名一致性
- **降低维护成本**：减少因命名不一致导致的潜在错误

### 目标用户

- [ ] 收银员
- [ ] 商户管理员
- [ ] 厨房人员
- [ ] 顾客
- [x] **开发人员**：主要受益者
- [ ] 其他: ________

---

## 💡 解决方案概述

### 方案描述

将 `GetMenuSnapshotResp.content` 字段重命名为 `GetMenuSnapshotResp.menu_data`，使其与 `SaveMenuSnapshotReq.menu_data` 保持一致。

**修改内容**：
1. 修改 `menu.proto` 文件中的字段定义
2. 重新生成 Go 代码（`menu.pb.go`）
3. 更新使用该字段的业务代码（`channel_menu.go`）

### 核心功能点

1. **Proto 定义修改**：将 `GetMenuSnapshotResp.content` 改为 `GetMenuSnapshotResp.menu_data`
2. **代码生成**：重新生成 protobuf Go 代码
3. **业务代码更新**：更新 `channel_menu.go` 中对 `Content` 字段的引用

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
- [ ] UI 组件
- [x] **API 接口**：gRPC 接口定义
- [ ] 数据模型
- [x] **业务逻辑**：`channel_menu.go` 中的字段引用
- [ ] 第三方集成
- [ ] 其他: ________

**影响文件**：
- `ttpos-bmp/app/ttpos-takeout/manifest/protobuf/menu/menu.proto`
- `ttpos-bmp/app/ttpos-takeout/api/menu/menu.pb.go`（自动生成）
- `ttpos-bmp/app/ttpos-takeout/internal/logic/channel_menu/channel_menu.go`

---

## 📊 初步评估

### 技术复杂度

- [x] **低**：纯字段重命名，无业务逻辑变更
- [ ] **中**：需要前后端联调，基础业务逻辑
- [ ] **高**：涉及架构调整、第三方集成、复杂算法

### 工作量预估

[粗略估算，具体 SP 在技术方案评审时确定]

- **预计天数**: 0.5 天
- **预估 SP**: 1（待技术评审确认）

### 风险识别

**潜在风险**：
1. **向后兼容性**：如果已有客户端使用该接口，需要协调更新
2. **代码生成**：需要确保 protobuf 代码生成工具正常工作

**缓解措施**：
1. 检查是否有外部系统依赖该接口，如有需要协调更新
2. 在开发环境验证 protobuf 代码生成流程
3. 充分测试相关业务逻辑

---

## 🔗 相关资源

### 参考需求

- 类似功能: 无
- 竞品分析: 无

### 相关文档

- Protobuf 规范: `ttpos-bmp/.cursor/rules/proto-rules.mdc`
- Go 代码规范: `ttpos-bmp/.cursor/rules/go-rules.mdc`

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

- [x] 创建 Spec：`task-bmp-menu-snapshot-field-rename`
- [ ] 分配负责人：{姓名}
- [ ] 目标 Sprint：Sprint {N}

---

## 📝 附录

### User Story（初稿）

**作为** 开发人员  
**我想** 统一菜单快照接口的字段命名  
**以便于** 提高代码可读性和维护性，减少因命名不一致导致的错误

### AC 验收标准（初稿）

1. **WHEN** 查看 `GetMenuSnapshotResp` 消息定义 **THEN** 字段名应为 `menu_data` **SHALL** 与 `SaveMenuSnapshotReq.menu_data` 保持一致
2. **IF** 调用 `GetMenuSnapshot` 接口 **THEN** 响应中的菜单数据字段 **SHALL** 使用 `menu_data` 而非 `content`
3. **WHEN** 重新生成 protobuf 代码 **THEN** Go 代码中的字段名和方法名 **SHALL** 正确更新为 `MenuData` 和 `GetMenuData()`

### 线框图/原型（可选）

无需 UI 变更

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
**创建日期**: 2025-12-15  
**维护者**: 产品组 + Scrum Master  
**相关规范**: `.cursor/rules/scrum_story_point.mdc`, `.cursor/rules/specs.mdc`
