# 角色权限服务 (Role Access Service)

## 概述

`role_access.go` 实现了角色权限管理服务（RBAC - Role-Based Access Control），负责管理和获取员工的访问权限。该服务支持基于角色的权限控制、权限树构建、根据商家配置动态筛选权限等功能，是系统安全和访问控制的核心模块。

**文件路径**: `ttpos-server-go/main/app/service/role_access.go`

## 核心功能

### 1. 权限获取
- 获取员工的数据库权限
- 区分超级管理员和普通员工
- 支持供应商用户类型

### 2. 权限筛选
- 根据商家配置动态筛选权限
- 过滤无效和废弃的权限
- 根据授权功能控制权限可见性

### 3. 权限树构建
- 构建层级化的权限树结构
- 支持按路由名称过滤
- 支持完整权限组返回

### 4. API权限获取
- 获取员工可访问的 API 路径列表
- 用于 API 级别的权限控制

## 接口定义

### IRoleAccessSrv 接口

```go
type IRoleAccessSrv interface {
    GetPermission(routerName constant.RouteName, staffUuid, companyUuid uint64) ([]*resp.Permission, error)
    GetPermissionGroup(staffUuid, companyUuid uint64) (resp.PermissionGroup, error)
    GetApiPermission(staffUuid, companyUuid uint64) ([]string, error)
}
```

### roleAccessSrv 结构体

```go
type roleAccessSrv struct {
    dbm *database.DBManager // 数据库管理器
}
```

## 依赖项

### 内部依赖
- **repository.AccessRepo**: 权限数据仓库
- **repository.StaffRepo**: 员工数据仓库
- **repository.StaffRoleRepo**: 员工角色关联仓库

### 外部依赖
- **database.DBManager**: 数据库管理器
- **copier**: 结构体拷贝库

## 核心方法详解

### 1. GetPermission - 获取特定路由的权限

**方法签名**:
```go
func (s *roleAccessSrv) GetPermission(routerName constant.RouteName, staffUuid, companyUuid uint64) ([]*resp.Permission, error)
```

**功能**: 获取员工在特定路由下的权限列表，返回树形结构。

**参数说明**:
- `routerName`: 路由名称（如 "cashier"、"shop"、"admin" 等）
- `staffUuid`: 员工 UUID
- `companyUuid`: 公司 UUID

**返回值**:
```go
type Permission struct {
    Uuid        uint64        // 权限UUID
    ID          int           // 权限ID
    ParentUuid  uint64        // 父级UUID
    Name        string        // 权限名称
    Title       string        // 权限标题
    Path        string        // API路径
    Icon        string        // 图标
    Type        int           // 类型（1-菜单，2-按钮）
    Sort        int           // 排序
    CreateTime  string        // 创建时间
    UpdateTime  string        // 更新时间
    Children    []*Permission // 子权限列表
}
```

**实现流程**:

```81:101:ttpos-server-go/main/app/service/role_access.go
func (s *roleAccessSrv) GetPermission(routerName constant.RouteName, staffUuid, companyUuid uint64) ([]*resp.Permission, error) {

	var permissions []resp.Permission
	dbPermissions, companySetting, err := s.getDbPermissions(staffUuid, companyUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	for _, dbPermission := range dbPermissions {
		var permission resp.Permission
		copier.Copy(&permission, dbPermission)
		permission.CreateTime = time.Unix(dbPermission.CreateTime, 0).Format(time.DateTime)
		permission.UpdateTime = time.Unix(dbPermission.UpdateTime, 0).Format(time.DateTime)
		permission.Children = []*resp.Permission{}
		permissions = append(permissions, permission)
	}

	permissions = s.filterPermission(permissions, companySetting)

	return s.buildPermissionTree(permissions, routerName), nil
}
```

**处理流程**:
1. 从数据库获取员工的原始权限列表
2. 转换数据格式，格式化时间字段
3. 根据商家配置筛选权限
4. 构建权限树并按路由名称过滤
5. 返回树形结构的权限列表

**使用场景**:
- 收银端获取菜单权限：`GetPermission("cashier", staffUuid, companyUuid)`
- 后台管理获取权限：`GetPermission("admin", staffUuid, companyUuid)`
- 商家后台获取权限：`GetPermission("shop", staffUuid, companyUuid)`

---

### 2. GetPermissionGroup - 获取完整权限组

**方法签名**:
```go
func (s *roleAccessSrv) GetPermissionGroup(staffUuid, companyUuid uint64) (resp.PermissionGroup, error)
```

**功能**: 获取员工的完整权限组，不按路由过滤，返回所有权限的树形结构。

