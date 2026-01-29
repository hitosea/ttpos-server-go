# 门店多渠道激活服务 需求文档

> 本文档定义门店多渠道激活服务的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                             |
| ----------------- | ---------------------------------------------------------------------------------------------------------------- |
| **来源 Proposal** | [docs/team/proposals/2026-01/v2.13.1-shop-activate-multi-provider.md](../../../../team/proposals/2026-01/v2.13.1-shop-activate-multi-provider.md) |
| **创建日期**      | 2026-01-07                                                                                                       |
| **负责人**        | rikugun                                                                                                          |
| **目标 Sprint**   | Sprint TBD                                                                                                       |
| **目标版本**      | v2.13.1                                                                                                          |
| **涉及技术栈**    | [x] Go (ttpos-bmp/) [ ] Go (main/) [ ] PHP (admin/) [ ] Vue (admin/views/)                                       |

## 📋 审核状态

| 项目         | 内容     |
| ------------ | -------- |
| **审核状态** | 待审核   |
| **审核人**   | -        |
| **审核日期** | -        |
| **审核意见** | -        |

---

## 📋 概述

当前系统中，门店激活外卖渠道的功能仅支持 Grab 渠道，且实现分散在 Grab 服务中。随着业务扩展，需要支持 LINE MAN 等更多外卖渠道。本需求旨在提供统一的门店激活服务，管理多个外卖渠道的激活流程，提升商户接入效率，降低维护成本。

**核心价值**：
- 支持多外卖渠道快速接入，提升业务扩展能力
- 统一门店激活流程，降低维护成本
- 为 LINE MAN 等新渠道提供标准化激活方案
- 提升商户接入效率，缩短上线周期

## 🎯 产品对齐

该功能支持公司外卖业务的战略扩展，通过统一的渠道接入机制，降低新渠道接入成本，加速业务规模化。为商户提供一站式多渠道管理能力，提升商户满意度和平台竞争力。

## 📝 用户故事

### Story 1: 激活外卖渠道

**作为** 商户管理员  
**我想** 通过统一的接口激活门店的外卖渠道（LINE MAN、Grab）  
**以便于** 快速接入多个外卖平台，提升门店线上订单能力

### Story 2: 查询渠道配置

**作为** 商户管理员  
**我想** 查询门店在各个外卖渠道的配置状态  
**以便于** 了解哪些渠道已激活、哪些待激活，方便管理和决策

---

## 功能需求

### Requirement 1: 门店渠道激活接口

**用户故事**: 作为商户管理员，我想通过统一接口激活门店的外卖渠道，以便于快速接入多个外卖平台

#### 验收标准

1. **WHEN** 调用 `ActivateShop` 且 `provider_name=lineman` **THEN** 系统 **SHALL** 创建 `shop_provider_cfg` 记录，状态为 `INACTIVE`
2. **WHEN** 调用 `ActivateShop` 且 `provider_name=grab` **THEN** 系统 **SHALL** 调用 Grab 自助激活服务，返回激活链接
3. **IF** `shop_uuid` 无效或为空 **THEN** 系统 **SHALL** 返回参数错误（400）
4. **IF** `provider_name` 不支持 **THEN** 系统 **SHALL** 返回不支持的渠道错误（400）
5. **WHEN** 激活成功 **THEN** 系统 **SHALL** 返回统一的 `ApiResponse` 格式，包含 shop_uuid、provider_name、updated_at（Grab 渠道额外返回 self_serve_url）

#### 具体要求

- [ ] 1.1 新增 `shop.ActivateShop` gRPC 服务
- [ ] 1.2 支持 lineman 和 grab 两个渠道的激活逻辑
- [ ] 1.3 Lineman 渠道：直接创建配置记录（状态 INACTIVE）
- [ ] 1.4 Grab 渠道：调用 `CreateSelfServeJourney` 创建激活链接（状态 SYNCING）
- [ ] 1.5 复用现有 `ShopProviderCfg` 服务管理配置
- [ ] 1.6 实现多渠道路由逻辑（策略模式）
- [ ] 1.7 统一错误处理和日志记录
- [ ] 1.8 返回格式统一使用 `takeout.ApiResponse` 包装

