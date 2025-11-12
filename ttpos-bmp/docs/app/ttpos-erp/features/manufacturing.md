# Manufacturing 服务功能说明

## 概述
Manufacturing 服务提供生产制造管理功能，主要包括物料清单（BOM）管理。

## 服务接口

### IBom - 物料清单管理

#### BOM 查询
- **GetBomList**: 获取 BOM 列表
  - 支持查询条件过滤
  - 返回 BOM 基本信息列表

- **GetBom**: 根据 BOM 名称获取单个 BOM 详细信息
  - 包含完整的物料清单结构
  - 包含子项明细

#### BOM 保存
- **SaveBom**: 保存 BOM 信息
  - 支持创建新 BOM
  - 支持更新现有 BOM
  - 包含 BOM 子项的保存

## 业务场景

### 物料清单管理
- 产品配方管理
- 生产用料清单
- 成本核算基础
- 生产计划依据

### BOM 结构
- 支持多层级 BOM
- 物料用量管理
- 替代物料配置
- 损耗率设置

## 使用说明

### 服务注册
```go
service.RegisterBom(bomImpl)
```

### 服务调用
```go
// 获取 BOM 服务实例
bom := service.Bom()

// 获取 BOM 列表
list, err := bom.GetBomList(ctx, &manufacturing.GetBomListReq{
    // 查询条件
})

// 获取 BOM 详情
bomInfo, err := bom.GetBom(ctx, &manufacturing.GetBomReq{
    Name: "BOM-001",
})

// 保存 BOM
result, err := bom.SaveBom(ctx, &manufacturing.SaveBomReq{
    // BOM 信息
})
```

## 数据结构

### BOM 主表
- BOM 编号
- 产品编码
- 产品数量
- 单位
- 是否默认 BOM

### BOM 子项
- 物料编码
- 物料数量
- 单位
- 损耗率
- 替代物料

## 业务流程

### BOM 创建流程
1. 选择产品
2. 添加物料子项
3. 设置物料用量
4. 配置损耗率
5. 保存 BOM

### BOM 使用场景
- 生产订单创建时引用 BOM
- 物料需求计划（MRP）计算
- 生产成本核算
- 库存预留

## 注意事项
1. BOM 名称需要保证唯一性
2. 一个产品可以有多个 BOM，但只能有一个默认 BOM
3. BOM 子项的物料必须是有效的库存物料
4. 修改已使用的 BOM 需要谨慎，可能影响在制订单
5. BOM 支持多层级嵌套，注意避免循环引用
