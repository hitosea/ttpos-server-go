# 新管理端-商家档案增加店铺编码 任务分解

> 本文档将需求分解为可执行的任务列表，用于开发跟踪。

## 📋 基本信息

| 项目              | 内容                                                                                                                       |
| ----------------- | -------------------------------------------------------------------------------------------------------------------------- |
| **关联 Requirements** | [requirements.md](./requirements.md)                                                                                    |
| **关联 Design**   | [design.md](./design.md)                                                                                                   |
| **创建日期**      | 2025-12-05                                                                                                                 |
| **负责人**        | weifashi                                                                                                                   |
| **Story Point**   | 2                                                                                                                          |
| **目标版本**      | v2.11.0                                                                                                                   |

---

## ✅ 任务总览

| 阶段             | 任务数 | 预估工时 | 状态   |
| ---------------- | ------ | -------- | ------ |
| **后端实现**     | 6      | 2-3h     | ✅ 完成 |
| **前端实现**     | 4      | 1.5-2h   | ❌ 取消（不需要） |
| **打印模块集成** | 3      | 1-1.5h   | ❌ 取消（不需要） |
| **测试与文档**   | 2      | 0.5h     | ⏳ 进行中 |
| **总计**         | **8**  | **2.5-3.5h** | **75%** |

---

## 📋 Phase 1: 后端实现（Go Main）

### Task 1.1: 修改请求 DTO 结构 ✅

**优先级**: P0  
**预估工时**: 10 分钟  
**实际工时**: 5 分钟  
**负责人**: weifashi  
**状态**: ✅ 已完成（2025-12-05）

**描述**: 在 `UpdateStoreSetting` 结构体中新增 `StoreCode` 字段。

**文件**: `main/app/dto/req/base.go`

**具体修改**:

```go
type UpdateStoreSetting struct {
	Name        string             `json:"name" binding:"required,max=100"`
	LogoUrl     string             `json:"logo_url" binding:"required"`
	TimeZone    string             `json:"time_zone" binding:"required"`
	CompanyName string             `json:"company_name" binding:"max=500"`
	Address     string             `json:"address" binding:"max=500"`
	Phone       string             `json:"phone" binding:"required,max=20"`
	TaxNumber   string             `json:"tax_number"`
	StoreCode   string             `json:"store_code" binding:"max=100"` // ← 新增
	Language    []dto.LanguageItem `json:"language" binding:"required,min=1"`
	Coordinates string             `json:"coordinates"`
}
```

**验收标准**:
- [x] `StoreCode` 字段定义正确
- [x] 验证规则 `max=100` 生效
- [x] JSON 标签为 `store_code`

---

### Task 1.2: 修改响应 DTO 结构 ✅

**优先级**: P0  
**预估工时**: 10 分钟  
**实际工时**: 5 分钟  
**负责人**: weifashi  
**状态**: ✅ 已完成（2025-12-05）

**描述**: 在 `ShopProfile` 结构体中新增 `StoreCode` 字段。

**文件**: `main/app/dto/resp/base.go`

**具体修改**:

```go
type ShopProfile struct {
	CompanyName     string                 `json:"company_name"`
	Address         string                 `json:"address"`
	Coordinates     string                 `json:"coordinates"`
	IpWhiteList     string                 `json:"ip_white_list"`
	Phone           string                 `json:"phone"`
	TaxNumber       string                 `json:"tax_number"`
	StoreCode       string                 `json:"store_code"` // ← 新增
	TimeZoneList    []setting.TimeZoneItem `json:"time_zone_list"`
	DefaultLanguage string                 `json:"default_language"`
	LanguageList    []dto.LanguageItem     `json:"language_list"`
	Language        []string               `json:"language"`
}
```

**验收标准**:
- [x] `StoreCode` 字段定义正确
- [x] JSON 标签为 `store_code`

---

### Task 1.3: 修改配置结构 ✅

