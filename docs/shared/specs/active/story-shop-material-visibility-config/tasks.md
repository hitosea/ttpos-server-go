# story-shop-material-visibility-config 任务清单

## 📊 进度总览

| 项目 | 数值 |
|------|------|
| 总 SP | 8 |
| 总任务数 | 12 |
| 已完成 | 0 |
| 完成率 | 0% |

---

## Phase 1: 数据层 (SP 2)

### 1.1 创建数据模型

| 项目 | 内容 |
|------|------|
| File | `main/app/model/material_category_visibility.go` |
| Purpose | 定义 MaterialCategoryVisibility 模型及 JSON 辅助方法 |
| Requirements | Req 2.1-2.3 配置数据存储 |
| Leverage | 参考 `main/app/model/material.go` 结构 |

**实现要点**:
- 定义 `MaterialCategoryVisibility` 结构体
- 实现 `GetCategoryUuids()` / `SetCategoryUuids()` JSON 解析方法
- 实现 `GetRoleConfigs()` / `SetRoleConfigs()` JSON 解析方法
- 定义 `RoleConfig` 结构体

- [ ] 完成

---

### 1.2 创建 Repository

| 项目 | 内容 |
|------|------|
| File | `main/app/repository/material_category_visibility.go` |
| Purpose | 可见性配置数据访问层 |
| Requirements | Req 1-4 所有数据操作 |
| Leverage | 参考 `main/app/repository/material.go` |

**实现要点**:
- 实现 `IMaterialCategoryVisibilityRepo` 接口
- GetList / GetByUuid / Create / Update / Delete
- GetByRoleAndCompany（按角色查询配置）
- GetAllConfiguredRoles（获取所有已配置的角色）

- [ ] 完成

---

### 1.3 创建数据库迁移

| 项目 | 内容 |
|------|------|
| File | `admin/database/migrations/{timestamp}_create_material_category_visibility.php` |
| Purpose | 创建 ttpos_material_category_visibility 表 |
| Requirements | 数据模型约束 |
| Leverage | 参考现有迁移文件格式 |

**实现要点**:
- 设置 `const TARGET = 'all';`
- 创建表结构（参考 design.md）
- 添加索引：uuid、headquarter_uuid

- [ ] 完成

---

### 1.4 更新种子文件

| 项目 | 内容 |
|------|------|
| File | `admin/database/seeds/shop_01.sql` |
| Purpose | 同步表结构到种子文件 |
| Requirements | CLAUDE.md 规范 |

- [ ] 完成

---

## Phase 2: 配置管理服务 (SP 3)

### 2.1 创建 Service - CRUD

| 项目 | 内容 |
|------|------|
| File | `main/app/service/material_category_visibility.go` |
| Purpose | 可见性配置管理业务逻辑 |
| Requirements | Req 1, Req 2 |
| Leverage | 参考 `main/app/service/material.go` |

**实现要点**:
- 定义 `IMaterialCategoryVisibilitySrv` 接口
- 实现 GetList / GetDetail / Create / Update / Delete
- 总店权限校验
- 参数校验（名称必填、角色至少选一个）

- [ ] 完成

---

### 2.2 实现子店角色查询

| 项目 | 内容 |
|------|------|
| File | `main/app/service/material_category_visibility.go` |
| Purpose | 跨库查询所有子店角色 |
| Requirements | Req 2.3 门店角色选择 |
| Leverage | `company.go` GetSubShopList + `role.go` GetRoleList |

**实现要点**:
- 实现 `GetSubShopRoles()` 方法
- 使用 `utils.Go` 并发查询各子店数据库
- 超时控制（3s）
- 排除总店角色

- [ ] 完成

---

### 2.3 实现同步机制

| 项目 | 内容 |
|------|------|
| File | `main/app/service/material_category_visibility.go` |
| Purpose | 配置同步到所有子店 |
| Requirements | Req 4 配置同步机制 |
| Leverage | 参考 `SyncMaterialCategory` 模式 |

**实现要点**:
- 实现 `SyncToSubShops()` 方法
- 实现 `SyncDeleteToSubShops()` 方法
- 使用 `utils.Go` 异步执行
- 失败日志记录（包含 company_uuid）
- 重试机制（可选）

- [ ] 完成

---

### 2.4 创建 DTO

| 项目 | 内容 |
|------|------|
| File | `main/app/dto/req/material_category_visibility.go` |
| File | `main/app/dto/resp/material_category_visibility.go` |
| Purpose | 请求和响应数据对象 |
| Requirements | Req 1, Req 2 |

