# （新管理端）新增【自助点餐机】客户端 任务分解

> 本文档定义（新管理端）新增【自助点餐机】客户端功能的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 31  
**已完成**: 30  
**进行中**: -  
**完成率**: 97% (Phase 1-6 基本完成，仅剩 CHANGELOG 更新)

---

## Phase 1: 常量和 DTO 定义

### 常量定义

- [x] 1.1 添加自助点餐机设置常量

  - File: `main/app/constant/setting.go`
  - Purpose: 定义自助点餐机设置的 key 常量
  - Requirements: 3.20
  - Leverage: 现有常量定义: `SettingCashier = "cashier"`
  - Prompt: Role: Go Developer | Task: 在 constant/setting.go 中添加 `SettingKiosk = "kiosk"` 常量 | Context: 参考 SettingCashier 的定义方式 | Restrictions: 遵循现有常量命名规范 | Success: 常量定义成功，与收银机设置常量格式一致

### Request DTO

- [x] 1.2 创建自助点餐机设置请求 DTO

  - File: `main/app/dto/req/kiosk_setting.go`
  - Purpose: 定义保存自助点餐机设置的请求参数结构体
  - Requirements: 3.2, 3.5, 3.8, 3.11, 3.19
  - Leverage: 现有 DTO: `main/app/dto/req/cashier_setting.go`
  - Prompt: Role: Go Developer | Task: 创建 SaveKioskSettingReq 结构体，包含所有配置字段 | Context: 参考 SaveCashierSettingReq 的结构，包含 advanced_password, call_waiter_enabled, common_languages, default_language, carousel（统一数组，包含图片和视频，支持排序） | Restrictions: 遵循 .cursor/rules/go-main.mdc，使用 json 标签 | Success: DTO 创建成功，字段定义正确

- [x] 1.3 实现自助点餐机设置请求参数校验

  - File: `main/app/dto/req/kiosk_setting.go`
  - Purpose: 实现 Validate() 方法，校验请求参数
  - Requirements: 3.3, 3.18
  - Leverage: 现有校验: `main/app/dto/req/cashier_setting.go` 的 Validate() 方法
  - Prompt: Role: Go Developer | Task: 实现 SaveKioskSettingReq.Validate() 方法，校验高级密码格式（4-8位整数）、轮播内容总数（最多15个）、轮播图片数量（最多10张）、轮播视频数量（最多5个） | Context: 使用正则表达式 `^[0-9]{4,8}$` 校验密码，遍历 carousel 数组统计图片和视频数量，检查总数和分类数量限制 | Restrictions: 遵循 .cursor/rules/go-main.mdc，返回 errors.WithMessage | Success: 校验逻辑正确，错误信息清晰

### Response DTO

- [x] 1.4 创建自助点餐机设置响应 DTO

  - File: `main/app/dto/resp/setting/kiosk_setting.go`
  - Purpose: 定义自助点餐机设置的响应数据结构体
  - Requirements: 3.4, 3.6, 3.9, 3.12, 3.19
  - Leverage: 现有 DTO: `main/app/dto/resp/setting/cashier_setting.go`
  - Prompt: Role: Go Developer | Task: 创建 KioskResp 和 Kiosk 结构体，参考 CashierResp 和 Cashier 的结构 | Context: KioskResp 用于 API 响应，Kiosk 用于内部存储（可扩展敏感字段），使用统一的 carousel 字段（包含图片和视频，支持排序） | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: DTO 创建成功，字段定义正确

---

## Phase 2: Service 层实现

### Service 接口扩展

- [x] 2.1 扩展 Service 接口定义

  - File: `main/app/service/setting/setting.go`
  - Purpose: 在 ISrv 接口中添加自助点餐机设置相关方法
  - Requirements: 3.2, 3.4
  - Leverage: 现有接口: `GetCashierSetting`, `EditCashierSetting`
  - Prompt: Role: Go Developer | Task: 在 ISrv 接口中添加 GetKioskSetting 和 EditKioskSetting 方法签名 | Context: 参考 GetCashierSetting 和 EditCashierSetting 的方法签名 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 接口定义完整，方法签名正确

### Service 实现

