# Setting Service 系统设置服务说明文档

## 📋 概述

`service/setting` 是 TTPOS 系统的核心设置管理服务，负责管理整个餐饮系统的各类配置，包括店铺信息、打印机配置、各端（收银、助手、平板、厨显、H5）设置、业务规则、支付方式等。该服务采用缓存机制提升性能，支持实时配置更新和多端同步。

**文件路径**: `/home/coder/workspaces/ttpos-server-go/main/app/service/setting/`  
**文件列表**:
- `setting.go` (2049行) - 主服务实现
- `default.go` (544行) - 默认配置定义
- `verify.go` (97行) - 密码验证

**接口定义**: `ISrv`  
**实现结构**: `Srv`

---

## 🏗️ 架构设计

### 接口定义 (ISrv)

```go
type ISrv interface {
    // 基础设置获取
    GetStoreSetting(ctx context.Context) (setting.Store, error)
    GetStoreLanguageList(ctx context.Context) ([]dto.LanguageItem, error)
    GetStoreLanguage(ctx context.Context) ([]string, error)
    GetCompanySetting(ctx context.Context) (model.CompanySetting, error)
    GetCloudBasicSetting(ctx context.Context) (setting.CloudBasic, error)
    
    // 各端设置获取
    GetCashierSetting(ctx context.Context, languageList []dto.LanguageItem) (setting.Cashier, error)
    GetAssistantSetting(ctx context.Context, languageList []dto.LanguageItem) (setting.Assistant, error)
    GetTabletSetting(ctx context.Context, languageList []dto.LanguageItem) (setting.Tablet, error)
    GetKitchenSetting(ctx context.Context, companySetting model.CompanySetting, languageList []dto.LanguageItem) (setting.Kitchen, error)
    GetH5Setting(ctx context.Context, languageList []dto.LanguageItem) (setting.H5, error)
    
    // 业务设置获取
    GetBusinessSetting(ctx context.Context) (setting.Business, error)
    GetBuffetSetting(ctx context.Context, companySetting model.CompanySetting) (setting.BuffetResp, error)
    GetPaymentSetting(ctx context.Context, companySetting model.CompanySetting) (setting.Payment, error)
    GetCurrencySetting(ctx context.Context) (setting.Currency, error)
    GetPointsSetting(ctx context.Context) (setting.Points, error)
    GetServiceFeeSetting(ctx context.Context) (setting.ServiceCharge, error)
    GetTaxRateSetting(ctx context.Context) (setting.TaxRate, error)
    
    // 打印机设置
    GetPrinterSetting(ctx context.Context, languageList []dto.LanguageItem) (setting.Printer, error)
    GetPrinterInfo(ctx context.Context, printerSetting setting.Printer, deviceId string) (setting.PrinterInfo, error)
    
    // 其他功能
    GetCashierLanguage(c context.Context) (resp.LanguageResp, error)
    GetCashierAd(ctx context.Context) (resp.Ads, error)
    GetCashierBaseSetting(ctx context.Context) (resp.CashierBaseSetting, error)
    GetAcceptOrderSetting(ctx context.Context) (*resp.AcceptOrderSetting, error)
    GetShopBusinessSetting(ctx context.Context) (setting.ShopBusiness, error)
    GetMenuQrcode(ctx context.Context) (string, error)
    GetPaymentMethodList(ctx context.Context) setting.PaymentMethodListResp
    
    // 设置更新
    UpdateSetting(ctx context.Context, settingKey string, values any) error
    EditAcceptOrderSetting(ctx context.Context, orderSetting req.UpdateAcceptOrderSetting) error
    EditAcceptMemberOrderSetting(ctx context.Context, orderSetting req.UpdateAcceptMemberOrderSetting) error
    EditSystemSetting(ctx context.Context, systemSetting req.UpdateSystemSetting) error
    EditStoreSetting(ctx context.Context, storeSetting req.UpdateStoreSetting) error
    EditBusinessSetting(ctx context.Context, businessSetting req.UpdateBusinessSetting) error
    
    // 密码验证
    VerifyPassword(ctx context.Context, source string, typ string, password string) bool
    VerifyAdvancedPassword(ctx context.Context, password string, options ...func(option *VerifyAdvancedPasswordOption)) error
    
    // 工具方法
    SymbolPosition(ctx context.Context, price float64) string
    CheckUpdate(ctx context.Context, appType int, brand string, language string) (resp.UpdateInfo, error)
}
```

### 依赖服务

```go
type Srv struct {
    dbm           *database.DBManager  // 数据库管理器
    cache         cache.Cache          // 缓存服务
    cacheKey      string              // 缓存键模板
    cloudCacheKey string              // 云端缓存键
}
```

---

## 🎯 核心功能

### 1. 缓存机制 (fromCache)

**功能描述**: 设置数据从缓存读取，不存在则从数据库加载并写入缓存。

#### 缓存策略

```go
cacheKey := fmt.Sprintf("setting:company_id:%d", companyUuid)
```

#### 实现逻辑

```
1. 尝试从缓存读取
   ↓
2. 缓存命中 → 返回数据
   ↓
3. 缓存未命中
   ↓
4. 从数据库查询所有设置
   ↓
5. 序列化为JSON
   ↓
6. 写入缓存（永久）
   ↓
7. 返回数据
```

#### 代码示例

```go
func (s *Srv) fromCache(ctx context.Context) ([]model.Setting, error) {
    companyUuid := ctx.GetCompanyUuid()
    var settings []model.Setting
    cacheKey := fmt.Sprintf(s.cacheKey, companyUuid)
    
    // 尝试从缓存读取
    if data, exists := s.cache.Get(cacheKey); exists {
        if dataValue, isString := data.(string); isString {
            if err := json.Unmarshal([]byte(dataValue), &settings); err != nil {
                return settings, errors.WithMessage(err)
            }
            return settings, nil
        }
    }
    
    // 从数据库读取
    settingRepo := repository.NewSettingRepo(s.dbm.GetDB(companyUuid))
    settings, err = settingRepo.GetAll()
    if err != nil {
        ctx.Log().Error("从数据库获取设置失败", zap.Error(err))
        return nil, errors.WithMessage(err, "获取设置失败")
    }
    
    // 写入缓存
    data, _ := json.Marshal(settings)
    s.cache.Set(cacheKey, string(data), 0)
    
    return settings, nil
}
```

---

### 2. 店铺设置 (GetStoreSetting)

**功能描述**: 获取店铺基本信息，包括名称、Logo、语言、时区等。

#### 返回数据结构

```go
type Store struct {
    Name          string          // 商城名称
    AvatarURL     string          // 默认头像
    LogoURL       string          // 商城logo
    Phone         string          // 联系电话
    Address       string          // 地址
    Coordinates   string          // 坐标 "lat,lng"
    Latitude      string          // 纬度
    Longitude     string          // 经度
    Company       string          // 公司名称
    TaxNumber     string          // 税号
    TimeZoneList  []TimeZoneItem  // 时区列表
    Language      []LanguageItem  // 语言列表
    IPWhiteList   string          // IP白名单
}
```

