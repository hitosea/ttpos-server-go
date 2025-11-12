# Core 服务功能说明

## 概述
Core 服务提供核心基础功能，包括 POS 价格表管理和用户信息管理。

## 服务接口

### IPosPriceList - POS 价格表管理

#### 价格表基础操作
- **CreatePosPriceList**: 创建 POS 价格表规则
- **GetPosPriceList**: 获取价格表规则详情（根据名称）
- **UpdatePosPriceList**: 更新价格表规则
- **DeletePosPriceList**: 删除价格表规则（根据名称）
- **ListPosPriceLists**: 获取价格表规则列表

#### 价格表配置
- **GetDefaultPosPriceList**: 从配置文件获取默认价格表
- **GetPosPriceListByCompany**: 根据公司获取默认价格表规则

### IUser - 用户管理

#### 用户信息查询
- **GetUserByUsername**: 根据用户名（邮箱）获取用户信息

#### 用户时区管理
- **MustGetUserTimeZone**: 获取用户时区
  - 根据用户邮箱查询用户时区设置
  - 用于时间显示和计算的本地化处理

## 业务场景

### 价格表管理
- 为不同公司或门店配置专属价格表
- 支持多价格表规则并存
- 价格表规则可以灵活创建、更新和删除

### 用户时区处理
- 多时区用户的时间本地化显示
- 确保时间数据的准确性和一致性

## 使用说明

### 服务注册
```go
service.RegisterPosPriceList(posPriceListImpl)
service.RegisterUser(userImpl)
```

### 服务调用
```go
// 获取价格表服务实例
priceList := service.PosPriceList()

// 获取用户服务实例
user := service.User()

// 获取公司默认价格表
priceList, err := priceList.GetPosPriceListByCompany(ctx, "公司名称")

// 获取用户时区
timezone := user.MustGetUserTimeZone(ctx, "user@example.com")
```

## 数据模型

### POS 价格表规则
- 支持按公司维度配置
- 可设置默认价格表
- 价格表规则可以被多个 POS 配置文件引用

### 用户信息
- 用户邮箱作为唯一标识
- 时区信息用于时间本地化

## 注意事项
1. 价格表名称需要保证唯一性
2. 删除价格表前需要确认没有被引用
3. MustGetUserTimeZone 方法会返回默认时区，不会返回错误
4. 价格表配置变更会影响 POS 系统的价格计算
