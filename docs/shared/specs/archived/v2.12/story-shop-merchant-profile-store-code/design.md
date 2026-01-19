# 新管理端-商家档案增加店铺编码 技术设计文档

> 本文档定义商家档案新增店铺编码字段的技术实现方案。

## 📋 基本信息

| 项目              | 内容                                                                                                                       |
| ----------------- | -------------------------------------------------------------------------------------------------------------------------- |
| **关联 Requirements** | [requirements.md](./requirements.md)                                                                                    |
| **创建日期**      | 2025-12-05                                                                                                                 |
| **设计者**        | weifashi                                                                                                                   |
| **目标版本**      | v2.11.0                                                                                                                   |

## 📋 审核状态

| 项目         | 内容                     |
| ------------ | ------------------------ |
| **审核状态** | 待审核                   |
| **审核人**   | -                        |
| **审核日期** | -                        |
| **审核意见** | -                        |
| **Story Point** | - (待评估)             |

---

## 📋 设计概述

本需求在商家档案配置中新增"店铺编码"字段，用于发票打印时显示店铺编号。技术实现涉及：

1. **后端 Go Main 模块**：扩展店铺设置相关的数据结构和 API
2. **前端 Flutter 新管理端**：在商家档案页面新增店铺编码输入框
3. **数据存储**：店铺编码存储在 `setting` 表的 `store_setting` JSON 字段中
4. **打印模块**：发票打印时读取并显示店铺编码

---

## 🏗️ 架构设计

### 系统层次

```
┌─────────────────────────────────────────────────┐
│         Flutter 新管理端（商家档案页面）          │
│         (ttpos-flutter/shop/profile)            │
└────────────────┬────────────────────────────────┘
                 │ HTTP API
                 ↓
┌─────────────────────────────────────────────────┐
│         Go Main API 层                           │
│         (main/app/api/v1/shop/)                 │
└────────────────┬────────────────────────────────┘
                 │
                 ↓
┌─────────────────────────────────────────────────┐
│         Go Main Service 层                       │
│         (main/app/service/setting/)             │
└────────────────┬────────────────────────────────┘
                 │
                 ↓
┌─────────────────────────────────────────────────┐
│         数据库 (MySQL)                           │
│         setting 表 (store_setting JSON字段)     │
└─────────────────────────────────────────────────┘

打印流程：
┌─────────────────────────────────────────────────┐
│         打印模块 (Printer Module)                │
│         读取 store_code → 渲染发票模板           │
└─────────────────────────────────────────────────┘
```

### 数据流向

```
[商家档案页面] 
    → 填写店铺编码 
    → POST /api/v1/shop/setting/store 
    → EditStoreSetting(req.UpdateStoreSetting) 
    → 更新 setting.store_setting JSON 
    → 清除缓存 
    → 返回成功

[发票打印] 
    → 读取 store_setting.store_code 
    → 渲染打印模板 
    → 输出发票（包含 SHOP: {code}）
```

---

## 💾 数据模型设计

### 1. 数据库设计

#### 影响的表

**表名**: `setting`  
**字段**: `values` (JSON 类型)  
**Key**: `store_setting`

```json
{
  "name": "店铺名称",
  "logo_url": "/uploads/...",
  "time_zone": "Asia/Shanghai",
  "company": "公司名称",
  "address": "地址",
  "phone": "联系电话",
  "tax_number": "税号",
  "store_code": "STORE-001",  // ← 新增字段
  "chain_number": "连锁编号",
  "language": [...],
  "coordinates": "经纬度"
}
```

#### 数据库迁移

由于 `store_code` 存储在 JSON 字段中，不需要创建数据库迁移脚本，直接在代码层面支持即可。

### 2. Go 数据结构设计

#### 2.1 请求结构（main/app/dto/req/base.go）

```go
type UpdateStoreSetting struct {
	Name        string             `json:"name" binding:"required,max=100"`   // 店铺名称
	LogoUrl     string             `json:"logo_url" binding:"required"`       // 店铺logo
	TimeZone    string             `json:"time_zone" binding:"required"`      // 时区
	CompanyName string             `json:"company_name" binding:"max=500"`    // 公司名称
	Address     string             `json:"address" binding:"max=500"`         // 地址
	Phone       string             `json:"phone" binding:"required,max=20"`   // 联系电话
	TaxNumber   string             `json:"tax_number"`                        // 税号
	StoreCode   string             `json:"store_code" binding:"max=100"`      // ← 新增：店铺编码
	Language    []dto.LanguageItem `json:"language" binding:"required,min=1"` // 系统语言
	Coordinates string             `json:"coordinates"`                       // 经纬度
}
```

