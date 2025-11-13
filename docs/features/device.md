# 设备管理服务 (Device Service)

## 概述

`device.go` 实现了设备管理服务，负责管理餐饮系统中各类终端设备的绑定、认证和配置。该服务支持多种设备类型（收银机、点餐助手、平板、厨显、商家后台），提供设备绑定上限控制、自动打印机配置、设备信息管理等功能，是多终端管理的核心模块。

**文件路径**: `ttpos-server-go/main/app/service/device.go`

## 核心功能

### 1. 设备绑定管理
- 添加/更新设备绑定
- 软删除设备的恢复
- 设备绑定数量限制
- 主设备自动标记

### 2. 设备类型支持
- 收银机 (Cashier)
- 点餐助手 (Assistant)
- 平板点餐 (Tablet)
- 厨房显示 (Kitchen)
- 商家后台 (Shop)

### 3. 自动化配置
- 收银机自带打印机自动绑定
- 设备平台自动识别
- 版本号自动更新

### 4. 设备信息管理
- 设备备注管理
- 设备绑定状态查询
- 最后登录信息记录

## 接口定义

### IDeviceSrv 接口

```go
type IDeviceSrv interface {
    AddDevice(ctx context.Context, addReq req.AddDeviceReq) (uint64, error)
    GetRemark(companyUuid uint64, source string, deviceId string) string
    IsDeviceBind(ctx context.Context, companyUuid uint64, source string, deviceId string) bool
    UpdateRemark(ctx context.Context, editSettingReq req.EditDeviceRemarkReq) error
}
```

### deviceSrv 结构体

```go
type deviceSrv struct {
    settingSrv setting.ISrv          // 设置服务
    dbm        *database.DBManager   // 数据库管理器
}
```

## 依赖项

### 内部依赖
- **repository.DeviceRepo**: 设备数据仓库
- **repository.CompanySettingRepo**: 商家配置仓库
- **setting.ISrv**: 设置服务，用于打印机配置

### 外部依赖
- **database.DBManager**: 数据库管理器
- **utils**: 工具包，提供平台识别等功能

## 支持的设备类型

| 设备类型 | 常量 | 说明 | 绑定上限字段 |
|---------|------|------|------------|
| 收银机 | SourceCashier | 收银端设备 | CashLimit |
| 点餐助手 | SourceAssistant | 服务员点餐设备 | AssistantLimit |
| 平板点餐 | SourceTablet | 顾客自助点餐平板 | TabletLimit |
| 厨房显示 | SourceKitchen | 后厨显示屏（KDS） | KitchenLimit |
| 商家后台 | SourceShop | 商家管理后台 | （无限制） |

## 核心方法详解

### 1. AddDevice - 添加/更新设备绑定

**方法签名**:
```go
func (s *deviceSrv) AddDevice(ctx context.Context, addReq req.AddDeviceReq) (uint64, error)
```

**功能**: 添加新设备绑定或更新已有设备信息，是设备管理的核心方法。

**请求参数**:
```go
type AddDeviceReq struct {
    CompanyUuid        uint64  // 公司UUID
    Source             string  // 设备来源类型
    DeviceId           string  // 设备ID（唯一标识）
    Remark             string  // 设备备注
    Brand              string  // 设备品牌
    FinallyLoginUuid   uint64  // 最后登录员工UUID
    FinallyLoginTime   int     // 最后登录时间
    ProductPrinterUuid uint64  // 产品打印机UUID
    KdsMode            *uint   // 厨显模式（仅厨显设备）
}
```

**返回值**: `uint64` - 设备 UUID

**实现流程**:

```42:158:ttpos-server-go/main/app/service/device.go
func (s *deviceSrv) AddDevice(ctx context.Context, addReq req.AddDeviceReq) (uint64, error) {
	if !slices.Contains([]string{
		constant.SourceCashier,
		constant.SourceAssistant,
		constant.SourceTablet,
		constant.SourceKitchen,
		constant.SourceShop,
	}, addReq.Source) ||
		addReq.CompanyUuid == 0 || addReq.DeviceId == "" {
		return 0, errors.New("来源设备错误")
	}
	// 判断厨显模式
	if addReq.Source == constant.SourceKitchen && addReq.KdsMode != nil {
		if !slices.Contains([]uint{constant.KdsModeDefault, constant.KdsModeMake, constant.KdsModeMakeAndSend}, *addReq.KdsMode) {
			return 0, errors.New("厨显工作模式错误")
		}
	}
	// 记录 ua 和 平台
	userAgent := ctx.GetGin().GetHeader("User-Agent") + ";" + ctx.GetGin().GetHeader("platform") // 记录平台
	platform := utils.GetPlatform(userAgent)

	db := s.dbm.GetDB(addReq.CompanyUuid)
	deviceRepo := repository.NewDeviceRepo(db)
	// 判断设备绑定上限
	companySetting := repository.NewCompanySettingRepo(db).Get()
	if err := s.reachBindLimit(deviceRepo, companySetting, addReq.Source, addReq.DeviceId); err != nil {
		return 0, err
	}
	// 获取绑定
	existsDevice, _ := deviceRepo.GetDeviceAll(deviceRepo.WhereSource(addReq.Source), deviceRepo.WhereSn(addReq.DeviceId))
	if existsDevice.ID != 0 {
		productPrinterUuid := addReq.ProductPrinterUuid
		if productPrinterUuid == 0 {
			productPrinterUuid = existsDevice.ProductPrinterUuid
		}
		kdsMode := existsDevice.KdsMode
		if addReq.KdsMode != nil {
			kdsMode = *addReq.KdsMode
		}
		remark := addReq.Remark
		if remark == "" {
			remark = existsDevice.Remark
		}
		finallyLoginTime := addReq.FinallyLoginTime
		if finallyLoginTime == 0 {
			finallyLoginTime = existsDevice.FinallyLoginTime
		}
		// 更新绑定
		err := deviceRepo.UpdateDevice(existsDevice.Uuid, map[string]any{
			"delete_time":          0,
			"product_printer_uuid": productPrinterUuid,
			"remark":               remark,
			"brand":                addReq.Brand,
			"platform":             platform,
			"user_agent":           userAgent,
			"finally_login_uuid":   addReq.FinallyLoginUuid,
			"finally_login_time":   finallyLoginTime,
			"kds_mode":             kdsMode,
			"version":              ctx.GetVersion(),
			"is_main": func() int {
				if addReq.Source == constant.SourceCashier {
					if !deviceRepo.IsExistCashierMain(constant.SourceCashier) {
						return 1
					}
				}
				return existsDevice.IsMain
			}(),
		})
		if err != nil {
			return 0, errors.WithMessage(err, "更新绑定信息失败")
		}
		// 已软删除收银机重新登录，如果自带打印，默认更新收银打印配置
		if existsDevice.DeleteTime != 0 &&
			addReq.Source == constant.SourceCashier && slices.Contains(constant.BrandsPrints, addReq.Brand) {
			if err := s.bindPrinter(ctx, addReq.DeviceId); err != nil {
				return 0, errors.WithMessage(err, "设置默认打印机失败")
			}
		}
		return existsDevice.Uuid, nil
	}

	// 绑定品牌，如果自带打印，默认更新收银打印配置
	if addReq.Source == constant.SourceCashier && slices.Contains(constant.BrandsPrints, addReq.Brand) {
		if err := s.bindPrinter(ctx, addReq.DeviceId); err != nil {
			return 0, errors.WithMessage(err, "设置默认打印机失败")
		}
	}

	kdsMode := uint(0)
	if addReq.KdsMode != nil {
		kdsMode = *addReq.KdsMode
	}
	device, err := deviceRepo.CreateDevice(model.Device{
		FinallyLoginUuid: addReq.FinallyLoginUuid,
		FinallyLoginTime: addReq.FinallyLoginTime,
		Source:           addReq.Source,
		DeviceId:         addReq.DeviceId,
		Remark:           addReq.Remark,
		Brand:            addReq.Brand,
		Platform:         platform,
		UserAgent:        userAgent,
		KdsMode:          kdsMode,
		Version:          ctx.GetVersion(),
		IsMain: func() int {
			if addReq.Source == constant.SourceCashier {
				if !deviceRepo.IsExistCashierMain(constant.SourceCashier) {
					return 1
				}
			}
			return 0
		}(),
	})
	if err != nil {
		return 0, errors.WithMessage(err)
	}
	return device.Uuid, nil
}
```

**处理流程**:

#### 第一步：参数验证
```go
// 1. 验证设备来源类型
if !slices.Contains([...], addReq.Source) {
    return errors.New("来源设备错误")
}

// 2. 验证厨显模式（仅厨显设备）
if addReq.Source == SourceKitchen && addReq.KdsMode != nil {
    if !slices.Contains([三种模式], *addReq.KdsMode) {
        return errors.New("厨显工作模式错误")
    }
}
```

**厨显工作模式**:
- `KdsModeDefault` (0): 默认模式
- `KdsModeMake` (1): 制作模式
- `KdsModeMakeAndSend` (2): 制作+送出模式

#### 第二步：自动识别平台
```go
userAgent := ctx.GetGin().GetHeader("User-Agent") + ";" + ctx.GetGin().GetHeader("platform")
platform := utils.GetPlatform(userAgent)
```

**平台识别**:
- 从 HTTP 请求头提取 `User-Agent` 和 `platform`
- 自动识别设备平台（Windows、macOS、Linux、Android、iOS 等）
- 记录完整的 UserAgent 信息

#### 第三步：检查绑定上限
```go
if err := s.reachBindLimit(deviceRepo, companySetting, addReq.Source, addReq.DeviceId); err != nil {
    return 0, err
}
```

**绑定上限控制**:
- 收银机：`CashLimit`
- 点餐助手：`AssistantLimit`
- 平板点餐：`TabletLimit`
- 厨房显示：`KitchenLimit`
- 商家后台：无限制

#### 第四步：查询现有设备
```go
existsDevice, _ := deviceRepo.GetDeviceAll(
    deviceRepo.WhereSource(addReq.Source), 
    deviceRepo.WhereSn(addReq.DeviceId)
)
```

**两种处理路径**:

##### 路径A: 设备已存在 - 更新
```go
if existsDevice.ID != 0 {
    // 1. 合并字段（新值优先，无新值保留旧值）
    productPrinterUuid := addReq.ProductPrinterUuid
    if productPrinterUuid == 0 {
        productPrinterUuid = existsDevice.ProductPrinterUuid
    }
    
    // 2. 更新设备信息
    deviceRepo.UpdateDevice(existsDevice.Uuid, map[string]any{
        "delete_time": 0,  // 恢复软删除
        // ... 其他字段
    })
    
    // 3. 软删除设备恢复时，自动配置打印机
    if existsDevice.DeleteTime != 0 && 
       addReq.Source == SourceCashier && 
       slices.Contains(BrandsPrints, addReq.Brand) {
        s.bindPrinter(ctx, addReq.DeviceId)
    }
}
```

**更新逻辑特点**:
1. **字段合并**: 新值优先，无新值时保留旧值
2. **软删除恢复**: 设置 `delete_time = 0`
3. **主设备判断**: 如果是收银机且不存在主设备，设为主设备
4. **打印机自动配置**: 软删除恢复且品牌支持打印时自动配置

##### 路径B: 新设备 - 创建
```go
// 1. 自带打印机自动配置
if addReq.Source == SourceCashier && 
   slices.Contains(BrandsPrints, addReq.Brand) {
    s.bindPrinter(ctx, addReq.DeviceId)
}

// 2. 创建设备记录
device, err := deviceRepo.CreateDevice(model.Device{
    Source:   addReq.Source,
    DeviceId: addReq.DeviceId,
    // ...
    IsMain: func() int {
        if addReq.Source == SourceCashier {
            if !deviceRepo.IsExistCashierMain(SourceCashier) {
                return 1  // 第一台收银机设为主设备
            }
        }
        return 0
    }(),
})
```

**创建逻辑特点**:
1. **打印机自动配置**: 创建前配置
2. **主设备自动标记**: 第一台收银机自动设为主设备
3. **版本号记录**: 记录客户端版本号

---

### 2. reachBindLimit - 检查绑定上限（私有方法）

**方法签名**:
```go
func (s *deviceSrv) reachBindLimit(deviceRepo repository.IDeviceRepo, companySetting model.CompanySetting, reqSource, deviceId string) error
```

**功能**: 检查设备绑定数量是否达到上限，防止超额绑定。

**实现流程**:

```161:182:ttpos-server-go/main/app/service/device.go
func (s *deviceSrv) reachBindLimit(deviceRepo repository.IDeviceRepo, companySetting model.CompanySetting, reqSource, deviceId string) error {
	type Source struct {
		Name      string
		Limit     uint
		ErrorCode int
	}
	sources := map[string]Source{
		constant.SourceCashier:   {"收银机", uint(companySetting.CashLimit), constant.CodeCashierLoginLimit},         // "收银机登录设备已达上限，请在其他设备上退出登录或联系销售代表"
		constant.SourceAssistant: {"点餐助手", uint(companySetting.AssistantLimit), constant.CodeAssistantLoginLimit}, // "点餐助手登录设备已达上限，请在其他设备上退出登录或联系销售代表"
		constant.SourceKitchen:   {"厨显", uint(companySetting.KitchenLimit), constant.CodeKitchenLoginLimit},       // "厨显登录设备已达上限，请在其他设备上退出登录或联系销售代表"
		constant.SourceTablet:    {"平板", uint(companySetting.TabletLimit), constant.CodeTabletLoginLimit},         // "平板登录设备已达上限，请在其他设备上退出登录或联系销售代表"
	}
	for sourceName, source := range sources {
		if sourceName != reqSource {
			continue
		}
		if deviceRepo.GetBindCountBySource(sourceName, deviceId) >= source.Limit {
			return errors.NewWithCode(source.ErrorCode, source.Name+"登录设备已达上限，请在其他设备上退出登录或联系销售代表")
		}
	}
	return nil
}
```

**绑定上限配置**:

| 设备类型 | 上限配置字段 | 错误码 | 错误消息 |
|---------|------------|--------|---------|
| 收银机 | CashLimit | CodeCashierLoginLimit | "收银机登录设备已达上限..." |
| 点餐助手 | AssistantLimit | CodeAssistantLoginLimit | "点餐助手登录设备已达上限..." |
| 厨房显示 | KitchenLimit | CodeKitchenLoginLimit | "厨显登录设备已达上限..." |
| 平板点餐 | TabletLimit | CodeTabletLoginLimit | "平板登录设备已达上限..." |

**检查逻辑**:
1. 根据设备类型查找对应的绑定上限
2. 统计当前该类型设备的绑定数量（排除当前设备）
3. 如果达到或超过上限，返回特定错误码
4. 商家后台不检查上限

**使用场景**:
- 商家购买了3个收银机授权，第4台设备登录时会被拒绝
- 提示用户在其他设备退出或联系销售

---

### 3. bindPrinter - 绑定打印机（私有方法）

**方法签名**:
```go
func (s *deviceSrv) bindPrinter(ctx context.Context, deviceId string) error
```

**功能**: 为自带打印功能的收银机自动配置打印机设置。

**实现流程**:

```185:207:ttpos-server-go/main/app/service/device.go
func (s *deviceSrv) bindPrinter(ctx context.Context, deviceId string) error {
	printerSetting, err := s.settingSrv.GetPrinterSetting(ctx, []dto.LanguageItem{})
	if err != nil {
		return errors.WithMessage(err)
	}
	if printerSetting.CashierOpen != "1" {
		return nil
	}
	var added bool
	for _, item := range printerSetting.CashierPrinter {
		if item.Key == deviceId {
			added = true
		}
	}
	if !added {
		printerSetting.CashierPrinter = append(printerSetting.CashierPrinter, setting2.CashierPrinterItem{
			Key:       deviceId,
			PrinterId: deviceId,
		})
	}
	// 设置默认打印机
	return s.settingSrv.UpdateSetting(ctx, constant.SettingPrinter, printerSetting)
}
```

**处理流程**:
1. 获取当前打印机配置
2. 检查收银打印是否开启（`CashierOpen = "1"`）
3. 检查设备是否已添加到打印机列表
4. 未添加则添加到 `CashierPrinter` 列表
5. 更新打印机配置

**触发条件**:
- 设备类型为收银机
- 设备品牌在 `BrandsPrints` 列表中（自带打印功能）
- 情况1: 新设备首次绑定
- 情况2: 软删除设备重新登录

**自带打印品牌示例**:
```go
// constant.BrandsPrints
var BrandsPrints = []string{
    "sunmi",      // 商米
    "smartpos",   // 智能POS
    "other_brands",
}
```

---

### 4. GetRemark - 获取设备备注

**方法签名**:
```go
func (s *deviceSrv) GetRemark(companyUuid uint64, source string, deviceId string) string
```

**功能**: 根据公司、设备类型和设备ID获取设备备注。

**实现流程**:

```209:213:ttpos-server-go/main/app/service/device.go
func (s *deviceSrv) GetRemark(companyUuid uint64, source string, deviceId string) string {
	deviceRepo := repository.NewDeviceRepo(s.dbm.GetDB(companyUuid))
	device, _ := deviceRepo.GetDevice(deviceRepo.WhereSource(source), deviceRepo.WhereSn(deviceId))
	return device.Remark
}
```

**参数说明**:
- `companyUuid`: 公司UUID
- `source`: 设备类型（cashier、assistant、tablet、kitchen、shop）
- `deviceId`: 设备ID

**返回值**: 设备备注字符串，未找到返回空字符串

**使用场景**:
- 显示设备列表时展示备注
- 设备管理界面显示设备名称
- 日志记录中标识设备

---

### 5. IsDeviceBind - 检查设备是否绑定

**方法签名**:
```go
func (s *deviceSrv) IsDeviceBind(ctx context.Context, companyUuid uint64, source string, deviceId string) bool
```

**功能**: 检查设备是否已绑定，并自动更新设备版本号。

**实现流程**:

```215:225:ttpos-server-go/main/app/service/device.go
func (s *deviceSrv) IsDeviceBind(ctx context.Context, companyUuid uint64, source string, deviceId string) bool {
	deviceRepo := repository.NewDeviceRepo(s.dbm.GetDB(companyUuid))
	device, _ := deviceRepo.GetDevice(deviceRepo.WhereSource(source), deviceRepo.WhereSn(deviceId))
	if device.Uuid == 0 {
		return false
	}
	if version := ctx.GetVersion(); device.Version != version {
		deviceRepo.UpdateDevice(device.Uuid, map[string]any{"version": version})
	}
	return true
}
```

**处理逻辑**:
1. 查询设备记录
2. 如果不存在（Uuid = 0），返回 false
3. 如果存在，检查版本号
4. 版本号不一致时自动更新为最新版本
5. 返回 true

**自动版本更新**:
- 每次调用时检查版本号
- 版本不一致自动更新
- 用于统计客户端版本分布

**使用场景**:
- 登录时验证设备是否已绑定
- 中间件验证设备合法性
- 设备管理界面显示绑定状态

---

### 6. UpdateRemark - 更新设备备注

**方法签名**:
```go
func (s *deviceSrv) UpdateRemark(ctx context.Context, editSettingReq req.EditDeviceRemarkReq) error
```

**功能**: 更新当前设备的备注信息。

**请求参数**:
```go
type EditDeviceRemarkReq struct {
    Remark string // 新的备注内容
}
```

**实现流程**:

```227:231:ttpos-server-go/main/app/service/device.go
func (s *deviceSrv) UpdateRemark(ctx context.Context, editSettingReq req.EditDeviceRemarkReq) error {
	return ctx.GetDB().Model(&model.Device{}).Where("uuid = ?", ctx.GetDeviceUuid()).Updates(map[string]any{
		"remark": editSettingReq.Remark,
	}).Error
}
```

**关键点**:
1. 使用 `ctx.GetDeviceUuid()` 获取当前设备 UUID
2. 只能更新当前登录设备的备注
3. 直接使用 GORM 更新数据库

**使用场景**:
- 用户在设置界面修改设备名称/备注
- 区分同类型的多个设备（如"前台收银机"、"后台收银机"）

---

## 数据模型

### Device - 设备表

```go
type Device struct {
    Uuid               uint64 `gorm:"primary_key"` // 设备UUID
    CompanyUuid        uint64                      // 公司UUID
    Source             string                      // 设备来源类型
    DeviceId           string                      // 设备ID（唯一标识）
    Remark             string                      // 设备备注
    Brand              string                      // 设备品牌
    Platform           string                      // 设备平台
    UserAgent          string                      // User-Agent
    FinallyLoginUuid   uint64                      // 最后登录员工UUID
    FinallyLoginTime   int                         // 最后登录时间
    ProductPrinterUuid uint64                      // 产品打印机UUID
    KdsMode            uint                        // 厨显模式
    Version            string                      // 客户端版本号
    IsMain             int                         // 是否主设备（0-否，1-是）
    CreateTime         int                         // 创建时间
    UpdateTime         int                         // 更新时间
    DeleteTime         int                         // 删除时间（软删除）
}
```

### 字段说明

#### Source - 设备来源类型
- `cashier`: 收银机
- `assistant`: 点餐助手
- `tablet`: 平板点餐
- `kitchen`: 厨房显示
- `shop`: 商家后台

#### DeviceId - 设备唯一标识
- 安卓：设备序列号或MAC地址
- iOS：UDID或广告标识符
- Windows/macOS：机器码

#### KdsMode - 厨显工作模式
- `0`: 默认模式
- `1`: 制作模式（只显示制作状态）
- `2`: 制作+送出模式（显示制作和送出两个状态）

#### IsMain - 主设备标记
- 仅收银机使用
- 第一台绑定的收银机自动设为主设备
- 用于区分主收银台和辅助收银台

---

## 业务规则

### 1. 设备绑定流程

```
开始
  ↓
参数验证（设备类型、公司UUID、设备ID）
  ↓
厨显模式验证（仅厨显）
  ↓
平台识别（UserAgent → Platform）
  ↓
检查绑定上限
  ↓
查询现有设备
  ↓
存在？
  ↓ 是              ↓ 否
更新设备信息      创建新设备
  ↓                 ↓
软删除恢复？      第一台收银机？
  ↓ 是              ↓ 是
配置打印机        设为主设备 + 配置打印机
  ↓                 ↓
返回设备UUID
```

### 2. 绑定上限规则

**检查时机**: 每次绑定设备时
**检查对象**: 除当前设备外的同类型设备数量
**限制类型**:
- 收银机: 根据 `CashLimit` 配置
- 点餐助手: 根据 `AssistantLimit` 配置
- 厨房显示: 根据 `KitchenLimit` 配置
- 平板点餐: 根据 `TabletLimit` 配置
- 商家后台: 无限制

**超限处理**:
- 返回特定错误码
- 提示用户在其他设备退出或联系销售

### 3. 主设备规则

**适用范围**: 仅收银机
**判断逻辑**:
```go
IsMain = 1 条件:
1. 设备类型为收银机 AND
2. 不存在其他主设备
```

**作用**:
- 标识主收银台
- 某些功能只能在主设备操作
- 报表统计时区分主辅收银台

### 4. 软删除与恢复

**软删除**: 设置 `delete_time` 为删除时间戳
**恢复条件**: 相同设备再次登录
**恢复操作**:
1. 设置 `delete_time = 0`
2. 更新设备信息
3. 如果是收银机且自带打印，重新配置打印机

### 5. 打印机自动配置

**触发条件**:
```go
1. 设备类型 = 收银机 AND
2. 设备品牌 ∈ BrandsPrints AND
3. (新设备首次绑定 OR 软删除设备恢复)
```

**配置内容**:
- 将设备ID添加到打印机列表
- 设置为默认打印机（如果打印功能已开启）

### 6. 版本号管理

**更新时机**:
- 设备绑定时
- 检查设备是否绑定时（如果版本不一致）

**用途**:
- 统计客户端版本分布
- 版本升级提醒
- 问题排查

---

## 使用场景

### 场景1: 新设备首次登录

```go
// 用户在新的收银机上登录
req := req.AddDeviceReq{
    CompanyUuid:      12345,
    Source:           "cashier",
    DeviceId:         "DEVICE-001",
    Remark:           "前台收银机",
    Brand:            "sunmi",  // 商米（自带打印）
    FinallyLoginUuid: 67890,
    FinallyLoginTime: 1699876543,
}

deviceUuid, err := deviceSrv.AddDevice(ctx, req)
// 结果：
// 1. 创建设备记录
// 2. 设为主设备（如果是第一台收银机）
// 3. 自动配置打印机（品牌支持打印）
```

### 场景2: 设备重新登录（更新信息）

```go
// 设备已绑定，再次登录更新信息
req := req.AddDeviceReq{
    CompanyUuid:      12345,
    Source:           "cashier",
    DeviceId:         "DEVICE-001",
    FinallyLoginUuid: 67891,  // 新员工登录
    FinallyLoginTime: 1699876600,
}

deviceUuid, err := deviceSrv.AddDevice(ctx, req)
// 结果：
// 1. 更新最后登录信息
// 2. 保留原有备注和其他配置
// 3. 更新版本号
```

### 场景3: 软删除设备恢复

```go
// 之前退出登录的设备重新登录
req := req.AddDeviceReq{
    CompanyUuid: 12345,
    Source:      "cashier",
    DeviceId:    "DEVICE-002",
    Brand:       "sunmi",
}

deviceUuid, err := deviceSrv.AddDevice(ctx, req)
// 结果：
// 1. 恢复设备（delete_time = 0）
// 2. 重新配置打印机（如果品牌支持）
// 3. 更新登录信息
```

### 场景4: 达到绑定上限

```go
// 商家有3个收银机授权，尝试绑定第4台
req := req.AddDeviceReq{
    CompanyUuid: 12345,
    Source:      "cashier",
    DeviceId:    "DEVICE-004",
}

deviceUuid, err := deviceSrv.AddDevice(ctx, req)
// 返回错误：
// Code: CodeCashierLoginLimit
// Message: "收银机登录设备已达上限，请在其他设备上退出登录或联系销售代表"
```

### 场景5: 厨显设备配置工作模式

```go
// 绑定厨显设备并设置工作模式
kdsMode := uint(constant.KdsModeMakeAndSend)
req := req.AddDeviceReq{
    CompanyUuid: 12345,
    Source:      "kitchen",
    DeviceId:    "KDS-001",
    Remark:      "前厨显示屏",
    KdsMode:     &kdsMode,  // 制作+送出模式
}

deviceUuid, err := deviceSrv.AddDevice(ctx, req)
// 结果：
// 1. 创建厨显设备
// 2. 设置工作模式为"制作+送出"
```

### 场景6: 设备备注管理

```go
// 获取设备备注
remark := deviceSrv.GetRemark(12345, "cashier", "DEVICE-001")
// 返回: "前台收银机"

// 更新设备备注
err := deviceSrv.UpdateRemark(ctx, req.EditDeviceRemarkReq{
    Remark: "前台收银机（新版）",
})
```

### 场景7: 中间件验证设备绑定

```go
// 在中间件中验证设备是否已绑定
func DeviceAuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        ctx := context.FromGinContext(c)
        companyUuid := ctx.GetCompanyUuid()
        source := c.GetHeader("Source")
        deviceId := c.GetHeader("Device-Id")
        
        // 检查设备是否绑定
        if !deviceSrv.IsDeviceBind(ctx, companyUuid, source, deviceId) {
            c.AbortWithStatusJSON(403, gin.H{
                "code":    403,
                "message": "设备未绑定",
            })
            return
        }
        
        c.Next()
    }
}
```

---

## 最佳实践

### 1. 设备ID生成建议

```go
// Android
deviceId := getAndroidId() // 优先使用 Android ID
if deviceId == "" {
    deviceId = getMacAddress() // 降级使用 MAC 地址
}

// iOS
deviceId := getUDID() // iOS 设备唯一标识

// Windows/macOS
deviceId := getMachineCode() // 机器码

// 要求：
// - 唯一性：同一设备始终相同
// - 稳定性：重装系统后不变
// - 合法性：符合隐私政策
```

### 2. 设备备注命名规范

```go
// 推荐格式：位置 + 设备类型 + 编号
"前台收银机1号"
"二楼点餐助手"
"大厅平板3号"
"后厨显示屏A"

// 避免：
"设备1"  // 不够具体
"iPad"   // 只有型号
"张三的平板" // 与员工绑定
```

### 3. 绑定上限配置建议

```go
// 小型餐厅（50平米以下）
CashLimit:      2  // 2台收银机
AssistantLimit: 2  // 2台点餐助手
TabletLimit:    5  // 5台平板
KitchenLimit:   1  // 1个厨显

// 中型餐厅（50-200平米）
CashLimit:      3
AssistantLimit: 5
TabletLimit:    10
KitchenLimit:   2

// 大型餐厅（200平米以上）
CashLimit:      5
AssistantLimit: 10
TabletLimit:    20
KitchenLimit:   4
```

### 4. 打印机配置检查

```go
// 在设备绑定后验证打印机配置
func verifyPrinterConfig(ctx context.Context, deviceId string, brand string) error {
    if !slices.Contains(constant.BrandsPrints, brand) {
        return nil // 不支持打印的品牌跳过
    }
    
    // 获取打印机配置
    printerSetting, _ := settingSrv.GetPrinterSetting(ctx)
    
    // 验证设备是否在打印机列表中
    found := false
    for _, printer := range printerSetting.CashierPrinter {
        if printer.PrinterId == deviceId {
            found = true
            break
        }
    }
    
    if !found {
        logger.Warn("设备未配置打印机", 
            zap.String("device_id", deviceId),
            zap.String("brand", brand),
        )
    }
    
    return nil
}
```

