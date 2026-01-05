# 简化菜单更新接口响应结构 需求提案

> 本文档用于需求发起阶段，经团队评审后创建正式 Spec。

---

## 📋 提案信息

| 项目       | 内容     |
| ---------- | -------- |
| **提案人** | rikugun   |
| **日期**   | 2025-12-16   |
| **目标版本** | - |
| **状态**   | 已创建 Spec   |
| **关联任务** | - |
| **关联 Spec** | [story-bmp-takeout-proto-simplify-response](../../shared/specs/archived/v2.12/story-bmp-takeout-proto-simplify-response/requirements.md) |

---

## 🎯 背景和动机

### 问题描述

当前 `menu.proto` 中的 `UpdateMenuItemResp` 和 `UpdateMenuModifierResp` 定义了独立的 `error_code` 和 `error_message` 字段：

```protobuf
message UpdateMenuItemResp {
  bool success = 1;
  string merchant_id = 2;
  string record_id = 3;
  string record_type = 4;
  string error_code = 5;     // ← 冗余
  string error_message = 6;  // ← 冗余
}
```

这与 `MenuService` 的 RPC 方法设计存在冗余：

```protobuf
rpc UpdateMenuItem (UpdateMenuItemReq) returns (takeout.ApiResponse) {}
```

`takeout.ApiResponse` 已经包含了统一的 `code` 和 `message` 字段用于错误处理，因此在 Resp 结构中再定义 `error_code` 和 `error_message` 是重复的。

### 业务价值

- **减少数据冗余**：避免同一信息在多个字段中重复
- **统一响应格式**：所有 API 使用一致的错误处理模式
- **简化客户端解析**：客户端只需处理 `ApiResponse` 的 `code/message`，无需额外解析 Resp 中的错误字段
- **减少维护成本**：少两个字段意味着更少的文档和测试

### 目标用户

- [x] 后端开发人员
- [x] 前端/客户端开发人员
- [ ] 收银员
- [ ] 商户管理员
- [ ] 其他: ________

---

## 💡 解决方案概述

### 方案描述

移除 `UpdateMenuItemResp` 和 `UpdateMenuModifierResp` 中的 `error_code` 和 `error_message` 字段，统一使用 `takeout.ApiResponse` 的 `code` 和 `message` 进行错误信息传递。

调整后的结构：

```protobuf
message UpdateMenuItemResp {
  bool success = 1;          // 是否成功
  string merchant_id = 2;    // 商户ID
  string record_id = 3;      // 记录ID (ItemID)
  string record_type = 4;    // 记录类型: ITEM
  // 移除 error_code 和 error_message，由 ApiResponse 统一处理
}

message UpdateMenuModifierResp {
  bool success = 1;          // 是否成功
  string merchant_id = 2;    // 商户ID
  string record_id = 3;      // 记录ID (ModifierID)
  string record_type = 4;    // 记录类型: MODIFIER
  // 移除 error_code 和 error_message，由 ApiResponse 统一处理
}
```

### 核心功能点

1. 移除 `UpdateMenuItemResp.error_code` 和 `UpdateMenuItemResp.error_message`
2. 移除 `UpdateMenuModifierResp.error_code` 和 `UpdateMenuModifierResp.error_message`
3. 重新生成 proto 代码 (`gf gen pb`)
4. 调整相关 DTO 和逻辑代码（如有必要）

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
- [x] 后端服务 (ttpos-takeout)

**涉及模块**：
- [ ] UI 组件
- [x] API 接口 (protobuf 定义)
- [x] 数据模型 (DTO)
- [ ] 业务逻辑
- [ ] 第三方集成
- [ ] 其他: ________

---

## 📊 初步评估

### 技术复杂度

- [x] **低**：纯接口结构调整，无业务逻辑变更
- [ ] **中**：需要前后端联调，基础业务逻辑
- [ ] **高**：涉及架构调整、第三方集成、复杂算法

### 工作量预估

- **预计天数**: 0.5 天
- **预估 SP**: 1（待技术评审确认）

### 风险识别

**潜在风险**：
1. 如果有客户端已经依赖 `error_code/error_message` 字段，可能需要调整
2. 已生成的 Go 代码需要重新生成

**缓解措施**：
1. 确认当前无客户端依赖这些字段（功能刚开发完成，尚未发布）
2. 执行 `gf gen pb` 重新生成代码

---

## 🔗 相关资源

### 参考文件

- Proto 文件: `ttpos-bmp/app/ttpos-takeout/manifest/protobuf/menu/menu.proto`
- 关联 Spec: `docs/shared/specs/archived/v2.12/story-bmp-grab-menu-update-item-modifier/`
- DTO 定义: `ttpos-bmp/app/ttpos-takeout/internal/model/dto/grab/menu_update.go`

### 相关文档

- ApiResponse 定义: `ttpos-bmp/app/ttpos-takeout/manifest/protobuf/takeout_api.proto`

---

## 🤝 需求评审

### 评审参与人

| 角色         | 姓名   | 签名/日期 |
| ------------ | ------ | --------- |
| 技术负责人   | rikugun |           |
| 开发代表     | rikugun |           |

### 评审结论

- [ ] ✅ **批准**：进入技术方案设计阶段
- [ ] 🔄 **修改后批准**：需补充以下内容
- [ ] ❌ **拒绝**：不符合产品规划或优先级

**评审意见**：

```
[记录评审会议的关键讨论和决策]
```

**下一步行动**：

- [ ] 直接执行修改（改动较小，无需创建 Spec）
- [ ] 分配负责人：rikugun
- [ ] 目标 Sprint：当前 Sprint

---

## 📝 附录

### 修改前后对比

**修改前**:
```protobuf
message UpdateMenuItemResp {
  bool success = 1;
  string merchant_id = 2;
  string record_id = 3;
  string record_type = 4;
  string error_code = 5;     // 冗余
  string error_message = 6;  // 冗余
}
```

**修改后**:
```protobuf
message UpdateMenuItemResp {
  bool success = 1;
  string merchant_id = 2;
  string record_id = 3;
  string record_type = 4;
}
```

---

**版本**: v1.0.0  
**创建日期**: 2025-12-16  
**维护者**: rikugun