**返回值**:
```go
type PermissionGroup struct {
    List []*Permission // 权限树列表（多个根节点）
}
```

**实现流程**:

```104:137:ttpos-server-go/main/app/service/role_access.go
func (s *roleAccessSrv) GetPermissionGroup(staffUuid, companyUuid uint64) (resp.PermissionGroup, error) {

	var permissions []resp.Permission
	groupPermission := resp.PermissionGroup{
		List: []*resp.Permission{},
	}
	dbPermissions, companySetting, err := s.getDbPermissions(staffUuid, companyUuid)
	if err != nil {
		return groupPermission, errors.WithMessage(err)
	}

	for _, dbPermission := range dbPermissions {
		var permission resp.Permission
		copier.Copy(&permission, dbPermission)
		// 手动设置ParentUuid字段，因为json标签不匹配导致copier无法正确复制
		permission.ParentUuid = dbPermission.ParentUuid
		permission.CreateTime = time.Unix(dbPermission.CreateTime, 0).Format(time.DateTime)
		permission.UpdateTime = time.Unix(dbPermission.UpdateTime, 0).Format(time.DateTime)
		permission.Children = []*resp.Permission{}
		permissions = append(permissions, permission)
	}

	// 筛选权限
	permissions = s.filterPermission(permissions, companySetting)

	// 构建权限树形结构
	roots := s.buildPermissionTreeWithoutFilter(permissions)
	for _, root := range roots {
		groupPermission.List = append(groupPermission.List, root)
	}

	// 返回权限组
	return groupPermission, nil
}
```

**关键点**:
1. 手动设置 `ParentUuid` 字段（copier 因 JSON 标签不匹配无法自动复制）
2. 不进行路由过滤，返回所有权限
3. 返回多个根节点的权限树

**与 GetPermission 的区别**:
- `GetPermission`: 返回特定路由的权限（单个路由的子权限）
- `GetPermissionGroup`: 返回所有权限（多个根节点）

**使用场景**:
- 权限配置页面：展示所有可分配的权限
- 角色管理：配置角色权限时使用
- 权限总览：查看员工的所有权限

---

### 3. GetApiPermission - 获取 API 权限列表

**方法签名**:
```go
func (s *roleAccessSrv) GetApiPermission(staffUuid, companyUuid uint64) ([]string, error)
```

**功能**: 获取员工可访问的 API 路径列表，用于 API 级别的权限验证。

**返回值**: `[]string` - API 路径列表
```go
// 示例
[
    "/api/v1/order/list",
    "/api/v1/product/create",
    "/api/v1/member/info"
]
```

**实现流程**:

```238:250:ttpos-server-go/main/app/service/role_access.go
func (s *roleAccessSrv) GetApiPermission(staffUuid, companyUuid uint64) ([]string, error) {
	accesses, _, err := s.getDbPermissions(staffUuid, companyUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	var permissions []string
	for _, access := range accesses {
		if !slices.Contains(permissions, access.Path) {
			permissions = append(permissions, access.Path)
		}
	}
	return permissions, nil
}
```

**关键点**:
1. 从数据库权限中提取 Path 字段
2. 自动去重（同一路径只返回一次）
3. 返回扁平化的路径列表

**使用场景**:
- 中间件权限验证：检查用户是否有权限访问某个 API
- 前端路由守卫：判断是否显示某个按钮或菜单
- API 网关：进行细粒度的访问控制

---

### 4. getDbPermissions - 从数据库获取权限（私有方法）

**方法签名**:
```go
func (s *roleAccessSrv) getDbPermissions(staffUuid, companyUuid uint64) ([]model.Access, model.CompanySetting, error)
```

**功能**: 从数据库获取员工的原始权限数据，是所有公开方法的基础。

**返回值**:
- `[]model.Access`: 权限列表
- `model.CompanySetting`: 商家配置
- `error`: 错误信息

**实现流程**:

```40:78:ttpos-server-go/main/app/service/role_access.go
func (s *roleAccessSrv) getDbPermissions(staffUuid, companyUuid uint64) ([]model.Access, model.CompanySetting, error) {
	db := s.dbm.GetDB(companyUuid)
	accessRepo := repository.NewAccessRepo(db)
	var companySetting model.CompanySetting
	staffRepo := repository.NewStaffRepo(db)
	staff, _ := staffRepo.GetStaff(staffRepo.WhereUuid(staffUuid), staffRepo.WithCompany(), staffRepo.WithCompanySetting())

	if staff.Company == nil || staff.Company.CompanySetting == nil {
		return nil, companySetting, errors.New("获取商家信息错误")
	}

	companySetting = *staff.Company.CompanySetting

	var options []repository.DBOption

	if staff.IsSuper == 1 { // 超级管理员
		if staff.UserType == 1 {
			options = append(options, accessRepo.WhereIsSupplier())
		}
	} else {
		roleUuids, err := repository.NewStaffRoleRepo(s.dbm.GetDB(staff.CompanyUuid)).GetRoleUuidsByStaffUuid(staff.Uuid)
		if err != nil {
			return nil, companySetting, errors.WithMessage(err, "获取用户角色失败")
		}
		accessUuids, err := accessRepo.GetAccessUuids(roleUuids)
		if err != nil {
			return nil, companySetting, errors.WithMessage(err, "获取角色权限失败")
		}
		options = append(options, accessRepo.WhereUuids(accessUuids))
	}

	dbPermissions, err := accessRepo.GetPermissions(options...)

	if err != nil {
		return nil, companySetting, errors.WithMessage(err, "获取权限失败")
	}

	return dbPermissions, companySetting, nil
}
```

**权限获取逻辑**:

#### 1. 超级管理员
```go
if staff.IsSuper == 1 {
    if staff.UserType == 1 {
        // 供应商类型的超级管理员
        options = append(options, accessRepo.WhereIsSupplier())
    } else {
        // 普通超级管理员 - 获取所有权限
    }
}
```

**超级管理员权限**:
- `IsSuper = 1` 表示超级管理员
- `UserType = 1` 表示供应商类型，只获取供应商相关权限
- `UserType != 1` 获取所有权限（无筛选条件）

#### 2. 普通员工
```go
else {
    // 1. 获取员工的角色列表
    roleUuids := GetRoleUuidsByStaffUuid(staff.Uuid)
    
    // 2. 根据角色获取权限UUID列表
    accessUuids := GetAccessUuids(roleUuids)
    
    // 3. 根据权限UUID获取权限详情
    options = append(options, accessRepo.WhereUuids(accessUuids))
}
```

**普通员工权限流程**:
1. 查询员工关联的所有角色
2. 查询这些角色关联的所有权限 UUID
3. 根据权限 UUID 查询完整的权限信息

**数据查询流程图**:
```
员工 → 员工角色关联表 → 角色 → 角色权限关联表 → 权限
Staff → StaffRole → Role → RoleAccess → Access
```

---

### 5. filterPermission - 筛选权限（私有方法）

**方法签名**:
```go
func (s *roleAccessSrv) filterPermission(permissions []resp.Permission, companySetting model.CompanySetting) []resp.Permission
```

**功能**: 根据商家配置和系统规则筛选权限，动态控制权限可见性。

**筛选规则**:

```140:191:ttpos-server-go/main/app/service/role_access.go
func (s *roleAccessSrv) filterPermission(permissions []resp.Permission, companySetting model.CompanySetting) []resp.Permission {
	var filteredPermissions []resp.Permission
	for _, permission := range permissions {
		// 删除无效数据
		if slices.Contains([]uint64{58, 124, 125, 128, 129, 160, 162, 1724320603, 1724320604, 1724320605}, permission.Uuid) {
			continue
		}

		// 暂时去掉外卖管理
		if permission.ID == 1626688443 {
			continue
		}
		// 授权无进销存权限
		if companySetting.SaleStock == 0 && slices.Contains([]int{1711006072, 1711009130}, permission.ID) {
			continue
		}
		// 授权无会员权限
		if companySetting.IsOpenMember == 0 && slices.Contains([]int{1636183779, 1704881218}, permission.ID) {
			continue
		}
		// 授权无平板点餐权限
		if companySetting.IsOpenTablet == 0 && permission.ID == 87 {
			continue
		}
		// 授权无H5点餐权限
		if companySetting.IsOpenH5 == 0 && permission.ID == 1724220505 {
			continue
		}
		// 授权无点餐助手权限
		if companySetting.IsOpenAssistant == 0 && permission.ID == 1720753338 {
			continue
		}
		// 授权无后厨权限
		if companySetting.IsOpenKitchenKds == 0 && permission.ID == 88 {
			continue
		}
		// 授权无自助餐权限
		if companySetting.IsOpenBuffet == 0 && permission.ID == 1708671616 {
			continue
		}
		// 授权无扫码点餐接单权限
		if companySetting.IsOpenH5Order == 0 && permission.ID == 1724320522 {
			continue
		}
		// 授权无外送权限
		if companySetting.DeliveryStatus != 1 && permission.Uuid == 1752716650 {
			continue
		}
		filteredPermissions = append(filteredPermissions, permission)
	}
	return filteredPermissions
}
```

**筛选规则详解**:

#### 1. 删除无效权限
```go
// 硬编码的无效权限UUID列表
invalidUuids := []uint64{58, 124, 125, 128, 129, 160, 162, 1724320603, 1724320604, 1724320605}
```
- 这些权限已废弃或数据异常
- 直接从结果中移除

#### 2. 功能模块权限控制

| 配置字段 | 条件 | 权限ID | 功能模块 |
|---------|------|--------|---------|
| `SaleStock` | = 0 | 1711006072, 1711009130 | 进销存管理 |
| `IsOpenMember` | = 0 | 1636183779, 1704881218 | 会员管理 |
| `IsOpenTablet` | = 0 | 87 | 平板点餐 |
| `IsOpenH5` | = 0 | 1724220505 | H5点餐 |
| `IsOpenAssistant` | = 0 | 1720753338 | 点餐助手 |
| `IsOpenKitchenKds` | = 0 | 88 | 后厨显示 |
| `IsOpenBuffet` | = 0 | 1708671616 | 自助餐 |
| `IsOpenH5Order` | = 0 | 1724320522 | 扫码点餐接单 |
| `DeliveryStatus` | != 1 | 1752716650 | 外送功能 |

#### 3. 临时移除功能
```go
// 暂时去掉外卖管理
if permission.ID == 1626688443 {
    continue
}
```
- 功能开发中或维护中
- 临时禁用

**筛选目的**:
1. **授权控制**: 商家未购买的功能不显示权限
2. **数据清理**: 移除历史遗留的无效权限
3. **功能开关**: 根据商家配置动态启用/禁用功能

---

### 6. buildPermissionTree - 构建权限树（私有方法）

**方法签名**:
```go
func (s *roleAccessSrv) buildPermissionTree(permissions []resp.Permission, routerName constant.RouteName) []*resp.Permission
```

**功能**: 构建权限树并按路由名称过滤，只返回指定路由的子权限。

**实现流程**:

```194:203:ttpos-server-go/main/app/service/role_access.go
func (s *roleAccessSrv) buildPermissionTree(permissions []resp.Permission, routerName constant.RouteName) []*resp.Permission {
	roots := s.buildPermissionTreeWithoutFilter(permissions)
	var filteredRoots []*resp.Permission
	for _, root := range roots {
		if root.Name == string(routerName) {
			filteredRoots = append(filteredRoots, root.Children...)
		}
	}
	return filteredRoots
}
```

**处理逻辑**:
1. 先构建完整的权限树
2. 查找名称匹配 `routerName` 的根节点
3. 返回该根节点的 `Children`（子权限）

**示例**:
```go
// 原始树结构
[
    {Name: "cashier", Children: [权限A, 权限B]},
    {Name: "shop", Children: [权限C, 权限D]},
    {Name: "admin", Children: [权限E, 权限F]}
]

// 调用 buildPermissionTree(permissions, "cashier")
// 返回: [权限A, 权限B]
```

---

### 7. buildPermissionTreeWithoutFilter - 构建完整权限树（私有方法）

**方法签名**:
```go
func (s *roleAccessSrv) buildPermissionTreeWithoutFilter(permissions []resp.Permission) []*resp.Permission
```

**功能**: 将扁平化的权限列表构建成树形结构，不进行任何过滤。

**实现流程**:

```206:236:ttpos-server-go/main/app/service/role_access.go
func (s *roleAccessSrv) buildPermissionTreeWithoutFilter(permissions []resp.Permission) []*resp.Permission {
	permissionMap := make(map[uint64]*resp.Permission)
	var roots []*resp.Permission
	var accessIds []string
	format := "%03d%020d"

	// 第一步：建立ID到节点的映射
	for i := range permissions {
		permission := &permissions[i]
		permissionMap[permission.Uuid] = permission
		accessIds = append(accessIds, fmt.Sprintf(format, permission.Sort, permission.Uuid))
	}

	sort.Strings(accessIds)
	// 第二步：构建树结构

	for _, accessId := range accessIds {
		for _, permission := range permissionMap {
			if fmt.Sprintf(format, permission.Sort, permission.Uuid) != accessId {
				continue
			}
			if parent, exists := permissionMap[permission.ParentUuid]; exists {
				parent.Children = append(parent.Children, permission)
			} else {
				roots = append(roots, permission)
			}
		}
	}

	return roots
}
```

**构建步骤**:

#### 第一步：建立映射和排序
```go
format := "%03d%020d"
// 格式：3位排序号 + 20位UUID
// 例如：001 00000000000000000123

for i := range permissions {
    permission := &permissions[i]
    permissionMap[permission.Uuid] = permission // UUID → Permission 映射
    accessIds = append(accessIds, fmt.Sprintf(format, permission.Sort, permission.Uuid))
}

sort.Strings(accessIds) // 按排序号+UUID排序
```