**验证规则**：
- `max=100`：最大长度 100 字符
- 非必填（没有 `required` 标签）

#### 2.2 响应结构（main/app/dto/resp/base.go）

```go
type ShopProfile struct {
	CompanyName     string                 `json:"company_name"`     // 公司名称
	Address         string                 `json:"address"`          // 地址
	Coordinates     string                 `json:"coordinates"`      // 经纬度
	IpWhiteList     string                 `json:"ip_white_list"`    // ip白名单
	Phone           string                 `json:"phone"`            // 联系电话
	TaxNumber       string                 `json:"tax_number"`       // 税号
	StoreCode       string                 `json:"store_code"`       // ← 新增：店铺编码
	TimeZoneList    []setting.TimeZoneItem `json:"time_zone_list"`   // 时区列表
	DefaultLanguage string                 `json:"default_language"` // 默认语言
	LanguageList    []dto.LanguageItem     `json:"language_list"`    // 语言列表
	Language        []string               `json:"language"`         // 可用语言列表
}
```

#### 2.3 配置结构（main/app/dto/resp/setting/store_setting.go）

```go
// Store 商城设置
type Store struct {
	Name          string             `json:"name"`           // 商城名称
	AvatarURL     string             `json:"avatar_url"`     // 默认头像
	LogoURL       string             `json:"logo_url"`       // 商城logo
	ZeroingMethod string             `json:"zeroing_method"` // 抹零方式
	IPWhiteList   string             `json:"ip_white_list"`  // ip白名单
	TimeZone      string             `json:"time_zone"`      // 时区
	NoClearTable  string             `json:"no_clear_table"` // 结账后不清台
	TimeZoneList  []TimeZoneItem     `json:"time_zone_list"` // 时区列表
	Company       string             `json:"company"`        // 公司名称
	Address       string             `json:"address"`        // 地址
	Phone         string             `json:"phone"`          // 联系电话
	TaxNumber     string             `json:"tax_number"`     // 税号
	StoreCode     string             `json:"store_code"`     // ← 新增：店铺编码
	ChainNumber   string             `json:"chain_number"`   // 连锁编号
	Language      []dto.LanguageItem `json:"language"`       // 系统语言
	AuthLanguage  string             `json:"auth_language"`  // 授权语言
	Coordinates   string             `json:"coordinates"`    // 经纬度
	Latitude      string             `json:"-"`              // 纬度
	Longitude     string             `json:"-"`              // 经度
}
```

---

## 🔌 API 设计

### API 端点

**现有端点**（无需新增）：
- `GET /api/v1/shop/basic` - 获取商家基本信息（包含 ShopProfile）
- `POST /api/v1/shop/setting/store` - 更新店铺设置

### API 请求示例

#### 更新店铺设置（包含店铺编码）

**请求**：
```http
POST /api/v1/shop/setting/store
Content-Type: application/json
Authorization: Bearer {token}

{
  "name": "示例餐厅",
  "logo_url": "/uploads/logo.png",
  "time_zone": "Asia/Shanghai",
  "company_name": "示例餐饮有限公司",
  "address": "北京市朝阳区XX路XX号",
  "phone": "010-12345678",
  "tax_number": "91110000XXXXXXXX",
  "store_code": "STORE-BJ-001",
  "language": [
    {"name": "简体中文", "value": "zh_CN"}
  ],
  "coordinates": "39.9042,116.4074"
}
```

**响应**：
```json
{
  "code": 200,
  "msg": "保存成功",
  "data": null
}
```

#### 获取商家基本信息（包含店铺编码）

**请求**：
```http
GET /api/v1/shop/basic
Authorization: Bearer {token}
```

**响应**（部分字段）：
```json
{
  "code": 200,
  "msg": "success",
  "data": {
    "profile": {
      "company_name": "示例餐饮有限公司",
      "address": "北京市朝阳区XX路XX号",
      "phone": "010-12345678",
      "tax_number": "91110000XXXXXXXX",
      "store_code": "STORE-BJ-001",
      "coordinates": "39.9042,116.4074",
      ...
    },
    ...
  }
}
```

