# Selling 服务功能说明

## 概述
Selling 服务提供销售管理功能，包括销售订单、POS 发票、客户管理等核心业务能力。

## 服务接口

### IAsyncSelling - 异步销售操作
- **SavePosInvoice**: 异步保存 POS 发票
- **ReturnPosInvoice**: 异步退货 POS 发票
- **CancelPosInvoice**: 异步取消 POS 发票
- **ClosePosEntry**: 异步关帐操作
- **GetLatestReceivePosInvoice**: 获取最新接收的 POS 发票

### ISelling - 销售管理

#### 销售订单管理
- **CreateSalesOrder**: 创建销售订单
- **UpdateSalesOrder**: 更新销售订单
- **GetSalesOrder**: 获取销售订单详情
- **GetSalesOrderList**: 获取销售订单列表
- **CountSalesOrder**: 统计销售订单数量
- **CancelSalesOrder**: 取消销售订单
- **SubmitSalesOrder**: 提交销售订单

#### POS 配置管理
- **GetPosProfileList**: 查询 POS 配置文件列表
- **CreatePosProfile**: 创建 POS 配置文件
- **CreateModePaymentAccount**: 创建支付方式账户

#### POS 开关帐
- **OpenPosEntry**: 开帐操作
- **ClosePosEntry**: 关帐操作
- **IsProfileOpening**: 查询 POS 配置文件是否开帐
- **GetPosOpeningEntry**: 获取 POS 开帐记录

#### POS 发票管理
- **GetPosInvoiceList**: 获取 POS 发票列表
- **SavePosInvoice**: 保存 POS 发票
- **SavePosInvoiceStep**: 保存 POS 发票步骤
- **ReturnPosInvoice**: 退货 POS 发票
- **CancelPosInvoice**: 取消 POS 发票

#### 支付方式管理
- **GetModeOfPaymentList**: 获取支付方式列表

#### 客户管理
- **CreateCustomer**: 创建客户
- **UpdateCustomer**: 更新客户
- **GetCustomer**: 获取客户详情
- **ListCustomers**: 获取客户列表
- **CountCustomer**: 统计客户数量
- **AddCompanyToCustomer**: 将公司添加到客户的允许交易公司列表

## 使用说明

### 服务注册
```go
service.RegisterSelling(sellingImpl)
service.RegisterAsyncSelling(asyncSellingImpl)
```

### 服务调用
```go
selling := service.Selling()
asyncSelling := service.AsyncSelling()
```