#### 处理逻辑

1. **JSON字段转换**: 处理 `logoUrl` → `logo_url`、`avatarUrl` → `avatar_url`
2. **语言key类型转换**: 将字符串key转换为整数
3. **合并默认值**: 使用copier合并默认店铺设置
4. **图片域名处理**: 为Logo和头像添加域名前缀
5. **坐标解析**: 将坐标字符串拆分为经纬度

---

### 3. 打印机设置 (GetPrinterSetting)

**功能描述**: 获取打印机相关配置，支持收银小票打印和厨房打印。

#### 返回数据结构

```go
type Printer struct {
    CashierOpen        string              // 是否开启打印 "0"/"1"
    CashierPrinterID   string              // 打印机ID
    CashierPrinter     []CashierPrinterItem // 打印机列表
    LanguageList       []LanguageItem      // 语言列表
    LanguageMethod     string              // 语言方式 "1"-单语言 "2"-多语言
    DefaultLanguage    string              // 默认打印语言
    PrintMethod        string              // 打印方式 "1"-文本 "2"-图片
    KitchenLanguage    string              // 厨房打印语言
    KitchenPrintMethod string              // 厨房打印方式
    ConsumptionTax     string              // 消费税显示
    BuffetSignOpen     string              // 自助餐标识
    MonetaryUnitOpen   string              // 货币单位显示
    DefaultCalendar    string              // 日历类型
    CalendarList       []CalendarItem      // 日历列表
    PrintList          []PrintItem         // 打印方式列表
}

type CashierPrinterItem struct {
    Key          string  // 设备标识
    PrinterId    string  // 打印机ID
    PrinterUsbId string  // USB打印机ID
    Sn           string  // 打印机SN
}
```

#### 打印机信息获取 (GetPrinterInfo)

**功能描述**: 根据设备ID获取对应的打印机详细信息。

```go
type PrinterInfo struct {
    PrinterType            string  // 打印机类型
    PrinterUuid            uint64  // 打印机UUID
    Copies                 uint    // 打印份数
    PrinterConfig          string  // 打印机配置JSON
    IsCashierPrinter       bool    // 是否收银机内置打印机
    IsCashierOpen          bool    // 是否开启打印
    PrinterCashierDeviceSn string  // 收银设备SN
    IsUsbPrinter           bool    // 是否USB打印机
    PrintMethod            int     // 打印方式
    PrinterSn              string  // 打印机SN
    PrinterWidth           int     // 打印机宽度(mm)
    EnableStatusCheck      int     // 是否启用状态检查
    EnableSound            int     // 是否启用打印提示音
}
```

#### 打印机类型

| 类型 | 说明 |
|-----|------|
| 普通打印机 | UUID < 18位纯数字，需要查询printer表 |
| USB打印机 | PrinterUsbId不为空 |
| 商米内置打印机 | brand在SunmiAllPrints中 |
| Compax内置打印机 | brand为A11510P |

---

### 4. 收银机设置 (GetCashierSetting)

**功能描述**: 获取收银端配置，包括用餐方式、自动送厨、锁屏、接单等设置。

#### 返回数据结构

```go
type Cashier struct {
    CashierResp
    AdvancedPassword string  // 高级设置密码
    CashierPassword  string  // 钱箱密码
    LockPassword     string  // 锁屏密码
}

type CashierResp struct {
    Carousel               []CarouselItem  // 轮播图/视频
    IsAutoSend             string          // 自动送厨 "0"/"1"
    OrderMethod            OrderMethod     // 用餐方式
    IsRemainColor          string          // 剩余时长颜色 "0"/"1"
    RemainColor            []string        // 颜色列表
    IsOpenCashierPassword  string          // 钱箱密码开关
    IsAutoLockScreen       string          // 自动锁屏开关
    AutoLockScreen         string          // 锁屏时间(秒)
    IsShowScanSoldOut      int             // H5端显示售罄
    IsShowAssistantSoldOut int             // 助手端显示售罄
    LanguageList           []LanguageItem  // 语言列表
    Language               []string        // 常用语言
    DefaultLanguage        string          // 默认语言
    IsAutoOrder            string          // 自动接单开关
    AutoOrderLimit         string          // 自动接单金额上限
    IsAutoVoice            string          // 接单语音播报
    IsAutoMemberOrder      string          // 自动接单会员订单
    AutoMemberOrderLimit   string          // 会员订单金额上限
    IsAutoVoiceMemberOrder string          // 会员订单语音播报
    MenuShowSoldOut        string          // 菜单显示售罄
    MemberShowSoldOut      string          // 会员端显示售罄
}

type OrderMethod struct {
    IsCashierOrder string  // 收银用餐 "0"/"1"
    IsTableOrder   string  // 桌台用餐 "0"/"1"
}
```

#### 特殊处理

1. **轮播图处理**: 自动添加域名前缀
2. **字段类型转换**: `is_show_scan_sold_out` 等字段从数字/字符串统一转为int
3. **语言过滤**: 仅返回有效的语言配置
4. **默认值合并**: 使用copier合并默认配置

---

### 5. 点餐助手设置 (GetAssistantSetting)

**功能描述**: 获取点餐助手端配置。

#### 返回数据结构

```go
type Assistant struct {
    AssistantResp
    AdvancedPassword string  // 高级设置密码
    LockPassword     string  // 锁屏密码
}

type AssistantResp struct {
    Server                 Server         // 服务器连接
    IsRemainColor          string         // 剩余时长颜色开关
    RemainColor            []string       // 颜色列表
    DefaultMode            string         // 默认模式 "0"-服务员 "1"-顾客
    IsAutoLockScreen       string         // 自动锁屏开关
    IsCheckOrder           string         // 下单校验高级密码
    AutoLockScreen         string         // 锁屏时间(秒)
    LanguageList           []LanguageItem // 语言列表
    Language               []string       // 常用语言
    DefaultLanguage        string         // 默认语言
    IsShowAssistantSoldOut int            // 显示售罄商品
}

type Server struct {
    IP   string  // 服务器IP
    Port string  // 服务器端口
}
```

#### IP和端口获取

```go
func (s *Srv) getIPAndPort() (string, string) {
    var serverIP, serverPort string
    serverIP = viper.GetString("HARDWARE_SERVER_URL")
    if serverIP == "" {
        serverIP, _ = utils.GetLocalIP()  // 自动获取本地IP
    }
    serverIP = strings.ReplaceAll(serverIP, "addr:", "")
    serverPort = viper.GetString("HARDWARE_SERVER_PORT")
    if serverPort == "" {
        serverPort = "8080"  // 默认端口
    }
    return serverIP, serverPort
}
```

---

### 6. 平板设置 (GetTabletSetting)

**功能描述**: 获取平板端配置，支持顾客自助点餐。

#### 返回数据结构