### 5. 设备列表展示

```go
// 设备列表界面展示建议
type DeviceListItem struct {
    Uuid         uint64 `json:"uuid"`
    Source       string `json:"source"`        // 设备类型
    DeviceId     string `json:"device_id"`     // 设备ID
    Remark       string `json:"remark"`        // 备注（主要显示）
    Brand        string `json:"brand"`         // 品牌
    Platform     string `json:"platform"`      // 平台
    IsMain       int    `json:"is_main"`       // 是否主设备
    IsOnline     bool   `json:"is_online"`     // 是否在线
    LastLogin    string `json:"last_login"`    // 最后登录时间
    LastStaff    string `json:"last_staff"`    // 最后登录员工
    Version      string `json:"version"`       // 客户端版本
}

// 在线状态判断
func isDeviceOnline(device model.Device) bool {
    // 5分钟内有登录记录视为在线
    return time.Now().Unix() - int64(device.FinallyLoginTime) < 300
}
```

---

## 错误处理

### 1. 常见错误

| 错误场景 | 错误消息 | 错误码 | 处理方式 |
|---------|---------|--------|---------|
| 设备类型无效 | "来源设备错误" | - | 检查 Source 参数 |
| 厨显模式错误 | "厨显工作模式错误" | - | 检查 KdsMode 参数 |
| 收银机超限 | "收银机登录设备已达上限..." | CodeCashierLoginLimit | 提示联系销售或退出其他设备 |
| 点餐助手超限 | "点餐助手登录设备已达上限..." | CodeAssistantLoginLimit | 同上 |
| 厨显超限 | "厨显登录设备已达上限..." | CodeKitchenLoginLimit | 同上 |
| 平板超限 | "平板登录设备已达上限..." | CodeTabletLoginLimit | 同上 |
| 更新失败 | "更新绑定信息失败" | - | 检查数据库连接 |
| 打印机配置失败 | "设置默认打印机失败" | - | 检查设置服务 |

### 2. 错误处理示例

```go
deviceUuid, err := deviceSrv.AddDevice(ctx, req)
if err != nil {
    // 根据错误类型处理
    if errors.IsWithCode(err) {
        // 业务错误（如超限）
        code := errors.GetCode(err)
        switch code {
        case constant.CodeCashierLoginLimit:
            // 提示用户购买更多授权或退出其他设备
            return showUpgradeDialog()
        case constant.CodeAssistantLoginLimit:
            // 同上
            return showUpgradeDialog()
        default:
            return showError(err.Error())
        }
    }
    
    // 系统错误
    logger.Error("设备绑定失败",
        zap.Error(err),
        zap.Any("request", req),
    )
    return showError("设备绑定失败，请稍后重试")
}
```

---

## 性能优化

### 1. 查询优化

**问题**: 每次绑定都查询商家配置
```go
// 不推荐：每次都查数据库
companySetting := repository.NewCompanySettingRepo(db).Get()
```

**优化**: 使用缓存
```go
// 推荐：从上下文获取（已缓存）
companySetting := ctx.GetCompanySetting()
```

### 2. 批量设备绑定

**场景**: 门店初始化时批量绑定多个设备

```go
// 优化前：逐个绑定
for _, deviceReq := range deviceReqs {
    deviceSrv.AddDevice(ctx, deviceReq)
}

// 优化后：批量处理
func (s *deviceSrv) BatchAddDevices(ctx context.Context, reqs []req.AddDeviceReq) ([]uint64, error) {
    // 1. 一次性查询所有设备
    // 2. 批量检查上限
    // 3. 批量创建/更新
    // 4. 批量配置打印机
}
```

### 3. 设备列表分页

```go
// 设备列表支持分页
type DeviceListReq struct {
    PageNo    int    // 页码
    PageSize  int    // 每页大小
    Source    string // 设备类型筛选
    IsOnline  *bool  // 在线状态筛选
}

func (s *deviceSrv) GetDeviceList(ctx context.Context, req DeviceListReq) (resp.DeviceListResp, error) {
    // 分页查询
    // 计算在线状态
    // 返回列表
}
```

---

## 安全考虑

### 1. 设备认证

```go
// 登录时验证设备合法性
func Login(ctx context.Context, req LoginReq) (LoginResp, error) {
    // 1. 验证用户名密码
    // 2. 验证设备是否绑定
    if !deviceSrv.IsDeviceBind(ctx, companyUuid, source, deviceId) {
        return nil, errors.New("设备未授权")
    }
    
    // 3. 生成 Token
    // 4. 记录登录信息
}
```

### 2. 防止设备伪造

```go
// 签名验证
func validateDeviceSign(deviceId string, timestamp int64, sign string) bool {
    // 使用密钥和时间戳验证签名
    expectedSign := generateSign(deviceId, timestamp, secretKey)
    return sign == expectedSign
}
```

