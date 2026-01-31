# 统一外卖订单数据结构字段 需求文档

> 本文档定义在 GetOrderInfo 接口中增加 order_data 字段的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2026-01/v2.14.0-structured-order-data.md](../../../../team/proposals/2026-01/v2.14.0-structured-order-data.md) |
| **创建日期**      | 2026-01-12                                                                                                 |
| **负责人**        | rikugun                                                                                                       |
| **目标 Sprint**   | Sprint 当前                                                                                                   |
| **涉及技术栈**    | [x] Go (ttpos-bmp/) [ ] Go (main/) [ ] PHP (admin/) [ ] Vue (admin/views/)                                   |

## 📋 审核状态

| 项目         | 内容                     |
| ------------ | ------------------------ |
| **审核状态** | 已通过 |
| **审核人**   | rikugun             |
| **审核日期** | 2026-01-12             |
| **审核意见** | 技术方案合理，可以开始实施         |

---

## 📋 概述

当前 `GetOrderInfoResp` 只返回 `raw_data` 字段（原始平台 JSON 数据），前端需要自行解析不同平台（Grab、Lineman）的数据格式差异。本需求通过增加 `order_data` 字段，返回经过 `takeout_converter.go` 转换后的统一 `TakeoutOrder` JSON 格式，降低前端开发复杂度，集中管理数据转换逻辑。

**核心价值**：
- 前端直接使用统一格式，无需关心平台差异
- 数据转换逻辑集中管理，平台变更时只需修改一处
- 提高系统可扩展性，新增外卖平台时不影响现有调用方
- 向后兼容，保留 `raw_data` 字段

## 🎯 产品对齐

**业务目标**：支持多外卖平台订单统一管理

**技术目标**：
1. 降低前端开发成本，提升开发效率
2. 集中管理数据转换逻辑，提高系统可维护性
3. 为后续外卖平台扩展提供标准化基础

**对齐 v2.14.0**：作为 Lineman 下单功能（story-takeout-lineman-place-order）的配套技术优化

## 📝 用户故事

**作为** 前端开发者  
**我想** 接收统一格式的外卖订单数据  
**以便于** 快速展示订单信息，无需关心不同平台的数据格式差异

**作为** 后端开发者  
**我想** 集中管理数据转换逻辑  
**以便于** 降低维护成本，提高系统可扩展性

---

## 功能需求

### Requirement 1: Protobuf 定义扩展

**用户故事**: 作为系统设计者，我想在 `GetOrderInfoResp` 中增加 `order_data` 字段，以便于提供统一格式的订单数据

#### 验收标准

1. **WHEN** 查看 `order.proto` **THEN** `GetOrderInfoResp` **SHALL** 包含 `order_data` 字段（string 类型）
2. **WHEN** 重新生成 Protobuf 代码 **THEN** 编译 **SHALL** 成功无错误
3. **WHEN** 查看字段注释 **THEN** **SHALL** 清晰说明字段用途（存储 TakeoutOrder JSON）

#### 具体要求

- [ ] 1.1 在 `GetOrderInfoResp` 消息中增加 `string order_data = 6;` 字段
- [ ] 1.2 添加注释：`// 转换后的统一订单数据（TakeoutOrder JSON）`
- [ ] 1.3 保留现有所有字段（向后兼容）
- [ ] 1.4 执行 `make proto` 重新生成 Go 代码

---
### Requirement 2: 数据转换逻辑集成

**用户故事**: 作为后端开发者，我想在 `GetOrderInfo` 接口中调用 converter 转换逻辑，以便于自动生成 `order_data` 字段

#### 验收标准

1. **WHEN** 调用 `GetOrderInfo` 接口 **THEN** 系统 **SHALL** 调用 `takeout_converter.go` 进行数据转换
2. **IF** `provider_name` 为 `grab` **THEN** 系统 **SHALL** 调用 `ConvertGrabToTakeoutOrder`
3. **IF** `provider_name` 为 `lineman` **THEN** 系统 **SHALL** 调用 `ConvertLinemanToTakeoutOrder`
4. **WHEN** 转换成功 **THEN** `order_data` **SHALL** 包含 `TakeoutOrder` JSON 字符串
5. **WHEN** 转换失败 **THEN** `order_data` **SHALL** 为空字符串，**AND** 系统 **SHALL** 记录错误日志

#### 具体要求

- [ ] 2.1 在 `GetOrderInfo` 实现中导入 `utility` 包
- [ ] 2.2 根据 `provider_name` 选择对应的转换函数
- [ ] 2.3 处理转换错误，不影响接口返回（优雅降级）
- [ ] 2.4 记录转换失败的错误日志（包含订单 ID、provider、错误信息）
- [ ] 2.5 序列化 `TakeoutOrder` 为 JSON 字符串
- [ ] 2.6 填充到 `GetOrderInfoResp.OrderData` 字段

---

### Requirement 3: 数据一致性验证

**用户故事**: 作为质量保证人员，我想验证 `order_data` 与 `raw_data` 的一致性，以便于确保数据转换正确

#### 验收标准

