# 新管理端-增加打印设置：自定义收银结账单打印联数 设计文档

> 本文档定义自定义收银结账单打印联数功能的技术设计和实现方案。

## 📋 概述

本功能通过扩展打印设置配置，允许商户管理员自定义收银结账单的打印联数（0-10份）。核心实现包括：

- 在打印设置中增加"票据类型打印联数设置"开关和结账单打印联数字段
- 修改打印联数获取逻辑，优先使用结账单打印联数配置
- 创建新管理端打印设置配置页面
- 实现配置同步机制，确保POS/Assistant端及时获取最新配置

---

## 🎯 规范对齐

### Go Main 规范 (go-main.mdc)

- ✅ Service 只依赖其他 Service 接口
- ✅ Repository 只持有 db 实例
- ✅ URL 使用 snake_case
- ✅ data 字段必须是对象
- ✅ 不使用 panic，返回 error
- ✅ 使用 errors.WithMessage 包装错误

### API 设计规范 (api.mdc)

- ✅ 响应格式统一: `{code, message, data{}}`
- ✅ data 不能为 null 或数组
- ✅ URL 使用 snake_case: `/api/v1/print_setting`

### 数据库规范 (database.mdc)

- ✅ 打印设置存储在 `ttpos_setting` 表，使用 JSON 格式
- ✅ 字段使用 snake_case
- ✅ 不新增数据库表，扩展现有 JSON 结构

### Vue 前端规范 (vue.mdc)

- ✅ 使用 Vue 3 + TypeScript + Composition API
- ✅ 使用 Element Plus 组件库
- ✅ 遵循组件化设计

---

## 🔄 代码复用分析

### 可复用的现有组件

- **SettingService**: `main/app/service/setting/setting.go` - 复用设置读取和更新逻辑
- **GetPrinterInfo**: `main/app/service/setting/setting.go` - 扩展打印联数获取逻辑
- **PrintingStatementOrder**: `main/app/printer/order_printer.go` - 复用结账单打印逻辑
- **PrinterSetting**: `main/app/dto/resp/setting/printer_setting.go` - 扩展打印设置结构体

### 集成点

- **打印设置模块**: 扩展 `Printer` 结构体，增加自定义打印联数字段
- **打印服务**: 修改 `GetPrinterInfo` 方法，优先使用结账单打印联数配置
- **配置同步**: 通过现有的设置缓存机制，配置变更后自动同步到POS/Assistant端

---

## 🏗️ 架构设计

### 分层设计

```
PrintSettingAPI (API 层)
  ↓ 调用
SettingService (业务层)
  ↓ 依赖
SettingRepository (数据层)
  ↓ 读取
ttpos_setting 表 (数据库)

PrintingStatementOrder (打印层)
  ↓ 调用
GetPrinterInfo (设置服务)
  ↓ 读取
PrinterSetting (打印设置)
  ↓ 优先级判断
结账单打印联数配置 > 打印机硬件联数
```

**依赖规则**：

- ✅ PrintingStatementOrder 依赖 SettingService 接口
- ✅ SettingService 依赖 SettingRepository
- ✅ 打印联数优先级：收银打印设置-结账单打印联数 > 打印设置-打印机打印联数

---

## 🗄️ 数据库设计

### 数据表设计

#### 表：ttpos_setting（扩展现有表）

打印设置存储在 `ttpos_setting` 表中，`key` 为 `printer`，`values` 为 JSON 格式。本次功能扩展 JSON 结构，不修改表结构。

**现有 JSON 结构**：

```json
{
  "cashier_open": "1",
  "cashier_printer_id": "-1",
  "cashier_printer": [],
  "language_list": [],
  "language_method": "1",
  "default_language": "en",
  "print_method": "1",
  "kitchen_language": "en",
  "kitchen_print_method": "1",
  "consumption_tax": "1",
  "buffet_sign_open": "1",
  "monetary_unit_open": "1",
  "calendar_list": [],
  "print_list": [],
  "default_calendar": "1",
  "language": ["en"]
}
```

**扩展后的 JSON 结构**：

```json
{
  "cashier_open": "1",
  "cashier_printer_id": "-1",
  "cashier_printer": [],
  "language_list": [],
  "language_method": "1",
  "default_language": "en",
  "print_method": "1",
  "kitchen_language": "en",
  "kitchen_print_method": "1",
  "consumption_tax": "1",
  "buffet_sign_open": "1",
  "monetary_unit_open": "1",
  "calendar_list": [],
  "print_list": [],
  "default_calendar": "1",
  "language": ["en"],
  "enable_custom_copies": "0",      // 新增：是否启用自定义打印联数 "0"-关闭 "1"-开启
  "checkout_slip_copies": 0         // 新增：结账单打印联数 0-10
}
```

