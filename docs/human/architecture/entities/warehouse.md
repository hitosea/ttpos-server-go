# Warehouse 实体模型说明

## 基本信息

- **实体名称**: Warehouse
- **表名**: ttpos_warehouse
- **所属模块**: model
- **描述**: 仓库管理实体，用于管理门店仓库信息，支持普通仓库和在途仓库，包含库存管理、出入库记录等完整功能

## 字段说明

| 字段名 | 类型 | 说明 | 备注 |
|--------|------|------|------|
| ID | uint | 主键ID | 继承自BaseModel |
| Uuid | uint64 | UUID | 继承自BaseModel |
| CreateTime | int64 | 创建时间 | 继承自BaseModel |
| UpdateTime | int64 | 更新时间 | 继承自BaseModel |
| DeleteTime | int64 | 删除时间 | 继承自BaseModel |
| Name | string | 名称 | 仓库名称 |
| MultiLanguageNameUuid | uint64 | 多语言名称UUID | 关联多语言名称表 |
| Type | string | 仓库类型 | normal-普通；transit-在途 |
| Code | string | 仓库编码 | 仓库唯一编码 |
| Status | int | 仓库状态 | 0-禁用 1-启用 |
| Contact | string | 联系人 | 仓库联系人 |
| Phone | string | 联系电话 | 仓库联系电话 |
| Address | string | 地址 | 仓库地址 |
| IsDefault | int | 是否默认仓库 | 0-否；1-是 |
| ErpCode | string | 关联erpnext | ERP系统编码 |
| HeadquarterUuid | uint64 | 总部Uuid | 总部仓库标识 |

## 关联关系

### 关联实体
- **MultiLanguageNameUuid** → MultiLanguageName 实体（通过 MultiLanguageNameUuid 关联）
- **HeadquarterUuid** → Warehouse 实体（通过 HeadquarterUuid 关联）

### 关联的子实体
- **WarehouseItem** → 仓库商品库存（一对多关系）
- **WarehouseForm** → 入库单（一对多关系）
- **WarehouseOutForm** → 出库单（一对多关系）
- **WarehouseInOutLog** → 出入库记录（一对多关系）

### 被关联关系
- **WarehouseItem** → 通过 WarehouseUuid 关联
- **WarehouseInOutLog** → 通过 WarehouseUuid 关联
- **WarehouseOutFormItem** → 通过 WarehouseUuid 关联

## 数据流分析

### 数据来源
- 管理员创建仓库信息
- ERP系统同步仓库数据
- 门店初始化默认仓库
- 第三方系统导入仓库信息

### 数据流向
1. **仓库创建流程**:
   - 管理员录入仓库基本信息
   - 设置仓库类型和状态
   - 配置联系人和地址信息
   - 创建仓库记录

2. **库存管理流程**:
   - 商品入库时更新库存数量
   - 销售出库时扣减库存
   - 库存调整时修改库存记录
   - 定期进行库存盘点

3. **出入库流程**:
   - 采购商品创建入库单
   - 销售商品创建出库单
   - 退菜入库处理
   - 损耗出库处理

4. **库存同步流程**:
   - 实时更新库存数量
   - 记录库存变更日志
   - 同步到ERP系统
   - 生成库存报表

### 业务场景
- 仓库基础信息管理
- 商品库存实时管理
- 采购入库处理
- 销售出库处理
- 库存调整和盘点
- 退菜入库处理
- 损耗出库处理
- 在途库存管理
- 多仓库库存调拨
- ERP系统集成

## 索引建议

- 主键索引: ID
- 唯一索引: Uuid
- 唯一索引: Code（仓库编码唯一）
- 普通索引: Type（仓库类型查询）
- 普通索引: Status（状态筛选）
- 普通索引: IsDefault（默认仓库查询）
- 普通索引: ErpCode（ERP编码查询）

## 业务方法

### IsTransit() bool
- **功能**: 判断是否为在途仓库
- **返回**: 是否为在途仓库
- **逻辑**: Type == "transit"
- **用途**: 仓库类型判断

