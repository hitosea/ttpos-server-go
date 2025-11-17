# ProductCategory 实体模型说明

## 基本信息

- **实体名称**: ProductCategory
- **表名**: product_category
- **所属模块**: model
- **描述**: 商品分类实体，用于管理商品的层级分类结构

## 字段说明

| 字段名 | 类型 | 说明 | 备注 |
|--------|------|------|------|
| ID | uint | 主键ID | 继承自BaseModel |
| Uuid | uint64 | UUID | 继承自BaseModel |
| CreateTime | int64 | 创建时间 | 继承自BaseModel |
| UpdateTime | int64 | 更新时间 | 继承自BaseModel |
| DeleteTime | int64 | 删除时间 | 继承自BaseModel |
| CompanyUuid | uint64 | 公司UUID | 多租户标识 |
| CategoryCode | string | 分类编码 | 业务编码 |
| CategoryName | string | 分类名称 | |
| ParentUuid | string | 父分类UUID | 支持层级结构 |
| Level | int | 层级深度 | 1为根分类 |
| SortOrder | int | 排序顺序 | |
| Description | string | 描述 | 分类详细描述 |
| ImageUrl | string | 图片URL | 分类图片 |
| IconUrl | string | 图标URL | 分类图标 |
| Status | string | 状态 | active激活, inactive停用 |
| IsShow | bool | 是否显示 | 前端显示控制 |
| ProductCount | int | 商品数量 | 该分类下的商品数量 |
| MaxDepth | int | 最大深度 | 限制子分类层级 |
| AllowProduct | bool | 允许添加商品 | 是否可直接添加商品 |
| Remark | string | 备注 | |

## 关联关系

### 关联实体
- **CompanyUuid** → Company 实体（通过 CompanyUuid 关联）
- **ParentUuid** → ProductCategory 实体（自关联，父分类）

### 被关联关系
- **ProductPackage** → 通过 CategoryUuid 关联
- **Product** → 通过 CategoryUuid 关联
- 子分类 → 通过 ParentUuid 关联

## 数据流分析

### 数据来源
- 商品管理系统创建的分类
- ERP系统同步的分类数据
- 管理后台配置的分类结构

### 数据流向
1. **分类创建流程**:
   - 商品管理员创建分类
   - 设置父分类关系和层级
   - 配置分类基本信息
   - 更新父分类的商品数量统计

2. **分类更新流程**:
   - 修改分类基本信息
   - 调整分类层级关系
   - 更新排序和显示状态
   - 同步更新相关统计数据

3. **分类删除流程**:
   - 检查是否有子分类
   - 检查是否有关联商品
   - 执行软删除操作
   - 更新父分类统计信息

4. **统计更新流程**:
   - 定期更新各分类的商品数量
   - 计算分类的层级深度
   - 验证分类结构的完整性

### 业务场景
- 商品分类管理
- 多级分类结构
- 分类商品统计
- 分类显示控制
- 分类权限管理

## 索引建议

- 主键索引: ID
- 唯一索引: Uuid
- 唯一索引: CompanyUuid + CategoryCode（公司内编码唯一）
- 普通索引: CompanyUuid（公司查询）
- 普通索引: ParentUuid（父分类查询）
- 普通索引: Level（层级查询）
- 普通索引: Status（状态筛选）
- 普通索引: SortOrder（排序查询）
- 普通索引: IsShow（显示控制查询）

## WebSocket推送

### AfterUpdate钩子
- **分类更新推送**: 当分类信息变更时，推送到相关终端
- **状态变更推送**: 当分类状态变更时，推送到前端显示

### 推送目标
- **收银端**: 分类信息更新、商品分类展示
- **移动端**: 分类结构更新、商品浏览
- **管理端**: 分类管理界面、统计信息

## 业务方法

### GetChildren() []ProductCategory
- **功能**: 获取所有子分类
- **返回**: 子分类列表
- **用途**: 分类树构建

### GetParent() *ProductCategory
- **功能**: 获取父分类
- **返回**: 父分类对象
- **用途**: 分类路径构建

### GetPath() []string
- **功能**: 获取分类路径
- **返回**: 分类名称路径
- **用途**: 面包屑导航

### GetLevel() int
- **功能**: 计算分类层级
- **返回**: 层级深度
- **用途**: 分类深度验证

### HasChildren() bool
- **功能**: 判断是否有子分类
- **返回**: 是否有子分类
- **用途**: 删除权限验证

### CanDelete() bool
- **功能**: 判断是否可以删除
- **返回**: 是否可删除
- **用途**: 删除权限验证

### UpdateProductCount() error
- **功能**: 更新商品数量统计
- **返回**: 错误信息
- **用途**: 统计数据维护

## 注意事项

1. **多租户支持**: 通过CompanyUuid实现数据隔离
2. **层级结构**: 支持多级分类，需要防止循环引用
3. **软删除机制**: 使用DeleteTime实现软删除
4. **统计维护**: 商品数量需要定期更新
5. **层级限制**: 需要限制分类的最大深度
6. **权限控制**: 不同分类可能有不同的操作权限
7. **实时同步**: 通过WebSocket确保分类信息实时更新

## 业务规则

1. **编码唯一性**: 分类编码在公司内必须唯一
2. **层级限制**: 分类层级不能超过系统设定的最大深度
3. **父级验证**: 父分类不能是自身或子分类
4. **删除限制**: 有子分类或关联商品的分类不能删除
5. **状态管理**: 只有激活状态的分类才能关联商品
6. **显示控制**: IsShow字段控制前端显示，不影响后端逻辑
7. **商品限制**: AllowProduct字段控制是否可直接添加商品

## 扩展功能

### 分类权限
- 支持基于角色的分类操作权限
- 不同用户只能操作特定分类
- 权限继承机制

### 分类模板
- 支持分类模板创建
- 快速复制分类结构
- 模板版本管理

### 分类标签
- 支持分类标签系统
- 多维度分类标记
- 标签搜索和过滤

### 分类统计
- 详细的分类销售统计
- 分类趋势分析
- 分类效益分析

## 树形结构处理

### 分类树构建
```go
// 构建分类树的示例逻辑
func BuildCategoryTree(categories []ProductCategory) []CategoryTree {
    categoryMap := make(map[string]*CategoryTree)
    var roots []CategoryTree
    
    // 创建节点映射
    for _, category := range categories {
        categoryMap[category.Uuid] = &CategoryTree{
            Category: category,
            Children: []CategoryTree{},
        }
    }
    
    // 构建树形结构
    for _, category := range categories {
        node := categoryMap[category.Uuid]
        if category.ParentUuid == "" || category.ParentUuid == "0" {
            roots = append(roots, *node)
        } else if parent, exists := categoryMap[category.ParentUuid]; exists {
            parent.Children = append(parent.Children, *node)
        }
    }
    
    return roots
}
```

### 分类路径获取
- 递归查找父分类
- 构建完整路径
- 缓存路径信息

### 分类验证
- 循环引用检测
- 层级深度验证
- 数据完整性检查

## 性能优化

### 查询优化
- 使用索引优化分类查询
- 批量加载减少数据库访问
- 缓存热点分类数据

### 统计优化
- 异步更新商品数量
- 批量统计更新
- 统计数据缓存

### 树形结构优化
- 预计算分类路径
- 缓存分类树结构
- 增量更新树形数据