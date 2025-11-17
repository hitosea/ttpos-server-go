# Production Service 厨房生产服务说明文档

## 📋 概述

`service/production.go` 是 TTPOS 系统的厨房生产管理服务，负责厨显端（Kitchen Display System, KDS）的核心业务逻辑。该服务管理送厨商品的展示、制作、传菜、恢复和退菜确认等功能，支持智能厨显模式（制作/传菜分离）、分批送厨、套餐商品同步等高级特性。

**文件路径**: `/home/coder/workspaces/ttpos-server-go/main/app/service/production.go`  
**文件大小**: 1097行  
**接口定义**: `IProductionSrv`  
**实现结构**: `productionSrv`

---

## 🏗️ 架构设计

### 接口定义 (IProductionSrv)

```go
type IProductionSrv interface {
    // 根据订单获取送厨商品
    GetProductListByOrder(ctx context.Context, req req.ProductionListReq) (resp.ProductionListWithPagination, error)
    
    // 根据分类获取送厨商品
    GetProductListByCategory(ctx context.Context, req req.ProductionListByCategoryReq) (resp.ProductionListWithPagination, error)
    
    // 获取制作完成、传菜完成历史
    GetHistory(ctx context.Context, req req.HistoryReq) (resp.ProductionHistory, error)
    
    // 完成制作、传菜
    Finish(ctx context.Context, req req.FinishReq) error
    
    // 恢复制作
    Recovery(ctx context.Context, req req.RecoveryReq) error
    
    // 厨显端确认退菜
    ConfirmReturn(ctx context.Context, productUuid uint64) error
    
    // 厨显端确认退菜整单
    ConfirmReturnAll(ctx context.Context, saleBillUuid uint64) error
}
```

### 依赖服务

```go
type productionSrv struct {
    dbm        *database.DBManager  // 数据库管理器
    settingSrv setting.ISrv         // 设置服务
}
```

### 服务初始化

```go
func NewProductionSrv(dbm *database.DBManager, settingSrv setting.ISrv) IProductionSrv {
    return NewProductionSrvImpl(dbm, settingSrv)
}

func NewProductionSrvImpl(dbm *database.DBManager, settingSrv setting.ISrv) IProductionSrv {
    return &productionSrv{
        dbm:        dbm,
        settingSrv: settingSrv,
    }
}
```

---

## 🎯 核心概念

### 1. 智能厨显模式

智能厨显支持三种工作模式，实现制作和传菜的分离管理：

| 模式 | 常量 | 说明 |
|-----|------|------|
| 单传菜模式 | `KdsModeDefault` | 仅显示待传菜的商品（已制作完成） |
| 仅制作模式 | `KdsModeMake` | 仅显示待制作的商品 |
| 制作+传菜模式 | `KdsModeMakeAndSend` | 同时支持制作和传菜 |

#### 模式判断逻辑

```go
func (s *productionSrv) getMode(ctx context.Context, reqMode uint) (*uint, error) {
    var mode *uint = nil
    
    // 获取厨显设置
    kitchenSetting, _ := s.settingSrv.GetKitchenSetting(ctx, ctx.GetCompanySetting(), nil)
    
    // 如果开启智能后厨
    if kitchenSetting.IsSmartKitchen == "1" {
        // 获取设备信息
        device, _ := deviceRepo.GetDevice(...)
        
        switch device.KdsMode {
        case constant.KdsModeDefault: // 单传菜模式
            if reqMode != constant.ReqModeSend {
                return nil, errors.New("本机不支持查看待制作的菜品")
            }
        case constant.KdsModeMake: // 仅制作模式
            if reqMode != constant.ReqModeMake {
                return nil, errors.New("本机不支持查看待传菜的菜品")
            }
        case constant.KdsModeMakeAndSend: // 制作+传菜模式
            if reqMode != constant.ReqModeMake && reqMode != constant.ReqModeSend {
                return nil, errors.New("本机不支持查看当前状态的菜品")
            }
        }
        mode = &reqMode
    }
    
    return mode, nil
}
```