### IsHeadquarter() bool
- **功能**: 判断是否为总部仓库
- **返回**: 是否为总部仓库
- **逻辑**: HeadquarterUuid > 0
- **用途**: 仓库层级判断

### IsDisabled() bool
- **功能**: 判断仓库是否禁用
- **返回**: 是否禁用
- **逻辑**: Status == 0
- **用途**: 仓库状态检查

## 扩展实体

### WarehouseItem 仓库商品库存表
#### 基本信息
- **表名**: ttpos_warehouse_item
- **描述**: 记录各仓库中商品的库存信息

#### 字段说明
| 字段名 | 类型 | 说明 | 备注 |
|--------|------|------|------|
| ID | uint | 主键ID | 继承自BaseModel |
| WarehouseUuid | uint64 | 仓库UUID | 索引:idx_warehouse_uuid |
| MaterialUuid | uint64 | 商品UUID | 索引:idx_material_uuid |
| MaterialCode | string | 商品编码 | 索引:idx_material_code |
| Stock | float64 | 库存数量 | |
| ReservedStock | float64 | 预留库存数量 | |
| Valuation | float64 | 估值单价 | |

#### 关联关系
- **WarehouseUuid** → Warehouse 实体（通过 WarehouseUuid 关联）
- **MaterialUuid** → Material 实体（通过 MaterialUuid 关联）

### WarehouseForm 入库单表
#### 基本信息
- **表名**: ttpos_warehouse_form
- **描述**: 记录商品入库的详细信息

#### 字段说明
| 字段名 | 类型 | 说明 | 备注 |
|--------|------|------|------|
| ID | uint | 主键ID | 继承自BaseModel |
| FormNo | string | 编号 | |
| Scene | int | 交易类型 | 0-purchase采购入库 1-add添加入库 2-adjust调整入库 3-退菜入库 |
| Num | int | 数量 | |
| Remark | string | 备注 | |
| Status | int | 状态 | 0-success已入库 1-canceled已撤销 |
| ProductBomUuid | uint64 | 商品BOM表uuid | |
| MaterialUuid | uint64 | 材料uuid | |
| PurchaseOrderUuid | uint64 | 采购订单uuid | |
| OperatorUuid | uint64 | 操作员uuid | |
| RevokeTime | int | 撤销时间 | 时间戳 |

#### 关联关系
- **ProductBomUuid** → ProductBom 实体（通过 ProductBomUuid 关联）
- **MaterialUuid** → Material 实体（通过 MaterialUuid 关联）
- **PurchaseOrderUuid** → PurchaseOrder 实体（通过 PurchaseOrderUuid 关联）
- **OperatorUuid** → Staff 实体（通过 OperatorUuid 关联）

#### 业务方法

##### SetNil()
- **功能**: 清空关联对象
- **用途**: 避免循环引用

### WarehouseFormItem 入库单明细表
#### 基本信息
- **表名**: ttpos_warehouse_form_item
- **描述**: 入库单的详细商品明细

#### 字段说明
| 字段名 | 类型 | 说明 | 备注 |
|--------|------|------|------|
| ID | uint | 主键ID | 继承自BaseModel |
| Num | float64 | 入库数量 | |
| Scene | int | 场景 | 0-采购 1-添加入库 2-调整入库 3-退菜入库 |
| AddStock | int | 是否已经加库存 | 0-未加库存 1-已加库存 |
| MaterialUuid | uint64 | 材料uuid | |
| ProductBomUuid | uint64 | 商品BOM表uuid | |
| WarehouseFormUuid | uint64 | 入库单uuid | |
| SaleOrderProductUuid | uint64 | 销售订单商品uuid | 用于退菜入库 |
| SaleBillUuid | uint64 | 销售账单uuid | 用于退菜入库 |

#### 业务方法

##### IsMaterial() bool
- **功能**: 判断是否是原材料
- **返回**: 是否是原材料
- **逻辑**: MaterialUuid != 0

##### IsProductBom() bool
- **功能**: 判断是否是规格商品或小料
- **返回**: 是否是规格商品或小料
- **逻辑**: ProductBomUuid != 0