**排序规则**:
- 首先按 `Sort` 字段排序（3位数字）
- `Sort` 相同时按 `Uuid` 排序
- 保证权限按预定义的顺序显示

#### 第二步：构建树结构
```go
for _, accessId := range accessIds {
    for _, permission := range permissionMap {
        if permission匹配accessId {
            if 存在父节点 {
                父节点.Children.append(当前节点)
            } else {
                roots.append(当前节点) // 根节点
            }
        }
    }
}
```

**树构建逻辑**:
1. 按排序后的顺序遍历权限
2. 查找每个权限的父节点
3. 如果找到父节点，添加到父节点的 `Children`
4. 如果找不到父节点（`ParentUuid` 不存在），则为根节点

**数据结构示例**:

**输入（扁平列表）**:
```go
[
    {Uuid: 1, ParentUuid: 0, Name: "cashier", Sort: 10},
    {Uuid: 2, ParentUuid: 1, Name: "order", Sort: 20},
    {Uuid: 3, ParentUuid: 1, Name: "product", Sort: 30},
    {Uuid: 4, ParentUuid: 2, Name: "order.create", Sort: 21},
]
```

**输出（树形结构）**:
```go
[
    {
        Uuid: 1, Name: "cashier", Sort: 10,
        Children: [
            {
                Uuid: 2, Name: "order", Sort: 20,
                Children: [
                    {Uuid: 4, Name: "order.create", Sort: 21, Children: []}
                ]
            },
            {Uuid: 3, Name: "product", Sort: 30, Children: []}
        ]
    }
]
```

---

## 权限数据模型

### Access - 权限表

```go
type Access struct {
    Uuid        uint64 `gorm:"primary_key"` // 权限UUID
    ID          int                         // 权限ID（业务ID）
    ParentUuid  uint64                      // 父级UUID
    Name        string                      // 权限名称（英文标识）
    Title       string                      // 权限标题（显示名称）
    Path        string                      // API路径
    Icon        string                      // 图标
    Type        int                         // 类型（1-菜单，2-按钮）
    Sort        int                         // 排序
    IsSupplier  int                         // 是否供应商权限
    CreateTime  int64                       // 创建时间
    UpdateTime  int64                       // 更新时间
}
```

### Staff - 员工表

```go
type Staff struct {
    Uuid        uint64 // 员工UUID
    CompanyUuid uint64 // 公司UUID
    IsSuper     int    // 是否超级管理员（0-否，1-是）
    UserType    int    // 用户类型（0-普通，1-供应商）
    
    // 关联
    Company *Company
}
```

### StaffRole - 员工角色关联表

```go
type StaffRole struct {
    StaffUuid uint64 // 员工UUID
    RoleUuid  uint64 // 角色UUID
}
```

### RoleAccess - 角色权限关联表

```go
type RoleAccess struct {
    RoleUuid   uint64 // 角色UUID
    AccessUuid uint64 // 权限UUID
}
```

### CompanySetting - 商家配置表

```go
type CompanySetting struct {
    SaleStock         int // 是否开启进销存（0-否，1-是）
    IsOpenMember      int // 是否开启会员（0-否，1-是）
    IsOpenTablet      int // 是否开启平板点餐（0-否，1-是）
    IsOpenH5          int // 是否开启H5点餐（0-否，1-是）
    IsOpenAssistant   int // 是否开启点餐助手（0-否，1-是）
    IsOpenKitchenKds  int // 是否开启后厨显示（0-否，1-是）
    IsOpenBuffet      int // 是否开启自助餐（0-否，1-是）
    IsOpenH5Order     int // 是否开启扫码点餐接单（0-否，1-是）
    DeliveryStatus    int // 外送状态（0-关闭，1-开启）
}
```

---

## RBAC 权限模型

### 权限模型结构

```
员工 (Staff) ←→ 角色 (Role) ←→ 权限 (Access)
     N : M           N : M
```

### 数据表关系

```
┌─────────┐     ┌──────────────┐     ┌──────┐     ┌─────────────┐     ┌────────┐
│  Staff  │────→│  StaffRole   │────→│ Role │────→│ RoleAccess  │────→│ Access │
└─────────┘     └──────────────┘     └──────┘     └─────────────┘     └────────┘
  员工表          员工角色关联表       角色表         角色权限关联表       权限表
```

### 权限类型

#### 1. 菜单权限 (Type = 1)
- 用于控制菜单显示
- 有层级关系（父菜单 → 子菜单）
- 示例：订单管理、商品管理