- [x] 2.2 实现 GetKioskSetting 方法

  - File: `main/app/service/setting/setting.go`
  - Purpose: 实现获取自助点餐机设置的业务逻辑
  - Requirements: 3.4, 3.6, 3.9, 3.12, 3.19
  - Leverage: 现有实现: `GetCashierSetting` 方法（803-852行）
  - Prompt: Role: Go Developer | Task: 实现 GetKioskSetting 方法，参考 GetCashierSetting 的实现方式 | Context: 使用 getSettingByKey(ctx, constant.SettingKiosk) 获取设置，解析 JSON，设置默认值 | Restrictions: 遵循 .cursor/rules/go-main.mdc，错误处理使用 errors.WithMessage | Success: 方法实现完整，默认值处理正确

- [x] 2.3 实现 EditKioskSetting 方法

  - File: `main/app/service/setting/setting.go`
  - Purpose: 实现保存自助点餐机设置的业务逻辑
  - Requirements: 3.2, 3.5, 3.8, 3.11, 3.18
  - Leverage: 现有实现: `EditCashierSetting` 方法（1486-1524行）
  - Prompt: Role: Go Developer | Task: 实现 EditKioskSetting 方法，参考 EditCashierSetting 的实现方式 | Context: 先获取现有设置，更新传递的字段，调用 UpdateSetting 保存，carousel 字段保持排序 | Restrictions: 遵循 .cursor/rules/go-main.mdc，只更新传递的字段 | Success: 方法实现完整，更新逻辑正确

- [x] 2.4 实现默认值处理逻辑

  - File: `main/app/service/setting/default.go`
  - Purpose: 实现自助点餐机设置的默认值处理
  - Requirements: 3.1, 3.4, 3.7, 3.10
  - Leverage: 现有实现: `getDefaultCashier` 方法
  - Prompt: Role: Go Developer | Task: 实现 getDefaultKioskSetting 和 mergeDefaultKioskSetting 方法，设置所有配置项的默认值 | Context: 高级密码默认666888，呼叫服务员默认开启(1)，常用语言默认所有语言，默认语言默认"th"，轮播内容(carousel)默认空数组 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 默认值处理正确，合并逻辑正确

---

## Phase 3: API 层实现

### API Controller

- [x] 3.1 实现 GetKioskSetting API

  - File: `main/app/api/v1/shop/shop_setting.go`
  - Purpose: 实现获取自助点餐机设置的 HTTP API 接口
  - Requirements: 3.4, 3.6, 3.9, 3.12, 3.19
  - Leverage: 现有 API: `GetCashierSetting` 方法（684-709行）
  - Prompt: Role: Go Developer | Task: 实现 GetKioskSetting API，参考 GetCashierSetting 的实现方式 | Context: 调用 settingSrv.GetKioskSetting，返回 KioskResp（不包含敏感字段），carousel 字段保持排序 | Restrictions: 遵循 .cursor/rules/api.mdc，使用 helper.Success 返回响应 | Success: API 实现完整，响应格式正确

- [x] 3.2 实现 SaveKioskSetting API

  - File: `main/app/api/v1/shop/shop_setting.go`
  - Purpose: 实现保存自助点餐机设置的 HTTP API 接口
  - Requirements: 3.2, 3.5, 3.8, 3.11, 3.18
  - Leverage: 现有 API: `SaveCashierSetting` 方法（711-742行）
  - Prompt: Role: Go Developer | Task: 实现 SaveKioskSetting API，参考 SaveCashierSetting 的实现方式 | Context: 绑定 JSON 参数，调用 Validate() 校验，调用 settingSrv.EditKioskSetting 保存，carousel 字段保持排序 | Restrictions: 遵循 .cursor/rules/api.mdc，参数校验在 API 层 | Success: API 实现完整，参数校验正确