### WarehouseOutForm 出库单表
#### 基本信息
- **表名**: ttpos_warehouse_out_form
- **描述**: 记录商品出库的详细信息

#### 字段说明
| 字段名 | 类型 | 说明 | 备注 |
|--------|------|------|------|
| ID | uint | 主键ID | 继承自BaseModel |
| FormNo | string | 编号 | |
| Scene | int | 出库类型 | 0-sales销售出库 1-adjust调整出库 2-loss损耗出库 3-lost丢失出库 4-delete删除出库 |
| Remark | string | 备注 | |
| Status | int | 状态 | 0-success已出库 1-canceled已撤销 |
| RevokeTime | int64 | 撤销时间 | 时间戳 |
| OperatorUuid | uint64 | 操作员uuid | |
| AssociatedOrderUuid | uint64 | 关联订单uuid | sale_bill_uuid |

#### 业务方法

##### RevokeForm()
- **功能**: 撤销出库
- **用途**: 反结账时撤销出库记录，将库存退还

### WarehouseOutFormItem 出库单明细表
#### 基本信息
- **表名**: ttpos_warehouse_out_form_item
- **描述**: 出库单的详细商品明细

#### 字段说明
| 字段名 | 类型 | 说明 | 备注 |
|--------|------|------|------|
| ID | uint | 主键ID | 继承自BaseModel |
| Num | float64 | 数量 | |
| Scene | int | 场景 | 0-销售出库 1-adjust调整 2-loss损耗 3-lost丢失 4-delete删除 |
| Status | int | 状态 | 0-预出库 1-已出库 |
| ReduceStock | int | 是否已经减库存 | 0-未减库存 1-已减库存 |
| RevokeTime | int64 | 撤销时间 | 时间戳 |
| WarehouseOutFormUuid | uint64 | 出库单uuid | |
| WarehouseUuid | uint64 | 仓库uuid | 出库的仓库 |
| ProductBomUuid | uint64 | 商品BOM表uuid | 规格商品或小料 |
| PackageUuid | uint64 | 套餐uuid | 只有套餐子商品才有 |
| MaterialUuid | uint64 | 材料uuid | 原材料 |
| SaleOrderProductUuid | uint64 | 销售订单商品uuid | |
| SaleOrderUuid | uint64 | 销售订单uuid | |
| SaleBillUuid | uint64 | 销售账单uuid | |
| StaffShiftLogUuid | uint64 | 员工交班记录ID | |

#### 业务方法

##### IsMaterial() bool
- **功能**: 判断是否是原材料
- **返回**: 是否是原材料
- **逻辑**: MaterialUuid != 0

##### IsProductBom() bool
- **功能**: 判断是否是规格商品或小料
- **返回**: 是否是规格商品或小料
- **逻辑**: ProductBomUuid != 0

### WarehouseInOutLog 仓库出入库记录表
#### 基本信息
- **表名**: ttpos_warehouse_in_out_log
- **描述**: 记录所有仓库出入库操作的详细日志

#### 字段说明
| 字段名 | 类型 | 说明 | 备注 |
|--------|------|------|------|
| ID | uint | 主键ID | 继承自BaseModel |
| LogType | int | 日志类型 | 0-入库 1-出库 |
| Scene | int | 场景 | 0-采购入库 1-销售出库 2-发货出库 3-盘盈入库 4-盘亏出库 20-在途入库 21-在途出库 |
| WarehouseUuid | uint64 | 仓库ID | |
| MaterialUuid | uint64 | 物品ID | |
| MaterialName | string | 物品名称 | JSON格式，记录当时物品名称 |
| MaterialBaseUnitUuid | uint64 | 物品基准单位ID | |
| MaterialBaseUnitName | string | 物品基准单位名称 | |
| Num | float64 | 数量 | |
| Price | float64 | 单价 | 物品基准单位单价 |
| Amount | float64 | 金额 | 单价*数量 |
| SupplierUuid | uint64 | 供应商ID | |
| SupplierErpCode | string | 供应商ERP编码 | |
| SupplierName | string | 供应商名称 | |
| OrderNo | string | 单据编号 | |
| OtherOrgUuid | uint64 | 对方机构ID | |
| OtherOrgType | uint64 | 对方机构类型 | 0:供应商 1:客户 |
| OtherOrgName | string | 对方机构名称 | |
| OpeningHours | string | 营业时段 | 仅用于Scene销售出库的场景 |

