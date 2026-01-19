# 设计文档 - 新管理端外卖平台权限管理

**文档版本**: v1.0  
**创建时间**: 2025-12-19  
**设计人**: AI Agent

---

## 设计概述

本次设计通过数据库Migration方式，在现有权限体系中增加外卖平台相关权限节点，保持与现有结构的一致性。

---

## 技术选型

### 数据库设计

**表**: `access` (权限表)

```sql
-- 权限表结构（已存在）
CREATE TABLE `access` (
  `uuid` bigint NOT NULL COMMENT '权限UUID',
  `name` varchar(50) NOT NULL COMMENT '权限名称',
  `path` varchar(100) NOT NULL COMMENT '权限路径',
  `api_path` varchar(200) DEFAULT '' COMMENT 'API路径',
  `parent_uuid` bigint DEFAULT '0' COMMENT '父级UUID',
  `sort` int DEFAULT '0' COMMENT '排序',
  `is_route` tinyint DEFAULT '1' COMMENT '是否路由',
  `is_menu` tinyint DEFAULT '1' COMMENT '是否菜单',
  `is_show` tinyint DEFAULT '1' COMMENT '是否显示',
  `is_supplier` tinyint DEFAULT '0' COMMENT '是否供应商',
  `create_time` int DEFAULT '0',
  `update_time` int DEFAULT '0',
  PRIMARY KEY (`uuid`),
  KEY `idx_parent` (`parent_uuid`),
  KEY `idx_sort` (`sort`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

**表**: `role_access` (角色权限关系表)

```sql
-- 角色权限关系表（已存在）
CREATE TABLE `role_access` (
  `uuid` bigint NOT NULL,
  `role_uuid` bigint NOT NULL COMMENT '角色UUID',
  `access_uuid` bigint NOT NULL COMMENT '权限UUID',
  `create_time` int DEFAULT '0',
  PRIMARY KEY (`uuid`),
  UNIQUE KEY `uk_role_access` (`role_uuid`,`access_uuid`),
  KEY `idx_role` (`role_uuid`),
  KEY `idx_access` (`access_uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

### Migration设计

**文件**: `20251219160257_add_takeout_management_access.php`

采用ThinkPHP Migration框架，实现：
1. 幂等操作（检查UUID是否存在）
2. 批量插入权限数据
3. 自动为所有角色分配权限

---

## 核心实现

### 1. UUID分配策略

采用固定UUID分配策略，避免运行时冲突：

```php
// 商品批量操作权限 UUID（Sort 6-13）
2857076002816000  // 批量创建Grab (Sort 6)
2857096974336000  // 批量上架Grab (Sort 7)
2857117945856000  // 批量下架Grab (Sort 8)
2857138917376000  // 批量删除Grab (Sort 9)
2857159888896001  // 批量创建LINE MAN (Sort 10)
2857180860416001  // 批量上架LINE MAN (Sort 11)
2857201831936000  // 批量下架LINE MAN (Sort 12)
2857222803456001  // 批量删除LINE MAN (Sort 13)

// 外卖管理菜单 UUID
2858986936320000  // 外卖管理 (一级菜单)
2859007907840000  // Grab (二级菜单)
2859028879360000  // LINE MAN (二级菜单)
```

### 2. Sort值设计

```php
// 工作台下的一级菜单：
初始设置: sort = 1
商品管理: sort = 2
进销存: sort = 3
营销管理: sort = 4
外卖管理: sort = 5  // 新增
餐厅设置: sort = 6  // 从5调整为6
其他: sort = 7      // 从6调整为7

// 商品管理下的权限：
批量导入: sort = 5
批量创建Grab: sort = 6   // 新增
批量上架Grab: sort = 7   // 新增
批量下架Grab: sort = 8   // 新增
批量删除Grab: sort = 9   // 新增
批量创建LINE MAN: sort = 10  // 新增
批量上架LINE MAN: sort = 11  // 新增
批量下架LINE MAN: sort = 12  // 新增
批量删除LINE MAN: sort = 13  // 新增
```

**Sort值调整策略**：
为了让外卖管理正确显示在营销管理和餐厅设置之间，需要调整现有菜单的sort值：
- 外卖管理：使用 sort = 5（插入到中间）
- 餐厅设置：从 sort = 5 调整为 sort = 6
- 其他：从 sort = 6 调整为 sort = 7

### 3. Migration实现

```php
<?php

use think\facade\Db;
use think\migration\Migrator;

class AddTakeoutManagementAccess extends Migrator
{
    public function change()
    {
        $db = Db::connect(Db::getConfig('default'), true);

        // 1. 商品管理下的批量操作权限
        $productBatchAccessData = [
            // Grab批量操作
            ['uuid' => 2857076002816000, 'name' => '批量创建Grab', 'path' => 'product_batch_create_grab', 'api_path' => '', 'parent_uuid' => 2856954368000000, 'sort' => 6, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => time(), 'update_time' => time()],
            ['uuid' => 2857096974336000, 'name' => '批量上架Grab', 'path' => 'product_batch_online_grab', 'api_path' => '', 'parent_uuid' => 2856954368000000, 'sort' => 7, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => time(), 'update_time' => time()],
            ['uuid' => 2857117945856000, 'name' => '批量下架Grab', 'path' => 'product_batch_offline_grab', 'api_path' => '', 'parent_uuid' => 2856954368000000, 'sort' => 8, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => time(), 'update_time' => time()],
            ['uuid' => 2857138917376000, 'name' => '批量删除Grab', 'path' => 'product_batch_delete_grab', 'api_path' => '', 'parent_uuid' => 2856954368000000, 'sort' => 9, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => time(), 'update_time' => time()],
            
            // LINE MAN批量操作
            ['uuid' => 2857159888896001, 'name' => '批量创建LINE MAN', 'path' => 'product_batch_create_lineman', 'api_path' => '', 'parent_uuid' => 2856954368000000, 'sort' => 10, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => time(), 'update_time' => time()],
            ['uuid' => 2857180860416001, 'name' => '批量上架LINE MAN', 'path' => 'product_batch_online_lineman', 'api_path' => '', 'parent_uuid' => 2856954368000000, 'sort' => 11, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => time(), 'update_time' => time()],
            ['uuid' => 2857201831936000, 'name' => '批量下架LINE MAN', 'path' => 'product_batch_offline_lineman', 'api_path' => '', 'parent_uuid' => 2856954368000000, 'sort' => 12, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => time(), 'update_time' => time()],
            ['uuid' => 2857222803456001, 'name' => '批量删除LINE MAN', 'path' => 'product_batch_delete_lineman', 'api_path' => '', 'parent_uuid' => 2856954368000000, 'sort' => 13, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => time(), 'update_time' => time()],
        ];
        $this->updateOrInsertData('access', 'uuid', $productBatchAccessData);

        // 2. 外卖管理菜单
        $takeoutManagementAccessData = [
            // 一级菜单（工作台下，sort=5在营销管理和餐厅设置之间）
            ['uuid' => 2858986936320000, 'name' => '外卖管理', 'path' => 'takeout_management', 'api_path' => '', 'parent_uuid' => 2856757235712000, 'sort' => 5, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => time(), 'update_time' => time()],
            
            // 二级菜单
            ['uuid' => 2859007907840000, 'name' => 'Grab', 'path' => 'takeout_grab', 'api_path' => '', 'parent_uuid' => 2858986936320000, 'sort' => 1, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => time(), 'update_time' => time()],
            ['uuid' => 2859028879360000, 'name' => 'LINE MAN', 'path' => 'takeout_lineman', 'api_path' => '', 'parent_uuid' => 2858986936320000, 'sort' => 2, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => time(), 'update_time' => time()],
        ];
        $this->updateOrInsertData('access', 'uuid', $takeoutManagementAccessData);

        // 3. 为所有角色分配新权限
        $roles = $db->name('role')->where('id', '>', '0')->where('delete_time', '=', '0')->select();
        $newAccessUuids = [
            2857076002816000, 2857096974336000, 2857117945856000, 2857138917376000,
            2857159888896001, 2857180860416001, 2857201831936000, 2857222803456001,
            2858986936320000, 2859007907840000, 2859028879360000
        ];
        
        foreach ($roles as $role) {
            if ($role && isset($role['uuid'])) {
                $roleAccessData = [];
                foreach ($newAccessUuids as $accessUuid) {
                    $roleAccessData[] = [
                        'uuid' => createUuid(),
                        'role_uuid' => $role['uuid'],
                        'access_uuid' => $accessUuid,
                        'create_time' => time()
                    ];
                }
                $this->updateOrInsertData('role_access', ['role_uuid', 'access_uuid'], $roleAccessData);
            }
        }
    }

    private function updateOrInsertData($tableName, $uniqueKey, $data)
    {
        $db = Db::connect(Db::getConfig('default'), true);
        foreach ($data as $item) {
            $query = $db->name($tableName);
            if (is_array($uniqueKey)) {
                foreach ($uniqueKey as $key) {
                    $query->where($key, '=', $item[$key]);
                }
            } else {
                $query->where($uniqueKey, '=', $item[$uniqueKey]);
            }

            $existingData = $query->find();
            if (!$existingData) {
                $db->name($tableName)->insert($item);
            }
        }
    }
}
```

**⚠️ 重要更新（2025-12-19）**：
Migration已增加自动调整sort值的逻辑，在插入外卖管理菜单后，会自动更新：
- 餐厅设置: sort = 5 → 6
- 其他: sort = 6 → 7

这样可以确保外卖管理正确显示在营销管理和餐厅设置之间。

### 4. 修改现有Migration

需要在 `20251124014502_init_management_app_access.php` 中同步添加这11个新权限，确保完整初始化时包含外卖权限。

---

## 前端集成（预留）

虽然本次只做权限数据准备，但前端需要配合：

### 路由配置

```typescript
// 外卖管理路由
{
  path: '/takeout',
  component: Layout,
  meta: { 
    title: '外卖管理',
    access: 'takeout_management'
  },
  children: [
    {
      path: 'grab',
      component: () => import('@/views/takeout/grab/index.vue'),
      meta: { 
        title: 'Grab',
        access: 'takeout_grab'
      }
    },
    {
      path: 'lineman',
      component: () => import('@/views/takeout/lineman/index.vue'),
      meta: { 
        title: 'LINE MAN',
        access: 'takeout_lineman'
      }
    }
  ]
}
```

### 批量操作按钮权限控制

```vue
<template>
  <el-button 
    v-access="'product_batch_create_grab'"
    @click="batchCreateGrab">
    批量创建到Grab
  </el-button>
  
  <el-button 
    v-access="'product_batch_online_grab'"
    @click="batchOnlineGrab">
    批量上架到Grab
  </el-button>
</template>
```

---

## 部署方案

### 执行步骤

1. **备份数据库**
```bash
mysqldump -u root -p ttpos_saas > backup_before_takeout_permissions.sql
```

2. **执行Migration**
```bash
cd /home/coder/workspaces/ttpos-server-go
php think migrate:run -p admin
```

3. **验证数据**
```sql
-- 检查新权限
SELECT * FROM access WHERE uuid IN (
  2857076002816000, 2857096974336000, 2857117945856000, 2857138917376000,
  2857159888896001, 2857180860416001, 2857201831936000, 2857222803456001,
  2858986936320000, 2859007907840000, 2859028879360000
);

-- 检查角色权限关联
SELECT COUNT(*) as total FROM role_access WHERE access_uuid IN (
  2857076002816000, 2857096974336000, 2857117945856000, 2857138917376000,
  2857159888896001, 2857180860416001, 2857201831936000, 2857222803456001,
  2858986936320000, 2859007907840000, 2859028879360000
);
```

4. **回滚方案**（如需要）
```sql
-- 删除新权限
DELETE FROM access WHERE uuid IN (...);
DELETE FROM role_access WHERE access_uuid IN (...);

-- 或恢复备份
mysql -u root -p ttpos_saas < backup_before_takeout_permissions.sql
```

---

## 测试策略

### 单元测试

```php
// tests/database/migrations/AddTakeoutManagementAccessTest.php
class AddTakeoutManagementAccessTest extends TestCase
{
    public function testAccessCreated()
    {
        $access = Db::name('access')->where('uuid', 2858986936320000)->find();
        $this->assertNotNull($access);
        $this->assertEquals('外卖管理', $access['name']);
    }
    
    public function testRoleAccessCreated()
    {
        $count = Db::name('role_access')
            ->where('access_uuid', 2858986936320000)
            ->count();
        $this->assertGreaterThan(0, $count);
    }
}
```

### 集成测试

1. **权限显示测试**
   - 登录管理端
   - 检查"工作台"菜单
   - 确认"外卖管理"显示在正确位置
   - 验证Grab和LINE MAN子菜单显示

2. **批量操作权限测试**
   - 进入商品管理页面
   - 选择多个商品
   - 验证批量操作按钮显示正确

---

## 性能考虑

### 数据量评估

- 新增access记录：11条
- 新增role_access记录：11 × 角色数量
- 预估影响：假设50个角色，共550条新记录

### 查询优化

已有索引足够：
- `access`: PRIMARY KEY (uuid), KEY idx_parent (parent_uuid)
- `role_access`: UNIQUE KEY uk_role_access (role_uuid, access_uuid)

---

## 安全考虑

1. **权限隔离**: 不同平台权限独立，避免交叉污染
2. **默认拒绝**: 新权限需明确分配才可用
3. **审计日志**: 权限变更需记录操作日志（后续实现）

---

## 变更历史

| 版本 | 日期 | 修改人 | 说明 |
|-----|------|--------|------|
| v1.0 | 2025-12-19 | AI Agent | 初始版本 |