### 2. 商品状态

#### 制作状态 (MakeStatus)

| 状态 | 常量 | 说明 |
|-----|------|------|
| 待制作 | `ProductionOrderProductMakeStatusDefault` | 初始状态 |
| 已制作 | `ProductionOrderProductMakeStatusFinished` | 制作完成 |
| 已恢复 | `ProductionOrderProductMakeStatusRecovery` | 从已制作恢复到待制作 |

#### 传菜状态 (Status)

| 状态 | 常量 | 说明 |
|-----|------|------|
| 制作中 | `ProductionOrderProductStatusCooking` | 正在制作或待传菜 |
| 已完成 | `ProductionOrderProductStatusFinished` | 传菜完成 |

### 3. 分批送厨

分批送厨功能允许某些商品分批次送到厨房，而不是一次性全部送厨。

#### 特点

- 分批商品有 `is_batch` 标识
- 每批有独立的送厨时间 `batch_time`
- 预送厨阶段（`batch_time = 0`）不显示
- 送厨时间以 `batch_time` 为准

#### 判断逻辑

```go
// 非分批商品、或者分批已送厨商品
productionRepo.WhereIsNotBatchOrBatchTimeGT0()

// 过滤预送厨阶段的分批商品
if businessSetting.OpenIsBatch() {
    if product.IsBatchBool() && product.IsPreCooking() {
        continue  // 不显示
    }
}
```

### 4. 套餐商品同步

套餐商品的状态需要与子商品同步：

- **子商品全部完成** → 套餐商品标记为完成
- **任一子商品未完成** → 套餐商品标记为制作中
- **时长计算**: 套餐时长 = 当前时间 - 送厨时间

---

## 🎯 核心功能

### 1. 根据订单获取送厨商品 (GetProductListByOrder)

**功能描述**: 按订单分组显示待制作/待传菜的商品列表，用于厨显端主界面。

#### 方法签名

```go
func (s *productionSrv) GetProductListByOrder(
    ctx context.Context, 
    req req.ProductionListReq
) (resp.ProductionListWithPagination, error)
```

#### 请求参数

```go
type ProductionListReq struct {
    Mode     uint `json:"mode"`      // 模式 0-传菜 1-制作
    PageNo   int  `json:"page_no"`   // 页码
    PageSize int  `json:"page_size"` // 每页大小
}
```

#### 返回数据结构

```go
type ProductionListWithPagination struct {
    SendKitchenNum int                `json:"send_kitchen_num"` // 送厨总数
    List           []ProductionGroup  `json:"list"`             // 商品分组列表
    FinishedList   ProductionList     `json:"finished_list"`    // 最近完成列表
    Meta           dto.PageResponse   `json:"meta"`             // 分页信息
}

type ProductionGroup struct {
    LocaleName        *dto.LocaleResponse `json:"locale_name"`         // 订单号（多语言）
    DiningMethod      int                 `json:"dining_method"`       // 用餐方式
    SaleBillUuid      uint64              `json:"sale_bill_uuid"`      // 销售账单UUID
    IsSaleBillDeleted bool                `json:"is_sale_bill_deleted"`// 是否已整单取消
    IsTakeoutBill     bool                `json:"is_takeout_bill"`     // 是否外送订单
    OrderRemark       OrderRemarkRes      `json:"order_remark"`        // 订单备注
    ProductionList    ProductionList      `json:"production_list"`     // 商品列表
}

type ProductionItem struct {
    Uuid                  uint64              `json:"uuid"`                    // 商品UUID
    LocaleName            dto.LocaleResponse  `json:"locale_name"`             // 商品名称（多语言）
    ProductAttributeNames string              `json:"product_attribute_names"` // 商品属性
    Num                   float64             `json:"num"`                     // 数量
    NumType               string              `json:"num_type"`                // 数量单位
    Remark                string              `json:"remark"`                  // 备注
    SerialNo              string              `json:"serial_no"`               // 订单号
    DiningMethod          int                 `json:"dining_method"`           // 用餐方式
    CreateTime            int64               `json:"create_time"`             // 送厨时间
    FinishedTime          int64               `json:"finished_time"`           // 完成时间
    MakeDuration          int64               `json:"make_duration"`           // 制作时长
    SendDuration          int64               `json:"send_duration"`           // 传菜时长
    IsSaleBillDeleted     bool                `json:"is_sale_bill_deleted"`    // 是否已整单取消
    BatchTag              BatchTagInfo        `json:"batch_tag"`               // 分批标签
    OrderRemark           OrderRemarkRes      `json:"order_remark"`            // 订单备注
}
```

