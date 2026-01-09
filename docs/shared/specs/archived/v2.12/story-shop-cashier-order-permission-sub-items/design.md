# 收银机接单权限子项管理 设计文档

> 本文档定义收银机接单权限子项管理功能的技术设计和实现方案。

## 📋 概述

在收银机权限下新增"接单"权限（与沽清权限平级），并将原有的接单权限和外送权限调整为该新权限的子项，同时新增外卖接单权限作为第三项子项，实现细粒度的接单权限控制。根据云平台外卖开关状态动态显示/隐藏"外卖"权限子项，并在云平台开启任意一个外卖平台时，自动为默认角色（Cashier、Store Manager）勾选"外卖"权限。

**实现范围**：
1. 数据库迁移：调整权限数据结构，新增接单权限和外卖权限
2. 权限筛选：在 Go Main 和 PHP Admin 模块中添加外卖权限过滤逻辑
3. 默认角色权限分配：在迁移脚本中为默认角色分配新权限
4. API 确认：确认 BaseInfo API 已返回外卖开关状态

**参考实现**：`admin/database/migrations/20250717013925_add_cashier_permission_delivery_order_to_access.php`

---

## 🎯 规范对齐

### Go Main 规范 (go-main.mdc)

- Service 只依赖其他 Service 接口
- Repository 只持有 db 实例
- 不使用 panic，返回 error
- 所有代码使用中文注释

### PHP 规范 (php.mdc)

- 遵循 MVC 分层
- Controller 不写业务逻辑
- 使用验证器验证参数
- 使用软删除

### 数据库规范 (database.mdc)

- 必需字段完整（id, uuid, create_time, update_time, delete_time）
- 时间字段使用 int 类型
- UUID 字段使用 bigint unsigned
- 字段名使用 snake_case

---

## 🔄 代码复用分析

### 可复用的现有组件

- **权限迁移脚本**: `admin/database/migrations/20250717013925_add_cashier_permission_delivery_order_to_access.php` - 外送权限迁移实现
- **权限筛选逻辑**: `main/app/service/role_access.go` - Go Main 权限筛选实现
- **权限筛选逻辑**: `admin/app/common/model/shop/Access.php` - PHP Admin 权限筛选实现
- **角色权限分配**: `admin/database/migrations/20250717013925_add_cashier_permission_delivery_order_to_access.php` - 默认角色权限分配实现
- **BaseInfo API**: `main/app/service/auth.go` - 商家基础信息接口

### 集成点

- **权限表**: `access` 表 - 新增和更新权限记录
- **角色权限关联表**: `role_access` 表 - 分配默认角色权限
- **权限筛选**: `filterPermission` 函数 - 根据云平台外卖开关过滤权限
- **BaseInfo API**: `/shop/base` 接口 - 返回外卖开关状态

---

## 🏗️ 架构设计

### 分层设计原则

**PHP Admin 模块**:

```
Migration Script (数据库迁移)
  ↓
Access Model (权限模型)
  ↓
Role Access (角色权限关联)
```

**Go Main 模块**:

```
Service Layer (role_access.go)
  ↓ filterPermission
Permission List (权限列表)
```

### 权限数据结构

**调整前**:
```
收银机权限 (1704880670)
├── 沽清 (1724220513, sort=10)
├── 接单 (1724320522, sort=20, parent=1704880670)
└── 外送 (1752716650, sort=20, parent=1704880670)
```

**调整后**:
```
收银机权限 (1704880670)
├── 沽清 (1724220513, sort=10)
└── 接单 (新UUID, sort=11, parent=1704880670)
    ├── 扫码接单 (1724320522, sort=10, parent=新接单UUID)
    ├── 外送 (1752716650, sort=20, parent=新接单UUID)
    └── 外卖 (新UUID, sort=30, parent=新接单UUID)
```

### 权限筛选流程

```
获取权限列表
  ↓
检查云平台外卖开关状态
  ↓
如果未开启 → 过滤掉外卖权限
  ↓
返回筛选后的权限列表
```

---

## 🗄️ 数据库设计

### 数据表设计

#### 表 1: access（修改现有表）

**新增权限记录**:

```sql
-- 新增接单权限（父级）
INSERT INTO `ttpos_access` (
    `uuid`, `name`, `path`, `api_path`, `parent_uuid`, `sort`,
    `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`,
    `plus_category_uuid`, `remark`, `is_supplier`, `create_time`, `update_time`
) VALUES (
    1734000000, '接单', 'cashier_accept_order', '', 1704880670, 11,
    '', '', 0, 0, 1, 0, '', 0, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()
);

-- 新增外卖权限（子级）
INSERT INTO `ttpos_access` (
    `uuid`, `name`, `path`, `api_path`, `parent_uuid`, `sort`,
    `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`,
    `plus_category_uuid`, `remark`, `is_supplier`, `create_time`, `update_time`
) VALUES (
    1734000001, '外卖', 'cashier_accept_delivery_platform', '/cashier/grab_order/list', 1734000000, 30,
    '', '', 0, 0, 1, 0, '', 0, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()
);
```

