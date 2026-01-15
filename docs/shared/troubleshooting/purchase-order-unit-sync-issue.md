# 品牌采购审批时子店单位同步问题

## 问题描述

品牌采购-总店在通过审批时添加了子店物品没有的采购单位，通过审批后子店未显示该单位及数量。

## 问题分析

### 问题场景

1. **总店审批品牌采购单**：总店使用更新接口添加新的采购单位
2. **同步到子店**：系统自动将采购明细同步到子店数据库
3. **子店查询**：子店查询采购单时，发现缺少总店新增的采购单位和数量

### 根本原因

在 `syncItemsToCompanyShop` 方法中同步采购明细到子店时，存在以下问题：

1. **子店已有物料但缺少单位**：
   - 当子店物料表中已存在该物料，但该物料的 `NotBaseUnitList` 中没有总店新增的采购单位
   - `buildPurchaseOrderItem` 方法在构建明细时，会遍历物料的 `NotBaseUnitList` 查找匹配的单位
   - 找不到匹配的单位时，该单位的 `PurchaseOrderItemUnit` 记录不会被创建

2. **原有逻辑的局限性**：
   - 原有逻辑只处理了"子店完全没有这个物料"的情况（第 1334-1355 行）
   - 没有处理"子店有物料但缺少单位"的情况

### 代码位置

文件：`main/app/service/purchase_order/purchase_order.go`

方法：`syncItemsToCompanyShop`（第 1223-1378 行）

## 解决方案

### 修复逻辑

在同步采购明细到子店时，增加单位补充逻辑：

1. **获取子店和总店的单位信息**：
   - 从子店数据库构建明细（可能缺少单位）
   - 从总店数据库构建完整的明细（包含所有单位）

2. **对比并补充缺失单位**：
   - 构建总店单位映射：`MaterialUuid -> Units`
   - 构建子店单位映射：`ItemUuid -> map[UnitUuid]bool`
   - 遍历子店明细，找出每个物料缺失的单位
   - 从总店单位信息中复制缺失的单位到子店明细

3. **保证数据完整性**：
   - 复制总店单位信息时，使用子店的 `ItemUuid` 和 `PurchaseOrderUuid`
   - 保留总店的单位属性（单位名称、换算率、数量等）

### 修复代码

```go
// 5. 检查子店铺物料是否缺少采购单位，补充缺失的单位
// 从总部数据库查询完整的单位信息
hqItemsWithUnits, _, err := s.validator.buildPurchaseOrderItems(ctx.GetDB(), companyOrder.Uuid, itemReqsToCreate)
if err != nil {
    return errors.WithMessage(errors.New("从总部查询完整单位信息失败"), err.Error())
}

// 构建总部单位映射：MaterialUuid -> Units
hqUnitsMap := make(map[uint64][]model.PurchaseOrderItemUnit)
for _, item := range hqItemsWithUnits {
    hqUnitsMap[item.MaterialUuid] = item.Units
}

// 构建子店已有单位映射：ItemUuid -> map[UnitUuid]bool
companyUnitsMap := make(map[uint64]map[uint64]bool)
for _, unit := range unitsFromCompany {
    if _, exists := companyUnitsMap[unit.ItemUuid]; !exists {
        companyUnitsMap[unit.ItemUuid] = make(map[uint64]bool)
    }
    companyUnitsMap[unit.ItemUuid][unit.UnitUuid] = true
}

// 为子店铺的明细补充缺失的单位
for i := range itemsFromCompany {
    item := &itemsFromCompany[i]
    hqUnits, hasHqUnits := hqUnitsMap[item.MaterialUuid]
    if !hasHqUnits {
        continue
    }

    companyUnits, hasCompanyUnits := companyUnitsMap[item.Uuid]

    // 找出子店铺缺失的单位
    for _, hqUnit := range hqUnits {
        // 如果子店铺没有这个单位，则从总部复制
        if !hasCompanyUnits || !companyUnits[hqUnit.UnitUuid] {
            // 复制总部单位信息，但使用子店铺的ItemUuid
            newUnit := model.PurchaseOrderItemUnit{
                ItemUuid:           item.Uuid,
                PurchaseOrderUuid:  companyOrder.Uuid,
                UnitUuid:           hqUnit.UnitUuid,
                Num:                hqUnit.Num,
                UnitName:           hqUnit.UnitName,
                UnitConversionRate: hqUnit.UnitConversionRate,
                ErpnextUom:         hqUnit.ErpnextUom,
                BaseUnitUuid:       hqUnit.BaseUnitUuid,
                BaseUnitName:       hqUnit.BaseUnitName,
            }
            item.Units = append(item.Units, newUnit)
        }
    }
}
```

## 影响范围

### 直接影响

- 品牌采购更新接口
- 子店采购单查询接口
- 子店收货接口

### 业务场景

- 总店在审批品牌采购时添加新的采购单位
- 子店查询采购单明细
- 子店创建收货单

## 测试建议

### 测试场景

1. **子店缺少采购单位**：
   - 前置条件：子店物料表中有物料A，但缺少单位"箱"
   - 操作：总店审批采购单，添加物料A的"箱"单位
   - 预期：子店采购单中显示物料A的"箱"单位及数量

2. **子店完全没有物料**：
   - 前置条件：子店物料表中没有物料B
   - 操作：总店审批采购单，添加物料B及其单位
   - 预期：子店采购单中显示物料B及所有单位

3. **子店已有所有单位**：
   - 前置条件：子店物料表中有物料C及其所有单位
   - 操作：总店审批采购单，添加物料C的已有单位
   - 预期：子店采购单中正常显示物料C及数量

### 验证步骤

1. 创建品牌采购单（总店）
2. 总店审批并添加新单位
3. 查询子店采购单明细
4. 验证单位和数量是否正确
5. 子店创建收货单
6. 验证收货是否正常

## 相关文档

- [品牌采购开发规范](../specs/active/story-purchase-order/)
- [数据库开发规范](/.cursor/rules/database.mdc)
- [Go Main 开发规范](/.cursor/rules/go-main.mdc)

## 修复时间

- 问题发现：2026-01-15
- 修复完成：2026-01-15
- 修复人员：AI Agent

## 备注

- 该问题影响品牌采购的核心功能，建议尽快测试验证
- 修复后需要关注子店收货时的单位验证逻辑
- 建议增加单元测试覆盖该场景