#### 执行流程

```
1. 验证工作模式
   ↓
2. 获取厨显绑定的商品打印机
   ↓
3. 获取打印机关联的商品UUID列表
   ↓
4. 获取打印机关联的销售账单UUID列表
   ↓
5. 构建查询条件
   ├─ 状态 = 制作中
   ├─ 商品UUID在列表中
   ├─ 账单UUID在列表中
   ├─ 制作模式：仅未制作完成的
   ├─ 版本<2.4.0：数量>0
   └─ 非分批或已分批送厨
   ↓
6. 分页查询订单（按sale_bill_uuid分组）
   ↓
7. 查询该页订单的所有商品
   ↓
8. 按订单分组组装数据
   ├─ 处理分批商品
   ├─ 处理套餐备注
   ├─ 处理多语言
   ├─ 处理打包状态
   └─ 处理分批标签
   ↓
9. 获取最近完成列表（3条）
   ↓
10. 返回分页结果
```

#### 代码示例

```go
// 获取待传菜商品列表
req := req.ProductionListReq{
    Mode:     constant.ReqModeSend,  // 传菜模式
    PageNo:   1,
    PageSize: 10,
}

result, err := productionSrv.GetProductListByOrder(ctx, req)
if err != nil {
    // 错误处理
}

// 遍历订单分组
for _, group := range result.List {
    fmt.Printf("订单号: %s\n", group.LocaleName.ZH)
    fmt.Printf("商品数量: %d\n", len(group.ProductionList.List))
    
    for _, item := range group.ProductionList.List {
        fmt.Printf("  - %s x %.2f\n", item.LocaleName.ZH, item.Num)
    }
}
```

---

### 2. 根据分类获取送厨商品 (GetProductListByCategory)

**功能描述**: 按商品分类分组显示待制作/待传菜的商品列表。

#### 方法签名

```go
func (s *productionSrv) GetProductListByCategory(
    ctx context.Context, 
    req req.ProductionListByCategoryReq
) (resp.ProductionListWithPagination, error)
```

#### 请求参数

```go
type ProductionListByCategoryReq struct {
    Mode         uint   `json:"mode"`          // 模式 0-传菜 1-制作
    CategoryUuid uint64 `json:"category_uuid"` // 分类UUID（0表示全部）
    PageNo       int    `json:"page_no"`       // 页码
    PageSize     int    `json:"page_size"`     // 每页大小
}
```

#### 执行流程

```
1. 验证工作模式
   ↓
2. 获取商品和账单列表
   ↓
3. 构建查询条件
   ├─ 基础条件同GetProductListByOrder
   └─ 指定分类UUID（如果有）
   ↓
4. 分页查询分类（按first_category_uuid分组）
   ↓
5. 查询该页分类的所有商品
   ↓
6. 预加载分类多语言名称
   ↓
7. 按分类分组组装数据
   ├─ 处理分批商品
   ├─ 处理套餐备注
   ├─ 处理多语言
   ├─ 处理打包状态
   └─ 处理分批标签
   ↓
8. 获取最近完成列表
   ↓
9. 返回分页结果
```

