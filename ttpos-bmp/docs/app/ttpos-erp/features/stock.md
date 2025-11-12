# Stock 服务功能说明

## 概述
Stock 服务提供库存管理功能，包括物品管理、仓库管理、库存调拨、库存盘点等核心业务能力。

## 服务接口

### IDeliveryNote - 送货单管理
- **CreateDeliveryNote**: 创建送货单
- **GetDeliveryNote**: 获取送货单详情
- **GetDeliveryNoteList**: 获取送货单列表
- **UpdateDeliveryNote**: 更新送货单
- **SubmitDeliveryNote**: 提交送货单
- **CancelDeliveryNote**: 取消送货单
- **DeleteDeliveryNote**: 删除送货单
- **CreateDeliveryNoteFromSaleOrder**: 从销售订单创建送货单

### IItem - 物品管理
- **GetItemList**: 获取物品列表
- **GetItem**: 获取物品详情
- **SaveItem**: 保存物品（创建或更新）
- **DeleteItem**: 删除物品（包括变体商品）
- **GetItemStock**: 获取物品库存信息
- **SavePosAttribute**: 保存 POS 属性物品
- **SavePosAddon**: 保存 POS 加料物品
- **CreateSingleVariantItem**: 创建多规格商品的单个规格
- **SyncDelay**: 同步延迟处理（推送到队列）

### IItemGroup - 物品分组管理
- **GetItemGroupList**: 获取物品分组列表
- **GetItemGroup**: 获取物品分组详情
- **SaveItemGroup**: 保存物品分组
- **DeleteItemGroup**: 删除物品分组
- **SaveAttributeGroup**: 保存物品属性分组
- **DeleteAttributeGroup**: 删除属性分组
- **SaveAddonGroup**: 保存加料组

### IMaterialTransfer - 物料调拨
- **MaterialTransfer**: 物料调拨
- **CreateInnerTransferReceipt**: 创建内部调拨收货单（通过内部销售单→采购单）

### IProduct - 产品管理
- **UpdateProduct**: 更新产品信息

### IStock - 库存管理
- **GetAttributeList**: 获取属性列表
- **GetAttributeValuesList**: 获取属性值列表
- **GetItemAttribute**: 获取属性详情
- **SaveAttribute**: 保存属性
- **CreateMaterialRequest**: 创建物料请求
- **GetMaterialRequestList**: 获取物料请求列表
- **GetStockLedger**: 获取库存分类账
- **SaveStockReconciliation**: 保存库存盘点
- **GetStockReconciliationList**: 获取库存盘点列表
- **SubmitStockReconciliation**: 提交库存盘点
- **CancelStockReconciliation**: 取消库存盘点

### IUom - 单位管理
- **GetUomList**: 获取单位列表
- **GetUom**: 获取单位详情
- **SaveUom**: 保存单位
- **DeleteUom**: 删除单位

### IWarehouse - 仓库管理
- **CreateWarehouse**: 创建仓库
- **GetWarehouseList**: 获取仓库列表
- **GetWarehouse**: 获取仓库详情
- **GetDefaultWarehouse**: 获取默认仓库
- **UpdateWarehouse**: 更新仓库
- **DeleteWarehouse**: 删除仓库

## 使用说明

### 服务注册
```go
service.RegisterItem(itemImpl)
service.RegisterItemGroup(itemGroupImpl)
service.RegisterStock(stockImpl)
service.RegisterWarehouse(warehouseImpl)
service.RegisterDeliveryNote(deliveryNoteImpl)
service.RegisterMaterialTransfer(materialTransferImpl)
service.RegisterUom(uomImpl)
service.RegisterProduct(productImpl)
```
