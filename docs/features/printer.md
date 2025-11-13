# Printer Service 打印服务说明文档

## 📋 概述

打印服务是 TTPOS 系统的核心功能之一，负责处理各类小票打印、模板管理、打印机管理等功能。该服务支持多种打印机类型（USB、网络、云打印），支持文本打印和图片打印两种方式，并提供灵活的模板定制功能。

**主要服务文件**:
- `/main/app/service/printer.go` - 打印机管理服务（1066行）
- `/main/app/printer/service/printer_log.go` - 打印日志服务（661行）
- `/main/app/printer/base.go` - 打印核心接口（356行）

**接口定义**: 
- `IPrinterSrv` - 打印机管理服务接口
- `IPrinterLogSrv` - 打印日志服务接口
- `PPrinterRepo` - 打印仓库接口

---

## 🏗️ 架构设计

### 接口定义

#### IPrinterSrv - 打印机管理服务

```go
type IPrinterSrv interface {
    // 打印机管理
    GetProductPrinterList(ctx context.Context) (resp.ProductPrinterList, error)                    // 获取打印档口列表
    UsbPrinterReport(ctx context.Context, reportReq req.UsbPrinterReportReq) (resp.PrinterReportResp, error) // USB打印机上报
    
    // 模板管理
    GetPrintTemplateList(ctx context.Context) (resp.PrintTemplateListResp, error)                  // 获取打印模板列表
    GetPrintTemplateDetail(ctx context.Context, id uint64) (resp.PrintTemplateDetailResp, error)   // 获取模板详情
    
    // 模板定制
    CreatePrinterCustomize(ctx context.Context, req req.CreatePrinterCustomizeReq) (resp.CreatePrinterCustomizeResp, error) // 创建打印机定制
    EditPrinterCustomize(ctx context.Context, req req.EditPrinterCustomizeReq) error              // 编辑打印机定制
    PreviewPrinterCustomize(ctx context.Context, req req.PreviewPrinterCustomizeReq) (resp.PreviewPrinterCustomizeResp, error) // 预览打印机定制
    DeletePrinterCustomize(ctx context.Context, customizeUuid uint64) error                       // 删除打印机定制
    UsePrinterCustomize(ctx context.Context, customizeUuid uint64) error                         // 使用打印机定制
    GetPrinterCustomizeConfigInfo(ctx context.Context, req req.PrinterGetConfigInfoReq) (resp.ConfigInfoResp, error) // 获取配置信息
}
```

#### IPrinterLogSrv - 打印日志服务

```go
type IPrinterLogSrv interface {
    // 打印日志管理
    AddLog(ctx context.Context, printer resp.PrinterInfo, printerLogData model.PrinterLog, controlDeviceId string) (model.PrinterLog, error) // 添加打印日志
    GetPrinterBase(ctx context.Context) (*resp.PrinterBaseResp, error)                                                                     // 获取基础数据
    GetPrinterLogList(ctx context.Context, req req.PrinterListReq) (*resp.PrinterListPaginationResp, error)                            // 获取打印列表
    GetPrinterData(ctx context.Context) (*resp.PrinterDataList, error)                                                                   // 获取打印数据
    PrinterPrint(ctx context.Context, req req.PrinterPrintReq) (*resp.PrinterData, error)                                                 // 打印
    PrinterReport(ctx context.Context, req req.PrinterReportReqs) error                                                                  // 打印报告
    
    // 特殊打印配置
    GetStaticOpenCashBoxPrinterConfig(ctx context.Context) (*resp.PrinterData, error)                                                    // 获取静态打开钱箱配置
    GetOldOrderPrinterConfig(ctx context.Context, data string) (*resp.PrinterData, error)                                               // 获取旧订单打印配置
}
```

#### PPrinterRepo - 打印仓库接口

