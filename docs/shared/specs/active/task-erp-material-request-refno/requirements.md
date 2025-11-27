# SaveMaterialRequestReq 增加 RefNo 字段 需求文档

> 本文档定义 ttpos-erp stock 模块 SaveMaterialRequestReq 新增 ref_no 字段的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                                         |
| ----------------- | ---------------------------------------------------------------------------------------------------------------------------- |
| **来源 Proposal** | [docs/team/proposals/2025-11/stock-material-request-refno.md](../../../../team/proposals/2025-11/stock-material-request-refno.md) |
| **创建日期**      | 2025-11-27                                                                                                                   |
| **负责人**        | rikugun                                                                                                                      |
| **目标 Sprint**   | 待定                                                                                                                         |
| **涉及技术栈**    | [x] Go (ttpos-bmp/) [ ] Go (main/) [ ] PHP (admin/) [ ] Vue (admin/views/)                                                   |

## 📋 审核状态

| 项目         | 内容       |
| ------------ | ---------- |
| **审核状态** | ✅ 已通过  |
| **审核人**   | rikugun    |
| **审核日期** | 2025-11-27 |
| **审核意见** | 简单字段新增，向后兼容，批准开发 |

---

## 📋 概述

为 `stock.SaveMaterialRequestReq` protobuf 消息新增 `ref_no` 字段，用于存储 ttpos 传递的原始订单号。该字段支持在 ERP 侧追溯 ttpos 订单来源，提升问题排查效率和数据可追溯性。

## 🎯 产品对齐

该功能支持系统运维和问题排查能力的提升：
- 建立 ttpos 与 ERP 之间的单据关联
- 降低生产问题排查成本
- 提高系统可维护性

## 📝 用户故事

**作为** 开发/运维人员  
**我想** 在 ERP 物料申请单中看到 ttpos 原始订单号  
**以便于** 快速定位和排查跨系统问题

---

## 功能需求

### Requirement 1: SaveMaterialRequestReq 新增 ref_no 字段

**用户故事**: 作为开发人员，我想在创建物料申请单时传入 ttpos 订单号，以便于后续排错追溯

#### 验收标准

1. **WHEN** ttpos 调用 SaveMaterialRequest 接口时传入 ref_no **THEN** 系统 **SHALL** 正确接收并处理该字段
2. **IF** ref_no 未传入 **THEN** 系统 **SHALL** 正常处理请求，不影响现有业务逻辑
3. **WHEN** 查询物料申请单详情时 **THEN** 系统 **SHALL** 返回关联的 ref_no 值（如已设置）

#### 具体要求

- [x] 1.1 在 `stock.proto` 的 `SaveMaterialRequestReq` 消息中新增 `string ref_no = 10;` 字段
- [x] 1.2 字段为可选，不传时默认为空字符串
- [x] 1.3 字段注释说明用途：`// 来源单据号，可选，用于跟踪 ttpos 原始订单号`
- [ ] 1.4 执行 `gf gen pb` 重新生成 Go 代码

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 遵循 ttpos-bmp 的 Controller → Logic → Service 分层
- **遵循规范**:
  - `ttpos-bmp/.cursor/rules/go-rules.mdc` - Go BMP 开发规范
  - `ttpos-bmp/.cursor/rules/proto-rules.mdc` - Protobuf 开发规范

### API 设计要求

- [x] 字段命名使用 snake_case（`ref_no`）
- [x] 字段为可选，向后兼容
- [x] 字段编号使用下一个可用编号（10）

### 性能要求

- [x] 无性能影响（仅字段新增，无额外逻辑）

### 测试要求

- [ ] 验证新字段能正确传递和接收
- [ ] 验证不传 ref_no 时接口正常工作

---

## 验收标准

### 功能验收

1. **字段定义**: `SaveMaterialRequestReq` 包含 `ref_no` 字段
2. **向后兼容**: 不传 `ref_no` 时接口正常工作
3. **数据传递**: 传入的 `ref_no` 能被正确接收

### 测试验收

1. **单元测试**: 验证字段序列化/反序列化
2. **集成测试**: 验证端到端调用正常

### 文档验收

1. **Protobuf 注释**: 字段有清晰的用途说明

---

## 约束条件

### 技术约束

#### Go BMP 模块

- 必须使用 GoFrame 2.x
- protobuf 修改后需执行 `gf gen pb` 重新生成
- 遵循 `ttpos-bmp/.cursor/rules/proto-rules.mdc`

### 业务约束

- 字段为可选，不影响现有调用方
- 字段仅用于跟踪追溯，不参与业务逻辑

### 资源约束

- 开发时间: 0.5 天
- Story Point: 1 SP

---

## 依赖关系

### 技术依赖

- `protoc` - protobuf 编译器
- `gf gen pb` - GoFrame protobuf 代码生成

### 服务依赖

- **ttpos → ttpos-erp**: gRPC 调用（调用方需更新传入 ref_no）

### 业务依赖

- 无前置依赖

---

## 风险和缓解

### 风险 1: 调用方未及时更新

**影响**: 低  
**概率**: 低  
**缓解措施**:

- 字段为可选，不影响现有调用方
- 调用方可按需更新

---

## 时间表

- **Phase 1 - Protobuf 修改**: 0.25 天
- **Phase 2 - 代码生成和验证**: 0.25 天
- **总计**: 0.5 天（SP = 1）

---

## 参考资料

### 核心规范

- `ttpos-bmp/.cursor/rules/go-rules.mdc` - Go BMP 开发规范
- `ttpos-bmp/.cursor/rules/proto-rules.mdc` - Protobuf 开发规范

### 涉及文件

- `ttpos-bmp/app/ttpos-erp/manifest/protobuf/stock/stock.proto` - Protobuf 定义
- `ttpos-bmp/app/ttpos-erp/api/stock/stock.pb.go` - 生成的 Go 代码
- `ttpos-bmp/app/ttpos-erp/api/stock/stock_grpc.pb.go` - 生成的 gRPC 代码

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/rikugun/2025-11/2025-11-27.md`

---

**版本**: v1.0.0  
**创建日期**: 2025-11-27  
**作者**: rikugun  
**审核者**: -