#### 使用场景

- 按菜品类型分区的厨房（如炒菜区、凉菜区）
- 根据分类安排不同工位
- 分类统计和管理

---

### 3. 获取历史记录 (GetHistory)

**功能描述**: 获取过去24小时内的制作完成或传菜完成历史记录。

#### 方法签名

```go
func (s *productionSrv) GetHistory(
    ctx context.Context, 
    req req.HistoryReq
) (resp.ProductionHistory, error)
```

#### 请求参数

```go
type HistoryReq struct {
    Mode uint `json:"mode"` // 模式 0-传菜历史 1-制作历史
}
```

#### 返回数据结构

```go
type ProductionHistory struct {
    List []ProductionGroup `json:"list"` // 历史记录列表
}
```

#### 执行逻辑

```go
// 获取过去24小时内的记录
var statusOpt, finishedTimeOpt repository.DBOption

if mode != nil && *mode == constant.KdsModeMake {
    // 制作历史
    statusOpt = productionRepo.WhereProductMakeStatus([]uint{
        constant.ProductionOrderProductMakeStatusFinished,
    })
    finishedTimeOpt = productionRepo.WhereProductMadeTime(
        time.Now().Add(-24 * time.Hour).Unix(),
    )
} else {
    // 传菜历史
    statusOpt = productionRepo.WhereProductStatus(
        constant.ProductionOrderProductStatusFinished,
    )
    finishedTimeOpt = productionRepo.WhereProductFinishedTime(
        time.Now().Add(-24 * time.Hour).Unix(),
    )
}

// 查询并按订单分组
limitProducts, _ := productionRepo.GetLimitedHistoryProducts(...)
_, products, _ := productionRepo.GetProducts(...)

return resp.ProductionHistory{
    List: s.groupByOrder(ctx, limitProducts, products, mode),
}
```

---

### 4. 完成制作/传菜 (Finish)

**功能描述**: 标记商品为制作完成或传菜完成，是厨显端最核心的操作。

#### 方法签名

```go
func (s *productionSrv) Finish(
    ctx context.Context, 
    req req.FinishReq
) error
```

#### 请求参数

```go
type FinishReq struct {
    Mode                      uint     `json:"mode"`                        // 模式
    ProductUuid               uint64   `json:"product_uuid"`                // 单个商品UUID
    ProductUuids              []uint64 `json:"product_uuids"`               // 批量商品UUID列表
    ConfirmReturnProductUuids []uint64 `json:"confirm_return_product_uuids"` // 同时确认退菜的商品UUID列表
}
```

#### 执行流程

```
1. 参数验证
   ├─ 至少有一个商品UUID
   └─ 商品必须存在
   ↓
2. 验证商品状态
   ├─ 状态必须是"制作中"
   ├─ 数量必须>0
   └─ 退菜商品数量必须=0
   ↓
3. 确认退菜商品（如果有）
   ↓
4. 判断工作模式
   │
   ├─ 制作模式 (Mode=1)
   │  ├─ 验证未制作完成
   │  ├─ 更新make_status = 已制作
   │  ├─ 记录made_time
   │  └─ 计算make_duration
   │
   └─ 传菜模式 (Mode=0)
      ├─ 验证已制作完成
      ├─ 更新status = 已完成
      ├─ 记录finished_time
      └─ 计算send_duration
   ↓
5. 更新套餐商品状态
   ↓
6. 提交事务
   ↓
7. 发布完成制作事件
   ↓
8. WebSocket推送更新
   ├─ 推送厨显更新
   └─ 推送订单更新
```

#### 时长计算逻辑

##### 智能厨显模式（mode != nil）

**制作模式完成**:
```sql
make_duration = CASE 
    WHEN is_batch = 1 THEN now - batch_time 
    ELSE now - create_time 
END
```

