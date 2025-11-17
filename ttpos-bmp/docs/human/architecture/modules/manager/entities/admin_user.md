# AdminUser 实体模型说明

## 基本信息

- **实体名称**: AdminUser
- **表名**: admin_user
- **所属模块**: ttpos-manager
- **描述**: 管理员用户实体，用于管理系统管理员账户

## 字段说明

| 字段名 | 类型 | 说明 | 备注 |
|--------|------|------|------|
| AdminUserId | uint | 自增ID | 主键 |
| Username | string | 用户名 | |
| Phone | string | 手机号 | |
| Password | string | 登录密码 | 加密存储 |
| RealName | string | 姓名 | |
| IsSuper | int | 是否超级管理员 | 0-否 1-是 |
| Status | int | 状态 | 0未启用 1已启用 |
| CreateTime | int | 创建时间 | 时间戳 |
| UpdateTime | int | 更新时间 | 时间戳 |
| DeleteTime | int | 删除时间 | 时间戳，软删除 |

## 关联关系

### 关联实体
- **AdminUserId** → AdminUserRole.AdminUserId（用户角色关联）
- **AdminUserId** → AdminUserLoginLog.AdminUserId（登录日志）
- **AdminUserId** → AdminUserOptLog.AdminUserId（操作日志）

## 数据流分析

### 数据来源
- 管理员账户创建
- 通过管理后台创建和配置

### 数据流向
1. **用户创建流程**:
   - 创建管理员账户
   - 设置用户名、密码、手机号等信息
   - 分配角色（通过 AdminUserRole）

2. **用户登录流程**:
   - 用户登录时验证用户名和密码
   - 记录登录日志（AdminUserLoginLog）
   - 检查用户状态（Status）和权限

3. **用户操作流程**:
   - 用户执行操作时记录操作日志（AdminUserOptLog）
   - 根据角色权限（AdminUserRole → AdminRole → AdminRoleAccess）控制访问

### 业务场景
- 管理员账户管理
- 用户权限控制
- 登录审计
- 操作审计

## 索引建议

- 主键索引: AdminUserId
- 唯一索引: Username（用户名唯一）
- 唯一索引: Phone（手机号唯一）
- 普通索引: Status（状态查询）
- 普通索引: IsSuper（超级管理员查询）

## 注意事项

1. Password 字段需要加密存储
2. IsSuper 字段标识超级管理员，拥有所有权限
3. Status 字段控制账户启用状态
4. 使用软删除机制（DeleteTime）

