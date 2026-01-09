# Takeout 模块服务层重构完成

**重构时间**: 2025-12-22  
**版本**: v1.0.0

---

## 📋 重构目标

将订单和配置服务从应用层移动到领域层：
- **问题**：Application 层直接调用 Repository，跳过了 Domain Service
- **方案**：将业务逻辑移到 Domain Service，符合 DDD 分层规范

---

## ✅ 已完成的工作

### 1️⃣ 移动服务文件

**从 `application/` 移动到 `domain/service/`**：

```bash
✅ application/takeout_order_srv.go
   → domain/service/takeout_order_service.go

✅ application/takeout_settings_srv.go
   → domain/service/takeout_settings_service.go
```

### 2️⃣ 更新 Package 声明

```go
// 旧 package
package application

// 新 package
package service
```

---

## 📂 最终目录结构

```
main/app/modules/takeout/
├── interfaces/                         # 接口层（HTTP API）
│   ├── request/
│   └── response/
├── application/                        # 应用层（用例编排）
│   └── takeout_app_service.go         # 菜单/绑定应用服务
├── domain/                             # 领域层（核心业务）
│   ├── model/                          # 领域模型
│   ├── value_object/                   # 值对象
│   └── service/                        # 领域服务 ✅
│       ├── takeout_order_service.go    # 订单领域服务（新位置）
│       ├── takeout_settings_service.go # 配置领域服务（新位置）
│       ├── takeout_domain_service.go   # 外卖状态领域服务
│       ├── import_progress_service.go  # 导入进度领域服务
│       └── platform_converter.go       # 平台转换器接口
└── infrastructure/                     # 基础设施层
    ├── persistence/                    # 数据持久化
    └── adapter/                        # 外部适配器
```

---

## 🎯 DDD 分层说明

### 正确的分层调用链

```
┌─────────────────────────────────────┐
│    interfaces/                      │  ← HTTP Handler / Controller
│    • 接收请求                        │
│    • 参数验证                        │
│    • 返回响应                        │
└─────────────────────────────────────┘
              ↓
┌─────────────────────────────────────┐
│    application/                     │  ← Application Service
│    • 用例编排                        │
│    • 事务管理                        │
│    • 调用多个 Domain Service          │
└─────────────────────────────────────┘
              ↓
┌─────────────────────────────────────┐
│    domain/service/                  │  ← Domain Service ✅
│    • 核心业务逻辑                     │
│    • 领域规则验证                     │
│    • 调用 Repository                  │
└─────────────────────────────────────┘
              ↓
┌─────────────────────────────────────┐
│    infrastructure/persistence/      │  ← Repository
│    • 数据库操作                       │
│    • 数据持久化                       │
└─────────────────────────────────────┘
```

### 本次重构前后对比

#### ❌ 重构前（跳过 Domain Service）

```
Controller → Application Service → Repository
                                      ↑
                              直接调用（不符合 DDD）
```

#### ✅ 重构后（符合 DDD 标准）

```
Controller → Application Service → Domain Service → Repository
                                          ↑
                                   正确的调用链
```

---

## 📝 服务职责划分

### Domain Service（领域服务）

**位置**：`domain/service/`

**职责**：
- ✅ 核心业务逻辑
- ✅ 领域规则验证
- ✅ 单一领域的完整操作
- ✅ 调用 Repository 进行数据持久化

**示例**：
```go
// takeout_order_service.go
type ITakeoutOrderSrv interface {
    GetList()        // 获取订单列表
    GetByUuid()      // 获取订单详情
    AcceptOrder()    // 接单（业务逻辑）
    RejectOrder()    // 拒单（业务逻辑）
    SyncNewOrder()   // 同步新订单
}
```

### Application Service（应用服务）

**位置**：`application/`

**职责**：
- ✅ 用例编排（协调多个 Domain Service）
- ✅ 事务管理
- ✅ 跨领域操作
- ✅ 第三方集成调用

**示例**：
```go
// takeout_app_service.go
type ITakeoutAppService interface {
    // 菜单相关（跨领域：菜单 + 商品 + 分类）
    ExportMenu()         // 导出菜单（协调多个服务）
    ImportMenu()         // 导入菜单（复杂流程编排）
    SyncMenuChanges()    // 同步菜单变更
    
    // 绑定相关（跨系统：Main + BMP）
    GetBindingLink()     // 获取绑定链接（RPC调用）
    CheckBindingStatus() // 检查绑定状态（RPC调用）
}
```

---

## 🔍 服务命名规范

### Domain Service（领域服务）

```
✅ 命名格式：{Entity}Service
✅ 文件名：{entity}_service.go

示例：
- TakeoutOrderService      → takeout_order_service.go
- TakeoutSettingsService   → takeout_settings_service.go
- ImportProgressService    → import_progress_service.go
```

### Application Service（应用服务）

```
✅ 命名格式：{Module}AppService
✅ 文件名：{module}_app_service.go

示例：
- TakeoutAppService        → takeout_app_service.go
```

---

## 📚 相关文档

- 领域服务: `main/app/modules/takeout/domain/service/`
- 应用服务: `main/app/modules/takeout/application/`
- Repository: `main/app/modules/takeout/infrastructure/persistence/`

---

## 🔄 后续工作

如果有 Controller 层，需要更新 import 路径：

```go
// 旧 import
import "ttpos-server-go/app/modules/takeout/application"

// 新 import
import "ttpos-server-go/app/modules/takeout/domain/service"

// 使用
orderSrv := service.NewTakeoutOrderSrv(dbm)
```

---

**维护者**: TTPOS Team  
**最后更新**: 2025-12-22

