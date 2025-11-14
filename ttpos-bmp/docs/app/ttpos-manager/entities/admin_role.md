# AdminRole 实体模型说明

## 基本信息

- **实体名称**: AdminRole
- **表名**: admin_role
- **所属模块**: ttpos-manager
- **描述**: 管理员角色实体，用于定义系统角色

## 字段说明

| 字段名 | 类型 | 说明 | 备注 |
|--------|------|------|------|
| Id | uint | 自增ID | 主键 |
| RoleName | string | 角色名称 | |
| Sort | int | 排序 | 数字越小越靠前 |
| CreateTime | int | 创建时间 | 时间戳 |
| UpdateTime | int | 更新时间 | 时间戳 |
| DeleteTime | int | 删除时间 | 时间戳，软删除 |

## 关联关系

### 关联实体
- **Id** → AdminRoleAccess.RoleId（角色权限关联）
- **Id** → AdminUserRole.RoleId（用户角色关联）

## 数据流分析

### 数据来源
- 角色配置信息
- 通过管理后台创建和配置

### 数据流向
1. **角色创建流程**:
   - 创建角色并设置角色名称
   - 分配权限（通过 AdminRoleAccess）
   - 设置排序（Sort）

2. **角色使用流程**:
   - 将角色分配给用户（通过 AdminUserRole）
   - 用户获得角色对应的权限（通过 AdminRoleAccess → AdminAccess）

### 业务场景
- 角色管理
- 权限分配
- 角色排序

## 索引建议

- 主键索引: Id
- 唯一索引: RoleName（角色名称唯一）
- 普通索引: Sort（排序查询）

## 注意事项

1. Sort 字段用于角色列表排序
2. 使用软删除机制（DeleteTime）
3. 角色通过权限关联表（AdminRoleAccess）分配权限