**传菜模式完成**:
```sql
send_duration = CASE 
    WHEN made_time = 0 THEN 
        CASE 
            WHEN is_batch = 1 THEN now - batch_time 
            ELSE now - create_time 
        END
    ELSE now - made_time 
END

all_duration = make_duration + send_duration
```

##### 非智能厨显模式（mode = nil）

```sql
make_duration = CASE 
    WHEN is_batch = 1 THEN now - batch_time 
    ELSE now - create_time 
END

all_duration = make_duration
```

#### 套餐商品同步

```go
func (s *productionSrv) updatePackageProduct(tx *gorm.DB, saleOrderProductUuids []uint64, finishedTime int64) error {
    // 1. 获取套餐商品
    saleOrderProducts, _ := saleOrderProductRepo.GetSaleOrderProductsByUuids(...)
    
    // 2. 提取套餐UUID
    var packageUuids []uint64
    for _, product := range saleOrderProducts {
        if product.IsPackageSubProduct() {
            packageUuids = append(packageUuids, product.PackageUuid)
        }
    }
    
    // 3. 遍历每个套餐
    for _, packageUuid := range packageUuids {
        // 获取套餐的所有子商品
        productionProducts, _ := productionRepo.GetProductsByPackageUuid(packageUuid)
        
        // 4. 检查是否有子商品未完成
        for _, product := range productionProducts {
            if product.Status == constant.ProductionOrderProductStatusCooking {
                // 套餐状态设为制作中
                productionRepo.UpdateProduct(...)
                return nil
            }
        }
        
        // 5. 所有子商品都已完成，套餐标记为完成
        makeDuration := finishedTime - createTime
        productionRepo.UpdateProduct(packageUuid, map[string]any{
            "status":        constant.ProductionOrderProductStatusFinished,
            "finished_time": finishedTime,
            "make_duration": makeDuration,
            "all_duration":  makeDuration,
        })
    }
    
    return nil
}
```

#### 事件发布

```go
// 发布完成制作事件
event.NewSystemBus().PublishFinishMenuEvent(event.FinishMenuPayload{
    BasePayload: event.BasePayload{
        Ctx:           ctx,
        CompanyUuid:   ctx.GetCompanyUuid(),
        SaleBillUuid:  saleBillUuid,
    },
    FinishedTime: finishedTime,
    Products: orderProducts,
})
```

---

### 5. 恢复制作 (Recovery)

**功能描述**: 将已完成的商品恢复到制作中状态，用于误操作纠正。

#### 方法签名

```go
func (s *productionSrv) Recovery(
    ctx context.Context, 
    req req.RecoveryReq
) error
```

#### 请求参数

```go
type RecoveryReq struct {
    Mode        uint   `json:"mode"`         // 模式
    ProductUuid uint64 `json:"product_uuid"` // 商品UUID
}
```

#### 执行流程

```
1. 查询商品
   ↓
2. 判断工作模式
   │
   ├─ 制作模式 (Mode=1)
   │  ├─ 验证已制作完成
   │  ├─ 验证未传菜
   │  ├─ 恢复make_status = 待制作
   │  └─ 清空made_time和时长
   │
   └─ 传菜模式 (Mode=0)
      ├─ 验证已传菜完成
      ├─ 恢复status = 制作中
      └─ 清空finished_time和send_duration
   ↓
3. 更新套餐商品状态
   ↓
4. WebSocket推送更新
   ├─ 推送厨显更新
   └─ 推送订单更新
```

#### 限制条件

| 模式 | 可恢复条件 | 不可恢复条件 |
|-----|-----------|-------------|
| 制作模式 | 已制作完成 且 未传菜 | 已传菜的商品 |
| 传菜模式 | 已传菜完成 | - |

#### 代码示例

```go
// 恢复误操作完成的商品
req := req.RecoveryReq{
    Mode:        constant.KdsModeMake,
    ProductUuid: 12345,
}

err := productionSrv.Recovery(ctx, req)
if err != nil {
    // 错误处理
    if err.Error() == "该菜品已传菜，不可恢复！" {
        // 提示用户无法恢复
    }
}
```

