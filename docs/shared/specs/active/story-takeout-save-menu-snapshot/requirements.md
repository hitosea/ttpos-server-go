# SaveMenuSnapshot 菜单快照保存 需求文档

> 本文档定义 SaveMenuSnapshot 菜单快照保存功能的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2025-12/save-menu-snapshot.md](../../../../team/proposals/2025-12/save-menu-snapshot.md) |
| **创建日期**      | 2025-12-11                                                                                                   |
| **负责人**        | rikugun                                                                                                      |
| **目标 Sprint**   | -                                                                                                            |
| **涉及技术栈**    | [x] Go (ttpos-bmp/)                                                                                          |

## 📋 审核状态

| 项目         | 内容     |
| ------------ | -------- |
| **审核状态** | 已审核   |
| **审核人**   | -        |
| **审核日期** | -        |
| **审核意见** | -        |

---

## 📋 概述

当前 `TakeoutService` 只有 `GetMenuSnapshot` 方法用于查询菜单快照，但缺少对应的保存方法。本功能新增 `SaveMenuSnapshot` gRPC 方法，用于接收和存储外部渠道（如 Grab、Lineman）推送的菜单快照数据，并在保存成功后触发相应渠道的菜单更新通知。

## 🎯 产品对齐

- 提供完整的菜单快照 CRUD 能力
- 支持外部渠道菜单数据的接收和存储
- 为后续菜单同步、对账等功能提供数据基础
- 与现有 `GetMenuSnapshot` 方法形成配套

## 📝 用户故事

**作为** 外送渠道集成服务  
**我想** 通过 gRPC 接口保存菜单快照数据  
**以便于** 后续查询和同步菜单信息

---

## 功能需求

### Requirement 1: SaveMenuSnapshot gRPC 接口

**用户故事**: 作为外送渠道集成服务，我想通过 gRPC 接口保存菜单快照，以便于统一管理渠道菜单数据

#### 验收标准

1. **WHEN** 调用 SaveMenuSnapshot 并提供有效参数 **THEN** 系统 **SHALL** 保存菜单快照到 channel_menu_snapshot 表并返回成功响应
2. **IF** provider_name 或 shop_uuid 为空 **THEN** 系统 **SHALL** 返回参数错误
3. **WHEN** 保存成功 **THEN** 系统 **SHALL** 更新快照的 updated_at 时间戳

#### 具体要求

- [ ] 1.1 在 `takeout.proto` 中新增 `SaveMenuSnapshotReq` 消息定义
- [ ] 1.2 在 `takeout.proto` 中新增 `SaveMenuSnapshotResp` 消息定义
- [ ] 1.3 在 `TakeoutService` 中新增 `SaveMenuSnapshot` RPC 方法
- [ ] 1.4 实现 Controller 层接收请求
- [ ] 1.5 实现 Logic 层业务逻辑
- [ ] 1.6 复用现有 DAO 层保存菜单快照

#### Proto 定义

```protobuf
message SaveMenuSnapshotReq {
  string provider_name = 1; // 渠道名称: grab,lineman
  string shop_uuid = 2;     // 店铺 UUID
  string menu_data = 3;     // 菜单数据 JSON 字符串
  string request_id = 4;    // 请求 ID
}

message SaveMenuSnapshotResp {
  ResponseInfo responseInfo = 1;
}

// 在 TakeoutService 中新增:
rpc SaveMenuSnapshot (SaveMenuSnapshotReq) returns (SaveMenuSnapshotResp) {}
```

---

### Requirement 2: Grab 菜单更新通知

**用户故事**: 作为系统，我想在保存 Grab 菜单快照后自动通知 Grab 菜单已更新，以便于 Grab 及时同步最新菜单

#### 验收标准

1. **WHEN** provider_name == "grab" 且保存成功 **THEN** 系统 **SHALL** 调用 Grab Update Menu Notification API 通知菜单更新
2. **IF** Grab API 调用失败 **THEN** 系统 **SHALL** 记录错误日志但不影响 SaveMenuSnapshot 主流程返回成功
3. **WHEN** 调用 Grab API **THEN** 系统 **SHALL** 使用现有 OAuth 认证机制获取 Bearer Token

#### 具体要求

- [ ] 2.1 根据 `shop_uuid` 从 `shop_provider_cfg` 获取 Grab `merchantID`
- [ ] 2.2 复用 `grab.go` 中的 OAuth 认证逻辑获取 access token
- [ ] 2.3 调用 Grab `POST /partner/v1/merchant/menu/notification` 接口
- [ ] 2.4 Grab API 调用失败时记录错误日志，不阻塞主流程
- [ ] 2.5 注意 Grab 120 秒分布式锁机制，避免频繁调用返回 409

