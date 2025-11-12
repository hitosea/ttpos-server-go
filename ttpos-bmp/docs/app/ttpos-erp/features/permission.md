# Permission 服务功能说明

## 概述
Permission 服务提供权限管理功能，主要用于 POS 系统的权限规则管理和权限校验。

## 服务接口

### IPermission - 权限管理

#### 权限规则管理
- **GetPosPermissionRuleList**: 获取 POS 权限规则列表
  - 根据查询条件过滤并返回权限规则信息列表

- **GetPosPermissionRule**: 获取 POS 权限规则详情
  - 根据规则代码获取权限规则的完整信息

- **CreatePosPermissionRule**: 创建 POS 权限规则
  - 创建新的权限规则记录

- **UpdatePosPermissionRule**: 更新 POS 权限规则
  - 更新现有的权限规则记录

- **DeletePosPermissionRule**: 删除 POS 权限规则
  - 删除指定的权限规则记录（根据规则代码）

#### 权限校验
- **CheckPermission**: 检查权限
  - 根据权限规则列表和公司名称检查是否有权限
  - 返回布尔值表示是否有权限

## 业务场景

### 权限规则配置
- 为不同角色配置 POS 操作权限
- 按公司维度控制权限
- 细粒度的功能权限控制

### 权限校验
- POS 操作前的权限验证
- 基于公司的权限隔离
- 多维度权限规则组合判断

## 使用说明

### 服务注册
```go
service.RegisterPermission(permissionImpl)
```

### 服务调用
```go
// 获取权限服务实例
permission := service.Permission()

// 创建权限规则
rule, err := permission.CreatePosPermissionRule(ctx, &erp.PosPermissionRule{
    RuleCode: "RULE-001",
    // 其他权限规则配置
})

// 检查权限
hasPermission, err := permission.CheckPermission(ctx, permissionList, "公司名称")
if !hasPermission {
    // 无权限处理
}
```

## 数据结构

### PosPermissionRule - 权限规则
- **RuleCode**: 规则代码（唯一标识）
- **RuleName**: 规则名称
- **Company**: 关联公司
- **Permissions**: 权限配置详情

### PermissionRule - 权限规则项
用于权限校验的规则项列表

## 权限校验逻辑

### 校验流程
1. 获取用户的权限规则列表
2. 传入需要校验的公司名称
3. 系统根据规则判断是否有权限
4. 返回校验结果

### 权限维度
- 公司维度：不同公司的权限隔离
- 功能维度：具体操作的权限控制
- 角色维度：基于角色的权限分配

## 注意事项
1. 权限规则代码（RuleCode）需要保证唯一性
2. 删除权限规则前需要确认没有被引用
3. 权限校验失败时应该有明确的提示信息
4. 权限规则变更会立即生效，需要谨慎操作
5. 建议使用缓存优化权限校验性能
6. 权限规则支持多条件组合判断