---

### 6. 确认退菜 (ConfirmReturn)

**功能描述**: 厨显端确认退菜商品，将商品从展示列表中移除。

#### 方法签名

```go
func (s *productionSrv) ConfirmReturn(
    ctx context.Context, 
    productUuid uint64
) error
```

#### 执行流程

```
1. 查询商品
   ↓
2. 验证商品数量=0（已退菜）
   ↓
3. 更新delete_time
   ↓
4. WebSocket推送更新厨显
```

#### 业务场景

- 收银端退菜后，商品数量变为0
- 厨显端显示退菜提示
- 厨房确认后，商品从列表消失

---

### 7. 确认退菜整单 (ConfirmReturnAll)

**功能描述**: 厨显端确认整单退菜，批量处理订单的所有商品。

#### 方法签名

```go
func (s *productionSrv) ConfirmReturnAll(
    ctx context.Context, 
    saleBillUuid uint64
) error
```

#### 执行流程

```
1. 查询销售账单
   ↓
2. 验证订单状态
   ├─ 订单已删除 或 状态=已取消
   └─ 厨显未确认（is_kitchen_confirm=0）
   ↓
3. 开启事务
   ├─ 更新订单is_kitchen_confirm=1
   └─ 更新所有商品delete_time
   ↓
4. 提交事务
   ↓
5. WebSocket推送更新厨显
```

#### 业务场景

- 顾客整单取消订单
- 厨显端显示整单退菜提示
- 厨房确认后，该订单所有商品消失

---

## 🔄 数据流转

### 1. 送厨到完成流程

```
收银端下单
  ↓
生成销售订单
  ↓
送厨（创建生产单）
  ├─ status = 制作中
  ├─ make_status = 待制作
  ├─ create_time = 当前时间
  └─ 分批商品：is_batch=1, batch_time=0（预送厨）
  ↓
厨显端查询商品列表
  ├─ 按订单分组 或 按分类分组
  └─ 过滤预送厨的分批商品
  ↓
【智能厨显-制作模式】
  ├─ 厨师完成制作
  ├─ 调用Finish(mode=1)
  ├─ make_status = 已制作
  ├─ made_time = 当前时间
  └─ make_duration = now - create_time
  ↓
【智能厨显-传菜模式】
  ├─ 传菜员传菜
  ├─ 调用Finish(mode=0)
  ├─ status = 已完成
  ├─ finished_time = 当前时间
  ├─ send_duration = now - made_time
  └─ all_duration = make_duration + send_duration
  ↓
【非智能厨显】
  ├─ 厨师完成制作+传菜
  ├─ 调用Finish(mode=nil)
  ├─ status = 已完成
  ├─ finished_time = 当前时间
  ├─ make_duration = now - create_time
  └─ all_duration = make_duration
  ↓
完成制作事件
  ├─ 发布FinishMenuEvent
  └─ 订单详情显示"已上菜"
  ↓
WebSocket推送
  ├─ 推送厨显更新
  └─ 推送订单更新
```

### 2. 分批送厨流程

```
收银端下单（分批商品）
  ↓
第一次送厨
  ├─ is_batch = 1
  ├─ batch_time = 0（预送厨）
  └─ 厨显端不显示
  ↓
后续分批送厨
  ├─ 更新batch_time = 当前时间
  └─ 厨显端开始显示
  ↓
厨显端显示
  ├─ 送厨时间使用batch_time
  └─ 制作时长 = now - batch_time
```

### 3. 套餐商品流程

```
收银端点套餐
  ↓
生成套餐主商品和子商品
  ├─ 主商品：product_type = package
  └─ 子商品：package_uuid = 主商品uuid
  ↓
送厨
  ├─ 创建主商品生产单
  └─ 创建子商品生产单
  ↓
厨显端完成子商品
  ↓
检查套餐状态
  ├─ 任一子商品未完成 → 套餐=制作中
  └─ 所有子商品已完成 → 套餐=已完成
  ↓
自动同步套餐状态
```

