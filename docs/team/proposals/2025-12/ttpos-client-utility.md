# TTPOS HTTP Client 工具类抽取 需求提案

> 本文档用于需求发起阶段，经团队评审后创建正式 Spec。

---

## 📋 提案信息

| 项目       | 内容     |
| ---------- | -------- |
| **提案人** | rikugun   |
| **日期**   | 2025-12-18   |
| **目标版本** | - |
| **状态**   | ✅ 已批准   |
| **关联任务** | - |
| **关联 Spec** | [task-bmp-ttpos-client-utility](../../../shared/specs/archived/v2.12/task-bmp-ttpos-client-utility/requirements.md) |

---

## 🎯 背景和动机

### 问题描述

当前 `ttpos-takeout` 模块中 `grab_menu.go` 的 `fetchMenuFromTTpos` 方法直接使用 `g.Client()` 创建 HTTP 客户端，包含以下硬编码逻辑：

```go
// 当前实现 (grab_menu.go:124-129)
client := g.Client().Timeout(10 * time.Second)
resp, err := client.
    SetHeader(consts.TTPOS_HEADER_SECRET, auth).
    ContentJson().
    Post(ctx, url, reqBody)
```

存在的问题：
1. **重复代码**：每次调用 TTPOS API 都需要重复设置超时、认证头、Content-Type
2. **不一致性**：不同调用点可能使用不同的超时配置
3. **维护困难**：如需修改认证逻辑或添加通用中间件，需要修改多处代码

### 业务价值

- 统一 TTPOS API 调用方式，减少代码重复
- 便于添加统一的日志、监控、重试等中间件
- 参考 `ttpos-erp` 模块的 `GetClient` 模式，保持项目一致性

### 目标用户

- [x] 开发人员

---

## 💡 解决方案概述

### 方案描述

在 `ttpos-bmp/app/ttpos-takeout/utility` 包中新增 `ttpos_client.go`，提供统一的 TTPOS HTTP Client 工厂方法，参考 `ttpos-erp` 模块的 `GetClient` 实现模式。

### 核心功能点

1. **GetTtposClient**: 创建预配置的 TTPOS HTTP Client
   - 自动设置 `app.ttposEndpoint` 为 prefix
   - 自动设置超时时间（可配置，默认 10s）
   - 自动设置 `ContentJson()`
   
2. **GetTtposClientWithAuth**: 创建带认证头的 Client
   - 包含 GetTtposClient 的所有功能
   - 自动生成并设置 `X-TTPOS-SECRET` 认证头

### 影响范围

**涉及终端**：
- [x] 其他: ttpos-takeout 服务端

**涉及模块**：
- [x] 业务逻辑
- [x] 第三方集成（内部服务调用）

---

## 📊 初步评估

### 技术复杂度

- [x] **低**：纯工具类封装，无业务逻辑变更

### 工作量预估

- **预计天数**: 0.5 天
- **预估 SP**: 1（待技术评审确认）

### 风险识别

**潜在风险**：
1. 无明显风险

**缓解措施**：
1. 保持向后兼容，原有代码可逐步迁移

---

## 🔗 相关资源

### 参考需求

- 参考实现: `ttpos-bmp/app/ttpos-erp/internal/logic/erpnext/erpnext.go` 中的 `GetClient`

### 相关文档

- 当前实现: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab_menu/grab_menu.go` 第 100-153 行

---

## 🤝 需求评审

### 评审参与人

| 角色         | 姓名   | 签名/日期 |
| ------------ | ------ | --------- |
| 技术负责人   |        |           |
| 开发代表     |        |           |

### 评审结论

- [ ] ✅ **批准**：进入技术方案设计阶段
- [ ] 🔄 **修改后批准**：需补充以下内容
- [ ] ❌ **拒绝**：不符合产品规划或优先级

**评审意见**：

```
[记录评审会议的关键讨论和决策]
```

**下一步行动**：

- [ ] 创建工具类文件
- [ ] 重构 `fetchMenuFromTTpos` 使用新工具类

---

## 📝 附录

### 代码示例

**新建 `utility/ttpos_client.go`**：

```go
package utility

import (
    "context"
    "net/http"
    "time"

    "github.com/gogf/gf/v2/frame/g"
    "github.com/gogf/gf/v2/net/gclient"
    "github.com/gogf/gf/v2/os/gctx"

    "ttpos-bmp/app/ttpos-takeout/internal/consts"
)

const defaultTimeout = 10 * time.Second

// GetTtposClient 获取预配置的 TTPOS HTTP Client
// 自动设置 endpoint prefix、超时、ContentJson
// 当 app.ttpos-client.dump 配置为 true 时，自动打印请求/响应详情
func GetTtposClient(ctx context.Context) *gclient.Client {
    c := g.Client()
    
    // 设置 prefix
    ttposEndpoint := g.Cfg().MustGet(gctx.GetInitCtx(), "app.ttposEndpoint").String()
    if ttposEndpoint != "" {
        c.SetPrefix(ttposEndpoint)
    }
    
    // 设置超时和 Content-Type
    c.Timeout(defaultTimeout)
    c.ContentJson()
    
    // 添加 dump 中间件
    c.Use(func(c *gclient.Client, r *http.Request) (resp *gclient.Response, err error) {
        resp, err = c.Next(r)
        if resp != nil && g.Cfg().MustGet(gctx.GetInitCtx(), "app.ttpos-client.dump").Bool() {
            resp.RawDump()
        }
        return resp, err
    })
    
    return c
}

// GetTtposClientWithAuth 获取带认证头的 TTPOS HTTP Client
// identifier: 用于生成认证头的标识符（如 shopUUID）
func GetTtposClientWithAuth(ctx context.Context, identifier string) (*gclient.Client, error) {
    c := GetTtposClient(ctx)
    
    // 生成并设置认证头
    auth, err := GenerateTtposAuth(identifier)
    if err != nil {
        return nil, err
    }
    c.SetHeader(consts.TTPOS_HEADER_SECRET, auth)
    
    return c, nil
}
```

**重构后的 `fetchMenuFromTTpos`**：

```go
func (s *sGrabMenu) fetchMenuFromTTpos(ctx context.Context, shopUUID uint64) (*grabfood.GetMenuNewResponse, error) {
    // 1. 获取带认证的 Client
    client, err := utility.GetTtposClientWithAuth(ctx, fmt.Sprintf("%d", shopUUID))
    if err != nil {
        return nil, gerror.Wrap(err, "failed to create TTPOS client")
    }

    // 2. 构建请求体
    reqBody := g.Map{
        "platform":     string(consts.ProviderGrab),
        "company_uuid": shopUUID,
    }

    // 3. 发起请求
    resp, err := client.Post(ctx, "/api/v1/takeout/menu/export", reqBody)
    if err != nil {
        return nil, gerror.Wrap(err, "failed to call TTPOS export API")
    }
    defer resp.Close()

    // ... 后续解析逻辑不变
}
```

---

**版本**: v1.0.0  
**创建日期**: 2025-12-18  
**维护者**: 开发组