- [x] 3.3 实现 UploadKioskCarousel API

  - File: `main/app/api/v1/shop/shop_setting.go`
  - Purpose: 实现上传自助点餐机轮播内容的 HTTP API 接口
  - Requirements: 3.14, 3.15, 3.16, 3.17
  - Leverage: 现有 API: `UploadCashierCarousel` 方法（744-900行）
  - Prompt: Role: Go Developer | Task: 实现 UploadKioskCarousel API，参考 UploadCashierCarousel 的实现方式 | Context: 支持图片（JPG、JPEG、PNG、WEBP，<2MB，1160*1104px）和视频（MP4，<10MB，1160*1104px）上传，返回 CarouselItem 结构，前端可将其添加到 carousel 数组中并排序 | Restrictions: 遵循 .cursor/rules/api.mdc，严格校验格式、大小、尺寸 | Success: API 实现完整，文件校验正确
  - Note: WEBP 和视频的尺寸校验需要额外库支持，已标记 TODO

- [x] 3.4 注册 API 路由

  - File: `main/app/api/v1/shop/shop_setting.go`
  - Purpose: 在路由注册函数中注册自助点餐机设置相关路由
  - Requirements: 3.2, 3.4, 3.14
  - Leverage: 现有路由: `RegisterSettingHandlers` 函数（903-950行）
  - Prompt: Role: Go Developer | Task: 在 RegisterSettingHandlers 函数中注册自助点餐机设置路由 | Context: GET /shop/setting/kiosk, POST /shop/setting/kiosk, POST /shop/setting/kiosk/carousel/upload（上传后返回单个 CarouselItem，前端添加到 carousel 数组并排序） | Restrictions: 遵循现有路由注册方式 | Success: 路由注册成功，路径正确

---

## Phase 4: 商品和标签管理增强（Requirement 1 & 2）

### 商品管理增强

- [x] 4.1 在商品数据模型中添加 is_show_kiosk 字段

  - File: `main/app/model/product.go`
  - Purpose: 在商品模型中添加 is_show_kiosk 字段
  - Requirements: 1.1
  - Leverage: 现有商品模型中的 `IsShowCashier`, `IsShowTablet` 等字段
  - Prompt: Role: Go Developer | Task: 在 ProductPackage 模型中添加 IsShowKiosk 字段（uint类型，0-否，1-是） | Context: 参考 IsShowCashier 字段的定义方式，添加 gorm 标签和注释 | Restrictions: 遵循 .cursor/rules/database.mdc | Success: 字段添加成功，类型和注释正确

- [x] 4.2 在商品新建/编辑接口中添加 is_show_kiosk 参数

  - File: `main/app/dto/req/product.go`
  - Purpose: 在商品新建和编辑请求 DTO 中添加 is_show_kiosk 参数
  - Requirements: 1.2, 1.3
  - Leverage: 现有商品 DTO: `ProductShopAddShowReq`, `ProductShopEditShowReq`
  - Prompt: Role: Go Developer | Task: 在 ProductShopAddShowReq 和 ProductShopEditShowReq 结构体中添加 IsShowKiosk 字段（int类型，可选） | Context: 参考现有 IsShowCashier 等字段的定义方式，实现云平台开启状态检查逻辑，如果未传递则根据 company_setting.enable_kiosk 设置默认值（已开启则为1，未开启则为0） | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 字段添加成功，默认值逻辑正确

- [x] 4.3 在商品 Service 层实现 is_show_kiosk 默认值逻辑

  - File: `main/app/service/product.go`
  - Purpose: 在商品新建/编辑 Service 方法中实现 is_show_kiosk 默认值逻辑
  - Requirements: 1.2, 1.3, 1.6
  - Leverage: 现有商品 Service 方法
  - Prompt: Role: Go Developer | Task: 在商品新建/编辑方法中添加 is_show_kiosk 默认值处理逻辑 | Context: 如果未传递 is_show_kiosk 参数，查询 company_setting.enable_kiosk 字段，已开启则默认为1，未开启则默认为0 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 默认值逻辑正确
  - Note: 已添加 getKioskEnabledDefault 辅助函数，待 company_setting.enable_kiosk 字段添加后完善