---

## 📊 版本兼容

### 2.4.0 版本差异

| 功能 | < 2.4.0 | >= 2.4.0 |
|-----|---------|----------|
| 商品数量过滤 | 仅显示数量>0的商品 | 显示所有商品（包括退菜） |
| 整单取消确认 | 通过delete_time判断 | 通过is_kitchen_confirm判断 |
| 退菜商品显示 | 自动过滤 | 显示并可确认 |

#### 版本判断

```go
// 2.4.0 之前，只显示大于0的商品
if !ctx.Version(context.GTE, "2.4.0") {
    opts = append(opts, productionRepo.WhereProductNumGT0())
}

// 2.4.0 及以后，获取未确认退菜整单的账单
if ctx.Version(context.GTE, "2.4.0") {
    opt = productPrinterRepo.WhereSaleBillIsKitchenConfirm(0)
} else {
    opt = productPrinterRepo.WhereSaleBillNotDeletedOrIsNotCanceled()
}
```

---

## 🎨 API接口示例

### 1. 获取待制作商品列表

#### 请求

```http
POST /api/v1/production/list_by_order
Authorization: Bearer {token}
Content-Type: application/json

{
  "mode": 1,
  "page_no": 1,
  "page_size": 10
}
```

#### 响应

```json
{
  "code": 1,
  "message": "success",
  "data": {
    "send_kitchen_num": 25,
    "list": [
      {
        "locale_name": {
          "zh": "T001",
          "en": "T001"
        },
        "dining_method": 1,
        "sale_bill_uuid": 12345,
        "is_sale_bill_deleted": false,
        "is_takeout_bill": false,
        "order_remark": {
          "text": "少辣",
          "color": "#FF0000"
        },
        "production_list": {
          "list": [
            {
              "uuid": 101,
              "locale_name": {
                "zh": "宫保鸡丁",
                "en": "Kung Pao Chicken"
              },
              "product_attribute_names": "中辣",
              "num": 1.0,
              "num_type": "份",
              "remark": "不要花生",
              "serial_no": "T001",
              "create_time": 1699000000,
              "batch_tag": {
                "uuid": 1,
                "locale_name": {
                  "zh": "第一批",
                  "en": "Batch 1"
                },
                "color": "#FF5733"
              }
            }
          ]
        }
      }
    ],
    "finished_list": {
      "list": [
        {
          "uuid": 99,
          "locale_name": {
            "zh": "麻婆豆腐",
            "en": "Mapo Tofu"
          },
          "serial_no": "T002",
          "finished_time": 1698999900
        }
      ]
    },
    "meta": {
      "page_no": 1,
      "page_size": 10,
      "total": 25
    }
  }
}
```

### 2. 完成制作

#### 请求

```http
POST /api/v1/production/finish
Authorization: Bearer {token}
Content-Type: application/json

{
  "mode": 1,
  "product_uuids": [101, 102, 103],
  "confirm_return_product_uuids": [104]
}
```

#### 响应

```json
{
  "code": 1,
  "message": "success",
  "data": {}
}
```

### 3. 恢复制作

#### 请求

```http
POST /api/v1/production/recovery
Authorization: Bearer {token}
Content-Type: application/json

{
  "mode": 1,
  "product_uuid": 101
}
```

#### 响应

```json
{
  "code": 1,
  "message": "success",
  "data": {}
}
```

---

## ⚡ 性能优化

### 1. 查询优化

```go
// ✅ 使用分页查询
limitedProducts, total, _ := productionRepo.GetLimitedProducts(
    "sale_bill_uuid",  // 分组字段
    req.PageNo,
    req.PageSize,
    opts...,
)

// ✅ 预加载关联数据
opts = append(opts, 
    productionRepo.WithProductCategory(),
    productionRepo.WithProductCategoryMultiLanguageName(),
)

// ✅ 使用索引字段
productionRepo.WhereProductPackageUuidIn(productPackageUuids)
productionRepo.WhereSaleBillUuidIn(saleBillUuids)
```

