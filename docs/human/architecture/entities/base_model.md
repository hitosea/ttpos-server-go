# BaseModel 实体模型说明

## 基本信息

- **实体名称**: BaseModel
- **表名**: 无（基础模型，不直接映射到表）
- **所属模块**: model
- **描述**: 基础模型，为所有实体提供通用字段和方法

## 字段说明

| 字段名 | 类型 | 说明 | 备注 |
|--------|------|------|------|
| ID | uint | 主键ID | 自增主键 |
| Uuid | uint64 | UUID | 绑定记录ID，业务唯一标识 |
| CreateTime | int64 | 创建时间 | 时间戳，自动生成 |
| UpdateTime | int64 | 更新时间 | 时间戳，自动更新 |
| DeleteTime | int64 | 删除时间 | 时间戳，软删除标记，0表示未删除 |
| isUpdate | bool | 更新标记 | 虚拟字段，用于判断该model是否需要更新 |
| operateSource | string | 操作来源 | 虚拟字段，用于标记添加来源 |

## 关联关系

### 关联实体
BaseModel 作为基础模型，被所有其他实体继承，包括：
- SaleBill（销售账单）
- SaleOrder（销售订单）
- ProductPackage（商品包）
- Desk（桌台）
- Company（公司）
- CustomerCall（客户呼叫）
- PrinterLog（打印日志）
- H5Order（H5订单）
- MemberSaleOrder（会员销售订单）
- 等等...

## 数据流分析

### 数据来源
- 所有实体的基础字段都继承自BaseModel
- 通过GORM的钩子函数自动管理生命周期

### 数据流向
1. **创建流程**:
   - 创建实体时，BeforeCreate钩子自动生成UUID
   - 使用雪花算法生成唯一ID
   - 设置创建时间戳

2. **更新流程**:
   - 更新实体时，BeforeUpdate钩子自动更新时间戳
   - 通过SetUpdate()方法标记需要更新

3. **删除流程**:
   - 软删除时，SetDelete()方法设置删除时间戳
   - IsDelete()方法检查是否已删除

4. **WebSocket推送流程**:
   - 通过getCompanyUuid()从数据库名提取公司UUID
   - 各种AfterCreate/AfterUpdate钩子推送实时数据
   - 支持多租户架构

### 业务场景
- 多租户数据隔离
- 软删除机制
- 实时数据同步
- 操作来源追踪
- 统一的生命周期管理

## 索引建议

由于BaseModel是基础模型，索引建议在具体实体中定义：
- 主键索引: ID（自增）
- 唯一索引: Uuid（业务唯一标识）
- 普通索引: CreateTime（创建时间查询）
- 普通索引: UpdateTime（更新时间查询）
- 普通索引: DeleteTime（软删除查询）

## 注意事项

1. **多租户支持**: 通过数据库名称提取公司UUID实现租户隔离
2. **软删除机制**: 使用DeleteTime字段实现软删除，避免数据丢失
3. **WebSocket集成**: 多个钩子函数支持实时数据推送
4. **UUID生成**: 使用雪花算法保证分布式环境下的唯一性
5. **操作来源追踪**: operateSource字段支持数据来源追踪
6. **继承模式**: 所有业务实体都继承BaseModel，确保数据结构一致性

## Hook方法说明

### BeforeCreate
- 自动生成UUID
- 处理UUID为0的特殊情况

### AfterUpdate (SaleBill)
- 推送订单更新到WebSocket客户端
- 推送桌台状态更新

### AfterCreate (CustomerCall)
- 推送客户呼叫信息到收银员、助手端、厨房端

### AfterCreate (PrinterLog)
- 推送打印数据到指定设备

### AfterCreate/AfterUpdate (H5Order)
- 推送H5订单信息到收银端

### AfterUpdate (Desk)
- 推送桌台状态更新

### AfterUpdate (MemberSaleOrder)
- 推送会员订单状态更新和呼叫消息

### AfterUpdate (ProductPackage)
- 推送商品包更新信息

### AfterUpdate (ProductCategory)
- 推送商品分类更新信息