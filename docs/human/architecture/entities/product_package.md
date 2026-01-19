# ProductPackage 实体模型说明

## 基本信息

- **实体名称**: ProductPackage
- **表名**: product_package
- **所属模块**: model
- **描述**: 商品包实体，用于管理组合商品和套餐商品

## 字段说明

| 字段名 | 类型 | 说明 | 备注 |
|--------|------|------|------|
| ID | uint | 主键ID | 继承自BaseModel |
| Uuid | uint64 | UUID | 继承自BaseModel |
| CreateTime | int64 | 创建时间 | 继承自BaseModel |
| UpdateTime | int64 | 更新时间 | 继承自BaseModel |
| DeleteTime | int64 | 删除时间 | 继承自BaseModel |
| CompanyUuid | uint64 | 公司UUID | 多租户标识 |
| PackageCode | string | 套餐编码 | 业务编码 |
| PackageName | string | 套餐名称 | |
| PackageType | string | 套餐类型 | combo套餐, bundle组合, gift礼品 |
| CategoryUuid | string | 分类UUID | 关联商品分类 |
| Description | string | 描述 | 套餐详细描述 |
| ImageUrl | string | 图片URL | 套餐图片 |
| Price | float64 | 套餐价格 | |
| OriginalPrice | float64 | 原价 | 用于显示折扣 |
| CostPrice | float64 | 成本价 | |
| Status | string | 状态 | active激活, inactive停用, discontinued停产 |
| IsRecommended | bool | 是否推荐 | 推荐套餐 |
| SortOrder | int | 排序顺序 | |
| Tags | string | 标签 | JSON格式存储 |
| ValidFrom | int64 | 有效期开始 | 时间戳 |
| ValidTo | int64 | 有效期结束 | 时间戳 |
| MaxDailyLimit | int | 每日限购 | 0表示不限 |
| MaxUserLimit | int | 每人限购 | 0表示不限 |
| SaleCount | int | 销售数量 | 累计销售统计 |
| SauceRequired | int | 是否必选小料 | 0-否 1-是（废弃字段，v2.12+） |
| SauceMinSelection | int | 小料最小选择数量 | v2.12新增，替代SauceRequired |
| SauceMaxSelection | int | 小料最大选择数量 | |
| Remark | string | 备注 | |

## 关联关系

### 关联实体
- **CompanyUuid** → Company 实体（通过 CompanyUuid 关联）
- **CategoryUuid** → ProductCategory 实体（通过 CategoryUuid 关联）

### 关联的子实体
- **ProductPackageItem** → 套餐项（一对多关系）
- **ProductPackageRule** → 套餐规则（一对多关系）

### 被关联关系
- **SaleOrderItem** → 通过商品包UUID关联
- **SaleBill** → 通过商品包UUID关联

## 数据流分析

### 数据来源
- 管理后台创建的套餐商品
- 商品管理系统配置的组合商品
- 营销活动创建的特价套餐

### 数据流向
1. **套餐创建流程**:
   - 商品管理员创建套餐
   - 设置套餐基本信息和价格
   - 添加套餐项商品
   - 配置套餐规则（如可选、必选等）
   - 设置有效期和限购规则

2. **套餐销售流程**:
   - 客户选择套餐商品
   - 系统验证套餐有效性
   - 检查限购规则
   - 创建订单时关联套餐信息
   - 更新销售统计

3. **套餐管理流程**:
   - 定期检查套餐有效期
   - 更新套餐状态
   - 调整套餐价格和规则
   - 统计套餐销售数据

### 业务场景
- 餐厅套餐管理（汉堡套餐、情侣套餐等）
- 商品组合销售（洗发水+护发素套装）
- 礼品套餐管理（节日礼品篮）
- 促销活动套餐（限时特价套餐）

## 索引建议

- 主键索引: ID
- 唯一索引: Uuid
- 唯一索引: PackageCode（套餐编码唯一）
- 普通索引: CompanyUuid（公司查询）
- 普通索引: CategoryUuid（分类查询）
- 普通索引: PackageType（套餐类型查询）
- 普通索引: Status（状态筛选）
- 普通索引: IsRecommended（推荐查询）
- 普通索引: ValidFrom/ValidTo（有效期查询）
- 普通索引: SortOrder（排序查询）

## WebSocket推送

### AfterUpdate钩子
- **套餐更新推送**: 当套餐信息变更时，推送到收银端
- **状态变更推送**: 当套餐状态变更时，推送到相关终端

### 推送目标
- **收银端**: 套餐信息更新、价格变更
- **移动端**: 推荐套餐展示、促销信息
- **管理端**: 套餐销售统计、库存提醒

## 业务方法

### IsActive() bool
- **功能**: 判断套餐是否当前有效
- **返回**: 是否有效
- **用途**: 套餐可用性验证

### CanPurchase(userUuid string, quantity int) bool
- **功能**: 判断是否可以购买
- **参数**: 用户UUID、购买数量
- **返回**: 是否可购买
- **用途**: 购买权限验证

### GetDailySales(date int64) int
- **功能**: 获取指定日期的销量
- **参数**: 日期时间戳
- **返回**: 销售数量
- **用途**: 销售统计分析

### GetUserSales(userUuid string) int
- **功能**: 获取指定用户的购买数量
- **参数**: 用户UUID
- **返回**: 购买数量
- **用途**: 限购验证

### GetItems() []ProductPackageItem
- **功能**: 获取套餐的所有商品项
- **返回**: 套餐项列表
- **用途**: 套餐详情展示

### GetRules() []ProductPackageRule
- **功能**: 获取套餐的所有规则
- **返回**: 套餐规则列表
- **用途**: 套餐验证逻辑

## 注意事项

1. **多租户支持**: 通过CompanyUuid实现数据隔离
2. **价格精度**: 使用float64需要注意精度问题，建议改用decimal
3. **有效期管理**: 需要定期检查和更新过期套餐
4. **限购规则**: 需要在订单创建时严格验证
5. **库存管理**: 套餐销售需要考虑包含商品的库存
6. **软删除机制**: 使用DeleteTime实现软删除
7. **实时同步**: 通过WebSocket确保套餐信息实时更新

## 业务规则

1. **套餐编码唯一性**: 套餐编码在公司内必须唯一
2. **价格验证**: 套餐价格应该合理，不能为负数
3. **有效期规则**: 有效期开始时间不能晚于结束时间
4. **限购验证**: 每日限购和每人限购需要严格执行
5. **状态管理**: 只有激活状态的套餐才能销售
6. **套餐项验证**: 套餐必须包含至少一个商品项
7. **库存检查**: 套餐销售时需要检查包含商品的库存

## 扩展功能

### 动态套餐
- 支持客户自定义套餐组合
- 动态价格计算
- 实时库存验证

### 套餐升级
- 支持套餐升级选项
- 差价计算
- 升级历史记录

### 套餐分享
- 支持套餐分享功能
- 分享链接生成
- 分享奖励机制

### 套餐预约
- 支持套餐预约购买
- 预约时间管理
- 预约库存预留

## 统计分析

### 销售统计
- **销量排行**: 套餐销量排名
- **销售额统计**: 套餐销售额分析
- **时间分布**: 套餐销售时间分布
- **客户偏好**: 客户套餐选择偏好

### 效益分析
- **利润分析**: 套餐利润率计算
- **成本分析**: 套餐成本结构分析
- **库存周转**: 套餐商品库存周转分析

### 促销效果
- **促销转化率**: 促销套餐的转化率
- **复购率**: 套餐复购率统计
- **客户留存**: 套餐对客户留存的影响