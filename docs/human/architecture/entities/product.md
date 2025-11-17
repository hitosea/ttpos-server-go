# Product 实体模型说明

## 基本信息

- **实体名称**: Product
- **表名**: product
- **所属模块**: model
- **描述**: 商品实体，用于管理商品的完整信息，包括基本信息、BOM、口味、酱料等

## 主要字段说明

| 字段名 | 类型 | 说明 | 备注 |
|--------|------|------|------|
| ID | uint | 主键ID | 继承自BaseModel |
| Uuid | uint64 | UUID | 继承自BaseModel |
| CreateTime | int64 | 创建时间 | 继承自BaseModel |
| UpdateTime | int64 | 更新时间 | 继承自BaseModel |
| DeleteTime | int64 | 删除时间 | 继承自BaseModel |
| CompanyUuid | uint64 | 公司UUID | 多租户标识 |
| ProductCode | string | 商品编码 | 业务编码 |
| ProductName | string | 商品名称 | |
| CategoryUuid | string | 分类UUID | 关联商品分类 |
| ProductType | string | 商品类型 | single单品, combo套餐, service服务 |
| Status | string | 状态 | active激活, inactive停用, discontinued停产 |
| Price | float64 | 销售价格 | |
| CostPrice | float64 | 成本价格 | |
| Unit | string | 单位 | 个、份、瓶等 |
| Description | string | 商品描述 | |
| ImageUrl | string | 商品图片 | |
| SortOrder | int | 排序顺序 | |
| Tags | string | 标签 | JSON格式存储 |
| IsRecommended | bool | 是否推荐 | |
| BarCode | string | 条形码 | |
| StockCount | int | 库存数量 | |
| MinStock | int | 最低库存 | |
| MaxStock | int | 最高库存 | |
| IsTrackStock | bool | 是否追踪库存 | |
| Remark | string | 备注 | |

## 关联实体

### 主要关联
- **CompanyUuid** → Company 实体（通过 CompanyUuid 关联）
- **CategoryUuid** → ProductCategory 实体（通过 CategoryUuid 关联）

### 子实体关联
- **ProductBom** → 商品BOM（一对多关系）
- **ProductFlavor** → 商品口味（一对多关系）
- **ProductSauce** → 商品酱料（一对多关系）

### 被关联关系
- **SaleOrderItem** → 通过商品UUID关联
- **ProductPackage** → 通过商品UUID关联

## ProductBom 子实体

### 基本信息
- **ParentUuid**: 父商品UUID
- **ChildUuid**: 子商品UUID
- **Quantity**: 数量
- **Unit**: 单位

### 数据流分析
- BOM关系定义商品的组成结构
- 支持多级BOM嵌套
- 用于套餐和组合商品管理

## ProductFlavor 子实体

### 基本信息
- **ProductUuid**: 商品UUID
- **FlavorName**: 口味名称
- **FlavorCode**: 口味编码
- **Price**: 价格差
- **Status**: 状态
- **SortOrder**: 排序

### 数据流分析
- 定义商品的可选口味
- 支持口味价格差异化
- 用于个性化定制

## ProductSauce 子实体

### 基本信息
- **ProductUuid**: 商品UUID
- **SauceName**: 酱料名称
- **SauceCode**: 酱料编码
- **Price**: 价格差
- **Status**: 状态
- **SortOrder**: 排序

### 数据流分析
- 定义商品的可选酱料
- 支持酱料价格差异化
- 用于个性化定制

## 数据流分析

### 数据来源
- 商品管理系统创建的商品信息
- ERP系统同步的商品数据
- 供应商提供的商品信息
- 库存管理系统同步

### 数据流向
1. **商品创建流程**:
   - 商品管理员创建商品
   - 设置商品基本信息和价格
   - 配置商品分类和属性
   - 设置库存信息
   - 配置BOM、口味、酱料等

2. **商品更新流程**:
   - 修改商品基本信息
   - 调整价格和库存
   - 更新商品分类
   - 修改BOM关系
   - 调整口味和酱料

3. **库存管理流程**:
   - 商品入库时增加库存
   - 商品出库时减少库存
   - 库存预警和补货提醒
   - 库存盘点和调整

4. **商品删除流程**:
   - 检查是否有订单关联
   - 执行软删除操作
   - 处理相关BOM关系
   - 更新分类统计

### 业务场景
- 餐厅商品管理
- 商品分类管理
- 库存管理
- BOM管理
- 个性化定制

## 索引建议