**优先级**: P0  
**预估工时**: 10 分钟  
**实际工时**: 5 分钟  
**负责人**: weifashi  
**状态**: ✅ 已完成（2025-12-05）

**描述**: 在 `setting.Store` 结构体中新增 `StoreCode` 字段。

**文件**: `main/app/dto/resp/setting/store_setting.go`

**具体修改**:

```go
type Store struct {
	Name          string             `json:"name"`
	AvatarURL     string             `json:"avatar_url"`
	LogoURL       string             `json:"logo_url"`
	ZeroingMethod string             `json:"zeroing_method"`
	IPWhiteList   string             `json:"ip_white_list"`
	TimeZone      string             `json:"time_zone"`
	NoClearTable  string             `json:"no_clear_table"`
	TimeZoneList  []TimeZoneItem     `json:"time_zone_list"`
	Company       string             `json:"company"`
	Address       string             `json:"address"`
	Phone         string             `json:"phone"`
	TaxNumber     string             `json:"tax_number"`
	StoreCode     string             `json:"store_code"` // ← 新增
	ChainNumber   string             `json:"chain_number"`
	Language      []dto.LanguageItem `json:"language"`
	AuthLanguage  string             `json:"auth_language"`
	Coordinates   string             `json:"coordinates"`
	Latitude      string             `json:"-"`
	Longitude     string             `json:"-"`
}
```

**验收标准**:
- [x] `StoreCode` 字段定义正确
- [x] JSON 标签为 `store_code`

---

### Task 1.4: 修改 EditStoreSetting 服务逻辑 ✅

**优先级**: P0  
**预估工时**: 20 分钟  
**实际工时**: 10 分钟  
**负责人**: weifashi  
**状态**: ✅ 已完成（2025-12-05）

**描述**: 在 `EditStoreSetting` 方法中处理 `StoreCode` 字段的保存。

**文件**: `main/app/service/setting/setting.go`

**具体修改**:

在 `EditStoreSetting` 方法中，找到以下代码：

```go
copier.CopyWithOption(&storeSetting, storeSettingReq, copier.Option{IgnoreEmpty: true})

storeSetting.LogoURL = storeSettingReq.LogoUrl
storeSetting.Company = storeSettingReq.CompanyName
```

修改为：

```go
copier.CopyWithOption(&storeSetting, storeSettingReq, copier.Option{IgnoreEmpty: true})

storeSetting.LogoURL = storeSettingReq.LogoUrl
storeSetting.Company = storeSettingReq.CompanyName
storeSetting.StoreCode = storeSettingReq.StoreCode // ← 新增
```

**验收标准**:
- [x] `StoreCode` 字段正确赋值
- [x] 保存到数据库后能正确读取
- [x] 空值允许保存

---

### Task 1.5: 修改登录接口返回 ShopProfile ✅

**优先级**: P0  
**预估工时**: 10 分钟  
**实际工时**: 5 分钟  
**负责人**: weifashi  
**状态**: ✅ 已完成（2025-12-05）

**描述**: 在登录接口中返回 `StoreCode` 字段。

**文件**: `main/app/service/auth.go`

**具体修改**:

在 `ShopBase` 返回的 `Profile` 中添加 `StoreCode`：

```go
Profile: resp.ShopProfile{
	Address:         storeSetting.Address,
	Coordinates:     storeSetting.Coordinates,
	IpWhiteList:     storeSetting.IPWhiteList,
	Phone:           storeSetting.Phone,
	TaxNumber:       storeSetting.TaxNumber,
	StoreCode:       storeSetting.StoreCode, // ← 新增
	TimeZoneList:    storeSetting.TimeZoneList,
	DefaultLanguage: storeSetting.Language[0].Name,
	LanguageList:    storeSetting.Language,
	Language:        companySetting.GetLanguages(),
	CompanyName:     storeSetting.Company,
},
```

**验收标准**:
- [x] 登录接口响应中包含 `store_code` 字段
- [x] 字段值正确

---

### Task 1.6: 编写后端单元测试 ✅