```go
type PPrinterRepo interface {
    // 打印功能
    PrintingDishes(printType int, saleBillUuid uint64, saleOrderUuid uint64, products printer_model.Products) bool
    PrintingStatementOrder(printType int, saleBill *model.SaleBill, saleOrderUuid uint64, firstExecution int, payMethodUuid uint64) (*resp.PrinterData, error)
    PrintingInvoice(saleBill *model.SaleBill, saleOrderUuid uint64, firstExecution int) (*resp.PrinterData, error)
    PrintingRechargeOrder(order model.MemberRechargeOrder, firstExecution int) (*resp.PrinterData, error)
    PrintingHandoverOrder(log *model.StaffShiftLog, businessData *business_data_resp.BusinessDataAll, firstExecution int, openMoneybox bool, deviceSnId ...string) (*resp.PrinterData, error)
    PrintingBusinessData(businessData *template.PrintingBusinessData, startTime int64, endTime int64, firstExecution int) (*resp.PrinterData, error)
    PrintingTakeoutOrder(memberSaleOrder *model.MemberSaleOrder, saleBill *model.SaleBill, saleOrderUuid uint64) (*resp.PrinterData, error)
    
    // 缓存管理
    DeleteProductPrinterListCache()     // 删除商品打印机列表缓存
    SetFinishedTime(finishedTime int64) // 设置完成时间
    GetFinishedTime() int64             // 获取完成时间
    SetPrinterWidth(printerWidth int)   // 设置打印机宽度
    GetPrinterWidth() int               // 获取打印机宽度
    Is58mmPrinter() bool                // 是否58mm打印机
}
```

### 依赖服务

```go
type printerSrv struct {
    dbm   *database.DBManager  // 数据库管理器
    cache cache.Cache          // 缓存
}

type printerLogSrv struct {
    dbm        *database.DBManager  // 数据库管理器
    settingSrv setting.ISrv         // 设置服务
}
```

---

## 🎯 核心功能

### 1. 打印类型

系统支持多种打印类型，每种类型对应不同的业务场景：

| 模板常量 | 说明 | 使用场景 |
|---------|------|---------|
| `PrinterTemplatePreBilling = 3` | 预结账单 | 订单预结时打印 |
| `PrinterTemplateBilling = 2` | 结账单 | 订单结算时打印 |
| `PrinterTemplateInvoice = 7` | 发票 | 开具发票时打印 |
| `PrinterTemplateRecharge = 8` | 充值单 | 会员充值时打印 |
| `PrinterTemplateHandoverSheet = 1` | 交班单 | 员工交班时打印 |
| `PrinterTemplateBusiness = 5` | 营业数据 | 查看营业数据时打印 |
| `PrinterTemplateOneDishOneMenu = 4` | 一菜一单 | 厨房打印，每个菜品一张小票 |
| `PrinterTemplateEntireOrder = 6` | 整单打印 | 厨房打印，整单一张小票 |
| `PrinterTemplateReturnDish = 9` | 退菜单 | 退菜时打印 |
| `PrinterTemplateOutMenu = 10` | 出菜单 | 出菜时打印 |
| `PrinterTemplateTakeoutOrder = 11` | 外送单 | 外送订单打印 |

### 2. 打印机类型

系统支持多种打印机品牌和连接方式：

| 打印机类型常量 | 说明 | 连接方式 |
|--------------|------|---------|
| `PrinterTypeFeiEYun` | 飞鹅打印机 | 云打印 |
| `PrinterTypeFeiEYunTag` | 飞鹅标签打印机 | 云打印 |
| `PrinterTypePrintCenter` | 365云打印 | 云打印 |
| `PrinterTypeSunmiLan` | 商米局域网打印 | 局域网 |
| `PrinterTypeSunmiCloud` | 商米云打印 | 云打印 |
| `PrinterTypeXPrinterLan` | 芯烨有线 | 有线网络 |
| `PrinterTypeXPrinterWifi` | 芯烨WIFI | 无线网络 |
| `PrinterTypeCodesoftLan` | Codesoft网口 | 有线网络 |
| `PrinterTypeCodesoftWifi` | Codesoft WIFI | 无线网络 |
| `PrinterTypeGpCloud` | 佳博云打印 | 云打印 |
| `PrinterTypeCashierCompax` | Compax收银打印机 | 内置 |
| `PrinterTypeCashierSunmi` | 商米收银打印机 | 内置 |
| `PrinterTypeCashierImmin` | IMIN收银打印机 | 内置 |