#### Grab API 详情

- **Endpoint**: `POST https://partner-api.grab.com/grabfood/partner/v1/merchant/menu/notification`
- **文档**: https://developer.grab.com/docs/grabfood/api/v1-1-3/#tag/update-menu-notification
- **请求体**:
  ```json
  {
    "merchantID": "<grab_merchant_id>"
  }
  ```

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 遵循 Controller → Logic → DAO 分层
- **单一职责原则**: 每个文件应有单一、明确的目的
- **遵循规范**:
  - `ttpos-bmp/.cursor/rules/go-rules.mdc` - Go BMP 开发规范
  - `ttpos-bmp/.cursor/rules/proto-rules.mdc` - Protobuf 开发规范

### API 设计要求

- [x] gRPC 接口定义在 `takeout.proto`
- [x] 复用现有 `ResponseInfo` 响应结构
- [x] 参数校验在 Controller 层完成

### 性能要求

- [ ] gRPC 响应时间 < 500ms（含 Grab API 调用）
- [ ] 菜单数据量较大时设置合理的 message size 限制

### 测试要求

- [ ] Logic 层测试覆盖率 ≥ 70%
- [ ] 包含 Grab API 调用的 mock 测试

### 安全要求

- [ ] Grab API 调用使用 OAuth 2.0 认证
- [ ] 敏感配置（client_id, client_secret）从环境变量读取

---

## 验收标准

### 功能验收

1. **保存功能**: 调用 SaveMenuSnapshot 能正确保存菜单快照到数据库
2. **参数校验**: provider_name 或 shop_uuid 为空时返回错误
3. **Grab 通知**: provider_name == "grab" 时自动调用 Grab API
4. **错误处理**: Grab API 失败不影响主流程

### 测试验收

1. **单元测试**: 覆盖率达标
2. **集成测试**: gRPC 接口测试通过
3. **Mock 测试**: Grab API 调用 mock 测试通过

---

## 约束条件

### 技术约束

#### Go BMP 模块

- 必须使用 GoFrame 2.x
- 禁止修改 dao/entity/do/ 目录（自动生成）
- gRPC 服务注册到 Nacos
- 遵循 `ttpos-bmp/.cursor/rules/go-rules.mdc`

### 业务约束

- Grab API 有 120 秒分布式锁机制
- 需要有效的 shop_uuid 到 Grab merchantID 映射

### 资源约束

- 开发时间: 1-2 天
- Story Point: 3 (待技术评审确认)

---

## 依赖关系

### 技术依赖

- `channel_menu_snapshot` 表 - 菜单快照存储
- `shop_provider_cfg` 表 - 店铺渠道配置（获取 Grab merchantID）
- Grab OAuth 认证模块 - `grab.go`

### 服务依赖

- **TakeoutService**: 现有 gRPC 服务
- **Grab API**: 外部依赖

---

## 风险和缓解

### 风险 1: Grab API 调用频率限制

**影响**: 中  
**概率**: 中  
**缓解措施**:

- 在业务层控制调用频率
- 记录最近调用时间，避免 120 秒内重复调用
- API 返回 409 时记录日志但不报错

### 风险 2: 菜单数据量过大

**影响**: 低  
**概率**: 低  
**缓解措施**:

- 设置合理的 gRPC message size 限制
- 监控大数据量请求的性能

---

## 时间表

- **Phase 1 - Proto 定义和代码生成**: 0.5 天
- **Phase 2 - 业务逻辑实现**: 1 天
- **Phase 3 - 测试和联调**: 0.5 天
- **总计**: 2 天（SP = 3）

---

## 参考资料

### 核心规范

- `ttpos-bmp/.cursor/rules/go-rules.mdc` - Go BMP 开发规范
- `ttpos-bmp/.cursor/rules/proto-rules.mdc` - Protobuf 开发规范

### 相关文件

- `ttpos-bmp/app/ttpos-takeout/manifest/protobuf/takeout/takeout.proto` - 现有 Proto 定义
- `ttpos-bmp/app/ttpos-takeout/internal/logic/grab/grab.go` - Grab 业务逻辑

### 外部参考

- [Grab Update Menu Notification API](https://developer.grab.com/docs/grabfood/api/v1-1-3/#tag/update-menu-notification)

---

**版本**: v1.0.0  
**创建日期**: 2025-12-11  
**作者**: rikugun  
**审核者**: -
