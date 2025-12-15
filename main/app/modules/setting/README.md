# Setting 模块

## 概述

Setting 模块是 TTPOS 系统中的配置管理模块，采用 DDD（领域驱动设计）架构重新设计和实现。该模块负责管理系统的各种配置项，包括门店设置、收银机设置、打印机设置、业务设置等。

## 架构设计

### 目录结构

```
modules/setting/
├── application/          # 应用服务层
│   ├── setting_app_service.go         # 应用服务接口
│   └── setting_app_service_impl.go    # 应用服务实现
├── domain/               # 领域层
│   ├── entity/           # 领域实体
│   │   ├── setting.go                 # 设置聚合根
│   │   ├── store_setting.go           # 门店设置实体
│   │   ├── cashier_setting.go         # 收银机设置实体
│   │   ├── printer_setting.go         # 打印机设置实体
│   │   └── business_setting.go        # 业务设置实体
│   ├── valueobject/      # 值对象
│   │   ├── language.go                # 语言值对象
│   │   ├── carousel.go                # 轮播图值对象
│   │   └── timezone.go                # 时区值对象
│   ├── service/         # 领域服务
│   │   ├── setting_domain_service.go              # 设置域服务接口
│   │   ├── setting_domain_service_impl.go         # 设置域服务实现
│   │   ├── store_setting_domain_service.go        # 门店设置域服务
│   │   ├── cashier_setting_domain_service.go      # 收银机设置域服务
│   │   ├── printer_setting_domain_service.go      # 打印机设置域服务
│   │   └── business_setting_domain_service.go     # 业务设置域服务
│   └── repository/      # 存储接口
│       └── setting_repository.go       # 设置存储接口
├── infrastructure/      # 基础设施层
│   └── persistence/     # 持久化实现
│       └── setting_repository_impl.go  # 设置存储实现
├── adapter/             # 适配器层（兼容旧代码）
│   └── legacy_adapter.go # 遗留代码适配器
├── module.go            # 模块配置
└── README.md            # 模块文档
```

### 核心概念

#### 聚合根 (Aggregate Root)
- **Setting**: 设置项的聚合根，管理设置项的基本信息和生命周期

#### 实体 (Entity)
- **StoreSetting**: 门店设置实体，包含门店基本信息、语言设置等
- **CashierSetting**: 收银机设置实体，包含收银机相关配置
- **PrinterSetting**: 打印机设置实体，包含打印机相关配置
- **BusinessSetting**: 业务设置实体，包含业务规则相关配置

#### 值对象 (Value Object)
- **LanguageItem**: 语言项值对象
- **CarouselItem**: 轮播图项值对象
- **TimeZoneItem**: 时区项值对象

#### 领域服务 (Domain Service)
- **SettingDomainService**: 处理跨实体的业务逻辑，如设置验证、数据转换等
- **StoreSettingDomainService**: 门店设置相关的业务逻辑
- **CashierSettingDomainService**: 收银机设置相关的业务逻辑
- **PrinterSettingDomainService**: 打印机设置相关的业务逻辑
- **BusinessSettingDomainService**: 业务设置相关的业务逻辑

### 应用服务 (Application Service)

应用服务层提供对外的API接口，协调领域对象完成业务用例：

- **GetStoreSetting**: 获取门店设置
- **GetCashierSetting**: 获取收银机设置
- **UpdateSetting**: 更新设置
- **VerifyPassword**: 密码验证

### 存储接口 (Repository)

提供数据访问的抽象接口：

- **ISettingRepository**: 设置存储接口
- **ICloudSettingRepository**: 云端设置存储接口
- **IPrinterRepository**: 打印机存储接口
- **ICompanySettingRepository**: 公司设置存储接口
- **IDeviceRepository**: 设备存储接口

## 重构说明

### 重构目标

1. **提高可维护性**: 将复杂的单体服务拆分为职责明确的模块
2. **遵循DDD原则**: 采用领域驱动设计，提高代码的业务表达力
3. **降低耦合度**: 通过依赖倒置和接口隔离降低模块间的耦合
4. **提高可测试性**: 分层架构使得单元测试更容易编写
5. **保持向后兼容**: 通过适配器层保证现有代码的正常运行

### 重构过程

1. **分析现有代码**: 梳理原有 setting 服务的功能和依赖关系
2. **设计领域模型**: 根据业务需求识别实体、值对象和领域服务
3. **定义接口**: 设计存储接口和应用服务接口
4. **实现基础设施**: 实现存储层和基础服务
5. **创建适配器**: 提供向后兼容的适配器层
6. **逐步迁移**: 逐步将原有代码迁移到新架构

### 兼容性保证

为了保证现有代码的正常运行，提供了一个适配器层：

```go
// 适配器提供与原接口相同的API
type LegacySettingSrv struct {
    appSvc application.ISettingAppService
}

// 保持原有的工厂函数
func NewSrv(dbm *database.DBManager, cache cache.Cache) ISrv {
    return NewLegacySettingSrv(dbm, cache)
}
```

## 使用方式

### 新架构使用

```go
// 创建设置模块
settingModule := setting.NewModule(dbManager, cache)

// 获取应用服务
settingAppSvc := settingModule.GetSettingAppService()

// 使用应用服务
storeSetting, err := settingAppSvc.GetStoreSetting(ctx)
```

### 兼容旧代码

```go
// 旧代码继续使用原有方式
settingSrv := setting.NewSrv(dbManager, cache)
storeSetting, err := settingSrv.GetStoreSetting(ctx)
```

## 测试

模块提供了完整的单元测试支持：

- **领域层测试**: 测试实体和领域服务的业务逻辑
- **应用层测试**: 测试应用服务的编排逻辑
- **基础设施测试**: 测试存储实现的正确性

## 扩展性

模块设计具有良好的扩展性：

1. **添加新设置类型**: 通过实现新的实体和领域服务
2. **扩展存储方式**: 通过实现新的 Repository 接口
3. **添加业务规则**: 通过扩展领域服务
4. **集成外部系统**: 通过适配器模式

## 注意事项

1. **事务管理**: 跨聚合的操作需要考虑分布式事务
2. **缓存策略**: 根据业务需求调整缓存策略
3. **性能优化**: 对高频查询进行性能优化
4. **数据迁移**: 确保数据结构变更时的向后兼容性

## 后续计划

1. **完善实现**: 补全所有方法的实现
2. **性能优化**: 添加缓存和查询优化
3. **监控告警**: 添加业务指标监控
4. **文档完善**: 补充API文档和使用指南