---

## 🔧 核心服务逻辑

### 1. EditStoreSetting 方法修改

**文件**: `main/app/service/setting/setting.go`

**修改点**：

```go
func (s *Srv) EditStoreSetting(ctx context.Context, storeSettingReq req.UpdateStoreSetting) error {
	// ... 现有代码 ...
	
	// 复制请求数据到 storeSetting
	copier.CopyWithOption(&storeSetting, storeSettingReq, copier.Option{IgnoreEmpty: true})
	
	storeSetting.LogoURL = storeSettingReq.LogoUrl
	storeSetting.Company = storeSettingReq.CompanyName
	storeSetting.StoreCode = storeSettingReq.StoreCode // ← 新增：设置店铺编码
	
	// ... 保存到数据库的逻辑 ...
	
	return nil
}
```

**关键点**：
- 使用 `copier.CopyWithOption` 自动复制同名字段
- 显式设置 `StoreCode` 字段确保正确映射
- 不需要额外的验证逻辑（Gin 框架自动验证 `binding` 标签）

### 2. GetStoreSetting 方法

**无需修改**，因为：
- `GetStoreSetting` 从数据库读取 JSON 并反序列化到 `setting.Store` 结构体
- 只要 `setting.Store` 结构体中有 `StoreCode` 字段，JSON 解析会自动填充
- 如果 JSON 中没有 `store_code`，字段值为空字符串（Go 默认零值）

### 3. 登录接口返回 ShopProfile

**文件**: `main/app/service/auth.go`

**修改点**：

```go
Profile: resp.ShopProfile{
	Address:         storeSetting.Address,
	Coordinates:     storeSetting.Coordinates,
	IpWhiteList:     storeSetting.IPWhiteList,
	Phone:           storeSetting.Phone,
	TaxNumber:       storeSetting.TaxNumber,
	StoreCode:       storeSetting.StoreCode,        // ← 新增：店铺编码
	TimeZoneList:    storeSetting.TimeZoneList,
	DefaultLanguage: storeSetting.Language[0].Name,
	LanguageList:    storeSetting.Language,
	Language:        companySetting.GetLanguages(),
	CompanyName:     storeSetting.Company,
},
```

---

## 🖨️ 打印模块集成

### 打印流程

1. 打印发票时，从 `store_setting` 中读取 `store_code`
2. 如果 `store_code` 不为空，在发票模板中添加 "SHOP: {store_code}"
3. 渲染并输出发票

### 代码位置（待确认）

**文件**: `main/app/modules/printer/invoice.go`（示例路径，需根据实际代码确定）

**伪代码**：

```go
func (p *PrinterService) PrintInvoice(ctx context.Context, order *model.SaleOrder) error {
	// 获取店铺设置
	storeSetting, err := p.settingSrv.GetStoreSetting(ctx)
	if err != nil {
		return err
	}
	
	// 构建打印数据
	printData := map[string]interface{}{
		"company_name": storeSetting.Company,
		"company_addr": storeSetting.Address,
		"company_phone": storeSetting.Phone,
		"tax_number": storeSetting.TaxNumber,
		"store_code": storeSetting.StoreCode, // ← 新增
		"order_info": order,
		// ... 其他字段
	}
	
	// 渲染模板
	invoice := p.renderTemplate("invoice", printData)
	
	// 发送到打印机
	return p.print(invoice)
}
```

### 打印模板调整

**模板文件**（示例，需根据实际确定）：

```
公司名称: {{.company_name}}
公司地址: {{.company_addr}}
公司电话: {{.company_phone}}
公司税号: {{.tax_number}}
{{if .store_code}}SHOP: {{.store_code}}{{end}}

==========================================
订单信息...
```

**注意事项**：
- 使用条件判断 `{{if .store_code}}`，只在有值时显示
- 确保模板布局美观，不影响其他内容

---

## 🎨 前端设计（Flutter）

### 页面位置

**路径**: `ttpos-flutter/lib/shop/pages/profile/merchant_profile_page.dart`（示例路径）

### UI 布局