**优先级**: P1  
**预估工时**: 1 小时  
**实际工时**: 45 分钟  
**负责人**: weifashi  
**状态**: ✅ 已完成（2025-12-05）

**描述**: 为店铺编码字段编写单元测试。

**文件**: `main/app/service/setting/store_code_test.go`（已创建）

**测试用例**:
1. ✅ 测试字段定义正确性
2. ✅ 测试保存空店铺编码
3. ✅ 测试保存 100 字符店铺编码（边界值）
4. ✅ 测试包含中文的店铺编码
5. ✅ 测试包含特殊字符的店铺编码

**测试结果**:
```
PASS: TestStoreCodeFieldExists
PASS: TestStoreCodeEmpty
PASS: TestStoreCodeMaxLength
PASS: TestStoreCodeWithChinese
PASS: TestStoreCodeWithSpecialChars
ok  	ttpos-server-go/app/service/setting	0.032s
```

**验收标准**:
- [x] 所有测试用例通过（5/5）
- [x] 测试覆盖关键场景

---

## ❌ Phase 2: 前端实现（已取消）

> **注意**: 根据项目需求调整，前端实现部分已取消，后端 API 已就绪，如需前端支持可随时对接。

### Task 2.1: 在商家档案页面新增店铺编码输入框 ❌

**优先级**: P0  
**预估工时**: 30 分钟  
**负责人**: 前端开发  
**状态**: ❌ 已取消（不需要）

**描述**: 在商家档案配置页面的税号下方新增店铺编码输入框。

**文件**: `ttpos-flutter/lib/shop/pages/profile/merchant_profile_page.dart`（示例路径）

**具体修改**:

在税号字段下方添加：

```dart
// 店铺编码
TextFormField(
  controller: _storeCodeController,
  decoration: InputDecoration(
    labelText: '店铺编码',
    hintText: '最多100个字符',
    helperText: '用于发票打印显示',
  ),
  textAlign: TextAlign.right,
  maxLength: 100,
  validator: (value) {
    if (value != null && value.length > 100) {
      return '店铺编码不能超过100个字符';
    }
    return null;
  },
),
```

**验收标准**:
- [ ] ~~输入框位于税号下方~~（已取消）
- [ ] ~~右对齐显示~~（已取消）
- [ ] ~~最大长度 100 字符~~（已取消）
- [ ] ~~超长时显示错误提示~~（已取消）

---

### Task 2.2: 绑定店铺编码数据 ❌

**优先级**: P0  
**预估工时**: 20 分钟  
**负责人**: 前端开发  
**状态**: ❌ 已取消（不需要）

**描述**: 在页面加载和保存时，正确读取和提交店铺编码数据。

**文件**: 同 Task 2.1

**具体修改**:

**加载数据**:
```dart
void _loadProfile() async {
  final response = await ApiService.getShopBasic();
  setState(() {
    // ... 其他字段 ...
    _storeCodeController.text = response.data.profile.storeCode ?? '';
  });
}
```

**保存数据**:
```dart
void _saveProfile() async {
  final req = UpdateStoreSettingReq(
    name: _nameController.text,
    // ... 其他字段 ...
    taxNumber: _taxNumberController.text,
    storeCode: _storeCodeController.text,
  );
  
  await ApiService.updateStoreSetting(req);
  showSuccessToast('保存成功');
}
```

**验收标准**:
- [ ] ~~页面加载时正确显示店铺编码~~（已取消）
- [ ] ~~保存时正确提交店铺编码~~（已取消）
- [ ] ~~空值允许保存~~（已取消）

---

### Task 2.3: 更新 API 请求和响应模型 ❌

**优先级**: P0  
**预估工时**: 20 分钟  
**负责人**: 前端开发  
**状态**: ❌ 已取消（不需要）

**描述**: 在 Flutter 的 API 模型中新增 `storeCode` 字段。

**文件**: 
- `ttpos-flutter/lib/shop/models/shop_profile.dart`（响应模型）
- `ttpos-flutter/lib/shop/models/update_store_setting_req.dart`（请求模型）