### 3. 打印方式

系统支持两种打印方式：

| 打印方式常量 | 说明 | 特点 |
|------------|------|-----|
| `PrinterLogPrintMethodText = 1` | 文本打印 | 速度快，兼容性好，但格式受限 |
| `PrinterLogPrintMethodImage = 2` | 图片打印 | 格式灵活，支持复杂布局，但速度较慢 |

### 4. 打印速度模式

系统支持三种打印速度模式，适用于不同的网络环境：

| 速度模式 | 常量值 | 说明 | 分片大小 |
|---------|--------|------|---------|
| 流畅模式 | `PrintSpeed = 1` | 不分片打印，适合稳定网络 | 5MB |
| 稳定模式 | `PrintSpeed = 2` | 分片大包打印，默认模式 | 20KB |
| 兼容模式 | `PrintSpeed = 3` | 分片小包打印，适合不稳定网络 | 4KB |

### 5. USB打印机管理 (UsbPrinterReport)

**功能描述**: 处理USB打印机的自动发现和状态管理。

#### 处理流程

```
1. 获取已有USB打印机列表
   ↓
2. 更新打印机状态
   - 检查上报列表中的打印机
   - 更新在线状态和心跳时间
   - 将不在列表中的打印机标记为离线
   ↓
3. 处理新打印机
   - 识别打印机类型（根据VID/PID）
   - 创建打印机记录
   - 自动关联到收银机
   ↓
4. 更新打印设置
   - 自动选择打印机
   - 更新收银机打印机配置
```

#### 打印机类型识别

系统根据USB打印机的VID和PID自动识别打印机类型：

```go
// VID=1137, PID=85 的打印机
if usbPrinter.Vid == 1137 && usbPrinter.Pid == 85 {
    if usbPrinter.M_name == "Zhuhai Howbest Label Printer Co.,Ltd." {
        printerTypeKey = constant.PRINTER_TYPE_GP_C200IV
    } else if usbPrinter.M_name == "ZHU HAI HOWBEST Receipt Printer Co.,Ltd." {
        printerTypeKey = constant.PRINTER_TYPE_GP_D300I
    }
}
```

#### 状态管理

- **在线状态**: 打印机在上报列表中且状态为1
- **离线状态**: 打印机不在上报列表中或状态为0
- **心跳时间**: 每次上报时更新 `last_heartbeat_time`

### 6. 打印模板管理

#### 获取模板列表 (GetPrintTemplateList)

**功能描述**: 获取系统支持的打印模板列表，按分组返回。

**模板分组**:
- **收银小票组** (`GroupType = 1`): 预结账单、结账单
- **厨房小票组** (`GroupType = 2`): 一菜一单、整单打印、退菜单、出菜单

**返回结构**:
```go
type PrintTemplateListResp struct {
    List []PrintTemplateGroup `json:"list"` // 模板分组列表
}

type PrintTemplateGroup struct {
    LocaleName dto.LocaleResponse `json:"locale_name"` // 分组名称（多语言）
    GroupType  int                `json:"group_type"`  // 分组类型
    List       []PrintTemplate    `json:"list"`        // 模板列表
}
```

#### 获取模板详情 (GetPrintTemplateDetail)

**功能描述**: 获取指定模板的详细信息，包括默认模板和高级模板列表。

**返回结构**:
```go
type PrintTemplateDetailResp struct {
    DefaultTpl      PrintTemplateDetail   `json:"default_tpl"`        // 默认模板
    AdvReceiptTpls  []PrintTemplateDetail `json:"adv_receipt_tpls"`   // 高级模板列表
    IsAdvReceiptTpl bool                  `json:"is_adv_receipt_tpl"` // 是否启用高级模板
}
```

**处理流程**:
1. 获取模板基本信息
2. 查询打印机定制列表
3. 解析默认模板（如果不存在则创建）
4. 解析高级模板列表
5. 返回模板详情

### 7. 模板定制功能

