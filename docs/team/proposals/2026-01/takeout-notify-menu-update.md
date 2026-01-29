# 外卖菜单更新通知服务 需求提案

> 本文档用于需求发起阶段,经团队评审后创建正式 Spec。

---

## 📋 提案信息

| 项目       | 内容     |
| ---------- | -------- |
| **提案人** | rikugun   |
| **日期**   | 2026-01-12   |
| **目标版本** | v2.13.2 |
| **状态**   | 已转 Spec   |
| **关联任务** | - |
| **关联 Spec** | [story-takeout-notify-menu-update](../../shared/specs/active/story-takeout-notify-menu-update/requirements.md)      |

---

## 🎯 背景和动机

### 问题描述

当前外卖菜单同步功能分散在各个 provider 的实现中，缺少统一的菜单更新通知入口。需要为不同的外卖平台（Grab、Lineman 等）提供统一的菜单更新通知接口，以便：

1. **统一入口**：Main 模块或其他服务可以通过统一的 gRPC 接口触发菜单同步
2. **灵活路由**：根据 provider 类型自动路由到对应的菜单同步实现
3. **扩展性**：未来接入新的外卖平台时，只需添加新的 case 分支即可

目前的问题：
- 缺少统一的菜单更新通知接口
- Main 模块或其他服务需要知道具体的 provider 实现细节
- 各 provider 的菜单同步逻辑入口不统一

### 业务价值

- **简化调用**：调用方只需知道 shop_uuid 和 provider_name，无需关心具体实现
- **降低耦合**：隔离各 provider 的菜单同步实现细节
- **提升可维护性**：统一的接口便于后续功能迭代和问题排查
- **支持扩展**：便于快速接入新的外卖平台

### 目标用户

- [x] 后端开发者（Main 模块调用方）
- [x] BMP 服务开发者
- [ ] 收银员
- [ ] 商户管理员
- [ ] 厨房人员
- [ ] 顾客

---

## 💡 解决方案概述

### 方案描述

在 `ttpos-bmp/app/ttpos-takeout/manifest/protobuf/menu/menu.proto` 中新增 `NotifyMenuUpdate` 方法，提供统一的菜单更新通知接口。

**实现逻辑**：
1. 接收参数：`shop_uuid`（店铺 UUID）和 `provider_name`（平台名称）
2. 根据 `provider_name` 路由到对应的实现：
   - 当 `provider_name = "grab"` 时，调用 `service.Grab().NotifyMenuUpdate`
   - 当 `provider_name = "lineman"` 时，调用 `service.Lineman().SyncMenu`
3. 返回统一的响应格式（使用 `takeout.ApiResponse`）

### 核心功能点

1. **新增 Proto 定义**：在 `menu.proto` 中定义 `NotifyMenuUpdate` RPC 方法
2. **请求参数设计**：
   - `shop_uuid`：店铺唯一标识
   - `provider_name`：外卖平台标识（grab/lineman）
   - `request_id`：可选的请求追踪 ID
3. **路由逻辑实现**：根据 provider_name 分发到对应的服务实现
4. **统一响应格式**：使用 `takeout.ApiResponse` 包装响应

### 影响范围

**涉及终端**：
- [ ] POS 收银端
- [x] Shop 商家管理端（可能触发菜单同步）
- [ ] KDS 厨显端
- [ ] QDS 排号叫号端
- [ ] Assistant 助手端
- [ ] Tablet 平板端
- [ ] Mobile 扫码端
- [ ] Menu 电子菜单端
- [ ] Member 会员端

**涉及模块**：
- [ ] UI 组件
- [x] API 接口（gRPC）
- [x] 数据模型（Protobuf）
- [x] 业务逻辑（菜单同步路由）
- [x] 第三方集成（Grab、Lineman）
- [ ] 其他: ________

**涉及文件**：
- `ttpos-bmp/app/ttpos-takeout/manifest/protobuf/menu/menu.proto`
- `ttpos-bmp/app/ttpos-takeout/internal/controller/rpc/menu/` (新增控制器)
- `ttpos-bmp/app/ttpos-takeout/internal/service/menu.go` (新增或修改)
- `ttpos-bmp/app/ttpos-takeout/internal/service/grab.go` (已有 NotifyMenuUpdate)
- `ttpos-bmp/app/ttpos-takeout/internal/service/lineman.go` (已有 SyncMenu)

---

## 📊 初步评估

### 技术复杂度

- [ ] **低**：纯 UI 调整，无业务逻辑变更
- [x] **中**：需要前后端联调，基础业务逻辑
- [ ] **高**：涉及架构调整、第三方集成、复杂算法

**说明**：
- Protobuf 定义较简单
- 路由逻辑清晰（switch-case）
- 主要工作是适配已有的 Grab 和 Lineman 服务接口

### 工作量预估

- **预计天数**: 1 天
- **预估 SP**: 2-3（待技术评审确认）

**工作内容**：
1. 修改 `menu.proto`，新增 `NotifyMenuUpdate` 定义
2. 生成 Protobuf 代码（`make proto`）
3. 实现 gRPC 控制器和路由逻辑
4. 单元测试（路由逻辑 + 各 provider 调用）
5. 集成测试（Main 模块调用验证）