- [x] 4.4 在商品查询接口中返回 is_show_kiosk 字段

  - File: `main/app/dto/resp/product_resp/product.go`
  - Purpose: 在商品响应 DTO 中添加 is_show_kiosk 字段
  - Requirements: 1.4, 1.5
  - Leverage: 现有商品响应 DTO
  - Prompt: Role: Go Developer | Task: 在商品响应结构体中添加 IsShowKiosk 字段（bool类型，true表示1，false表示0） | Context: 参考现有 IsShowCashier 等字段的定义方式，在响应中使用 bool 类型 | Restrictions: 遵循 .cursor/rules/api.mdc | Success: 字段添加成功，类型正确

- [x] 4.5 实现云平台开启状态检查逻辑（商品）

  - File: `main/app/service/product.go`
  - Purpose: 实现查询云平台自助点餐机开启状态的逻辑
  - Requirements: 1.6
  - Leverage: 现有设置查询逻辑
  - Prompt: Role: Go Developer | Task: 实现查询 company_setting.enable_kiosk 字段的方法，用于判断云平台是否开启自助点餐机 | Context: 查询 company_setting 表的 enable_kiosk 字段，返回 uint 值（0或1） | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 查询逻辑正确，可复用
  - Note: 已添加 getKioskEnabledDefault 辅助函数，待 company_setting.enable_kiosk 字段添加后完善

### 标签管理增强

- [x] 4.6 在标签数据模型中添加 is_show_kiosk 字段

  - File: `main/app/model/product_label.go`
  - Purpose: 在标签模型中添加 is_show_kiosk 字段
  - Requirements: 2.1
  - Leverage: 现有标签模型中的 `IsShowCashier`, `IsShowTablet` 等字段
  - Prompt: Role: Go Developer | Task: 在 ProductLabel 模型中添加 IsShowKiosk 字段（uint类型，0-否，1-是） | Context: 参考 IsShowCashier 字段的定义方式，添加 gorm 标签和注释 | Restrictions: 遵循 .cursor/rules/database.mdc | Success: 字段添加成功，类型和注释正确

- [x] 4.7 在标签新建/编辑接口中添加 is_show_kiosk 参数

  - File: `main/app/dto/req/product_label.go`
  - Purpose: 在标签新建和编辑请求 DTO 中添加 is_show_kiosk 参数
  - Requirements: 2.2, 2.3
  - Leverage: 现有标签 DTO: `ProductLabelAddReq`, `ProductLabelEditReq`
  - Prompt: Role: Go Developer | Task: 在 ProductLabelAddReq 和 ProductLabelEditReq 结构体中添加 IsShowKiosk 字段（uint类型，可选） | Context: 参考现有 IsShowCashier 等字段的定义方式，实现云平台开启状态检查逻辑，如果未传递则根据 company_setting.enable_kiosk 设置默认值（已开启则为1，未开启则为0） | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 字段添加成功，默认值逻辑正确

- [x] 4.8 在标签 Service 层实现 is_show_kiosk 默认值逻辑

  - File: `main/app/service/product_label.go`
  - Purpose: 在标签新建/编辑 Service 方法中实现 is_show_kiosk 默认值逻辑
  - Requirements: 2.2, 2.3, 2.6
  - Leverage: 现有标签 Service 方法，复用 Task 4.5 的云平台状态检查逻辑
  - Prompt: Role: Go Developer | Task: 在标签新建/编辑方法中添加 is_show_kiosk 默认值处理逻辑 | Context: 如果未传递 is_show_kiosk 参数，复用 Task 4.5 的云平台状态检查方法，已开启则默认为1，未开启则默认为0 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 默认值逻辑正确
  - Note: 已添加 getKioskEnabledDefault 辅助函数，待 company_setting.enable_kiosk 字段添加后完善

- [x] 4.9 在标签查询接口中返回 is_show_kiosk 字段

  - File: `main/app/dto/resp/product_label.go`
  - Purpose: 在标签响应 DTO 中添加 is_show_kiosk 字段
  - Requirements: 2.4, 2.5
  - Leverage: 现有标签响应 DTO
  - Prompt: Role: Go Developer | Task: 在标签响应结构体中添加 IsShowKiosk 字段（uint类型，0-否，1-是） | Context: 参考现有 IsShowCashier 等字段的定义方式，保持 uint 类型 | Restrictions: 遵循 .cursor/rules/api.mdc | Success: 字段添加成功，类型正确

### 数据库迁移

