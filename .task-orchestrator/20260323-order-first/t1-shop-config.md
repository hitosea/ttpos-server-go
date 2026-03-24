# Shop 端手机点餐设置和会员端堂食配置代码分析

## 1. Shop 端门店点餐（手机点餐）配置

### 1.1 API 路由
- **获取配置**: `GET /shop/setting/store_scan_order` (第1228行)
- **保存配置**: `POST /shop/setting/store_scan_order` (第1248行)

### 1.2 文件位置

#### API 层
- **文件**: `/home/ttpos_602666178/ttpos-server-go/.dev/worktree/quiet-garden/main/app/api/v1/shop/shop_setting.go`
- **获取接口**: 第1219-1236行 `GetStoreScanOrderSetting()`
- **保存接口**: 第1238-1262行 `SaveStoreScanOrderSetting()`
- **路由注册**: 第1595-1596行

#### Service 层
- **文件**: `/home/ttpos_602666178/ttpos-server-go/.dev/worktree/quiet-garden/main/app/service/setting/setting.go`
- **获取方法**: 第2606-2636行 `GetStoreScanOrderSetting()`
- **保存方法**: 第2638-2667行 `SaveStoreScanOrderSetting()`
- **接口定义**: 第95-96行 ISrv 接口中定义

#### DTO 层
- **请求对象**: `/home/ttpos_602666178/ttpos-server-go/.dev/worktree/quiet-garden/main/app/dto/req/store_scan_order_setting.go`
  ```go
  type SaveStoreScanOrderSettingReq struct {
    IsEnabled        int `json:"is_enabled"`         // 启用状态：0-关闭，1-开启
    EnableDelivery   int `json:"enable_delivery"`    // 外送服务：0-关闭，1-开启
    EnableSelfPickup int `json:"enable_self_pickup"` // 到店自取：0-关闭，1-开启
  }
  ```

- **响应对象**: `/home/ttpos_602666178/ttpos-server-go/.dev/worktree/quiet-garden/main/app/dto/resp/setting/store_scan_order_setting.go`
  ```go
  type StoreScanOrderSettingResp struct {
    IsEnabled           int `json:"is_enabled"`             // 启用状态
    EnableDelivery      int `json:"enable_delivery"`        // 外送服务
    EnableSelfPickup    int `json:"enable_self_pickup"`     // 到店自取
    DeliveryAvailable   int `json:"delivery_available"`     // 外送是否可用
    SelfPickupAvailable int `json:"self_pickup_available"`  // 自取是否可用
  }
  ```

### 1.3 数据库存储

#### 表名
- **主表**: `ttpos_setting` 
- **存储Key**: `store_scan_order` (常量在constant/setting.go第27行)

#### 字段结构
- `key`: varchar(30) = "store_scan_order"
- `values`: mediumtext (JSON格式存储配置)
- `describe`: varchar(255) = "门店点餐设置"
- `create_time`: int(10)
- `update_time`: int(10)
- `delete_time`: int(10)

### 1.4 配置获取逻辑
1. 从 `ttpos_setting` 表读取 key="store_scan_order" 的记录
2. 解析 JSON 格式的 values 字段
3. 默认值：IsEnabled=1, EnableDelivery=1, EnableSelfPickup=1
4. 实时计算可用性：
   - `DeliveryAvailable`: 取决于 company_setting.is_open_rider（外送状态）
   - `SelfPickupAvailable`: 取决于 company_setting.is_open_member_instant（会员端即时点餐功能）

### 1.5 配置保存逻辑
1. 验证外送服务：若启用则需要 company_setting.delivery_status=1
2. 验证到店自取：若启用则需要 company_setting.is_open_member_instant=1
3. 通过 `UpdateSetting()` 方法保存到数据库

---

## 2. 会员端堂食配置

### 2.1 堂食是否开启的控制字段

#### CompanySetting 中的关键字段
- **字段名**: `is_open_member_instant`
- **表名**: `ttpos_company_setting` (SaaS 主库表)
- **类型**: int(10), default=0
- **含义**: 是否开启会员端即时点餐功能（扫码点餐到店自取）0不开启, 1开启
- **文件**: `/home/ttpos_602666178/ttpos-server-go/.dev/worktree/quiet-garden/main/app/model/company.go` 第115行
- **JSON标签**: `json:"is_open_member_instant"`

### 2.2 会员端堂食 API 路由

#### 订单操作
- **创建堂食订单**: `POST /member/order/dine_in/create` (第76行)
- **获取堂食订单表单**: `GET /member/order/dine_in/form_info` (第144行)
- **设置用餐方式**: `POST /member/order/dine_in/dining_method` (第175行)
- **提交支付**: `POST /member/order/dine_in/pay` (第210行)
- **获取支付信息**: `GET /member/order/dine_in/pay/info` (第245行)
- **获取支付状态**: `GET /member/order/dine_in/pay/status` (第274行)

#### 查询操作
- **堂食订单检查**: `GET /member/order/dine_in/check` (第611行)
- **堂食订单列表**: `GET /member/order/dine_in/list` (第647行)
- **堂食订单详情**: `GET /member/order/dine_in/detail` (第676行)
- **取消订单**: `POST /member/order/dine_in/cancel` (第734行)

### 2.3 文件位置

