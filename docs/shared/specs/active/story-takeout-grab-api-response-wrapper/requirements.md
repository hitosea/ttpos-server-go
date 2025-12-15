# grab-api-response-wrapper 需求文档

> 本文档定义 将 Grab 服务 API 响应格式统一为 takeout.ApiResponse 的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2025-12/story-takeout-grab-api-response-wrapper.md](../../../../team/proposals/2025-12/story-takeout-grab-api-response-wrapper.md) |
| **创建日期**      | 2025-12-15                                                                                                 |
| **负责人**        | rikugun                                                                                                       |
| **目标 Sprint**   | Sprint 当前                                                                                                   |
| **涉及技术栈**    | [x] Go (ttpos-bmp/) [ ] Go (main/) [ ] PHP (admin/) [ ] Vue (admin/views/)                                   |

## 📋 审核状态

| 项目         | 内容                     |
| ------------ | ------------------------ |
| **审核状态** | 已通过                   |
| **审核人**   | rikugun                  |
| **审核日期** | 2025-12-15               |
| **审核意见** | 快速原型开发，直接通过 |

---

## 📋 概述

修改 `ttpos-bmp/app/ttpos-takeout/manifest/protobuf/grab/grab.proto` 中的 `CreateSelfServeJourney` 和 `GetShopProviderCfg` 方法，将返回值从自定义响应消息改为使用统一的 `takeout.ApiResponse` 包装器，确保 API 响应格式的一致性。

## 🎯 产品对齐

该功能支持 TTPOS 技术架构的统一性，提升 API 接口的标准化水平，为后续的微服务治理和前端集成提供更好的基础。

## 📝 用户故事

**作为** 技术开发人员  
**我想** 统一 Grab 服务 API 的响应格式  
**以便于** 提高代码维护性和 API 一致性

---

## 功能需求

### Requirement 1: 修改 CreateSelfServeJourney 接口响应格式

**用户故事**: 作为技术开发人员，我想将 CreateSelfServeJourney 接口的响应格式改为统一的 ApiResponse，以保持 API 一致性。

#### 验收标准

1. **WHEN** 调用 CreateSelfServeJourney 接口 **THEN** 系统 **SHALL** 返回 takeout.ApiResponse 格式的响应
2. **IF** 接口调用成功 **THEN** 系统 **SHALL** 在 data 字段中返回 CreateSelfServeJourneyResp 的数据结构
3. **IF** 接口调用失败 **THEN** 系统 **SHALL** 在 code 和 message 字段中返回错误信息

#### 具体要求

- [ ] 1.1 修改 protobuf 文件中的方法定义，将返回值改为 takeout.ApiResponse
- [ ] 1.2 更新服务实现代码，适配新的响应格式
- [ ] 1.3 重新生成 protobuf Go 代码
- [ ] 1.4 验证前后端集成正常工作

---

### Requirement 2: 修改 GetShopProviderCfg 接口响应格式

**用户故事**: 作为技术开发人员，我想将 GetShopProviderCfg 接口的响应格式改为统一的 ApiResponse，以保持 API 一致性。

#### 验收标准

1. **WHEN** 调用 GetShopProviderCfg 接口 **THEN** 系统 **SHALL** 返回 takeout.ApiResponse 格式的响应
2. **IF** 接口调用成功 **THEN** 系统 **SHALL** 在 data 字段中返回 GetShopProviderCfgResp 的数据结构
3. **IF** 接口调用失败 **THEN** 系统 **SHALL** 在 code 和 message 字段中返回错误信息

#### 具体要求

- [ ] 2.1 修改 protobuf 文件中的方法定义，将返回值改为 takeout.ApiResponse
- [ ] 2.2 更新服务实现代码，适配新的响应格式
- [ ] 2.3 重新生成 protobuf Go 代码
- [ ] 2.4 验证前后端集成正常工作

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 严格遵循 Controller → Service → Repository 分层
- **单一职责原则**: 每个文件应有单一、明确的目的
- **模块化设计**: Service 和 Repository 应独立且可复用
- **依赖管理**: Service 只能依赖其他 Service 接口，不能直接依赖 Repository
- **遵循规范**:
  - `.cursor/rules/go-bmp.mdc` - Go BMP 微服务规范
  - `ttpos-bmp/.cursor/rules/go-rules.mdc` - Go BMP 开发规范
  - `ttpos-bmp/.cursor/rules/proto-rules.mdc` - Protobuf 开发规范

### API 设计要求

- [x] URL 使用 snake_case 命名（如：`/api/v1/order_info`）
- [x] data 字段必须是对象，不能是 null 或数组
- [ ] 分页信息统一放在 meta 中
- [x] 响应格式：`{code, message, data{}}`
- [ ] 参考: `.cursor/rules/api.mdc` - API 设计规范

### 数据库设计要求

- [ ] 无数据库变更，无需遵循数据库规范

### 性能要求

