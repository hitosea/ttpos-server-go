# TTPOS HTTP Client 工具类抽取 设计文档

> 本文档定义 TTPOS HTTP Client 工具类抽取的技术设计和实现方案。

## 📋 概述

在 `ttpos-bmp/app/ttpos-takeout/utility` 包中新增 `ttpos_client.go`，提供统一的 TTPOS HTTP Client 工厂方法，参考 `ttpos-erp` 模块的 `GetClient` 实现模式。

---

## 🎯 规范对齐

### Go BMP 规范 (go-bmp.mdc)

- 遵循 GoFrame 项目结构
- 工具类放置在 `utility/` 目录

---

## 🔄 代码复用分析

### 可复用的现有组件

- **GenerateTtposAuth**: `ttpos-bmp/app/ttpos-takeout/utility/ttpos_auth.go` - 认证头生成
- **GetClient 模式**: `ttpos-bmp/app/ttpos-erp/internal/logic/erpnext/erpnext.go` - 参考实现

### 集成点

- **fetchMenuFromTTpos**: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab_menu/grab_menu.go` - 重构使用新工具类

---

## 🏗️ 架构设计

### 模块划分

```
ttpos-bmp/app/ttpos-takeout/utility/
├── ttpos_auth.go        # 现有：认证头生成
├── ttpos_auth_test.go   # 现有：认证测试
└── ttpos_client.go      # 新增：HTTP Client 工厂
```

---

## 🧩 组件和接口

### GetTtposClient

```go
// ttpos-bmp/app/ttpos-takeout/utility/ttpos_client.go
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
```

### GetTtposClientWithAuth

```go
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

---

## 📚 实现清单

### Phase 1: 工具类开发

- [ ] 创建 `utility/ttpos_client.go`
- [ ] 实现 `GetTtposClient`
- [ ] 实现 `GetTtposClientWithAuth`

### Phase 2: 重构现有代码

- [ ] 重构 `grab_menu.go` 中的 `fetchMenuFromTTpos`

**详细任务**: 参见 `tasks.md`

---

**版本**: v1.0.0  
**创建日期**: 2025-12-18  
**作者**: rikugun  
**审核者**: -

