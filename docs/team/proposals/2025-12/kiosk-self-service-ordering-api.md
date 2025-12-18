# Kiosk 自助点餐机终端 API 接口列表

> 本文档定义 Kiosk 自助点餐机终端功能所需的 API 接口列表。
> 
> **说明**：本文档仅列出接口名称和路径，不包含具体的请求参数和返回值定义。详细的接口规范将在技术设计阶段补充。

---

## 📋 接口说明

- **基础路径**: `/api/v1/kiosk`
- **认证方式**: JWT Token（除登录接口外）
- **参考实现**: 收银机（Cashier）、平板端（Tablet）、会员端（Member）等终端的接口实现
- **全局接口**: 验证码接口使用全局路径 `/api/v1/passport/captcha`（不在基础路径下）

---

## 1. 登录认证模块

> 参考收银机（Cashier）的登录认证接口实现

| 接口名称 | 接口路径 | 请求方法 | 说明 |
|---------|---------|---------|------|
| 获取验证码 | `/api/v1/passport/captcha` | GET | 获取图形验证码（返回 sign 和 base64 图片，全局接口） |
| 登录 | `/api/v1/kiosk/login` | POST | 使用商家员工账号登录，支持邮箱/手机号登录，验证码验证 |
| 刷新Token | `/api/v1/kiosk/refresh_token` | GET | 刷新访问令牌 |
| 退出登录 | `/api/v1/kiosk/logout` | POST | 退出登录，清除登录状态 |

**参考接口**:
- `/api/v1/passport/captcha` - 通用验证码接口（全局接口，所有终端共用）
- `/api/v1/cashier/login` - 收银机登录接口
- `/api/v1/cashier/refresh_token` - 收银机刷新Token接口
- `/api/v1/cashier/logout` - 收银机退出登录接口

---

## 2. 首页功能模块

| 接口名称 | 接口路径 | 请求方法 | 说明 |
|---------|---------|---------|------|
| 获取基本信息 | `/api/v1/kiosk/base` | GET | 获取基本信息（包含轮播广告、呼叫服务员、语言设置等配置，参考其他终端的 base 接口） |
| 呼叫服务员 | `/api/v1/kiosk/call` | POST | 发起呼叫服务员请求 |

**参考接口**:
- `/api/v1/cashier/base` - 收银端基本信息接口（包含设置信息）
- `/api/v1/tablet/base` - 平板端基本信息接口（包含设置信息）
- `/api/v1/assistant/base` - 助手端基本信息接口（包含设置信息）
- `/api/v1/shop/setting/kiosk` - 商家端自助点餐机设置接口（响应中包含 `language_list` 字段，用于参考数据结构）
- `/api/v1/tablet/call` - 平板端呼叫服务员接口
- `/api/v1/h5/call` - H5端呼叫服务员接口

---

## 3. 商品浏览与选择模块

| 接口名称 | 接口路径 | 请求方法 | 说明 |
|---------|---------|---------|------|
| 获取商品分类列表 | `/api/v1/kiosk/product/category/list` | GET | 获取商品分类列表，用于分类导航 |
| 获取商品列表 | `/api/v1/kiosk/product/list` | GET | 获取商品列表，支持分页和分类筛选 |
| 获取商品详情 | `/api/v1/kiosk/product/detail` | GET | 获取商品详细信息，包括规格、属性、加料等选项 |

**参考接口**:
- `/tablet/product/category/list` - 平板端商品分类列表接口
- `/tablet/product/list` - 平板端商品列表接口
- `/member/product/detail` - 会员端商品详情接口

---

## 4. 购物车管理模块

> 参考 POS 端即时点餐模块的购物车管理接口

| 接口名称 | 接口路径 | 请求方法 | 说明 |
|---------|---------|---------|------|
| 查询购物车信息 | `/api/v1/kiosk/order/cart/info` | GET | 查询购物车信息（包含商品列表、价格计算、订单总价等） |
| 添加商品到购物车 | `/api/v1/kiosk/order/cart/product/add` | POST | 向购物车添加商品（若没有订单则自动创建） |
| 添加套餐到购物车 | `/api/v1/kiosk/order/cart/product_package/add` | POST | 向购物车添加套餐 |
| 修改商品数量 | `/api/v1/kiosk/order/cart/product/num` | POST | 修改购物车中商品的数量 |
| 查询商品规格属性 | `/api/v1/kiosk/order/cart/product/flavor_and_attribute` | GET | 查询购物车商品的规格和属性信息 |
| 修改商品规格属性 | `/api/v1/kiosk/order/cart/product/flavor_and_attribute` | POST | 修改购物车商品的规格和属性 |
| 删除购物车商品 | `/api/v1/kiosk/order/cart/product/delete` | DELETE | 删除购物车中的商品 |

**参考接口**:
- `/cashier/instant/order/cart/info` - POS 端查询购物车信息接口
- `/cashier/instant/order/cart/product/add` - POS 端添加商品接口
- `/cashier/instant/order/cart/product_package/add` - POS 端添加套餐接口
- `/cashier/instant/order/cart/product/num` - POS 端修改商品数量接口
- `/cashier/instant/order/cart/product/flavor_and_attribute` - POS 端查询/修改商品规格属性接口
- `/cashier/instant/order/product/delete` - POS 端删除商品接口

---

## 5. 订单确认与创建模块

> 参考 POS 端即时点餐模块的订单确认接口

| 接口名称 | 接口路径 | 请求方法 | 说明 |
|---------|---------|---------|------|
| 订单检查 | `/api/v1/kiosk/order/check` | GET | 订单检查（点击结账时检查订单是否可以结账） |
| 获取结账页面信息 | `/api/v1/kiosk/order/payment/info` | GET | 获取结账页面信息（包含支付方式列表、订单总价、优惠信息等） |

**参考接口**:
- `/cashier/instant/order/check` - POS 端订单检查接口
- `/cashier/instant/order/payment/info` - POS 端获取结账页面信息接口（包含支付方式列表、价格计算等）

---

## 6. 支付功能模块

> 参考会员端的支付接口实现

| 接口名称 | 接口路径 | 请求方法 | 说明 |
|---------|---------|---------|------|
| 提交支付 | `/api/v1/kiosk/order/pay` | POST | 提交支付（选择支付方式，订单状态变为待支付） |
| 获取支付信息 | `/api/v1/kiosk/order/pay/info` | GET | 获取支付信息（创建支付订单，返回二维码和支付信息，可轮询此接口查询支付状态） |
| 获取支付状态 | `/api/v1/kiosk/order/pay/status` | GET | 获取支付状态（查询订单的支付状态） |

**参考接口**:
- `/member/order/pay` - 会员端提交支付接口
- `/member/order/pay/info` - 会员端获取支付信息接口（包含二维码、支付状态等）
- `/member/order/pay/status` - 会员端获取支付状态接口

---

## 7. 异常处理模块

> 参考 POS 端即时点餐模块的异常处理接口

| 接口名称 | 接口路径 | 请求方法 | 说明 |
|---------|---------|---------|------|
| 取消订单 | `/api/v1/kiosk/order/cancel` | POST | 取消订单（用于异常情况下的订单取消） |

**参考接口**:
- `/cashier/instant/order/cancel` - POS 端取消订单接口

---

## 📊 接口统计

| 模块 | 接口数量 | 说明 |
|-----|---------|------|
| 登录认证模块 | 3 | 登录、刷新Token、退出登录（验证码使用全局接口 `/passport/captcha`） |
| 首页功能模块 | 2 | 设置（包含语言列表）、呼叫服务员 |
| 商品浏览与选择模块 | 3 | 分类列表、商品列表、商品详情 |
| 购物车管理模块 | 7 | 购物车信息、添加商品/套餐、修改数量、规格属性、删除商品 |
| 订单确认与创建模块 | 2 | 订单检查、结账页面信息 |
| 支付功能模块 | 3 | 提交支付、获取支付信息（含二维码）、获取支付状态 |
| 异常处理模块 | 1 | 取消订单 |
| **总计** | **21** | - |

---

## 🔗 相关文档

- [Kiosk 自助点餐机终端功能需求提案](./v2.12.0-kiosk-self-service-ordering.md)
- [API 设计规范](../../../.cursor/rules/api.mdc)
- [收银端 API 参考](../../../shared/api/)

---

**版本**: v1.0.0  
**创建日期**: 2025-12-17  
**维护者**: 开发团队