#### 2. 按钮权限 (Type = 2)
- 用于控制按钮/操作显示
- 通常是菜单的子权限
- 示例：新增订单、删除商品

### 权限层级

```
根权限（路由）
├── 一级菜单
│   ├── 二级菜单
│   │   ├── 按钮权限1
│   │   └── 按钮权限2
│   └── 二级菜单
└── 一级菜单
```

**示例**:
```
cashier（收银端）
├── 订单管理
│   ├── 订单列表
│   │   ├── 查看订单
│   │   ├── 创建订单
│   │   └── 取消订单
│   └── 订单统计
└── 商品管理
    ├── 商品列表
    └── 商品分类
```

---

## 权限判断流程

### 1. 超级管理员判断

```
开始
  ↓
检查 IsSuper = 1?
  ↓ 是
检查 UserType = 1?
  ↓ 是              ↓ 否
供应商权限        所有权限
  ↓                 ↓
返回权限列表
```

### 2. 普通员工权限获取

```
开始
  ↓
查询员工角色 (StaffRole)
  ↓
查询角色权限 (RoleAccess)
  ↓
查询权限详情 (Access)
  ↓
筛选权限 (filterPermission)
  ↓
构建权限树
  ↓
返回权限列表
```

### 3. 权限筛选流程

```
原始权限列表
  ↓
移除无效权限（硬编码列表）
  ↓
检查商家配置
  ↓
移除未授权功能的权限
  ↓
返回筛选后权限列表
```

---

## 使用场景

### 场景1: 收银端菜单加载

```go
// 用户登录收银端
staffUuid := 12345
companyUuid := 67890

// 获取收银端权限
permissions, err := roleAccessSrv.GetPermission("cashier", staffUuid, companyUuid)

// 前端根据权限渲染菜单
for _, permission := range permissions {
    if permission.Type == 1 { // 菜单类型
        renderMenu(permission)
    }
}
```

### 场景2: 按钮权限控制

```go
// 检查是否有"创建订单"权限
permissions, _ := roleAccessSrv.GetPermission("cashier", staffUuid, companyUuid)

hasCreateOrderPermission := false
for _, permission := range permissions {
    if permission.Path == "/api/v1/order/create" {
        hasCreateOrderPermission = true
        break
    }
}

if hasCreateOrderPermission {
    showCreateOrderButton()
}
```

### 场景3: API 权限验证

```go
// 中间件验证 API 权限
func PermissionMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        staffUuid := c.GetUint64("staff_uuid")
        companyUuid := c.GetUint64("company_uuid")
        requestPath := c.Request.URL.Path
        
        // 获取员工的 API 权限列表
        apiPermissions, err := roleAccessSrv.GetApiPermission(staffUuid, companyUuid)
        if err != nil {
            c.AbortWithStatus(403)
            return
        }
        
        // 检查是否有权限访问
        if !slices.Contains(apiPermissions, requestPath) {
            c.AbortWithStatus(403)
            return
        }
        
        c.Next()
    }
}
```

### 场景4: 角色权限配置

```go
// 管理员配置角色权限
// 1. 获取所有权限组
permissionGroup, err := roleAccessSrv.GetPermissionGroup(adminStaffUuid, companyUuid)

// 2. 展示权限树供选择
displayPermissionTree(permissionGroup.List)

// 3. 用户选择权限后保存到角色
selectedAccessUuids := []uint64{1, 2, 3, 4, 5}
saveRoleAccess(roleUuid, selectedAccessUuids)
```

### 场景5: 多端权限隔离

```go
// 收银端
cashierPermissions, _ := roleAccessSrv.GetPermission("cashier", staffUuid, companyUuid)

// 商家后台
shopPermissions, _ := roleAccessSrv.GetPermission("shop", staffUuid, companyUuid)

// 系统后台
adminPermissions, _ := roleAccessSrv.GetPermission("admin", staffUuid, companyUuid)

// 每个端只能看到自己路由下的权限
```

---

## 最佳实践

### 1. 权限缓存

```go
// 推荐：缓存员工权限，避免频繁查询
type CachedPermissionSrv struct {
    roleAccessSrv IRoleAccessSrv
    cache         cache.Cache
}

func (s *CachedPermissionSrv) GetPermission(routerName constant.RouteName, staffUuid, companyUuid uint64) ([]*resp.Permission, error) {
    // 缓存键
    cacheKey := fmt.Sprintf("permission:%d:%d:%s", companyUuid, staffUuid, routerName)
    
    // 尝试从缓存获取
    if cached, err := s.cache.Get(cacheKey); err == nil {
        return cached.([]*resp.Permission), nil
    }
    
    // 从数据库获取
    permissions, err := s.roleAccessSrv.GetPermission(routerName, staffUuid, companyUuid)
    if err != nil {
        return nil, err
    }
    
    // 缓存5分钟
    s.cache.Set(cacheKey, permissions, 5*time.Minute)
    
    return permissions, nil
}
```

