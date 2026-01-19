# SaveMaterialRequestReq RefNo 字段补充提案

> 本文档用于补充说明 SaveMaterialRequestReq 增加 ref_no 字段的必要性和价值。

---

## 📋 提案信息

| 项目       | 内容     |
| ---------- | -------- |
| **提案人** | weifashi |
| **日期**   | 2025-11-27 |
| **目标版本** | v2.x.x |
| **状态**   | 待评审   |
| **关联任务** | - |
| **关联 Spec** | docs/shared/specs/archived/v2.10.0/task-erp-material-request-refno/ |

---

## 🎯 背景和动机

### 问题描述

当前 `stock.SaveMaterialRequestReq` 消息结构中缺少来源单据号字段，导致以下问题：

1. **排错困难**：当 ttpos 调用 ERP 创建物料申请单时，无法在 ERP 侧追溯到原始的 ttpos 订单号
2. **数据关联断裂**：ttpos 与 ERP 之间的单据无法建立明确的对应关系
3. **问题定位耗时**：出现问题时，需要通过时间、金额等间接条件反向匹配，效率低下

**实际场景案例**：

> 运维人员接到报障："某商户反馈物料申请单数据异常"。由于 ERP 侧只能看到物料申请单编号（如 MAT-REQ-001234），无法直接对应回 ttpos 的原始订单号，运维人员需要：
>
> 1. 查询 ERP 物料申请单的创建时间
> 2. 回到 ttpos 系统，按时间范围、商户、金额等条件反向搜索
> 3. 逐一比对可能的订单，耗时 15-30 分钟
>
> 如果有 `ref_no` 字段存储 ttpos 订单号，可以在 5 秒内直接定位问题源头。

### 业务价值

- **提升排错效率**：通过 RefNo 可直接定位 ttpos 原始订单，将问题排查时间从 15-30 分钟缩短到 5 秒
- **增强数据可追溯性**：建立 ttpos 订单与 ERP 物料申请单的明确关联，形成完整的数据链路
- **降低运维成本**：减少问题排查时间，提高系统可维护性，降低人工介入成本
- **支持数据审计**：为后续数据审计、业务分析提供基础支撑

### 目标用户

- [x] 运维人员（主要受益者）
- [x] 开发人员（调试和问题定位）
- [x] 商户管理员（查询单据关联）
- [ ] 厨房人员
- [ ] 顾客
- [x] 其他: 数据分析团队

---

## 💡 解决方案概述

### 方案描述

在 `stock.SaveMaterialRequestReq` protobuf 消息中新增可选字段 `ref_no`，用于存储 ttpos 传递的原始订单号。该字段完全向后兼容，不影响现有调用方，调用方可按需选择是否传入该字段。

**技术实现**：
- 修改 `ttpos-bmp/app/ttpos-erp/manifest/protobuf/stock/stock.proto`
- 新增 `string ref_no = 10;` 字段
- 执行 `gf gen pb` 重新生成 Go 代码
- 无需修改业务逻辑代码（字段仅用于追溯，不参与业务逻辑）

### 核心功能点

1. **字段新增**：在 `SaveMaterialRequestReq` 中新增 `ref_no` 字段（string 类型，可选）
2. **向后兼容**：字段为可选，不传时默认为空字符串，现有调用方无需修改
3. **数据传递**：ttpos 调用 ERP 时可选传入原始订单号（如 `ORDER-20251127-001234`）
4. **可选日志**：ERP 侧可选择在日志中记录 `ref_no`，便于问题排查

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
- [x] 其他: ttpos-erp 微服务（后端基础设施）

**涉及模块**：
- [ ] UI 组件
- [x] API 接口（protobuf 定义）
- [x] 数据模型（自动生成）
- [ ] 业务逻辑（无需变更）
- [ ] 第三方集成
- [x] 其他: gRPC 接口定义

---

## 📊 初步评估

### 技术复杂度

- [x] **低**：纯 protobuf 字段新增，无业务逻辑变更
- [ ] **中**：需要前后端联调，基础业务逻辑
- [ ] **高**：涉及架构调整、第三方集成、复杂算法

**复杂度说明**：
- 仅涉及 protobuf 字段新增，执行 `gf gen pb` 自动生成代码
- 无需修改 Controller、Logic、Service 层
- 无数据库变更，无性能影响
- 完全向后兼容，无破坏性变更

### 工作量预估

- **预计天数**: 0.5 天
- **预估 SP**: 1 SP（待技术评审确认）

**工作量拆解**：
- Protobuf 修改：0.25 天
- 代码生成和验证：0.25 天

### 风险识别

**潜在风险**：

1. **风险1**：调用方未及时更新，导致 `ref_no` 字段为空
   - **影响**: 低（字段为可选，不影响功能）
   - **概率**: 低
   - **缓解措施**: 字段为可选，调用方可按需更新，不强制要求

2. **风险2**：字段编号冲突
   - **影响**: 中（protobuf 字段编号冲突会导致编译失败）
   - **概率**: 极低
   - **缓解措施**: 已确认使用下一个可用编号（10）