---

### Requirement 2: 门店渠道配置查询接口

**用户故事**: 作为商户管理员，我想查询门店在各个外卖渠道的配置状态，以便于了解激活情况并做出决策

#### 验收标准

1. **WHEN** 调用 `GetShopProviderCfg` 且 `provider_name` 为空 **THEN** 系统 **SHALL** 返回门店所有渠道的配置列表（lineman、grab）
2. **WHEN** 调用 `GetShopProviderCfg` 且 `provider_name` 不为空 **THEN** 系统 **SHALL** 仅返回该指定渠道的配置
3. **WHEN** 渠道配置存在 **THEN** 系统 **SHALL** 返回该渠道的配置信息（包含 merchant_id、status、updated_at）
4. **WHEN** 渠道配置不存在 **THEN** 系统 **SHALL** 返回该渠道状态为 `INACTIVE` 的默认配置
5. **IF** `shop_uuid` 为 0 **THEN** 系统 **SHALL** 返回参数错误（400）
6. **WHEN** 查询成功 **THEN** 系统 **SHALL** 返回统一的 `ApiResponse` 格式，Data 字段包含配置列表

#### 具体要求

- [ ] 2.1 新增 `shop.GetShopProviderCfg` gRPC 服务
- [ ] 2.2 支持两种查询模式：查询所有渠道 / 查询指定渠道
- [ ] 2.3 `provider_name` 为空时：返回所有支持渠道的配置列表
- [ ] 2.4 `provider_name` 不为空时：仅返回指定渠道的配置
- [ ] 2.5 配置不存在时返回默认 INACTIVE 状态
- [ ] 2.6 新增 `GetShopProviderCfgForRPC` Logic 层方法
- [ ] 2.7 统一响应格式，使用 `repeated ShopProviderCfgItem`

---

### Requirement 3: 常量和数据模型扩展

**用户故事**: 作为开发人员，我需要扩展系统支持的渠道类型，以便于支持 LINE MAN 渠道

#### 验收标准

1. **WHEN** 添加新渠道常量 **THEN** 系统 **SHALL** 在 `consts.go` 中定义 `ProviderLineman`
2. **WHEN** 使用新渠道常量 **THEN** 系统 **SHALL** 在所有相关逻辑中正确识别

#### 具体要求

- [ ] 3.1 在 `internal/consts/consts.go` 中新增 `ProviderLineman` 常量
- [ ] 3.2 在渠道路由逻辑中支持 lineman 分支
- [ ] 3.3 确保 `shop_provider_cfg` 表支持 lineman 渠道数据

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 严格遵循 Controller → Service/Logic → DAO 分层（GoFrame 规范）
- **单一职责原则**: 每个文件应有单一、明确的目的
- **模块化设计**: Service/Logic 应独立且可复用
- **遵循规范**:
  - `ttpos-bmp/.cursor/rules/go-rules.mdc` - Go BMP 开发规范
  - `ttpos-bmp/.cursor/rules/proto-rules.mdc` - Protobuf 开发规范
  - `.cursor/rules/go-ttpos-takeout` - ttpos-takeout 子模块规则
  - `.cursor/rules/structs.mdc` - 项目结构规范

### API 设计要求

- [x] 使用 gRPC 协议
- [x] 响应格式统一使用 `takeout.ApiResponse` 包装
- [x] Controller 层返回 `ApiResponse`，Logic 层返回具体业务数据类型
- [x] 请求参数使用 Protobuf 定义
- [x] 参考: `.cursor/rules/api.mdc` - API 设计规范

### 数据库设计要求

- [x] 复用现有 `shop_provider_cfg` 表
- [x] 必须包含: `uuid`, `shop_uuid`, `provider_name`, `provider_merchant_id`, `provider_shop_status`
- [x] 时间字段: `created_at`, `updated_at`, `deleted_at`（int 类型）
- [x] 字段名使用 snake_case
- [x] 参考: `.cursor/rules/database.mdc` - 数据库开发规范

