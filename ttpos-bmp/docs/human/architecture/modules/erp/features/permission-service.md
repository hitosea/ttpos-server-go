# Permission 权限服务说明文档

## 📋 服务概览

Permission 权限服务是 ttpos-erp 模块的权限管理服务，负责 POS 权限规则的管理和权限检查。该服务提供权限规则的 CRUD 操作和权限检查功能，支持白名单和黑名单机制。

## 🎯 主要功能

### 权限规则管理
- **权限规则列表**: 查询权限规则列表
- **权限规则详情**: 获取权限规则完整信息
- **创建权限规则**: 创建新的权限规则
- **更新权限规则**: 更新现有权限规则
- **删除权限规则**: 删除权限规则

### 权限检查
- **权限检查**: 根据权限规则列表和公司名称检查是否有权限
- **白名单机制**: 支持白名单规则（只允许列表中的公司）
- **黑名单机制**: 支持黑名单规则（禁止列表中的公司）

## 📁 文件结构

```
internal/logic/permission/
└── permission.go             # 权限服务主逻辑
```

## 🔧 接口定义

### IPermission 接口

```go
type IPermission interface {
    // GetPosPermissionRuleList 获取POS权限规则列表
    GetPosPermissionRuleList(ctx context.Context, req *erp.PosPermissionRule) ([]*erp.PosPermissionRule, error)
    
    // GetPosPermissionRule 获取POS权限规则详情
    GetPosPermissionRule(ctx context.Context, ruleCode string) (*erp.PosPermissionRule, error)
    
    // CreatePosPermissionRule 创建POS权限规则
    CreatePosPermissionRule(ctx context.Context, req *erp.PosPermissionRule) (*erp.PosPermissionRule, error)
    
    // UpdatePosPermissionRule 更新POS权限规则
    UpdatePosPermissionRule(ctx context.Context, req *erp.PosPermissionRule) (*erp.PosPermissionRule, error)
    
    // DeletePosPermissionRule 删除POS权限规则
    DeletePosPermissionRule(ctx context.Context, ruleCode string) error
    
    // CheckPermission 检查权限
    CheckPermission(ctx context.Context, permissionList []erp.PermissionRule, company string) (bool, error)
}
```

## 🏗️ 实现细节

### 权限检查逻辑

权限检查遵循以下规则：

1. **如果没有权限规则**: 默认允许访问
2. **黑名单规则**: 如果公司在黑名单中，直接拒绝
3. **白名单规则**: 如果存在白名单规则，只有在白名单中的公司才能访问
4. **优先级**: 白名单优先于黑名单

```go
func (s *sPermission) CheckPermission(ctx context.Context, permissionList []erp.PermissionRule, company string) (bool, error) {
    // 如果没有权限规则，默认允许访问
    if len(permissionList) == 0 {
        return true, nil
    }
    
    // 获取所有权限规则详情
    permissionRuleList := make([]*erp.PosPermissionRule, 0)
    for _, rule := range permissionList {
        permissionRule, err := service.Permission().GetPosPermissionRule(ctx, rule.PermissionRule)
        permissionRuleList = append(permissionRuleList, permissionRule)
    }
    
    // 标记是否存在白名单规则
    hasWhiteRule := false
    inWhiteList := false
    
    // 遍历权限规则列表
    for _, rule := range permissionRuleList {
        // 检查黑名单规则
        if rule.RuleType == "Black" {
            if s.isCompanyInList(rule.CompanyList, company) {
                return false, nil // 在黑名单中，拒绝访问
            }
        }
        
        // 检查白名单规则
        if rule.RuleType == "White" {
            hasWhiteRule = true
            if s.isCompanyInList(rule.CompanyList, company) {
                inWhiteList = true // 在白名单中
            }
        }
    }
    
    // 如果存在白名单规则，则只有在白名单中的公司才能访问
    if hasWhiteRule {
        return inWhiteList, nil
    }
    
    // 如果没有白名单规则，且没有被黑名单拒绝，则允许访问
    return true, nil
}
```