**缓解措施**：
1. 字段为可选，完全向后兼容，无强制要求
2. 提前验证 protobuf 编译正确性
3. 提供调用方集成示例文档

---

## 🔗 相关资源

### 参考需求

- 类似功能: ERP 系统常见的单据关联字段设计
- 竞品分析: 其他 ERP 系统普遍提供来源单据号字段

### 相关文档

- 需求文档 (Requirements): `docs/shared/specs/archived/v2.10.0/task-erp-material-request-refno/requirements.md`
- 设计文档 (Design): `docs/shared/specs/archived/v2.10.0/task-erp-material-request-refno/design.md`
- 任务清单 (Tasks): `docs/shared/specs/archived/v2.10.0/task-erp-material-request-refno/tasks.md`
- Protobuf 规范: `ttpos-bmp/.cursor/rules/proto-rules.mdc`
- Go BMP 规范: `ttpos-bmp/.cursor/rules/go-rules.mdc`

---

## 🤝 需求评审

### 评审参与人

| 角色         | 姓名   | 签名/日期 |
| ------------ | ------ | --------- |
| 产品经理     | weifashi |           |
| 技术负责人   | rikugun |           |
| 开发代表     | -      |           |
| 测试代表     | -      |           |
| UI/UX 设计师 | -      |           |

### 评审结论

- [ ] ✅ **批准**：进入技术方案设计阶段（已完成）
- [ ] 🔄 **修改后批准**：需补充以下内容
- [ ] ❌ **拒绝**：不符合产品规划或优先级

**评审意见**：

```
[待补充评审会议的关键讨论和决策]
```

**下一步行动**：

- [x] 创建 Spec：`task-erp-material-request-refno`（已完成）
- [x] 分配负责人：rikugun
- [ ] 目标 Sprint：Sprint {N}（待确定）
- [ ] 开始开发实现

---

## 📝 附录

### User Story（已确认）

**作为** 开发/运维人员  
**我想** 在 ERP 物料申请单中看到 ttpos 原始订单号  
**以便于** 快速定位和排查跨系统问题

### AC 验收标准（已确认）

1. **WHEN** ttpos 调用 SaveMaterialRequest 接口时传入 ref_no **THEN** 系统 **SHALL** 正确接收并处理该字段
2. **IF** ref_no 未传入 **THEN** 系统 **SHALL** 正常处理请求，不影响现有业务逻辑
3. **WHEN** 查询物料申请单详情时 **THEN** 系统 **SHALL** 返回关联的 ref_no 值（如已设置）

### 实现时间表

- **Phase 1 - Protobuf 修改**: 0.25 天
  - 修改 `stock.proto`，新增 `ref_no` 字段
  - 添加中文注释说明用途

- **Phase 2 - 代码生成和验证**: 0.25 天
  - 执行 `gf gen pb` 重新生成 Go 代码
  - 验证生成的代码正确
  - 验证向后兼容性

- **总计**: 0.5 天（SP = 1）

### 技术细节补充

**Protobuf 字段定义**：

```protobuf
message SaveMaterialRequestReq {
  // ... 现有字段 ...
  string ref_no = 10;  // 来源单据号，可选，用于跟踪 ttpos 原始订单号
}
```

**调用示例**（ttpos → ttpos-erp）：

```go
// ttpos 调用方
req := &stock.SaveMaterialRequestReq{
    CompanyAbbr:      "ACME",
    Branch:           "main-branch",
    TransactionDate:  time.Now().Unix(),
    RequiredBy:       time.Now().Add(24*time.Hour).Unix(),
    SourceWarehouse:  "WH-001",
    TargetWarehouse:  "WH-002",
    Purpose:          "Purchase",
    RefNo:            "ORDER-20251127-001234",  // 新增字段
    Items:            items,
}
```

---

## 📄 模板使用说明

### 何时使用此模板

- ✅ 产品经理提出新功能想法
- ✅ 用户反馈需求建议
- ✅ 技术团队提出改进方案
- ✅ 需要团队讨论和评审的需求
- ✅ **补充说明现有 Spec 的必要性和价值**

### 与 Spec 的关系

本提案作为 Spec `task-erp-material-request-refno` 的补充说明文档，重点阐述：

1. **问题背景**：为什么需要这个字段（实际痛点和场景）
2. **业务价值**：解决问题后带来的具体收益（量化指标）
3. **风险评估**：潜在风险和缓解措施（向后兼容性保证）
4. **评审记录**：团队讨论和决策过程

### 流转路径

```
提案 (Proposal) 
  ↓ 评审批准
需求文档 (Requirements) ← 已完成
  ↓ 技术评审
设计文档 (Design) ← 已完成
  ↓ SP 评估 = 1
任务分解 (Tasks) ← 已完成
  ↓
开发实现 ← 待执行
```

---

**版本**: v1.0.0  
**创建日期**: 2025-11-27  
**维护者**: weifashi  
**相关规范**: `.cursor/rules/scrum_story_point.mdc`, `.cursor/rules/specs.mdc`