### 2. 权限变更清理缓存

```go
// 角色权限变更时清理缓存
func UpdateRoleAccess(roleUuid uint64, accessUuids []uint64) error {
    // 更新数据库
    err := roleAccessRepo.UpdateRoleAccess(roleUuid, accessUuids)
    if err != nil {
        return err
    }
    
    // 清理该角色下所有员工的权限缓存
    staffUuids := staffRoleRepo.GetStaffUuidsByRoleUuid(roleUuid)
    for _, staffUuid := range staffUuids {
        cache.Delete(fmt.Sprintf("permission:*:%d:*", staffUuid))
    }
    
    return nil
}
```

### 3. 前端权限指令

```javascript
// Vue 自定义指令
Vue.directive('permission', {
    mounted(el, binding) {
        const requiredPermission = binding.value
        const userPermissions = store.state.permissions
        
        const hasPermission = userPermissions.some(p => p.path === requiredPermission)
        
        if (!hasPermission) {
            el.parentNode?.removeChild(el)
        }
    }
})

// 使用
<button v-permission="'/api/v1/order/create'">创建订单</button>
```

### 4. 动态路由生成

```javascript
// 根据权限动态生成路由
function generateRoutes(permissions) {
    const routes = []
    
    permissions.forEach(permission => {
        if (permission.type === 1) { // 菜单类型
            routes.push({
                path: permission.path,
                name: permission.name,
                component: () => import(`@/views${permission.path}`),
                meta: {
                    title: permission.title,
                    icon: permission.icon
                },
                children: generateRoutes(permission.children)
            })
        }
    })
    
    return routes
}
```

---

## 性能优化

### 1. 查询优化

**问题**: 多次数据库查询
```go
// 不推荐：N+1 查询
for _, staffUuid := range staffUuids {
    permissions, _ := GetPermission("cashier", staffUuid, companyUuid)
    // 处理权限
}
```

**优化**: 批量查询
```go
// 推荐：批量查询
staffPermissions := BatchGetPermissions(staffUuids, companyUuid)
for staffUuid, permissions := range staffPermissions {
    // 处理权限
}
```

### 2. 预加载关联数据

```go
// 获取员工时预加载公司和配置
staff, _ := staffRepo.GetStaff(
    staffRepo.WhereUuid(staffUuid),
    staffRepo.WithCompany(),        // 预加载公司
    staffRepo.WithCompanySetting(), // 预加载配置
)
```

### 3. 权限树构建优化

当前实现的时间复杂度: O(n²)
```go
// 当前实现
for _, accessId := range accessIds {           // O(n)
    for _, permission := range permissionMap { // O(n)
        // 查找和添加
    }
}
```

**优化建议**: 使用哈希查找
```go
// 优化后: O(n)
for _, accessId := range accessIds {
    permission := permissionMap[extractUuid(accessId)] // O(1)
    if parent, exists := permissionMap[permission.ParentUuid]; exists {
        parent.Children = append(parent.Children, permission)
    } else {
        roots = append(roots, permission)
    }
}
```

---

## 错误处理

### 1. 常见错误

| 错误场景 | 错误消息 | 处理方式 |
|---------|---------|---------|
| 商家信息缺失 | "获取商家信息错误" | 检查员工数据完整性 |
| 获取角色失败 | "获取用户角色失败" | 检查 StaffRole 表 |
| 获取权限失败 | "获取角色权限失败" | 检查 RoleAccess 表 |
| 权限查询失败 | "获取权限失败" | 检查 Access 表 |

### 2. 错误处理示例

```go
permissions, err := roleAccessSrv.GetPermission("cashier", staffUuid, companyUuid)
if err != nil {
    // 记录详细日志
    logger.Logger.Error("获取权限失败",
        zap.Error(err),
        zap.Uint64("staff_uuid", staffUuid),
        zap.Uint64("company_uuid", companyUuid),
        zap.String("router_name", "cashier"),
    )
    
    // 返回默认权限或拒绝访问
    return []resp.Permission{}, err
}
```

---

## 安全考虑

### 1. 权限验证层次

```
┌─────────────────────┐
│   前端权限控制      │ ← 用户体验（可绕过）
├─────────────────────┤
│   API权限中间件     │ ← 第一道防线
├─────────────────────┤
│   业务逻辑权限检查   │ ← 第二道防线
├─────────────────────┤
│   数据库权限控制    │ ← 最后防线
└─────────────────────┘
```