#### 创建打印机定制 (CreatePrinterCustomize)

**功能描述**: 创建高级打印模板定制。

**前置条件**:
- 必须开启高级模板打印功能 (`IsOpenAdvancedTicketPrint == 1`)
- 模板名称不能重复

**处理流程**:
```
1. 验证是否开启高级模板
   ↓
2. 验证模板是否存在
   ↓
3. 检查模板名称是否重复
   ↓
4. 生成定制UUID
   ↓
5. 创建打印机定制记录
   - Name: 模板名称
   - Data: 模板JSON数据
   - TemplateId: 模板ID
   - IsAdv: 1 (高级模板)
```

#### 编辑打印机定制 (EditPrinterCustomize)

**功能描述**: 编辑已存在的打印机定制。

**处理流程**:
```
1. 检查定制是否存在
   ↓
2. 验证模板数据（解析测试）
   ↓
3. 更新定制信息
   - 高级模板：更新名称和数据
   - 默认模板：只更新数据
   ↓
4. 如果正在使用，同步更新模板
```

#### 预览打印机定制 (PreviewPrinterCustomize)

**功能描述**: 预览模板定制效果，返回生成的图片URL。

**处理流程**:
```
1. 获取模板信息
   ↓
2. 获取测试数据
   ↓
3. 解析模板生成图片
   ↓
4. 返回Base64编码的图片URL
```

#### 使用打印机定制 (UsePrinterCustomize)

**功能描述**: 将指定的定制模板设置为当前使用的模板。

**处理流程**:
```
1. 检查定制是否存在
   ↓
2. 将同模板的其他定制设为未使用
   ↓
3. 将当前定制设为使用中
   ↓
4. 更新模板的定制UUID和数据
```

#### 删除打印机定制 (DeletePrinterCustomize)

**功能描述**: 删除高级模板定制。

**限制条件**:
- 只能删除高级模板（`IsAdv == 1`）
- 不能删除正在使用的模板（`IsUse == 1`）

### 8. 打印日志管理

#### 添加打印日志 (AddLog)

**功能描述**: 创建打印任务并添加到打印队列。

**处理流程**:
```
1. 设置初始状态为"进行中"
   ↓
2. 判断打印类型
   - 商米云打印 + 未开启本地打印 → 队列打印
   - 其他情况 → 本地打印
   ↓
3. 保存打印日志数据（压缩存储）
   ↓
4. 保存打印日志记录
   ↓
5. 提交事务
   ↓
6. 加入打印队列（异步执行）
```

**数据压缩**:
- 打印数据使用GZIP压缩后Base64编码存储
- 格式: `GZIP:{base64_encoded_data}`

#### 获取打印数据 (GetPrinterData)

**功能描述**: 获取设备待打印的数据列表。

**处理流程**:
```
1. 加锁防止并发操作
   ↓
2. 验证设备（网页版设备不能获取）
   ↓
3. 查询待打印日志
   ↓
4. 转换为打印数据格式
   - 解压数据
   - 组装打印机配置
   - 计算打印耗时
   ↓
5. 返回打印数据列表
```

**返回结构**:
```go
type PrinterData struct {
    Uuid              uint64 `json:"uuid"`                // 打印日志UUID
    Data              string `json:"data"`                // 打印数据（压缩）
    PrintMethod       int    `json:"print_method"`        // 打印方式
    Copies            uint   `json:"copies"`             // 打印份数
    PrinterType       string `json:"printer_type"`        // 打印机类型
    PrinterConfig     string `json:"printer_config"`      // 打印机配置
    IsCashierPrinter  bool   `json:"is_cashier_printer"`  // 是否收银打印机
    IsUsbPrinter      bool   `json:"is_usb_printer"`      // 是否USB打印机
    PrintingTime      int64  `json:"printing_time"`       // 打印耗时（毫秒）
    EnableStatusCheck int    `json:"enable_status_check"` // 是否启用状态检查
    TradeNo           string `json:"trade_no"`            // 交易号
    PrintChunkSize    int    `json:"print_chunk_size"`    // 打印分片大小
}
```

#### 打印报告 (PrinterReport)

**功能描述**: 客户端上报打印结果（成功/失败）。

**处理流程**:
```
1. 验证请求参数
   ↓
2. 批量查询打印日志
   ↓
3. 更新打印状态
   - 成功: Status = 2, Num++
   - 失败: Status = 0, Num++, 记录失败原因
   ↓
4. 批量更新数据库
```

**状态说明**:
- `Status = 0`: 打印失败
- `Status = 1`: 进行中
- `Status = 2`: 打印成功
- `Num > 1`: 补打

### 9. 打印功能实现

#### 打印结账单 (PrintingStatementOrder)

**功能描述**: 打印订单结账单（预结账单或结账单）。

**参数说明**:
- `printType`: 打印类型（2=结账单, 3=预结账单）
- `saleBill`: 销售账单
- `saleOrderUuid`: 销售订单UUID
- `firstExecution`: 是否首次执行（1=是, 0=否）
- `payMethodUuid`: 支付方式UUID

**处理流程**:
```
1. 获取打印模板
   ↓
2. 构建打印数据
   - 订单信息
   - 商品列表
   - 支付信息
   - 门店信息
   ↓
3. 根据打印方式生成数据
   - 文本打印: 生成ESC/POS指令
   - 图片打印: 生成图片数据
   ↓
4. 创建打印日志
   ↓
5. 返回打印数据
```

#### 打印厨房小票 (PrintingDishes)

**功能描述**: 打印厨房小票（一菜一单或整单打印）。

**参数说明**:
- `printType`: 打印类型（1=送厨打印, 0=付款打印, -1=退菜打印, -2=出菜单打印）
- `saleBillUuid`: 销售账单UUID
- `saleOrderUuid`: 销售订单UUID
- `products`: 商品列表

**处理流程**:
```
1. 获取商品打印机列表（缓存）
   ↓
2. 根据打印规则筛选打印机
   - 按区域筛选
   - 按商品分类筛选
   - 按打印标签筛选
   ↓
3. 遍历打印机列表
   ↓
4. 根据打印模式生成小票
   - 整单打印: 一张小票包含所有商品
   - 一菜一单: 每个商品一张小票
   ↓
5. 创建打印日志
   ↓
6. 返回打印结果
```

#### 打印发票 (PrintingInvoice)

**功能描述**: 打印订单发票。

**处理流程**:
```
1. 获取发票模板
   ↓
2. 构建发票数据
   - 发票抬头
   - 订单信息
   - 商品明细
   - 税额信息
   ↓
3. 生成打印数据
   ↓
4. 创建打印日志
```

#### 打印充值单 (PrintingRechargeOrder)

**功能描述**: 打印会员充值单。

**处理流程**:
```
1. 获取充值单模板
   ↓
2. 构建充值数据
   - 会员信息
   - 充值金额
   - 充值方式
   - 门店信息
   ↓
3. 生成打印数据
   ↓
4. 创建打印日志
```

#### 打印交班单 (PrintingHandoverOrder)

**功能描述**: 打印员工交班单。

**参数说明**:
- `log`: 交班日志
- `businessData`: 营业数据
- `firstExecution`: 是否首次执行
- `openMoneybox`: 是否打开钱箱
- `deviceSnId`: 设备SN列表

**处理流程**:
```
1. 获取交班单模板
   ↓
2. 构建交班数据
   - 交班员工信息
   - 营业数据统计
   - 支付方式统计
   - 现金盘点
   ↓
3. 生成打印数据
   ↓
4. 创建打印日志
   ↓
5. 如果需要，添加开钱箱指令
```

#### 打印营业数据 (PrintingBusinessData)

**功能描述**: 打印营业数据报表。

**处理流程**:
```
1. 获取营业数据模板
   ↓
2. 构建报表数据
   - 时间范围
   - 销售统计
   - 商品统计
   - 支付统计
   ↓
3. 生成打印数据
   ↓
4. 创建打印日志
```

---

## 🔄 数据流转

### 打印任务流转