#### API 层
- **文件**: `/home/ttpos_602666178/ttpos-server-go/.dev/worktree/quiet-garden/main/app/api/v1/member/member_order.go`
- **主要处理器**: OrderHandler (第20-22行)
- **堂食创建接口**: 第66-101行 `CreateDineInOrder()`
- **路由注册**: 第784-797行

#### Service 层
- **文件**: `/home/ttpos_602666178/ttpos-server-go/.dev/worktree/quiet-garden/main/app/service/order_member.go`
- **关键方法**: 
  - `CreateMemberDineInOrder()` 
  - 堂食支付回调处理 (第648行)
  - 订单列表查询 (第3195行)

### 2.4 堂食订单来源标记
- **常量**: `OrderSourceMemberDineIn = "member_dine_in"`
- **文件**: `/home/ttpos_602666178/ttpos-server-go/.dev/worktree/quiet-garden/main/app/constant/order.go` 第12行

### 2.5 堂食相关操作日志
- `ActionMemberDineInSetDiningMethod = "member_dine_in_set_dining_method"` // 设置用餐方式
- `ActionMemberDineInPay = "member_dine_in_pay"` // 订单支付
- `ActionMemberDineInCancel = "member_dine_in_cancel"` // 取消订单
- **文件**: `/home/ttpos_602666178/ttpos-server-go/.dev/worktree/quiet-garden/main/app/constant/operation_duration.go` 第126-128行

### 2.6 堂食可用性检查
会员端创建堂食订单前需要检查：
```go
// 来源：service/setting/setting.go 第2625行
if companySetting.IsOpenMemberInstant == 1 {
    selfPickupAvailable = 1  // 堂食可用
}

// 来源：service/setting/setting.go 第2651行
if settingReq.EnableSelfPickup == 1 && companySetting.IsOpenMemberInstant != 1 {
    return errors.New("暂未开启，请联系销售代表")
}
```

---

## 3. 配置调用链

### 3.1 Shop 端门店点餐配置调用链

```
HTTP GET /shop/setting/store_scan_order
    ↓
shop/shop_setting.go: SettingHandler.GetStoreScanOrderSetting()
    ↓
service/setting/setting.go: Srv.GetStoreScanOrderSetting()
    ↓
从 ttpos_setting 表读取 key="store_scan_order"
    ↓
解析 JSON 并组装响应
    ├─ CompanySetting.IsOpenRider() → DeliveryAvailable
    └─ CompanySetting.IsOpenMemberInstant → SelfPickupAvailable
    ↓
返回 StoreScanOrderSettingResp
```

### 3.2 Shop 端保存门店点餐配置调用链

```
HTTP POST /shop/setting/store_scan_order
    ↓
shop/shop_setting.go: SettingHandler.SaveStoreScanOrderSetting()
    ↓
service/setting/setting.go: Srv.SaveStoreScanOrderSetting()
    ├─ 验证 EnableDelivery：检查 CompanySetting.DeliveryStatus==1
    └─ 验证 EnableSelfPickup：检查 CompanySetting.IsOpenMemberInstant==1
    ↓
UpdateSetting() → 更新 ttpos_setting 表
    ↓
删除缓存，推送配置更新
```

### 3.3 会员端堂食配置流程

```
会员端扫码 / 即时点餐
    ↓
检查 CompanySetting.IsOpenMemberInstant
    ├─ = 0: 堂食功能不可用
    └─ = 1: 允许创建堂食订单
    ↓
POST /member/order/dine_in/create
    ↓
service/order_member.go: CreateMemberDineInOrder()
    ├─ 读取 StoreScanOrderSetting 检查 EnableSelfPickup
    └─ 创建订单，OrderSource = "member_dine_in"
    ↓
返回订单详情
```

---

## 4. 关键常量

### Setting Key 常量
- **文件**: `constant/setting.go`
- `SettingStore = "store"`
- `SettingBusiness = "business"`
- `SettingStoreScanOrder = "store_scan_order"`

### 订单来源
- `OrderSourceMemberDineIn = "member_dine_in"` // 会员堂食

### 商户功能开关（Company Setting）
- `IsOpenMemberInstant` (int): 会员端即时点餐（堂食）
- `DeliveryStatus` (int): 外送状态
- `IsOpenRider` (bool): 骑手配送开启状态

---

## 5. 缓存策略

配置保存时会清除以下缓存：
1. 系统设置缓存：`setting:company_id:{companyUuid}`
2. 通用设置缓存标签：`common_get_settingLanguages`
3. 收银端缓存标签：`cashier`
4. 对象存储缓存（如启用）：`BusinessSettingCache`

配置变更时会通过 WebSocket 推送更新：
- 事件类型：`UPDATE_CONFIG`
- 推送数据包含：`update_time` (Unix时间戳)

---

## 6. 总结

- **Shop 端手机点餐**: 通过 `store_scan_order` 设置项管理，存储于 `ttpos_setting` 表
- **会员端堂食**: 通过 `company_setting.is_open_member_instant` 字段控制，存储于 `ttpos_company_setting` 表
- **核心API**: `/shop/setting/store_scan_order` (查询/保存) 和 `/member/order/dine_in/*` (堂食订单操作)
- **依赖关系**: 会员堂食功能依赖于云平台的 `is_open_member_instant` 开启状态