**具体修改**:

**响应模型**:
```dart
class ShopProfile {
  final String companyName;
  final String address;
  final String phone;
  final String taxNumber;
  final String? storeCode; // ← 新增
  // ... 其他字段 ...
  
  ShopProfile.fromJson(Map<String, dynamic> json)
      : companyName = json['company_name'] ?? '',
        address = json['address'] ?? '',
        phone = json['phone'] ?? '',
        taxNumber = json['tax_number'] ?? '',
        storeCode = json['store_code'], // ← 新增
        // ... 其他字段 ...
}
```

**请求模型**:
```dart
class UpdateStoreSettingReq {
  final String name;
  // ... 其他字段 ...
  final String? taxNumber;
  final String? storeCode; // ← 新增
  
  Map<String, dynamic> toJson() {
    return {
      'name': name,
      // ... 其他字段 ...
      'tax_number': taxNumber,
      'store_code': storeCode, // ← 新增
    };
  }
}
```

**验收标准**:
- [ ] ~~响应模型正确解析 `store_code`~~（已取消）
- [ ] ~~请求模型正确序列化 `store_code`~~（已取消）

---

### Task 2.4: 前端 Widget 测试 ❌

**优先级**: P1  
**预估工时**: 30 分钟  
**负责人**: 前端开发  
**状态**: ❌ 已取消（不需要）

**描述**: 编写商家档案页面的 Widget 测试。

**文件**: `ttpos-flutter/test/shop/pages/profile/merchant_profile_page_test.dart`

**测试用例**:
1. 页面加载时显示店铺编码
2. 输入正常店铺编码并保存
3. 输入超长店铺编码显示错误
4. 清空店铺编码并保存

**验收标准**:
- [ ] ~~所有测试用例通过~~（已取消）
- [ ] ~~测试覆盖关键交互逻辑~~（已取消）

---

## ❌ Phase 3: 打印模块集成（已取消）

> **注意**: 根据项目需求调整，打印模块集成部分已取消。如需在发票打印中显示店铺编码，后端数据已就绪，可随时对接。

### Task 3.1: 在打印服务中读取店铺编码 ❌

**优先级**: P0  
**预估工时**: 20 分钟  
**负责人**: weifashi  
**状态**: ❌ 已取消（不需要）

**描述**: 在发票打印逻辑中读取 `store_code`。

**文件**: `main/app/modules/printer/invoice.go`（待确认实际路径）

**具体修改**:

在打印发票的方法中，获取店铺设置：

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
	
	// 渲染并打印
	return p.renderAndPrint("invoice", printData)
}
```

**验收标准**:
- [ ] ~~能正确读取 `store_code`~~（已取消）
- [ ] ~~传递给打印模板~~（已取消）

---

### Task 3.2: 修改发票打印模板 ❌

**优先级**: P0  
**预估工时**: 20 分钟  
**负责人**: weifashi  
**状态**: ❌ 已取消（不需要）

**描述**: 在发票模板中添加店铺编号显示。

**文件**: 打印模板文件（待确认实际路径）

**具体修改**:

在公司信息区域添加：

```
公司名称: {{.company_name}}
公司地址: {{.company_addr}}
公司电话: {{.company_phone}}
公司税号: {{.tax_number}}
{{if .store_code}}SHOP: {{.store_code}}{{end}}
```

**验收标准**:
- [ ] ~~有店铺编码时显示 "SHOP: {code}"~~（已取消）
- [ ] ~~无店铺编码时不显示该行~~（已取消）
- [ ] ~~布局美观，不影响其他内容~~（已取消）

---

### Task 3.3: 打印功能测试 ❌

**优先级**: P0  
**预估工时**: 30 分钟  
**负责人**: weifashi  
**状态**: ❌ 已取消（不需要）

**描述**: 测试发票打印功能。

**测试用例**:
1. ~~有店铺编码时打印发票~~（已取消）
2. ~~无店铺编码时打印发票~~（已取消）
3. ~~包含中文的店铺编码打印~~（已取消）
4. ~~包含特殊字符的店铺编码打印~~（已取消）

**验收标准**:
- [ ] ~~所有测试场景打印正确~~（已取消）
- [ ] ~~布局美观，无乱码~~（已取消）

---

## 📋 Phase 4: 测试与文档

### Task 4.1: API 集成测试 ❌

**优先级**: P0  
**预估工时**: 30 分钟  
**负责人**: 测试工程师  
**状态**: ❌ 已取消（前端不需要，暂不测试）

**描述**: 端到端测试 API 功能。

**备注**: 由于前端和打印模块已取消，暂不进行端到端测试。后端单元测试已完成并通过。

---

### Task 4.2: 端到端测试 ❌

**优先级**: P1  
**预估工时**: 30 分钟  
**负责人**: 测试工程师  
**状态**: ❌ 已取消（前端不需要）

**描述**: 从前端到后端的完整流程测试。

**备注**: 由于前端实现已取消，暂不进行端到端测试。

---

### Task 4.3: 更新 API 文档 ⏳

**优先级**: P1  
**预估工时**: 20 分钟  
**负责人**: weifashi  
**状态**: ⏳ 待完成

**描述**: 更新 API 文档，说明新增的 `store_code` 字段。

**文件**: `docs/shared/api/shop-management.md`

**更新内容**:
- 更新 `UpdateStoreSetting` 请求参数说明
- 更新 `ShopProfile` 响应字段说明
- 新增字段示例

**验收标准**:
- [ ] 文档准确描述新字段
- [ ] 包含示例代码

---

### Task 4.4: 更新用户手册 ❌

**优先级**: P2  
**预估工时**: 10 分钟  
**负责人**: 产品经理  
**状态**: ❌ 已取消（前端不需要）

**描述**: 更新用户手册，说明店铺编码的用途和配置方法。

**备注**: 由于前端实现已取消，用户手册暂不更新。

---

## 📊 进度追踪

### 开发进度

| Phase            | 进度    | 完成时间       | 备注               |
| ---------------- | ------- | -------------- | ------------------ |
| Phase 1: 后端    | ✅ 100% | 2025-12-05 14:06 | 全部任务已完成     |
| Phase 2: 前端    | ❌ 取消 | -              | 不需要             |
| Phase 3: 打印    | ❌ 取消 | -              | 不需要             |
| Phase 4: 测试    | 75%     | -              | API文档待更新      |

**总体进度**: 85% （6/7 个需要的任务已完成）

### 任务状态

| 任务ID  | 任务名称                     | 状态         | 负责人     | 完成时间       |
| ------- | ---------------------------- | ------------ | ---------- | -------------- |
| Task1.1 | 修改请求 DTO                 | ✅ 已完成    | weifashi   | 2025-12-05     |
| Task1.2 | 修改响应 DTO                 | ✅ 已完成    | weifashi   | 2025-12-05     |
| Task1.3 | 修改配置结构                 | ✅ 已完成    | weifashi   | 2025-12-05     |
| Task1.4 | 修改服务逻辑                 | ✅ 已完成    | weifashi   | 2025-12-05     |
| Task1.5 | 修改登录接口                 | ✅ 已完成    | weifashi   | 2025-12-05     |
| Task1.6 | 后端单元测试                 | ✅ 已完成    | weifashi   | 2025-12-05     |
| Task2.1 | 前端新增输入框               | ❌ 已取消    | -          | -              |
| Task2.2 | 前端数据绑定                 | ❌ 已取消    | -          | -              |
| Task2.3 | 前端 API 模型                | ❌ 已取消    | -          | -              |
| Task2.4 | 前端 Widget 测试             | ❌ 已取消    | -          | -              |
| Task3.1 | 打印服务读取                 | ❌ 已取消    | -          | -              |
| Task3.2 | 打印模板修改                 | ❌ 已取消    | -          | -              |
| Task3.3 | 打印功能测试                 | ❌ 已取消    | -          | -              |
| Task4.1 | API 集成测试                 | ❌ 已取消    | -          | -              |
| Task4.2 | 端到端测试                   | ❌ 已取消    | -          | -              |
| Task4.3 | 更新 API 文档                | ⏳ 待完成    | weifashi   | -              |
| Task4.4 | 更新用户手册                 | ❌ 已取消    | -          | -              |

---

## 🔗 相关资源

### 文档链接

- [Requirements](./requirements.md)
- [Design](./design.md)
- [Proposal](../../../../team/proposals/2025-12/v2.111.0-shop-merchant-profile-store-code.md)
- [DooTask #37456](https://dootask.{domain}/project/368/task/37456)

### 开发规范

- [Go Main 核心约束](.cursor/rules/go-main.mdc)
- [打印模块开发规范](.cursor/rules/go-printer.mdc)
- [API 设计规范](.cursor/rules/api.mdc)

---

## ⚠️ 注意事项

### 开发注意事项

1. **数据兼容性**：✅ 已确保空值处理正确，不影响现有数据
2. ~~**打印模块**：需要确认实际的打印模块代码路径和模板位置~~（已取消）
3. ~~**前端路径**：Flutter 前端文件路径为示例，需根据实际项目确定~~（已取消）
4. **多语言**：✅ 单元测试已验证中英文、特殊字符正确处理
5. **缓存清理**：✅ 修改设置后会自动清除相关缓存（在 EditStoreSetting 中实现）

### 测试注意事项

1. ✅ 已测试各种边界情况（空值、边界值、100字符）
2. ✅ 已测试中英文、特殊字符的正确处理
3. ~~测试打印效果（实际打印机输出）~~（已取消）
4. **向后兼容性**：✅ 旧版本客户端不受影响（新字段为可选）

### 后端 API 就绪

**已完成的 API 端点**：
- ✅ `POST /api/v1/shop/setting/store` - 支持 `store_code` 参数
- ✅ `GET /api/v1/shop/basic` - 响应包含 `profile.store_code`

**数据格式**：
```json
// 请求
{
  "store_code": "STORE-001"  // 可选，最大100字符
}