**实现要点**:
- VisibilityListReq / VisibilityCreateReq / VisibilityUpdateReq
- VisibilityListResp / VisibilityDetailResp / SubShopRolesResp
- 响应切片使用 `make` 初始化

- [ ] 完成

---

## Phase 3: 过滤逻辑 (SP 3)

### 3.1 实现可见性过滤方法

| 项目 | 内容 |
|------|------|
| File | `main/app/service/material_category_visibility.go` |
| Purpose | 核心可见性判断逻辑 |
| Requirements | Req 3 可见性判断逻辑 |

**实现要点**:
- 实现 `GetVisibleCategoryUuids(staffRoleUuids, companyUuid) ([]uint64, bool)`
- 场景1：所有角色未配置 → 返回 nil, false（全部可见）
- 场景2：部分角色已配置 → 返回类别并集, true（白名单模式）
- 场景3：IsAllRoles=1 的配置对所有角色生效

- [ ] 完成

---

### 3.2 集成到 MaterialService

| 项目 | 内容 |
|------|------|
| File | `main/app/service/material.go` |
| Purpose | 在 GetMaterialList 中集成过滤 |
| Requirements | Req 3.3 常规业务模块过滤 |

**实现要点**:
- 修改 `GetMaterialList` 方法
- 获取当前用户角色 UUID 列表
- 调用 `GetVisibleCategoryUuids` 获取可见类别
- 如果 needFilter=true，添加 `WhereCategoryIn` 过滤
- 移除/替换现有的 `WhereAllowSubstoreVisible` 逻辑

- [ ] 完成

---

### 3.3 集成盘点特殊处理

| 项目 | 内容 |
|------|------|
| File | `main/app/service/inventory_check.go` (或相关盘点服务) |
| Purpose | 盘点场景可见性处理 |
| Requirements | Req 3.4 盘点场景特殊处理 |

**实现要点**:
- 全盘/日盘/周盘/月盘：跳过可见性过滤
- 指定物品盘点：应用可见性过滤
- 根据盘点类型参数判断

- [ ] 完成

---

### 3.4 创建 API 控制器

| 项目 | 内容 |
|------|------|
| File | `main/app/api/v1/shop/shop_material_category_visibility.go` |
| Purpose | HTTP API 端点 |
| Requirements | Req 1, Req 2 |

**实现要点**:
- 实现 6 个 API 端点（参考 design.md API 设计）
- 总店权限校验
- 参数校验（使用 binding tag）
- 标准响应格式

- [ ] 完成

---

### 3.5 注册路由

| 项目 | 内容 |
|------|------|
| File | `main/app/api/v1/shop/router.go` |
| Purpose | 添加新路由 |
| Requirements | API 层集成 |

**实现要点**:
- 添加 `/material_category_visibility` 路由组
- 注册 list / detail / add / edit / delete / sub_shop_roles 路由

- [ ] 完成

---

## 提交清单

### 代码质量

- [ ] `go mod tidy` 执行
- [ ] `go fmt ./...` 执行
- [ ] `go vet ./...` 通过
- [ ] 测试通过: `go test ./...`

### 功能完整性

- [ ] 所有验收标准通过（Req 1-5）
- [ ] API 响应格式正确（data 为对象）
- [ ] 响应切片使用 `make` 初始化
- [ ] 日志包含 `company_uuid` 字段

### 迁移同步

- [ ] 迁移文件已创建
- [ ] shop_01.sql 已更新

### 测试覆盖

- [ ] Service 层测试覆盖率 ≥ 80%
- [ ] 可见性判断逻辑全场景覆盖
- [ ] 同步机制测试

---

## 开发顺序建议

```
Phase 1 (数据层)
    │
    ├── 1.1 Model ─────────────┐
    │                          │
    ├── 1.2 Repository ────────┤
    │                          │
    ├── 1.3 Migration ─────────┤
    │                          │
    └── 1.4 Seeds ─────────────┘
                               │
Phase 2 (配置管理)             ▼
    │
    ├── 2.1 Service CRUD ──────┐
    │                          │
    ├── 2.2 子店角色查询 ───────┤
    │                          │
    ├── 2.3 同步机制 ───────────┤
    │                          │
    └── 2.4 DTO ───────────────┘
                               │
Phase 3 (过滤逻辑)             ▼
    │
    ├── 3.1 过滤方法 ───────────┐
    │                          │
    ├── 3.2 集成 Material ──────┤
    │                          │
    ├── 3.3 盘点特殊处理 ────────┤
    │                          │
    ├── 3.4 API 控制器 ─────────┤
    │                          │
    └── 3.5 路由注册 ───────────┘
                               │
                               ▼
                         提交 & 测试
```

---

**版本**: v1.0.0
**创建日期**: 2026-02-03