#### 关联关系
- **MaterialUuid** → Material 实体（通过 MaterialUuid 关联）
- **SupplierUuid** → Supplier 实体（通过 SupplierUuid 关联）
- **WarehouseUuid** → Warehouse 实体（通过 WarehouseUuid 关联）

#### 业务方法

##### TableName() string
- **功能**: 获取表名
- **返回**: 表名
- **用途**: 自定义表名映射

##### SetNil()
- **功能**: 清空关联对象
- **用途**: 避免循环引用

## 业务规则

1. **仓库管理规则**:
   - 仓库编码在系统内必须唯一
   - 默认仓库只能有一个
   - 禁用的仓库不能进行出入库操作
   - 在途仓库用于中转库存

2. **库存管理规则**:
   - 库存数量不能为负数
   - 预留库存需要单独管理
   - 库存变更需要记录详细日志
   - 支持多仓库库存调拨

3. **入库管理规则**:
   - 入库单必须有明确的入库场景
   - 采购入库需要关联采购订单
   - 退菜入库需要关联销售订单
   - 入库后需要实时更新库存

4. **出库管理规则**:
   - 出库单必须有明确的出库场景
   - 销售出库需要关联销售订单
   - 损耗出库需要记录损耗原因
   - 出库后需要实时扣减库存

5. **日志记录规则**:
   - 所有出入库操作必须记录日志
   - 日志包含完整的快照信息
   - 支持按时间、仓库、商品查询
   - 日志数据不可修改

## 注意事项

1. **多仓库支持**: 系统支持多个仓库同时管理
2. **库存实时性**: 库存数据需要实时更新，确保准确性
3. **数据一致性**: 出入库操作需要保证数据一致性
4. **日志完整性**: 所有库存变更都需要记录详细日志
5. **预留库存**: 预留库存需要单独管理，避免重复出库
6. **ERP集成**: 支持与ERP系统的数据同步
7. **多语言支持**: 仓库名称支持多语言
8. **软删除**: 使用DeleteTime实现软删除机制

## 扩展功能

### 智能库存管理
- 基于销售数据的库存预测
- 自动补货提醒功能
- 库存预警和报警
- 库存周转率分析

### 多仓库调拨
- 仓库间库存调拨管理
- 调拨单生成和审批
- 在途库存跟踪
- 调拨成本计算

### 库存盘点
- 定期库存盘点计划
- 盘盈盘亏处理
- 盘点差异分析
- 盘点报表生成

### 成本核算
- 移动平均成本计算
- 批次成本管理
- 库存估值分析
- 成本变动跟踪

## 性能优化

### 查询优化
- 库存查询使用复合索引
- 出入库日志按时间分区
- 仓库列表查询缓存
- 商品库存预加载

### 数据归档
- 历史出入库日志定期归档
- 过期的入库单出库单清理
- 库存变更历史压缩存储

## 统计分析

### 库存分析
- **库存周转率**: 各商品库存周转效率分析
- **库存结构分析**: 库存金额和数量分布统计
- **库存预警分析**: 低库存和高库存预警统计
- **呆滞库存分析**: 长期不动库存商品统计

### 出入库分析
- **出入库趋势**: 按时间统计出入库趋势
- **出入库类型分析**: 各类型出入库占比统计
- **仓库效率分析**: 各仓库出入库效率对比
- **供应商分析**: 供应商供货质量和效率统计

### 成本分析
- **库存成本分析**: 库存总成本和单位成本分析
- **成本变动分析**: 成本变动趋势和原因分析
- **损耗成本分析**: 各类损耗成本统计
- **调拨成本分析**: 仓库间调拨成本统计