### 性能要求

- [ ] gRPC 响应时间 < 200ms
- [ ] 数据库查询优化（使用索引）
- [ ] 缓存策略（如需要）

### 测试要求

- [ ] Logic 层测试覆盖率 ≥ 70%
- [ ] gRPC 接口测试覆盖所有场景
- [ ] 集成测试覆盖核心流程（激活 + 查询）
- [ ] 错误场景测试（参数错误、渠道不支持等）

### 安全要求

- [ ] 所有 gRPC 接口需要身份验证
- [ ] 参数校验（shop_uuid、provider_name）
- [ ] 防止 SQL 注入（使用 ORM）
- [ ] 错误信息不暴露敏感数据
- [ ] 参考: `.cursor/rules/security.mdc` - 安全开发规范

### 可靠性要求

- [ ] 网络异常时优雅降级
- [ ] 事务管理（保证数据一致性）
- [ ] 错误日志记录（使用 g.Log()）
- [ ] Grab API 调用失败时的重试机制

---

## 验收标准

### 功能验收

1. **ActivateShop 接口**: 成功激活 lineman 和 grab 渠道，返回正确的响应
2. **GetShopProviderCfg 接口**: 支持查询单个或所有渠道，正确返回配置状态
3. **渠道路由**: 根据 provider_name 正确路由到对应的激活逻辑
4. **错误处理**: 参数错误、不支持的渠道等场景正确返回错误信息
5. **数据持久化**: 激活后的配置正确保存到数据库

### 测试验收

1. **单元测试**: Logic 层测试覆盖率达标
2. **gRPC 测试**: 所有接口测试通过
3. **集成测试**: 端到端流程测试通过
4. **性能测试**: 响应时间符合要求

### 文档验收

1. **技术文档**: design.md 完整且准确
2. **API 文档**: Protobuf 定义文档完整
3. **测试文档**: tasks.md 中的测试任务完成

---

## 约束条件

### 技术约束

#### Go BMP 模块（ttpos-takeout）

- 必须使用 GoFrame 2.x 框架
- 禁止修改 dao/entity/do/service 目录（自动生成）
- gRPC 服务必须注册到 Nacos
- 遵循 `ttpos-bmp/.cursor/rules/go-rules.mdc`
- 遵循 `ttpos-bmp/.cursor/rules/proto-rules.mdc`
- Controller 层返回 `takeout.ApiResponse`
- Logic 层返回具体业务数据类型（不返回 `ApiResponse`）

#### Protobuf 规范

- 请求消息以 `Req` 结尾
- 响应消息以 `Resp` 结尾
- 字段名使用 snake_case
- 服务名以 `Service` 结尾
- 方法名使用大驼峰命名法（PascalCase）
- 添加详细的中文注释

### 业务约束

- LINE MAN 渠道激活流程仅创建配置记录（状态 INACTIVE），不调用第三方 API
- Grab 渠道激活需要调用 `CreateSelfServeJourney` 创建自助激活链接
- 配置不存在时返回默认 INACTIVE 状态，不报错
- 支持的渠道列表：lineman、grab

### 资源约束

- 开发时间: 2 天
- Story Point: 3 (必须 ≤ 5)

---

## 依赖关系

### 技术依赖

- `github.com/gogf/gf/v2` - GoFrame 核心框架
- `github.com/gogf/gf/contrib/rpc/grpcx/v2` - gRPC 扩展
- `google.golang.org/protobuf` - Protobuf 支持
- `ttpos-bmp/app/ttpos-takeout/internal/logic/shop_provider_cfg` - 门店配置管理服务
- `ttpos-bmp/app/ttpos-takeout/internal/logic/grab_self_serve` - Grab 自助激活服务

### 服务依赖

- **复用服务**: `ShopProviderCfg` - 门店第三方配置管理
- **复用服务**: `GrabSelfServe` - Grab 自助激活链接服务
- **数据库**: `shop_provider_cfg` 表

### 业务依赖

