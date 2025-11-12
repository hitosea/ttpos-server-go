# Setup 服务功能说明

## 概述
Setup 服务提供系统初始化和配置功能，包括店铺初始化、用户创建、数据迁移等。

## 服务接口

### ISetup - 系统配置

#### 店铺初始化
- **InitShop**: 初始化店铺（完整流程）
- **CreateBranch**: 创建分店
- **CreateDefaultPosProfile**: 创建默认的 POS 配置文件

#### 用户管理
- **CreateUser**: 创建网站用户、门店收银账户
- **GetUserApiKeySecret**: 获取用户的 API 密钥和密钥

#### 数据迁移
- **InitErpDocTypeWithDirname**: 初始化 ERP 文档类型
- **InitCustomFields**: 初始化自定义字段（遍历 JSON 文件创建）
- **InitCustomers**: 初始化客户数据（遍历 JSON 文件创建）
- **InitModeOfPayment**: 初始化支付方式（遍历 JSON 文件创建）

#### 时区管理
- **GetDefaultTimeZone**: 获取默认时区
- **MustGetLocalDateTime**: 获取本地时间（UTC 转本地时区）

## 业务场景

### 店铺初始化流程
1. 创建分店（Branch）
2. 创建仓库（Warehouse）
3. 创建用户账户
4. 创建 POS 配置文件
5. 初始化基础数据

### 数据迁移
- 从 manifest 目录读取 JSON 文件
- 批量创建 ERPNext 文档
- 支持自定义字段、客户、支付方式等数据迁移

## 使用说明

### 服务注册
```go
service.RegisterSetup(setupImpl)
```

### 服务调用
```go
setup := service.Setup()

// 初始化店铺
resp, err := setup.InitShop(ctx, &setup.InitShopReq{
    // 店铺信息
})
```

## 注意事项
1. 初始化操作通常只执行一次
2. 数据迁移需要确保 JSON 文件格式正确
3. 时区转换失败时返回 UTC 时间