### 2. 防止越权访问

```go
// 不仅检查权限存在，还要检查资源归属
func GetOrder(staffUuid, companyUuid, orderUuid uint64) (*Order, error) {
    // 1. 检查API权限
    apiPermissions, _ := roleAccessSrv.GetApiPermission(staffUuid, companyUuid)
    if !slices.Contains(apiPermissions, "/api/v1/order/detail") {
        return nil, errors.New("无权限")
    }
    
    // 2. 检查资源归属
    order, _ := orderRepo.GetOrder(orderUuid)
    if order.CompanyUuid != companyUuid {
        return nil, errors.New("无权访问该订单")
    }
    
    return order, nil
}
```

### 3. 超级管理员限制

```go
// 限制超级管理员的某些危险操作
func DeleteCompany(staffUuid, companyUuid uint64) error {
    staff, _ := staffRepo.GetStaff(staffUuid)
    
    // 即使是超级管理员，也需要额外验证
    if staff.IsSuper == 1 {
        // 要求二次验证
        if !verifySecondaryAuth(staffUuid) {
            return errors.New("需要二次验证")
        }
    }
    
    // 执行删除
    return companyRepo.Delete(companyUuid)
}
```

---

## 潜在改进点

### 1. 权限表达式支持

**当前**: 硬编码权限ID
**改进**: 支持权限表达式
```go
// 当前
if hasPermission("/api/v1/order/create") { }

// 改进后
if hasPermission("order:create") { }
if hasPermission("order:*") { } // 通配符
if hasPermission("order:create,update") { } // 多个权限
```

### 2. 动态权限规则

**当前**: 静态权限配置
**改进**: 支持动态规则
```go
type PermissionRule struct {
    Resource   string // 资源类型
    Action     string // 操作
    Condition  string // 条件（如：owner_only）
}

// 示例：只能查看自己创建的订单
{
    Resource: "order",
    Action: "view",
    Condition: "created_by = current_user"
}
```

### 3. 权限继承

**改进**: 支持权限继承
```go
// 角色继承
type Role struct {
    Uuid       uint64
    Name       string
    ParentUuid uint64 // 继承父角色的权限
}
```

### 4. 临时权限授予

**改进**: 支持临时权限
```go
type TemporaryPermission struct {
    StaffUuid  uint64
    AccessUuid uint64
    ExpireTime time.Time // 过期时间
}
```

### 5. 权限审计日志

**改进**: 记录权限变更历史
```go
type PermissionAuditLog struct {
    StaffUuid     uint64
    Action        string // grant, revoke
    AccessUuid    uint64
    OperatorUuid  uint64
    OperationTime time.Time
}
```

### 6. 权限组管理

**改进**: 支持权限组（预设权限集合）
```go
type PermissionGroup struct {
    Name        string
    Description string
    AccessUuids []uint64
}

// 预设权限组
groups := []PermissionGroup{
    {Name: "收银员", AccessUuids: [...]},
    {Name: "店长", AccessUuids: [...]},
    {Name: "财务", AccessUuids: [...]},
}
```

---

## 相关文件

### DTO 定义
- `ttpos-server-go/app/dto/resp/permission.go` - 权限响应数据

### 数据仓库
- `ttpos-server-go/app/repository/access.go` - 权限仓库
- `ttpos-server-go/app/repository/staff.go` - 员工仓库
- `ttpos-server-go/app/repository/staff_role.go` - 员工角色仓库

### 数据模型
- `ttpos-server-go/app/model/access.go` - 权限模型
- `ttpos-server-go/app/model/staff.go` - 员工模型
- `ttpos-server-go/app/model/role.go` - 角色模型

### 常量定义
- `ttpos-server-go/app/constant/router.go` - 路由名称常量

---

## 总结

角色权限服务是系统安全和访问控制的核心，具有以下特点：

1. **灵活的RBAC模型**: 支持员工-角色-权限的多对多关系
2. **超级管理员支持**: 区分普通超管和供应商超管
3. **动态权限筛选**: 根据商家配置自动控制权限可见性
4. **树形结构构建**: 支持层级化的权限展示
5. **多端权限隔离**: 不同客户端（收银、商家后台、系统后台）权限独立
6. **完善的排序机制**: 权限按 Sort 字段有序展示
7. **API权限控制**: 支持接口级别的细粒度权限验证
8. **商家授权管理**: 根据商家购买的功能模块控制权限

该服务为整个系统提供了统一的权限管理能力，确保不同角色的员工只能访问被授权的功能和数据。