**字段说明**：

| 字段 | 类型 | 说明 | 默认值 | 约束 |
|------|------|------|--------|------|
| enable_custom_copies | string | 是否启用自定义打印联数 | "0" | "0"关闭，"1"开启 |
| checkout_slip_copies | int | 结账单打印联数 | 0 | 0-10的整数，0表示不打印 |

**兼容性处理**：

- 读取时，如果字段不存在，使用默认值（`enable_custom_copies = "0"`, `checkout_slip_copies = 0`）
- 写入时，始终包含这两个字段

---

## 📊 数据模型

### Go Model

#### 扩展 Printer 结构体

```go
// main/app/dto/resp/setting/printer_setting.go
type Printer struct {
	CashierOpen        string               `json:"cashier_open"`
	CashierPrinterID   string               `json:"cashier_printer_id"`
	CashierPrinter     []CashierPrinterItem `json:"cashier_printer"`
	LanguageList       []dto.LanguageItem   `json:"language_list"`
	LanguageMethod     string               `json:"language_method"`
	DefaultLanguage    string               `json:"default_language"`
	PrintMethod        string               `json:"print_method"`
	KitchenLanguage    string               `json:"kitchen_language"`
	KitchenPrintMethod string               `json:"kitchen_print_method"`
	ConsumptionTax     string               `json:"consumption_tax"`
	BuffetSignOpen     string               `json:"buffet_sign_open"`
	MonetaryUnitOpen   string               `json:"monetary_unit_open"`
	CalendarList       []CalendarItem       `json:"calendar_list"`
	PrintList          []PrintItem          `json:"print_list"`
	DefaultCalendar    string               `json:"default_calendar"`
	Language           []string             `json:"language"`
	// 新增字段
	EnableCustomCopies string `json:"enable_custom_copies"` // 是否启用自定义打印联数 "0"-关闭 "1"-开启
	CheckoutSlipCopies int    `json:"checkout_slip_copies"` // 结账单打印联数 0-10
}
```

### DTO 定义

#### Request DTO

```go
// main/app/dto/req/print_setting_req.go
type UpdatePrintSettingReq struct {
	EnableCustomCopies string `json:"enable_custom_copies" binding:"required,oneof=0 1"` // 是否启用自定义打印联数
	CheckoutSlipCopies int    `json:"checkout_slip_copies" binding:"omitempty,gte=0,lte=10"` // 结账单打印联数 0-10
}
```

#### Response DTO

```go
// main/app/dto/resp/setting/printer_setting.go
// 复用现有的 Printer 结构体，已包含新增字段
```

---

## 🔌 API 设计

### RESTful API

#### API 1: 更新打印设置

**请求**:

- **URL**: `/api/v1/print_setting/update`
- **Method**: `POST`
- **Headers**:
  ```json
  {
    "Authorization": "Bearer {token}",
    "Content-Type": "application/json"
  }
  ```
- **Body**:
  ```json
  {
    "enable_custom_copies": "1",
    "checkout_slip_copies": 2
  }
  ```

**响应**:

```json
{
  "code": 1,
  "message": "success",
  "data": {
    "enable_custom_copies": "1",
    "checkout_slip_copies": 2
  }
}
```

**错误响应**:

```json
{
  "code": 0,
  "message": "结账单打印联数必须在0-10之间",
  "data": {}
}
```

#### API 2: 获取打印设置

**请求**:

- **URL**: `/api/v1/print_setting/get`
- **Method**: `GET`
- **Headers**:
  ```json
  {
    "Authorization": "Bearer {token}"
  }
  ```

**响应**:

```json
{
  "code": 1,
  "message": "success",
  "data": {
    "enable_custom_copies": "0",
    "checkout_slip_copies": 0,
    "cashier_open": "1",
    // ... 其他打印设置字段
  }
}
```

---

## 🧩 组件和接口

### Service 层

#### Service 接口扩展

```go
// main/app/service/setting/setting.go
type ISrv interface {
	// ... 现有方法
	GetPrinterSetting(ctx context.Context, languageList []dto.LanguageItem) (setting.Printer, error)
	GetPrinterInfo(ctx context.Context, printerSetting setting.Printer, deviceId string) (setting.PrinterInfo, error)
	UpdateSetting(ctx context.Context, settingKey string, values any) error
	// 新增方法
	UpdatePrintSetting(ctx context.Context, req *dto_req.UpdatePrintSettingReq) error
}
```