```
业务触发打印
    ↓
创建打印日志（状态：进行中）
    ↓
保存打印数据（压缩存储）
    ↓
加入打印队列
    ↓
客户端拉取打印数据
    ↓
执行打印
    ↓
上报打印结果
    ↓
更新打印状态（成功/失败）
```

### 模板定制流转

```
获取模板列表
    ↓
选择模板
    ↓
获取模板详情
    ↓
编辑模板（可选）
    ↓
预览模板
    ↓
保存定制
    ↓
使用定制
    ↓
更新模板配置
```

---

## 🔐 权限控制

### 打印模板管理

- **查看模板列表**: 所有登录用户
- **查看模板详情**: 所有登录用户
- **编辑模板**: 需要相应权限
- **创建高级模板**: 需要开启高级模板功能

### 打印日志管理

- **查看打印日志**: 需要相应权限
- **补打**: 需要相应权限
- **删除日志**: 需要管理员权限

---

## ⚠️ 错误处理

### 常见错误

| 错误信息 | 原因 | 解决方案 |
|---------|------|---------|
| "打印机不存在" | 打印机已被删除 | 检查打印机配置 |
| "打印数据为空" | 模板解析失败 | 检查模板数据 |
| "打印机类型为空" | 打印机类型未配置 | 配置打印机类型 |
| "模板解析失败" | 模板JSON格式错误 | 检查模板JSON格式 |
| "高级模版名称已存在" | 模板名称重复 | 使用其他名称 |
| "未开启高级模版打印" | 功能未开启 | 在设置中开启功能 |
| "默认模版不能删除" | 尝试删除默认模板 | 只能删除高级模板 |

### 错误处理机制

1. **参数验证**: 所有接口都进行严格的参数验证
2. **事务处理**: 关键操作使用数据库事务保证一致性
3. **错误包装**: 使用 `errors.WithMessage` 包装错误信息
4. **日志记录**: 关键错误记录到日志系统

---

## 📊 数据模型

### Printer - 打印机

```go
type Printer struct {
    BaseModel
    Name              string // 打印机名称
    PrinterTypeUuid   uint64 // 打印机类型UUID
    ConfigJson        string // 打印机配置JSON
    Sn                string // 打印机SN
    Copies            uint   // 打印份数
    Sort              uint   // 排序
    IsUsb             int    // 是否USB打印机
    Status            int    // 状态（1=在线, 0=离线）
    LastHeartbeatTime uint   // 最后心跳时间
    SourceDeviceSn    string // 源设备SN
    PrintMethod       int    // 打印方式（1=文本, 2=图片）
    Width             int    // 纸张宽度（mm）
    EnableStatusCheck int    // 是否启用状态检查
    EnableSound       int    // 是否启用打印提示音
    PrintSpeed        int    // 打印速度模式
}
```

### PrinterLog - 打印日志

```go
type PrinterLog struct {
    BaseModel
    PrinterUuid        uint64 // 打印机UUID
    CashierDeviceId    string // 收银机设备ID
    RelatedType        int    // 关联类型（0=销售账单, 1=销售订单, 2=充值订单, 3=交班单）
    RelatedUuid        uint64 // 关联UUID
    Data               string // 打印数据
    Type               int    // 类型（0=系统队列, 1=云服务）
    DataType           int    // 数据类型（模板类型）
    PrintMethod        int    // 打印方式
    Num                int    // 打印次数
    Status             int    // 状态（0=失败, 1=进行中, 2=成功）
    Reason             string // 失败原因
    PrinterTime        int64  // 打印时间
    FirstExecution     int    // 是否首次执行
    ReadDeviceId       string // 读取设备ID
    ProductPrinterUuid uint64 // 商品打印机UUID
    PrinterType        string // 打印机类型
    PrintingTime       int64  // 打印耗时
    Copies             uint   // 打印份数
    PrintSpeed         int    // 打印速度模式
}
```

### PrinterTemplate - 打印模板

```go
type PrinterTemplate struct {
    BaseModel
    Name     string // 模板名称
    Template int    // 模板类型
    TmpUuid  uint64 // 定制UUID
    TmpData  string // 定制数据
}
```