- Grab 平台配置已完成（ClientID、ClientSecret、Environment）
- `shop_provider_cfg` 表已存在
- Nacos 服务注册中心已部署

---

## 风险和缓解

### 风险 1: LINE MAN 渠道激活流程可能需要额外的验证逻辑

**影响**: 中  
**概率**: 中  
**缓解措施**:

- 与产品确认 LINE MAN 激活流程的完整需求
- 预留扩展点，便于后续增强 lineman 激活逻辑
- 当前实现保持简单，仅创建配置记录

### 风险 2: 多渠道扩展时可能需要调整路由策略

**影响**: 低  
**概率**: 中  
**缓解措施**:

- 设计可扩展的渠道注册机制（策略模式）
- 使用 switch-case 实现路由，便于后续新增渠道
- 统一错误处理和日志记录

### 风险 3: Grab 自助激活链接服务的错误处理需要传递到新服务

**影响**: 中  
**概率**: 低  
**缓解措施**:

- 统一错误处理，将 Grab API 错误正确映射到业务错误码
- 添加详细的错误日志
- 实现重试机制（如需要）

---

## 时间表

- **Phase 1 - Protobuf 定义和代码生成**: 0.5 天
  - 定义 `ActivateShopReq/Resp` 和 `GetShopProviderCfgReq/Resp` 消息
  - 定义 `Shop` 服务
  - 执行 `gf gen pb` 生成代码
  - 执行 `gf gen service` 生成服务接口

- **Phase 2 - Logic 层实现**: 0.5 天
  - 实现 `shop_activate` Logic（多渠道路由）
  - 扩展 `shop_provider_cfg` Logic（新增 `GetShopProviderCfgForRPC` 方法）
  - 添加常量 `ProviderLineman`

- **Phase 3 - Controller 层和测试**: 0.5 天
  - 实现 `ActivateShop` Controller
  - 实现 `GetShopProviderCfg` Controller
  - 单元测试和集成测试

- **Phase 4 - 联调和文档**: 0.5 天
  - 前后端联调
  - 完善技术文档
  - 代码审查

- **总计**: 2 天（SP = 3）

---

## 参考资料

### 核心规范

- `ttpos-bmp/.cursor/rules/go-rules.mdc` - Go BMP 开发规范
- `ttpos-bmp/.cursor/rules/proto-rules.mdc` - Protobuf 开发规范
- `.cursor/rules/go-ttpos-takeout` - ttpos-takeout 子模块规则
- `.cursor/rules/api.mdc` - API 设计规范
- `.cursor/rules/database.mdc` - 数据库开发规范
- `.cursor/rules/security.mdc` - 安全开发规范
- `.cursor/rules/structs.mdc` - 项目结构规范

### 现有实现参考

- `ttpos-bmp/app/ttpos-takeout/internal/logic/grab_self_serve/grab_self_serve.go` - Grab 自助激活实现
- `ttpos-bmp/app/ttpos-takeout/internal/logic/shop_provider_cfg/shop_provider_cfg.go` - 门店配置管理服务
- `ttpos-bmp/app/ttpos-takeout/manifest/protobuf/grab/grab.proto` - Grab 服务定义
- `ttpos-bmp/app/ttpos-takeout/manifest/protobuf/takeout/takeout.proto` - 通用消息定义

### 架构文档

- `ttpos-bmp/README.md` - BMP 模块说明
- `ttpos-bmp/MIGRATION_QUICK_START.md` - 数据库迁移快速入门
- `ttpos-bmp/DATABASE_MIGRATION_RULES.md` - 数据库迁移规则

### 外部参考

- [GoFrame 官方文档](https://goframe.org.cn)
- [Protocol Buffers 文档](https://protobuf.dev)
- [gRPC Go 文档](https://grpc.io/docs/languages/go/)

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/rikugun/2026-01/2026-01-07.md`
- 提醒：需求评审或范围调整若形成经验，应同步更新 Episode 并在 Proposal/Spec 互链。

---

**版本**: v1.0.0  
**创建日期**: 2026-01-07  
**作者**: rikugun  
**审核者**: 待指定