- [x] 本地响应时间 < 200ms
- [ ] 数据库查询优化（使用索引）
- [ ] 缓存策略（Redis）
- [ ] 并发处理（使用 UUID 锁）

### 浏览器兼容性（管理后台）

- [ ] 无前端变更，无需浏览器兼容性要求

### 测试要求

- [x] Service 层测试覆盖率 ≥ 70%
- [x] Repository 层测试覆盖率 ≥ 80%
- [ ] **Payment/Order 相关模块测试覆盖率 100%**（高风险）
- [x] 集成测试覆盖核心流程
- [x] API 测试覆盖所有接口
- [ ] 参考: `ttpos-bmp/.cursor/rules/go-rules.mdc` - 测试规范

### 国际化要求

- [ ] 无用户界面变更，无需国际化要求

### 安全要求

- [x] 所有 API 需要身份验证
- [ ] 敏感数据加密存储
- [ ] SQL 注入防护（使用参数化查询）
- [ ] XSS 防护（前端输入校验）
- [ ] CSRF 防护（Token 验证）
- [ ] 参考: `.cursor/rules/security.mdc` - 安全开发规范

### 可靠性要求

- [x] 网络异常时优雅降级
- [ ] 事务管理（保证数据一致性）
- [x] 错误日志记录（使用 Logger）
- [x] 故障恢复机制

---

## 验收标准

### 功能验收

1. **接口响应格式统一**: CreateSelfServeJourney 和 GetShopProviderCfg 接口都返回 takeout.ApiResponse 格式
2. **数据结构完整性**: 成功响应的 data 字段包含完整的业务数据
3. **错误处理一致性**: 失败响应包含正确的错误码和错误信息
4. **向后兼容性**: 现有客户端可以通过适配器正常使用新接口

### 测试验收

1. **单元测试**: 覆盖率达标，Service 层 ≥ 70%
2. **API 测试**: 所有接口测试通过，包括成功和失败场景
3. **集成测试**: 端到端流程测试通过
4. **手动测试**: 验证实际调用效果

### 文档验收

1. **技术文档**: 更新后的 protobuf 文件和实现代码文档完整
2. **API 文档**: 接口变更说明文档完整
3. **测试文档**: 测试用例和测试结果文档完整

---

## 约束条件

### 技术约束

#### Go BMP 模块

- 必须使用 GoFrame 2.x
- 禁止修改 dao/entity/do/ 目录（自动生成）
- gRPC 服务必须注册到 Nacos
- 遵循 `ttpos-bmp/.cursor/rules/go-rules.mdc`
- 遵循 `ttpos-bmp/.cursor/rules/proto-rules.mdc`

### 业务约束

- 不影响现有的业务逻辑和数据处理
- 保持接口的功能完整性
- 确保向后兼容性

### 资源约束

- 开发时间: 1 天
- Story Point: 2 (必须 ≤ 5)

---

## 依赖关系

### 技术依赖

- `google.golang.org/grpc` - gRPC 通信框架
- `github.com/gogf/gf/v2` - GoFrame 2.x 框架
- `takeout.ApiResponse` - 统一的 API 响应格式

### 服务依赖

- **BMP → BMP**: gRPC 调用（内部服务间通信）

### 业务依赖

- ttpos-takeout 模块的 Grab 服务已正常运行
- takeout.ApiResponse 类型已定义

---

## 风险和缓解

### 风险 1: 前端客户端适配困难

**影响**: 中  
**概率**: 中  
**缓解措施**:

- 提供详细的 API 变更说明文档
- 编写客户端适配示例代码
- 安排技术支持协助前端团队适配

### 风险 2: 现有测试用例失效

**影响**: 中  
**概率**: 低  
**缓解措施**:

- 在修改前备份所有测试用例
- 重新编写测试用例以适配新的响应格式
- 进行充分的回归测试

---

## 时间表

- **Phase 1 - 接口定义修改**: 0.5 天
- **Phase 2 - 代码实现更新**: 0.3 天
- **Phase 3 - 测试验证**: 0.2 天
- **总计**: 1 天（SP = 2）

---

## 参考资料

### 核心规范

- `ttpos-bmp/.cursor/rules/go-rules.mdc` - Go BMP 开发规范
- `ttpos-bmp/.cursor/rules/proto-rules.mdc` - Protobuf 开发规范
- `.cursor/rules/api.mdc` - API 设计规范

### 架构文档

- `docs/human/architecture/go-bmp-architecture.md` - Go BMP 架构

### 开发指南

- `docs/human/guides/go-bmp-development.md` - Go BMP 开发指南
- `docs/human/guides/api-design-guide.md` - API 设计指南

### 外部参考

- [Protocol Buffers Documentation](https://developers.google.com/protocol-buffers)
- [gRPC Documentation](https://grpc.io/docs/)

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 提醒：需求评审或范围调整若形成经验，应同步更新 Episode 并在 Proposal/Spec 互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-12-15  
**作者**: rikugun  
**审核者**:
