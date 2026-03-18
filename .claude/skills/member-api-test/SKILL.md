---
name: member-api-test
description: 会员端堂食订单 API 手动测试。当用户要求测试会员端接口、发起堂食订单、查询订单列表/详情、收银端接单/拒单时触发。
allowed-tools: Read, Bash, TodoWrite, AskUserQuestion
---

# 会员端堂食订单 API 测试 Skill

## 环境信息

- **API 基础地址**: `http://localhost:8080/api/v1`
- **公司 UUID**: `7199984230400000`
- **公共请求头**:
  ```
  Device-Id: test-member-device
  TZ: Asia/Bangkok
  X-TTPOS-Company-Id: 7199984230400000
  Content-Type: application/json
  ```

## Token 获取

### 会员端 Token

通过访客登录获取，**无需认证**：

```bash
curl -s -X POST "http://localhost:8080/api/v1/member/visitor/login" \
  -H "Content-Type: application/json" \
  -H "X-TTPOS-Company-Id: 7199984230400000" \
  -d '{"company_uuid": 7199984230400000}'
```

从响应 `data.token` 获取 token，后续请求使用 `Authorization: Bearer <token>`。

### 收银端 Token

需要通过 Go 脚本生成 JWT Token（验证码登录流程复杂，直接生成更高效）：

```bash
cat > /home/ttpos_602666178/ttpos-server-go/main/gen_token.go << 'EOF'
//go:build ignore

package main

import (
    "fmt"
    "time"
    "github.com/golang-jwt/jwt/v5"
)

type Claims struct {
    Source      string `json:"source"`
    CompanyUuid uint64 `json:"company_uuid"`
    StaffUuid   uint64 `json:"staff_uuid"`
    MemberUuid  uint64 `json:"member_uuid"`
    DeviceUuid  uint64 `json:"device_uuid"`
    DeviceId    string `json:"device_id"`
    IsRefreshToken bool `json:"is_refresh_token"`
    Brand       string `json:"brand"`
    jwt.RegisteredClaims
}

func main() {
    secret := "dkjhd00a08"
    claims := Claims{
        Source:      "cashier",
        CompanyUuid: 7199984230400000,
        StaffUuid:   7204597964800000,
        DeviceUuid:  3717765477173250,
        DeviceId:    "2abb4b5626e3beb7fc1225c55c26c780",
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour * 100)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
        },
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenString, _ := token.SignedString([]byte(secret))
    fmt.Println(tokenString)
}
EOF

cd /home/ttpos_602666178/ttpos-server-go/main && go run gen_token.go
rm -f gen_token.go
```

收银端请求头额外需要：
```
Device-Id: 2abb4b5626e3beb7fc1225c55c26c780
```

## 会员端堂食订单完整流程

### Step 1: 获取商品列表

```bash
GET /member/product/list?company_uuid=7199984230400000
```

响应结构中每个商品关键字段：
- `uuid` - 商品 UUID
- `product_type` - 0=普通商品，1=套餐
- `price` - 价格
- `flavors.list[0].uuid` - **flavor_uuid**（创建订单时使用）
- `package_group_list.list` - 套餐分组（套餐商品才有）

### Step 2: 创建堂食订单

```bash
POST /member/order/dine_in/create
```

**普通商品请求体**：
```json
{
  "products": [
    {
      "flavor_uuid": <flavor_uuid>,
      "num": 1,
      "price": <price>,
      "product_type": 0,
      "sauce_uuid": [],
      "attribute_uuid": []
    }
  ]
}
```

**套餐商品请求体**：
```json
{
  "products": [
    {
      "flavor_uuid": <套餐flavor_uuid>,
      "num": 1,
      "price": <套餐price>,
      "product_type": 1,
      "sauce_uuid": [],
      "attribute_uuid": [],
      "products": [
        {
          "product_package_group_uuid": <分组uuid>,
          "flavor_uuid": <子商品flavor_uuid>,
          "num": 1,
          "unit_num": 1,
          "sauce_uuid": [],
          "attribute_uuid": []
        }
      ]
    }
  ]
}
```

响应返回 `sale_bill_uuid` 和 `sale_order_uuid`。

### Step 3: 获取表单信息（可选）

```bash
GET /member/order/dine_in/form_info?sale_bill_uuid=<sale_bill_uuid>
```

### Step 4: 设置用餐方式

```bash
POST /member/order/dine_in/dining_method
Body: {"sale_bill_uuid": <sale_bill_uuid>, "dining_method": 0}
```

`dining_method`: 0=堂食, 1=打包自取

### Step 5: 获取支付方式列表

```bash
GET /member/order/payment/method/list?sale_bill_uuid=<sale_bill_uuid>
```

### Step 6: 提交支付

```bash
POST /member/order/dine_in/pay
Body: {"sale_bill_uuid": <sale_bill_uuid>, "payment_method_uuid": <payment_method_uuid>}
```

### Step 7: 模拟支付回调（测试用）

```bash
POST /member/order/dine_in/pay/mock_callback
Body: {"sale_bill_uuid": <sale_bill_uuid>}
```

**注意**：此接口直接将订单标记为已支付，跳过真实支付流程。可在 Step 6 之后或直接替代 Step 5-6 使用。

### Step 8: 查询支付状态（可选）

```bash
GET /member/order/dine_in/pay/status?sale_bill_uuid=<sale_bill_uuid>
```

### Step 9: 查询订单列表

```bash
GET /member/order/dine_in/list
GET /member/order/dine_in/list?status=<status>
```

`status` 可选值：`unpaid`、`pending`、`preparing`、`completed`、`cancelled`、`rejected`

### Step 10: 查询订单详情

```bash
GET /member/order/dine_in/detail?sale_bill_uuid=<sale_bill_uuid>
```

### Step 11: 取消订单（可选）

```bash
POST /member/order/dine_in/cancel
Body: {"sale_bill_uuid": <sale_bill_uuid>}
```

## 收银端接单/拒单流程

### 获取待接单列表

```bash
GET /cashier/h5_order/list?status=0&order_type=1
```

`order_type=1` 表示会员端堂食订单。响应中每个订单有 `h5_order_uuid`。

### 获取已处理列表

```bash
GET /cashier/h5_order/list?status=1&order_type=1
```

### 查看订单详情

```bash
GET /cashier/h5_order/detail?h5_order_uuid=<h5_order_uuid>
```

### 接单

```bash
POST /cashier/h5_order/accept
Body: {"h5_order_uuid": <h5_order_uuid>}
```

接单后会员端订单状态变为 `preparing`（备餐中）。

### 拒单

```bash
POST /cashier/h5_order/reject
Body: {"h5_order_uuid": <h5_order_uuid>}
```

拒单后会员端订单状态变为 `rejected`（已拒单）。

## 订单状态流转

```
创建订单 → unpaid（待支付）
    ↓ 支付成功
pending（待接单）
    ├── 收银端接单 → preparing（备餐中） → completed（已完成）
    ├── 收银端拒单 → rejected（已拒单）
    └── 会员取消 → cancelled（已取消）
```

## 测试数据参考

### 当前可用商品

| 商品 | product_uuid | flavor_uuid | 价格 | 类型 |
|------|-------------|-------------|------|------|
| 苹果派 | 3717245672884226 | 3717245672884228 | ¥16 | 普通 |
| 香辣鸡肉卷 | 3717245645621250 | - | ¥44 | 普通 |
| 拿铁咖啡 | 3717245576415234 | - | ¥36 | 普通 |
| 四人餐 | 3718708423821314 | 3718708423821316 | ¥124 | 套餐 |

### 四人餐套餐子商品

分组: "任选"（uuid=3718708423821320），需选 2 个：
- 巨无霸 flavor_uuid=3717245297494017
- 麦辣鸡腿堡 flavor_uuid=3717245473654788

### 数据库信息

- **DB Host**: 10.180.10.10:13306
- **DB User**: saas
- **DB Password**: 11ca2c16594c7878
- **商户数据库**: shop7199984230400000
- **JWT Secret**: dkjhd00a08

## 注意事项

1. curl 输出通过管道传给 python3 时可能丢失数据，建议先 `> /tmp/xxx.json` 再解析
2. 会员端 Token 有效期很长（~100年），但可能因服务重启失效，失效时重新调用 visitor/login
3. 收银端 Token 通过 JWT 直接生成，需要使用已有的设备 UUID 和 Device-Id
4. 所有 API 请求必须通过 `http://localhost:8080` 代理访问，禁止直接访问 Go 服务端口