**更新权限记录**:

```sql
-- 更新原有接单权限为扫码接单
UPDATE `ttpos_access` SET
    `name` = '扫码接单',
    `path` = 'cashier_accept_scan_order',
    `parent_uuid` = 1734000000,
    `sort` = 10,
    `update_time` = UNIX_TIMESTAMP()
WHERE `uuid` = 1724320522;

-- 更新外送权限的 parent_uuid
UPDATE `ttpos_access` SET
    `parent_uuid` = 1734000000,
    `sort` = 20,
    `update_time` = UNIX_TIMESTAMP()
WHERE `uuid` = 1752716650;
```

#### 表 2: role_access（新增记录）

**为默认角色分配权限**:

```sql
-- 为新接单权限分配默认角色
INSERT INTO `ttpos_role_access` (`uuid`, `role_uuid`, `access_uuid`, `create_time`, `update_time`, `delete_time`)
SELECT createUuid(), 1, 1734000000, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0
WHERE NOT EXISTS (
    SELECT 1 FROM `ttpos_role_access` WHERE `role_uuid` = 1 AND `access_uuid` = 1734000000
);

INSERT INTO `ttpos_role_access` (`uuid`, `role_uuid`, `access_uuid`, `create_time`, `update_time`, `delete_time`)
SELECT createUuid(), 2, 1734000000, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0
WHERE NOT EXISTS (
    SELECT 1 FROM `ttpos_role_access` WHERE `role_uuid` = 2 AND `access_uuid` = 1734000000
);

-- 为外卖权限分配默认角色
INSERT INTO `ttpos_role_access` (`uuid`, `role_uuid`, `access_uuid`, `create_time`, `update_time`, `delete_time`)
SELECT createUuid(), 1, 1734000001, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0
WHERE NOT EXISTS (
    SELECT 1 FROM `ttpos_role_access` WHERE `role_uuid` = 1 AND `access_uuid` = 1734000001
);

INSERT INTO `ttpos_role_access` (`uuid`, `role_uuid`, `access_uuid`, `create_time`, `update_time`, `delete_time`)
SELECT createUuid(), 2, 1734000001, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0
WHERE NOT EXISTS (
    SELECT 1 FROM `ttpos_role_access` WHERE `role_uuid` = 2 AND `access_uuid` = 1734000001
);
```

**迁移文件**: `admin/database/migrations/{YYYYMMDDHHMMSS}_add_cashier_order_permission_sub_items.php`

**参考实现**: `admin/database/migrations/20250717013925_add_cashier_permission_delivery_order_to_access.php`

---

## 📊 数据模型

### 权限 UUID 设计

| 权限名称 | UUID | parent_uuid | sort | 说明 |
|---------|------|-------------|------|------|
| 接单（新） | 1734000000 | 1704880670 | 11 | 新增父级权限 |
| 扫码接单 | 1724320522 | 1734000000 | 10 | 原有接单权限，调整 parent_uuid |
| 外送 | 1752716650 | 1734000000 | 20 | 原有外送权限，调整 parent_uuid |
| 外卖（新） | 1734000001 | 1734000000 | 30 | 新增子级权限 |

**注意**: UUID 使用时间戳生成，确保唯一性。实际迁移脚本中应使用 `time()` 或 `createUuid()` 函数生成。

---

## 🔌 API 设计

### BaseInfo API（已存在，无需修改）

**接口**: `/shop/base` (GET)

**响应结构**:

```json
{
  "code": 1,
  "message": "success",
  "data": {
    "company": {
      "uuid": 123456,
      "name": "商家名称",
      "is_open_grab_delivery": true
    }
  }
}
```

**字段说明**:
- `is_open_grab_delivery`: 是否开启Grab外卖功能（bool类型）
- 该字段已存在于 `main/app/dto/resp/base.go` 的 `Company` 结构体中
- 该字段已正确映射到 `company_setting.enable_grab_delivery`

---

## 🧩 组件和接口

### 权限筛选逻辑

#### Go Main 模块

**文件**: `main/app/service/role_access.go`

**函数**: `filterPermission`

**实现逻辑**:

```go
// 授权无Grab外卖权限（未开启Grab外卖时，隐藏外卖接单权限）
if !companySetting.IsOpenGrabDelivery() {
    if permission.Uuid == 1734000001 { // 外卖权限UUID
        continue
    }
}
```

**参考**: 外送权限过滤逻辑（第189-192行）

#### PHP Admin 模块

**文件**: `admin/app/common/model/shop/Access.php`

**函数**: `recursiveMenuArray`

**实现逻辑**:

```php
// 授权无Grab外卖权限（未开启Grab外卖时，隐藏外卖接单权限）
if (($licenses['enable_grab_delivery'] ?? 0) == 0) {
    if ($value['uuid'] == 1734000001) { // 外卖权限UUID
        continue;
    }
}
```

**参考**: 外送权限过滤逻辑（第379-384行）

### 默认角色权限分配

**文件**: `admin/database/migrations/{timestamp}_add_cashier_order_permission_sub_items.php`

**实现逻辑**:

```php
// 获取默认角色
$storeManager = $db->name('role')->where('name', '=', 'Store Manager')->find();
$cashier = $db->name('role')->where('name', '=', 'Cashier')->find();

// 为新接单权限分配默认角色
if ($storeManager && isset($storeManager['uuid'])) {
    $this->updateOrInsertData('role_access', ['role_uuid', 'access_uuid'], [
        ['uuid' => createUuid(), 'role_uuid' => $storeManager['uuid'], 'access_uuid' => '1734000000', 'create_time' => time()],
    ]);
}

if ($cashier && isset($cashier['uuid'])) {
    $this->updateOrInsertData('role_access', ['role_uuid', 'access_uuid'], [
        ['uuid' => createUuid(), 'role_uuid' => $cashier['uuid'], 'access_uuid' => '1734000000', 'create_time' => time()],
    ]);
}

// 为外卖权限分配默认角色（同上，access_uuid 改为 1734000001）
```

**参考**: `admin/database/migrations/20250717013925_add_cashier_permission_delivery_order_to_access.php` 第38-45行

---

## ⚡ 缓存设计

### Redis 缓存（如适用）

**缓存策略**:
- 权限列表缓存：`ttpos:permission:{company_uuid}:{route_name}`
- 过期时间：5分钟
- 更新策略：Cache-Aside Pattern

**缓存失效**:
- 权限配置变更时清除相关缓存
- 云平台外卖开关状态变更时清除权限缓存

---

## 🚨 错误处理

### 错误场景

#### 场景 1: 权限迁移失败

- **处理方式**: 使用事务确保数据一致性，迁移失败时回滚
- **用户影响**: 迁移脚本执行失败，不影响现有权限配置
- **代码示例**:
  ```php
  $db->startTrans();
  try {
      // 执行迁移操作
      $db->commit();
  } catch (\Exception $e) {
      $db->rollback();
      throw $e;
  }
  ```

#### 场景 2: 权限筛选逻辑错误

- **处理方式**: 添加日志记录，确保筛选逻辑正确
- **用户影响**: 权限显示不正确，需要检查日志排查
- **缓解措施**: 充分测试权限筛选逻辑，覆盖所有场景

---

## 🔒 安全设计

### 权限验证

- **RBAC**: 基于角色的访问控制
- **API 权限**: 每个 API 检查用户权限
- **权限筛选**: 根据云平台开关状态动态过滤权限

### 数据安全

- **SQL 注入防护**: 使用参数化查询
- **权限数据完整性**: 使用事务确保数据一致性

---

## 🧪 测试策略

### 单元测试

**目标覆盖率**:
- 权限筛选逻辑: 100%（关键业务逻辑）

**测试内容**:
- 权限筛选逻辑（Go Main 和 PHP Admin）
- 默认角色权限分配
- 权限数据结构调整

### 集成测试

**测试流程**:
- 权限迁移脚本执行
- 权限列表查询
- 权限筛选功能
- 默认角色权限分配

---

## 📈 性能优化

### 优化策略

1. **数据库优化**:
   - 权限查询使用索引（uuid, parent_uuid）
   - 避免全表扫描

2. **缓存优化**:
   - 权限列表缓存（如适用）
   - 减少数据库查询

### 性能指标

- 权限查询响应时间: < 200ms
- 权限筛选逻辑执行时间: < 10ms

---

## 📚 实现清单

### Phase 1: 数据库迁移和权限数据

- [ ] 创建数据库迁移脚本
- [ ] 新增接单权限
- [ ] 更新原有接单权限和外送权限
- [ ] 新增外卖权限
- [ ] 为默认角色分配权限

### Phase 2: 权限筛选逻辑

- [ ] Go Main 模块权限筛选
- [ ] PHP Admin 模块权限筛选
- [ ] 在 getLicense() 中添加 enable_grab_delivery 字段

### Phase 3: API 确认

- [ ] 确认 BaseInfo API 返回外卖开关状态
- [ ] 确认相关接口字段映射正确

### Phase 4: 测试

- [ ] 权限迁移脚本测试
- [ ] 权限筛选逻辑测试
- [ ] 默认角色权限分配测试
- [ ] 集成测试

**详细任务**: 参见 `tasks.md`

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{user}/{YYYY-MM}/{YYYY-MM-DD}.md`
- 在技术方案评审、关键架构决策或踩坑总结后，立即补充 Episode 并在设计文档尾部互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-12-09  
**作者**: 开发团队  
**审核者**: 待审核
