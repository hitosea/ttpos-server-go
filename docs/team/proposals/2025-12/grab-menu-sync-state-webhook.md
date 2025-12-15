# Grab Menu Sync State Webhook 实现

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

当前 `grab_v1_menu_sync_state.go` Controller 只有空壳实现，返回 `NotImplemented` 错误：

```go
func (c *ControllerV1) MenuSyncState(ctx context.Context, req *v1.MenuSyncStateReq) (res *v1.MenuSyncStateRes, err error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}
```

而业务逻辑层 `MenuService.HandleMenuSyncState()` 已经完整实现，需要在 Controller 层完成调用链路。

### 业务价值

- 完成 GrabFood Menu Sync State Webhook 集成
- 接收 Grab 平台推送的菜单同步结果通知
- 实现菜单同步状态追踪（QUEUEING → PROCESSING → SUCCESS/FAILED）
- 支持 GrabFood Self-Serve Activation 全流程

### 目标用户

- [x] 商户管理员（通过外卖后台查看菜单同步状态）
- [x] 其他: Grab 平台（作为 Webhook 调用方）

---

## 💡 解决方案概述

### 方案描述

在 Controller 层调用已实现的 `service.Grab().HandleMenuSyncState()` 方法，完成 Webhook 处理链路。

参考 GrabFood API 文档：https://developer.grab.com/docs/grabfood/api/v1-1-3/#tag/menu-sync-state-webhook

### 核心功能点

1. 接收 Grab 菜单同步状态回调请求
2. 调用 Service 层处理业务逻辑（已实现）
3. 返回 HTTP 200/204 响应

### 现有实现分析

| 层级 | 文件 | 状态 |
|------|------|------|
| API 定义 | `api/grab/v1/menu_sync_state.go` | ✅ 已完成 |
| Service 接口 | `internal/service/grab.go` | ⚠️ 需确认 |
| 业务逻辑 | `internal/logic/grab/menu_service.go` | ✅ 已实现 `HandleMenuSyncState()` |
| Controller | `internal/controller/grab/grab_v1_menu_sync_state.go` | ❌ 未实现 |

### 影响范围

**涉及终端**：
- [ ] POS 收银端
- [x] Shop 商家管理端（外卖菜单管理）
- [ ] KDS 厨显端
- [ ] QDS 排号叫号端
- [ ] Assistant 助手端
- [ ] Tablet 平板端
- [ ] Mobile 扫码端
- [ ] Menu 电子菜单端
- [ ] Member 会员端

**涉及模块**：
- [ ] UI 组件
- [x] API 接口
- [ ] 数据模型
- [x] 业务逻辑
- [x] 第三方集成（GrabFood）

---

## 📊 初步评估

### 技术复杂度

- [x] **低**：纯 Controller 层调用，业务逻辑已实现
- [ ] **中**：需要前后端联调，基础业务逻辑
- [ ] **高**：涉及架构调整、第三方集成、复杂算法

### 工作量预估

- **预计天数**: 0.5 天
- **预估 SP**: 1（待技术评审确认）

### 实现要点

1. **Controller 层**：调用 `service.Grab().HandleMenuSyncState()`
2. **Service 注册**：确认 `HandleMenuSyncState` 已在 Service 接口中声明
3. **错误处理**：参考现有 Webhook 实现（记录日志但不中断流程）

### 风险识别

**潜在风险**：
1. Service 接口可能未声明 `HandleMenuSyncState` 方法

**缓解措施**：
1. 检查并补充 Service 接口定义

---

## 🔗 相关资源

### 参考需求

- GrabFood API 文档: https://developer.grab.com/docs/grabfood/api/v1-1-3/#tag/menu-sync-state-webhook

### 相关代码

- API 定义: `ttpos-bmp/app/ttpos-takeout/api/grab/v1/menu_sync_state.go`
- 业务逻辑: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab/menu_service.go`
- 参考实现: `ttpos-bmp/app/ttpos-takeout/internal/controller/grab/grab_v1_integration_status.go`

---

## 🤝 需求评审

### 评审结论

- [x] ✅ **批准**：直接实现（工作量 < 1 SP，无需创建 Spec）
- [ ] 🔄 **修改后批准**：需补充以下内容
- [ ] ❌ **拒绝**：不符合产品规划或优先级

**下一步行动**：

- [x] 直接实现 Controller 层代码
- [ ] 分配负责人：rikugun

---

## 📝 附录

### 预期实现代码

```go
// MenuSyncState 处理 Grab 菜单同步状态回调
// GrabFood 在菜单同步状态变化时调用此端点推送状态通知
// 参考: https://developer.grab.com/docs/grabfood/api/v1-1-3/#tag/menu-sync-state-webhook
func (c *ControllerV1) MenuSyncState(ctx context.Context, req *v1.MenuSyncStateReq) (res *v1.MenuSyncStateRes, err error) {
	// 调用 Service 处理菜单同步状态
	if err := service.Grab().HandleMenuSyncState(ctx, req.MenuSyncWebhookRequest); err != nil {
		g.Log().Errorf(ctx, "[Grab] HandleMenuSyncState failed: %v", err)
		// 记录错误但返回成功，避免 Grab 重试
	}

	// 返回 200 OK
	return &v1.MenuSyncStateRes{}, nil
}
```

---

**版本**: v1.0.0  
**创建日期**: 2025-12-11  
**维护者**: rikugun