```go
type Tablet struct {
    TabletResp
    AdvancedPassword string  // 高级设置密码
}

type TabletResp struct {
    Carousel           []CarouselItem     // 轮播图/视频
    IsCallService      string             // 呼叫服务员开关
    IsCustomerOrder    string             // 顾客自助下单开关
    IsVoiceRemind      string             // 声音提醒开关
    IsShowSoldOut      string             // 显示售罄商品
    IsBuffetOrderLimit string             // 自助餐下单限制
    BuffetOrderLimit   BuffetOrderLimit   // 自助餐限制详情
    IsOrderLimit       string             // 非自助餐下单限制
    OrderLimit         OrderLimit         // 非自助餐限制详情
    Server             Server             // 服务器连接
    LanguageList       []LanguageItem     // 语言列表
    Language           []string           // 常用语言
    DefaultLanguage    string             // 默认语言
}

type BuffetOrderLimit struct {
    IsLimitTime string  // 时间限制开关
    LimitTime   string  // 时间限制(分钟)
    IsLimitNum  string  // 数量限制开关
    LimitNum    string  // 数量限制
}

type OrderLimit struct {
    IsLimitTime string  // 时间限制开关
    LimitTime   string  // 时间限制(分钟)
    IsLimitNum  string  // 数量限制开关
    LimitNum    string  // 数量限制
}
```

#### 字段类型转换

```go
func (s *Srv) convertTabletFormat(oldVal string) ([]byte, error) {
    tabletMap := map[string]any{}
    json.Unmarshal([]byte(oldVal), &tabletMap)
    
    // 将数字字段转换为字符串
    if v, ok := tabletMap["is_call_service"].(float64); ok {
        tabletMap["is_call_service"] = fmt.Sprintf("%.0f", v)
    }
    // ... 其他字段同理
    
    return json.Marshal(tabletMap)
}
```

---

### 7. 厨显设置 (GetKitchenSetting)

**功能描述**: 获取厨房显示端配置。

#### 返回数据结构

```go
type Kitchen struct {
    KitchenResp
    AdvancedPassword string  // 高级设置密码
}

type KitchenResp struct {
    IsOpen          string         // 功能开关
    IsComeDish      string         // 来菜提醒开关
    IsCallService   string         // 顾客呼叫提醒开关
    Server          Server         // 服务器连接
    IsWaitColor     string         // 等待时长颜色开关
    WaitColor       []string       // 颜色列表
    LanguageList    []LanguageItem // 语言列表
    Language        []string       // 常用语言
    DefaultLanguage string         // 默认语言
    IsSmartKitchen  string         // 智能后厨开关
}
```

#### 权限控制

```go
// 总权限控制 - 如果公司设置中未开启厨显
if companySetting.IsOpenKitchenKds == 0 {
    kitchen.IsOpen = "0"
}
```

---

### 8. H5设置 (GetH5Setting)

**功能描述**: 获取H5扫码点餐端配置。

#### 返回数据结构

```go
type H5 struct {
    IsCallService      string           // 呼叫服务员开关
    IsCustomerOrder    string           // 顾客自助下单开关
    IsVoiceRemind      string           // 声音提醒开关
    IsShowScanSoldOut  int              // 显示售罄商品
    IsBuffetOrderLimit string           // 自助餐下单限制
    BuffetOrderLimit   BuffetOrderLimit // 自助餐限制详情
    IsOrderLimit       string           // 非自助餐下单限制
    OrderLimit         OrderLimit       // 非自助餐限制详情
    LanguageList       []LanguageItem   // 语言列表
    Language           []string         // 常用语言
    DefaultLanguage    string           // 默认语言
}
```

#### 特殊处理

```go
// 处理空数组转对象
if strings.Contains(st.Values, "\"buffet_order_limit\":[]") {
    st.Values = strings.Replace(st.Values, "\"buffet_order_limit\":[]", "\"buffet_order_limit\":{}", -1)
}
```

---

### 9. 业务设置 (GetBusinessSetting)

**功能描述**: 获取门店业务规则配置。

#### 返回数据结构

```go
type Business struct {
    ZeroingMethodList         []ZeroingMethodItem         // 抹零方式列表
    CheckoutZeroingMethodList []CheckoutZeroingMethodItem // 结账抹零方式列表
    ZeroingMethod             string                      // 优惠折扣抹零方式
    CheckoutZeroingMethod     string                      // 结账抹零方式
    GiftMethodList            []GiftMethodItem            // 赠菜计算方式列表
    GiftMethod                string                      // 赠菜计算方式
    FreeMethodList            []FreeMethodItem            // 免单计算方式列表
    FreeMethod                string                      // 免单计算方式
    DiscountMethod            string                      // 折扣计算方式
    QrCode                    string                      // 电子菜单二维码
    NoClearTable              string                      // 结账后不清台
    IsNeedPassword            string                      // 取消订单/退菜需要密码
    DishCardStyle             string                      // 菜品卡片样式
    DishCardStyleTime         string                      // 卡片样式更新时间
    IsInvoice                 string                      // 开票信息
    OpeningHours              string                      // 营业时间
    DeliveryPriceRatio        int                         // 外送价格比例
    StartSerialNo             string                      // 开始序列号
    IsBatch                   string                      // 是否分批商品
    BatchProductUuids         []uint64                    // 分批商品UUID列表
    BatchTagNum               uint                        // 分批类型数量
    SafetyStockType           string                      // 安全库存类型
    RequiredParentCompanyApproval string                  // 调拨规则-上级审批
    ViaParentCompanyWarehouse     string                  // 调拨规则-上级仓库
}
```

#### 抹零方式

| Key | 说明 |
|-----|------|
| "0" | 实款实收 |
| "1" | 抹分 |
| "2" | 抹角 |
| "3" | 四舍五入保留一位小数 |
| "4" | 四舍五入到整数 |
| "5" | 抹元 |

#### 赠菜计算方式

| Key | 说明 |
|-----|------|
| "10" | 计入总销售额、优惠折扣 |
| "20" | 不计入总销售额、优惠折扣 |

#### 免单计算方式

| Key | 说明 |
|-----|------|
| "10" | 计入总销售额、优惠折扣、服务费、税费 |
| "20" | 不计入总销售额、优惠折扣、服务费、税费 |

#### 兼容性处理

```go
// 兼容v1.0版本字段为数字的情况
{
    re := regexp.MustCompile(`"dish_card_style":(\s*)(\d+)`)
    st.Values = re.ReplaceAllString(st.Values, `"dish_card_style":"$2"`)
}
```

#### 分批商品统计

```go
// 分批商品数量
batchProductUuids, err := productPackageRepo.GetProductPackageBatchTagCount()
defaultBusiness.BatchProductUuids = batchProductUuids

// 分批类型数量
batchTagNum, err := repository.NewBatchTagRepo(db).GetBatchTagCount()
defaultBusiness.BatchTagNum = uint(batchTagNum)
```

---

### 10. 自助餐设置 (GetBuffetSetting)

**功能描述**: 获取自助餐相关配置。

#### 返回数据结构

```go
type BuffetResp struct {
    IsOpen                   string        // 功能开关
    TabletEndTime            int           // 平板结束时间提醒(分钟)
    IsRemainContinue         string        // 剩余时间不可继续点餐开关
    RemainContinueTime       string        // 剩余时间不可继续点餐(分钟)
    RemainContinueNoticeTime string        // 剩余时间提醒(分钟)
    IsBuyContinue            string        // 非自助餐商品到时可继续选购
    IsAddClock               string        // 加钟开关
    IsBuffetDiscount         string        // 自助餐折扣开关
    AddClock                 []AddClockItem // 加钟列表
}

type AddClockItem struct {
    Name  string  // 加钟名称
    Time  string  // 加钟时间(分钟)
    Price string  // 加钟价格
}
```

#### 权限控制

```go
// 如果公司设置中未开启自助餐
if companySetting.IsOpenBuffet == 0 {
    buffet.IsOpen = "0"
}
```

---

### 11. 支付方式设置 (GetPaymentSetting)

**功能描述**: 获取支付方式配置。

#### 返回数据结构

```go
type Payment struct {
    IsCash    string  // 现金支付开关
    IsBalance string  // 余额支付开关
    IsOther   string  // 其他方式支付开关
}
```

#### 权限联动

```go
// 会员关闭时，余额支付也关闭
if companySetting.IsOpenMember == 0 {
    payment.IsBalance = "0"
}
```

---

### 12. 货币设置 (GetCurrencySetting)

**功能描述**: 获取货币单位和显示配置。

#### 返回数据结构

```go
type Currency struct {
    Unit             string  // 主货币单位 (如: ฿)
    PrintUnit        string  // 打印专用货币单位
    UnitPosition     string  // 主货币显示位置 "0"-前 "1"-后
    IsOpen           string  // 副货币单位开关
    ViceUnit         string  // 副货币单位
    ViceUnitPosition string  // 副货币显示位置
}
```

#### 货币符号位置格式化

```go
func (s *Srv) SymbolPosition(ctx context.Context, amount float64) string {
    currencySetting, _ := s.GetCurrencySetting(ctx)
    if currencySetting.UnitPosition == "0" {
        return currencySetting.Unit + " " + utils.FormatAmount(amount)
    } else {
        return utils.FormatAmount(amount) + " " + currencySetting.Unit
    }
}
```

**示例**:
- UnitPosition="0": `฿ 100.00`
- UnitPosition="1": `100.00 ฿`

---

### 13. 积分设置 (GetPointsSetting)

**功能描述**: 获取会员积分规则配置。

#### 返回数据结构

```go
type Points struct {
    DeductionOrder     string              // 扣款顺序
    DeductRatioMain    string              // 主账户扣款比例
    DeductRatioGift    string              // 赠送账户扣款比例
    PointsName         string              // 积分名称
    IsShoppingGift     string              // 购物送积分开关
    GiftRatio          string              // 积分赠送比例
    IsShoppingDiscount string              // 积分抵扣开关
    Discount           DiscountItem        // 抵扣规则
    Describe           string              // 积分说明
    ShoppingGiftRules  []ShoppingGiftRule  // 购物送积分规则
    Exchange           PointsExchange      // 积分兑换规则
}

type ShoppingGiftRule struct {
    Type                     string            // 规则类型
    IsOpen                   string            // 开关
    IsMemberLevelRelated     string            // 会员等级关联
    Value                    string            // 值
    PaymentAmountRequirement string            // 支付金额要求
    MealType                 []string          // 用餐类型
    BalancePaymentGetPoints  string            // 余额支付是否赠送积分
    RefundReturnPoints       string            // 退款是否返还积分
    MemberLevels             []MemberLevelItem // 会员等级列表
}
```

#### 规则类型

| 类型 | 说明 |
|-----|------|
| `RuleTypePaymentAmount` | 按支付金额赠送 |
| `RuleTypeDesk` | 按桌台数量赠送 |

#### 会员等级处理

```go
memberLevels := memberRepo.GetMemberLevels()
for _, level := range memberLevels {
    rateMemberLevels = append(rateMemberLevels, setting.MemberLevelItem{
        Uuid:  level.Uuid,
        Name:  level.Name,
        Value: level.PointsRate,      // 积分倍率
    })
    quantityMemberLevels = append(quantityMemberLevels, setting.MemberLevelItem{
        Uuid:  level.Uuid,
        Name:  level.Name,
        Value: level.PointsQuantity,  // 积分数量
    })
}
```

#### 用餐类型过滤

```go
// 如果未开启自助餐，过滤掉buffet类型
if companySetting.IsOpenBuffet == 0 {
    newMealType := slice.Filter(rule.MealType, func(index int, item string) bool {
        return item != "buffet"
    })
    defaultPoints.ShoppingGiftRules[i].MealType = newMealType
}
```

---

### 14. 服务费设置 (GetServiceFeeSetting)

**功能描述**: 获取服务费收取配置。

#### 返回数据结构

```go
type ServiceCharge struct {
    IsOpen              string   // 功能开关
    ChargeType          string   // 服务费类型 "1"-固定金额 "2"-百分比
    ServiceCharge       string   // 服务费值
    IsOpenTax           string   // 是否收取税费
    ApplyScope          string   // 适用范围 "1"-全部 "2"-部分
    ApplyScopeOrdering  string   // 适用范围-点餐
    ApplyScopeTable     string   // 适用范围-桌台
    ApplyScopeTableList []int64  // 适用范围-桌台ID列表
}
```

#### 字段类型转换

```go
func (s *Srv) convertServiceFeeFormat(oldVal string) ([]byte, error) {
    serviceFeeMap := map[string]any{}
    json.Unmarshal([]byte(oldVal), &serviceFeeMap)
    
    // 将数字转换为字符串
    if v, ok := serviceFeeMap["service_charge"].(float64); ok {
        serviceFeeMap["service_charge"] = fmt.Sprintf("%f", v)
    }
    
    return json.Marshal(serviceFeeMap)
}
```

---

### 15. 税率设置 (GetTaxRateSetting)

**功能描述**: 获取税率管理配置。

#### 返回数据结构

```go
type TaxRate struct {
    IsOpen         string              // 功能开关
    CalcType       string              // 计算类型 "1"-已含税 "2"-未含税
    AddTaxCategory []AddTaxCategoryItem // 增值税分类
}

type AddTaxCategoryItem struct {
    CategoryName string  // 分类名称
    TaxRate      string  // 税率
}
```

---

### 16. 云端基础设置 (GetCloudBasicSetting)

**功能描述**: 获取云端平台的基础配置，如品牌信息等。

#### 数据来源

1. **缓存优先**: 先从缓存读取
2. **远程获取**: 缓存不存在时，调用云端API获取
3. **本地缓存**: 获取后写入本地缓存

#### 远程API调用

```go
type Base struct {
    Code int    `json:"code"`
    Msg  string `json:"msg"`
    Data struct {
        Base struct {
            BrandName          string `json:"brand_name"`
            BrandLogo          string `json:"brand_logo"`
            BrandLogoLong      string `json:"brand_logo_long"`
            BrowserLogo        string `json:"browser_logo"`
            BrowserTitle       string `json:"browser_title"`
            ExpirationReminder int    `json:"expiration_reminder"`
        } `json:"base"`
    } `json:"data"`
}

res, err := gohttp.NewRequest().Post(
    viper.GetString("CLOUD_PLATFORM_HOST") + "/api/admin/setting.service/info"
)
```

#### 返回数据结构

```go
type CloudBasic struct {
    BrandName          string  // 品牌名称
    BrandLogo          string  // 品牌Logo
    BrandLogoLong      string  // 品牌长Logo
    BrowserLogo        string  // 浏览器Logo
    BrowserTitle       string  // 浏览器标题
    ExpirationReminder int     // 到期提醒(天数)
}
```

#### 图片处理

```go
// 自动添加域名前缀
cloudBasic.BrandLogo = utils.AddImageDomain(
    utils.RemoveDomain(cloudBasic.BrandLogo), 
    utils.GetBaseURL(ctx.GetGin().Request), 
    true
)
```

---

## 🔧 设置更新功能

### 1. 通用设置更新 (UpdateSetting)

**功能描述**: 更新任意设置项的值。

#### 更新流程

```
1. 序列化values为JSON
   ↓
2. 特殊字段处理 (如Store的logoUrl)
   ↓
3. 查询设置是否存在
   ↓
4. 不存在 → 创建新记录
   ↓
5. 存在 → 更新记录
   ↓
6. 删除缓存
   ↓
7. 返回成功
```

#### 代码实现

```go
func (s *Srv) UpdateSetting(ctx context.Context, settingKey string, values any) error {
    value := utils.ToJson(values)
    
    // 特殊字段处理
    if settingKey == constant.SettingStore {
        value = strings.ReplaceAll(value, "\"logo_url\"", "\"logoUrl\"")
        value = strings.ReplaceAll(value, "\"avatar_url\"", "\"avatarUrl\"")
    }
    
    db := ctx.GetDB()
    if db == nil {
        db = s.dbm.GetDB(ctx.GetCompanyUuid())
    }
    
    settingRepo := repository.NewSettingRepo(db)
    set := settingRepo.GetByKey(settingKey)
    
    if set.Key == "" {
        // 创建新记录
        if _, err := settingRepo.Create(model.Setting{
            Key:    settingKey,
            Values: value,
        }); err != nil {
            return errors.New("更新设置失败")
        }
    } else {
        // 更新记录
        if err := settingRepo.Updates(settingKey, value); err != nil {
            return errors.New("更新设置失败")
        }
    }
    
    // 删除缓存
    s.cache.Del(fmt.Sprintf(s.cacheKey, ctx.GetCompanyUuid()))
    return nil
}
```

---

### 2. 修改接单设置 (EditAcceptOrderSetting)

**功能描述**: 修改H5订单自动接单配置。

#### 更新字段

```go
type UpdateAcceptOrderSetting struct {
    IsAutoOrder    string  // 自动接单开关
    AutoOrderLimit string  // 自动接单金额上限
}
```

#### 实现逻辑

```go
func (s *Srv) EditAcceptOrderSetting(ctx context.Context, orderSetting req.UpdateAcceptOrderSetting) error {
    cashierSetting, err := s.GetCashierSetting(ctx, nil)
    if err != nil {
        return errors.WithMessage(err)
    }
    
    cashierSetting.IsAutoOrder = orderSetting.IsAutoOrder
    cashierSetting.AutoOrderLimit = orderSetting.AutoOrderLimit
    
    return s.UpdateSetting(ctx, constant.SettingCashier, cashierSetting)
}
```

---

### 3. 修改会员订单接单设置 (EditAcceptMemberOrderSetting)

**功能描述**: 修改会员外送订单自动接单配置。

#### 更新字段

```go
type UpdateAcceptMemberOrderSetting struct {
    IsAutoMemberOrder    string  // 自动接单开关
    AutoMemberOrderLimit string  // 自动接单金额上限
}
```

---

### 4. 修改系统设置 (EditSystemSetting)

**功能描述**: 批量修改系统相关设置，涉及多个设置项。

#### 更新字段

```go
type UpdateSystemSetting struct {
    IsShowAssistantSoldOut *int    // 助手端显示售罄
    IsShowScanSoldOut      *int    // H5端显示售罄
    MenuShowSoldOut        *int    // 菜单显示售罄
    MemberShowSoldOut      *int    // 会员端显示售罄
    DishCardStyle          string  // 菜品卡片样式
    IsShowSoldOut          *int    // 平板端显示售罄
    DeviceRemark           string  // 设备备注
}
```

#### 实现逻辑

```go
func (s *Srv) EditSystemSetting(ctx context.Context, systemSetting req.UpdateSystemSetting) error {
    // 1. 更新收银机设置
    cashierSetting, _ := s.GetCashierSetting(ctx, nil)
    cashierSetting.IsShowAssistantSoldOut = *systemSetting.IsShowAssistantSoldOut
    cashierSetting.IsShowScanSoldOut = *systemSetting.IsShowScanSoldOut
    cashierSetting.MenuShowSoldOut = strconv.Itoa(*systemSetting.MenuShowSoldOut)
    cashierSetting.MemberShowSoldOut = strconv.Itoa(*systemSetting.MemberShowSoldOut)
    s.UpdateSetting(ctx, constant.SettingCashier, cashierSetting)
    
    // 2. 更新业务设置
    businessSetting, _ := s.GetBusinessSetting(ctx)
    businessSetting.DishCardStyle = systemSetting.DishCardStyle
    s.UpdateSetting(ctx, constant.SettingBusiness, businessSetting)
    
    // 3. 更新平板设置
    tabletSetting, _ := s.GetTabletSetting(ctx, nil)
    tabletSetting.IsShowSoldOut = strconv.Itoa(*systemSetting.IsShowSoldOut)
    s.UpdateSetting(ctx, constant.SettingTablet, tabletSetting)
    
    // 4. 更新设备备注
    repository.NewDeviceRepo(s.dbm.GetDB(ctx.GetCompanyUuid())).UpdateDevice(
        ctx.GetDeviceUuid(), 
        map[string]any{"remark": systemSetting.DeviceRemark},
    )
    
    return nil
}
```

---

### 5. 修改店铺设置 (EditStoreSetting)

**功能描述**: 修改店铺基本信息，是最复杂的更新操作之一。

#### 更新字段

```go
type UpdateStoreSetting struct {
    Name        string          // 店铺名称
    Phone       string          // 联系电话
    Address     string          // 地址
    Coordinates string          // 坐标
    TimeZone    string          // 时区
    Language    []LanguageItem  // 语言列表
    LogoUrl     string          // Logo地址
    CompanyName string          // 公司名称
}
```

#### 更新流程

```
1. 获取当前店铺设置
   ↓
2. 验证时区是否有效
   ↓
3. 验证语言是否有效
   ↓
4. 处理图片域名
   ↓
5. 更新各端设置中的语言配置
   - 收银机设置
   - 平板设置
   - H5设置
   - 厨显设置
   - 点餐助手设置
   - 打印机设置
   ↓
6. 开启事务
   ↓
7. 更新saas库的company和company_setting
   ↓
8. 更新商家库的company和company_setting
   ↓
9. 更新setting表中的各项设置
   ↓
10. 如果是总部且开启ERP，更新供应商名称
   ↓
11. 提交事务
   ↓
12. 删除缓存
   ↓
13. 推送配置更新消息到WebSocket
```

#### 语言配置同步逻辑

```go
// 以收银机设置为例
cashierSetting, _ := s.GetCashierSetting(ctx, storeSettingReq.Language)

// 1. 删除不在新语言列表中的语言
for _, language := range cashierSetting.Language {
    if !slices.ContainsFunc(storeSettingReq.Language, func(item dto.LanguageItem) bool {
        return item.Name == language
    }) {
        cashierSetting.Language = slices.DeleteFunc(cashierSetting.Language, func(item string) bool {
            return item == language
        })
    }
}

// 2. 确保至少有一种语言
if len(cashierSetting.Language) == 0 {
    cashierSetting.Language = []string{storeSettingReq.Language[0].Name}
}

// 3. 如果默认语言不在列表中，设置为第一个语言
if !slices.Contains(cashierSetting.Language, cashierSetting.DefaultLanguage) {
    cashierSetting.DefaultLanguage = cashierSetting.Language[0]
}

// 4. 清除LanguageList（会自动补充）
cashierSetting.LanguageList = []dto.LanguageItem{}
```

#### WebSocket推送

```go
utils.Go(func() {
    websocket.PushClient(companyUuid, websocket.SourceAll, websocket.SourceAll, websocket.UPDATE_CONFIG, map[string]any{
        "update_time": time.Now().Unix(),
    })
})
```

---

### 6. 修改业务设置 (EditBusinessSetting)

**功能描述**: 修改门店业务规则配置。

#### 更新流程

```
1. 获取当前业务设置
   ↓
2. 判断菜品卡片样式是否更新
   - 更新则记录时间戳
   ↓
3. 验证外送功能权限
   ↓
4. 合并新旧设置
   ↓
5. 如果开启分批，验证至少有一个分批类型
   ↓
6. 保存设置
   ↓
7. 删除缓存
   ↓
8. 推送配置更新
   ↓
9. 删除非当前安全库存类型的预警记录
```

#### 特殊处理

```go
// 菜品卡片样式更新时间
if oldDishCardStyle != businessSetting.DishCardStyle {
    businessSetting.DishCardStyleTime = strconv.Itoa(int(time.Now().Unix()))
}

// 外送功能权限验证
if businessSetting.DeliveryPriceRatio != businessSettingReq.DeliveryPriceRatio && 
   companySetting.DeliveryStatus != 1 {
    return errors.New("当前没有权限使用此功能")
}

// 分批类型验证
if businessSetting.IsBatch == "1" {
    batchTagNum, _ := repository.NewBatchTagRepo(db).GetBatchTagCount()
    if batchTagNum == 0 {
        return errors.New("开启分批送厨商品时，必须至少有一个分批类型")
    }
}
```

---

## 🔐 密码验证功能

### 1. 通用密码验证 (VerifyPassword)

**功能描述**: 验证各端的密码（钱箱密码、高级密码、锁屏密码）。

#### 密码类型

| 常量 | 说明 |
|-----|------|
| `PasswordTypeCashBox` | 钱箱密码 |
| `PasswordTypeAdvanced` | 高级设置密码 |
| `PasswordTypeLock` | 锁屏密码 |

#### 支持的来源

| 来源 | 支持的密码类型 |
|-----|--------------|
| `SourceCashier` | 钱箱密码、高级密码、锁屏密码 |
| `SourceAssistant` | 高级密码、锁屏密码 |
| `SourceTablet` | 高级密码 |
| `SourceKitchen` | 高级密码 |

#### 实现逻辑

```go
func (s *Srv) VerifyPassword(ctx context.Context, source string, typ string, password string) bool {
    passwordMap := make(map[string]string)
    
    switch source {
    case constant.SourceCashier:
        cashierSetting, _ := s.GetCashierSetting(ctx, nil)
        passwordMap = map[string]string{
            constant.PasswordTypeCashBox:  cashierSetting.CashierPassword,
            constant.PasswordTypeAdvanced: cashierSetting.AdvancedPassword,
            constant.PasswordTypeLock:     cashierSetting.LockPassword,
        }
    case constant.SourceAssistant:
        assistantSetting, _ := s.GetAssistantSetting(ctx, nil)
        passwordMap = map[string]string{
            constant.PasswordTypeAdvanced: assistantSetting.AdvancedPassword,
            constant.PasswordTypeLock:     assistantSetting.LockPassword,
        }
    // ... 其他来源
    }
    
    if truePassword, exits := passwordMap[typ]; exits {
        return password == truePassword
    }
    return false
}
```

---

### 2. 高级密码验证 (VerifyAdvancedPassword)

**功能描述**: 验证高级密码，支持可选的验证场景。

#### 选项模式

```go
type VerifyAdvancedPasswordOption struct {
    IsAssistantCheckOrder bool  // 是否是助手端下单校验
}

func WithIsAssistantCheckOrder() func(option *VerifyAdvancedPasswordOption) {
    return func(option *VerifyAdvancedPasswordOption) {
        option.IsAssistantCheckOrder = true
    }
}
```

#### 使用示例

```go
// 场景1：助手端下单校验
err := settingSrv.VerifyAdvancedPassword(ctx, password, WithIsAssistantCheckOrder())

// 场景2：取消订单/退菜校验
err := settingSrv.VerifyAdvancedPassword(ctx, password)
```

#### 验证逻辑

```go
func (s *Srv) VerifyAdvancedPassword(ctx context.Context, password string, options ...func(option *VerifyAdvancedPasswordOption)) error {
    option := &VerifyAdvancedPasswordOption{}
    for _, optionFunc := range options {
        optionFunc(option)
    }
    
    // 助手端下单校验
    if option.IsAssistantCheckOrder {
        assistant, _ := s.GetAssistantSetting(ctx, nil)
        if assistant.IsCheckOrderPassword() {
            if password == "" {
                return errors.New("请输入确认密码")
            }
            if password != assistant.AdvancedPassword {
                return errors.New("确认密码错误")
            }
        }
        return nil
    }
    
    // 取消订单/退菜校验
    businessSetting, _ := s.GetBusinessSetting(ctx)
    if businessSetting.IsNeedPassword == "1" {
        if password == "" {
            return errors.New("请输入确认密码")
        }
        
        switch ctx.GetSource() {
        case constant.SourceCashier:
            cashier, _ := s.GetCashierSetting(ctx, nil)
            if password != cashier.AdvancedPassword {
                return errors.New("确认密码错误")
            }
        case constant.SourceAssistant:
            assistant, _ := s.GetAssistantSetting(ctx, nil)
            if password != assistant.AdvancedPassword {
                return errors.New("确认密码错误")
            }
        // ... 其他来源
        }
    }
    return nil
}
```

---

## 🛠️ 辅助功能

### 1. 检查更新 (CheckUpdate)

**功能描述**: 检查客户端应用是否有新版本。

#### 请求参数

| 参数 | 说明 |
|-----|------|
| `appType` | 应用类型（收银、助手、平板等） |
| `brand` | 设备品牌 |
| `language` | 语言 |