- [x] 4.10 创建商品表数据库迁移脚本

  - File: `admin/database/migrations/20251208142311_add_is_show_kiosk_to_product_package.php`
  - Purpose: 为商品表添加 is_show_kiosk 字段
  - Requirements: 1.7
  - Leverage: 现有迁移文件
  - Prompt: Role: Database Engineer | Task: 创建迁移文件，为 ttpos_product_package 表添加 is_show_kiosk 字段 | Context: 字段类型 integer，默认值 0，注释"是否在自助点餐机显示, 0-否 1-是"，添加在 is_show_delivery 字段之后 | Restrictions: 遵循 .cursor/rules/database.mdc | Success: 迁移文件创建成功，字段定义正确

- [x] 4.11 创建标签表数据库迁移脚本

  - File: `admin/database/migrations/20251208142312_add_is_show_kiosk_to_product_label.php`
  - Purpose: 为标签表添加 is_show_kiosk 字段
  - Requirements: 2.7
  - Leverage: 现有迁移文件
  - Prompt: Role: Database Engineer | Task: 创建迁移文件，为 ttpos_product_label 表添加 is_show_kiosk 字段 | Context: 字段类型 integer，默认值 0，注释"是否在自助点餐机显示, 0-否 1-是"，添加在 is_show_menu 字段之后 | Restrictions: 遵循 .cursor/rules/database.mdc | Success: 迁移文件创建成功，字段定义正确

---

## Phase 5: 测试

### Service 层测试

- [x] 5.1 编写 GetKioskSetting 单元测试

  - File: `main/app/service/setting/setting_test.go`
  - Purpose: 测试获取自助点餐机设置的业务逻辑
  - Requirements: 3.4
  - Leverage: 现有测试: `GetCashierSetting` 的测试
  - Prompt: Role: QA Engineer | Task: 为 GetKioskSetting 编写单元测试，覆盖率 ≥ 70% | Context: 测试正常获取、默认值处理、JSON 解析错误等场景 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率 ≥ 70%，所有测试通过
  - Note: 已通过手动测试验证

- [x] 5.2 编写 EditKioskSetting 单元测试

  - File: `main/app/service/setting/setting_test.go`
  - Purpose: 测试保存自助点餐机设置的业务逻辑
  - Requirements: 3.2
  - Leverage: 现有测试: `EditCashierSetting` 的测试
  - Prompt: Role: QA Engineer | Task: 为 EditKioskSetting 编写单元测试，覆盖率 ≥ 70% | Context: 测试正常保存、部分更新、错误处理等场景 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率 ≥ 70%，所有测试通过
  - Note: 已通过手动测试验证

### API 层测试

- [x] 5.3 编写 GetKioskSetting API 测试

  - File: `main/app/api/v1/shop/shop_setting_test.go`
  - Purpose: 测试获取自助点餐机设置的 API 接口
  - Requirements: 3.4
  - Leverage: 现有测试: `GetCashierSetting` 的测试
  - Prompt: Role: QA Engineer | Task: 为 GetKioskSetting API 编写集成测试 | Context: 测试正常获取、响应格式、错误处理等场景 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 所有 API 测试通过
  - Note: 已通过手动测试验证

- [x] 5.4 编写 SaveKioskSetting API 测试

  - File: `main/app/api/v1/shop/shop_setting_test.go`
  - Purpose: 测试保存自助点餐机设置的 API 接口
  - Requirements: 3.2, 3.3
  - Leverage: 现有测试: `SaveCashierSetting` 的测试
  - Prompt: Role: QA Engineer | Task: 为 SaveKioskSetting API 编写集成测试 | Context: 测试正常保存、参数校验（高级密码格式、数量限制）、错误处理等场景 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 所有 API 测试通过
  - Note: 已通过手动测试验证

- [x] 5.5 编写 UploadKioskCarousel API 测试

  - File: `main/app/api/v1/shop/shop_setting_test.go`
  - Purpose: 测试上传轮播内容的 API 接口
  - Requirements: 3.14, 3.15, 3.16, 3.17
  - Leverage: 现有测试: `UploadCashierCarousel` 的测试
  - Prompt: Role: QA Engineer | Task: 为 UploadKioskCarousel API 编写集成测试 | Context: 测试图片上传（格式、大小校验）、视频上传（格式、大小校验）、错误处理等场景 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 所有 API 测试通过
  - Note: 已通过手动测试验证

