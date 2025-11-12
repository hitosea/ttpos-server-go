# Buying 服务功能说明

## 概述
Buying 服务提供采购管理相关功能，包括采购订单管理、供应商管理等核心业务能力。

## 服务接口

### IBuying - 采购订单管理

#### 采购订单创建
- **CreatePurchaseFromMq**: 根据材料请求从消息队列创建采购订单
- **CreatePurchaseOrder**: 创建采购订单
- **CreatePurchaseOrderFromSalesOrder**: 从销售订单创建采购订单

#### 采购订单查询
- **GetPurchaseOrder**: 获取单个采购订单详情
- **GetPurchaseOrderList**: 获取采购订单列表，支持查询条件过滤
- **GetPurchaseOrderCount**: 统计采购订单数量

#### 采购订单更新
- **UpdatePurchaseOrder**: 更新采购订单信息

#### 采购收货
- **CreatePurchaseReceiptFromOrder**: 从采购订单创建采购收货单

#### 内部订单流转
- **CreateInnerSaleOrderFromPurchaseOrder**: 从采购订单创建内部销售订单
- **CreateDeliveryNoteFromInnerSaleOrder**: 从内部销售订单创建送货单

### ISupplier - 供应商管理

#### 供应商基础操作
- **CreateSupplier**: 创建供应商
- **GetSupplier**: 获取供应商详情
- **UpdateSupplier**: 更新供应商信息
- **DeleteSupplier**: 删除供应商
- **ListSuppliers**: 获取供应商列表，支持分页和过滤
- **CountSupplier**: 统计供应商数量

#### 内部供应商管理
- **GetInnerSupplierList**: 获取内部供应商列表（根据公司缩写）

#### 供应商关系管理
- **AddSupplerTransactCompany**: 为供应商添加允许交易的公司
- **AddCompanyToSupplier**: 将公司添加到供应商的允许交易公司列表

#### 供应商商品管理
- **GetSupplierItemList**: 获取供应商商品列表

## 业务流程

### 采购流程
1. 接收材料请求（MQ）→ 创建采购订单
2. 采购订单审核 → 创建采购收货单
3. 采购收货 → 更新库存

### 内部调拨流程
1. 采购订单 → 创建内部销售订单
2. 内部销售订单 → 创建送货单
3. 送货单提交 → 完成调拨

## 使用说明

### 服务注册
```go
service.RegisterBuying(buyingImpl)
service.RegisterSupplier(supplierImpl)
```

### 服务调用
```go
// 获取采购服务实例
buying := service.Buying()

// 获取供应商服务实例
supplier := service.Supplier()
```

## 注意事项
1. 所有接口都需要传入 context 上下文对象
2. 供应商与公司的关联关系需要通过 AddSupplerTransactCompany 维护
3. 内部订单流转涉及多个服务协同，需要确保事务一致性
