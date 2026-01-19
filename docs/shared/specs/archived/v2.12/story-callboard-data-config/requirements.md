> ⚠️ **已归档** - 此 Spec 已随 v2.12 发布。
>
> - 归档时间: 2026-01-04
> - 归档人: weifashi

# 叫号系统返回配置信息 需求文档

> 本文档定义 叫号系统返回配置信息 的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2025-12/callboard-data-config.md](../../../../team/proposals/2025-12/callboard-data-config.md) |
| **创建日期**      | 2025-12-11                                                                                                 |
| **负责人**        | 王昱                                                                                                       |
| **目标 Sprint**   | Sprint {N}                                                                                                   |
| **涉及技术栈**    | [x] Go (main/) [ ] Go (ttpos-bmp/) [ ] PHP (admin/) [ ] Vue (admin/views/)                                   |

## 📋 审核状态

| 项目         | 内容                     |
| ------------ | ------------------------ |
| **审核状态** | 待审核 |
| **审核人**   | {审核人姓名}             |
| **审核日期** | {YYYY-MM-DD}             |
| **审核意见** | {审核意见，如有}         |

---

## 📋 概述

扩展 `/callboard/data` 接口，在现有队列数据基础上返回叫号系统配置信息（系统名称、背景图片、超时限制、语音叫号开关、叫号次数），使设备端能够从单一接口获取完整数据，减少接口调用次数，提升性能和用户体验。

## 🎯 产品对齐

该功能支持叫号展示设备端统一获取配置信息，简化设备端逻辑，提升系统整体性能和用户体验。

## 📝 用户故事

**作为** 叫号展示设备端  
**我想** 从 `/callboard/data` 接口获取配置信息（系统名称、背景图片、超时限制、语音叫号开关、叫号次数）  
**以便于** 正确展示界面和功能，无需额外调用其他接口

---

## 功能需求

### Requirement 1: 扩展 `/callboard/data` 接口返回配置信息

**用户故事**: 作为叫号展示设备端，我想从 `/callboard/data` 接口获取配置信息，以便于正确展示界面和功能

#### 验收标准

1. **WHEN** 设备端调用 `/callboard/data` 接口 **THEN** 响应中 **SHALL** 必须包含 `name`、`background_image_url`、`timeout_limit`、`voice_call_enabled`、`call_count` 字段（必返字段）
2. **IF** 配置信息不存在或为空 **THEN** 系统 **SHALL** 返回默认值：
   - `name`: 默认为 "WALLACE"（如果为空）
   - `background_image_url`: 默认为空字符串 ""
   - `timeout_limit`: 默认为 0（如果为 nil）
   - `voice_call_enabled`: 默认为 false（如果为 nil）
   - `call_count`: 默认为 1（如果为 0）
3. **IF** 配置信息读取失败 **THEN** 系统 **SHALL** 仍返回队列数据，配置字段使用默认值，不影响队列数据返回
4. **WHEN** 配置信息存在 **THEN** 系统 **SHALL** 返回实际配置值，不使用默认值

#### 具体要求

- [ ] 1.1 扩展 `QueueDataResp` 响应结构，新增配置信息字段（必返字段，不使用 `omitempty`）
- [ ] 1.2 修改 `GetQueueData` 服务方法，从 Redis `DeviceBindInfo` 读取配置信息
- [ ] 1.3 实现默认值逻辑：
  - `name` 为空时设置为 "WALLACE"
  - `background_image_url` 为空时设置为空字符串
  - `timeout_limit` 为 nil 时设置为 0
  - `voice_call_enabled` 为 nil 时设置为 false
  - `call_count` 为 0 时设置为 1
- [ ] 1.4 确保配置信息读取失败时不影响队列数据返回，使用默认值填充配置字段
- [ ] 1.5 保持接口向后兼容，新增字段不影响现有功能

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 严格遵循 Controller → Service → Repository 分层
- **单一职责原则**: 每个文件应有单一、明确的目的
- **模块化设计**: Service 和 Repository 应独立且可复用
- **依赖管理**: Service 只能依赖其他 Service 接口，不能直接依赖 Repository
- **遵循规范**:
  - `.cursor/rules/go-main.mdc` - Go Main 开发规范
  - `.cursor/rules/go-bmp.mdc` - Go BMP 微服务规范
  - `.cursor/rules/php.mdc` - PHP 开发规范
  - `.cursor/rules/vue.mdc` - Vue 前端规范
  - `.cursor/rules/structs.mdc` - 项目结构规范

### API 设计要求

- [x] URL 使用 snake_case 命名（`/callboard/data`）
- [x] data 字段必须是对象，不能是 null 或数组
- [x] 响应格式：`{code, message, data{}}`
- [x] 参考: `.cursor/rules/api.mdc` - API 设计规范