### PrinterCustomize - 打印机定制

```go
type PrinterCustomize struct {
    BaseModel
    Name       string // 定制名称
    Data       string // 定制数据（JSON）
    TemplateId uint64 // 模板ID
    IsAdv      int    // 是否高级模板（1=是, 0=否）
    IsUse      int    // 是否使用中（1=是, 0=否）
}
```

### ProductPrinter - 商品打印机（档口）

```go
type ProductPrinter struct {
    BaseModel
    Name               string // 名称（档口名称）
    Status             int    // 状态（1=开启, 0=关闭）
    PrintMode          int    // 打印模式（0=付款打印, 1=送厨打印）
    PrintMethod        int    // 打印方式（0=整单, 1=一菜一单）
    PrintProductSelect int    // 打印商品选择（0=按分类, 1=按标签）
    PrintModeScene     int    // 打印模式场景（0=合并, 1=分开）
    Copies             uint   // 打印份数
}
```

---

## 🚀 性能优化

### 缓存策略

1. **商品打印机列表缓存**:
   - 缓存键: `product_printer_list_v2:{company_uuid}:{width_print_mode}`
   - 缓存时间: 24小时
   - 数据压缩: GZIP压缩 + Base64编码

2. **打印设置缓存**:
   - 缓存键: `setting:company_id:{company_uuid}`
   - 更新时自动清除缓存

### 数据压缩

- **打印数据**: 使用GZIP压缩，减少存储空间和传输时间
- **缓存数据**: 使用GZIP压缩，提高缓存效率

### 并发控制

- **获取打印数据**: 使用UUID锁防止并发操作
- **打印任务**: 使用队列异步处理，避免阻塞

---

## 🧪 测试建议

### 单元测试

1. **模板解析测试**:
   - 测试各种模板类型的解析
   - 测试模板数据验证
   - 测试错误处理

2. **打印数据生成测试**:
   - 测试文本打印数据生成
   - 测试图片打印数据生成
   - 测试数据压缩和解压

3. **打印机管理测试**:
   - 测试USB打印机上报
   - 测试打印机状态更新
   - 测试打印机类型识别

### 集成测试

1. **打印流程测试**:
   - 测试完整打印流程
   - 测试打印队列处理
   - 测试打印报告上报

2. **模板定制测试**:
   - 测试模板创建、编辑、删除
   - 测试模板预览
   - 测试模板使用

3. **多打印机测试**:
   - 测试多种打印机类型
   - 测试打印速度模式
   - 测试打印分片处理

### 性能测试

1. **打印数据生成性能**:
   - 测试大量订单打印数据生成
   - 测试数据压缩性能

2. **打印队列性能**:
   - 测试高并发打印任务处理
   - 测试打印队列容量

3. **缓存性能**:
   - 测试缓存命中率
   - 测试缓存更新性能

---

## 📝 注意事项

1. **打印数据存储**:
   - 打印数据使用压缩存储，减少数据库空间
   - 数据格式: `GZIP:{base64_encoded_data}`

2. **打印速度模式**:
   - 流畅模式适合稳定网络，但可能因数据量大而失败
   - 稳定模式是默认模式，适合大多数场景
   - 兼容模式适合不稳定网络，但速度较慢

3. **USB打印机**:
   - USB打印机需要客户端定期上报
   - 长时间未上报的打印机会被标记为离线
   - 新发现的USB打印机会自动创建记录

4. **模板定制**:
   - 默认模板不能删除
   - 正在使用的模板不能删除
   - 高级模板需要开启相应功能

5. **打印队列**:
   - 打印任务异步执行，不阻塞业务
   - 打印结果通过报告接口上报
   - 支持补打功能

---

## 🔗 相关文档

- [设置服务文档](./setting.md) - 打印设置相关配置
- [订单服务文档](./order.md) - 订单打印相关功能
- [会员服务文档](./member.md) - 充值单打印相关功能

---

**文档版本**: v1.0  
**最后更新**: 2025-01-27  
**维护者**: TTPOS开发团队

