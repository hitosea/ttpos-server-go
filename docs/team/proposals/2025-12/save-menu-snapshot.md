# SaveMenuSnapshot 菜单快照保存 需求提案

> 本文档用于需求发起阶段，经团队评审后创建正式 Spec。

---

## 📋 提案信息

| 项目       | 内容     |
| ---------- | -------- |
| **提案人** | rikugun   |
| **日期**   | 2025-12-11   |
| **目标版本** | - |
| **状态**   | 待评审   |
| **关联任务** | - |
| **关联 Spec** | -      |

---

## 🎯 背景和动机

### 问题描述

当前 `TakeoutService` 只有 `GetMenuSnapshot` 方法用于查询菜单快照，但缺少对应的保存方法。外部渠道（如 Grab、Lineman）在推送菜单更新时，需要有一个统一的 gRPC 接口来保存菜单快照数据。

### 业务价值

- 提供完整的菜单快照 CRUD 能力
- 支持外部渠道菜单数据的接收和存储
- 为后续菜单同步、对账等功能提供数据基础
- 与现有 `GetMenuSnapshot` 方法形成配套

### 目标用户

- [x] 其他: 外部渠道集成服务（Grab、Lineman 等）

---

## 💡 解决方案概述

### 方案描述

在 `takeout.proto` 中新增 `SaveMenuSnapshot` RPC 方法，用于保存外部渠道推送的菜单快照数据。请求参数包含渠道名称、店铺 UUID、请求 ID 和菜单数据（JSON 字符串格式）。

### 核心功能点

1. 新增 `SaveMenuSnapshotReq` 消息定义
2. 新增 `SaveMenuSnapshotResp` 消息定义
3. 在 `TakeoutService` 中新增 `SaveMenuSnapshot` RPC 方法
4. **保存后触发 Grab 菜单更新通知**：当 `provider_name == "grab"` 时，调用 Grab API 通知菜单已更新

### Proto 定义

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

### 业务流程

```
SaveMenuSnapshot 请求
        ↓
保存菜单数据到 channel_menu_snapshot 表
        ↓
判断 provider_name == "grab"?
        ↓ Yes
调用 Grab Update Menu Notification API
  POST https://partner-api.grab.com/grabfood/partner/v1/merchant/menu/notification
  Body: { "merchantID": "<grab_merchant_id>" }
        ↓
返回响应
```

### Grab API 集成

当 `provider_name == "grab"` 时，需要调用 Grab 的 **Update menu notification** 接口：

- **Endpoint**: `POST /partner/v1/merchant/menu/notification`
- **文档**: https://developer.grab.com/docs/grabfood/api/v1-1-3/#tag/update-menu-notification
- **认证**: Bearer Token（通过现有 Grab OAuth 机制获取）
- **请求体**:
  ```json
  {
    "merchantID": "<grab_merchant_id>"
  }
  ```
- **说明**: 
  - 需要根据 `shop_uuid` 获取对应的 Grab `merchantID`
  - 复用现有 `grab.go` 中的 Grab SDK 认证逻辑
  - 注意 Grab 有 120 秒分布式锁机制，避免短时间内重复调用

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
- [x] 其他: 内部服务调用（ttpos-bmp 模块间）

**涉及模块**：
- [ ] UI 组件
- [x] API 接口
- [x] 数据模型
- [x] 业务逻辑
- [x] 第三方集成
- [ ] 其他: ________

---

## 📊 初步评估

### 技术复杂度

- [ ] **低**：纯 UI 调整，无业务逻辑变更
- [x] **中**：需要前后端联调，基础业务逻辑
- [ ] **高**：涉及架构调整、第三方集成、复杂算法

### 工作量预估

- **预计天数**: 1-2 天
- **预估 SP**: 3（待技术评审确认）

### 风险识别

**潜在风险**：
1. 菜单数据量较大时可能影响 gRPC 传输性能
2. 需要与现有 channel_menu_snapshot 表结构对齐
3. Grab API 有 120 秒分布式锁，频繁调用会返回 409 错误
4. 需要根据 shop_uuid 映射到 Grab merchantID

**缓解措施**：
1. 设置合理的 gRPC message size 限制
2. 复用现有 DAO 层逻辑
3. 在业务层控制调用频率，或记录调用状态避免重复通知
4. 通过 shop_provider_cfg 表获取 Grab merchantID 映射

---

## 🔗 相关资源

### 参考需求

- 现有 GetMenuSnapshot 方法: `ttpos-bmp/app/ttpos-takeout/manifest/protobuf/takeout/takeout.proto`
- 菜单快照存储提案: `docs/team/proposals/2025-12/takeout-channel-menu-storage.md`

### 相关文档

- Proto 规范: `ttpos-bmp/.cursor/rules/proto-rules.mdc`

---

## 🤝 需求评审

### 评审参与人

| 角色         | 姓名   | 签名/日期 |
| ------------ | ------ | --------- |
| 产品经理     |  |           |
| 技术负责人   |  |           |
| 开发代表     |  |           |
| 测试代表     |  |           |
| UI/UX 设计师 | N/A |           |

### 评审结论

- [ ] ✅ **批准**：进入技术方案设计阶段
- [ ] 🔄 **修改后批准**：需补充以下内容
- [ ] ❌ **拒绝**：不符合产品规划或优先级

**评审意见**：

```
[待填写]
```

**下一步行动**：

- [ ] 创建 Spec：`story-takeout-save-menu-snapshot`
- [ ] 分配负责人：
- [ ] 目标 Sprint：

---

## 📝 附录

### User Story（初稿）

**作为** 外送渠道集成服务  
**我想** 通过 gRPC 接口保存菜单快照数据  
**以便于** 后续查询和同步菜单信息

### AC 验收标准（初稿）

1. **WHEN** 调用 SaveMenuSnapshot 并提供有效参数 **THEN** 系统 **SHALL** 保存菜单快照并返回成功响应
2. **IF** provider_name 或 shop_uuid 为空 **THEN** 系统 **SHALL** 返回参数错误
3. **WHEN** provider_name == "grab" 且保存成功 **THEN** 系统 **SHALL** 调用 Grab Update Menu Notification API 通知菜单更新
4. **IF** Grab API 调用失败 **THEN** 系统 **SHALL** 记录错误日志但不影响 SaveMenuSnapshot 主流程返回成功

---

**版本**: v1.0.0  
**创建日期**: 2025-12-11  
**维护者**: rikugun