### 数据库设计要求

- [x] 配置信息存储在 Redis 中，无需数据库变更
- [x] 使用现有 `DeviceBindInfo` 结构存储配置信息

### 性能要求

- [x] 本地响应时间 < 200ms（配置信息从 Redis 读取，性能影响可忽略）
- [x] 缓存策略（配置信息已存储在 Redis）
- [x] 不影响现有队列数据查询性能

### 测试要求

- [ ] Service 层测试覆盖率 ≥ 70%
- [ ] API 测试覆盖所有接口
- [ ] 测试默认值逻辑（配置为空、nil、0 的情况）
- [ ] 测试配置信息读取失败时的降级处理
- [ ] 参考: `.cursor/rules/go-main.mdc` - 测试规范

### 可靠性要求

- [x] 配置信息读取失败时优雅降级（使用默认值）
- [x] 错误日志记录（使用 Logger）
- [x] 不影响现有队列数据返回功能

---

## 验收标准

### 功能验收

1. **接口响应结构**: `/callboard/data` 接口响应中包含所有配置字段（必返）
2. **默认值处理**: 配置信息不存在或为空时，正确返回默认值
3. **降级处理**: 配置信息读取失败时，仍返回队列数据，配置字段使用默认值
4. **向后兼容**: 新增字段不影响现有设备端功能

### 测试验收

1. **单元测试**: Service 层测试覆盖默认值逻辑
2. **API 测试**: `/callboard/data` 接口测试通过，验证配置字段返回
3. **集成测试**: 端到端流程测试通过（设备端获取配置信息）
4. **边界测试**: 测试配置为空、nil、0 等各种边界情况

### 文档验收

1. **技术文档**: design.md 完整且准确（待创建）
2. **API 文档**: API 接口文档更新（Swagger 注释）
3. **代码注释**: 代码中包含默认值逻辑的注释说明

---

## 约束条件

### 技术约束

#### Go Main 模块

- 必须使用 Gin 框架
- 接口以 `I` 开头，实现以 `Impl` 结尾
- Service 只能依赖其他 Service 接口
- Repository 只能持有 db 实例，不能持有 DBManager
- 不使用 panic，返回 error
- 配置信息从 Redis 读取，不修改数据库

### 业务约束

- 配置字段必须返回（不使用 `omitempty`）
- 默认值必须明确且一致
- 不影响现有队列数据返回功能

### 资源约束

- 开发时间: 0.5 天
- Story Point: 2 (必须 ≤ 5)

---

## 依赖关系

### 技术依赖

- `main/app/service/callboard` - 叫号系统服务
- `main/app/dto/resp/callboard` - 响应 DTO
- Redis - 配置信息存储

### 服务依赖

- 无外部服务依赖

### 业务依赖

- 依赖 `story-shop-callboard-settings` Spec（叫号系统配置管理功能）
- 配置信息已通过商家管理端设置并存储在 Redis

---

## 风险和缓解

### 风险 1: 配置信息读取失败影响队列数据返回

**影响**: 中  
**概率**: 低  
**缓解措施**:

- 配置信息读取失败时使用默认值，不影响队列数据返回
- 添加错误日志记录，便于排查问题
- 确保 Redis 连接稳定

### 风险 2: 默认值不一致导致设备端显示异常

**影响**: 中  
**概率**: 低  
**缓解措施**:

- 明确文档化所有默认值
- 代码中统一使用常量定义默认值
- 添加单元测试验证默认值逻辑

---

## 时间表

- **Phase 1 - 需求评审**: 0.5 天
- **Phase 2 - 技术设计**: 0.5 天
- **Phase 3 - 开发实现**: 0.5 天
- **Phase 4 - 测试验证**: 0.5 天
- **总计**: 2 天（SP = 2）

---

## 参考资料

### 核心规范

- `.cursor/rules/go-main.mdc` - Go Main 核心约束
- `.cursor/rules/api.mdc` - API 设计规范
- `.cursor/rules/structs.mdc` - 项目结构规范

### 相关文档

- `docs/shared/specs/active/story-shop-callboard-settings/` - 叫号系统配置管理 Spec
- `main/app/api/v1/callboard/handler.go` - 叫号系统接口处理器
- `main/app/service/callboard/service.go` - 叫号系统服务实现
- `main/app/dto/resp/callboard.go` - 响应 DTO 定义

### 外部参考

- 无

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/2025-12/2025-12-11.md`
- 提醒：需求评审或范围调整若形成经验，应同步更新 Episode 并在 Proposal/Spec 互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-12-11  
**作者**: 王昱  
**审核者**: {审核者}