### 权限规则类型

- **White（白名单）**: 只允许列表中的公司访问
- **Black（黑名单）**: 禁止列表中的公司访问

### 权限规则结构

```go
type PosPermissionRule struct {
    Name       string                  // 规则名称
    RuleCode   string                  // 规则代码
    RuleName   string                  // 规则名称
    RuleType   string                  // 规则类型（White/Black）
    CompanyList []PermissionCompanyList // 公司列表
    // ... 更多字段
}
```

## 📊 数据模型

### PosPermissionRule POS 权限规则

```go
type PosPermissionRule struct {
    Name       string                  // 规则名称
    RuleCode   string                  // 规则代码
    RuleName   string                  // 规则名称
    RuleType   string                  // 规则类型（White/Black）
    Owner      string                  // 所有者
    Creation   string                  // 创建时间
    Modified   string                  // 修改时间
    ModifiedBy string                  // 修改人
    Docstatus  int                     // 单据状态
    Idx        int                     // 索引
    Doctype    string                  // 文档类型
    CompanyList []PermissionCompanyList // 公司列表
}
```

### PermissionCompanyList 权限公司列表

```go
type PermissionCompanyList struct {
    Name        string // 名称
    Company     string // 公司名称
    Parent      string // 父级
    Parentfield string // 父级字段
    Parenttype  string // 父级类型
    Owner       string // 所有者
    Creation    string // 创建时间
    Modified    string // 修改时间
    ModifiedBy  string // 修改人
    Docstatus   int    // 单据状态
    Idx         int    // 索引
    Doctype     string // 文档类型
}
```

### PermissionRule 权限规则（简化版）

```go
type PermissionRule struct {
    PermissionRule string // 权限规则代码
}
```

## 🔄 使用流程

### 1. 创建权限规则

```go
rule, err := permissionService.CreatePosPermissionRule(ctx, &erp.PosPermissionRule{
    RuleCode: "RULE-001",
    RuleName: "测试规则",
    RuleType: "White",
    CompanyList: []erp.PermissionCompanyList{
        {Company: "CFG Company"},
        {Company: "CFG2 Company"},
    },
})
```

### 2. 查询权限规则列表

```go
rules, err := permissionService.GetPosPermissionRuleList(ctx, &erp.PosPermissionRule{
    RuleCode: "RULE-001",
})

for _, rule := range rules {
    fmt.Printf("规则: %s (%s)\n", rule.RuleName, rule.RuleType)
}
```

### 3. 检查权限

```go
hasPermission, err := permissionService.CheckPermission(ctx, []erp.PermissionRule{
    {PermissionRule: "RULE-001"},
}, "CFG Company")

if hasPermission {
    fmt.Println("有权限访问")
} else {
    fmt.Println("无权限访问")
}
```

### 4. 更新权限规则

```go
rule, err := permissionService.UpdatePosPermissionRule(ctx, &erp.PosPermissionRule{
    Name:     "RULE-001",
    RuleName: "更新后的规则名称",
    RuleType: "Black",
    CompanyList: []erp.PermissionCompanyList{
        {Company: "CFG3 Company"},
    },
})
```

## ⚠️ 注意事项

1. **规则类型**: 规则类型必须为 "White" 或 "Black"
2. **白名单优先**: 如果存在白名单规则，白名单优先
3. **黑名单检查**: 黑名单规则会直接拒绝访问
4. **空规则**: 如果没有权限规则，默认允许访问
5. **公司列表**: 权限规则可以包含多个公司

## 📝 总结

Permission 权限服务提供了完整的权限管理能力。

### 技术特点

- **白名单/黑名单**: 支持两种权限控制机制
- **优先级处理**: 白名单优先于黑名单
- **灵活配置**: 支持多个公司配置

### 设计优势

- **简单高效**: 权限检查逻辑清晰
- **易于扩展**: 支持多种权限规则类型