### Product主表
- 主键索引: ID
- 唯一索引: Uuid
- 唯一索引: CompanyUuid + ProductCode（公司内编码唯一）
- 普通索引: CompanyUuid（公司查询）
- 普通索引: CategoryUuid（分类查询）
- 普通索引: ProductType（商品类型查询）
- 普通索引: Status（状态筛选）
- 普通索引: BarCode（条形码查询）

### ProductBom表
- 主键索引: ID
- 普通索引: ParentUuid（父商品查询）
- 普通索引: ChildUuid（子商品查询）

### ProductFlavor表
- 主键索引: ID
- 普通索引: ProductUuid（商品查询）
- 普通索引: FlavorCode（口味编码查询）

### ProductSauce表
- 主键索引: ID
- 普通索引: ProductUuid（商品查询）
- 普通索引: SauceCode（酱料编码查询）

## WebSocket推送

### AfterUpdate钩子
- **商品更新推送**: 当商品信息变更时，推送到相关终端
- **库存变更推送**: 当库存变化时，推送到前端

### 推送目标
- **收银端**: 商品信息更新、价格变更、库存预警
- **厨房端**: 商品制作信息、BOM结构
- **移动端**: 商品展示、个性化选项

## 业务方法

### IsActive() bool
- **功能**: 判断商品是否激活
- **返回**: 是否激活
- **用途**: 商品可用性验证

### GetPrice(flavors []string, sauces []string) float64
- **功能**: 计算定制商品价格
- **参数**: 口味列表、酱料列表
- **返回**: 最终价格
- **用途**: 价格计算

### GetBomItems() []ProductBom
- **功能**: 获取BOM项
- **返回**: BOM项列表
- **用途**: BOM展示

### GetAvailableFlavors() []ProductFlavor
- **功能**: 获取可用口味
- **返回**: 可用口味列表
- **用途**: 口味选择

### GetAvailableSauces() []ProductSauce
- **功能**: 获取可用酱料
- **返回**: 可用酱料列表
- **用途**: 酱料选择

### CheckStock(quantity int) bool
- **功能**: 检查库存是否充足
- **参数**: 需要数量
- **返回**: 是否充足
- **用途**: 库存验证

### UpdateStock(delta int) error
- **功能**: 更新库存
- **参数**: 库存变化量
- **返回**: 错误信息
- **用途**: 库存管理

## 注意事项

1. **多租户支持**: 通过CompanyUuid实现数据隔离
2. **价格精度**: 使用float64需要注意精度问题，建议改用decimal
3. **库存一致性**: 库存变化需要考虑BOM子商品
4. **软删除机制**: 使用DeleteTime实现软删除
5. **BOM循环**: 需要防止BOM循环引用
6. **实时同步**: 通过WebSocket确保商品信息实时更新
7. **权限控制**: 不同商品可能有不同的操作权限

## 业务规则

1. **编码唯一性**: 商品编码在公司内必须唯一
2. **价格验证**: 商品价格不能为负数
3. **库存限制**: 库存数量不能为负数
4. **BOM验证**: BOM子商品不能是自身
5. **状态管理**: 只有激活状态的商品才能销售
6. **分类关联**: 商品必须关联到有效分类
7. **条形码唯一**: 条形码在公司内建议唯一

## 扩展功能

### 商品规格
- 支持多规格商品管理
- 规格价格差异化
- 规格库存独立管理

### 商品属性
- 支持动态商品属性
- 属性值管理
- 属性搜索和过滤

### 商品图片
- 支持多图片管理
- 图片分类和排序
- 图片压缩和优化

### 商品评价
- 支持商品评价系统
- 评分统计和展示
- 评价内容管理

## 性能优化

### 查询优化
- 使用索引优化商品查询
- 批量加载关联数据
- 缓存热点商品数据

### 库存优化
- 库存预计算
- 库存变化队列
- 分布式库存管理

### BOM优化
- BOM结构缓存
- 递归BOM查询优化
- BOM变更批量处理

## 统计分析

### 商品统计
- **商品数量统计**: 各分类商品数量
- **价格分布分析**: 商品价格区间分布
- **库存周转分析**: 商品库存周转率

### 销售统计
- **销量排行**: 商品销量排名
- **销售额分析**: 商品销售额统计
- **利润分析**: 商品利润率计算

### 库存分析
- **库存预警**: 低库存商品预警
- **滞销分析**: 滞销商品识别
- **补货建议**: 基于销售数据的补货建议

## 集成接口

### ERP集成
- 商品信息同步
- 库存数据同步
- 价格数据同步

### 供应商集成
- 供应商商品信息
- 采购订单管理
- 供货状态跟踪

### 第三方平台集成
- 电商平台商品同步
- 价格信息同步
- 库存状态同步