#### 返回数据

```go
type UpdateInfo struct {
    VersionName  string  // 版本号
    ForcedUpdate int     // 是否强制更新
    UpdateLog    string  // 更新日志（国际化）
    DownloadURL  string  // 下载地址
}
```

#### 平台限制

```go
userAgent := ctx.GetGin().GetHeader("User-Agent") + ";" + ctx.GetGin().GetHeader("platform")
if utils.GetPlatform(userAgent) != 1 {  // 1=Android
    return resp.UpdateInfo{}, errors.NewWithCode(constant.CodeSystemError, "当前平台暂不支持应用内更新")
}
```

#### 更新日志国际化

```go
var updateLogMultilanguage dto.LocaleResponse
if updateData.Data.UpdateLog != "" {
    err := json.Unmarshal([]byte(updateData.Data.UpdateLog), &updateLogMultilanguage)
    if err != nil {
        updateLog = updateData.Data.UpdateLog  // 单语言
    } else {
        updateLog = updateLogMultilanguage.GetLocale(language)  // 多语言
    }
}
```

---

### 2. 获取电子菜单二维码 (GetMenuQrcode)

**功能描述**: 生成电子菜单访问链接。

#### 生成逻辑

```go
func (s *Srv) GetMenuQrcode(ctx context.Context) (string, error) {
    businessSetting, _ := s.GetBusinessSetting(ctx)
    token := s.getMenuQrcodeToken(ctx, businessSetting)
    return viper.GetString("MENU_BASE_URL") + "/home?token=" + token, nil
}

func (s *Srv) getMenuQrcodeToken(ctx context.Context, businessSetting setting.Business) string {
    type Qrcode struct {
        CompanyUuid uint64 `json:"a"`
        Qrcode      string `json:"q"`
    }
    qrcode := Qrcode{
        CompanyUuid: ctx.GetCompanyUuid(),
        Qrcode:      businessSetting.QrCode,
    }
    qrcodeString := utils.ToJson(qrcode)
    hash := md5.Sum([]byte(qrcodeString))
    token := fmt.Sprintf("%x.%s", hash, base64.StdEncoding.EncodeToString([]byte(qrcodeString)))
    
    return base64.StdEncoding.EncodeToString([]byte(token))
}
```

#### Token结构

```
1. 构造数据: {CompanyUuid, QrCode}
   ↓
2. JSON序列化
   ↓
3. 计算MD5哈希
   ↓
4. 组合: hash.base64(json)
   ↓
5. Base64编码整体
```

---

### 3. 获取支付方式列表 (GetPaymentMethodList)

**功能描述**: 获取店铺配置的支付方式。

#### 返回数据

```go
type PaymentMethodListResp struct {
    List []PaymentMethod
}

type PaymentMethod struct {
    Uuid uint64  // 支付方式UUID
    Name string  // 支付方式名称
}
```

#### 查询逻辑

```go
paymentMethodList := paymentRepo.GetPaymentMethodList(
    commonRepo.WhereBySoftDelete(),       // 排除已删除
    commonRepo.SortWithSort("asc"),       // 按sort排序
    commonRepo.SortWithCreateTime("desc"), // 按创建时间排序
)
```

---

### 4. 货币符号位置格式化 (SymbolPosition)

**功能描述**: 根据配置格式化金额显示。

#### 使用示例

```go
// UnitPosition="0" (符号在前)
result := settingSrv.SymbolPosition(ctx, 100.00)
// 输出: ฿ 100.00

// UnitPosition="1" (符号在后)
result := settingSrv.SymbolPosition(ctx, 100.00)
// 输出: 100.00 ฿
```

---

### 5. 获取收银端基础设置 (GetCashierBaseSetting)

**功能描述**: 获取收银端所有基础配置的聚合视图。

#### 返回数据

```go
type CashierBaseSetting struct {
    AcceptOrder       AcceptOrderSetting       // 接单设置
    AcceptMemberOrder AcceptMemberOrderSetting // 会员订单接单设置
    System            SystemSetting            // 系统设置
    UsbPrinter        UsbPrinterList           // USB打印机列表
}

type SystemSetting struct {
    IsShowScanSoldOut      int     // H5端显示售罄
    IsShowAssistantSoldOut int     // 助手端显示售罄
    MenuShowSoldOut        int     // 菜单显示售罄
    MemberShowSoldOut      int     // 会员端显示售罄
    DishCardStyle          string  // 菜品卡片样式
    IsShowSoldOut          int     // 平板端显示售罄
    DefaultLanguage        string  // 默认语言
    SecondLanguage         string  // 第二语言
    DeviceId               string  // 设备ID
    DeviceRemark           string  // 设备备注
    ClientVersion          string  // 客户端版本
    ServerVersion          string  // 服务端版本
}
```

---

### 6. 获取商家业务设置 (GetShopBusinessSetting)

**功能描述**: 获取移动管理端的业务设置视图。

#### 返回数据

```go
type ShopBusiness struct {
    Business                                 Business  // 业务设置
    FreeReasonCount                          int       // 免单原因数量
    ReturnFoodReasonCount                    int       // 退菜原因数量
    OrderRemarkCount                         int       // 订单备注数量
    HeadquarterRequiredParentCompanyApproval string    // 总部-上级审批规则
    HeadquarterViaParentCompanyWarehouse     string    // 总部-上级仓库规则
}
```

#### 统计查询

```go
// 统计免单原因数量
db.Model(&model.FreeReason{}).Scopes(repository.NotDeleted).Select("count(*)").Scan(&freeReasonCount)

// 统计退菜原因数量
db.Model(&model.ReturnFoodReason{}).Scopes(repository.NotDeleted).Select("count(*)").Scan(&returnFoodReasonCount)

// 统计订单备注数量
db.Model(&model.OrderRemark{}).Scopes(repository.NotDeleted).Select("count(*)").Scan(&orderRemarkCount)
```

#### 总部规则获取

```go
if companySetting.HeadquarterUuid > 0 {
    ctx2 := ctx.Copy()
    ctx2.SetCompanyUuid(companySetting.HeadquarterUuid)
    headquarterBusinessSetting, _ := s.GetBusinessSetting(ctx2)
    
    headquarterRequiredParentCompanyApproval = headquarterBusinessSetting.RequiredParentCompanyApproval
    headquarterViaParentCompanyWarehouse = headquarterBusinessSetting.ViaParentCompanyWarehouse
}
```

---

## 📊 默认配置值

### 默认配置文件 (default.go)

该文件定义了所有设置的默认值，当数据库中没有配置时使用。

#### 主要默认配置方法