```dart
// 在税号字段下方新增店铺编码字段
Column(
  children: [
    // ... 现有字段 ...
    
    // 税号
    TextFormField(
      controller: _taxNumberController,
      decoration: InputDecoration(labelText: '税号'),
      textAlign: TextAlign.right,
    ),
    
    SizedBox(height: 16),
    
    // 店铺编码（新增）
    TextFormField(
      controller: _storeCodeController,
      decoration: InputDecoration(
        labelText: '店铺编码',
        hintText: '最多100个字符',
      ),
      textAlign: TextAlign.right, // 右对齐
      maxLength: 100,
      validator: (value) {
        if (value != null && value.length > 100) {
          return '店铺编码不能超过100个字符';
        }
        return null;
      },
    ),
    
    // ... 其他字段 ...
  ],
)
```

### 数据交互

```dart
// 加载数据
void _loadProfile() async {
  final profile = await ApiService.getShopBasic();
  setState(() {
    _storeCodeController.text = profile.profile.storeCode ?? '';
  });
}

// 保存数据
void _saveProfile() async {
  final req = UpdateStoreSettingReq(
    name: _nameController.text,
    // ... 其他字段 ...
    taxNumber: _taxNumberController.text,
    storeCode: _storeCodeController.text, // 新增
  );
  
  await ApiService.updateStoreSetting(req);
  showSuccessToast('保存成功');
}
```

---

## 🧪 测试方案

### 1. 单元测试

#### 后端 Go 单元测试

**文件**: `main/app/service/setting/setting_test.go`

```go
func TestEditStoreSetting_WithStoreCode(t *testing.T) {
	// 测试保存店铺编码
	req := req.UpdateStoreSetting{
		// ... 必填字段 ...
		StoreCode: "STORE-001",
	}
	
	err := settingSrv.EditStoreSetting(ctx, req)
	assert.NoError(t, err)
	
	// 验证保存成功
	storeSetting, _ := settingSrv.GetStoreSetting(ctx)
	assert.Equal(t, "STORE-001", storeSetting.StoreCode)
}

func TestEditStoreSetting_StoreCodeTooLong(t *testing.T) {
	// 测试超长店铺编码
	req := req.UpdateStoreSetting{
		// ... 必填字段 ...
		StoreCode: strings.Repeat("A", 101), // 101 字符
	}
	
	// 应该在 Gin 验证层被拦截
	err := validateStruct(req)
	assert.Error(t, err)
}

func TestEditStoreSetting_EmptyStoreCode(t *testing.T) {
	// 测试空店铺编码
	req := req.UpdateStoreSetting{
		// ... 必填字段 ...
		StoreCode: "",
	}
	
	err := settingSrv.EditStoreSetting(ctx, req)
	assert.NoError(t, err) // 允许为空
}
```

### 2. 集成测试

#### API 测试

**测试用例**：

| 测试用例                     | 请求参数                       | 预期结果                     |
| ---------------------------- | ------------------------------ | ---------------------------- |
| 保存正常店铺编码             | `store_code: "STORE-001"`      | 200 OK，保存成功             |
| 保存空店铺编码               | `store_code: ""`               | 200 OK，允许为空             |
| 保存 100 字符店铺编码        | `store_code: "{100个字符}"`    | 200 OK，边界值正常           |
| 保存 101 字符店铺编码        | `store_code: "{101个字符}"`    | 400 Bad Request，参数错误    |
| 保存包含中文的店铺编码       | `store_code: "门店-北京-001"`  | 200 OK，支持中文             |
| 保存包含特殊字符的店铺编码   | `store_code: "STORE#001-ABC!"` | 200 OK，支持特殊字符         |
| 获取商家信息（有店铺编码）   | -                              | 响应包含 `store_code` 字段   |
| 获取商家信息（无店铺编码）   | -                              | `store_code` 为空字符串      |

### 3. 前端测试

#### Flutter Widget 测试

**测试用例**：

| 测试用例                     | 操作                           | 预期结果                     |
| ---------------------------- | ------------------------------ | ---------------------------- |
| 页面加载显示店铺编码         | 打开商家档案页面               | 店铺编码输入框显示正确值     |
| 输入正常店铺编码             | 输入 "STORE-001" 并保存        | 保存成功，刷新后显示正确     |
| 输入超长店铺编码             | 输入 101 个字符                | 前端显示错误提示             |
| 清空店铺编码                 | 删除店铺编码并保存             | 保存成功，刷新后为空         |

### 4. 打印测试

#### 发票打印测试

**测试用例**：

