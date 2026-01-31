# LINE MAN 菜单同步通知入库设计文档

> 本文档定义 LINE MAN 菜单同步通知入库功能的技术设计和实现方案。

## 📋 概述

本功能实现 MenuSyncNotification 接口，接收 LINE MAN 菜单同步完成后的通知回调。当 LINE MAN 平台调用该接口时，系统将：
1. 解析请求参数（partnerId、storeId、menuSyncRequestId、status、error）
2. 记录同步通知到 `menu_log`（`sync_type=NOTIFY`，`status=SUCCESS/FAIL`）
3. 返回标准化响应（避免第三方重试）

该功能属于 **ttpos-takeout** 微服务，采用 GoFrame 2.x 框架实现。

---

## 🎯 规范对齐

### Go BMP 规范 (go-bmp.mdc)

本设计严格遵循 Go BMP 开发规范：

- ✅ 使用 GoFrame 2.x 框架
- ✅ Controller → Logic → Service → DAO 分层架构
- ✅ **禁止修改** `dao/`, `model/entity/`, `model/do/` 目录（自动生成）
- ✅ 使用 `gerror` 处理错误（不用标准库 errors）
- ✅ 使用 `g.Log()` 记录日志
- ✅ 业务逻辑写在 `internal/logic/` 目录
- ✅ 数据库操作通过 `dao` 层

### API 设计规范 (api.mdc)

符合 Lineman API 定义规范：