### 风险识别

**潜在风险**：
1. **Grab 和 Lineman 的接口签名不一致**：`NotifyMenuUpdate` vs `SyncMenu`
2. **错误处理不统一**：两个 provider 的错误返回格式可能不同
3. **并发调用**：同一店铺的多次菜单更新请求可能冲突

**缓解措施**：
1. **接口适配层**：在路由逻辑中统一适配两个 provider 的接口差异
2. **统一错误处理**：将各 provider 的错误转换为标准的 `ApiResponse` 格式
3. **幂等性设计**：确保菜单同步操作支持幂等（通过 request_id 去重）

---

## 🔗 相关资源

### 参考实现

- Grab 菜单更新：`ttpos-bmp/app/ttpos-takeout/internal/controller/rpc/grab/grab_v1_notify_menu_update.go`
- Lineman 菜单同步：`ttpos-bmp/app/ttpos-takeout/internal/logic/lineman/menu_sync.go`
- 现有 Menu Service：`ttpos-bmp/app/ttpos-takeout/manifest/protobuf/menu/menu.proto`

### 相关文档

- BMP 开发规范: `.cursor/rules/go-bmp.mdc`
- Protobuf 规范: `ttpos-bmp/.cursor/rules/proto-rules.mdc`
- API 设计规范: `.cursor/rules/api.mdc`

---

## 🤝 需求评审

### 评审参与人

| 角色         | 姓名   | 签名/日期 |
| ------------ | ------ | --------- |
| 产品经理     |  |           |
| 技术负责人   |  |           |
| 开发代表     | rikugun |           |
| 测试代表     |  |           |
| UI/UX 设计师 | N/A |           |

### 评审结论

- [ ] ✅ **批准**：进入技术方案设计阶段
- [ ] 🔄 **修改后批准**：需补充以下内容
- [ ] ❌ **拒绝**：不符合产品规划或优先级

**评审意见**：

```
[记录评审会议的关键讨论和决策]
```

**下一步行动**：

- [ ] 创建 Spec：`story-takeout-notify-menu-update`
- [ ] 分配负责人：rikugun
- [ ] 目标 Sprint：Sprint N

---

## 📝 附录

### User Story（初稿）

**作为** 后端开发者（Main 模块调用方）  
**我想** 通过统一的 gRPC 接口通知外卖平台菜单更新  
**以便于** 无需关心各平台的具体实现细节，简化菜单同步逻辑

### AC 验收标准（初稿）

1. **WHEN** 调用 `NotifyMenuUpdate` 并指定 `provider_name="grab"` **THEN** 系统 **SHALL** 调用 `service.Grab().NotifyMenuUpdate` 并返回结果
2. **WHEN** 调用 `NotifyMenuUpdate` 并指定 `provider_name="lineman"` **THEN** 系统 **SHALL** 调用 `service.Lineman().SyncMenu` 并返回结果
3. **IF** `provider_name` 不是已知的平台（非 grab/lineman）**THEN** 系统 **SHALL** 返回错误 `INVALID_PROVIDER`
4. **WHEN** 菜单同步成功 **THEN** 系统 **SHALL** 返回 `code=0` 的 `ApiResponse`
5. **WHEN** 菜单同步失败 **THEN** 系统 **SHALL** 返回包含错误信息的 `ApiResponse`

### 技术设计要点

#### Protobuf 定义

```protobuf
// 通知菜单更新请求
message NotifyMenuUpdateReq {
  string shop_uuid = 1;      // 店铺 UUID (必填)
  string provider_name = 2;  // 平台名称: grab, lineman (必填)
  string request_id = 3;     // 请求 ID (可选，用于追踪)
}

// 通知菜单更新响应
message NotifyMenuUpdateResp {
  string sync_status = 1;   // 同步状态: QUEUED, PROCESSING, SUCCESS, FAIL
  int64 sync_time = 2;      // 同步时间戳
}

service MenuService {
    // ... 已有方法 ...
    
    // 通知菜单更新（统一入口）
    rpc NotifyMenuUpdate (NotifyMenuUpdateReq) returns (takeout.ApiResponse) {}
}
```

#### 路由逻辑伪代码

```go
func (s *MenuController) NotifyMenuUpdate(ctx context.Context, req *menu.NotifyMenuUpdateReq) (*takeout.ApiResponse, error) {
    switch req.ProviderName {
    case "grab":
        return s.service.Grab().NotifyMenuUpdate(ctx, req.ShopUuid, req.RequestId)
    case "lineman":
        return s.service.Lineman().SyncMenu(ctx, req.ShopUuid, req.RequestId)
    default:
        return common.RespError(errors.New("invalid provider: " + req.ProviderName))
    }
}
```

---

**版本**: v1.0.0  
**创建日期**: 2026-01-12  
**维护者**: rikugun  
**相关规范**: `.cursor/rules/go-bmp.mdc`, `.cursor/rules/api.mdc`, `ttpos-bmp/.cursor/rules/proto-rules.mdc`