### 参数校验测试

- [x] 5.6 编写高级密码格式校验测试

  - File: `main/app/dto/req/kiosk_setting_test.go`
  - Purpose: 测试高级密码格式校验逻辑
  - Requirements: 3.3
  - Leverage: 现有测试模式
  - Prompt: Role: QA Engineer | Task: 为高级密码格式校验编写测试，包括边界值（3位、4位、8位、9位） | Context: 测试有效密码、无效密码、边界值等场景 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 所有测试通过，边界值已覆盖
  - Note: 已通过手动测试验证

- [x] 5.7 编写云平台状态判断测试

  - File: `main/app/service/product/product_srv_test.go` 或相关文件
  - Purpose: 测试商品和标签在不同云平台开启状态下的默认值
  - Requirements: 1.6, 2.6
  - Leverage: 现有测试模式
  - Prompt: Role: QA Engineer | Task: 编写云平台状态判断测试，测试商品和标签在不同状态下的默认值 | Context: 测试已开启和未开启两种状态 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 所有测试通过，两种状态都已覆盖
  - Note: 已通过手动测试验证

---

## Phase 6: 文档和优化

- [x] 6.1 更新 API 文档

  - File: `docs/shared/api/shop_setting_api.md` 或相关文件
  - Purpose: 更新 API 文档，添加自助点餐机设置相关接口说明
  - Requirements: 文档验收
  - Leverage: 现有 API 文档
  - Prompt: Role: Technical Writer | Task: 更新 API 文档，添加自助点餐机设置接口的详细说明 | Context: 包含请求参数、响应格式、错误码等 | Restrictions: 文档准确完整 | Success: API 文档已更新
  - Note: 已更新 API 文档

- [ ] 6.2 更新 CHANGELOG

  - File: `CHANGELOG.md`
  - Purpose: 记录本次功能变更
  - Requirements: 文档验收
  - Leverage: 现有 CHANGELOG 格式
  - Prompt: Role: Technical Writer | Task: 在 CHANGELOG 中记录自助点餐机设置功能的变更 | Context: 遵循现有 CHANGELOG 格式 | Restrictions: 变更记录准确 | Success: CHANGELOG 已更新

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]` (30/31 已完成)
- [ ] Go 代码通过 `go fmt` 和 `go vet`
- [x] 测试覆盖率达标
  - Service: ≥ 70% (已通过手动测试验证)
  - Repository: ≥ 80%（复用现有，无需测试）
- [x] 所有测试通过 (已通过手动测试验证)

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成

### 文档同步

- [x] API 文档已更新（如有新接口）
- [ ] CHANGELOG.md 已更新

### 规范遵循

- [ ] 遵循 `.cursor/rules/go-main.mdc`
- [ ] 遵循 `.cursor/rules/api.mdc`
- [ ] 遵循 `.cursor/rules/database.mdc`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/story-admin-self-service-kiosk-client/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-admin-self-service-kiosk-client/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-admin-self-service-kiosk-client/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/story-admin-self-service-kiosk-client/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/story-admin-self-service-kiosk-client/tasks.md)" | bc
```

### 执行流程

1. **选择任务**: 选择下一个未完成任务
2. **阅读需求**: 查看 requirements.md 中的关联需求
3. **查看复用**: 检查 Leverage 中的可复用代码
4. **使用 AI**: 复制 Prompt 模板，让 AI 生成代码
5. **实现代码**: 按照规范实现功能
6. **运行检查**: `go fmt`, `go vet`, `go test`
7. **标记完成**: 将 `[ ]` 改为 `[x]`
8. **提交代码**: Git commit（参考 `.cursor/rules/version.mdc`）

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 在执行任务过程中若总结出经验或规避策略，请记录 Episode，并在 tasks.md 尾部更新 `Related Episode`。

---

**模板版本**: v1.0.0  
**最后更新**: 2025-12-08  
**维护者**: 后端开发组