1. **WHEN** 对比 `raw_data` 和 `order_data` **THEN** 核心字段（订单号、商品数量、总价）**SHALL** 一致
2. **WHEN** `raw_data` 为 Grab 订单 **THEN** `order_data.orderID` **SHALL** 等于 `raw_data.orderID`
3. **WHEN** `raw_data` 为 Lineman 订单 **THEN** `order_data.orderID` **SHALL** 等于 `raw_data.orderId`

#### 具体要求

- [ ] 3.1 编写单元测试，验证 Grab 订单转换一致性
- [ ] 3.2 编写单元测试，验证 Lineman 订单转换一致性
- [ ] 3.3 测试边界情况（空商品列表、缺失字段、null 值）
- [ ] 3.4 验证 JSON 格式正确（可反序列化）

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 严格遵循 Controller → Service → Logic 分层（GoFrame 架构）
- **单一职责原则**: converter 只负责数据转换，不包含业务逻辑
- **模块化设计**: 转换逻辑独立于接口实现，可复用

### 测试要求

- [ ] Converter 测试覆盖率 100%（高风险转换逻辑）
- [ ] 集成测试覆盖 Grab 和 Lineman 两种场景
- [ ] 性能测试验证响应时间 < 50ms

### 性能要求

- [ ] 数据转换耗时 < 50ms（P99）
- [ ] 接口响应时间增加 < 20ms
- [ ] JSON 序列化使用标准库
- [ ] 不引入额外数据库查询

---

## 验收标准

### 功能验收

1. **Protobuf 定义**: `order.proto` 包含 `order_data` 字段，代码生成成功
2. **Grab 订单**: 调用接口返回正确的 `order_data` JSON，字段映射正确
3. **Lineman 订单**: 调用接口返回正确的 `order_data` JSON，字段映射正确
4. **转换失败**: 返回空字符串，记录错误日志，不影响接口正常返回
5. **向后兼容**: 现有调用方不受影响，`raw_data` 正常返回

### 测试验收

1. **单元测试**: Converter 测试覆盖率 100%，所有测试通过
2. **集成测试**: Grab 和 Lineman 两种场景的端到端测试通过
3. **性能测试**: 数据转换耗时 < 50ms，接口响应时间增加 < 20ms

### 文档验收

1. **技术文档**: design.md 完整且准确
2. **代码注释**: Protobuf、converter、service 层注释完整
3. **测试文档**: tasks.md 中的测试任务完成
4. **README 更新**: `ttpos-api/ttpos-takeout/message/README.md` 更新使用说明

---

## 约束条件

### 技术约束

- 必须使用 GoFrame 2.x
- 遵循 `ttpos-bmp/.cursor/rules/go-rules.mdc`
- 不修改 `dao/entity/do/` 目录（自动生成）
- 使用 `g.Log()` 记录日志
- 不使用 panic，返回 error

### 业务约束

- 只支持 Grab 和 Lineman 两种平台（当前版本）
- 转换失败不影响接口返回（优雅降级）

### 资源约束

- 开发时间: 1-2 天
- Story Point: 3

---

## 依赖关系

### 技术依赖

- `ttpos-api/ttpos-takeout/message` - TakeoutOrder 统一模型
- `ttpos-bmp/app/ttpos-takeout/utility` - 数据转换工具
- `encoding/json` - JSON 序列化

### 业务依赖

- `story-takeout-lineman-place-order` - Lineman 下单功能（已实现 converter）
- 现有 Grab 订单功能

---

## 风险和缓解

### 风险 1: 转换性能影响接口响应时间

**影响**: 中  
**概率**: 中  
**缓解措施**:
- 性能测试验证耗时 < 50ms
- 添加性能监控日志
- 如不达标，后期可增加缓存优化

### 风险 2: 转换逻辑覆盖不全导致数据丢失

**影响**: 高  
**概率**: 低  
**缓解措施**:
- 使用已测试通过的 converter
- 编写完整的单元测试
- 集成测试验证真实订单数据

---

## 时间表

- **Phase 1 - Protobuf 定义扩展**: 0.5 天
- **Phase 2 - 数据转换逻辑集成**: 0.5 天
- **Phase 3 - 测试和验证**: 0.5 天
- **Phase 4 - 文档和代码审查**: 0.5 天
- **总计**: 2 天（SP = 3）

---

## 参考资料

### 相关文件

- `ttpos-bmp/app/ttpos-takeout/manifest/protobuf/order/order.proto` - Protobuf 定义
- `ttpos-bmp/app/ttpos-takeout/utility/takeout_converter.go` - 数据转换工具
- `ttpos-api/ttpos-takeout/message/takeout_order.go` - TakeoutOrder 模型
- `ttpos-api/ttpos-takeout/message/README.md` - 使用指南
- `ttpos-api/ttpos-takeout/message/MAPPING.md` - 字段映射表

### 核心规范

- `.cursor/rules/go-bmp.mdc` - Go BMP 开发规范
- `ttpos-bmp/.cursor/rules/go-rules.mdc` - Go BMP 代码规范
- `.cursor/rules/api.mdc` - API 设计规范

---

**版本**: v1.0.0  
**创建日期**: 2026-01-12  
**作者**: rikugun  
**审核者**: 待审核
