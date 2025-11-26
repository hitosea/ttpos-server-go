# Item 实体模型说明

## 基本信息

- **实体名称**: Item
- **表名**: item
- **所属模块**: ttpos-erp
- **描述**: 商品实体，用于存储本地商品信息（与 ERPNext Item 不同，这是本地缓存表）

## 字段说明

| 字段名 | 类型 | 说明 | 备注 |
|--------|------|------|------|
| ID | uint | 主键ID | 自增 |
| Uuid | uint64 | UUID | 商品唯一标识 |
| Name | string | 商品名称 | |
| ImageName | string | 图片名称 | |
| Status | uint | 状态 | 0-上架，1-下架 |
| CategoryUuid | uint64 | 分类UUID | 关联商品分类 |
| CreateTime | int64 | 创建时间 | Unix 时间戳 |
| UpdateTime | int64 | 更新时间 | Unix 时间戳 |
| DeleteTime | int64 | 删除时间 | 软删除时间戳 |

## 关联关系

### 关联实体
- **CategoryUuid** → Category（商品分类）

### 被引用
- 无直接引用（本地缓存表）

## 数据流分析

### 数据来源
- 本地商品管理
- 与 ERPNext Item 同步（可选）

### 数据流向
1. **商品创建流程**:
   - 在本地创建商品记录
   - 可选：同步到 ERPNext

2. **商品更新流程**:
   - 更新本地商品信息
   - 可选：同步到 ERPNext

### 业务场景
- 本地商品管理
- 商品缓存
- 商品状态管理

## 索引建议

- 主键索引: ID
- 唯一索引: Uuid
- 普通索引: CategoryUuid（分类查询）
- 普通索引: Status（状态查询）
- 普通索引: DeleteTime（软删除过滤）

## 注意事项

1. 这是本地商品表，与 ERPNext 的 Item DocType 不同
2. 使用软删除机制（DeleteTime）
3. Status 字段：0-上架，1-下架
4. 可选与 ERPNext Item 同步