#### Service 实现

**修改 GetPrinterInfo 方法**：

```go
// main/app/service/setting/setting.go
func (s *Srv) GetPrinterInfo(ctx context.Context, printerSetting setting.Printer, deviceSn string) (setting.PrinterInfo, error) {
	var (
		// ... 现有变量
		copies uint = 1
	)

	// ... 现有逻辑：获取打印机硬件联数
	if printer.Uuid > 0 {
		copies = printer.Copies
		// ... 其他逻辑
	}

	// 新增：优先级判断
	// 如果启用了自定义打印联数，且是结账单打印，优先使用结账单打印联数配置
	if printerSetting.EnableCustomCopies == "1" {
		// 注意：这里需要判断是否是结账单打印
		// 由于 GetPrinterInfo 方法不知道打印类型，需要在调用方判断
		// 或者增加一个参数来标识打印类型
		// 暂时在 PrintingStatementOrder 中处理优先级
	}

	return setting.PrinterInfo{
		// ... 现有字段
		Copies: copies,
	}, nil
}
```

**更好的方案：在 PrintingStatementOrder 中处理优先级**：

```go
// main/app/printer/order_printer.go
func (p *PrinterRepoImpl) PrintingStatementOrder(...) (*resp.PrinterData, error) {
	// ... 现有逻辑

	// 获取打印设置
	settingPrinterInfo, err := p.setting.GetPrinterInfo(p.ctx, p.printerSetting, deviceSn)
	if err != nil {
		return nil, errors.WithMessage(err, "获取打印设置失败")
	}

	// 新增：优先级判断
	// 如果启用了自定义打印联数，优先使用结账单打印联数配置
	if p.printerSetting.EnableCustomCopies == "1" {
		if p.printerSetting.CheckoutSlipCopies > 0 {
			settingPrinterInfo.Copies = uint(p.printerSetting.CheckoutSlipCopies)
		} else {
			// 如果设置为0，表示不打印
			settingPrinterInfo.Copies = 0
		}
	}

	// ... 后续逻辑
}
```

**新增 UpdatePrintSetting 方法**：

```go
// main/app/service/setting/setting.go
func (s *Srv) UpdatePrintSetting(ctx context.Context, req *dto_req.UpdatePrintSettingReq) error {
	// 获取当前打印设置
	printerSetting, err := s.GetPrinterSetting(ctx, nil)
	if err != nil {
		return errors.WithMessage(err, "获取打印设置失败")
	}

	// 更新字段
	printerSetting.EnableCustomCopies = req.EnableCustomCopies
	if req.EnableCustomCopies == "1" {
		printerSetting.CheckoutSlipCopies = req.CheckoutSlipCopies
	} else {
		// 关闭时重置为0
		printerSetting.CheckoutSlipCopies = 0
	}

	// 保存设置
	return s.UpdateSetting(ctx, constant.SettingPrinter, printerSetting)
}
```

### API 层

```go
// main/app/api/print_setting_api.go
type PrintSettingAPI struct {
	settingSrv service.ISrv
}

func NewPrintSettingAPI(settingSrv service.ISrv) *PrintSettingAPI {
	return &PrintSettingAPI{settingSrv: settingSrv}
}

// POST /api/v1/print_setting/update
func (api *PrintSettingAPI) Update(c *gin.Context) {
	var req dto_req.UpdatePrintSettingReq
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeInvalidParam, err)
		return
	}

	// 验证：如果启用自定义打印联数，必须提供结账单打印联数
	if req.EnableCustomCopies == "1" && req.CheckoutSlipCopies < 0 || req.CheckoutSlipCopies > 10 {
		helper.Error(c, constant.CodeInvalidParam, "结账单打印联数必须在0-10之间")
		return
	}

	ctx := context.NewContext(c)
	if err := api.settingSrv.UpdatePrintSetting(ctx, &req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}

	helper.Success(c, gin.H{
		"data": gin.H{
			"enable_custom_copies": req.EnableCustomCopies,
			"checkout_slip_copies": req.CheckoutSlipCopies,
		},
	})
}

// GET /api/v1/print_setting/get
func (api *PrintSettingAPI) Get(c *gin.Context) {
	ctx := context.NewContext(c)
	printerSetting, err := api.settingSrv.GetPrinterSetting(ctx, nil)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}

	helper.Success(c, gin.H{
		"data": printerSetting,
	})
}
```

---

## ⚡ 缓存设计