| 测试用例                     | 店铺编码值                     | 预期打印结果                 |
| ---------------------------- | ------------------------------ | ---------------------------- |
| 打印发票（有店铺编码）       | "STORE-001"                    | 显示 "SHOP: STORE-001"       |
| 打印发票（无店铺编码）       | ""                             | 不显示 "SHOP:" 字段          |
| 打印发票（中文店铺编码）     | "门店-北京-001"                | 正确显示中文字符             |
| 打印发票（超长店铺编码）     | 101 个字符                     | 截断到 100 字符显示          |

---

## 📊 Story Point 评估

### 复杂度分析

| 维度               | 评分 | 说明                                                         |
| ------------------ | ---- | ------------------------------------------------------------ |
| **业务复杂度**     | 1    | 简单的字段新增，业务逻辑清晰                                 |
| **技术复杂度**     | 2    | 涉及前后端、打印模块，但都是常规操作                         |
| **工作量**         | 2    | 需要修改多个文件，但每个修改都较简单                         |
| **风险**           | 1    | 低风险，不影响现有功能                                       |
| **不确定性**       | 1    | 需求清晰，实现方案明确                                       |

### 工作量估算

| 任务                         | 预估工时 | 说明                                       |
| ---------------------------- | -------- | ------------------------------------------ |
| 后端数据结构修改             | 0.5h     | 新增字段到 DTO 和配置结构                  |
| 后端服务逻辑修改             | 0.5h     | 修改 EditStoreSetting 和登录接口           |
| 后端单元测试                 | 1h       | 编写测试用例                               |
| 前端 UI 实现                 | 1h       | 新增输入框、验证逻辑、数据绑定             |
| 前端测试                     | 0.5h     | Widget 测试                                |
| 打印模块集成                 | 1h       | 读取字段、渲染模板（需确认实际代码）       |
| 打印测试                     | 0.5h     | 测试各种场景的打印效果                     |
| 集成测试                     | 1h       | API 测试、端到端测试                       |
| 文档更新                     | 0.5h     | API 文档、用户手册                         |
| **总计**                     | **6.5h** | **约 1 个工作日**                          |

### Story Point 建议

**推荐 SP**: **2**

**理由**：
- 需求简单清晰，实现方案明确
- 涉及前后端联调，但都是常规操作
- 工作量约 1 个工作日
- 低风险，不影响现有功能

---

## 🚀 实施计划

### Phase 1: 后端实现（2-3h）

1. 修改数据结构（DTO、配置）
2. 修改服务逻辑（EditStoreSetting、登录接口）
3. 编写单元测试
4. API 测试

### Phase 2: 前端实现（1.5-2h）

1. 新增店铺编码输入框
2. 数据绑定和验证
3. Widget 测试
4. UI 测试

### Phase 3: 打印模块集成（1-1.5h）

1. 读取店铺编码
2. 渲染打印模板
3. 打印测试

### Phase 4: 集成测试与文档（1-1.5h）

1. 端到端测试
2. 更新 API 文档
3. 更新用户手册（如有）

---

## 🔒 安全考虑

### 输入验证

- **长度限制**: 最大 100 字符（后端 Gin 验证 + 前端验证）
- **XSS 防护**: 输出时进行 HTML 转义（打印模板中）
- **SQL 注入**: 不涉及（数据存储在 JSON 字段中）

### 权限控制

- **查看权限**: 商家管理员
- **修改权限**: 商家管理员
- **API 鉴权**: 必须通过 JWT 认证

---

## 📝 文档更新

### 需要更新的文档

- [ ] API 文档：`docs/shared/api/shop-management.md`
- [ ] 用户手册：商家档案配置说明
- [ ] 数据字典：新增 `store_code` 字段说明

---

## 🔗 相关资源

### 参考代码

- 店铺设置 DTO: `main/app/dto/req/base.go`、`main/app/dto/resp/setting/store_setting.go`
- 设置服务: `main/app/service/setting/setting.go`
- 登录接口: `main/app/service/auth.go`
- 打印模块: `main/app/modules/printer/`（待确认）

### 参考规范

- Go Main 核心约束: `.cursor/rules/go-main.mdc`
- 打印模块开发规范: `.cursor/rules/go-printer.mdc`
- API 设计规范: `.cursor/rules/api.mdc`

---

**版本**: v1.0.0  
**创建日期**: 2025-12-05  
**最后更新**: 2025-12-05  
**设计者**: weifashi  
**审核状态**: 待审核