- ✅ URL: `POST /v1/partners/{partnerId}/stores/{storeId}/menus/notification`
- ✅ 响应格式: `{ "status": "ok/fail", "code": "string", "message": "string" }`
- ✅ 标准 HTTP 状态码: 200/400/404/500
- ✅ 参考协议: [Lineman API 定义及 TTPOS 映射](https://docs.google.com/spreadsheets/d/1CKRl7tRLtp6dCAcXQqWhPvS_0M378-vdKpucR6ZtNbg/edit?gid=571121603#gid=571121603)

### 数据库规范 (database.mdc)

复用现有 `menu_log` 表，无需新建表：

- ✅ 表名: `takeout_menu_log`（已存在）
- ✅ 必需字段: `id`, `uuid`, `create_time`, `update_time`, `delete_time`
- ✅ 时间字段使用 int 类型
- ✅ 通过 DAO 层操作数据库

---

## 🔄 代码复用分析

### 可复用的现有组件

1. **API 定义**
   - 路径: `api/lineman/v1/menu.go`
   - 结构体: `MenuSyncNotificationReq`, `MenuSyncNotificationRes`
   - 状态: ✅ 已定义

2. **Controller 骨架**
   - 路径: `internal/controller/lineman/lineman_v1_menu_sync_notification.go`
   - 状态: ✅ 已存在骨架，已实现

3. **Logic 层实现**
   - 路径: `internal/logic/lineman/menu_sync_notification.go`
   - 方法: `HandleMenuSyncNotification()`
   - 状态: ✅ 已实现

4. **ChannelMenu Service**
   - 路径: `internal/service/channel_menu.go`
   - 方法: `LogMenuSync()` - 记录菜单同步日志
   - 状态: ✅ 已实现，直接调用

5. **常量定义**
   - 路径: `internal/consts/consts.go`
   - 常量: `ProviderLineman`, `MenuSyncTypeNotify`（新增）
   - 状态: ⚠️ 需新增 `MenuSyncTypeNotify`

6. **数据库表**
   - 表名: `takeout_menu_log`
   - 状态: ✅ 已存在

### 集成点

- **Lineman API 路由**: 已注册在 `api/lineman/v1/menu.go`
- **Controller 入口**: ✅ 已实现
- **数据库表**: `takeout_menu_log` 已存在
- **Service 依赖**: `service.ChannelMenu()` 已注册

---

## 🏗️ 架构设计

### 分层设计原则

**Go BMP 四层架构**:

```
HTTP Controller 层
  ↓ 依赖
Logic 层（业务编排）
  ↓ 依赖
Service 层（核心业务）
  ↓ 依赖
DAO 层（数据访问，自动生成 ❌ 禁止修改）
```

### 调用链路

```
Lineman Platform
  ↓ HTTP POST
┌─────────────────────────────────────────────────┐
│ Controller: MenuSyncNotification()              │
│ - 解析路径参数 (partnerId, storeId)             │
│ - 解析请求体 (menuSyncRequestId, status, error) │
│ - 类型断言调用 Logic 层                         │
└─────────────────────────────────────────────────┘
  ↓
┌─────────────────────────────────────────────────┐
│ Logic: HandleMenuSyncNotification()             │
│ - 参数校验（必填字段、状态枚举）                 │
│ - 状态映射（SUCCESS → SUCCESS, FAILED → FAIL）  │
│ - 调用 ChannelMenu.LogMenuSync() 写入日志       │
└─────────────────────────────────────────────────┘
  ↓
┌─────────────────────────────────────────────────┐
│ Service: ChannelMenu.LogMenuSync()              │
│ - 生成 UUID                                     │
│ - 构建 menu_log 记录                            │
│ - 调用 DAO 层写入数据库                          │
└─────────────────────────────────────────────────┘
  ↓
┌─────────────────────────────────────────────────┐
│ DAO: dao.MenuLog.Ctx(ctx).Data(...).Insert()   │
│ - ORM 操作（自动生成，禁止修改）                 │
└─────────────────────────────────────────────────┘
  ↓
MySQL: takeout_menu_log 表
```

---

## 📦 模块设计

### 1. Controller 层

**文件**: `ttpos-bmp/app/ttpos-takeout/internal/controller/lineman/lineman_v1_menu_sync_notification.go`

**职责**:
- 解析 HTTP 请求参数
- 调用 Logic 层处理业务
- 构建 HTTP 响应

**实现要点**:
```go
func (c *ControllerV1) MenuSyncNotification(ctx context.Context, req *v1.MenuSyncNotificationReq) (res *v1.MenuSyncNotificationRes, err error) {
    // 1. 类型断言调用 Logic 层
    lineman, ok := service.Lineman().(interface {
        HandleMenuSyncNotification(context.Context, *v1.MenuSyncNotificationReq) error
    })
    
    // 2. 调用 Logic
    err = lineman.HandleMenuSyncNotification(ctx, req)
    
    // 3. 错误码映射
    if err != nil {
        if gerror.Code(err) == gcode.CodeInvalidParameter {
            // 返回 400
        }
        if gerror.Code(err) == gcode.CodeNotFound {
            // 返回 404
        }
        // 返回 500
    }
    
    // 4. 返回成功响应
    return &v1.MenuSyncNotificationRes{
        LinemanCommonResData: v1.LinemanCommonResData{
            Status:  "ok",
            Code:    "200",
            Message: "Menu sync notification received successfully",
        },
    }, nil
}
```

**错误处理**:
| 错误码 | HTTP 状态码 | 返回 status | 返回 code |
|--------|-------------|-------------|-----------|
| CodeInvalidParameter | 200 | fail | 400 |
| CodeNotFound | 200 | fail | 404 |
| CodeInternalError | 200 | fail | 500 |

**注意**: Lineman API 约定所有响应都返回 HTTP 200，通过 `status` 字段区分成功/失败。

---

### 2. Logic 层

**文件**: `ttpos-bmp/app/ttpos-takeout/internal/logic/lineman/menu_sync_notification.go`

**职责**:
- 参数校验
- 业务逻辑编排
- 调用 Service 层

**实现要点**:
```go
func (s *sLineman) HandleMenuSyncNotification(ctx context.Context, req *v1.MenuSyncNotificationReq) error {
    // 1. 参数校验
    if req.MenuSyncRequestId == "" {
        return gerror.NewCode(gcode.CodeInvalidParameter, "menuSyncRequestId 不能为空")
    }
    if req.Status == "" {
        return gerror.NewCode(gcode.CodeInvalidParameter, "status 不能为空")
    }
    if req.Status != "SUCCESS" && req.Status != "FAILED" {
        return gerror.NewCode(gcode.CodeInvalidParameter, "status 必须为 SUCCESS 或 FAILED")
    }
    
    // 2. 状态映射
    var menuLogStatus string
    if req.Status == "SUCCESS" {
        menuLogStatus = "SUCCESS"
    } else {
        menuLogStatus = "FAIL"
    }
    
    // 3. 构建错误信息
    var errorMsg string
    if req.Status == "FAILED" {
        errorMsg = req.Error
    }
    
    // 4. 调用 ChannelMenu Service 写入日志
    err := service.ChannelMenu().LogMenuSync(
        ctx,
        req.StoreId,                              // merchantID
        string(consts.ProviderLineman),           // providerName
        string(consts.MenuSyncTypeNotify),        // syncType
        req.MenuSyncRequestId,                    // requestID
        menuLogStatus == "SUCCESS",               // success
        "",                                       // menuSnapshot (空)
        errorMsg,                                 // errMsg
    )
    
    // 5. 错误处理
    if err != nil {
        g.Log().Errorf(ctx, "[Lineman] 记录菜单同步通知失败: storeId=%s, requestId=%s, status=%s, error=%v",
            req.StoreId, req.MenuSyncRequestId, req.Status, err)
        return gerror.WrapCode(gcode.CodeInternalError, err, "记录菜单同步通知失败")
    }
    
    // 6. 记录成功日志
    g.Log().Infof(ctx, "[Lineman] 菜单同步通知已记录: storeId=%s, requestId=%s, status=%s",
        req.StoreId, req.MenuSyncRequestId, req.Status)
    
    return nil
}
```

**参数校验规则**:
| 字段 | 校验规则 | 错误码 |
|------|----------|--------|
| menuSyncRequestId | 必填 | CodeInvalidParameter |
| status | 必填 | CodeInvalidParameter |
| status | 枚举 (SUCCESS/FAILED) | CodeInvalidParameter |
| error | 可选（status=FAILED 时记录） | - |

---

### 3. Service 层（复用）

**文件**: `ttpos-bmp/app/ttpos-takeout/internal/logic/channel_menu/channel_menu.go`

**方法**: `LogMenuSync()`

**参数**:
```go
func (s *sChannelMenu) LogMenuSync(
    ctx context.Context,
    merchantID string,      // storeId
    providerName string,    // "lineman"
    syncType string,        // "NOTIFY"
    requestID string,       // menuSyncRequestId
    success bool,           // true (SUCCESS) / false (FAILED)
    menuSnapshot string,    // "" (不需要)
    errMsg string,          // error 字段内容
) error
```

**复用说明**: 该方法已存在，无需修改。

---

### 4. 常量层（新增）

**文件**: `ttpos-bmp/app/ttpos-takeout/internal/consts/consts.go`

**新增常量**:
```go
const (
    // 已存在
    MenuSyncTypeFull                MenuSyncType = "FULL"
    MenuSyncTypeBatchUpdateItem     MenuSyncType = "BATCH_UPDATE_ITEM"
    MenuSyncTypeBatchUpdateModifier MenuSyncType = "BATCH_UPDATE_MODIFIER"
    
    // 新增 ✅
    MenuSyncTypeNotify              MenuSyncType = "NOTIFY"
)
```

---

## 📊 数据流设计

### 请求流

```
POST /v1/partners/{partnerId}/stores/{storeId}/menus/notification
Authorization: Bearer {token}
Content-Type: application/json

{
  "menuSyncRequestId": "req_123456",
  "updatedAt": "2022-11-01T13:08:06+07:00",
  "status": "SUCCESS",
  "error": ""
}
```

### 响应流（成功）

```
HTTP/1.1 200 OK
Content-Type: application/json

{
  "status": "ok",
  "code": "200",
  "message": "Menu sync notification received successfully"
}
```

### 响应流（失败）

```
HTTP/1.1 200 OK
Content-Type: application/json

{
  "status": "fail",
  "code": "400",
  "message": "menuSyncRequestId 不能为空"
}
```

### 数据库记录

**SUCCESS 场景**:
```sql
INSERT INTO takeout_menu_log (
  uuid, merchant_id, provider_name, sync_type, request_id, 
  status, menu_snapshot, error_msg, created_at, updated_at
) VALUES (
  1234567890, 'store123', 'lineman', 'NOTIFY', 'req_123456',
  'SUCCESS', '', '', 1704067200, 1704067200
);
```

**FAILED 场景**:
```sql
INSERT INTO takeout_menu_log (
  uuid, merchant_id, provider_name, sync_type, request_id, 
  status, menu_snapshot, error_msg, created_at, updated_at
) VALUES (
  1234567891, 'store123', 'lineman', 'NOTIFY', 'req_123456',
  'FAIL', '', 'Invalid menu format', 1704067200, 1704067200
);
```

---

## 🔐 安全设计

### 认证

- **机制**: Bearer Token（由网关/中间件处理）
- **Header**: `Authorization: Bearer {access_token}`
- **验证**: GoFrame 路由层自动校验

### 授权

- **校验**: `partnerId` 和 `storeId` 有效性
- **实现**: 由 Logic 层校验（可选，当前未实现）

### 审计

- **日志**: 所有请求记录到 `menu_log` 表
- **内容**: `merchant_id`, `request_id`, `status`, `error_msg`

---

## 📈 可观测性设计

### 日志规范

**成功日志**:
```go
g.Log().Infof(ctx, "[Lineman] 菜单同步通知已记录: storeId=%s, requestId=%s, status=%s",
    req.StoreId, req.MenuSyncRequestId, req.Status)
```

**失败日志**:
```go
g.Log().Errorf(ctx, "[Lineman] 记录菜单同步通知失败: storeId=%s, requestId=%s, status=%s, error=%v",
    req.StoreId, req.MenuSyncRequestId, req.Status, err)
```

### 监控指标

- **指标**: 无需新增（复用现有日志监控）
- **告警**: 失败率 > 10% 触发告警

---

## 🧪 测试策略

### 单元测试

**文件**: `ttpos-bmp/app/ttpos-takeout/internal/logic/lineman/menu_sync_notification_test.go`

**测试用例**:
1. 测试参数校验（必填字段、状态枚举）
2. 测试状态映射（SUCCESS → SUCCESS, FAILED → FAIL）
3. 测试错误信息记录（status=FAILED 时）
4. 测试 Service 调用失败场景

### 集成测试

**工具**: Postman / curl

**测试场景**:
1. 成功通知（status=SUCCESS）
2. 失败通知（status=FAILED, error="xxx"）
3. 参数缺失（400 错误）
4. 门店不存在（404 错误）

---

## 📝 部署清单

### 代码变更

- [x] `api/lineman/v1/menu.go` - API 定义（已存在）
- [x] `internal/controller/lineman/lineman_v1_menu_sync_notification.go` - Controller（已实现）
- [x] `internal/logic/lineman/menu_sync_notification.go` - Logic（已实现）
- [x] `internal/consts/consts.go` - 新增 `MenuSyncTypeNotify` 常量（已实现）

### 配置变更

- 无需配置变更

### 数据库变更

- 无需数据库变更（复用现有表）

### 依赖变更

- 无需依赖变更

---

## 🔄 回滚计划

**回滚策略**: 代码回滚

**回滚步骤**:
1. 回滚 Controller 代码
2. 回滚 Logic 代码
3. 回滚常量定义

**影响范围**: 无数据库变更，回滚无风险

---

## 📚 相关文档

- [Go BMP 开发规范](../../../../.cursor/rules/go-bmp.mdc)
- [API 设计规范](../../../../.cursor/rules/api.mdc)
- [数据库开发规范](../../../../.cursor/rules/database.mdc)
- [Lineman API 协议](https://docs.google.com/spreadsheets/d/1CKRl7tRLtp6dCAcXQqWhPvS_0M378-vdKpucR6ZtNbg/edit?gid=571121603#gid=571121603)
- [requirements.md](./requirements.md)

---

**版本**: v1.0.0  
**创建日期**: 2026-01-15  
**最后更新**: 2026-01-15