### Redis 缓存

**缓存策略**:

- **Key 命名**: `setting:company_id:{company_uuid}`
- **过期时间**: 无过期时间（手动删除）
- **更新策略**: Cache-Aside Pattern

**缓存更新**:

```go
// main/app/service/setting/setting.go
func (s *Srv) UpdatePrintSetting(ctx context.Context, req *dto_req.UpdatePrintSettingReq) error {
	// ... 更新数据库

	// 删除缓存
	companyUuid := ctx.GetCompanyUuid()
	cacheKey := fmt.Sprintf("setting:company_id:%d", companyUuid)
	s.cache.Delete(cacheKey)

	// ... 返回
}
```

---

## 🚨 错误处理

### 错误场景

#### 场景 1: 输入验证失败

- **处理方式**: 返回参数错误，提示用户输入范围
- **用户影响**: 显示错误提示"结账单打印联数必须在0-10之间"
- **代码示例**:
  ```go
  if req.CheckoutSlipCopies < 0 || req.CheckoutSlipCopies > 10 {
      helper.Error(c, constant.CodeInvalidParam, "结账单打印联数必须在0-10之间")
      return
  }
  ```

#### 场景 2: 配置读取失败

- **处理方式**: 降级到数据库查询，记录错误日志
- **用户影响**: 使用默认值（打印机硬件联数）
- **代码示例**:
  ```go
  printerSetting, err := s.GetPrinterSetting(ctx, nil)
  if err != nil {
      logger.Logger.Error("获取打印设置失败", zap.Error(err))
      // 使用默认值
      copies = printer.Copies
  }
  ```

#### 场景 3: 配置保存失败

- **处理方式**: 事务回滚，返回错误
- **用户影响**: 显示错误提示"保存失败，请重试"
- **代码示例**:
  ```go
  if err := s.UpdateSetting(ctx, constant.SettingPrinter, printerSetting); err != nil {
      return errors.WithMessage(err, "保存打印设置失败")
  }
  ```

---

## 🔒 安全设计

### 身份验证

- **JWT Token**: 所有 API 需要 Token 验证
- **权限验证**: 需要"打印设置"权限

### 数据安全

- **参数验证**: 使用 binding 标签验证输入范围
- **SQL 注入防护**: 使用 GORM 参数化查询
- **XSS 防护**: 前端输入校验

---

## 🧪 测试策略

### 单元测试

**目标覆盖率**:

- SettingService: 70%+
- PrintSettingAPI: 70%+

**测试内容**:

- 打印联数优先级判断逻辑
- 配置更新逻辑
- 参数验证逻辑
- 错误处理逻辑

### API 测试

**测试内容**:

- API 接口调用
- 参数验证
- 响应格式
- 错误处理

### 集成测试

**测试流程**:

- 配置更新 → 缓存更新 → 打印联数生效
- 多语言打印联数一致性
- 补打功能联数正确性

---

## 📈 性能优化

### 优化策略

1. **缓存优化**:
   - 配置读取优先从缓存获取
   - 配置更新后立即删除缓存

2. **数据库优化**:
   - 使用现有索引（company_uuid）

### 性能指标

- 配置读取响应时间: < 100ms（从缓存）
- 配置更新响应时间: < 200ms
- 打印联数判断: < 1ms

---

## 🌐 浏览器兼容性

### 前端兼容性（Vue）

- Chrome 90+
- Safari 14+
- Firefox 88+
- Edge 90+

---

## 📚 实现清单

### Phase 1: 数据模型和 Service 层

- [ ] 扩展 Printer 结构体，增加自定义打印联数字段
- [ ] 实现 UpdatePrintSetting 方法
- [ ] 修改 GetPrinterInfo 方法，支持优先级判断
- [ ] 修改 PrintingStatementOrder 方法，应用优先级规则

### Phase 2: API 层

- [ ] 创建 PrintSettingAPI
- [ ] 实现 Update 和 Get 接口
- [ ] 注册 API 路由

### Phase 3: 前端实现

- [ ] 创建打印设置配置页面
- [ ] 实现开关和输入框组件
- [ ] 实现表单验证
- [ ] 实现 API 调用

### Phase 4: 测试

- [ ] 单元测试
- [ ] API 测试
- [ ] 集成测试
- [ ] 手动测试

**详细任务**: 参见 `tasks.md`

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 在技术方案评审、关键架构决策或踩坑总结后，立即补充 Episode 并在设计文档尾部互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-12-24  
**作者**: 后端开发组  
**审核者**: {待审核}