// 响应
{
  "data": {
    "profile": {
      "store_code": "STORE-001"
    }
  }
}
```

---

## 📝 开发日志

### 2025-12-05

- ✅ 创建 Spec 文档（requirements.md、design.md、tasks.md）
- ✅ Phase 1 后端实现完成
  - ✅ 修改 3 个 DTO 文件
  - ✅ 修改 2 个服务文件
  - ✅ 编写 5 个单元测试（全部通过）
  - ✅ Lint 检查无错误
  - ✅ 编译检查通过
- ⏳ 取消 Phase 2（前端）和 Phase 3（打印）
- ⏳ Phase 4 部分完成（API 文档待更新）

**实际工时**: 约 1.75 小时（比预估 2.5 小时更快）

---

## 🎯 后续建议

### 如果需要前端支持

后端 API 已就绪，如果未来需要前端实现，可以直接对接：

1. **前端调用 API**：
   - 获取店铺编码：调用 `GET /api/v1/shop/basic`，读取 `profile.store_code`
   - 保存店铺编码：调用 `POST /api/v1/shop/setting/store`，传递 `store_code` 参数

2. **UI 实现建议**：
   - 在商家档案页面税号下方添加"店铺编码"输入框
   - 右对齐显示，最大 100 字符
   - 支持空值

### 如果需要打印模块支持

后端数据已就绪，如果需要在发票打印中显示店铺编码：

1. **数据来源**：从 `storeSetting.StoreCode` 读取
2. **显示格式**：建议显示为 "SHOP: {store_code}"
3. **模板修改**：在发票模板的公司信息区域添加店铺编码

---

**版本**: v1.0.0  
**创建日期**: 2025-12-05  
**最后更新**: 2025-12-05  
**维护者**: weifashi  
**状态**: 待开始