### 2. 批量操作

```go
// ✅ 批量完成多个商品
if len(req.ProductUuids) > 0 {
    productionRepo.UpdateProduct(
        []repository.DBOption{
            productionRepo.WhereProductUuidIn(productUuids),
        },
        updateData,
    )
}
```

### 3. 异步处理

```go
// ✅ WebSocket推送异步执行
utils.Go(func() {
    websocket.PushClient(...)
})

// ✅ 事件发布异步执行
utils.Go(func() {
    event.NewSystemBus().PublishFinishMenuEvent(...)
})
```

---

## 🛡️ 最佳实践

### 1. 模式验证

```go
// ✅ 正确：使用getMode验证模式
mode, err := s.getMode(ctx, req.Mode)
if err != nil {
    return err
}

// ❌ 错误：不验证直接使用
if req.Mode == 1 {
    // 可能与设备配置不匹配
}
```

### 2. 套餐商品处理

```go
// ✅ 正确：完成商品后更新套餐状态
if len(saleOrderProductUuids) > 0 {
    if err := s.updatePackageProduct(tx, saleOrderProductUuids, finishedTime); err != nil {
        return err
    }
}

// ❌ 错误：忘记更新套餐状态
// 会导致套餐状态不一致
```

### 3. 事务使用

```go
// ✅ 正确：使用事务确保一致性
err = db.Transaction(func(tx *gorm.DB) error {
    // 更新商品状态
    productionRepo.UpdateProduct(...)
    
    // 更新套餐状态
    s.updatePackageProduct(tx, ...)
    
    // 确认退菜
    confirmReturn(tx)
    
    return nil
})
```

### 4. WebSocket推送

```go
// ✅ 正确：完成操作后推送更新
utils.Go(func() {
    // 推送厨显更新
    websocket.PushClient(companyUuid, websocket.SourceKitchen, ...)
    
    // 推送订单更新
    websocket.PushClient(companyUuid, websocket.SourceAll, ...)
})
```

---

## ⚠️ 注意事项

### 1. 智能厨显模式

- 开启智能厨显后，设备必须配置工作模式
- 制作和传菜分离，需要两次完成操作
- 时长计算逻辑不同

### 2. 分批送厨

- 预送厨阶段（batch_time=0）不显示
- 送厨时间以batch_time为准
- 时长计算使用batch_time

### 3. 套餐商品

- 子商品完成会影响套餐状态
- 套餐时长以最后一个子商品为准
- 套餐备注需要特殊标识

### 4. 版本兼容

- 注意2.4.0版本的差异
- 退菜商品处理逻辑不同
- 使用ctx.Version()判断

### 5. 并发安全

- 使用事务保证数据一致性
- WebSocket推送使用异步
- 事件发布使用协程

---

## 📚 相关文档

- [系统设置](setting.md) - 获取厨显设置
- [Order Service](./order.md) - 订单管理
- [WebSocket Service](../pkg/websocket/websocket_service.md) - 实时推送
- [Event Bus](../pkg/eventbus/eventbus.md) - 事件总线

---

## 📊 服务特点总结

| 特点 | 说明 |
|-----|------|
| 复杂 | 1097行代码，业务逻辑复杂 |
| 灵活 | 支持多种工作模式 |
| 高级 | 智能厨显、分批送厨、套餐同步 |
| 实时 | WebSocket推送更新 |
| 可靠 | 事务保证数据一致性 |
| 兼容 | 支持多版本差异 |

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

**注意**: 本文档基于代码自动生成，如有代码变更，请及时更新文档。生产服务是厨房管理的核心，修改时需充分测试各种模式和场景。