### 3. 敏感操作权限控制

```go
// 某些操作只能在主设备执行
func CloseShift(ctx context.Context) error {
    device := getDeviceInfo(ctx)
    if device.Source == constant.SourceCashier && device.IsMain != 1 {
        return errors.New("交班操作只能在主收银机执行")
    }
    // 执行交班
}
```

### 4. 设备解绑审计

```go
// 记录设备解绑日志
func UnbindDevice(ctx context.Context, deviceUuid uint64) error {
    device := deviceRepo.GetDevice(deviceUuid)
    
    // 记录审计日志
    auditLog := model.AuditLog{
        Action:      "unbind_device",
        DeviceUuid:  deviceUuid,
        DeviceId:    device.DeviceId,
        OperatorId:  ctx.GetStaffUuid(),
        OperateTime: time.Now(),
    }
    auditLogRepo.Create(auditLog)
    
    // 软删除设备
    deviceRepo.UpdateDevice(deviceUuid, map[string]any{
        "delete_time": time.Now().Unix(),
    })
}
```

---

## 潜在改进点

### 1. 设备分组管理

**当前**: 无分组概念
**改进**: 支持设备分组
```go
type DeviceGroup struct {
    Name    string   // 分组名称（如"一楼"、"二楼"）
    Devices []uint64 // 设备UUID列表
}

// 应用场景：
// - 按楼层分组
// - 按区域分组
// - 批量操作同组设备
```

### 2. 设备状态监控

**改进**: 实时监控设备状态
```go
type DeviceStatus struct {
    DeviceUuid uint64
    IsOnline   bool      // 在线状态
    LastBeat   time.Time // 最后心跳时间
    CpuUsage   float64   // CPU使用率
    MemUsage   float64   // 内存使用率
    DiskUsage  float64   // 磁盘使用率
}

// 心跳上报
func (s *deviceSrv) ReportHeartbeat(ctx context.Context, status DeviceStatus) error
```

### 3. 设备远程控制

**改进**: 支持远程重启、更新等操作
```go
type RemoteCommand struct {
    DeviceUuid uint64
    Command    string // restart, update, clearCache
    Params     map[string]interface{}
}

func (s *deviceSrv) SendCommand(ctx context.Context, cmd RemoteCommand) error {
    // 通过 WebSocket 或推送发送命令
}
```

### 4. 设备使用统计

**改进**: 统计设备使用情况
```go
type DeviceStatistics struct {
    DeviceUuid    uint64
    LoginCount    int     // 登录次数
    AvgOnlineTime int     // 平均在线时长
    OrderCount    int     // 处理订单数
    PrintCount    int     // 打印次数
    ErrorCount    int     // 错误次数
}
```

### 5. 设备权限细化

**改进**: 不同设备不同权限
```go
type DevicePermission struct {
    DeviceUuid  uint64
    Permissions []string // ["order.create", "order.refund", ...]
}

// 示例：
// - 前台收银机：完整权限
// - 后台收银机：只能查询，不能退款
// - 平板：只能下单，不能改价
```

### 6. 设备使用报告

**改进**: 生成设备使用分析报告
```go
type DeviceUsageReport struct {
    Period      string // 统计周期
    TopDevices  []DeviceStatistics // 使用最多的设备
    IdleDevices []Device // 长期未使用的设备
    Suggestions []string // 优化建议
}
```

---

## 相关文件

### DTO 定义
- `ttpos-server-go/app/dto/req/device.go` - 设备请求参数
- `ttpos-server-go/app/dto/resp/device.go` - 设备响应数据

### 数据仓库
- `ttpos-server-go/app/repository/device.go` - 设备数据仓库
- `ttpos-server-go/app/repository/company_setting.go` - 商家配置仓库

### 数据模型
- `ttpos-server-go/app/model/device.go` - 设备模型
- `ttpos-server-go/app/model/company_setting.go` - 商家配置模型

### 设置服务
- `ttpos-server-go/app/service/setting/setting.go` - 设置服务

### 常量定义
- `ttpos-server-go/app/constant/device.go` - 设备常量（设备类型、工作模式等）
- `ttpos-server-go/app/constant/error_code.go` - 错误码定义

---

## 总结

设备管理服务是多终端餐饮系统的基础设施，具有以下特点：

1. **多设备类型支持**: 支持5种设备类型，满足不同场景需求
2. **灵活的绑定策略**: 新建、更新、恢复三种模式自动判断
3. **智能上限控制**: 根据授权动态控制设备数量
4. **自动化配置**: 打印机自动配置，减少人工操作
5. **主设备管理**: 收银机主辅区分，支持功能差异化
6. **软删除机制**: 设备解绑不删除记录，支持快速恢复
7. **版本号管理**: 自动记录和更新客户端版本
8. **平台识别**: 自动识别设备平台和浏览器信息
9. **厨显模式支持**: 灵活配置厨房显示工作模式
10. **完善的错误提示**: 特定错误码和友好提示信息

该服务为多终端协作提供了稳定可靠的设备管理能力，确保各类终端设备的有序接入和管理。