| 方法 | 说明 |
|-----|------|
| `getDefaultCashier()` | 收银机默认设置 |
| `getDefaultTablet()` | 平板默认设置 |
| `getDefaultH5()` | H5默认设置 |
| `getDefaultKitchen()` | 厨显默认设置 |
| `getDefaultAssistant()` | 助手默认设置 |
| `getDefaultBuffet()` | 自助餐默认设置 |
| `getDefaultPayment()` | 支付方式默认设置 |
| `getDefaultBusiness()` | 业务规则默认设置 |
| `getDefaultStore()` | 店铺默认设置 |
| `getDefaultPrinter()` | 打印机默认设置 |
| `getDefaultPoints()` | 积分默认设置 |
| `getDefaultCurrency()` | 货币默认设置 |
| `getDefaultTaxRate()` | 税率默认设置 |
| `getDefaultServiceCharge()` | 服务费默认设置 |

#### 默认密码

所有端的默认密码统一为: `666888`

```go
AdvancedPassword: "666888"  // 高级设置密码
CashierPassword:  "666888"  // 钱箱密码
LockPassword:     "666888"  // 锁屏密码
```

---

## 🔄 数据处理

### 1. JSON字段类型转换

#### 收银机设置解析

```go
func (s *Srv) parseCashierSetting(values string) ([]byte, error) {
    var jsonMap map[string]interface{}
    json.Unmarshal([]byte(values), &jsonMap)
    
    // is_show_scan_sold_out: float64/string → int
    if isShowScanSoldOut, ok := jsonMap["is_show_scan_sold_out"]; ok {
        if numVal, ok := isShowScanSoldOut.(float64); ok {
            jsonMap["is_show_scan_sold_out"] = int(numVal)
        } else if strVal, ok := isShowScanSoldOut.(string); ok {
            jsonMap["is_show_scan_sold_out"], _ = strconv.Atoi(strVal)
        }
    }
    
    // is_auto_order: any → string
    if isAutoOrder, ok := jsonMap["is_auto_order"]; ok {
        switch v := isAutoOrder.(type) {
        case float64:
            jsonMap["is_auto_order"] = strconv.Itoa(int(v))
        case int:
            jsonMap["is_auto_order"] = strconv.Itoa(v)
        case string:
            jsonMap["is_auto_order"] = v
        default:
            jsonMap["is_auto_order"] = "0"
        }
    }
    
    return json.Marshal(jsonMap)
}
```

#### 平板设置转换

```go
func (s *Srv) convertTabletFormat(oldVal string) ([]byte, error) {
    tabletMap := map[string]any{}
    json.Unmarshal([]byte(oldVal), &tabletMap)
    
    // float64 → string
    if v, ok := tabletMap["is_call_service"].(float64); ok {
        tabletMap["is_call_service"] = fmt.Sprintf("%.0f", v)
    }
    
    // []any → struct{}
    if v, ok := tabletMap["buffet_order_limit"].([]any); ok {
        tabletMap["buffet_order_limit"] = struct{}{}
    }
    
    return json.Marshal(tabletMap)
}
```

---

### 2. 语言列表处理

#### 语言过滤和去重

```go
validLanguageList := make([]dto.LanguageItem, 0)
var languageNames []string

for _, item := range cashierSetting.LanguageList {
    // 只保留在Language中的语言，且去重
    if slices.Contains(cashierSetting.Language, item.Name) && 
       !slices.Contains(languageNames, item.Name) {
        validLanguageList = append(validLanguageList, item)
        languageNames = append(languageNames, item.Name)
    }
}

cashierSetting.LanguageList = validLanguageList
```

---

### 3. 图片域名处理

#### 添加域名前缀

```go
// Logo图片
if defaultStore.LogoURL != "" && ginContext != nil {
    defaultStore.LogoURL = utils.GetBaseURL(ginContext.Request) + defaultStore.LogoURL
}

// 轮播图
if len(cashier.Carousel) > 0 && ginContext != nil {
    for i, item := range cashier.Carousel {
        cashier.Carousel[i].FilePath = utils.AddImageDomain(
            item.FilePath, 
            utils.GetBaseURL(ginContext.Request), 
            true
        )
    }
}
```

---

## 🎯 最佳实践

### 1. 获取设置的正确姿势

```go
// ✅ 正确：使用Setting服务获取
settingSrv := setting.NewSrv(dbm, cache)
cashierSetting, err := settingSrv.GetCashierSetting(ctx, nil)

// ❌ 错误：直接查询数据库
settingRepo.GetByKey(constant.SettingCashier)
```

### 2. 更新设置后清理缓存

```go
// ✅ 正确：UpdateSetting会自动清理缓存
err := settingSrv.UpdateSetting(ctx, constant.SettingCashier, cashierSetting)

// ✅ 正确：手动更新后清理缓存
settingRepo.Updates(key, value)
s.cache.Del(fmt.Sprintf(s.cacheKey, companyUuid))
```

### 3. 语言列表传递

```go
// ✅ 正确：传nil自动获取商家语言
cashierSetting, _ := settingSrv.GetCashierSetting(ctx, nil)

// ✅ 正确：传递指定语言列表
languageList := []dto.LanguageItem{{Name: "th", Value: "泰语"}}
cashierSetting, _ := settingSrv.GetCashierSetting(ctx, languageList)
```

### 4. 合并默认值

```go
// ✅ 正确：使用copier合并，IgnoreEmpty保留已有值
defaultCashier := s.getDefaultCashier(languageList)
copier.CopyWithOption(&defaultCashier, cashier, copier.Option{
    IgnoreEmpty: true,
    DeepCopy: true,
})
```

### 5. Context使用

```go
// ✅ 正确：在协程中使用副本
ctx2 := ctx.Copy()
go func() {
    headquarterBusinessSetting, _ := s.GetBusinessSetting(ctx2)
}()

// ❌ 错误：直接使用原context
go func() {
    s.GetBusinessSetting(ctx)  // 可能有并发问题
}()
```

---

## ⚠️ 注意事项

### 1. 字段类型兼容性

- 数据库中存储的JSON字段类型可能不一致（字符串/数字）
- 使用类型转换函数统一处理
- 新版本应统一字段类型

### 2. 缓存失效

- 更新设置后必须删除缓存
- 使用 `UpdateSetting` 方法会自动处理
- 直接操作数据库需手动清理缓存

### 3. 语言配置联动

- 修改店铺语言时，需同步更新所有端的语言配置
- 确保每个端至少有一种语言
- 验证默认语言在语言列表中

### 4. 权限控制

- 部分功能受公司设置控制（如自助餐、厨显）
- 修改设置时需验证权限
- 会员关闭时，余额支付自动关闭

### 5. 事务处理

- 涉及多表更新的操作使用事务
- `EditStoreSetting` 需要更新多个表
- 事务失败时自动回滚

---

## 📚 相关文档

- [缓存服务](../cache/cache_service.md)
- [国际化服务](../i18n/i18n.md)
- [WebSocket推送](../websocket/websocket_service.md)
- [图片处理工具](../utils/image_utils.md)

---

## 📄 更新日志

| 日期 | 版本 | 说明 |
|-----|------|-----|
| 2025-11-12 | 1.0 | 初始文档创建 |

---

## 👥 维护者

- 开发团队：Backend Team
- 文档维护：AI Assistant

---

**注意**: 本文档基于代码自动生成，如有代码变更，请及时更新文档。设置服务是系统的核心基础服务，修改时需特别谨慎，建议充分测试后再上线。

